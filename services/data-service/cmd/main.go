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

	"QLP/services/data-service/internal/handlers"
	"QLP/services/data-service/internal/repository"
	"QLP/internal/config"
	"QLP/internal/logger"
)

func main() {
	// Initialize logger
	if err := logger.InitFromEnv(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.WithComponent("data-service").Info("Starting QLP Data Service")

	// Load environment configuration
	config.LoadEnv()

	// Initialize database connection
	databaseURL := config.GetEnvOrDefault("DATABASE_URL", "postgres://localhost/qlp_dev?sslmode=disable")
	db, err := repository.NewConnection(databaseURL)
	if err != nil {
		logger.WithComponent("data-service").Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	intentRepo := repository.NewIntentRepository(db)
	vectorRepo := repository.NewVectorRepository(db)

	// Initialize handlers
	intentHandler := handlers.NewIntentHandler(intentRepo)
	vectorHandler := handlers.NewVectorHandler(vectorRepo)

	// Setup routes
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tenants/{tenantId}", func(r chi.Router) {
			// Intent endpoints
			r.Post("/intents", intentHandler.CreateIntent)
			r.Get("/intents", intentHandler.ListIntents)
			r.Get("/intents/{intentId}", intentHandler.GetIntent)
			r.Put("/intents/{intentId}", intentHandler.UpdateIntent)
			r.Delete("/intents/{intentId}", intentHandler.DeleteIntent)
			
			// Vector search endpoints
			r.Post("/vector/similar", vectorHandler.FindSimilar)
			r.Post("/vector/embed", vectorHandler.CreateEmbedding)
		})
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithComponent("data-service").Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("data-service").Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.WithComponent("data-service").Info("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithComponent("data-service").Error("Server shutdown failed", zap.Error(err))
	}
	logger.WithComponent("data-service").Info("Server stopped")
}