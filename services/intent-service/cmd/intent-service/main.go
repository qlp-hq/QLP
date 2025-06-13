package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QLP/internal/config"
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/services/intent-service/internal/handler"
)

func main() {
	// 1. Load Configuration
	config.LoadEnv()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Initialize Dependencies
	// LLM Client (using Azure OpenAI)
	llmClient, err := llm.NewAzureOpenAIClient(os.Getenv("AZURE_OPENAI_ENDPOINT"), os.Getenv("AZURE_OPENAI_API_KEY"))
	if err != nil {
		log.Fatalf("FATAL: Failed to create LLM client: %v", err)
	}

	// Kafka Event Manager
	eventManager, err := events.NewKafkaEventManager()
	if err != nil {
		log.Fatalf("FATAL: Failed to create Kafka Event Manager: %v", err)
	}
	defer eventManager.Close()

	// 3. Setup HTTP Server & Handlers
	intentHandler := handler.NewIntentHandler(llmClient, eventManager)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/intent", intentHandler.Handle)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// 4. Start Server with Graceful Shutdown
	go func() {
		log.Printf("Intent service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("FATAL: Could not start server: %v", err)
		}
	}()

	// 5. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down intent service...")

	// 6. Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server shutdown failed: %v", err)
	}

	log.Println("Intent service shut down gracefully.")
}
