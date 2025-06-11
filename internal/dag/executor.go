package dag

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"QLP/internal/agents"
	"QLP/internal/events"
	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/validation"
)

type TaskResult struct {
	AgentID          string
	Status           models.TaskStatus
	Output           string
	ExecutionTime    time.Duration
	SandboxResult    *sandbox.SandboxExecutionResult
	ValidationResult *validation.ValidationResult
	Error            error
	StartTime        time.Time
	EndTime          time.Time
}

type DAGExecutor struct {
	eventBus       *events.EventBus
	agentFactory   *agents.AgentFactory
	taskStates     map[string]models.TaskStatus
	taskResults    map[string]*TaskResult
	mu             sync.RWMutex
	waitingTasks   chan models.Task
	projectContext agents.ProjectContext
	maxConcurrency int
	semaphore      chan struct{}
}

func NewDAGExecutor(eventBus *events.EventBus, agentFactory *agents.AgentFactory) *DAGExecutor {
	projectContext := agents.ProjectContext{
		ProjectType:  "web_api",
		TechStack:    []string{"Go", "HTTP", "JSON"},
		Requirements: []string{"RESTful API", "Authentication", "Error handling"},
		Constraints: map[string]string{
			"performance": "high",
			"security":    "required",
			"scalability": "horizontal",
		},
		Architecture: "microservices",
	}

	maxConcurrency := 4 // Limit to 4 concurrent agents
	
	return &DAGExecutor{
		eventBus:       eventBus,
		agentFactory:   agentFactory,
		taskStates:     make(map[string]models.TaskStatus),
		taskResults:    make(map[string]*TaskResult),
		waitingTasks:   make(chan models.Task, 100),
		projectContext: projectContext,
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

func (de *DAGExecutor) ExecuteTaskGraph(ctx context.Context, taskGraph *models.TaskGraph) error {
	log.Printf("Starting DAG execution with %d tasks", len(taskGraph.Tasks))

	for _, task := range taskGraph.Tasks {
		de.mu.Lock()
		de.taskStates[task.ID] = models.TaskStatusPending
		de.mu.Unlock()
	}

	completedChan := make(chan string, len(taskGraph.Tasks))

	var executeTasksRecursively func([]models.Task)
	executeTasksRecursively = func(tasks []models.Task) {
		var wg sync.WaitGroup
		for _, task := range tasks {
			// Check if task is already being executed or completed
			de.mu.RLock()
			status, exists := de.taskStates[task.ID]
			de.mu.RUnlock()
			
			if exists && (status == models.TaskStatusInProgress || status == models.TaskStatusCompleted) {
				continue // Skip tasks that are already running or completed
			}
			
			wg.Add(1)
			go func(t models.Task) {
				defer wg.Done()
				
				// Acquire semaphore (limit concurrency)
				de.semaphore <- struct{}{}
				defer func() { <-de.semaphore }()
				
				if err := de.executeTaskWithDynamicAgent(ctx, t, completedChan); err != nil {
					log.Printf("Task %s failed: %v", t.ID, err)
				}
			}(task)
		}
		wg.Wait()
	}

	readyTasks := de.findReadyTasks(taskGraph.Tasks)
	executeTasksRecursively(readyTasks)

	completedCount := 0
	for completedCount < len(taskGraph.Tasks) {
		select {
		case taskID := <-completedChan:
			completedCount++
			log.Printf("Task completed: %s (%d/%d)", taskID, completedCount, len(taskGraph.Tasks))

			nextTasks := de.findNextReadyTasks(taskID, taskGraph)
			if len(nextTasks) > 0 {
				go executeTasksRecursively(nextTasks)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	log.Println("All tasks completed successfully")
	return nil
}

func (de *DAGExecutor) executeTaskWithDynamicAgent(ctx context.Context, task models.Task, completedChan chan<- string) error {
	startTime := time.Now()
	
	// Double-check task state to prevent race conditions
	de.mu.Lock()
	if status, exists := de.taskStates[task.ID]; exists && (status == models.TaskStatusInProgress || status == models.TaskStatusCompleted) {
		de.mu.Unlock()
		return nil // Task already being executed or completed
	}
	de.taskStates[task.ID] = models.TaskStatusInProgress
	de.mu.Unlock()

	de.eventBus.Publish(events.Event{
		ID:        fmt.Sprintf("event_%s_started", task.ID),
		Type:      events.EventTaskStarted,
		Timestamp: time.Now(),
		Source:    "dag_executor",
		Payload: map[string]interface{}{
			"task_id":     task.ID,
			"task_type":   task.Type,
			"description": task.Description,
		},
	})

	log.Printf("Creating dynamic agent for task: %s - %s", task.ID, task.Description)

	agent, err := de.agentFactory.CreateAgent(ctx, task, de.projectContext)
	if err != nil {
		de.mu.Lock()
		de.taskStates[task.ID] = models.TaskStatusFailed
		de.taskResults[task.ID] = &TaskResult{
			AgentID:       "",
			Status:        models.TaskStatusFailed,
			Output:        "",
			ExecutionTime: time.Since(startTime),
			Error:         err,
			StartTime:     startTime,
			EndTime:       time.Now(),
		}
		de.mu.Unlock()

		de.eventBus.Publish(events.Event{
			ID:        fmt.Sprintf("event_%s_failed", task.ID),
			Type:      events.EventTaskFailed,
			Timestamp: time.Now(),
			Source:    "dag_executor",
			Payload: map[string]interface{}{
				"task_id": task.ID,
				"error":   err.Error(),
			},
		})

		return fmt.Errorf("failed to create agent: %w", err)
	}

	if err := de.agentFactory.ExecuteAgent(ctx, agent); err != nil {
		de.mu.Lock()
		de.taskStates[task.ID] = models.TaskStatusFailed
		de.taskResults[task.ID] = &TaskResult{
			AgentID:       agent.ID,
			Status:        models.TaskStatusFailed,
			Output:        agent.GetOutput(),
			ExecutionTime: time.Since(startTime),
			Error:         err,
			StartTime:     startTime,
			EndTime:       time.Now(),
		}
		de.mu.Unlock()

		de.eventBus.Publish(events.Event{
			ID:        fmt.Sprintf("event_%s_failed", task.ID),
			Type:      events.EventTaskFailed,
			Timestamp: time.Now(),
			Source:    "dag_executor",
			Payload: map[string]interface{}{
				"task_id": task.ID,
				"error":   err.Error(),
			},
		})

		return fmt.Errorf("agent execution failed: %w", err)
	}

	de.mu.Lock()
	de.taskStates[task.ID] = models.TaskStatusCompleted
	de.taskResults[task.ID] = &TaskResult{
		AgentID:          agent.ID,
		Status:           models.TaskStatusCompleted,
		Output:           agent.GetOutput(),
		ExecutionTime:    time.Since(startTime),
		SandboxResult:    agent.SandboxResult,
		ValidationResult: agent.ValidationResult,
		Error:            nil,
		StartTime:        startTime,
		EndTime:          time.Now(),
	}
	de.mu.Unlock()

	de.eventBus.Publish(events.Event{
		ID:        fmt.Sprintf("event_%s_completed", task.ID),
		Type:      events.EventTaskCompleted,
		Timestamp: time.Now(),
		Source:    "dag_executor",
		Payload: map[string]interface{}{
			"task_id":     task.ID,
			"agent_id":    agent.ID,
			"output_size": len(agent.GetOutput()),
		},
	})

	de.agentFactory.CleanupAgent(agent.ID)
	completedChan <- task.ID

	return nil
}

func (de *DAGExecutor) findReadyTasks(tasks []models.Task) []models.Task {
	var readyTasks []models.Task

	for _, task := range tasks {
		if len(task.Dependencies) == 0 {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks
}

func (de *DAGExecutor) findNextReadyTasks(_ string, taskGraph *models.TaskGraph) []models.Task {
	var readyTasks []models.Task

	for _, task := range taskGraph.Tasks {
		if de.taskStates[task.ID] != models.TaskStatusPending {
			continue
		}

		if de.dependenciesCompleted(task.Dependencies) {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks
}

func (de *DAGExecutor) dependenciesCompleted(dependencies []string) bool {
	de.mu.RLock()
	defer de.mu.RUnlock()

	for _, depID := range dependencies {
		if de.taskStates[depID] != models.TaskStatusCompleted {
			return false
		}
	}

	return true
}

func (de *DAGExecutor) allTasksCompleted() bool {
	de.mu.RLock()
	defer de.mu.RUnlock()

	for _, status := range de.taskStates {
		if status != models.TaskStatusCompleted {
			return false
		}
	}

	return true
}

func (de *DAGExecutor) GetTaskResult(taskID string) *TaskResult {
	de.mu.RLock()
	defer de.mu.RUnlock()
	return de.taskResults[taskID]
}
