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

// DAGEngine handles DAG validation and execution planning
type DAGEngine struct {
	mu sync.RWMutex
}

// NewDAGEngine creates a new DAG engine
func NewDAGEngine() *DAGEngine {
	return &DAGEngine{}
}

// ValidateDAG validates a DAG structure for cycles and dependencies
func (de *DAGEngine) ValidateDAG(ctx context.Context, req *contracts.DAGValidationRequest) (*contracts.DAGValidationResponse, error) {
	de.mu.Lock()
	defer de.mu.Unlock()

	logger.WithComponent("dag-engine").Info("Validating DAG structure",
		zap.Int("task_count", len(req.Tasks)),
		zap.Int("dependency_count", len(req.Dependencies)))

	response := &contracts.DAGValidationResponse{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check for valid task IDs
	taskIDs := make(map[string]bool)
	for _, task := range req.Tasks {
		if task.ID == "" {
			response.Valid = false
			response.Errors = append(response.Errors, "Task ID cannot be empty")
			continue
		}
		if taskIDs[task.ID] {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Duplicate task ID: %s", task.ID))
		}
		taskIDs[task.ID] = true
	}

	// Validate dependencies exist
	for _, dep := range req.Dependencies {
		if !taskIDs[dep.From] {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Dependency references non-existent task: %s", dep.From))
		}
		if !taskIDs[dep.To] {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Dependency references non-existent task: %s", dep.To))
		}
		if dep.From == dep.To {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Self-dependency detected: %s", dep.From))
		}
	}

	// Check for cycles using DFS
	if response.Valid {
		if hasCycle, cyclePath := de.detectCycles(req.Tasks, req.Dependencies); hasCycle {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Cycle detected: %s", cyclePath))
		}
	}

	// Generate execution order if valid
	if response.Valid {
		executionOrder, err := de.generateExecutionOrder(req.Tasks, req.Dependencies)
		if err != nil {
			response.Valid = false
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to generate execution order: %v", err))
		} else {
			response.ExecutionOrder = executionOrder
		}
	}

	// Add warnings for potential issues
	if len(req.Tasks) > 50 {
		response.Warnings = append(response.Warnings, "Large number of tasks may impact performance")
	}

	for _, task := range req.Tasks {
		if task.Timeout > 30*time.Minute {
			response.Warnings = append(response.Warnings, fmt.Sprintf("Task %s has very long timeout: %v", task.ID, task.Timeout))
		}
	}

	logger.WithComponent("dag-engine").Info("DAG validation completed",
		zap.Bool("valid", response.Valid),
		zap.Int("errors", len(response.Errors)),
		zap.Int("warnings", len(response.Warnings)))

	return response, nil
}

// detectCycles uses DFS to detect cycles in the task dependency graph
func (de *DAGEngine) detectCycles(tasks []contracts.Task, dependencies []contracts.TaskDependency) (bool, string) {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = []string{}
	}
	for _, dep := range dependencies {
		graph[dep.From] = append(graph[dep.From], dep.To)
	}

	// Track visit states: 0=unvisited, 1=visiting, 2=visited
	state := make(map[string]int)
	path := []string{}

	var dfs func(string) (bool, string)
	dfs = func(taskID string) (bool, string) {
		if state[taskID] == 1 {
			// Found a back edge - cycle detected
			cycleStart := -1
			for i, id := range path {
				if id == taskID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := append(path[cycleStart:], taskID)
				return true, fmt.Sprintf("%v", cyclePath)
			}
			return true, fmt.Sprintf("cycle involving %s", taskID)
		}

		if state[taskID] == 2 {
			return false, ""
		}

		state[taskID] = 1
		path = append(path, taskID)

		for _, neighbor := range graph[taskID] {
			if hasCycle, cyclePath := dfs(neighbor); hasCycle {
				return true, cyclePath
			}
		}

		state[taskID] = 2
		path = path[:len(path)-1]
		return false, ""
	}

	for _, task := range tasks {
		if state[task.ID] == 0 {
			if hasCycle, cyclePath := dfs(task.ID); hasCycle {
				return true, cyclePath
			}
		}
	}

	return false, ""
}

// generateExecutionOrder creates a topological ordering of tasks
func (de *DAGEngine) generateExecutionOrder(tasks []contracts.Task, dependencies []contracts.TaskDependency) ([]string, error) {
	// Build adjacency list and in-degree count
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, task := range tasks {
		graph[task.ID] = []string{}
		inDegree[task.ID] = 0
	}

	for _, dep := range dependencies {
		graph[dep.From] = append(graph[dep.From], dep.To)
		inDegree[dep.To]++
	}

	// Use Kahn's algorithm for topological sorting
	var queue []string
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(tasks) {
		return nil, fmt.Errorf("cycle detected - cannot generate execution order")
	}

	return result, nil
}

// EstimateExecutionTime estimates total execution time based on task timeouts and dependencies
func (de *DAGEngine) EstimateExecutionTime(tasks []contracts.Task, dependencies []contracts.TaskDependency, maxConcurrency int) time.Duration {
	if len(tasks) == 0 {
		return 0
	}

	// Generate execution order
	executionOrder, err := de.generateExecutionOrder(tasks, dependencies)
	if err != nil {
		// If we can't generate order, return sum of all timeouts
		var total time.Duration
		for _, task := range tasks {
			total += task.Timeout
		}
		return total
	}

	// Build task map for quick lookup
	taskMap := make(map[string]contracts.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Build dependency map
	dependsOn := make(map[string][]string)
	for _, dep := range dependencies {
		dependsOn[dep.To] = append(dependsOn[dep.To], dep.From)
	}

	// Simulate execution with limited concurrency
	taskStartTimes := make(map[string]time.Duration)
	taskEndTimes := make(map[string]time.Duration)
	runningSlots := make([]time.Duration, maxConcurrency)

	for _, taskID := range executionOrder {
		task := taskMap[taskID]

		// Find earliest start time based on dependencies
		var earliestStart time.Duration
		for _, depID := range dependsOn[taskID] {
			if endTime, exists := taskEndTimes[depID]; exists {
				if endTime > earliestStart {
					earliestStart = endTime
				}
			}
		}

		// Find available slot considering concurrency limit
		slotIndex := 0
		availableAt := runningSlots[0]
		for i, slotTime := range runningSlots {
			if slotTime < availableAt {
				availableAt = slotTime
				slotIndex = i
			}
		}

		// Task starts when both dependencies are done and slot is available
		startTime := earliestStart
		if availableAt > startTime {
			startTime = availableAt
		}

		endTime := startTime + task.Timeout
		taskStartTimes[taskID] = startTime
		taskEndTimes[taskID] = endTime
		runningSlots[slotIndex] = endTime
	}

	// Find maximum end time
	var maxEndTime time.Duration
	for _, endTime := range taskEndTimes {
		if endTime > maxEndTime {
			maxEndTime = endTime
		}
	}

	return maxEndTime
}