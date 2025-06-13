package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QLP/internal/agents"
	"QLP/internal/config"
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/models"
	promptclient "QLP/services/prompt-service/pkg/client"
)

const (
	EventTaskReady       events.EventType = "task.ready"
	EventArtifactCreated events.EventType = "artifact.created"
)

// Global agent factory
var agentFactory *agents.AgentFactory

func main() {
	config.LoadEnv()

	// Initialize Dependencies
	llmClient, err := llm.NewLLMClient()
	if err != nil {
		log.Fatalf("FATAL: Failed to create LLM client: %v", err)
	}

	// The factory needs a map of clients and a default, but since NewLLMClient
	// returns a single fallback client, we'll wrap it in a map.
	llmClients := map[string]llm.Client{"default": llmClient}

	promptServiceURL := os.Getenv("PROMPT_SERVICE_URL")
	if promptServiceURL == "" {
		promptServiceURL = "http://prompt-service:8081"
	}
	promptClient := promptclient.New(promptServiceURL)

	agentFactory, err = agents.NewAgentFactory(llmClients, llmClient, promptClient)
	if err != nil {
		log.Fatalf("FATAL: Failed to create AgentFactory: %v", err)
	}

	eventManager, err := events.NewKafkaEventManager()
	if err != nil {
		log.Fatalf("FATAL: Failed to create Kafka Event Manager: %v", err)
	}
	defer eventManager.Close()

	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	log.Println("Agent Worker started. Subscribing to ready tasks...")
	err = eventManager.Subscribe(ctx, EventTaskReady, func(ctx context.Context, event events.Event) error {
		return handleReadyTask(ctx, event, eventManager)
	})

	if err != nil {
		log.Fatalf("FATAL: Subscription failed: %v", err)
	}

	<-ctx.Done()
	log.Println("Agent Worker service has been shut down.")
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Agent Worker service shutting down...")
		cancel()
	}()
}

func handleReadyTask(ctx context.Context, event events.Event, eventManager events.Manager) error {
	var task models.Task
	if err := json.Unmarshal(event.Payload, &task); err != nil {
		log.Printf("ERROR: Failed to unmarshal task from event payload: %v", err)
		return nil // Acknowledge the message to avoid reprocessing a malformed one
	}

	log.Printf("Processing task %s of type %s", task.ID, task.Type)

	agent, err := agentFactory.GetAgent(task)
	if err != nil {
		log.Printf("ERROR: Could not get agent for task %s: %v", task.ID, err)
		// Here we might want to publish a "task.failed" event
		return nil
	}

	artifact, err := agent.Execute(ctx, task)
	if err != nil {
		log.Printf("ERROR: Agent failed to execute task %s: %v", task.ID, err)
		// Here we might want to publish a "task.failed" event
		return nil
	}

	log.Printf("Successfully executed task %s, created artifact %s", task.ID, artifact.ID)

	return publishArtifact(ctx, eventManager, artifact)
}

func publishArtifact(ctx context.Context, eventManager events.Manager, artifact *models.Artifact) error {
	payload, err := json.Marshal(artifact)
	if err != nil {
		log.Printf("ERROR: Failed to marshal artifact %s: %v", artifact.ID, err)
		return err // Return error to signal failure
	}

	event := events.Event{
		ID:        artifact.ID,
		Type:      EventArtifactCreated,
		Source:    "agent-worker",
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if err := eventManager.Publish(ctx, event); err != nil {
		log.Printf("ERROR: Failed to publish artifact.created event for artifact %s: %v", artifact.ID, err)
		return err
	}

	log.Printf("Published event %s for artifact %s", EventArtifactCreated, artifact.ID)
	return nil
}
