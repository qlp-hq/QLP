package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/worker-runtime/internal/executor"
	"QLP/services/worker-runtime/pkg/contracts"
)

type WorkerHandler struct {
	executor *executor.TaskExecutor
}

func NewWorkerHandler(executor *executor.TaskExecutor) *WorkerHandler {
	return &WorkerHandler{executor: executor}
}

func (wh *WorkerHandler) ExecuteTask(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	var req contracts.ExecuteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("worker-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set tenant ID from URL
	req.Task.TenantID = tenantID

	// Validate request
	if req.Task.ID == "" {
		http.Error(w, "task.id is required", http.StatusBadRequest)
		return
	}
	if req.Task.Type == "" {
		http.Error(w, "task.type is required", http.StatusBadRequest)
		return
	}
	if req.Task.Description == "" {
		http.Error(w, "task.description is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Task.Priority == "" {
		req.Task.Priority = contracts.PriorityMedium
	}
	if req.Task.TimeoutSeconds == 0 {
		req.Task.TimeoutSeconds = 300 // 5 minutes default
	}
	if req.Task.Metadata == nil {
		req.Task.Metadata = make(map[string]string)
	}

	// Execute task
	response, err := wh.executor.ExecuteTask(r.Context(), &req)
	if err != nil {
		logger.WithComponent("worker-handler").Error("Failed to execute task",
			zap.String("tenant_id", tenantID),
			zap.String("task_id", req.Task.ID),
			zap.Error(err))
		http.Error(w, "Failed to execute task", http.StatusInternalServerError)
		return
	}

	logger.WithComponent("worker-handler").Info("Task execution started",
		zap.String("tenant_id", tenantID),
		zap.String("task_id", req.Task.ID),
		zap.String("execution_id", response.ExecutionID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (wh *WorkerHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	executionID := chi.URLParam(r, "executionId")

	if tenantID == "" || executionID == "" {
		http.Error(w, "tenant_id and execution_id are required", http.StatusBadRequest)
		return
	}

	execution, err := wh.executor.GetExecution(executionID, tenantID)
	if err != nil {
		if err.Error() == "execution not found" || err.Error() == "execution not found for tenant" {
			http.Error(w, "Execution not found", http.StatusNotFound)
			return
		}
		logger.WithComponent("worker-handler").Error("Failed to get execution",
			zap.String("execution_id", executionID),
			zap.Error(err))
		http.Error(w, "Failed to get execution", http.StatusInternalServerError)
		return
	}

	response := contracts.GetExecutionResponse{Execution: *execution}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (wh *WorkerHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	req := &contracts.ListExecutionsRequest{
		TenantID: tenantID,
		Limit:    parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset:   parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	// Parse status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		req.Status = contracts.ExecutionStatus(statusStr)
	}

	// Parse task type filter
	if taskTypeStr := r.URL.Query().Get("task_type"); taskTypeStr != "" {
		req.TaskType = contracts.TaskType(taskTypeStr)
	}

	// Parse since filter
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			req.Since = &since
		}
	}

	response, err := wh.executor.ListExecutions(req)
	if err != nil {
		logger.WithComponent("worker-handler").Error("Failed to list executions",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		http.Error(w, "Failed to list executions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (wh *WorkerHandler) CancelExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	executionID := chi.URLParam(r, "executionId")

	if tenantID == "" || executionID == "" {
		http.Error(w, "tenant_id and execution_id are required", http.StatusBadRequest)
		return
	}

	err := wh.executor.CancelExecution(executionID, tenantID)
	if err != nil {
		if err.Error() == "execution not found" || err.Error() == "execution not found for tenant" {
			http.Error(w, "Execution not found", http.StatusNotFound)
			return
		}
		logger.WithComponent("worker-handler").Error("Failed to cancel execution",
			zap.String("execution_id", executionID),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := contracts.CancelExecutionResponse{
		Success: true,
		Message: "Execution canceled successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (wh *WorkerHandler) StreamExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	executionID := chi.URLParam(r, "executionId")

	if tenantID == "" || executionID == "" {
		http.Error(w, "tenant_id and execution_id are required", http.StatusBadRequest)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial connection event
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"execution_id\":\"%s\"}\n\n", executionID)
	w.(http.Flusher).Flush()

	// Mock streaming - in real implementation, this would stream actual updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 10; i++ {
		select {
		case <-ticker.C:
			update := contracts.ExecutionUpdate{
				ExecutionID: executionID,
				Status:      contracts.ExecutionStatusRunning,
				Output:      fmt.Sprintf("Processing step %d...", i+1),
				Timestamp:   time.Now(),
				ProgressPct: (i + 1) * 10,
			}

			updateJSON, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", updateJSON)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			return
		}
	}

	// Send completion event
	finalUpdate := contracts.ExecutionUpdate{
		ExecutionID: executionID,
		Status:      contracts.ExecutionStatusCompleted,
		Output:      "Task completed successfully",
		Timestamp:   time.Now(),
		ProgressPct: 100,
	}
	updateJSON, _ := json.Marshal(finalUpdate)
	fmt.Fprintf(w, "data: %s\n\n", updateJSON)
	w.(http.Flusher).Flush()
}

func (wh *WorkerHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "worker-runtime",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (wh *WorkerHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// Mock metrics - in real implementation, this would return actual metrics
	metrics := map[string]interface{}{
		"executions_total":      150,
		"executions_running":    5,
		"executions_completed":  140,
		"executions_failed":     5,
		"avg_execution_time_ms": 2500,
		"cpu_usage_percent":     45.2,
		"memory_usage_mb":       512,
		"active_containers":     3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (wh *WorkerHandler) GetAgentTypes(w http.ResponseWriter, r *http.Request) {
	agentTypes := []map[string]interface{}{
		{
			"type":         "codegen",
			"description":  "Generates code based on natural language descriptions",
			"capabilities": []string{"go", "python", "javascript", "typescript"},
		},
		{
			"type":         "infra",
			"description":  "Creates infrastructure as code",
			"capabilities": []string{"terraform", "arm", "bicep", "cloudformation"},
		},
		{
			"type":         "doc",
			"description":  "Generates documentation",
			"capabilities": []string{"markdown", "rst", "asciidoc"},
		},
		{
			"type":         "test",
			"description":  "Creates test cases and test suites",
			"capabilities": []string{"go_test", "pytest", "jest", "mocha"},
		},
		{
			"type":         "analyze",
			"description":  "Performs code analysis and quality assessment",
			"capabilities": []string{"static_analysis", "security_scan", "performance_analysis"},
		},
		{
			"type":         "validate",
			"description":  "Validates code quality and security",
			"capabilities": []string{"syntax", "security", "quality"},
		},
		{
			"type":         "package",
			"description":  "Packages code into deployable artifacts",
			"capabilities": []string{"qlp_capsule", "docker", "zip"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_types": agentTypes,
		"total":       len(agentTypes),
	})
}

// Helper function
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	
	return value
}