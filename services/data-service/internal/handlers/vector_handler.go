package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/data-service/internal/repository"
	"QLP/services/data-service/pkg/contracts"
)

type VectorHandler struct {
	repo *repository.VectorRepository
}

func NewVectorHandler(repo *repository.VectorRepository) *VectorHandler {
	return &VectorHandler{repo: repo}
}

func (vh *VectorHandler) FindSimilar(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	var req contracts.VectorSimilarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("vector-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set tenant ID from URL
	req.TenantID = tenantID

	if len(req.Embedding) == 0 && req.Query == "" {
		http.Error(w, "either embedding or query is required", http.StatusBadRequest)
		return
	}

	// If query is provided but no embedding, this would typically call an embedding service
	// For now, we'll require the embedding to be provided
	if len(req.Embedding) == 0 {
		http.Error(w, "embedding generation from query not yet implemented", http.StatusNotImplemented)
		return
	}

	response, err := vh.repo.FindSimilar(r.Context(), &req)
	if err != nil {
		logger.WithComponent("vector-handler").Error("Failed to find similar intents", zap.Error(err))
		http.Error(w, "Failed to find similar intents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (vh *VectorHandler) CreateEmbedding(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	var req contracts.EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("vector-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set tenant ID from URL
	req.TenantID = tenantID

	if req.Text == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	// This would typically call an embedding service (OpenAI, etc.)
	// For now, return a placeholder response
	response := contracts.EmbeddingResponse{
		Embedding: make([]float64, 1536), // Placeholder embedding
		Model:     "text-embedding-ada-002",
	}

	logger.WithComponent("vector-handler").Info("Embedding generation requested",
		zap.String("tenant_id", tenantID),
		zap.Int("text_length", len(req.Text)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}