package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/validation-service/internal/engines"
	"QLP/services/validation-service/pkg/contracts"
)

type ValidationHandler struct {
	engine *engines.ValidationEngine
}

func NewValidationHandler(engine *engines.ValidationEngine) *ValidationHandler {
	return &ValidationHandler{
		engine: engine,
	}
}

// ValidateContent validates content for a specific tenant
func (h *ValidationHandler) ValidateContent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	var req contracts.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("validation-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	if req.TaskType == "" {
		req.TaskType = contracts.TaskTypeCodegen // Default
	}

	// Start validation
	resp, err := h.engine.ValidateContent(r.Context(), &req, tenantID)
	if err != nil {
		logger.WithComponent("validation-handler").Error("Validation failed", 
			zap.Error(err), 
			zap.String("tenant_id", tenantID))
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// GetValidation retrieves validation results by ID
func (h *ValidationHandler) GetValidation(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	validationID := chi.URLParam(r, "validationId")

	if tenantID == "" || validationID == "" {
		http.Error(w, "Tenant ID and Validation ID are required", http.StatusBadRequest)
		return
	}

	validation, err := h.engine.GetValidation(validationID, tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Validation not found", http.StatusNotFound)
		} else {
			logger.WithComponent("validation-handler").Error("Failed to get validation", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

// ListValidations lists validations for a tenant with filtering
func (h *ValidationHandler) ListValidations(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	req := &contracts.ListValidationsRequest{
		TenantID: tenantID,
		Status:   query.Get("status"),
		Offset:   0,
		Limit:    20, // Default limit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if sinceStr := query.Get("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			req.Since = &since
		}
	}

	resp, err := h.engine.ListValidations(req)
	if err != nil {
		logger.WithComponent("validation-handler").Error("Failed to list validations", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CancelValidation cancels a running validation
func (h *ValidationHandler) CancelValidation(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	validationID := chi.URLParam(r, "validationId")

	if tenantID == "" || validationID == "" {
		http.Error(w, "Tenant ID and Validation ID are required", http.StatusBadRequest)
		return
	}

	// For now, just return not implemented
	// In a full implementation, this would cancel the running validation
	http.Error(w, "Cancel validation not implemented", http.StatusNotImplemented)
}

// ValidateBatch validates multiple items in batch
func (h *ValidationHandler) ValidateBatch(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	var req contracts.BatchValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "No items to validate", http.StatusBadRequest)
		return
	}

	if len(req.Items) > 50 {
		http.Error(w, "Too many items (max 50)", http.StatusBadRequest)
		return
	}

	// Start batch validation
	var responses []contracts.ValidateResponse
	for _, item := range req.Items {
		resp, err := h.engine.ValidateContent(r.Context(), &item, tenantID)
		if err != nil {
			logger.WithComponent("validation-handler").Error("Batch validation item failed", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	batchResp := contracts.BatchValidateResponse{
		BatchID:     fmt.Sprintf("batch_%d", time.Now().Unix()),
		TotalItems:  len(req.Items),
		Validations: responses,
		Status:      "processing",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(batchResp)
}

// GetBatchStatus gets the status of a batch validation
func (h *ValidationHandler) GetBatchStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	batchID := chi.URLParam(r, "batchId")

	if tenantID == "" || batchID == "" {
		http.Error(w, "Tenant ID and Batch ID are required", http.StatusBadRequest)
		return
	}

	// For now, return not implemented
	// In a full implementation, this would track batch validation status
	http.Error(w, "Batch status tracking not implemented", http.StatusNotImplemented)
}

// StreamValidation provides real-time validation updates via Server-Sent Events
func (h *ValidationHandler) StreamValidation(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	validationID := chi.URLParam(r, "validationId")

	if tenantID == "" || validationID == "" {
		http.Error(w, "Tenant ID and Validation ID are required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the validation
	validation, err := h.engine.GetValidation(validationID, tenantID)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Validation not found\"}\n\n")
		return
	}

	// Stream current status
	statusData, _ := json.Marshal(map[string]interface{}{
		"id":           validation.ID,
		"status":       validation.Status,
		"overall_score": validation.OverallScore,
		"passed":       validation.Passed,
		"completed_at": validation.CompletedAt,
	})

	fmt.Fprintf(w, "event: status\ndata: %s\n\n", statusData)

	// If validation is complete, send final results and close
	if validation.Status == contracts.ValidationStatusCompleted ||
		validation.Status == contracts.ValidationStatusFailed ||
		validation.Status == contracts.ValidationStatusTimeout {
		
		resultData, _ := json.Marshal(validation)
		fmt.Fprintf(w, "event: complete\ndata: %s\n\n", resultData)
		return
	}

	// For ongoing validations, we would poll or use a pub/sub mechanism
	// For now, just send periodic updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Get updated validation
			updatedValidation, err := h.engine.GetValidation(validationID, tenantID)
			if err != nil {
				return
			}

			// Send update
			updateData, _ := json.Marshal(map[string]interface{}{
				"id":           updatedValidation.ID,
				"status":       updatedValidation.Status,
				"overall_score": updatedValidation.OverallScore,
				"passed":       updatedValidation.Passed,
			})

			fmt.Fprintf(w, "event: update\ndata: %s\n\n", updateData)

			// Check if completed
			if updatedValidation.Status == contracts.ValidationStatusCompleted ||
				updatedValidation.Status == contracts.ValidationStatusFailed ||
				updatedValidation.Status == contracts.ValidationStatusTimeout {
				
				finalData, _ := json.Marshal(updatedValidation)
				fmt.Fprintf(w, "event: complete\ndata: %s\n\n", finalData)
				return
			}

			// Flush the response
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// HealthCheck returns service health status
func (h *ValidationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "validation-service",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetMetrics returns validation service metrics
func (h *ValidationHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"total_validations": 0,
		"active_validations": 0,
		"average_duration": "0s",
		"success_rate": 0.0,
		"uptime": time.Since(time.Now()).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetValidators returns available validators and their capabilities
func (h *ValidationHandler) GetValidators(w http.ResponseWriter, r *http.Request) {
	validators := map[string]interface{}{
		"syntax_validators": []string{"go", "python", "javascript", "typescript", "hcl", "markdown", "yaml", "json"},
		"security_scanners": []string{"fast", "standard", "comprehensive"},
		"quality_analyzers": []string{"fast", "standard", "comprehensive"},
		"validation_levels": []string{"fast", "standard", "comprehensive"},
		"supported_checks": []string{"syntax", "security", "quality", "performance", "compliance", "llm_critique", "accessibility"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validators)
}

// GetRules returns validation rules and their descriptions
func (h *ValidationHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	rules := map[string]interface{}{
		"syntax_rules": map[string]string{
			"go-package":         "Go files must have package declaration",
			"go-braces":          "Proper brace matching",
			"python-colon":       "Control statements must end with colon",
			"js-semicolon":       "JavaScript statements should end with semicolon",
			"yaml-no-tabs":       "YAML files must use spaces, not tabs",
		},
		"security_rules": map[string]string{
			"secret-exposure":    "No hardcoded secrets or API keys",
			"sql-injection":      "No SQL injection vulnerabilities",
			"xss-prevention":     "Prevent cross-site scripting",
			"weak-crypto":        "Use strong cryptographic algorithms",
		},
		"quality_rules": map[string]string{
			"documentation":      "Adequate code documentation",
			"error-handling":     "Proper error handling",
			"test-coverage":      "Sufficient test coverage",
			"complexity":         "Manageable code complexity",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}