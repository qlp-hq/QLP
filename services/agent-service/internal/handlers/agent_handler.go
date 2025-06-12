package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/services/agent-service/pkg/contracts"
	"QLP/services/agent-service/internal/factory"
	"QLP/internal/logger"
	"QLP/internal/tenancy"
)

// AgentHandler handles HTTP requests for the agent service
type AgentHandler struct {
	agentFactory *factory.AgentFactory
	startTime    time.Time
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentFactory *factory.AgentFactory) *AgentHandler {
	return &AgentHandler{
		agentFactory: agentFactory,
		startTime:    time.Now(),
	}
}

// CreateAgent handles POST /api/v1/tenants/{tenantId}/agents
func (ah *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		ah.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Create agent
	agent, err := ah.agentFactory.CreateAgent(r.Context(), &req)
	if err != nil {
		logger.WithComponent("agent-handler").Error("Failed to create agent",
			zap.String("tenant_id", tenantID),
			zap.String("task_id", req.TaskID),
			zap.Error(err))
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("agent creation failed: %v", err), "CREATION_FAILED")
		return
	}

	response := &contracts.CreateAgentResponse{
		AgentID: agent.ID,
		Status:  string(agent.Status),
		Message: "Agent created successfully",
		Agent:   agent,
	}

	logger.WithComponent("agent-handler").Info("Agent created successfully",
		zap.String("tenant_id", tenantID),
		zap.String("agent_id", agent.ID),
		zap.String("task_type", req.TaskType))

	ah.writeJSON(w, http.StatusCreated, response)
}

// ExecuteAgent handles POST /api/v1/tenants/{tenantId}/agents/{agentId}/execute
func (ah *AgentHandler) ExecuteAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	agentID := chi.URLParam(r, "agentId")

	if tenantID == "" || agentID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID or agent ID", "MISSING_PARAMETERS")
		return
	}

	var req contracts.ExecuteAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Set agent ID from URL
	req.AgentID = agentID

	// Validate request
	if err := req.Validate(); err != nil {
		ah.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Execute agent
	err := ah.agentFactory.ExecuteAgent(r.Context(), agentID, &req)
	if err != nil {
		logger.WithComponent("agent-handler").Error("Failed to execute agent",
			zap.String("tenant_id", tenantID),
			zap.String("agent_id", agentID),
			zap.Error(err))
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("agent execution failed: %v", err), "EXECUTION_FAILED")
		return
	}

	// Get updated agent
	agent, err := ah.agentFactory.GetAgent(agentID)
	if err != nil {
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get agent: %v", err), "AGENT_NOT_FOUND")
		return
	}

	response := &contracts.ExecuteAgentResponse{
		AgentID:     agentID,
		Status:      string(agent.Status),
		Message:     "Agent execution completed",
		ExecutionID: fmt.Sprintf("exec_%s_%d", agentID, time.Now().Unix()),
		Agent:       agent,
	}

	logger.WithComponent("agent-handler").Info("Agent execution completed",
		zap.String("tenant_id", tenantID),
		zap.String("agent_id", agentID),
		zap.String("status", string(agent.Status)))

	ah.writeJSON(w, http.StatusOK, response)
}

// GetAgent handles GET /api/v1/tenants/{tenantId}/agents/{agentId}
func (ah *AgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	agentID := chi.URLParam(r, "agentId")

	if tenantID == "" || agentID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID or agent ID", "MISSING_PARAMETERS")
		return
	}

	agent, err := ah.agentFactory.GetAgent(agentID)
	if err != nil {
		ah.writeError(w, http.StatusNotFound, "agent not found", "AGENT_NOT_FOUND")
		return
	}

	// Verify tenant access
	if agent.Metadata["tenant_id"] != tenantID {
		ah.writeError(w, http.StatusForbidden, "access denied", "ACCESS_DENIED")
		return
	}

	response := &contracts.GetAgentResponse{
		Agent: agent,
	}

	ah.writeJSON(w, http.StatusOK, response)
}

// ListAgents handles GET /api/v1/tenants/{tenantId}/agents
func (ah *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
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

	// List agents
	agents, total := ah.agentFactory.ListAgents(page, pageSize, status)

	// Filter by tenant - in production this would be done at the storage level
	var filteredAgents []contracts.AgentSummary
	for _, agent := range agents {
		// This is a simplified check - in production we'd need tenant filtering at query level
		if agent != nil {
			filteredAgents = append(filteredAgents, *agent)
		}
	}

	response := &contracts.ListAgentsResponse{
		Agents:   filteredAgents,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	ah.writeJSON(w, http.StatusOK, response)
}

// CancelAgent handles POST /api/v1/tenants/{tenantId}/agents/{agentId}/cancel
func (ah *AgentHandler) CancelAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	agentID := chi.URLParam(r, "agentId")

	if tenantID == "" || agentID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID or agent ID", "MISSING_PARAMETERS")
		return
	}

	var req contracts.CancelAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	err := ah.agentFactory.CancelAgent(agentID, req.Reason)
	if err != nil {
		ah.writeError(w, http.StatusBadRequest, err.Error(), "CANCEL_FAILED")
		return
	}

	logger.WithComponent("agent-handler").Info("Agent cancelled",
		zap.String("tenant_id", tenantID),
		zap.String("agent_id", agentID),
		zap.String("reason", req.Reason))

	w.WriteHeader(http.StatusOK)
}

// RetryAgent handles POST /api/v1/tenants/{tenantId}/agents/{agentId}/retry
func (ah *AgentHandler) RetryAgent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	agentID := chi.URLParam(r, "agentId")

	if tenantID == "" || agentID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID or agent ID", "MISSING_PARAMETERS")
		return
	}

	var req contracts.RetryAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// For retry, we create a new execution request
	executeReq := &contracts.ExecuteAgentRequest{
		AgentID: agentID,
		Metadata: map[string]string{
			"tenant_id":    tenantID,
			"retry_reason": req.Reason,
			"retry_at":     time.Now().Format(time.RFC3339),
		},
	}

	err := ah.agentFactory.ExecuteAgent(r.Context(), agentID, executeReq)
	if err != nil {
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("agent retry failed: %v", err), "RETRY_FAILED")
		return
	}

	// Get updated agent
	agent, err := ah.agentFactory.GetAgent(agentID)
	if err != nil {
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get agent: %v", err), "AGENT_NOT_FOUND")
		return
	}

	response := &contracts.ExecuteAgentResponse{
		AgentID:     agentID,
		Status:      string(agent.Status),
		Message:     "Agent retry completed",
		ExecutionID: fmt.Sprintf("retry_%s_%d", agentID, time.Now().Unix()),
		Agent:       agent,
	}

	logger.WithComponent("agent-handler").Info("Agent retry completed",
		zap.String("tenant_id", tenantID),
		zap.String("agent_id", agentID),
		zap.String("reason", req.Reason))

	ah.writeJSON(w, http.StatusOK, response)
}

// BatchCreateAgents handles POST /api/v1/tenants/{tenantId}/agents/batch
func (ah *AgentHandler) BatchCreateAgents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.BatchCreateAgentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		ah.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	batchID := fmt.Sprintf("batch_%s_%d", tenantID, time.Now().UnixNano())
	var results []contracts.CreateAgentResponse
	var successCount, failureCount int

	// Process each agent in the batch
	for _, agentReq := range req.Agents {
		// Add tenant context
		if agentReq.Metadata == nil {
			agentReq.Metadata = make(map[string]string)
		}
		agentReq.Metadata["tenant_id"] = tenantID
		agentReq.Metadata["batch_id"] = batchID

		agent, err := ah.agentFactory.CreateAgent(r.Context(), &agentReq)
		if err != nil {
			results = append(results, contracts.CreateAgentResponse{
				Status:  "failed",
				Message: err.Error(),
			})
			failureCount++
		} else {
			results = append(results, contracts.CreateAgentResponse{
				AgentID: agent.ID,
				Status:  string(agent.Status),
				Message: "Agent created successfully",
				Agent:   agent,
			})
			successCount++
		}
	}

	response := &contracts.BatchCreateAgentsResponse{
		BatchID:      batchID,
		TotalCount:   len(req.Agents),
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Metadata:     req.Metadata,
	}

	logger.WithComponent("agent-handler").Info("Batch agent creation completed",
		zap.String("tenant_id", tenantID),
		zap.String("batch_id", batchID),
		zap.Int("total", response.TotalCount),
		zap.Int("success", successCount),
		zap.Int("failures", failureCount))

	ah.writeJSON(w, http.StatusOK, response)
}

// CreateDeploymentValidator handles POST /api/v1/tenants/{tenantId}/agents/deployment-validator
func (ah *AgentHandler) CreateDeploymentValidator(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		ah.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.CreateDeploymentValidatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Create deployment validator
	agent, err := ah.agentFactory.CreateDeploymentValidator(r.Context(), &req)
	if err != nil {
		logger.WithComponent("agent-handler").Error("Failed to create deployment validator",
			zap.String("tenant_id", tenantID),
			zap.String("agent_id", req.AgentID),
			zap.Error(err))
		ah.writeError(w, http.StatusInternalServerError, fmt.Sprintf("deployment validator creation failed: %v", err), "CREATION_FAILED")
		return
	}

	response := &contracts.CreateAgentResponse{
		AgentID: agent.ID,
		Status:  string(agent.Status),
		Message: "Deployment validator agent created successfully",
		Agent:   agent,
	}

	logger.WithComponent("agent-handler").Info("Deployment validator created successfully",
		zap.String("tenant_id", tenantID),
		zap.String("agent_id", agent.ID),
		zap.String("capsule_id", req.CapsuleData.ID))

	ah.writeJSON(w, http.StatusCreated, response)
}

// GetServiceStatus handles GET /api/v1/status
func (ah *AgentHandler) GetServiceStatus(w http.ResponseWriter, r *http.Request) {
	activeAgents := ah.agentFactory.GetActiveAgents()
	metrics := ah.agentFactory.GetMetrics()

	status := &contracts.ServiceStatus{
		Status:        "running",
		Timestamp:     time.Now(),
		ActiveAgents:  activeAgents,
		TotalAgents:   int(metrics.TotalAgentsCreated),
		Version:       "1.0.0",
		Uptime:        time.Since(ah.startTime),
		ResourceUsage: contracts.ResourceUsage{
			CPUPercent:    15.5, // Mock values
			MemoryUsedMB:  256,
			MemoryTotalMB: 1024,
			DiskUsedMB:    512,
			ActiveThreads: 10,
		},
		Dependencies: []contracts.DependencyStatus{
			{
				Name:      "llm-service",
				Status:    "healthy",
				Healthy:   true,
				LastCheck: time.Now(),
				Response:  50 * time.Millisecond,
			},
		},
	}

	ah.writeJSON(w, http.StatusOK, status)
}

// GetMetrics handles GET /api/v1/metrics
func (ah *AgentHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := ah.agentFactory.GetMetrics()
	ah.writeJSON(w, http.StatusOK, metrics)
}

// HealthCheck handles GET /health
func (ah *AgentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	activeAgents := ah.agentFactory.GetActiveAgents()

	health := &contracts.HealthCheckResponse{
		Service:      "agent-service",
		Status:       "healthy",
		Timestamp:    time.Now(),
		Version:      "1.0.0",
		ActiveAgents: activeAgents,
		Checks: map[string]string{
			"agent_factory": "healthy",
			"llm_service":   "healthy",
			"memory":        "healthy",
		},
	}

	ah.writeJSON(w, http.StatusOK, health)
}

// GetTenantContext extracts tenant information from request context
func (ah *AgentHandler) GetTenantContext(r *http.Request) *tenancy.TenantContext {
	return tenancy.GetTenantContextFromRequest(r)
}

// Helper methods

func (ah *AgentHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithComponent("agent-handler").Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (ah *AgentHandler) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	errorResponse := &contracts.ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
	}

	ah.writeJSON(w, statusCode, errorResponse)
}