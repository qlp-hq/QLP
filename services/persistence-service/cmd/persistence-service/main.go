package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"QLP/internal/config"
	"QLP/internal/events"
	"QLP/internal/models"
	"QLP/services/persistence-service/internal/storage"
)

const (
	EventArtifactValidated events.EventType = "artifact.validated"
	EventArtifactPersisted events.EventType = "artifact.persisted"
)

type PersistenceService struct {
	storage      storage.Store
	eventManager events.Manager
}

func main() {
	config.LoadEnv()

	storage, err := storage.NewLocalStorage()
	if err != nil {
		log.Fatalf("FATAL: Failed to create local storage: %v", err)
	}

	eventManager, err := events.NewKafkaEventManager()
	if err != nil {
		log.Fatalf("FATAL: Failed to create Kafka Event Manager: %v", err)
	}
	defer eventManager.Close()

	service := &PersistenceService{
		storage:      storage,
		eventManager: eventManager,
	}

	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	log.Println("Persistence service started. Subscribing to 'artifact.validated' topic...")
	if err := eventManager.Subscribe(ctx, EventArtifactValidated, service.handleArtifactValidated); err != nil {
		log.Fatalf("FATAL: Subscription to 'artifact.validated' failed: %v", err)
	}

	<-ctx.Done()
	log.Println("Persistence service shut down.")
}

func (s *PersistenceService) handleArtifactValidated(ctx context.Context, event events.Event) error {
	var result models.ValidationResult
	if err := json.Unmarshal(event.Payload, &result); err != nil {
		log.Printf("ERROR: Failed to unmarshal validation result: %v", err)
		return nil // Acknowledge message
	}

	// Only persist artifacts that passed validation
	if !result.Passed {
		log.Printf("Skipping persistence for failed artifact %s", result.Artifact.ID)
		return nil
	}

	artifact := result.Artifact
	// Define a structured path: <intent_id>/<task_id>/<artifact_id>.<ext>
	// This requires file extension logic.
	fileName := fmt.Sprintf("%s.txt", artifact.ID) // Basic extension
	if lang, ok := artifact.Metadata["language"]; ok {
		// A real implementation would have a map of language to file extension
		if lang == "python" {
			fileName = fmt.Sprintf("%s.py", artifact.ID)
		} else if lang == "go" {
			fileName = fmt.Sprintf("%s.go", artifact.ID)
		}
	}

	path := filepath.Join(artifact.Task.IntentID, artifact.Task.ID, fileName)

	log.Printf("Persisting artifact %s to %s", artifact.ID, path)

	storagePath, err := s.storage.Save(ctx, path, []byte(artifact.Content))
	if err != nil {
		log.Printf("ERROR: Failed to save artifact %s: %v", artifact.ID, err)
		return nil // NACK and retry later? For now, we drop it.
	}

	// Publish the final event
	payload, err := json.Marshal(map[string]string{
		"artifact_id":  artifact.ID,
		"storage_path": storagePath,
	})
	if err != nil {
		log.Printf("ERROR: Failed to marshal persistence info: %v", err)
		return nil
	}

	newEvent := events.Event{
		ID:        artifact.ID,
		Type:      EventArtifactPersisted,
		Source:    "persistence-service",
		Timestamp: time.Now(),
		Payload:   payload,
	}

	return s.eventManager.Publish(ctx, newEvent)
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down persistence service...")
		cancel()
	}()
}
