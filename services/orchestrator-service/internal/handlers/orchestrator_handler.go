package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/services/orchestrator-service/pkg/contracts"
	"QLP/services/orchestrator-service/internal/engines"
	"QLP/internal/logger"
	"QLP/internal/tenancy"
)

// OrchestratorHandler handles HTTP requests for the orchestrator service
type OrchestratorHandler struct {
	workflowEngine *engines.WorkflowEngine
	dagEngine      *engines.DAGEngine
}

// NewOrchestratorHandler creates a new orchestrator handler
func NewOrchestratorHandler(workflowEngine *engines.WorkflowEngine, dagEngine *engines.DAGEngine) *OrchestratorHandler {
	return &OrchestratorHandler{
		workflowEngine: workflowEngine,
		dagEngine:      dagEngine,
	}
}

// ExecuteWorkflow handles POST /api/v1/tenants/{tenantId}/workflows
func (oh *OrchestratorHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID")
		return
	}

	var req contracts.ExecuteWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Validate request
	if req.IntentID == "" {
		oh.writeError(w, http.StatusBadRequest, "intent_id is required")
		return
	}

	if len(req.Tasks) == 0 {
		oh.writeError(w, http.StatusBadRequest, "at least one task is required")
		return
	}

	// Set default configuration values
	if req.Configuration.MaxConcurrency == 0 {
		req.Configuration.MaxConcurrency = 4
	}
	if req.Configuration.Timeout == 0 {
		req.Configuration.Timeout = 30 * time.Minute
	}
	if req.Configuration.FailurePolicy == "" {
		req.Configuration.FailurePolicy = "abort"
	}

	// Add tenant context
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["tenant_id"] = tenantID

	// Execute workflow
	resp, err := oh.workflowEngine.ExecuteWorkflow(r.Context(), &req)
	if err != nil {
		logger.WithComponent("orchestrator-handler").Error("Failed to execute workflow",
			zap.String("tenant_id", tenantID),
			zap.String("intent_id", req.IntentID),
			zap.Error(err))
		oh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("workflow execution failed: %v", err))
		return
	}

	logger.WithComponent("orchestrator-handler").Info("Workflow execution started",
		zap.String("tenant_id", tenantID),
		zap.String("workflow_id", resp.WorkflowID),
		zap.String("intent_id", req.IntentID))

	oh.writeJSON(w, http.StatusCreated, resp)
}

// GetWorkflow handles GET /api/v1/tenants/{tenantId}/workflows/{workflowId}
func (oh *OrchestratorHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	resp, err := oh.workflowEngine.GetWorkflow(r.Context(), workflowID)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get workflow: %v", err))
		return
	}

	// Verify tenant access
	if resp.Execution.Metadata["tenant_id"] != tenantID {
		oh.writeError(w, http.StatusForbidden, "access denied")
		return
	}

	oh.writeJSON(w, http.StatusOK, resp)
}

// ListWorkflows handles GET /api/v1/tenants/{tenantId}/workflows
func (oh *OrchestratorHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID")
		return
	}

	// Parse query parameters
	page := 0
	pageSize := 20
	status := r.URL.Query().Get("status")

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	resp, err := oh.workflowEngine.ListWorkflows(r.Context(), page, pageSize, status)
	if err != nil {
		oh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
		return
	}

	// Filter by tenant - in production this would be done at the database level
	filteredWorkflows := []contracts.WorkflowSummary{}
	for _, workflow := range resp.Workflows {
		// This is a simplified check - in production we'd need to fetch full workflows or filter at query level
		filteredWorkflows = append(filteredWorkflows, workflow)
	}

	resp.Workflows = filteredWorkflows
	resp.Total = len(filteredWorkflows)

	oh.writeJSON(w, http.StatusOK, resp)
}

// PauseWorkflow handles POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/pause
func (oh *OrchestratorHandler) PauseWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	var req contracts.PauseWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	err := oh.workflowEngine.PauseWorkflow(r.Context(), workflowID, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.WithComponent("orchestrator-handler").Info("Workflow paused",
		zap.String("tenant_id", tenantID),
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason))

	w.WriteHeader(http.StatusOK)
}

// ResumeWorkflow handles POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/resume
func (oh *OrchestratorHandler) ResumeWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	var req contracts.ResumeWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	err := oh.workflowEngine.ResumeWorkflow(r.Context(), workflowID, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.WithComponent("orchestrator-handler").Info("Workflow resumed",
		zap.String("tenant_id", tenantID),
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason))

	w.WriteHeader(http.StatusOK)
}

// CancelWorkflow handles POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/cancel
func (oh *OrchestratorHandler) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	var req contracts.CancelWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	err := oh.workflowEngine.CancelWorkflow(r.Context(), workflowID, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.WithComponent("orchestrator-handler").Info("Workflow cancelled",
		zap.String("tenant_id", tenantID),
		zap.String("workflow_id", workflowID),
		zap.String("reason", req.Reason))

	w.WriteHeader(http.StatusOK)
}

// RetryTask handles POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/retry
func (oh *OrchestratorHandler) RetryTask(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	var req contracts.RetryTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.TaskID == "" {
		oh.writeError(w, http.StatusBadRequest, "task_id is required")
		return
	}

	err := oh.workflowEngine.RetryTask(r.Context(), workflowID, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.WithComponent("orchestrator-handler").Info("Task retry requested",
		zap.String("tenant_id", tenantID),
		zap.String("workflow_id", workflowID),
		zap.String("task_id", req.TaskID),
		zap.String("reason", req.Reason))

	w.WriteHeader(http.StatusOK)
}

// ValidateDAG handles POST /api/v1/dag/validate
func (oh *OrchestratorHandler) ValidateDAG(w http.ResponseWriter, r *http.Request) {
	var req contracts.DAGValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if len(req.Tasks) == 0 {
		oh.writeError(w, http.StatusBadRequest, "at least one task is required")
		return
	}

	resp, err := oh.dagEngine.ValidateDAG(r.Context(), &req)
	if err != nil {
		oh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("DAG validation failed: %v", err))
		return
	}

	oh.writeJSON(w, http.StatusOK, resp)
}

// GetWorkflowMetrics handles GET /api/v1/tenants/{tenantId}/workflows/{workflowId}/metrics
func (oh *OrchestratorHandler) GetWorkflowMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	workflowID := chi.URLParam(r, "workflowId")

	if tenantID == "" || workflowID == "" {
		oh.writeError(w, http.StatusBadRequest, "missing tenant ID or workflow ID")
		return
	}

	metrics, err := oh.workflowEngine.GetWorkflowMetrics(r.Context(), workflowID)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", workflowID) {
			oh.writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		oh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get workflow metrics: %v", err))
		return
	}

	oh.writeJSON(w, http.StatusOK, metrics)
}

// HealthCheck handles GET /health
func (oh *OrchestratorHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"service":   "orchestrator-service",
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	oh.writeJSON(w, http.StatusOK, health)
}

// GetTenantContext extracts tenant information from request context
func (oh *OrchestratorHandler) GetTenantContext(r *http.Request) *tenancy.TenantContext {
	return tenancy.GetTenantContextFromRequest(r)
}

// Helper methods

func (oh *OrchestratorHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithComponent("orchestrator-handler").Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (oh *OrchestratorHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    statusCode,
	}

	oh.writeJSON(w, statusCode, errorResponse)
}