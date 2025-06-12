package engines

import (
	"context"
	"fmt"
	"sync"
	"time"

	"QLP/services/orchestrator-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// WorkflowEngine manages workflow execution state and control
type WorkflowEngine struct {
	workflows map[string]*contracts.WorkflowExecution
	mu        sync.RWMutex
	dagEngine *DAGEngine
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(dagEngine *DAGEngine) *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]*contracts.WorkflowExecution),
		dagEngine: dagEngine,
	}
}

// ExecuteWorkflow starts a new workflow execution
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, req *contracts.ExecuteWorkflowRequest) (*contracts.ExecuteWorkflowResponse, error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	logger.WithComponent("workflow-engine").Info("Starting workflow execution",
		zap.String("intent_id", req.IntentID),
		zap.Int("task_count", len(req.Tasks)))

	// Validate DAG structure first
	dagReq := &contracts.DAGValidationRequest{
		Tasks:        req.Tasks,
		Dependencies: req.Dependencies,
	}

	dagResp, err := we.dagEngine.ValidateDAG(ctx, dagReq)
	if err != nil {
		return nil, fmt.Errorf("DAG validation failed: %w", err)
	}

	if !dagResp.Valid {
		return &contracts.ExecuteWorkflowResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid DAG structure: %v", dagResp.Errors),
		}, fmt.Errorf("invalid DAG: %v", dagResp.Errors)
	}

	// Generate workflow ID
	workflowID := fmt.Sprintf("wf_%s_%d", req.IntentID, time.Now().Unix())

	// Create workflow execution
	now := time.Now()
	workflow := &contracts.WorkflowExecution{
		ID:            workflowID,
		IntentID:      req.IntentID,
		Status:        contracts.WorkflowStatusPending,
		Tasks:         we.convertToTaskExecutions(req.Tasks),
		Dependencies:  req.Dependencies,
		Configuration: req.Configuration,
		Context:       req.Context,
		Progress:      we.calculateInitialProgress(req.Tasks),
		Results:       make(map[string]contracts.TaskResult),
		Errors:        []contracts.WorkflowError{},
		CreatedAt:     now,
		Duration:      0,
		Metadata:      req.Metadata,
	}

	// Store workflow
	we.workflows[workflowID] = workflow

	// Start execution in background
	go we.executeWorkflowAsync(context.Background(), workflow)

	logger.WithComponent("workflow-engine").Info("Workflow started",
		zap.String("workflow_id", workflowID),
		zap.String("status", string(workflow.Status)))

	return &contracts.ExecuteWorkflowResponse{
		WorkflowID: workflowID,
		Status:     string(workflow.Status),
		Message:    "Workflow execution started",
		Execution:  workflow,
	}, nil
}

// GetWorkflow retrieves a workflow execution by ID
func (we *WorkflowEngine) GetWorkflow(ctx context.Context, workflowID string) (*contracts.GetWorkflowResponse, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return &contracts.GetWorkflowResponse{
		WorkflowID: workflowID,
		Execution:  workflow,
	}, nil
}

// ListWorkflows lists workflows with filtering and pagination
func (we *WorkflowEngine) ListWorkflows(ctx context.Context, page, pageSize int, status string) (*contracts.ListWorkflowsResponse, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	var filteredWorkflows []*contracts.WorkflowExecution
	for _, workflow := range we.workflows {
		if status == "" || string(workflow.Status) == status {
			filteredWorkflows = append(filteredWorkflows, workflow)
		}
	}

	// Apply pagination
	total := len(filteredWorkflows)
	start := page * pageSize
	end := start + pageSize

	if start >= total {
		return &contracts.ListWorkflowsResponse{
			Workflows: []contracts.WorkflowSummary{},
			Total:     total,
			Page:      page,
			PageSize:  pageSize,
		}, nil
	}

	if end > total {
		end = total
	}

	// Convert to summaries
	var summaries []contracts.WorkflowSummary
	for i := start; i < end; i++ {
		workflow := filteredWorkflows[i]
		summary := contracts.WorkflowSummary{
			ID:          workflow.ID,
			IntentID:    workflow.IntentID,
			Status:      workflow.Status,
			Progress:    workflow.Progress,
			CreatedAt:   workflow.CreatedAt,
			CompletedAt: workflow.CompletedAt,
			Duration:    workflow.Duration,
		}
		summaries = append(summaries, summary)
	}

	return &contracts.ListWorkflowsResponse{
		Workflows: summaries,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// PauseWorkflow pauses a running workflow
func (we *WorkflowEngine) PauseWorkflow(ctx context.Context, workflowID string, req *contracts.PauseWorkflowRequest) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if workflow.Status != contracts.WorkflowStatusRunning {
		return fmt.Errorf("workflow is not running, current status: %s", workflow.Status)
	}

	workflow.Status = contracts.WorkflowStatusPaused
	workflow.Metadata["pause_reason"] = req.Reason
	workflow.Metadata["paused_at"] = time.Now().Format(time.RFC3339)

	logger.WithComponent("workflow-engine").Info("Workflow paused",
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason))

	return nil
}

// ResumeWorkflow resumes a paused workflow
func (we *WorkflowEngine) ResumeWorkflow(ctx context.Context, workflowID string, req *contracts.ResumeWorkflowRequest) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if workflow.Status != contracts.WorkflowStatusPaused {
		return fmt.Errorf("workflow is not paused, current status: %s", workflow.Status)
	}

	workflow.Status = contracts.WorkflowStatusRunning
	workflow.Metadata["resume_reason"] = req.Reason
	workflow.Metadata["resumed_at"] = time.Now().Format(time.RFC3339)

	// Resume execution in background
	go we.executeWorkflowAsync(context.Background(), workflow)

	logger.WithComponent("workflow-engine").Info("Workflow resumed",
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason))

	return nil
}

// CancelWorkflow cancels a workflow execution
func (we *WorkflowEngine) CancelWorkflow(ctx context.Context, workflowID string, req *contracts.CancelWorkflowRequest) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if workflow.Status == contracts.WorkflowStatusCompleted || workflow.Status == contracts.WorkflowStatusCancelled {
		return fmt.Errorf("workflow is already completed or cancelled")
	}

	workflow.Status = contracts.WorkflowStatusCancelled
	workflow.Metadata["cancel_reason"] = req.Reason
	workflow.Metadata["cancelled_at"] = time.Now().Format(time.RFC3339)
	workflow.Metadata["force_cancel"] = req.Force

	// Cancel running tasks
	for i := range workflow.Tasks {
		task := &workflow.Tasks[i]
		if task.Status == contracts.TaskStatusRunning || task.Status == contracts.TaskStatusQueued {
			task.Status = contracts.TaskStatusCancelled
			now := time.Now()
			task.CompletedAt = &now
		}
	}

	// Update progress
	workflow.Progress = we.calculateProgress(workflow.Tasks)
	
	now := time.Now()
	workflow.CompletedAt = &now
	workflow.Duration = now.Sub(workflow.CreatedAt)

	logger.WithComponent("workflow-engine").Info("Workflow cancelled",
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason),
		zap.Bool("force", req.Force))

	return nil
}

// RetryTask retries a failed task
func (we *WorkflowEngine) RetryTask(ctx context.Context, workflowID string, req *contracts.RetryTaskRequest) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Find the task
	var taskIndex = -1
	for i, task := range workflow.Tasks {
		if task.Task.ID == req.TaskID {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("task not found: %s", req.TaskID)
	}

	task := &workflow.Tasks[taskIndex]
	if task.Status != contracts.TaskStatusFailed {
		return fmt.Errorf("task is not in failed state: %s", task.Status)
	}

	// Reset task state for retry
	task.Status = contracts.TaskStatusQueued
	task.StartedAt = nil
	task.CompletedAt = nil
	task.Duration = 0
	task.Attempts++
	task.Error = nil
	task.Result = nil

	// Update workflow progress
	workflow.Progress = we.calculateProgress(workflow.Tasks)

	logger.WithComponent("workflow-engine").Info("Task queued for retry",
		zap.String("workflow_id", workflowID),
		zap.String("task_id", req.TaskID),
		zap.String("reason", req.Reason),
		zap.Int("attempt", task.Attempts))

	return nil
}

// GetWorkflowMetrics retrieves metrics for a workflow
func (we *WorkflowEngine) GetWorkflowMetrics(ctx context.Context, workflowID string) (*contracts.WorkflowMetrics, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Calculate task metrics
	var taskMetrics []contracts.TaskMetrics
	var totalRetries int
	for _, taskExec := range workflow.Tasks {
		taskMetrics = append(taskMetrics, contracts.TaskMetrics{
			TaskID:        taskExec.Task.ID,
			ExecutionTime: taskExec.Duration,
			ResourceUsage: contracts.ResourceUsage{
				CPUTime:    taskExec.Duration / 2, // Estimate
				MemoryUsed: 64 * 1024 * 1024,     // 64MB estimate
				DiskUsed:   10 * 1024 * 1024,     // 10MB estimate
				NetworkIO:  1024 * 1024,          // 1MB estimate
			},
			RetryCount:    taskExec.Attempts - 1,
			CustomMetrics: taskExec.Metadata,
		})
		totalRetries += taskExec.Attempts - 1
	}

	// Calculate error rate
	var failedTasks int
	for _, taskExec := range workflow.Tasks {
		if taskExec.Status == contracts.TaskStatusFailed {
			failedTasks++
		}
	}
	
	errorRate := float64(failedTasks) / float64(len(workflow.Tasks))
	
	// Calculate throughput
	var throughput float64
	if workflow.Duration > 0 {
		throughput = float64(len(workflow.Tasks)) / workflow.Duration.Hours()
	}

	metrics := &contracts.WorkflowMetrics{
		WorkflowID:    workflowID,
		ExecutionTime: workflow.Duration,
		TaskMetrics:   taskMetrics,
		ResourceUsage: contracts.ResourceUsage{
			CPUTime:    workflow.Duration,
			MemoryUsed: int64(len(workflow.Tasks)) * 64 * 1024 * 1024,
			DiskUsed:   int64(len(workflow.Tasks)) * 10 * 1024 * 1024,
			NetworkIO:  int64(len(workflow.Tasks)) * 1024 * 1024,
		},
		ErrorRate:       errorRate,
		ThroughputTasks: throughput,
		CustomMetrics: map[string]interface{}{
			"total_retries": totalRetries,
			"total_tasks":   len(workflow.Tasks),
			"completed_tasks": workflow.Progress.CompletedTasks,
			"failed_tasks":    workflow.Progress.FailedTasks,
		},
	}

	return metrics, nil
}

// executeWorkflowAsync executes workflow tasks asynchronously
func (we *WorkflowEngine) executeWorkflowAsync(ctx context.Context, workflow *contracts.WorkflowExecution) {
	logger.WithComponent("workflow-engine").Info("Starting async workflow execution",
		zap.String("workflow_id", workflow.ID))

	we.mu.Lock()
	workflow.Status = contracts.WorkflowStatusRunning
	startTime := time.Now()
	workflow.StartedAt = &startTime
	we.mu.Unlock()

	// Simulate task execution based on DAG order
	dagReq := &contracts.DAGValidationRequest{
		Tasks:        we.extractTasks(workflow.Tasks),
		Dependencies: workflow.Dependencies,
	}

	dagResp, err := we.dagEngine.ValidateDAG(ctx, dagReq)
	if err != nil || !dagResp.Valid {
		we.failWorkflow(workflow, fmt.Errorf("DAG validation failed during execution: %v", err))
		return
	}

	// Execute tasks in topological order
	for _, taskID := range dagResp.ExecutionOrder {
		// Check if workflow is still running
		we.mu.RLock()
		status := workflow.Status
		we.mu.RUnlock()

		if status != contracts.WorkflowStatusRunning {
			logger.WithComponent("workflow-engine").Info("Workflow execution stopped",
				zap.String("workflow_id", workflow.ID),
				zap.String("status", string(status)))
			return
		}

		// Find and execute task
		taskIndex := we.findTaskIndex(workflow.Tasks, taskID)
		if taskIndex >= 0 {
			we.executeTask(workflow, taskIndex)
		}
	}

	// Complete workflow
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow.Status = contracts.WorkflowStatusCompleted
	now := time.Now()
	workflow.CompletedAt = &now
	workflow.Duration = now.Sub(workflow.CreatedAt)
	workflow.Progress = we.calculateProgress(workflow.Tasks)

	logger.WithComponent("workflow-engine").Info("Workflow completed",
		zap.String("workflow_id", workflow.ID),
		zap.Duration("duration", workflow.Duration))
}

// Helper methods

func (we *WorkflowEngine) convertToTaskExecutions(tasks []contracts.Task) []contracts.TaskExecution {
	var executions []contracts.TaskExecution
	for _, task := range tasks {
		execution := contracts.TaskExecution{
			Task:      task,
			Status:    contracts.TaskStatusPending,
			AgentID:   "",
			Duration:  0,
			Attempts:  1,
			Result:    nil,
			Error:     nil,
			Logs:      []string{},
			Metadata:  make(map[string]interface{}),
		}
		executions = append(executions, execution)
	}
	return executions
}

func (we *WorkflowEngine) extractTasks(executions []contracts.TaskExecution) []contracts.Task {
	var tasks []contracts.Task
	for _, exec := range executions {
		tasks = append(tasks, exec.Task)
	}
	return tasks
}

func (we *WorkflowEngine) calculateInitialProgress(tasks []contracts.Task) contracts.WorkflowProgress {
	return contracts.WorkflowProgress{
		TotalTasks:        len(tasks),
		CompletedTasks:    0,
		FailedTasks:       0,
		RunningTasks:      0,
		PendingTasks:      len(tasks),
		PercentComplete:   0.0,
		EstimatedTimeLeft: 0,
	}
}

func (we *WorkflowEngine) calculateProgress(executions []contracts.TaskExecution) contracts.WorkflowProgress {
	progress := contracts.WorkflowProgress{
		TotalTasks: len(executions),
	}

	for _, exec := range executions {
		switch exec.Status {
		case contracts.TaskStatusCompleted:
			progress.CompletedTasks++
		case contracts.TaskStatusFailed:
			progress.FailedTasks++
		case contracts.TaskStatusRunning:
			progress.RunningTasks++
		case contracts.TaskStatusPending, contracts.TaskStatusQueued:
			progress.PendingTasks++
		}
	}

	if progress.TotalTasks > 0 {
		progress.PercentComplete = float64(progress.CompletedTasks) / float64(progress.TotalTasks) * 100
	}

	return progress
}

func (we *WorkflowEngine) findTaskIndex(executions []contracts.TaskExecution, taskID string) int {
	for i, exec := range executions {
		if exec.Task.ID == taskID {
			return i
		}
	}
	return -1
}

func (we *WorkflowEngine) executeTask(workflow *contracts.WorkflowExecution, taskIndex int) {
	we.mu.Lock()
	defer we.mu.Unlock()

	task := &workflow.Tasks[taskIndex]
	task.Status = contracts.TaskStatusRunning
	startTime := time.Now()
	task.StartedAt = &startTime
	task.AgentID = fmt.Sprintf("agent_%s_%d", task.Task.ID, time.Now().Unix())

	logger.WithComponent("workflow-engine").Info("Executing task",
		zap.String("workflow_id", workflow.ID),
		zap.String("task_id", task.Task.ID),
		zap.String("agent_id", task.AgentID))

	// Simulate task execution
	go func() {
		time.Sleep(1 * time.Second) // Simulate work

		we.mu.Lock()
		defer we.mu.Unlock()

		// Complete task
		task.Status = contracts.TaskStatusCompleted
		now := time.Now()
		task.CompletedAt = &now
		task.Duration = now.Sub(*task.StartedAt)

		// Create mock result
		task.Result = &contracts.TaskResult{
			Output:    fmt.Sprintf("Task %s completed successfully", task.Task.ID),
			Files:     map[string]string{},
			Artifacts: []string{},
			Metrics:   map[string]interface{}{"execution_time_ms": task.Duration.Milliseconds()},
		}

		// Update workflow progress
		workflow.Progress = we.calculateProgress(workflow.Tasks)

		logger.WithComponent("workflow-engine").Info("Task completed",
			zap.String("workflow_id", workflow.ID),
			zap.String("task_id", task.Task.ID),
			zap.Duration("duration", task.Duration))
	}()
}

func (we *WorkflowEngine) failWorkflow(workflow *contracts.WorkflowExecution, err error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow.Status = contracts.WorkflowStatusFailed
	workflow.Errors = append(workflow.Errors, contracts.WorkflowError{
		Type:      "execution_error",
		Message:   err.Error(),
		Code:      "WORKFLOW_EXECUTION_FAILED",
		Details:   err.Error(),
		Timestamp: time.Now(),
	})

	now := time.Now()
	workflow.CompletedAt = &now
	workflow.Duration = now.Sub(workflow.CreatedAt)

	logger.WithComponent("workflow-engine").Error("Workflow failed",
		zap.String("workflow_id", workflow.ID),
		zap.Error(err))
}