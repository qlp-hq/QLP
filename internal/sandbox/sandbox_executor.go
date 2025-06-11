package sandbox

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"QLP/internal/models"
)

type SandboxedExecutor struct {
	defaultConfig *SandboxConfig
}

func NewSandboxedExecutor() *SandboxedExecutor {
	return &SandboxedExecutor{
		defaultConfig: DefaultSandboxConfig(),
	}
}

func (se *SandboxedExecutor) Execute(ctx context.Context, task models.Task, agentOutput string) (*SandboxExecutionResult, error) {
	config := se.buildTaskSpecificConfig(task)
	
	sandbox, err := NewContainerSandbox(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	commands := se.parseAgentOutputToCommands(task, agentOutput)
	if len(commands) == 0 {
		return &SandboxExecutionResult{
			TaskID:        task.ID,
			Success:       true,
			Output:        agentOutput,
			ExecutionTime: 0,
			SecurityScore: 100,
			Message:       "No executable commands found, treating as documentation/analysis output",
		}, nil
	}

	log.Printf("Executing %d commands in sandbox for task %s", len(commands), task.ID)

	var results []CommandResult
	var totalDuration time.Duration
	securityScore := 100

	for i, cmd := range commands {
		log.Printf("Executing command %d/%d: %s", i+1, len(commands), cmd.Description)
		
		result, err := sandbox.Execute(ctx, cmd.Command, cmd.Stdin)
		if err != nil {
			return &SandboxExecutionResult{
				TaskID:        task.ID,
				Success:       false,
				Output:        fmt.Sprintf("Command failed: %s", err.Error()),
				ExecutionTime: totalDuration,
				SecurityScore: 0,
				Message:       fmt.Sprintf("Sandbox execution failed at command %d", i+1),
				Results:       results,
			}, nil
		}

		totalDuration += result.Duration
		
		cmdResult := CommandResult{
			Command:     strings.Join(cmd.Command, " "),
			ExitCode:    result.ExitCode,
			Stdout:      result.Stdout,
			Stderr:      result.Stderr,
			Duration:    result.Duration,
			Metrics:     result.Metrics,
		}
		results = append(results, cmdResult)

		if result.ExitCode != 0 {
			securityScore -= 10
		}
		
		if se.detectSuspiciousActivity(result) {
			securityScore -= 20
		}
	}

	success := len(results) > 0 && results[len(results)-1].ExitCode == 0
	output := se.aggregateOutput(results)

	return &SandboxExecutionResult{
		TaskID:        task.ID,
		Success:       success,
		Output:        output,
		ExecutionTime: totalDuration,
		SecurityScore: securityScore,
		Message:       fmt.Sprintf("Executed %d commands successfully", len(results)),
		Results:       results,
	}, nil
}

func (se *SandboxedExecutor) buildTaskSpecificConfig(task models.Task) *SandboxConfig {
	config := &SandboxConfig{
		Image:          se.getImageForTaskType(task.Type),
		WorkingDir:     "/workspace",
		Environment:    se.getEnvironmentForTaskType(task.Type),
		ResourceLimits: se.getResourceLimitsForTaskType(task.Type),
		NetworkPolicy:  se.getNetworkPolicyForTaskType(task.Type),
		TimeoutSeconds: se.getTimeoutForTaskType(task.Type),
		ReadOnly:       false, // Allow writes for code generation
		NoNetwork:      se.shouldDisableNetwork(task.Type),
	}

	return config
}

func (se *SandboxedExecutor) getImageForTaskType(taskType models.TaskType) string {
	switch taskType {
	case models.TaskTypeCodegen:
		return "golang:1.21-alpine"
	case models.TaskTypeTest:
		return "golang:1.21-alpine"
	case models.TaskTypeInfra:
		return "alpine/terragrunt:latest"
	case models.TaskTypeDoc:
		return "pandoc/core:latest"
	case models.TaskTypeAnalyze:
		return "sonarsource/sonar-scanner-cli:latest"
	default:
		return "alpine:latest"
	}
}

func (se *SandboxedExecutor) getEnvironmentForTaskType(taskType models.TaskType) []string {
	baseEnv := []string{
		"HOME=/tmp",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"TERM=xterm",
	}

	switch taskType {
	case models.TaskTypeCodegen, models.TaskTypeTest:
		return append(baseEnv,
			"GOPATH=/workspace/go",
			"GOCACHE=/tmp/go-cache",
			"GOMODCACHE=/tmp/go-mod-cache",
			"CGO_ENABLED=0",
		)
	case models.TaskTypeInfra:
		return append(baseEnv,
			"TF_IN_AUTOMATION=1",
			"TF_LOG=ERROR",
		)
	default:
		return baseEnv
	}
}

func (se *SandboxedExecutor) getResourceLimitsForTaskType(taskType models.TaskType) ResourceLimits {
	switch taskType {
	case models.TaskTypeCodegen, models.TaskTypeTest:
		return ResourceLimits{
			CPUQuota:   100000,     // 1.0 CPU cores for compilation
			CPUPeriod:  100000,
			Memory:     1024 * 1024 * 1024, // 1GB for Go compilation
			MemorySwap: 1024 * 1024 * 1024,
			PidsLimit:  int64Ptr(512),
			DiskQuota:  2048 * 1024 * 1024, // 2GB for code and dependencies
		}
	case models.TaskTypeInfra:
		return ResourceLimits{
			CPUQuota:   50000,      // 0.5 CPU cores
			CPUPeriod:  100000,
			Memory:     512 * 1024 * 1024, // 512MB
			MemorySwap: 512 * 1024 * 1024,
			PidsLimit:  int64Ptr(256),
			DiskQuota:  1024 * 1024 * 1024, // 1GB
		}
	case models.TaskTypeAnalyze:
		return ResourceLimits{
			CPUQuota:   200000,     // 2.0 CPU cores for analysis
			CPUPeriod:  100000,
			Memory:     2048 * 1024 * 1024, // 2GB for analysis tools
			MemorySwap: 2048 * 1024 * 1024,
			PidsLimit:  int64Ptr(1024),
			DiskQuota:  4096 * 1024 * 1024, // 4GB
		}
	default:
		return se.defaultConfig.ResourceLimits
	}
}

func (se *SandboxedExecutor) getNetworkPolicyForTaskType(taskType models.TaskType) NetworkPolicy {
	switch taskType {
	case models.TaskTypeCodegen, models.TaskTypeTest:
		return NetworkPolicy{
			AllowOutbound: true, // Need to download dependencies
			AllowedHosts:  []string{"proxy.golang.org", "sum.golang.org", "github.com"},
			BlockedPorts:  []string{"22", "23", "25"},
		}
	case models.TaskTypeInfra:
		return NetworkPolicy{
			AllowOutbound: true, // Need cloud provider APIs
			AllowedHosts:  []string{"amazonaws.com", "azure.com", "googleapis.com"},
			BlockedPorts:  []string{"22", "23", "25"},
		}
	default:
		return NetworkPolicy{
			AllowOutbound: false,
			AllowedHosts:  []string{},
			BlockedPorts:  []string{"22", "23", "25", "53", "80", "443"},
		}
	}
}

func (se *SandboxedExecutor) getTimeoutForTaskType(taskType models.TaskType) int64 {
	switch taskType {
	case models.TaskTypeCodegen:
		return 600 // 10 minutes for compilation
	case models.TaskTypeTest:
		return 900 // 15 minutes for tests
	case models.TaskTypeInfra:
		return 1800 // 30 minutes for infrastructure
	case models.TaskTypeAnalyze:
		return 1200 // 20 minutes for analysis
	default:
		return 300 // 5 minutes default
	}
}

func (se *SandboxedExecutor) shouldDisableNetwork(taskType models.TaskType) bool {
	switch taskType {
	case models.TaskTypeDoc, models.TaskTypeAnalyze:
		return true // Documentation and analysis don't need network
	default:
		return false // Others may need dependency downloads
	}
}

func (se *SandboxedExecutor) parseAgentOutputToCommands(task models.Task, output string) []SandboxCommand {
	var commands []SandboxCommand

	switch task.Type {
	case models.TaskTypeCodegen, models.TaskTypeTest:
		commands = se.parseGoCommands(output)
	case models.TaskTypeInfra:
		commands = se.parseInfraCommands(output)
	case models.TaskTypeDoc:
		commands = se.parseDocCommands(output)
	case models.TaskTypeAnalyze:
		commands = se.parseAnalysisCommands(output)
	}

	return commands
}

func (se *SandboxedExecutor) parseGoCommands(output string) []SandboxCommand {
	var commands []SandboxCommand

	// Extract Go code blocks and create files
	if codeBlocks := extractCodeBlocks(output, "go"); len(codeBlocks) > 0 {
		for i, block := range codeBlocks {
			filename := fmt.Sprintf("main_%d.go", i)
			if strings.Contains(block, "package main") {
				filename = "main.go"
			} else if strings.Contains(block, "_test.go") {
				filename = fmt.Sprintf("test_%d_test.go", i)
			}

			commands = append(commands, SandboxCommand{
				Command:     []string{"sh", "-c", fmt.Sprintf("cat > %s", filename)},
				Stdin:       block,
				Description: fmt.Sprintf("Create %s", filename),
			})
		}

		// Initialize Go module if needed
		if !strings.Contains(output, "go.mod") {
			commands = append(commands, SandboxCommand{
				Command:     []string{"go", "mod", "init", "sandbox"},
				Description: "Initialize Go module",
			})
		}

		// Try to build and run
		commands = append(commands, SandboxCommand{
			Command:     []string{"go", "mod", "tidy"},
			Description: "Download dependencies",
		})

		commands = append(commands, SandboxCommand{
			Command:     []string{"go", "build", "-o", "output", "."},
			Description: "Build Go program",
		})

		// Run tests if test files exist
		if strings.Contains(output, "_test.go") {
			commands = append(commands, SandboxCommand{
				Command:     []string{"go", "test", "-v", "./..."},
				Description: "Run tests",
			})
		}
	}

	return commands
}

func (se *SandboxedExecutor) parseInfraCommands(output string) []SandboxCommand {
	var commands []SandboxCommand

	// Extract Terraform/infrastructure files
	if tfBlocks := extractCodeBlocks(output, "hcl"); len(tfBlocks) > 0 {
		for i, block := range tfBlocks {
			filename := fmt.Sprintf("main_%d.tf", i)
			commands = append(commands, SandboxCommand{
				Command:     []string{"sh", "-c", fmt.Sprintf("cat > %s", filename)},
				Stdin:       block,
				Description: fmt.Sprintf("Create %s", filename),
			})
		}

		commands = append(commands, SandboxCommand{
			Command:     []string{"terraform", "init"},
			Description: "Initialize Terraform",
		})

		commands = append(commands, SandboxCommand{
			Command:     []string{"terraform", "validate"},
			Description: "Validate Terraform configuration",
		})

		commands = append(commands, SandboxCommand{
			Command:     []string{"terraform", "plan"},
			Description: "Generate Terraform plan",
		})
	}

	return commands
}

func (se *SandboxedExecutor) parseDocCommands(output string) []SandboxCommand {
	var commands []SandboxCommand

	// Extract markdown and convert to different formats
	if mdBlocks := extractCodeBlocks(output, "markdown"); len(mdBlocks) > 0 {
		for i, block := range mdBlocks {
			filename := fmt.Sprintf("doc_%d.md", i)
			commands = append(commands, SandboxCommand{
				Command:     []string{"sh", "-c", fmt.Sprintf("cat > %s", filename)},
				Stdin:       block,
				Description: fmt.Sprintf("Create %s", filename),
			})
		}

		commands = append(commands, SandboxCommand{
			Command:     []string{"pandoc", "*.md", "-o", "output.pdf"},
			Description: "Convert to PDF",
		})
	}

	return commands
}

func (se *SandboxedExecutor) parseAnalysisCommands(output string) []SandboxCommand {
	var commands []SandboxCommand

	// Analysis tasks typically don't need execution, just validation
	commands = append(commands, SandboxCommand{
		Command:     []string{"echo", "Analysis completed"},
		Description: "Validate analysis output",
	})

	return commands
}

func (se *SandboxedExecutor) detectSuspiciousActivity(result *ExecutionResult) bool {
	suspicious := []string{
		"curl", "wget", "nc", "netcat", "ssh", "scp", "rsync",
		"rm -rf", "chmod 777", "su ", "sudo ", "passwd",
		"/etc/", "/var/", "/usr/", "/root/", "/home/",
		"base64", "eval", "exec", "system",
	}

	output := result.Stdout + result.Stderr
	for _, pattern := range suspicious {
		if strings.Contains(strings.ToLower(output), pattern) {
			log.Printf("Detected suspicious activity: %s", pattern)
			return true
		}
	}

	return false
}

func (se *SandboxedExecutor) aggregateOutput(results []CommandResult) string {
	var output strings.Builder
	
	for i, result := range results {
		output.WriteString(fmt.Sprintf("=== Command %d: %s ===\n", i+1, result.Command))
		output.WriteString(fmt.Sprintf("Exit Code: %d\n", result.ExitCode))
		output.WriteString(fmt.Sprintf("Duration: %v\n", result.Duration))
		
		if result.Stdout != "" {
			output.WriteString("STDOUT:\n")
			output.WriteString(result.Stdout)
			output.WriteString("\n")
		}
		
		if result.Stderr != "" {
			output.WriteString("STDERR:\n")
			output.WriteString(result.Stderr)
			output.WriteString("\n")
		}
		
		output.WriteString("\n")
	}
	
	return output.String()
}

func extractCodeBlocks(text, language string) []string {
	var blocks []string
	lines := strings.Split(text, "\n")
	
	var currentBlock strings.Builder
	inBlock := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```"+language) || strings.HasPrefix(line, "```") {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}
	
	return blocks
}

type SandboxCommand struct {
	Command     []string
	Stdin       string
	Description string
}

type CommandResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Metrics  *ResourceMetrics
}

type SandboxExecutionResult struct {
	TaskID        string
	Success       bool
	Output        string
	ExecutionTime time.Duration
	SecurityScore int
	Message       string
	Results       []CommandResult
}