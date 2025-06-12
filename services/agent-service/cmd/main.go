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

	"QLP/services/agent-service/internal/handlers"
	"QLP/services/agent-service/internal/factory"
	"QLP/services/llm-service/pkg/client"
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

	logger.WithComponent("agent-service").Info("Starting QLP Agent Service")

	// Load environment configuration
	config.LoadEnv()

	// Get configuration from environment
	port := config.GetEnvOrDefault("PORT", "8086")
	llmServiceURL := config.GetEnvOrDefault("LLM_SERVICE_URL", "http://localhost:8085")

	// Initialize LLM client
	llmClient := client.NewLLMClient(llmServiceURL)

	// Initialize agent factory
	agentFactory := factory.NewAgentFactory(llmClient)

	// Initialize handlers
	agentHandler := handlers.NewAgentHandler(agentFactory)

	// Initialize tenant resolver (if using tenant middleware)
	var tenantResolver *tenancy.TenantResolver
	if databaseURL := config.GetEnvOrDefault("DATABASE_URL", ""); databaseURL != "" {
		// TODO: Initialize tenant repository and resolver
		// This would be implemented when we have the tenant repository
		logger.WithComponent("agent-service").Info("Tenant resolution available")
	}

	// Setup routes
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(600 * time.Second)) // 10 minutes for agent operations
	router.Use(middleware.RequestID)

	// Add tenant middleware if available
	if tenantResolver != nil {
		tenantMiddleware := tenancy.NewTenantMiddleware(tenantResolver)
		router.Use(tenantMiddleware.ResolveTenantFromURL)
	}

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Tenant-specific agent endpoints
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Core agent operations
			r.Post("/agents", agentHandler.CreateAgent)
			r.Get("/agents", agentHandler.ListAgents)
			r.Get("/agents/{agentId}", agentHandler.GetAgent)
			
			// Agent execution and control
			r.Post("/agents/{agentId}/execute", agentHandler.ExecuteAgent)
			r.Post("/agents/{agentId}/cancel", agentHandler.CancelAgent)
			r.Post("/agents/{agentId}/retry", agentHandler.RetryAgent)
			
			// Batch operations
			r.Post("/agents/batch", agentHandler.BatchCreateAgents)
			
			// Specialized agents
			r.Post("/agents/deployment-validator", agentHandler.CreateDeploymentValidator)
		})

		// Global service endpoints (tenant-independent)
		r.Get("/status", agentHandler.GetServiceStatus)
		r.Get("/metrics", agentHandler.GetMetrics)
	})

	// Health check endpoint
	router.Get("/health", agentHandler.HealthCheck)

	// Metrics endpoint (Prometheus-compatible)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Basic metrics endpoint - in production this would be Prometheus metrics
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP agent_service_status Service status\n# TYPE agent_service_status gauge\nagent_service_status 1\n"))
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second, // 10 minutes for agent operations
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("agent-service").Info("Server starting",
			zap.String("port", port),
			zap.String("llm_service_url", llmServiceURL))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("agent-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("agent-service").Info("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("agent-service").Error("Server shutdown failed", zap.Error(err))
	}

	logger.WithComponent("agent-service").Info("Server stopped")
}