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

	"QLP/services/worker-runtime/internal/handlers"
	"QLP/services/worker-runtime/internal/executor"
	"QLP/services/worker-runtime/internal/agents"
	"QLP/services/worker-runtime/internal/sandbox"
	"QLP/internal/config"
	"QLP/internal/logger"
	"QLP/internal/llm"
)

func main() {
	// Initialize logger
	if err := logger.InitFromEnv(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.WithComponent("worker-runtime").Info("Starting QLP Worker Runtime Service")

	// Load environment configuration
	config.LoadEnv()

	// Initialize sandbox environment
	sandboxMgr, err := sandbox.NewManager(sandbox.Config{
		Runtime:           "docker", // or "firecracker" for production
		BaseImage:         "alpine:latest",
		NetworkIsolation:  true,
		FileSystemIsolation: true,
		ResourceLimits: sandbox.DefaultResourceLimits(),
	})
	if err != nil {
		logger.WithComponent("worker-runtime").Fatal("Failed to initialize sandbox manager", zap.Error(err))
	}
	defer sandboxMgr.Cleanup()

	// Initialize LLM client
	llmClient := llm.NewLLMClient()

	// Initialize agent factory
	agentFactory := agents.NewFactory(agents.FactoryConfig{
		LLMClient: llmClient,
		Timeout:   30 * time.Second,
	})

	// Initialize task executor
	taskExecutor := executor.NewTaskExecutor(executor.Config{
		AgentFactory:    agentFactory,
		SandboxManager:  sandboxMgr,
		MaxConcurrent:   10,
		DefaultTimeout:  300 * time.Second,
	})

	// Initialize handlers
	workHandler := handlers.NewWorkerHandler(taskExecutor)

	// Setup routes
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) 
	router.Use(middleware.Timeout(300 * time.Second)) // 5 minute timeout for long-running tasks
	router.Use(middleware.Compress(5))

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Task execution endpoints
			r.Post("/tasks/execute", workHandler.ExecuteTask)
			r.Get("/executions", workHandler.ListExecutions)
			r.Get("/executions/{executionId}", workHandler.GetExecution)
			r.Delete("/executions/{executionId}", workHandler.CancelExecution)
			
			// Real-time streaming endpoint
			r.Get("/executions/{executionId}/stream", workHandler.StreamExecution)
		})
		
		// System endpoints (no tenant scope)
		r.Get("/health", workHandler.HealthCheck)
		r.Get("/metrics", workHandler.GetMetrics)
		r.Get("/agents/types", workHandler.GetAgentTypes)
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second, // Allow long-running tasks
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("worker-runtime").Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("worker-runtime").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("worker-runtime").Info("Shutting down server...")
	
	// Cancel running tasks
	taskExecutor.Shutdown(ctx)
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("worker-runtime").Error("Server shutdown failed", zap.Error(err))
	}
	logger.WithComponent("worker-runtime").Info("Server stopped")
}