package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/worker-runtime/internal/agents"
	"QLP/services/worker-runtime/internal/sandbox"
	"QLP/services/worker-runtime/pkg/contracts"
)

type TaskExecutor struct {
	config       Config
	agentFactory *agents.Factory
	sandboxMgr   *sandbox.Manager
	executions   map[string]*contracts.WorkerExecution
	executionsMu sync.RWMutex
	semaphore    chan struct{}
}

type Config struct {
	AgentFactory   *agents.Factory
	SandboxManager *sandbox.Manager
	MaxConcurrent  int
	DefaultTimeout time.Duration
}

func NewTaskExecutor(config Config) *TaskExecutor {
	return &TaskExecutor{
		config:       config,
		agentFactory: config.AgentFactory,
		sandboxMgr:   config.SandboxManager,
		executions:   make(map[string]*contracts.WorkerExecution),
		semaphore:    make(chan struct{}, config.MaxConcurrent),
	}
}

func (te *TaskExecutor) ExecuteTask(ctx context.Context, req *contracts.ExecuteTaskRequest) (*contracts.ExecuteTaskResponse, error) {
	executionID := uuid.New().String()
	
	logger.WithComponent("task-executor").Info("Starting task execution",
		zap.String("execution_id", executionID),
		zap.String("task_id", req.Task.ID),
		zap.String("task_type", string(req.Task.Type)),
		zap.String("tenant_id", req.Task.TenantID))

	// Create execution record
	execution := &contracts.WorkerExecution{
		ID:        executionID,
		TaskID:    req.Task.ID,
		TenantID:  req.Task.TenantID,
		Status:    contracts.ExecutionStatusPending,
		StartTime: time.Now(),
	}

	// Store execution
	te.executionsMu.Lock()
	te.executions[executionID] = execution
	te.executionsMu.Unlock()

	// Execute task asynchronously
	go te.executeTaskAsync(ctx, req, execution)

	return &contracts.ExecuteTaskResponse{
		ExecutionID: executionID,
		Status:      string(contracts.ExecutionStatusPending),
		Message:     "Task execution started",
	}, nil
}

func (te *TaskExecutor) executeTaskAsync(ctx context.Context, req *contracts.ExecuteTaskRequest, execution *contracts.WorkerExecution) {
	// Acquire semaphore for concurrency control
	te.semaphore <- struct{}{}
	defer func() { <-te.semaphore }()

	// Update status to running
	te.updateExecutionStatus(execution.ID, contracts.ExecutionStatusRunning)

	// Set timeout
	timeout := te.config.DefaultTimeout
	if req.Task.TimeoutSeconds > 0 {
		timeout = time.Duration(req.Task.TimeoutSeconds) * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the task
	err := te.runTask(execCtx, req, execution)
	if err != nil {
		logger.WithComponent("task-executor").Error("Task execution failed",
			zap.String("execution_id", execution.ID),
			zap.Error(err))
		
		te.updateExecutionError(execution.ID, err.Error())
		te.updateExecutionStatus(execution.ID, contracts.ExecutionStatusFailed)
		return
	}

	// Mark as completed
	te.updateExecutionStatus(execution.ID, contracts.ExecutionStatusCompleted)
	
	logger.WithComponent("task-executor").Info("Task execution completed",
		zap.String("execution_id", execution.ID),
		zap.String("task_id", req.Task.ID),
		zap.Duration("duration", time.Since(execution.StartTime)))
}

func (te *TaskExecutor) runTask(ctx context.Context, req *contracts.ExecuteTaskRequest, execution *contracts.WorkerExecution) error {
	// Step 1: Create agent for the task
	agentCtx := &contracts.AgentContext{
		ProjectType:  "microservice", // Default
		TechStack:    []string{"go"}, // Default
		Requirements: []string{},
		Constraints:  make(map[string]string),
		Architecture: "microservices",
	}

	agent, err := te.agentFactory.CreateAgent(ctx, &req.Task, agentCtx)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	execution.AgentID = fmt.Sprintf("%s-agent", req.Task.Type)

	// Step 2: Execute agent to generate code/output
	agentResult, err := agent.Execute(ctx, &req.Task, agentCtx)
	if err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	execution.Output = agentResult.Output

	// Step 3: If there's code to execute, run it in sandbox
	if agentResult.Code != "" {
		// Create task with generated code
		sandboxTask := &contracts.WorkerTask{
			ID:             req.Task.ID + "-sandbox",
			Type:           req.Task.Type,
			Description:    req.Task.Description,
			Code:           agentResult.Code,
			Language:       agentResult.Language,
			TenantID:       req.Task.TenantID,
			ResourceLimits: req.Task.ResourceLimits,
			TimeoutSeconds: req.Task.TimeoutSeconds,
		}

		sandboxResult, err := te.sandboxMgr.ExecuteTask(ctx, sandboxTask)
		if err != nil {
			return fmt.Errorf("sandbox execution failed: %w", err)
		}

		execution.SandboxResult = sandboxResult

		// Update output with sandbox results if successful
		if sandboxResult.ExitCode == 0 && sandboxResult.Stdout != "" {
			execution.Output += "\n\nExecution Output:\n" + sandboxResult.Stdout
		} else if sandboxResult.ExitCode != 0 {
			execution.Output += "\n\nExecution Failed:\n" + sandboxResult.Stderr
		}
	}

	// Step 4: Validation (if requested)
	if req.ValidateOutput {
		validation := te.performValidation(agentResult, execution.SandboxResult)
		execution.ValidationResult = validation
	}

	// Step 5: Calculate execution time
	now := time.Now()
	execution.EndTime = &now
	execution.ExecutionTime = now.Sub(execution.StartTime)

	return nil
}

func (te *TaskExecutor) performValidation(agentResult *agents.AgentResult, sandboxResult *contracts.SandboxResult) *contracts.ValidationResult {
	// Mock validation - in real system, this would call validation service
	score := 85
	passed := true

	if sandboxResult != nil && sandboxResult.ExitCode != 0 {
		score -= 20
		passed = false
	}

	validation := &contracts.ValidationResult{
		OverallScore:   score,
		SecurityScore:  90,
		QualityScore:   score,
		Passed:         passed,
		Issues:         []contracts.ValidationIssue{},
		Warnings:       []contracts.ValidationWarning{},
		ValidationTime: 100 * time.Millisecond,
	}

	// Add issues based on sandbox result
	if sandboxResult != nil && sandboxResult.ExitCode != 0 {
		validation.Issues = append(validation.Issues, contracts.ValidationIssue{
			Type:     "execution",
			Severity: "error",
			Message:  "Code execution failed",
		})
	}

	return validation
}

func (te *TaskExecutor) GetExecution(executionID, tenantID string) (*contracts.WorkerExecution, error) {
	te.executionsMu.RLock()
	defer te.executionsMu.RUnlock()

	execution, exists := te.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found")
	}

	if execution.TenantID != tenantID {
		return nil, fmt.Errorf("execution not found for tenant")
	}

	return execution, nil
}

func (te *TaskExecutor) ListExecutions(req *contracts.ListExecutionsRequest) (*contracts.ListExecutionsResponse, error) {
	te.executionsMu.RLock()
	defer te.executionsMu.RUnlock()

	var filtered []*contracts.WorkerExecution
	for _, execution := range te.executions {
		if execution.TenantID != req.TenantID {
			continue
		}

		if req.Status != "" && execution.Status != req.Status {
			continue
		}

		if req.Since != nil && execution.StartTime.Before(*req.Since) {
			continue
		}

		filtered = append(filtered, execution)
	}

	// Apply pagination
	total := len(filtered)
	start := req.Offset
	end := start + req.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	if req.Limit == 0 {
		end = total
	}

	var paginatedExecutions []contracts.WorkerExecution
	for i := start; i < end; i++ {
		paginatedExecutions = append(paginatedExecutions, *filtered[i])
	}

	return &contracts.ListExecutionsResponse{
		Executions: paginatedExecutions,
		Total:      total,
	}, nil
}

func (te *TaskExecutor) CancelExecution(executionID, tenantID string) error {
	te.executionsMu.Lock()
	defer te.executionsMu.Unlock()

	execution, exists := te.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found")
	}

	if execution.TenantID != tenantID {
		return fmt.Errorf("execution not found for tenant")
	}

	if execution.Status == contracts.ExecutionStatusCompleted ||
		execution.Status == contracts.ExecutionStatusFailed ||
		execution.Status == contracts.ExecutionStatusCanceled {
		return fmt.Errorf("execution cannot be canceled in current status: %s", execution.Status)
	}

	execution.Status = contracts.ExecutionStatusCanceled
	now := time.Now()
	execution.EndTime = &now
	execution.ExecutionTime = now.Sub(execution.StartTime)

	logger.WithComponent("task-executor").Info("Execution canceled",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", tenantID))

	return nil
}

func (te *TaskExecutor) Shutdown(ctx context.Context) {
	logger.WithComponent("task-executor").Info("Shutting down task executor")

	// Cancel all running executions
	te.executionsMu.Lock()
	for _, execution := range te.executions {
		if execution.Status == contracts.ExecutionStatusRunning ||
			execution.Status == contracts.ExecutionStatusPending {
			execution.Status = contracts.ExecutionStatusCanceled
		}
	}
	te.executionsMu.Unlock()

	// Wait for all workers to finish or timeout
	for i := 0; i < te.config.MaxConcurrent; i++ {
		select {
		case te.semaphore <- struct{}{}:
			// Successfully acquired, worker finished
			<-te.semaphore
		case <-ctx.Done():
			// Timeout, force shutdown
			logger.WithComponent("task-executor").Warn("Force shutdown due to timeout")
			return
		}
	}

	logger.WithComponent("task-executor").Info("Task executor shutdown complete")
}

// Helper methods
func (te *TaskExecutor) updateExecutionStatus(executionID string, status contracts.ExecutionStatus) {
	te.executionsMu.Lock()
	defer te.executionsMu.Unlock()

	if execution, exists := te.executions[executionID]; exists {
		execution.Status = status
		if status == contracts.ExecutionStatusCompleted || 
		   status == contracts.ExecutionStatusFailed || 
		   status == contracts.ExecutionStatusCanceled {
			now := time.Now()
			execution.EndTime = &now
			execution.ExecutionTime = now.Sub(execution.StartTime)
		}
	}
}

func (te *TaskExecutor) updateExecutionError(executionID, errorMsg string) {
	te.executionsMu.Lock()
	defer te.executionsMu.Unlock()

	if execution, exists := te.executions[executionID]; exists {
		execution.Error = errorMsg
	}
}