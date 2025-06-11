package sandbox

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type ContainerSandbox struct {
	client      *client.Client
	containerID string
	config      *SandboxConfig
	metrics     *ResourceMetrics
}

type SandboxConfig struct {
	Image          string
	WorkingDir     string
	Environment    []string
	ResourceLimits ResourceLimits
	NetworkPolicy  NetworkPolicy
	TimeoutSeconds int64
	ReadOnly       bool
	NoNetwork      bool
}

type ResourceLimits struct {
	CPUQuota   int64  // CPU quota in microseconds (100000 = 1 CPU)
	CPUPeriod  int64  // CPU period in microseconds  
	Memory     int64  // Memory limit in bytes
	MemorySwap int64  // Memory + swap limit in bytes
	PidsLimit  *int64 // Maximum number of processes
	DiskQuota  int64  // Disk quota in bytes
}

type NetworkPolicy struct {
	AllowOutbound bool
	AllowedHosts  []string
	BlockedPorts  []string
}

type ResourceMetrics struct {
	CPUUsagePercent float64
	MemoryUsageBytes int64
	NetworkRxBytes  int64
	NetworkTxBytes  int64
	DiskUsageBytes  int64
	ProcessCount    int
	StartTime       time.Time
	EndTime         *time.Time
}

func NewContainerSandbox(config *SandboxConfig) (*ContainerSandbox, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &ContainerSandbox{
		client:  cli,
		config:  config,
		metrics: &ResourceMetrics{},
	}, nil
}

func (cs *ContainerSandbox) Execute(ctx context.Context, command []string, stdin string) (*ExecutionResult, error) {
	if err := cs.pullImage(ctx); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	containerConfig := cs.buildContainerConfig(command)
	hostConfig := cs.buildHostConfig()
	networkConfig := cs.buildNetworkConfig()

	resp, err := cs.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	cs.containerID = resp.ID
	defer cs.cleanup(context.Background())

	if err := cs.client.ContainerStart(ctx, cs.containerID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	cs.metrics.StartTime = time.Now()

	if stdin != "" {
		if err := cs.writeStdin(ctx, stdin); err != nil {
			return nil, fmt.Errorf("failed to write stdin: %w", err)
		}
	}

	result, err := cs.waitForCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	if err := cs.collectMetrics(ctx); err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	return result, nil
}

func (cs *ContainerSandbox) pullImage(ctx context.Context) error {
	reader, err := cs.client.ImagePull(ctx, cs.config.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (cs *ContainerSandbox) buildContainerConfig(command []string) *container.Config {
	config := &container.Config{
		Image:        cs.config.Image,
		Cmd:          command,
		Env:          cs.config.Environment,
		WorkingDir:   cs.config.WorkingDir,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		StdinOnce:    true,
		Tty:          false,
	}

	if cs.config.NoNetwork {
		config.NetworkDisabled = true
	}

	return config
}

func (cs *ContainerSandbox) buildHostConfig() *container.HostConfig {
	hostConfig := &container.HostConfig{
		ReadonlyRootfs: cs.config.ReadOnly,
		Resources: container.Resources{
			CPUQuota:   cs.config.ResourceLimits.CPUQuota,
			CPUPeriod:  cs.config.ResourceLimits.CPUPeriod,
			Memory:     cs.config.ResourceLimits.Memory,
			MemorySwap: cs.config.ResourceLimits.MemorySwap,
			PidsLimit:  cs.config.ResourceLimits.PidsLimit,
		},
		SecurityOpt: []string{
			"no-new-privileges:true",
			"seccomp:unconfined",
		},
		CapDrop: []string{"ALL"},
		CapAdd:  []string{"CHOWN", "DAC_OVERRIDE", "FOWNER", "SETGID", "SETUID"},
	}

	if cs.config.WorkingDir != "" {
		hostConfig.Mounts = []mount.Mount{
			{
				Type:     mount.TypeTmpfs,
				Target:   cs.config.WorkingDir,
				TmpfsOptions: &mount.TmpfsOptions{
					SizeBytes: cs.config.ResourceLimits.DiskQuota,
					Mode:      0755,
				},
			},
		}
	}

	if cs.config.NoNetwork {
		hostConfig.NetworkMode = "none"
	}

	return hostConfig
}

func (cs *ContainerSandbox) buildNetworkConfig() *network.NetworkingConfig {
	if cs.config.NoNetwork {
		return &network.NetworkingConfig{}
	}

	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"bridge": {},
		},
	}
}

func (cs *ContainerSandbox) writeStdin(ctx context.Context, stdin string) error {
	hijackedResp, err := cs.client.ContainerAttach(ctx, cs.containerID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return err
	}
	defer hijackedResp.Close()

	_, err = hijackedResp.Conn.Write([]byte(stdin))
	if err != nil {
		return err
	}

	return hijackedResp.CloseWrite()
}

func (cs *ContainerSandbox) waitForCompletion(ctx context.Context) (*ExecutionResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cs.config.TimeoutSeconds)*time.Second)
	defer cancel()

	statusCh, errCh := cs.client.ContainerWait(timeoutCtx, cs.containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("container wait error: %w", err)
		}
	case status := <-statusCh:
		now := time.Now()
		cs.metrics.EndTime = &now

		stdout, stderr, err := cs.getLogs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs: %w", err)
		}

		return &ExecutionResult{
			ExitCode: int(status.StatusCode),
			Stdout:   stdout,
			Stderr:   stderr,
			Duration: now.Sub(cs.metrics.StartTime),
			Metrics:  cs.metrics,
		}, nil
	case <-timeoutCtx.Done():
		cs.kill(context.Background())
		return nil, fmt.Errorf("execution timeout after %d seconds", cs.config.TimeoutSeconds)
	}

	return nil, fmt.Errorf("unexpected wait completion")
}

func (cs *ContainerSandbox) getLogs(ctx context.Context) (string, string, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	logs, err := cs.client.ContainerLogs(ctx, cs.containerID, options)
	if err != nil {
		return "", "", err
	}
	defer logs.Close()

	content, err := io.ReadAll(logs)
	if err != nil {
		return "", "", err
	}

	logContent := string(content)
	parts := strings.SplitN(logContent, "\n", 2)
	
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	
	return logContent, "", nil
}

func (cs *ContainerSandbox) collectMetrics(ctx context.Context) error {
	stats, err := cs.client.ContainerStats(ctx, cs.containerID, false)
	if err != nil {
		return err
	}
	defer stats.Body.Close()

	var containerStats types.StatsJSON
	if err := stats.Body.Close(); err != nil {
		return err
	}

	cs.metrics.CPUUsagePercent = calculateCPUPercent(&containerStats)
	cs.metrics.MemoryUsageBytes = int64(containerStats.MemoryStats.Usage)
	cs.metrics.NetworkRxBytes = calculateNetworkRx(&containerStats)
	cs.metrics.NetworkTxBytes = calculateNetworkTx(&containerStats)
	cs.metrics.ProcessCount = int(containerStats.PidsStats.Current)

	return nil
}

func (cs *ContainerSandbox) kill(ctx context.Context) error {
	return cs.client.ContainerKill(ctx, cs.containerID, "SIGKILL")
}

func (cs *ContainerSandbox) cleanup(ctx context.Context) error {
	return cs.client.ContainerRemove(ctx, cs.containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (cs *ContainerSandbox) GetMetrics() *ResourceMetrics {
	return cs.metrics
}

type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Metrics  *ResourceMetrics
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return 0.0
}

func calculateNetworkRx(stats *types.StatsJSON) int64 {
	var rx int64
	for _, network := range stats.Networks {
		rx += int64(network.RxBytes)
	}
	return rx
}

func calculateNetworkTx(stats *types.StatsJSON) int64 {
	var tx int64
	for _, network := range stats.Networks {
		tx += int64(network.TxBytes)
	}
	return tx
}

func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		Image:      "alpine:latest",
		WorkingDir: "/workspace",
		Environment: []string{
			"HOME=/tmp",
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		ResourceLimits: ResourceLimits{
			CPUQuota:   50000,      // 0.5 CPU cores
			CPUPeriod:  100000,     // Standard period
			Memory:     512 * 1024 * 1024, // 512MB
			MemorySwap: 512 * 1024 * 1024, // No swap
			PidsLimit:  int64Ptr(256),      // Max 256 processes
			DiskQuota:  1024 * 1024 * 1024, // 1GB
		},
		NetworkPolicy: NetworkPolicy{
			AllowOutbound: false,
			AllowedHosts:  []string{},
			BlockedPorts:  []string{"22", "23", "25", "53", "80", "443"},
		},
		TimeoutSeconds: 300, // 5 minutes
		ReadOnly:       true,
		NoNetwork:      true,
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}