package agents

import (
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/models"
	"context"
	"testing"
	"time"
)

func TestDynamicAgent_Lifecycle(t *testing.T) {
	// Create test components
	task := models.Task{
		ID:           "test_task_lifecycle",
		Type:         models.TaskTypeCodegen,
		Description:  "Test task for lifecycle validation",
		Priority:     models.PriorityHigh,
		Dependencies: []string{},
		CreatedAt:    time.Now(),
	}

	eventBus := events.NewEventBus()
	mockClient := llm.NewMockClient()

	agentContext := AgentContext{
		ProjectType: "test_project",
		TechStack:   []string{"Go"},
	}

	// Create agent
	agent := NewDynamicAgent(task, mockClient, eventBus, agentContext)

	// Test initial status
	if agent.GetStatus() != AgentStatusInitializing {
		t.Errorf("Expected initial status to be %s, got %s", AgentStatusInitializing, agent.GetStatus())
	}

	// Test initialization
	ctx := context.Background()
	err := agent.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	if agent.GetStatus() != AgentStatusReady {
		t.Errorf("Expected status after init to be %s, got %s", AgentStatusReady, agent.GetStatus())
	}

	if agent.GeneratedPrompt == "" {
		t.Error("Expected non-empty generated prompt after initialization")
	}

	// Test execution
	err = agent.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute agent: %v", err)
	}

	if agent.GetStatus() != AgentStatusCompleted {
		t.Errorf("Expected status after execution to be %s, got %s", AgentStatusCompleted, agent.GetStatus())
	}

	if agent.GetOutput() == "" {
		t.Error("Expected non-empty output after execution")
	}
}

func TestDynamicAgent_ExecutionPromptBuilding(t *testing.T) {
	task := models.Task{
		ID:          "test_prompt_building",
		Type:        models.TaskTypeCodegen,
		Description: "Build execution prompt test",
		Priority:    models.PriorityMedium,
	}

	eventBus := events.NewEventBus()
	mockClient := llm.NewMockClient()

	agentContext := AgentContext{
		ProjectType: "web_service",
		TechStack:   []string{"Go", "PostgreSQL"},
		PreviousOutputs: map[string]string{
			"prev_task": "Previous task output",
		},
	}

	agent := NewDynamicAgent(task, mockClient, eventBus, agentContext)
	agent.GeneratedPrompt = "Test generated prompt"

	prompt := agent.buildExecutionPrompt()

	// Validate execution prompt contains required elements
	if !contains(prompt, "Test generated prompt") {
		t.Error("Expected execution prompt to contain generated prompt")
	}

	if !contains(prompt, "test_prompt_building") {
		t.Error("Expected execution prompt to contain task ID")
	}

	if !contains(prompt, "Build execution prompt test") {
		t.Error("Expected execution prompt to contain task description")
	}

	if !contains(prompt, "web_service") {
		t.Error("Expected execution prompt to contain project type")
	}
}

func TestDynamicAgent_ErrorHandling(t *testing.T) {
	task := models.Task{
		ID:          "test_error_handling",
		Type:        models.TaskTypeCodegen,
		Description: "Test error handling",
		Priority:    models.PriorityLow,
	}

	eventBus := events.NewEventBus()

	// Create a failing mock client
	failingClient := &FailingMockClient{}

	agentContext := AgentContext{
		ProjectType: "test_project",
		TechStack:   []string{"Go"},
	}

	agent := NewDynamicAgent(task, failingClient, eventBus, agentContext)

	// Test initialization failure
	ctx := context.Background()
	err := agent.Initialize(ctx)
	if err == nil {
		t.Error("Expected initialization to fail with failing client")
	}

	if agent.GetStatus() != AgentStatusFailed {
		t.Errorf("Expected status to be %s after failure, got %s", AgentStatusFailed, agent.GetStatus())
	}

	if agent.GetError() == nil {
		t.Error("Expected error to be set after failure")
	}
}

// FailingMockClient for testing error scenarios
type FailingMockClient struct{}

func (f *FailingMockClient) Complete(ctx context.Context, prompt string) (string, error) {
	return "", &MockError{message: "Mock client intentional failure"}
}

type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
