package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/packaging-service/internal/engines"
	"QLP/services/packaging-service/pkg/contracts"
)

// PackagingHandler handles HTTP requests for packaging operations
type PackagingHandler struct {
	capsuleEngine     *engines.CapsuleEngine
	quantumDropEngine *engines.QuantumDropsEngine
	capsuleStorage    map[string]*contracts.QLCapsule // In-memory storage for demo
	dropStorage       map[string]*contracts.QuantumDrop
}

// NewPackagingHandler creates a new packaging handler
func NewPackagingHandler(capsuleEngine *engines.CapsuleEngine, quantumDropEngine *engines.QuantumDropsEngine) *PackagingHandler {
	return &PackagingHandler{
		capsuleEngine:     capsuleEngine,
		quantumDropEngine: quantumDropEngine,
		capsuleStorage:    make(map[string]*contracts.QLCapsule),
		dropStorage:       make(map[string]*contracts.QuantumDrop),
	}
}

// CreateCapsule handles POST /api/v1/tenants/{tenantId}/capsules
func (ph *PackagingHandler) CreateCapsule(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var req contracts.CreateCapsuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("packaging-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.IntentID == "" {
		http.Error(w, "intent_id is required", http.StatusBadRequest)
		return
	}

	logger.WithComponent("packaging-handler").Info("Creating capsule",
		zap.String("tenant_id", tenantID),
		zap.String("intent_id", req.IntentID))

	// Create capsule using the engine
	capsule, err := ph.capsuleEngine.CreateCapsule(r.Context(), tenantID, &req)
	if err != nil {
		logger.WithComponent("packaging-handler").Error("Failed to create capsule", zap.Error(err))
		http.Error(w, "failed to create capsule", http.StatusInternalServerError)
		return
	}

	// Store capsule (in production, this would be in a database)
	ph.capsuleStorage[capsule.Metadata.CapsuleID] = capsule

	// Generate download URL
	downloadURL := fmt.Sprintf("/api/v1/tenants/%s/capsules/%s/download", tenantID, capsule.Metadata.CapsuleID)

	response := contracts.CreateCapsuleResponse{
		CapsuleID:   capsule.Metadata.CapsuleID,
		Status:      "created",
		Message:     "Capsule created successfully",
		Capsule:     capsule,
		DownloadURL: downloadURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CreateQuantumDrops handles POST /api/v1/tenants/{tenantId}/quantum-drops
func (ph *PackagingHandler) CreateQuantumDrops(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var req contracts.CreateQuantumDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithComponent("packaging-handler").Error("Failed to decode request", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.IntentID == "" {
		http.Error(w, "intent_id is required", http.StatusBadRequest)
		return
	}

	logger.WithComponent("packaging-handler").Info("Creating quantum drops",
		zap.String("tenant_id", tenantID),
		zap.String("intent_id", req.IntentID))

	// Create quantum drops using the engine
	drops, err := ph.quantumDropEngine.CreateQuantumDrops(r.Context(), tenantID, &req)
	if err != nil {
		logger.WithComponent("packaging-handler").Error("Failed to create quantum drops", zap.Error(err))
		http.Error(w, "failed to create quantum drops", http.StatusInternalServerError)
		return
	}

	// Store drops (in production, this would be in a database)
	dropID := fmt.Sprintf("batch-%s", req.IntentID)
	for i := range drops {
		ph.dropStorage[drops[i].ID] = &drops[i]
	}

	response := contracts.CreateQuantumDropResponse{
		DropID:  dropID,
		Status:  "created",
		Message: fmt.Sprintf("Created %d quantum drops", len(drops)),
		Drops:   drops,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetCapsule handles GET /api/v1/tenants/{tenantId}/capsules/{capsuleId}
func (ph *PackagingHandler) GetCapsule(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	capsuleID := chi.URLParam(r, "capsuleId")

	if tenantID == "" || capsuleID == "" {
		http.Error(w, "tenant ID and capsule ID are required", http.StatusBadRequest)
		return
	}

	// Retrieve capsule from storage
	capsule, exists := ph.capsuleStorage[capsuleID]
	if !exists {
		http.Error(w, "capsule not found", http.StatusNotFound)
		return
	}

	// Verify tenant access (basic check)
	if capsule.Metadata.TenantID != tenantID {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	response := contracts.GetCapsuleResponse{
		CapsuleID: capsuleID,
		Capsule:   capsule,
		Status:    "available",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListCapsules handles GET /api/v1/tenants/{tenantId}/capsules
func (ph *PackagingHandler) ListCapsules(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Filter capsules by tenant
	var tenantCapsules []contracts.CapsuleMetadata
	for _, capsule := range ph.capsuleStorage {
		if capsule.Metadata.TenantID == tenantID {
			tenantCapsules = append(tenantCapsules, capsule.Metadata)
		}
	}

	// Apply pagination
	total := len(tenantCapsules)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		tenantCapsules = []contracts.CapsuleMetadata{}
	} else {
		if end > total {
			end = total
		}
		tenantCapsules = tenantCapsules[start:end]
	}

	response := contracts.ListCapsulesResponse{
		Capsules: tenantCapsules,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadCapsule handles GET /api/v1/tenants/{tenantId}/capsules/{capsuleId}/download
func (ph *PackagingHandler) DownloadCapsule(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	capsuleID := chi.URLParam(r, "capsuleId")

	if tenantID == "" || capsuleID == "" {
		http.Error(w, "tenant ID and capsule ID are required", http.StatusBadRequest)
		return
	}

	// Retrieve capsule from storage
	capsule, exists := ph.capsuleStorage[capsuleID]
	if !exists {
		http.Error(w, "capsule not found", http.StatusNotFound)
		return
	}

	// Verify tenant access
	if capsule.Metadata.TenantID != tenantID {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	logger.WithComponent("packaging-handler").Info("Packaging capsule for download",
		zap.String("capsule_id", capsuleID))

	// Package capsule into ZIP
	zipData, err := ph.capsuleEngine.PackageCapsule(r.Context(), capsule)
	if err != nil {
		logger.WithComponent("packaging-handler").Error("Failed to package capsule", zap.Error(err))
		http.Error(w, "failed to package capsule", http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("%s.qlcapsule", capsuleID)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(zipData)))

	// Write the ZIP data
	w.WriteHeader(http.StatusOK)
	w.Write(zipData)
}

// GetQuantumDrop handles GET /api/v1/tenants/{tenantId}/quantum-drops/{dropId}
func (ph *PackagingHandler) GetQuantumDrop(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	dropID := chi.URLParam(r, "dropId")

	if tenantID == "" || dropID == "" {
		http.Error(w, "tenant ID and drop ID are required", http.StatusBadRequest)
		return
	}

	// Retrieve drop from storage
	drop, exists := ph.dropStorage[dropID]
	if !exists {
		http.Error(w, "quantum drop not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drop)
}

// HealthCheck handles GET /health
func (ph *PackagingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "packaging-service",
	})
}