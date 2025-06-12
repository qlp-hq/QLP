package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/worker-runtime/pkg/contracts"
)

type Manager struct {
	config       Config
	dockerClient *client.Client
	activeJobs   map[string]*ExecutionJob
}

type Config struct {
	Runtime             string
	BaseImage           string
	NetworkIsolation    bool
	FileSystemIsolation bool
	ResourceLimits      ResourceLimits
	WorkDir             string
}

type ResourceLimits struct {
	CPUQuota    int64  // CPU quota in microseconds (100000 = 1 CPU)
	MemoryMB    int64  // Memory limit in MB
	TimeoutSec  int    // Execution timeout
	DiskSpaceMB int64  // Disk space limit
}

type ExecutionJob struct {
	ID           string
	ContainerID  string
	TenantID     string
	StartTime    time.Time
	Context      context.Context
	Cancel       context.CancelFunc
	OutputChan   chan string
	ErrorChan    chan error
}

func NewManager(cfg Config) (*Manager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	if cfg.WorkDir == "" {
		cfg.WorkDir = "/tmp/qlp-sandbox"
	}

	// Ensure work directory exists
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	manager := &Manager{
		config:       cfg,
		dockerClient: dockerClient,
		activeJobs:   make(map[string]*ExecutionJob),
	}

	// Pull base image if not exists
	if err := manager.ensureBaseImage(); err != nil {
		logger.WithComponent("sandbox").Warn("Failed to pull base image", zap.Error(err))
	}

	return manager, nil
}

func (m *Manager) ExecuteTask(ctx context.Context, task *contracts.WorkerTask) (*contracts.SandboxResult, error) {
	jobID := uuid.New().String()
	
	logger.WithComponent("sandbox").Info("Starting task execution",
		zap.String("job_id", jobID),
		zap.String("task_id", task.ID),
		zap.String("tenant_id", task.TenantID),
		zap.String("task_type", string(task.Type)))

	// Create execution context with timeout
	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = time.Duration(m.config.ResourceLimits.TimeoutSec) * time.Second
	}
	
	jobCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create work directory for this job
	jobWorkDir := filepath.Join(m.config.WorkDir, jobID)
	if err := os.MkdirAll(jobWorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job work directory: %w", err)
	}
	defer os.RemoveAll(jobWorkDir)

	// Write task code to file if provided
	var scriptPath string
	if task.Code != "" {
		scriptPath = filepath.Join(jobWorkDir, "main."+getFileExtension(task.Language))
		if err := os.WriteFile(scriptPath, []byte(task.Code), 0644); err != nil {
			return nil, fmt.Errorf("failed to write task code: %w", err)
		}
	}

	// Create container
	containerID, err := m.createContainer(jobCtx, task, jobWorkDir, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	defer m.cleanupContainer(containerID)

	// Register job
	job := &ExecutionJob{
		ID:          jobID,
		ContainerID: containerID,
		TenantID:    task.TenantID,
		StartTime:   time.Now(),
		Context:     jobCtx,
		Cancel:      cancel,
		OutputChan:  make(chan string, 100),
		ErrorChan:   make(chan error, 1),
	}
	m.activeJobs[jobID] = job
	defer delete(m.activeJobs, jobID)

	// Start container
	if err := m.dockerClient.ContainerStart(jobCtx, containerID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Collect results
	return m.collectResults(jobCtx, job, task)
}

func (m *Manager) createContainer(ctx context.Context, task *contracts.WorkerTask, workDir, scriptPath string) (string, error) {
	// Determine command based on language
	cmd := m.buildCommand(task, scriptPath)
	
	// Resource limits
	limits := m.getResourceLimits(task.ResourceLimits)
	
	// Container config
	containerConfig := &container.Config{
		Image:        m.getImageForLanguage(task.Language),
		Cmd:          cmd,
		WorkingDir:   "/workspace",
		Env:          []string{"QLP_TENANT_ID=" + task.TenantID},
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    false,
		Tty:          false,
	}

	// Host config with resource limits
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   limits.MemoryMB * 1024 * 1024, // Convert MB to bytes
			CPUQuota: limits.CPUQuota,
			CPUPeriod: 100000, // 100ms period
		},
		NetworkMode: container.NetworkMode("none"), // No network by default
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			"/tmp": "rw,size=100m",
		},
		Binds: []string{
			workDir + ":/workspace:ro", // Read-only workspace
		},
	}

	// Allow network access if specified
	if task.ResourceLimits != nil && task.ResourceLimits.NetworkAccess {
		hostConfig.NetworkMode = container.NetworkMode("bridge")
	}

	// Allow filesystem writes if specified
	if task.ResourceLimits != nil && task.ResourceLimits.FileSystemRW {
		hostConfig.ReadonlyRootfs = false
		hostConfig.Binds[0] = strings.Replace(hostConfig.Binds[0], ":ro", ":rw", 1)
	}

	// Network config
	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := m.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (m *Manager) collectResults(ctx context.Context, job *ExecutionJob, task *contracts.WorkerTask) (*contracts.SandboxResult, error) {
	startTime := time.Now()
	
	// Get container logs
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logs, err := m.dockerClient.ContainerLogs(ctx, job.ContainerID, logOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	// Read logs
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}

	// Docker multiplexes stdout/stderr, need to demux
	_, err = io.Copy(stdout, logs)
	if err != nil && err != io.EOF {
		logger.WithComponent("sandbox").Warn("Error reading logs", zap.Error(err))
	}

	// Wait for container to finish
	statusCh, errCh := m.dockerClient.ContainerWait(ctx, job.ContainerID, container.WaitConditionNotRunning)
	
	var exitCode int64
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("container wait error: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	case <-ctx.Done():
		// Timeout or cancellation
		m.dockerClient.ContainerKill(context.Background(), job.ContainerID, "SIGKILL")
		return &contracts.SandboxResult{
			ExitCode:      124, // Timeout exit code
			Stdout:        stdout.String(),
			Stderr:        "Execution timed out",
			ExecutionTime: time.Since(startTime),
			ResourceUsage: &contracts.ResourceUsage{},
		}, nil
	}

	// Get container stats for resource usage
	resourceUsage, err := m.getResourceUsage(job.ContainerID)
	if err != nil {
		logger.WithComponent("sandbox").Warn("Failed to get resource usage", zap.Error(err))
		resourceUsage = &contracts.ResourceUsage{}
	}

	result := &contracts.SandboxResult{
		ExitCode:      int(exitCode),
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: time.Since(startTime),
		ResourceUsage: resourceUsage,
		Files:         []contracts.FileOutput{}, // TODO: Implement file collection
		NetworkCalls:  []contracts.NetworkCall{}, // TODO: Implement network monitoring
		SecurityViolations: []contracts.SecurityViolation{}, // TODO: Implement security monitoring
	}

	logger.WithComponent("sandbox").Info("Task execution completed",
		zap.String("job_id", job.ID),
		zap.String("task_id", task.ID),
		zap.Int("exit_code", result.ExitCode),
		zap.Duration("execution_time", result.ExecutionTime))

	return result, nil
}

func (m *Manager) getResourceUsage(containerID string) (*contracts.ResourceUsage, error) {
	stats, err := m.dockerClient.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return nil, err
	}

	return &contracts.ResourceUsage{
		CPUTimeMs:    int64(statsJSON.CPUStats.CPUUsage.TotalUsage / 1000000), // Convert nanoseconds to milliseconds
		MemoryPeakMB: int64(statsJSON.MemoryStats.MaxUsage / 1024 / 1024),     // Convert bytes to MB
		DiskReadMB:   int64(statsJSON.BlkioStats.IoServiceBytesRecursive[0].Value / 1024 / 1024),
		DiskWriteMB:  int64(statsJSON.BlkioStats.IoServiceBytesRecursive[1].Value / 1024 / 1024),
		NetworkInMB:  int64(statsJSON.Networks["eth0"].RxBytes / 1024 / 1024),
		NetworkOutMB: int64(statsJSON.Networks["eth0"].TxBytes / 1024 / 1024),
	}, nil
}

func (m *Manager) cleanupContainer(containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Remove container
	err := m.dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		logger.WithComponent("sandbox").Warn("Failed to remove container", 
			zap.String("container_id", containerID),
			zap.Error(err))
	}
}

func (m *Manager) Cleanup() {
	// Cancel all active jobs
	for _, job := range m.activeJobs {
		job.Cancel()
		m.cleanupContainer(job.ContainerID)
	}
	
	// Close Docker client
	if m.dockerClient != nil {
		m.dockerClient.Close()
	}
}

// Helper functions
func (m *Manager) ensureBaseImage() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	reader, err := m.dockerClient.ImagePull(ctx, m.config.BaseImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	
	// Consume the pull output
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (m *Manager) buildCommand(task *contracts.WorkerTask, scriptPath string) []string {
	switch task.Language {
	case "go":
		return []string{"go", "run", "/workspace/main.go"}
	case "python":
		return []string{"python3", "/workspace/main.py"}
	case "node", "javascript":
		return []string{"node", "/workspace/main.js"}
	case "bash", "shell":
		return []string{"bash", "/workspace/main.sh"}
	default:
		// Generic execution
		if scriptPath != "" {
			return []string{"cat", "/workspace/" + filepath.Base(scriptPath)}
		}
		return []string{"echo", task.Description}
	}
}

func (m *Manager) getImageForLanguage(language string) string {
	switch language {
	case "go":
		return "golang:1.21-alpine"
	case "python":
		return "python:3.11-alpine"
	case "node", "javascript":
		return "node:18-alpine"
	case "bash", "shell":
		return "alpine:latest"
	default:
		return m.config.BaseImage
	}
}

func (m *Manager) getResourceLimits(taskLimits *contracts.ResourceLimits) ResourceLimits {
	limits := m.config.ResourceLimits
	
	if taskLimits != nil {
		if taskLimits.CPUMillicores > 0 {
			limits.CPUQuota = taskLimits.CPUMillicores * 100 // Convert millicores to microseconds
		}
		if taskLimits.MemoryMB > 0 {
			limits.MemoryMB = taskLimits.MemoryMB
		}
		if taskLimits.TimeoutSec > 0 {
			limits.TimeoutSec = taskLimits.TimeoutSec
		}
	}
	
	return limits
}

func getFileExtension(language string) string {
	switch language {
	case "go":
		return "go"
	case "python":
		return "py"
	case "node", "javascript":
		return "js"
	case "bash", "shell":
		return "sh"
	default:
		return "txt"
	}
}

func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		CPUQuota:    100000, // 1 CPU
		MemoryMB:    256,    // 256MB
		TimeoutSec:  300,    // 5 minutes
		DiskSpaceMB: 1024,   // 1GB
	}
}