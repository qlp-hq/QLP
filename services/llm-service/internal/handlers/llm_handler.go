package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/services/llm-service/pkg/contracts"
	"QLP/services/llm-service/internal/providers"
	"QLP/internal/logger"
	"QLP/internal/tenancy"
)

// LLMHandler handles HTTP requests for the LLM service
type LLMHandler struct {
	providerManager *providers.ProviderManager
	startTime       time.Time
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(providerManager *providers.ProviderManager) *LLMHandler {
	return &LLMHandler{
		providerManager: providerManager,
		startTime:       time.Now(),
	}
}

// Complete handles POST /api/v1/tenants/{tenantId}/completion
func (lh *LLMHandler) Complete(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		lh.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		lh.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Process completion
	resp, err := lh.providerManager.Complete(r.Context(), &req)
	if err != nil {
		logger.WithComponent("llm-handler").Error("Completion failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		lh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("completion failed: %v", err), "COMPLETION_FAILED")
		return
	}

	logger.WithComponent("llm-handler").Info("Completion successful",
		zap.String("tenant_id", tenantID),
		zap.String("provider", resp.Provider),
		zap.Duration("response_time", resp.ResponseTime),
		zap.Int("tokens", resp.Usage.TotalTokens))

	lh.writeJSON(w, http.StatusOK, resp)
}

// GenerateEmbedding handles POST /api/v1/tenants/{tenantId}/embedding
func (lh *LLMHandler) GenerateEmbedding(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		lh.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		lh.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Process embedding
	resp, err := lh.providerManager.GenerateEmbedding(r.Context(), &req)
	if err != nil {
		logger.WithComponent("llm-handler").Error("Embedding generation failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		lh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("embedding failed: %v", err), "EMBEDDING_FAILED")
		return
	}

	logger.WithComponent("llm-handler").Info("Embedding generation successful",
		zap.String("tenant_id", tenantID),
		zap.String("provider", resp.Provider),
		zap.Duration("response_time", resp.ResponseTime),
		zap.Int("dimensions", resp.Dimensions))

	lh.writeJSON(w, http.StatusOK, resp)
}

// ChatCompletion handles POST /api/v1/tenants/{tenantId}/chat/completion
func (lh *LLMHandler) ChatCompletion(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		lh.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		lh.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Add tenant context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["tenant_id"] = tenantID

	// Process chat completion
	resp, err := lh.providerManager.ChatCompletion(r.Context(), &req)
	if err != nil {
		logger.WithComponent("llm-handler").Error("Chat completion failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		lh.writeError(w, http.StatusInternalServerError, fmt.Sprintf("chat completion failed: %v", err), "CHAT_COMPLETION_FAILED")
		return
	}

	logger.WithComponent("llm-handler").Info("Chat completion successful",
		zap.String("tenant_id", tenantID),
		zap.String("provider", resp.Provider),
		zap.Duration("response_time", resp.ResponseTime),
		zap.Int("tokens", resp.Usage.TotalTokens))

	lh.writeJSON(w, http.StatusOK, resp)
}

// BatchProcess handles POST /api/v1/tenants/{tenantId}/batch
func (lh *LLMHandler) BatchProcess(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		lh.writeError(w, http.StatusBadRequest, "missing tenant ID", "MISSING_TENANT_ID")
		return
	}

	var req contracts.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), "INVALID_REQUEST")
		return
	}

	if len(req.Requests) == 0 {
		lh.writeError(w, http.StatusBadRequest, "no requests in batch", "EMPTY_BATCH")
		return
	}

	if len(req.Requests) > 100 {
		lh.writeError(w, http.StatusBadRequest, "batch size exceeds limit of 100", "BATCH_TOO_LARGE")
		return
	}

	startTime := time.Now()
	batchID := fmt.Sprintf("batch_%s_%d", tenantID, startTime.UnixNano())
	
	var responses []interface{}
	var successCount, failureCount int

	// Process each request in the batch
	for _, reqInterface := range req.Requests {
		switch req.Type {
		case "completion":
			// Convert to completion request
			reqBytes, _ := json.Marshal(reqInterface)
			var completionReq contracts.CompletionRequest
			if err := json.Unmarshal(reqBytes, &completionReq); err != nil {
				responses = append(responses, map[string]string{"error": "invalid completion request"})
				failureCount++
				continue
			}

			resp, err := lh.providerManager.Complete(r.Context(), &completionReq)
			if err != nil {
				responses = append(responses, map[string]string{"error": err.Error()})
				failureCount++
			} else {
				responses = append(responses, resp)
				successCount++
			}

		case "embedding":
			// Convert to embedding request
			reqBytes, _ := json.Marshal(reqInterface)
			var embeddingReq contracts.EmbeddingRequest
			if err := json.Unmarshal(reqBytes, &embeddingReq); err != nil {
				responses = append(responses, map[string]string{"error": "invalid embedding request"})
				failureCount++
				continue
			}

			resp, err := lh.providerManager.GenerateEmbedding(r.Context(), &embeddingReq)
			if err != nil {
				responses = append(responses, map[string]string{"error": err.Error()})
				failureCount++
			} else {
				responses = append(responses, resp)
				successCount++
			}

		case "chat":
			// Convert to chat completion request
			reqBytes, _ := json.Marshal(reqInterface)
			var chatReq contracts.ChatCompletionRequest
			if err := json.Unmarshal(reqBytes, &chatReq); err != nil {
				responses = append(responses, map[string]string{"error": "invalid chat request"})
				failureCount++
				continue
			}

			resp, err := lh.providerManager.ChatCompletion(r.Context(), &chatReq)
			if err != nil {
				responses = append(responses, map[string]string{"error": err.Error()})
				failureCount++
			} else {
				responses = append(responses, resp)
				successCount++
			}

		default:
			responses = append(responses, map[string]string{"error": "unsupported request type"})
			failureCount++
		}
	}

	processingTime := time.Since(startTime)

	batchResp := &contracts.BatchResponse{
		Responses:      responses,
		BatchID:        batchID,
		TotalCount:     len(req.Requests),
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		ProcessingTime: processingTime,
		Metadata:       req.Metadata,
	}

	logger.WithComponent("llm-handler").Info("Batch processing completed",
		zap.String("tenant_id", tenantID),
		zap.String("batch_id", batchID),
		zap.Int("total", batchResp.TotalCount),
		zap.Int("success", successCount),
		zap.Int("failures", failureCount),
		zap.Duration("processing_time", processingTime))

	lh.writeJSON(w, http.StatusOK, batchResp)
}

// ListProviders handles GET /api/v1/providers
func (lh *LLMHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	statuses := lh.providerManager.GetProviderStatus()

	resp := &contracts.ListProvidersResponse{
		Providers: statuses,
		Total:     len(statuses),
	}

	lh.writeJSON(w, http.StatusOK, resp)
}

// GetServiceStatus handles GET /api/v1/status
func (lh *LLMHandler) GetServiceStatus(w http.ResponseWriter, r *http.Request) {
	statuses := lh.providerManager.GetProviderStatus()
	activeModel := lh.providerManager.GetActiveModel()

	status := &contracts.ServiceStatus{
		Status:      "running",
		Timestamp:   time.Now(),
		Providers:   statuses,
		ActiveModel: activeModel,
		Version:     "1.0.0",
	}

	lh.writeJSON(w, http.StatusOK, status)
}

// GetMetrics handles GET /api/v1/metrics
func (lh *LLMHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := lh.providerManager.GetMetrics()
	lh.writeJSON(w, http.StatusOK, metrics)
}

// HealthCheck handles GET /health
func (lh *LLMHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	statuses := lh.providerManager.GetProviderStatus()
	
	checks := make(map[string]string)
	hasHealthyProvider := false
	
	for _, status := range statuses {
		if status.Available && status.Healthy {
			hasHealthyProvider = true
			checks[status.Name] = "healthy"
		} else {
			checks[status.Name] = "unhealthy"
		}
	}

	serviceStatus := "unhealthy"
	if hasHealthyProvider {
		serviceStatus = "healthy"
	}

	health := &contracts.HealthCheckResponse{
		Service:   "llm-service",
		Status:    serviceStatus,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if serviceStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	lh.writeJSON(w, statusCode, health)
}

// GetTenantContext extracts tenant information from request context
func (lh *LLMHandler) GetTenantContext(r *http.Request) *tenancy.TenantContext {
	return tenancy.GetTenantContextFromRequest(r)
}

// Helper methods

func (lh *LLMHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithComponent("llm-handler").Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (lh *LLMHandler) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	errorResponse := &contracts.ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
	}

	lh.writeJSON(w, statusCode, errorResponse)
}