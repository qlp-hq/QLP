package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"QLP/services/orchestrator-service/internal/handlers"
	"QLP/services/orchestrator-service/internal/engines"
	"QLP/internal/config"
	"QLP/internal/logger"
	"QLP/internal/tenancy"
)

func main() {
	// Initialize logger
	if err := logger.InitFromEnv(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.WithComponent("orchestrator-service").Info("Starting QLP Orchestrator Service")

	// Load environment configuration
	config.LoadEnv()

	// Get configuration from environment
	port := config.GetEnvOrDefault("PORT", "8084")

	// Initialize engines
	dagEngine := engines.NewDAGEngine()
	workflowEngine := engines.NewWorkflowEngine(dagEngine)

	// Initialize handlers
	orchestratorHandler := handlers.NewOrchestratorHandler(workflowEngine, dagEngine)

	// Initialize tenant resolver (if using tenant middleware)
	var tenantResolver *tenancy.TenantResolver
	if databaseURL := config.GetEnvOrDefault("DATABASE_URL", ""); databaseURL != "" {
		// TODO: Initialize tenant repository and resolver
		// This would be implemented when we have the tenant repository
		logger.WithComponent("orchestrator-service").Info("Tenant resolution available")
	}

	// Setup routes
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(300 * time.Second)) // 5 minutes for workflow operations
	router.Use(middleware.RequestID)

	// Add tenant middleware if available
	if tenantResolver != nil {
		tenantMiddleware := tenancy.NewTenantMiddleware(tenantResolver)
		router.Use(tenantMiddleware.ResolveTenantFromURL)
	}

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Tenant-specific workflow endpoints
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Workflow management
			r.Post("/workflows", orchestratorHandler.ExecuteWorkflow)
			r.Get("/workflows", orchestratorHandler.ListWorkflows)
			r.Get("/workflows/{workflowId}", orchestratorHandler.GetWorkflow)
			
			// Workflow control
			r.Post("/workflows/{workflowId}/pause", orchestratorHandler.PauseWorkflow)
			r.Post("/workflows/{workflowId}/resume", orchestratorHandler.ResumeWorkflow)
			r.Post("/workflows/{workflowId}/cancel", orchestratorHandler.CancelWorkflow)
			r.Post("/workflows/{workflowId}/retry", orchestratorHandler.RetryTask)
			
			// Workflow metrics
			r.Get("/workflows/{workflowId}/metrics", orchestratorHandler.GetWorkflowMetrics)
		})

		// Global DAG validation (tenant-independent)
		r.Post("/dag/validate", orchestratorHandler.ValidateDAG)
	})

	// Health check endpoint
	router.Get("/health", orchestratorHandler.HealthCheck)

	// Metrics endpoint (basic)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"orchestrator-service","status":"running"}`))
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // 5 minutes for long-running workflows
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("orchestrator-service").Info("Server starting",
			zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("orchestrator-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("orchestrator-service").Info("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("orchestrator-service").Error("Server shutdown failed", zap.Error(err))
	}

	logger.WithComponent("orchestrator-service").Info("Server stopped")
}