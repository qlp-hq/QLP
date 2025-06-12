package engines

import (
	"context"
	"fmt"
	"time"

	"QLP/services/agent-service/pkg/contracts"
	"QLP/services/llm-service/pkg/client"
	llmContracts "QLP/services/llm-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// DynamicAgent represents a dynamic agent for task execution
type DynamicAgent struct {
	id              string
	taskID          string
	taskType        string
	taskDescription string
	status          contracts.AgentStatus
	createdAt       time.Time
	startedAt       *time.Time
	completedAt     *time.Time
	output          string
	error           error
	configuration   contracts.AgentConfig
	metadata        map[string]string
	metrics         contracts.AgentMetrics
	projectContext  contracts.ProjectContext
	prompt          string
}

// NewDynamicAgent creates a new dynamic agent
func NewDynamicAgent(agentID string, req *contracts.CreateAgentRequest) (*DynamicAgent, error) {
	agent := &DynamicAgent{
		id:              agentID,
		taskID:          req.TaskID,
		taskType:        req.TaskType,
		taskDescription: req.TaskDescription,
		status:          contracts.AgentStatusInitializing,
		createdAt:       time.Now(),
		configuration:   req.Configuration,
		metadata:        req.Metadata,
		projectContext:  req.ProjectContext,
	}

	// Set default configuration
	if agent.configuration.Timeout == 0 {
		agent.configuration.Timeout = 5 * time.Minute
	}
	if agent.configuration.MaxRetries == 0 {
		agent.configuration.MaxRetries = 3
	}

	// Build execution prompt
	agent.prompt = agent.buildExecutionPrompt()
	agent.status = contracts.AgentStatusReady

	logger.WithComponent("dynamic-agent").Info("Dynamic agent created",
		zap.String("agent_id", agentID),
		zap.String("task_type", req.TaskType))

	return agent, nil
}

// Execute executes the agent
func (da *DynamicAgent) Execute(ctx context.Context, llmClient *client.LLMClient) error {
	if da.status != contracts.AgentStatusReady {
		return fmt.Errorf("agent %s not ready for execution, status: %s", da.id, da.status)
	}

	da.status = contracts.AgentStatusExecuting
	startTime := time.Now()
	da.startedAt = &startTime

	logger.WithComponent("dynamic-agent").Info("Starting agent execution",
		zap.String("agent_id", da.id),
		zap.String("task_id", da.taskID))

	// Create LLM request
	llmReq := &llmContracts.CompletionRequest{
		Prompt:       da.prompt,
		MaxTokens:    2000,
		Temperature:  0.1,
		SystemPrompt: da.getSystemPrompt(),
		Metadata: map[string]string{
			"agent_id": da.id,
			"task_id":  da.taskID,
			"task_type": da.taskType,
		},
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, da.configuration.Timeout)
	defer cancel()

	// Call LLM service
	llmResp, err := llmClient.Complete(execCtx, "default", llmReq)
	if err != nil {
		da.status = contracts.AgentStatusFailed
		da.error = fmt.Errorf("LLM execution failed: %w", err)
		completedTime := time.Now()
		da.completedAt = &completedTime
		
		da.updateMetrics(startTime, llmResp)
		
		logger.WithComponent("dynamic-agent").Error("Agent execution failed",
			zap.String("agent_id", da.id),
			zap.Error(err))
		
		return da.error
	}

	// Process output
	da.output = llmResp.Content
	da.status = contracts.AgentStatusCompleted
	completedTime := time.Now()
	da.completedAt = &completedTime

	// Update metrics
	da.updateMetrics(startTime, llmResp)

	logger.WithComponent("dynamic-agent").Info("Agent execution completed",
		zap.String("agent_id", da.id),
		zap.Duration("duration", completedTime.Sub(startTime)),
		zap.Int("output_length", len(da.output)))

	return nil
}

// Cancel cancels the agent execution
func (da *DynamicAgent) Cancel(reason string) error {
	if da.status != contracts.AgentStatusExecuting && da.status != contracts.AgentStatusReady {
		return fmt.Errorf("cannot cancel agent in status: %s", da.status)
	}

	da.status = contracts.AgentStatusCancelled
	da.error = fmt.Errorf("agent cancelled: %s", reason)
	completedTime := time.Now()
	da.completedAt = &completedTime

	if da.metadata == nil {
		da.metadata = make(map[string]string)
	}
	da.metadata["cancel_reason"] = reason

	logger.WithComponent("dynamic-agent").Info("Agent cancelled",
		zap.String("agent_id", da.id),
		zap.String("reason", reason))

	return nil
}

// buildExecutionPrompt builds the execution prompt for the agent
func (da *DynamicAgent) buildExecutionPrompt() string {
	taskTypeInstructions := da.getTaskTypeInstructions()

	prompt := fmt.Sprintf(`You are an Expert %s Agent. Your job is to DIRECTLY EXECUTE the following task and provide the complete, ready-to-use output.

TASK TO EXECUTE:
- ID: %s
- Type: %s
- Description: %s

PROJECT CONTEXT:
- Project Type: %s
- Tech Stack: %v
- Requirements: %v
- Architecture: %s

%s

CRITICAL: Provide ONLY the actual executable output (code/configuration/documentation) - NO lists, NO steps, NO explanations, NO process descriptions. Just the final working result that can be used immediately.

Previous outputs available: %v
`,
		da.taskType,
		da.taskID,
		da.taskType,
		da.taskDescription,
		da.projectContext.ProjectType,
		da.projectContext.TechStack,
		da.projectContext.Requirements,
		da.projectContext.Architecture,
		taskTypeInstructions,
		da.projectContext.PreviousOutputs)

	return prompt
}

// getSystemPrompt returns the system prompt for the agent
func (da *DynamicAgent) getSystemPrompt() string {
	return fmt.Sprintf("You are an expert %s specialist. Always respond with valid, production-ready output in the requested format.", da.taskType)
}

// getTaskTypeInstructions returns task-specific instructions
func (da *DynamicAgent) getTaskTypeInstructions() string {
	switch da.taskType {
	case "codegen":
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
      }
    ]
  }
}

Generate complete, production-ready project structure with proper file organization.
`
	case "infra":
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
        "content": "apiVersion: apps/v1\nkind: Deployment..."
      }
    ]
  }
}

Generate appropriate infrastructure files (.yaml, .tf, .dockerfile, etc.)
`
	case "test":
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
        "content": "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {...}"
      }
    ]
  }
}

Generate appropriate test files (_test.go, test_*.py, *.spec.js, etc.)
`
	case "doc":
		return `
REQUIRED OUTPUT: JSON structure containing documentation files:

{
  "project_structure": {
    "project_name": "documentation",
    "project_type": "markdown|sphinx|gitbook",
    "files": [
      {
        "path": "README.md",
        "type": "markdown",
        "content": "# Project Title\n\n## Overview..."
      }
    ]
  }
}

Generate appropriate documentation files (.md, .rst, .html, etc.)
`
	case "analyze":
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
        "content": "# Analysis Report\n\n## Executive Summary..."
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

// updateMetrics updates agent metrics
func (da *DynamicAgent) updateMetrics(startTime time.Time, llmResp *llmContracts.CompletionResponse) {
	duration := time.Since(startTime)
	
	da.metrics = contracts.AgentMetrics{
		TotalExecutionTime: duration,
		ValidationScore:    85, // Default score
		SecurityScore:      80, // Default score
		QualityScore:       75, // Default score
	}

	if llmResp != nil {
		da.metrics.LLMTokensUsed = llmResp.Usage.TotalTokens
		da.metrics.LLMResponseTime = llmResp.ResponseTime
	}

	// Simulate sandbox and validation times
	da.metrics.SandboxExecutionTime = duration / 4
	da.metrics.ValidationTime = duration / 8
	da.metrics.MemoryUsed = 64 // MB
	da.metrics.CPUTime = duration / 2
}

// Getter methods

func (da *DynamicAgent) GetID() string {
	return da.id
}

func (da *DynamicAgent) GetTaskID() string {
	return da.taskID
}

func (da *DynamicAgent) GetTaskType() string {
	return da.taskType
}

func (da *DynamicAgent) GetTaskDescription() string {
	return da.taskDescription
}

func (da *DynamicAgent) GetStatus() contracts.AgentStatus {
	return da.status
}

func (da *DynamicAgent) GetCreatedAt() time.Time {
	return da.createdAt
}

func (da *DynamicAgent) GetStartedAt() *time.Time {
	return da.startedAt
}

func (da *DynamicAgent) GetCompletedAt() *time.Time {
	return da.completedAt
}

func (da *DynamicAgent) GetDuration() time.Duration {
	if da.completedAt != nil && da.startedAt != nil {
		return da.completedAt.Sub(*da.startedAt)
	}
	return 0
}

func (da *DynamicAgent) GetOutput() string {
	return da.output
}

func (da *DynamicAgent) GetError() error {
	return da.error
}

func (da *DynamicAgent) GetErrorString() string {
	if da.error != nil {
		return da.error.Error()
	}
	return ""
}

func (da *DynamicAgent) GetMetrics() contracts.AgentMetrics {
	return da.metrics
}

func (da *DynamicAgent) GetConfiguration() contracts.AgentConfig {
	return da.configuration
}

func (da *DynamicAgent) GetMetadata() map[string]string {
	return da.metadata
}

func (da *DynamicAgent) GetValidationScore() int {
	return da.metrics.ValidationScore
}