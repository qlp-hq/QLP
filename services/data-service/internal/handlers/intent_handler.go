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

type IntentHandler struct {
	repo *repository.IntentRepository
}

func NewIntentHandler(repo *repository.IntentRepository) *IntentHandler {
	return &IntentHandler{repo: repo}
}

func (ih *IntentHandler) CreateIntent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	var req contracts.CreateIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("intent-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserInput == "" {
		http.Error(w, "user_input is required", http.StatusBadRequest)
		return
	}

	intent, err := ih.repo.Create(r.Context(), &req, tenantID)
	if err != nil {
		logger.WithComponent("intent-handler").Error("Failed to create intent", zap.Error(err))
		http.Error(w, "Failed to create intent", http.StatusInternalServerError)
		return
	}

	response := contracts.CreateIntentResponse{Intent: *intent}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (ih *IntentHandler) GetIntent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	intentID := chi.URLParam(r, "intentId")

	if tenantID == "" || intentID == "" {
		http.Error(w, "tenant_id and intent_id are required", http.StatusBadRequest)
		return
	}

	intent, err := ih.repo.GetByID(r.Context(), intentID, tenantID)
	if err != nil {
		if err.Error() == "intent not found" {
			http.Error(w, "Intent not found", http.StatusNotFound)
			return
		}
		logger.WithComponent("intent-handler").Error("Failed to get intent", zap.Error(err))
		http.Error(w, "Failed to get intent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(intent)
}

func (ih *IntentHandler) ListIntents(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")

	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	req := &contracts.ListIntentsRequest{
		Status: r.URL.Query().Get("status"),
		Limit:  parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	intents, total, err := ih.repo.List(r.Context(), tenantID, req)
	if err != nil {
		logger.WithComponent("intent-handler").Error("Failed to list intents", zap.Error(err))
		http.Error(w, "Failed to list intents", http.StatusInternalServerError)
		return
	}

	response := contracts.ListIntentsResponse{
		Intents: intents,
		Total:   total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ih *IntentHandler) UpdateIntent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	intentID := chi.URLParam(r, "intentId")

	if tenantID == "" || intentID == "" {
		http.Error(w, "tenant_id and intent_id are required", http.StatusBadRequest)
		return
	}

	var req contracts.UpdateIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("intent-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	intent, err := ih.repo.Update(r.Context(), intentID, tenantID, &req)
	if err != nil {
		if err.Error() == "intent not found" {
			http.Error(w, "Intent not found", http.StatusNotFound)
			return
		}
		logger.WithComponent("intent-handler").Error("Failed to update intent", zap.Error(err))
		http.Error(w, "Failed to update intent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(intent)
}

func (ih *IntentHandler) DeleteIntent(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	intentID := chi.URLParam(r, "intentId")

	if tenantID == "" || intentID == "" {
		http.Error(w, "tenant_id and intent_id are required", http.StatusBadRequest)
		return
	}

	err := ih.repo.Delete(r.Context(), intentID, tenantID)
	if err != nil {
		if err.Error() == "intent not found" {
			http.Error(w, "Intent not found", http.StatusNotFound)
			return
		}
		logger.WithComponent("intent-handler").Error("Failed to delete intent", zap.Error(err))
		http.Error(w, "Failed to delete intent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}