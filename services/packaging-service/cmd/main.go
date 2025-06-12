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

	"QLP/services/packaging-service/internal/handlers"
	"QLP/services/packaging-service/internal/engines"
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

	logger.WithComponent("packaging-service").Info("Starting QLP Packaging Service")

	// Load environment configuration
	config.LoadEnv()

	// Get configuration from environment
	outputDir := config.GetEnvOrDefault("OUTPUT_DIR", "./output")
	port := config.GetEnvOrDefault("PORT", "8083")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.WithComponent("packaging-service").Fatal("Failed to create output directory",
			zap.String("dir", outputDir),
			zap.Error(err))
	}

	// Initialize engines
	capsuleEngine := engines.NewCapsuleEngine(outputDir)
	quantumDropEngine := engines.NewQuantumDropsEngine(outputDir)

	// Initialize handlers
	packagingHandler := handlers.NewPackagingHandler(capsuleEngine, quantumDropEngine)

	// Initialize tenant resolver (if using tenant middleware)
	var tenantResolver *tenancy.TenantResolver
	if databaseURL := config.GetEnvOrDefault("DATABASE_URL", ""); databaseURL != "" {
		// TODO: Initialize tenant repository and resolver
		// This would be implemented when we have the tenant repository
		logger.WithComponent("packaging-service").Info("Tenant resolution available")
	}

	// Setup routes
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(120 * time.Second)) // Longer timeout for packaging operations
	router.Use(middleware.RequestID)

	// Add tenant middleware if available
	if tenantResolver != nil {
		tenantMiddleware := tenancy.NewTenantMiddleware(tenantResolver)
		router.Use(tenantMiddleware.ResolveTenantFromURL)
	}

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Capsule endpoints
			r.Post("/capsules", packagingHandler.CreateCapsule)
			r.Get("/capsules", packagingHandler.ListCapsules)
			r.Get("/capsules/{capsuleId}", packagingHandler.GetCapsule)
			r.Get("/capsules/{capsuleId}/download", packagingHandler.DownloadCapsule)

			// Quantum drops endpoints
			r.Post("/quantum-drops", packagingHandler.CreateQuantumDrops)
			r.Get("/quantum-drops/{dropId}", packagingHandler.GetQuantumDrop)
		})
	})

	// Health check endpoint
	router.Get("/health", packagingHandler.HealthCheck)

	// Metrics endpoint (basic)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"packaging-service","status":"running"}`))
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Longer write timeout for large packages
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("packaging-service").Info("Server starting",
			zap.String("port", port),
			zap.String("output_dir", outputDir))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("packaging-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("packaging-service").Info("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("packaging-service").Error("Server shutdown failed", zap.Error(err))
	}

	logger.WithComponent("packaging-service").Info("Server stopped")
}