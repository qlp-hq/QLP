package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
	"QLP/services/common/validation"
	promptclient "QLP/services/prompt-service/pkg/client"

	"go.uber.org/zap"
)

type DynamicAgent struct {
	ID               string
	Task             models.Task
	LLMClient        llm.Client
	EventManager     events.Manager
	PromptClient     *promptclient.PromptServiceClient
	Context          AgentContext
	SandboxExecutor  *sandbox.SandboxedExecutor
	ValidationEngine *validation.ValidationEngine
	GeneratedPrompt  string
	Status           AgentStatus
	StartTime        time.Time
	Output           string
	SandboxResult    *sandbox.SandboxExecutionResult
	ValidationResult *types.ValidationResult
	Error            error
}

type AgentStatus string

const (
	AgentStatusInitializing AgentStatus = "initializing"
	AgentStatusReady        AgentStatus = "ready"
	AgentStatusExecuting    AgentStatus = "executing"
	AgentStatusCompleted    AgentStatus = "completed"
	AgentStatusFailed       AgentStatus = "failed"
)

func NewDynamicAgent(task models.Task, llmClient llm.Client, eventManager events.Manager, agentContext AgentContext, promptClient *promptclient.PromptServiceClient) *DynamicAgent {
	sandboxExecutor := sandbox.NewSandboxedExecutor()
	validationEngine := validation.NewValidationEngine(llmClient)

	return &DynamicAgent{
		ID:               generateProfessionalAgentID(task),
		Task:             task,
		LLMClient:        llmClient,
		EventManager:     eventManager,
		PromptClient:     promptClient,
		Context:          agentContext,
		SandboxExecutor:  sandboxExecutor,
		ValidationEngine: validationEngine,
		Status:           AgentStatusInitializing,
	}
}

func (da *DynamicAgent) Initialize(ctx context.Context) error {
	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Initializing dynamic agent",
		zap.String("task_id", da.Task.ID),
		zap.String("task_type", string(da.Task.Type)))

	prompt, err := da.PromptClient.GetActivePromptByTaskType(ctx, string(da.Task.Type))
	if err != nil {
		return fmt.Errorf("failed to get prompt from prompt-service: %w", err)
	}

	// Here you would inject the context into the prompt.
	// For now, we'll just use the text directly.
	da.GeneratedPrompt = prompt.PromptText
	da.Status = AgentStatusReady

	payload, _ := json.Marshal(map[string]interface{}{
		"agent_id":  da.ID,
		"task_id":   da.Task.ID,
		"task_type": string(da.Task.Type),
		"status":    string(da.Status),
	})
	da.EventManager.Publish(ctx, events.Event{
		ID:        fmt.Sprintf("agent_%s_initialized", da.ID),
		Type:      "agent.spawned",
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload:   payload,
	})

	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Agent initialized with specialized prompt")
	return nil
}

// Execute is the entrypoint that satisfies the Agent interface.
// It orchestrates the initialization and execution of the dynamic agent's lifecycle.
func (da *DynamicAgent) Execute(ctx context.Context, task models.Task) (*models.Artifact, error) {
	da.Task = task // Assign the task
	if err := da.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("agent initialization failed: %w", err)
	}

	// The dynamic agent's own Execute method runs the main logic
	if err := da.executeInternal(ctx); err != nil {
		// Even if execution fails, we might have a partial artifact or output
		// that we want to save. For now, we'll just return the error.
		return nil, err
	}

	artifact := &models.Artifact{
		ID:      fmt.Sprintf("art-%s", da.Task.ID),
		Task:    da.Task,
		Type:    models.ArtifactType(da.Task.Type),
		Content: da.GetOutput(),
		Metadata: map[string]string{
			"agent_id":          da.ID,
			"validation_score":  fmt.Sprintf("%d", da.ValidationResult.OverallScore),
			"validation_passed": fmt.Sprintf("%t", da.ValidationResult.Passed),
		},
		CreatedAt: time.Now(),
	}

	return artifact, nil
}

// executeInternal contains the primary logic for the agent's operation.
func (da *DynamicAgent) executeInternal(ctx context.Context) error {
	if da.Status != AgentStatusReady {
		return fmt.Errorf("agent %s not ready for execution, status: %s", da.ID, da.Status)
	}

	da.Status = AgentStatusExecuting
	da.StartTime = time.Now()

	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Agent executing task",
		zap.String("task_id", da.Task.ID),
		zap.String("task_description", da.Task.Description))

	payload, _ := json.Marshal(map[string]interface{}{
		"agent_id":    da.ID,
		"task_id":     da.Task.ID,
		"description": da.Task.Description,
	})
	da.EventManager.Publish(ctx, events.Event{
		ID:        fmt.Sprintf("agent_%s_started", da.ID),
		Type:      "task.started",
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload:   payload,
	})

	llmOutput, err := da.LLMClient.Complete(ctx, da.GeneratedPrompt)
	if err != nil {
		da.Status = AgentStatusFailed
		da.Error = err

		payload, _ = json.Marshal(map[string]interface{}{
			"agent_id": da.ID,
			"task_id":  da.Task.ID,
			"error":    err.Error(),
		})
		da.EventManager.Publish(ctx, events.Event{
			ID:        fmt.Sprintf("agent_%s_failed", da.ID),
			Type:      "task.failed",
			Timestamp: time.Now(),
			Source:    da.ID,
			Payload:   payload,
		})

		return fmt.Errorf("agent execution failed: %w", err)
	}

	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Agent received LLM output, executing in sandbox",
		zap.Int("llm_output_length", len(llmOutput)))

	sandboxResult, err := da.SandboxExecutor.Execute(ctx, da.Task, llmOutput)
	if err != nil {
		da.Status = AgentStatusFailed
		da.Error = err
		da.Output = llmOutput // Store LLM output even if sandbox fails

		payload, _ = json.Marshal(map[string]interface{}{
			"agent_id": da.ID,
			"task_id":  da.Task.ID,
			"error":    err.Error(),
			"phase":    "sandbox_execution",
		})
		da.EventManager.Publish(ctx, events.Event{
			ID:        fmt.Sprintf("agent_%s_sandbox_failed", da.ID),
			Type:      "task.failed",
			Timestamp: time.Now(),
			Source:    da.ID,
			Payload:   payload,
		})

		return fmt.Errorf("sandbox execution failed: %w", err)
	}

	da.SandboxResult = sandboxResult

	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Sandbox execution completed, starting validation",
		zap.Bool("sandbox_success", sandboxResult.Success),
		zap.Duration("execution_time", sandboxResult.ExecutionTime))

	// Validate the output
	validationResult, err := da.ValidationEngine.ValidateTaskOutput(ctx, da.Task, llmOutput, sandboxResult)
	if err != nil {
		logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Warn("Validation failed",
			zap.Error(err))
		// Continue execution even if validation fails
		validationResult = &types.ValidationResult{
			OverallScore: 50,
			Passed:       false,
			ValidatedAt:  time.Now(),
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
		0, // syntax score not available
		getScoreOrDefault(validationResult.SecurityResult),
		getStringOrDefault(validationResult.SecurityResult),
		getScoreOrDefault(validationResult.QualityResult),
		0, // critique score not available
		validationResult.ValidationTime,
	)

	da.Status = AgentStatusCompleted

	payload, _ = json.Marshal(map[string]interface{}{
		"agent_id":          da.ID,
		"task_id":           da.Task.ID,
		"output_size":       len(da.Output),
		"duration_ms":       time.Since(da.StartTime).Milliseconds(),
		"sandbox_success":   sandboxResult.Success,
		"security_score":    sandboxResult.SecurityScore,
		"execution_time":    sandboxResult.ExecutionTime.Milliseconds(),
		"validation_score":  validationResult.OverallScore,
		"validation_passed": validationResult.Passed,
	})
	da.EventManager.Publish(ctx, events.Event{
		ID:        fmt.Sprintf("agent_%s_completed", da.ID),
		Type:      "task.completed",
		Timestamp: time.Now(),
		Source:    da.ID,
		Payload:   payload,
	})

	logger.WithComponent("agents").With(zap.String("agent_id", da.ID)).Info("Task completed",
		zap.String("task_id", da.Task.ID),
		zap.Duration("total_duration", time.Since(da.StartTime)),
		zap.Int("validation_score", validationResult.OverallScore),
		zap.Bool("validation_passed", validationResult.Passed))
	return nil
}

func (da *DynamicAgent) formatExecutionContext() string {
	var context strings.Builder
	context.WriteString(fmt.Sprintf("Project Type: %s\n", da.Context.ProjectType))
	context.WriteString(fmt.Sprintf("Tech Stack: %v\n", da.Context.TechStack))
	context.WriteString(fmt.Sprintf("Dependencies: %v\n", da.Task.Dependencies))

	if len(da.Context.PreviousOutputs) > 0 {
		context.WriteString("Previous Task Outputs Available:\n")
		for taskID := range da.Context.PreviousOutputs {
			context.WriteString(fmt.Sprintf("- %s\n", taskID))
		}
	}

	return context.String()
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
	case *types.SecurityResult:
		if r != nil {
			return r.Score
		}
	case *types.QualityResult:
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
	case *types.SecurityResult:
		if r != nil {
			return string(r.RiskLevel)
		}
	}
	return "unknown"
}
