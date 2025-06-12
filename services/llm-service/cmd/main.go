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

	"QLP/services/llm-service/internal/handlers"
	"QLP/services/llm-service/internal/providers"
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

	logger.WithComponent("llm-service").Info("Starting QLP LLM Service")

	// Load environment configuration
	config.LoadEnv()

	// Get configuration from environment
	port := config.GetEnvOrDefault("PORT", "8085")

	// Initialize provider manager
	providerManager := providers.NewProviderManager()

	// Register providers based on configuration
	registerProviders(providerManager)

	// Start health checks for providers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go providerManager.StartHealthChecks(ctx)

	// Initialize handlers
	llmHandler := handlers.NewLLMHandler(providerManager)

	// Initialize tenant resolver (if using tenant middleware)
	var tenantResolver *tenancy.TenantResolver
	if databaseURL := config.GetEnvOrDefault("DATABASE_URL", ""); databaseURL != "" {
		// TODO: Initialize tenant repository and resolver
		// This would be implemented when we have the tenant repository
		logger.WithComponent("llm-service").Info("Tenant resolution available")
	}

	// Setup routes
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(180 * time.Second)) // 3 minutes for LLM operations
	router.Use(middleware.RequestID)

	// Add tenant middleware if available
	if tenantResolver != nil {
		tenantMiddleware := tenancy.NewTenantMiddleware(tenantResolver)
		router.Use(tenantMiddleware.ResolveTenantFromURL)
	}

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Tenant-specific LLM endpoints
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Core LLM operations
			r.Post("/completion", llmHandler.Complete)
			r.Post("/embedding", llmHandler.GenerateEmbedding)
			r.Post("/chat/completion", llmHandler.ChatCompletion)
			
			// Batch processing
			r.Post("/batch", llmHandler.BatchProcess)
		})

		// Global service endpoints (tenant-independent)
		r.Get("/providers", llmHandler.ListProviders)
		r.Get("/status", llmHandler.GetServiceStatus)
		r.Get("/metrics", llmHandler.GetMetrics)
	})

	// Health check endpoint
	router.Get("/health", llmHandler.HealthCheck)

	// Metrics endpoint (Prometheus-compatible)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Basic metrics endpoint - in production this would be Prometheus metrics
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP llm_service_status Service status\n# TYPE llm_service_status gauge\nllm_service_status 1\n"))
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 180 * time.Second, // 3 minutes for LLM operations
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("llm-service").Info("Server starting",
			zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("llm-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.WithComponent("llm-service").Info("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.WithComponent("llm-service").Error("Server shutdown failed", zap.Error(err))
	}

	logger.WithComponent("llm-service").Info("Server stopped")
}

// registerProviders registers all available LLM providers
func registerProviders(providerManager *providers.ProviderManager) {
	logger.WithComponent("llm-service").Info("Registering LLM providers")

	// Register Azure OpenAI provider (if configured)
	if azureProvider := providers.NewAzureOpenAIProviderFromEnv(); azureProvider != nil {
		providerManager.RegisterProvider(azureProvider)
		logger.WithComponent("llm-service").Info("Azure OpenAI provider registered")
	} else {
		logger.WithComponent("llm-service").Info("Azure OpenAI provider not configured (missing environment variables)")
	}

	// Register Ollama provider
	ollamaProvider := providers.NewOllamaProviderFromEnv()
	providerManager.RegisterProvider(ollamaProvider)
	logger.WithComponent("llm-service").Info("Ollama provider registered")

	// Register Mock provider (always available for development)
	mockProvider := providers.NewMockProvider("mock")
	providerManager.RegisterProvider(mockProvider)
	logger.WithComponent("llm-service").Info("Mock provider registered")

	logger.WithComponent("llm-service").Info("Provider registration completed")
}