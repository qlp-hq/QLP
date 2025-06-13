package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"QLP/internal/config"
	"QLP/internal/events"
	"QLP/internal/models"
	"QLP/services/common/validation"
	"QLP/services/orchestrator-service/internal/dag"
	"QLP/services/orchestrator-service/internal/statemanager"
)

const (
	EventIntentReceived    events.EventType = "intent.received"
	EventArtifactValidated events.EventType = "artifact.validated"
	EventTaskReady         events.EventType = "task.ready"
	EventIntentCompleted   events.EventType = "intent.completed"
)

type Orchestrator struct {
	stateManager statemanager.StateManager
	eventManager events.Manager
}

func main() {
	config.LoadEnv()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	stateManager := statemanager.NewRedisStateManager(redisAddr)
	eventManager, err := events.NewKafkaEventManager()
	if err != nil {
		log.Fatalf("FATAL: Failed to create Kafka Event Manager: %v", err)
	}
	defer eventManager.Close()

	orchestrator := &Orchestrator{
		stateManager: stateManager,
		eventManager: eventManager,
	}

	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		log.Println("Subscribing to 'intent.received' topic...")
		if err := eventManager.Subscribe(ctx, EventIntentReceived, orchestrator.handleIntentReceived); err != nil {
			log.Printf("ERROR: Subscription to 'intent.received' failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Subscribing to 'artifact.validated' topic...")
		if err := eventManager.Subscribe(ctx, EventArtifactValidated, orchestrator.handleArtifactValidated); err != nil {
			log.Printf("ERROR: Subscription to 'artifact.validated' failed: %v", err)
		}
	}()

	log.Println("Orchestrator service started.")
	wg.Wait()
	<-ctx.Done()
	log.Println("Orchestrator service shut down.")
}

func (o *Orchestrator) handleIntentReceived(ctx context.Context, event events.Event) error {
	var intent models.Intent
	if err := json.Unmarshal(event.Payload, &intent); err != nil {
		log.Printf("ERROR: Failed to unmarshal intent: %v", err)
		return nil
	}

	log.Printf("Orchestrating new intent: %s", intent.ID)

	graph := dag.NewDAG()
	var finalTasks []models.Task
	taskMap := make(map[string]models.Task)

	for _, task := range intent.Tasks {
		taskMap[task.ID] = task
	}

	for _, task := range intent.Tasks {
		if task.Ensemble {
			log.Printf("Detected ensemble task: %s. Creating child tasks.", task.ID)

			providers := []string{"azure", "anthropic", "groq"}
			var childTaskIDs []string

			for _, provider := range providers {
				childTask := task // copy
				childTask.ID = fmt.Sprintf("%s-%s", task.ID, provider)
				childTask.Model = provider
				childTask.Ensemble = false // It's a child, not an ensemble itself
				childTask.Description = fmt.Sprintf("[%s] %s", provider, task.Description)

				finalTasks = append(finalTasks, childTask)
				childTaskIDs = append(childTaskIDs, childTask.ID)
				taskMap[childTask.ID] = childTask
			}

			judgementTask := models.Task{
				ID:           fmt.Sprintf("%s-judgement", task.ID),
				IntentID:     task.IntentID,
				Type:         "judgement",
				Description:  fmt.Sprintf("Select the best output from child tasks for original task: %s", task.Description),
				Dependencies: childTaskIDs,
				Priority:     task.Priority,
				Status:       models.TaskStatusPending,
				CreatedAt:    time.Now(),
			}
			finalTasks = append(finalTasks, judgementTask)
			taskMap[judgementTask.ID] = judgementTask

			// Update dependencies of other tasks that depended on the original ensemble task
			for i, otherTask := range finalTasks {
				for j, depID := range otherTask.Dependencies {
					if depID == task.ID {
						finalTasks[i].Dependencies[j] = judgementTask.ID
					}
				}
			}

		} else {
			finalTasks = append(finalTasks, task)
		}
	}
	intent.Tasks = finalTasks

	for _, task := range intent.Tasks {
		graph.AddTask(task)
	}
	for _, task := range intent.Tasks {
		for _, depID := range task.Dependencies {
			if err := graph.AddEdge(depID, task.ID); err != nil {
				log.Printf("ERROR: Failed to add edge for intent %s: %v", intent.ID, err)
				return nil
			}
		}
	}

	o.stateManager.Set(intent.ID, graph)
	log.Printf("Built and saved DAG for intent %s with %d tasks.", intent.ID, len(intent.Tasks))

	return o.dispatchReadyTasks(ctx, intent.ID, graph)
}

func (o *Orchestrator) handleArtifactValidated(ctx context.Context, event events.Event) error {
	var validationResult validation.ValidationResult
	if err := json.Unmarshal(event.Payload, &validationResult); err != nil {
		log.Printf("ERROR: Failed to unmarshal validation result: %v", err)
		return nil
	}

	if !validationResult.Passed {
		log.Printf("WARN: Artifact %s failed validation with score %d. Stopping this branch of the graph.",
			validationResult.Artifact.ID, validationResult.OverallScore)
		// In a future step, this is where we would trigger a refinement loop
		// by publishing a "refinement.required" event.
		return nil
	}

	intentID := validationResult.Artifact.Task.IntentID
	taskID := validationResult.Artifact.Task.ID

	log.Printf("Artifact %s for task %s (intent: %s) has been validated and passed.", validationResult.Artifact.ID, taskID, intentID)

	graph, found := o.stateManager.Get(intentID)
	if !found {
		log.Printf("WARN: Received validation for unknown or completed intent: %s", intentID)
		return nil
	}

	graph.MarkTaskComplete(taskID)
	log.Printf("Task %s marked as complete for intent %s", taskID, intentID)

	if graph.IsEmpty() {
		log.Printf("SUCCESS: All tasks for intent %s have been completed.", intentID)
		o.stateManager.Delete(intentID)
		o.publishIntentCompleted(ctx, intentID)
		return nil
	}

	o.stateManager.Set(intentID, graph)
	return o.dispatchReadyTasks(ctx, intentID, graph)
}

func (o *Orchestrator) dispatchReadyTasks(ctx context.Context, intentID string, graph *dag.DAG) error {
	readyTasks := graph.GetReadyTasks()
	if len(readyTasks) == 0 {
		return nil
	}

	log.Printf("Dispatching %d ready tasks for intent %s", len(readyTasks), intentID)
	for _, task := range readyTasks {
		payload, _ := json.Marshal(task)
		event := events.Event{
			ID:        task.ID,
			Type:      EventTaskReady,
			Source:    "orchestrator-service",
			Timestamp: time.Now(),
			Payload:   payload,
		}
		if err := o.eventManager.Publish(ctx, event); err != nil {
			log.Printf("ERROR: Failed to publish task.ready event for task %s: %v", task.ID, err)
		}
	}
	return nil
}

func (o *Orchestrator) publishIntentCompleted(ctx context.Context, intentID string) {
	payload, _ := json.Marshal(map[string]string{"intent_id": intentID, "status": "completed"})
	event := events.Event{
		ID:        intentID,
		Type:      EventIntentCompleted,
		Source:    "orchestrator-service",
		Timestamp: time.Now(),
		Payload:   payload,
	}
	if err := o.eventManager.Publish(ctx, event); err != nil {
		log.Printf("ERROR: Failed to publish intent.completed event for intent %s: %v", intentID, err)
	}
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down orchestrator service...")
		cancel()
	}()
}
