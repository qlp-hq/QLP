package agents

import (
	"context"
	"fmt"
	"log"
	"time"

	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/validation"
)

type DynamicAgent struct {
	ID                string
	Task              models.Task
	LLMClient         llm.Client
	EventBus          *events.EventBus
	MetaPromptGen     *MetaPromptGenerator
	Context           AgentContext
	SandboxExecutor   *sandbox.SandboxedExecutor
	ValidationEngine  *validation.ValidationEngine
	GeneratedPrompt   string
	Status            AgentStatus
	StartTime         time.Time
	Output            string
	SandboxResult     *sandbox.SandboxExecutionResult
	ValidationResult  *validation.ValidationResult
	Error             error
}

type AgentStatus string

const (
	AgentStatusInitializing AgentStatus = "initializing"
	AgentStatusReady        AgentStatus = "ready"
	AgentStatusExecuting    AgentStatus = "executing"
	AgentStatusCompleted    AgentStatus = "completed"
	AgentStatusFailed       AgentStatus = "failed"
)

func NewDynamicAgent(task models.Task, llmClient llm.Client, eventBus *events.EventBus, agentContext AgentContext) *DynamicAgent {
	metaPromptGen := NewMetaPromptGenerator(llmClient)
	sandboxExecutor := sandbox.NewSandboxedExecutor()
	validationEngine := validation.NewValidationEngine(llmClient)

	return &DynamicAgent{
		ID:               generateProfessionalAgentID(task),
		Task:             task,
		LLMClient:        llmClient,
		EventBus:         eventBus,
		MetaPromptGen:    metaPromptGen,
		Context:          agentContext,
		SandboxExecutor:  sandboxExecutor,
		ValidationEngine: validationEngine,
		Status:           AgentStatusInitializing,
	}
}

func (da *DynamicAgent) Initialize(ctx context.Context) error {
	log.Printf("Initializing dynamic agent %s for task %s", da.ID, da.Task.ID)

	// Skip meta-prompt generation and use direct execution prompt
	da.GeneratedPrompt = da.buildDirectExecutionPrompt()
	da.Status = AgentStatusReady

	da.EventBus.Publish(events.Event{
		ID:        fmt.Sprintf("agent_%s_initialized", da.ID),
		Type:      events.EventAgentSpawned,
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload: map[string]interface{}{
			"agent_id":  da.ID,
			"task_id":   da.Task.ID,
			"task_type": da.Task.Type,
			"status":    da.Status,
		},
	})

	log.Printf("Agent %s initialized with specialized prompt", da.ID)
	return nil
}

func (da *DynamicAgent) Execute(ctx context.Context) error {
	if da.Status != AgentStatusReady {
		return fmt.Errorf("agent %s not ready for execution, status: %s", da.ID, da.Status)
	}

	da.Status = AgentStatusExecuting
	da.StartTime = time.Now()

	log.Printf("Agent %s executing task %s", da.ID, da.Task.ID)

	da.EventBus.Publish(events.Event{
		ID:        fmt.Sprintf("agent_%s_started", da.ID),
		Type:      events.EventTaskStarted,
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload: map[string]interface{}{
			"agent_id":    da.ID,
			"task_id":     da.Task.ID,
			"description": da.Task.Description,
		},
	})

	executionPrompt := da.buildExecutionPrompt()

	llmOutput, err := da.LLMClient.Complete(ctx, executionPrompt)
	if err != nil {
		da.Status = AgentStatusFailed
		da.Error = err

		da.EventBus.Publish(events.Event{
			ID:        fmt.Sprintf("agent_%s_failed", da.ID),
			Type:      events.EventTaskFailed,
			Timestamp: time.Now(),
			Source:    da.ID,
			Payload: map[string]interface{}{
				"agent_id": da.ID,
				"task_id":  da.Task.ID,
				"error":    err.Error(),
			},
		})

		return fmt.Errorf("agent execution failed: %w", err)
	}

	log.Printf("Agent %s received LLM output, executing in sandbox", da.ID)

	sandboxResult, err := da.SandboxExecutor.Execute(ctx, da.Task, llmOutput)
	if err != nil {
		da.Status = AgentStatusFailed
		da.Error = err
		da.Output = llmOutput // Store LLM output even if sandbox fails

		da.EventBus.Publish(events.Event{
			ID:        fmt.Sprintf("agent_%s_sandbox_failed", da.ID),
			Type:      events.EventTaskFailed,
			Timestamp: time.Now(),
			Source:    da.ID,
			Payload: map[string]interface{}{
				"agent_id": da.ID,
				"task_id":  da.Task.ID,
				"error":    err.Error(),
				"phase":    "sandbox_execution",
			},
		})

		return fmt.Errorf("sandbox execution failed: %w", err)
	}

	da.SandboxResult = sandboxResult

	log.Printf("Agent %s sandbox execution completed, starting validation", da.ID)

	// Validate the output
	validationResult, err := da.ValidationEngine.ValidateTaskOutput(ctx, da.Task, llmOutput, sandboxResult)
	if err != nil {
		log.Printf("Agent %s validation failed: %v", da.ID, err)
		// Continue execution even if validation fails
		validationResult = &validation.ValidationResult{
			TaskID:       da.Task.ID,
			OverallScore: 50,
			Passed:       false,
			Timestamp:    time.Now(),
		}
	}

	da.ValidationResult = validationResult

	// Build comprehensive output
	da.Output = fmt.Sprintf(`=== LLM OUTPUT ===
%s

=== SANDBOX EXECUTION ===
%s

=== VALIDATION RESULTS ===
Overall Score: %d/100 (%s)
Syntax Score: %d/100
Security Score: %d/100 (Risk: %s)
Quality Score: %d/100
LLM Critique Score: %d/100
Validation Time: %v
`, 
		llmOutput, 
		sandboxResult.Output,
		validationResult.OverallScore,
		map[bool]string{true: "PASSED", false: "FAILED"}[validationResult.Passed],
		getScoreOrDefault(validationResult.SyntaxResult),
		getScoreOrDefault(validationResult.SecurityResult),
		getStringOrDefault(validationResult.SecurityResult),
		getScoreOrDefault(validationResult.QualityResult),
		getScoreOrDefault(validationResult.LLMCritiqueResult),
		validationResult.ValidationTime,
	)

	da.Status = AgentStatusCompleted

	da.EventBus.Publish(events.Event{
		ID:        fmt.Sprintf("agent_%s_completed", da.ID),
		Type:      events.EventTaskCompleted,
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload: map[string]interface{}{
			"agent_id":         da.ID,
			"task_id":          da.Task.ID,
			"output_size":      len(da.Output),
			"duration_ms":      time.Since(da.StartTime).Milliseconds(),
			"sandbox_success":  sandboxResult.Success,
			"security_score":   sandboxResult.SecurityScore,
			"execution_time":   sandboxResult.ExecutionTime.Milliseconds(),
			"validation_score": validationResult.OverallScore,
			"validation_passed": validationResult.Passed,
		},
	})

	log.Printf("Agent %s completed task %s in %v", da.ID, da.Task.ID, time.Since(da.StartTime))
	return nil
}

func (da *DynamicAgent) buildDirectExecutionPrompt() string {
	taskTypeInstructions := da.getTaskTypeExecutionInstructions()
	
	return fmt.Sprintf(`You are an Expert %s Agent. Your job is to DIRECTLY EXECUTE the following task and provide the complete, ready-to-use output.

TASK TO EXECUTE:
- ID: %s
- Type: %s
- Description: %s
- Priority: %s

PROJECT CONTEXT:
- Project Type: %s
- Tech Stack: %v
- Dependencies: %v

%s

CRITICAL: Provide ONLY the actual executable output (code/configuration/documentation) - NO lists, NO steps, NO explanations, NO process descriptions. Just the final working result that can be used immediately.
`,
		da.Task.Type,
		da.Task.ID,
		da.Task.Type,
		da.Task.Description,
		da.Task.Priority,
		da.Context.ProjectType,
		da.Context.TechStack,
		da.Task.Dependencies,
		taskTypeInstructions,
	)
}

func (da *DynamicAgent) buildExecutionPrompt() string {
	// Use the direct execution prompt - no double-wrapping
	return da.GeneratedPrompt
}

func (da *DynamicAgent) getTaskTypeExecutionInstructions() string {
	switch da.Task.Type {
	case models.TaskTypeCodegen:
		return `
REQUIRED OUTPUT: JSON structure containing file information and code:

{
  "project_structure": {
    "project_name": "descriptive-project-name",
    "project_type": "go-api|python-script|node-app|etc",
    "files": [
      {
        "path": "main.go",
        "type": "go",
        "content": "package main\n\nimport (\n    \"fmt\"\n)\n\nfunc main() {\n    fmt.Println(\"Hello World\")\n}"
      },
      {
        "path": "go.mod",
        "type": "mod",
        "content": "module project-name\n\ngo 1.21"
      },
      {
        "path": "README.md",
        "type": "markdown",
        "content": "# Project Name\n\nDescription and usage instructions"
      }
    ]
  }
}

The LLM should determine appropriate:
- File extensions (.go, .py, .js, .yaml, .dockerfile, etc.)
- Project structure (src/, cmd/, pkg/, tests/, docs/)
- Configuration files (go.mod, package.json, requirements.txt, etc.)
- Documentation files (README.md, API docs)
- Build/deployment files (Dockerfile, Makefile, etc.)

Generate complete, production-ready project structure with proper file organization.
`
	case models.TaskTypeInfra:
		return `
REQUIRED OUTPUT: JSON structure containing infrastructure files:

{
  "project_structure": {
    "project_name": "infrastructure-project",
    "project_type": "kubernetes|terraform|docker-compose|helm",
    "files": [
      {
        "path": "deployment.yaml",
        "type": "yaml",
        "content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: app\nspec:\n  replicas: 3"
      },
      {
        "path": "service.yaml",
        "type": "yaml", 
        "content": "apiVersion: v1\nkind: Service\nmetadata:\n  name: app-service"
      },
      {
        "path": "Dockerfile",
        "type": "dockerfile",
        "content": "FROM alpine:latest\nWORKDIR /app\nCOPY . .\nEXPOSE 8080"
      }
    ]
  }
}

Generate appropriate infrastructure files (.yaml, .tf, .dockerfile, docker-compose.yml, etc.)
`
	case models.TaskTypeTest:
		return `
REQUIRED OUTPUT: JSON structure containing test files:

{
  "project_structure": {
    "project_name": "test-suite",
    "project_type": "go-test|python-pytest|jest|etc",
    "files": [
      {
        "path": "main_test.go",
        "type": "go",
        "content": "package main\n\nimport (\n\t\"testing\"\n)\n\nfunc TestMain(t *testing.T) {\n\t// Test implementation\n}"
      },
      {
        "path": "test_data.json",
        "type": "json",
        "content": "{\n\t\"test_cases\": []\n}"
      }
    ]
  }
}

Generate appropriate test files (_test.go, test_*.py, *.spec.js, etc.)
`
	case models.TaskTypeDoc:
		return `
REQUIRED OUTPUT: JSON structure containing documentation files:

{
  "project_structure": {
    "project_name": "documentation",
    "project_type": "markdown|sphinx|gitbook|etc",
    "files": [
      {
        "path": "README.md",
        "type": "markdown",
        "content": "# Project Title\n\n## Overview\nDescription of the project\n\n## Installation\nInstallation instructions"
      },
      {
        "path": "docs/api.md",
        "type": "markdown", 
        "content": "# API Documentation\n\n## Endpoints\n\n### GET /api/users\nReturns list of users"
      },
      {
        "path": "docs/examples.md",
        "type": "markdown",
        "content": "# Examples\n\n## Basic Usage\nCode examples and usage patterns"
      }
    ]
  }
}

Generate appropriate documentation files (.md, .rst, .html, etc.)
`
	case models.TaskTypeAnalyze:
		return `
REQUIRED OUTPUT: JSON structure containing analysis files:

{
  "project_structure": {
    "project_name": "analysis-report",
    "project_type": "analysis|report|research",
    "files": [
      {
        "path": "analysis_report.md",
        "type": "markdown",
        "content": "# Analysis Report\n\n## Executive Summary\nKey findings and recommendations\n\n## Detailed Analysis\nIn-depth analysis with data"
      },
      {
        "path": "data/metrics.json",
        "type": "json",
        "content": "{\n  \"performance_metrics\": {},\n  \"security_findings\": {}\n}"
      },
      {
        "path": "charts/performance.svg",
        "type": "svg",
        "content": "<svg>Performance chart data</svg>"
      }
    ]
  }
}

Generate appropriate analysis files (.md, .json, .csv, .svg, etc.)
`
	default:
		return `
REQUIRED OUTPUT: JSON structure containing project files:

{
  "project_structure": {
    "project_name": "generic-project",
    "project_type": "general",
    "files": [
      {
        "path": "output.txt",
        "type": "text",
        "content": "Complete, production-ready deliverable content"
      }
    ]
  }
}
`
	}
}

func (da *DynamicAgent) formatExecutionContext() string {
	ctxInfo := fmt.Sprintf(`
Project Type: %s
Tech Stack: %v
Dependencies: %v
`,
		da.Context.ProjectType,
		da.Context.TechStack,
		da.Task.Dependencies,
	)

	if len(da.Context.PreviousOutputs) > 0 {
		ctxInfo += "\nPrevious Task Outputs Available:\n"
		for taskID := range da.Context.PreviousOutputs {
			ctxInfo += fmt.Sprintf("- %s\n", taskID)
		}
	}

	return ctxInfo
}

func (da *DynamicAgent) GetOutput() string {
	return da.Output
}

func (da *DynamicAgent) GetStatus() AgentStatus {
	return da.Status
}

func (da *DynamicAgent) GetError() error {
	return da.Error
}

func generateProfessionalAgentID(task models.Task) string {
	typePrefix := map[models.TaskType]string{
		models.TaskTypeInfra:   "QLI",
		models.TaskTypeCodegen: "QLD", 
		models.TaskTypeTest:    "QLT",
		models.TaskTypeDoc:     "QLC",
		models.TaskTypeAnalyze: "QLA",
	}
	
	prefix, exists := typePrefix[task.Type]
	if !exists {
		prefix = "QLG"
	}
	
	timestamp := time.Now().Format("150405")
	sequence := time.Now().UnixNano() % 1000
	return fmt.Sprintf("%s-AGT-%s-%03d", prefix, timestamp, sequence)
}

func getScoreOrDefault(result interface{}) int {
	switch r := result.(type) {
	case *validation.SyntaxValidationResult:
		if r != nil {
			return r.Score
		}
	case *validation.TaskSecurityValidationResult:
		if r != nil {
			return r.Score
		}
	case *validation.QualityValidationResult:
		if r != nil {
			return r.Score
		}
	case *validation.LLMCritiqueResult:
		if r != nil {
			return r.Score
		}
	}
	return 0
}

func getStringOrDefault(result interface{}) string {
	switch r := result.(type) {
	case *validation.TaskSecurityValidationResult:
		if r != nil {
			return string(r.RiskLevel)
		}
	}
	return "unknown"
}
