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

	"QLP/services/validation-service/internal/handlers"
	"QLP/services/validation-service/internal/engines"
	"QLP/services/validation-service/internal/validators"
	"QLP/services/validation-service/internal/scanners"
	"QLP/internal/logger"
	"QLP/internal/llm"
	"QLP/internal/tenancy"
	"QLP/internal/database"
)

func main() {
	// Initialize logger
	if err := logger.InitFromEnv(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.WithComponent("validation-service").Info("Starting QLP Validation Service")

	// Initialize database connection for tenancy
	dbInstance, err := database.New()
	if err != nil {
		logger.WithComponent("validation-service").Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbInstance.Close()

	// Initialize tenant repository and resolver
	tenantRepository := tenancy.NewPostgresTenantRepository(dbInstance.GetConnection())
	tenantResolver := tenancy.NewTenantResolver(tenantRepository)
	tenantMiddleware := tenancy.NewTenantMiddleware(tenantResolver)

	// Initialize LLM client for critique
	llmClient := llm.NewLLMClient()

	// Initialize syntax validators
	syntaxValidators := validators.NewSyntaxValidatorRegistry()

	// Initialize security scanners
	securityScanners := scanners.NewSecurityScannerRegistry()

	// Initialize quality analyzers
	qualityAnalyzers := validators.NewQualityAnalyzerRegistry()

	// Initialize validation engine
	validationEngine := engines.NewValidationEngine(engines.Config{
		LLMClient:        llmClient,
		SyntaxValidators: syntaxValidators,
		SecurityScanners: securityScanners,
		QualityAnalyzers: qualityAnalyzers,
		DefaultTimeout:   300 * time.Second,
		MaxConcurrent:    10,
	})

	// Initialize handlers
	validationHandler := handlers.NewValidationHandler(validationEngine)

	// Setup routes
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(300 * time.Second)) // 5 minute timeout for comprehensive validation
	router.Use(middleware.Compress(5))

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Apply tenant resolution middleware
			r.Use(tenantMiddleware.ResolveTenantFromURL)
			
			// Validation endpoints
			r.Post("/validate", validationHandler.ValidateContent)
			r.Get("/validations", validationHandler.ListValidations)
			r.Get("/validations/{validationId}", validationHandler.GetValidation)
			r.Delete("/validations/{validationId}", validationHandler.CancelValidation)
			
			// Batch validation
			r.Post("/validate/batch", validationHandler.ValidateBatch)
			r.Get("/validate/batch/{batchId}", validationHandler.GetBatchStatus)
			
			// Real-time streaming endpoint
			r.Get("/validations/{validationId}/stream", validationHandler.StreamValidation)
			
			// Tenant-specific endpoints
			r.Handle("/tenant/metrics", tenantMiddleware.TenantMetricsHandler())
		})
		
		// System endpoints (no tenant scope)
		r.Get("/health", validationHandler.HealthCheck)
		r.Get("/metrics", validationHandler.GetMetrics)
		r.Get("/validators", validationHandler.GetValidators)
		r.Get("/rules", validationHandler.GetRules)
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second, // Allow long-running validations
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("validation-service").Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("validation-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("validation-service").Info("Shutting down server...")
	
	// Shutdown validation engine
	validationEngine.Shutdown(ctx)
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("validation-service").Error("Server shutdown failed", zap.Error(err))
	}
	logger.WithComponent("validation-service").Info("Server stopped")
}