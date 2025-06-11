package main

import (
	"QLP/internal/models"
	"QLP/internal/orchestrator"
	"context"
	"testing"
	"time"
)

func TestOrchestratorBasicFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orch := orchestrator.New()

	// Test intent processing
	intent, err := orch.ProcessIntent(ctx, "Create a simple web API with user authentication")
	if err != nil {
		t.Fatalf("Failed to process intent: %v", err)
	}

	if intent == nil {
		t.Fatal("Expected non-nil intent")
	}

	if intent.UserInput != "Create a simple web API with user authentication" {
		t.Error("Expected intent to preserve original user input")
	}

	if intent.Status != models.IntentStatusProcessing {
		t.Errorf("Expected intent status to be %s, got %s", models.IntentStatusProcessing, intent.Status)
	}

	if len(intent.ParsedTasks) == 0 {
		t.Error("Expected tasks to be generated from intent")
	}

	t.Logf("Generated %d tasks from intent", len(intent.ParsedTasks))
	for i, task := range intent.ParsedTasks {
		t.Logf("Task %d: %s - %s (%s)", i+1, task.ID, task.Description, task.Type)

		// Validate task structure
		if task.ID == "" {
			t.Error("Expected task to have non-empty ID")
		}

		if task.Description == "" {
			t.Error("Expected task to have non-empty description")
		}

		if task.Status != models.TaskStatusPending {
			t.Errorf("Expected new task status to be %s, got %s", models.TaskStatusPending, task.Status)
		}
	}

	// Validate task types are reasonable
	foundValidType := false
	validTypes := []models.TaskType{
		models.TaskTypeCodegen,
		models.TaskTypeInfra,
		models.TaskTypeDoc,
		models.TaskTypeTest,
		models.TaskTypeAnalyze,
	}

	for _, task := range intent.ParsedTasks {
		for _, validType := range validTypes {
			if task.Type == validType {
				foundValidType = true
				break
			}
		}
	}

	if !foundValidType {
		t.Error("Expected at least one task with a valid task type")
	}
}

func TestOrchestratorTaskGraphBuilding(t *testing.T) {
	orch := orchestrator.New()

	// Create test tasks with dependencies
	tasks := []models.Task{
		{
			ID:           "task_1",
			Type:         models.TaskTypeCodegen,
			Description:  "Setup project structure",
			Dependencies: []string{},
			Priority:     models.PriorityHigh,
		},
		{
			ID:           "task_2",
			Type:         models.TaskTypeCodegen,
			Description:  "Implement API endpoints",
			Dependencies: []string{"task_1"},
			Priority:     models.PriorityHigh,
		},
		{
			ID:           "task_3",
			Type:         models.TaskTypeTest,
			Description:  "Write tests",
			Dependencies: []string{"task_2"},
			Priority:     models.PriorityMedium,
		},
	}

	// Note: We can't directly test buildTaskGraph as it's private,
	// but we can test it through ProcessIntent
	ctx := context.Background()
	intent, err := orch.ProcessIntent(ctx, "Create a web API")
	if err != nil {
		t.Fatalf("Failed to process intent: %v", err)
	}

	// The processed intent should have proper task relationships
	if len(intent.ParsedTasks) == 0 {
		t.Error("Expected tasks to be generated")
	}

	// Check that dependencies are properly structured
	for _, task := range intent.ParsedTasks {
		for _, depID := range task.Dependencies {
			// Verify dependency exists in task list
			found := false
			for _, depTask := range intent.ParsedTasks {
				if depTask.ID == depID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Task %s has dependency %s that doesn't exist in task list", task.ID, depID)
			}
		}
	}
}

func TestOrchestratorMultipleIntents(t *testing.T) {
	orch := orchestrator.New()
	ctx := context.Background()

	intents := []string{
		"Create a REST API for user management",
		"Build a microservice with database integration",
		"Implement a CLI tool for file processing",
		"Create a gRPC service with monitoring",
	}

	for i, intentText := range intents {
		t.Run(fmt.Sprintf("Intent_%d", i+1), func(t *testing.T) {
			intent, err := orch.ProcessIntent(ctx, intentText)
			if err != nil {
				t.Fatalf("Failed to process intent '%s': %v", intentText, err)
			}

			if len(intent.ParsedTasks) == 0 {
				t.Errorf("Expected tasks for intent: %s", intentText)
			}

			t.Logf("Intent '%s' generated %d tasks", intentText, len(intent.ParsedTasks))
		})
	}
}

// Helper function for string formatting in tests
func fmt_Sprintf(format string, args ...interface{}) string {
	// Simple sprintf implementation for testing
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			// Replace first %d with the integer
			for i := 0; i < len(result)-1; i++ {
				if result[i] == '%' && result[i+1] == 'd' {
					// Convert int to string manually
					intStr := ""
					num := v
					if num == 0 {
						intStr = "0"
					} else {
						for num > 0 {
							intStr = string(rune('0'+num%10)) + intStr
							num /= 10
						}
					}
					result = result[:i] + intStr + result[i+2:]
					break
				}
			}
		case string:
			// Replace first %s with the string
			for i := 0; i < len(result)-1; i++ {
				if result[i] == '%' && result[i+1] == 's' {
					result = result[:i] + v + result[i+2:]
					break
				}
			}
		}
	}
	return result
}
