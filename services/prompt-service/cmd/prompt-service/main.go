package main

import (
	"QLP/internal/database"
	"QLP/services/prompt-service/internal/handler"
	"QLP/services/prompt-service/internal/repository"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PROMPT_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}

	db, err := database.New()
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer db.Close()

	if !db.IsConnected() {
		log.Println("Database is not connected. The service might not function as expected.")
	}

	promptRepo := repository.NewPromptRepository(db)
	promptHandler := handler.NewPromptHandler(promptRepo)

	router := mux.NewRouter()
	promptHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Prompt service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("FATAL: Could not start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down prompt service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server shutdown failed: %v", err)
	}

	log.Println("Prompt service shut down gracefully.")
}
