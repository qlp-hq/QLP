package handler

import (
	"QLP/services/prompt-service/internal/models"
	"QLP/services/prompt-service/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PromptHandler struct {
	repo *repository.PromptRepository
}

func NewPromptHandler(repo *repository.PromptRepository) *PromptHandler {
	return &PromptHandler{repo: repo}
}

func (h *PromptHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/prompts", h.CreatePrompt).Methods("POST")
	router.HandleFunc("/prompts", h.ListPrompts).Methods("GET")
	router.HandleFunc("/prompts/{id}", h.GetPrompt).Methods("GET")
	router.HandleFunc("/prompts/{id}", h.UpdatePrompt).Methods("PUT")
	router.HandleFunc("/prompts/{id}", h.DeactivatePrompt).Methods("DELETE")
}

func (h *PromptHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	var prompt models.Prompt
	if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.repo.Create(&prompt); err != nil {
		http.Error(w, "Failed to create prompt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prompt)
}

func (h *PromptHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	prompt, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Prompt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

func (h *PromptHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		PromptText string `json:"prompt_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	prompt, err := h.repo.Update(id, reqBody.PromptText)
	if err != nil {
		http.Error(w, "Failed to update prompt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

func (h *PromptHandler) DeactivatePrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Deactivate(id); err != nil {
		http.Error(w, "Failed to deactivate prompt", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PromptHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	taskType := r.URL.Query().Get("task_type")
	isActive, _ := strconv.ParseBool(r.URL.Query().Get("active"))

	if limit == 0 {
		limit = 20
	}

	if taskType != "" && isActive {
		prompts, err := h.repo.GetByTaskType(taskType)
		if err != nil {
			http.Error(w, "Failed to get prompts by task type", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prompts)
		return
	}

	prompts, err := h.repo.List(limit, offset)
	if err != nil {
		http.Error(w, "Failed to list prompts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}
