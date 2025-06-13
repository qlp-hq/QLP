package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QLP/internal/config"
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/models"
	"QLP/services/validation-service/internal/validator"
)

const (
	EventArtifactCreated   events.EventType = "artifact.created"
	EventArtifactValidated events.EventType = "artifact.validated"
)

type ValidationService struct {
	validator    *validator.Validator
	eventManager events.Manager
}

func main() {
	config.LoadEnv()

	eventManager, err := events.NewKafkaEventManager()
	if err != nil {
		log.Fatalf("FATAL: Failed to create Kafka Event Manager: %v", err)
	}
	defer eventManager.Close()

	llmClient, err := llm.NewAzureOpenAIClient(os.Getenv("AZURE_OPENAI_ENDPOINT"), os.Getenv("AZURE_OPENAI_API_KEY"))
	if err != nil {
		log.Fatalf("FATAL: Failed to create LLM client: %v", err)
	}

	service := &ValidationService{
		validator:    validator.New(llmClient),
		eventManager: eventManager,
	}

	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	log.Println("Validation service started. Subscribing to 'artifact.created' topic...")
	if err := eventManager.Subscribe(ctx, EventArtifactCreated, service.handleArtifactCreated); err != nil {
		log.Fatalf("FATAL: Subscription to 'artifact.created' failed: %v", err)
	}

	<-ctx.Done()
	log.Println("Validation service shut down.")
}

func (s *ValidationService) handleArtifactCreated(ctx context.Context, event events.Event) error {
	var artifact models.Artifact
	if err := json.Unmarshal(event.Payload, &artifact); err != nil {
		log.Printf("ERROR: Failed to unmarshal artifact: %v", err)
		return nil // Acknowledge message
	}

	log.Printf("Validating artifact %s for task %s", artifact.ID, artifact.Task.ID)

	result := s.validator.Validate(ctx, &artifact)

	log.Printf("Validation complete for artifact %s. Passed: %t, Score: %d", result.Artifact.ID, result.Passed, result.OverallScore)

	// Publish the result
	payload, err := json.Marshal(result)
	if err != nil {
		log.Printf("ERROR: Failed to marshal validation result: %v", err)
		return nil // Don't block processing
	}

	newEvent := events.Event{
		ID:        result.Artifact.ID,
		Type:      EventArtifactValidated,
		Source:    "validation-service",
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if err := s.eventManager.Publish(ctx, newEvent); err != nil {
		log.Printf("ERROR: Failed to publish artifact.validated event: %v", err)
		// Potentially retry or send to a dead-letter queue
	}

	return nil
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down validation service...")
		cancel()
	}()
}
