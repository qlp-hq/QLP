package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// UniversalValidator leverages LLM intelligence for truly universal project validation
type UniversalValidator struct {
	llmClient llm.Client
}

// ProjectAnalysis contains LLM-powered analysis of any project
type ProjectAnalysis struct {
	Language            string            `json:"language"`
	Framework           string            `json:"framework"`
	ProjectType         string            `json:"project_type"`
	BuildTool           string            `json:"build_tool"`
	PackageManager      string            `json:"package_manager"`
	TestFramework       string            `json:"test_framework"`
	EntryPoint          string            `json:"entry_point"`
	BuildCommands       []string          `json:"build_commands"`
	TestCommands        []string          `json:"test_commands"`
	RunCommands         []string          `json:"run_commands"`
	Dependencies        []string          `json:"dependencies"`
	DevDependencies     []string          `json:"dev_dependencies"`
	ConfigFiles         []string          `json:"config_files"`
	SourceDirectories   []string          `json:"source_directories"`
	OutputDirectories   []string          `json:"output_directories"`
	RequiredTools       []string          `json:"required_tools"`
	EnvironmentSetup    []string          `json:"environment_setup"`
	DeploymentStrategy  string            `json:"deployment_strategy"`
	Confidence          float64           `json:"confidence"`
	Recommendations     []string          `json:"recommendations"`
	PotentialIssues     []string          `json:"potential_issues"`
	BestPractices       []string          `json:"best_practices"`
	SecurityConsiderations []string       `json:"security_considerations"`
}

// UniversalBuildResult contains results from LLM-guided universal build
type UniversalBuildResult struct {
	Success         bool                 `json:"success"`
	Language        string               `json:"language"`
	Framework       string               `json:"framework"`
	BuildTool       string               `json:"build_tool"`
	ExecutedCommands []ExecutedCommand   `json:"executed_commands"`
	OutputArtifacts []string            `json:"output_artifacts"`
	BuildTime       time.Duration       `json:"build_time"`
	Issues          []string            `json:"issues"`
	Warnings        []string            `json:"warnings"`
	Recommendations []string            `json:"recommendations"`
	NextSteps       []string            `json:"next_steps"`
}

// ExecutedCommand tracks command execution details
type ExecutedCommand struct {
	Command     string        `json:"command"`
	Directory   string        `json:"directory"`
	ExitCode    int           `json:"exit_code"`
	Output      string        `json:"output"`
	Error       string        `json:"error"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
}

// NewUniversalValidator creates a new LLM-powered universal validator
func NewUniversalValidator(llmClient llm.Client) *UniversalValidator {
	return &UniversalValidator{
		llmClient: llmClient,
	}
}

// AnalyzeProject uses LLM to analyze any programming language project
func (uv *UniversalValidator) AnalyzeProject(ctx context.Context, projectPath string, files map[string]string) (*ProjectAnalysis, error) {
	logger.WithComponent("validation").Info("Analyzing project with LLM intelligence",
		zap.String("project_path", projectPath),
		zap.Int("file_count", len(files)))

	// Build comprehensive project context for LLM
	projectContext := uv.buildProjectContext(files)

	prompt := fmt.Sprintf(`You are a senior software architect and DevOps expert with deep knowledge of ALL programming languages, frameworks, and build systems. Analyze this project and provide comprehensive build and deployment guidance.

PROJECT ANALYSIS REQUEST:
Analyze the following project files and provide detailed information about how to build, test, and deploy this project.

PROJECT FILES:
%s

ANALYSIS REQUIREMENTS:

1. LANGUAGE & FRAMEWORK DETECTION:
   - Primary programming language
   - Framework/runtime (if any)
   - Project type (web app, CLI tool, library, microservice, etc.)

2. BUILD SYSTEM ANALYSIS:
   - Build tool (Maven, Gradle, npm, cargo, mix, stack, sbt, etc.)
   - Package manager (pip, npm, yarn, cargo, gem, composer, etc.)
   - Dependency management approach

3. BUILD COMMANDS:
   - Exact commands to install dependencies
   - Exact commands to build the project
   - Commands in correct execution order
   - Any environment setup required

4. TESTING STRATEGY:
   - Test framework used (if detectable)
   - Commands to run tests
   - Test file patterns

5. EXECUTION & DEPLOYMENT:
   - How to run the application
   - Default ports/endpoints (if applicable)
   - Deployment strategy recommendations
   - Required runtime environment

6. PROJECT HEALTH:
   - Code quality assessment
   - Potential issues or missing files
   - Security considerations
   - Best practices compliance

7. TOOL REQUIREMENTS:
   - Required tools/runtimes that must be installed
   - Version requirements (if specified)
   - Environment variables needed

IMPORTANT GUIDELINES:
- Support ALL programming languages (Rust, Go, Python, Java, C#, Ruby, PHP, Swift, Kotlin, Dart, Elixir, Haskell, Scala, Clojure, F#, etc.)
- Provide specific, executable commands
- Consider cross-platform compatibility
- Include confidence score (0.0-1.0)
- Suggest improvements and best practices
- Identify potential security issues

RESPOND WITH JSON:
{
  "language": "rust",
  "framework": "actix-web",
  "project_type": "web_api",
  "build_tool": "cargo",
  "package_manager": "cargo",
  "test_framework": "built-in",
  "entry_point": "src/main.rs",
  "build_commands": [
    "cargo fetch",
    "cargo build --release"
  ],
  "test_commands": [
    "cargo test"
  ],
  "run_commands": [
    "cargo run",
    "./target/release/app-name"
  ],
  "dependencies": ["actix-web", "serde", "tokio"],
  "dev_dependencies": ["cargo-watch"],
  "config_files": ["Cargo.toml", "Cargo.lock"],
  "source_directories": ["src/", "tests/"],
  "output_directories": ["target/"],
  "required_tools": ["rust", "cargo"],
  "environment_setup": ["curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"],
  "deployment_strategy": "binary_deployment",
  "confidence": 0.95,
  "recommendations": [
    "Add Dockerfile for containerized deployment",
    "Consider adding CI/CD configuration"
  ],
  "potential_issues": [
    "No environment configuration file detected"
  ],
  "best_practices": [
    "Good use of Cargo.toml for dependency management",
    "Proper project structure follows Rust conventions"
  ],
  "security_considerations": [
    "Review dependencies for known vulnerabilities",
    "Add input validation for web endpoints"
  ]
}`, projectContext)

	response, err := uv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed for project %s: %w", projectPath, err)
	}

	var analysis ProjectAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		logger.WithComponent("validation").Warn("Failed to parse LLM project analysis response",
			zap.Error(err),
			zap.String("response", response))
		
		// Fallback to basic analysis
		return uv.fallbackProjectAnalysis(files), nil
	}

	logger.WithComponent("validation").Info("Project analysis completed",
		zap.String("language", analysis.Language),
		zap.String("framework", analysis.Framework),
		zap.String("build_tool", analysis.BuildTool),
		zap.Float64("confidence", analysis.Confidence))

	return &analysis, nil
}

// BuildProject uses LLM-guided commands to build any project
func (uv *UniversalValidator) BuildProject(ctx context.Context, projectPath string, analysis *ProjectAnalysis) (*UniversalBuildResult, error) {
	startTime := time.Now()
	
	logger.WithComponent("validation").Info("Starting LLM-guided universal build",
		zap.String("project_path", projectPath),
		zap.String("language", analysis.Language),
		zap.String("build_tool", analysis.BuildTool))

	result := &UniversalBuildResult{
		Language:         analysis.Language,
		Framework:        analysis.Framework,
		BuildTool:        analysis.BuildTool,
		ExecutedCommands: make([]ExecutedCommand, 0),
		OutputArtifacts:  make([]string, 0),
		Issues:           make([]string, 0),
		Warnings:         make([]string, 0),
		Recommendations:  make([]string, 0),
		NextSteps:        make([]string, 0),
	}

	// Execute environment setup commands first
	for _, setupCmd := range analysis.EnvironmentSetup {
		if setupCmd != "" {
			executed := uv.executeCommand(projectPath, setupCmd)
			result.ExecutedCommands = append(result.ExecutedCommands, executed)
			if !executed.Success {
				logger.WithComponent("validation").Warn("Environment setup command failed",
					zap.String("command", setupCmd),
					zap.String("error", executed.Error))
				result.Warnings = append(result.Warnings, fmt.Sprintf("Environment setup warning: %s", executed.Error))
			}
		}
	}

	// Execute build commands in sequence
	for _, buildCmd := range analysis.BuildCommands {
		if buildCmd != "" {
			executed := uv.executeCommand(projectPath, buildCmd)
			result.ExecutedCommands = append(result.ExecutedCommands, executed)
			
			if !executed.Success {
				logger.WithComponent("validation").Error("Build command failed",
					zap.String("command", buildCmd),
					zap.String("error", executed.Error))
				result.Issues = append(result.Issues, fmt.Sprintf("Build failed: %s", executed.Error))
				result.Success = false
				result.BuildTime = time.Since(startTime)
				return result, fmt.Errorf("build command failed for %s project: %s", analysis.Language, executed.Error)
			}
		}
	}

	// Check for output artifacts
	for _, outputDir := range analysis.OutputDirectories {
		artifacts := uv.findArtifacts(filepath.Join(projectPath, outputDir))
		result.OutputArtifacts = append(result.OutputArtifacts, artifacts...)
	}

	result.Success = true
	result.BuildTime = time.Since(startTime)
	
	// Generate next steps based on successful build
	result.NextSteps = append(result.NextSteps, "Run tests to validate functionality")
	result.NextSteps = append(result.NextSteps, "Perform security scan on built artifacts")
	result.NextSteps = append(result.NextSteps, "Prepare deployment configuration")

	logger.WithComponent("validation").Info("Universal build completed successfully",
		zap.String("language", analysis.Language),
		zap.Duration("build_time", result.BuildTime),
		zap.Int("artifacts", len(result.OutputArtifacts)))

	return result, nil
}

// TestProject runs LLM-suggested test commands for any project
func (uv *UniversalValidator) TestProject(ctx context.Context, projectPath string, analysis *ProjectAnalysis) (*UniversalBuildResult, error) {
	startTime := time.Now()
	
	logger.WithComponent("validation").Info("Starting LLM-guided universal testing",
		zap.String("project_path", projectPath),
		zap.String("language", analysis.Language),
		zap.String("test_framework", analysis.TestFramework))

	result := &UniversalBuildResult{
		Language:         analysis.Language,
		Framework:        analysis.Framework,
		BuildTool:        analysis.BuildTool,
		ExecutedCommands: make([]ExecutedCommand, 0),
		Issues:           make([]string, 0),
		Warnings:         make([]string, 0),
	}

	// Execute test commands
	for _, testCmd := range analysis.TestCommands {
		if testCmd != "" {
			executed := uv.executeCommand(projectPath, testCmd)
			result.ExecutedCommands = append(result.ExecutedCommands, executed)
			
			if !executed.Success {
				logger.WithComponent("validation").Warn("Test command failed",
					zap.String("command", testCmd),
					zap.String("error", executed.Error))
				result.Issues = append(result.Issues, fmt.Sprintf("Test failed: %s", executed.Error))
			}
		}
	}

	// Determine overall success
	result.Success = len(result.Issues) == 0
	result.BuildTime = time.Since(startTime)

	logger.WithComponent("validation").Info("Universal testing completed",
		zap.String("language", analysis.Language),
		zap.Bool("success", result.Success),
		zap.Duration("test_time", result.BuildTime))

	return result, nil
}

// Helper methods

func (uv *UniversalValidator) buildProjectContext(files map[string]string) string {
	var contextBuilder strings.Builder
	
	contextBuilder.WriteString("PROJECT FILE STRUCTURE AND CONTENTS:\n\n")
	
	for filePath, content := range files {
		contextBuilder.WriteString(fmt.Sprintf("=== FILE: %s ===\n", filePath))
		
		// Limit content size for each file to avoid token limits
		if len(content) > 2000 {
			contextBuilder.WriteString(content[:2000])
			contextBuilder.WriteString("\n... [truncated] ...\n")
		} else {
			contextBuilder.WriteString(content)
		}
		contextBuilder.WriteString("\n\n")
	}
	
	return contextBuilder.String()
}

func (uv *UniversalValidator) executeCommand(projectPath, command string) ExecutedCommand {
	startTime := time.Now()
	
	// Parse command (handle shell commands)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ExecutedCommand{
			Command:   command,
			Directory: projectPath,
			ExitCode:  1,
			Error:     "empty command",
			Duration:  0,
			Success:   false,
		}
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = projectPath
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	executed := ExecutedCommand{
		Command:   command,
		Directory: projectPath,
		Output:    string(output),
		Duration:  duration,
		Success:   err == nil,
	}
	
	if err != nil {
		executed.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			executed.ExitCode = exitError.ExitCode()
		} else {
			executed.ExitCode = 1
		}
	}
	
	return executed
}

func (uv *UniversalValidator) findArtifacts(outputPath string) []string {
	// This would scan for common build artifacts
	// Implementation would use filepath.Walk to find executables, JARs, etc.
	return []string{} // Placeholder
}

func (uv *UniversalValidator) fallbackProjectAnalysis(files map[string]string) *ProjectAnalysis {
	// Basic fallback analysis when LLM fails
	analysis := &ProjectAnalysis{
		Language:     "unknown",
		Framework:    "unknown",
		ProjectType:  "unknown",
		BuildTool:    "unknown",
		Confidence:   0.1,
		Recommendations: []string{"Manual analysis required - AI analysis failed"},
	}
	
	// Simple heuristics
	for filePath := range files {
		switch {
		case strings.HasSuffix(filePath, ".go") && strings.Contains(filePath, "go.mod"):
			analysis.Language = "go"
			analysis.BuildTool = "go"
			analysis.BuildCommands = []string{"go mod download", "go build"}
		case strings.HasSuffix(filePath, ".py") && strings.Contains(filePath, "requirements.txt"):
			analysis.Language = "python"
			analysis.BuildTool = "pip"
			analysis.BuildCommands = []string{"pip install -r requirements.txt"}
		case strings.HasSuffix(filePath, ".js") && strings.Contains(filePath, "package.json"):
			analysis.Language = "javascript"
			analysis.BuildTool = "npm"
			analysis.BuildCommands = []string{"npm install", "npm run build"}
		}
	}
	
	return analysis
}