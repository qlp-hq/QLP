package agents

import (
	"QLP/internal/models"
	"context"
	"testing"
	"time"
)

func TestMetaPromptGenerator_BuildMetaPrompt(t *testing.T) {
	generator := &MetaPromptGenerator{}

	task := models.Task{
		ID:           "test_task_001",
		Type:         models.TaskTypeCodegen,
		Description:  "Create a REST API endpoint",
		Priority:     models.PriorityHigh,
		Dependencies: []string{},
		CreatedAt:    time.Now(),
	}

	context := AgentContext{
		ProjectType:        "web_api",
		TechStack:          []string{"Go", "HTTP", "JSON"},
		Dependencies:       []models.Task{},
		OutputRequirements: []string{"Complete code", "Error handling"},
		Constraints: map[string]string{
			"performance": "high",
			"security":    "required",
		},
		PreviousOutputs: map[string]string{},
	}

	prompt := generator.buildMetaPrompt(task, context)

	// Validate prompt contains essential elements
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	if !contains(prompt, "test_task_001") {
		t.Error("Expected prompt to contain task ID")
	}

	if !contains(prompt, "codegen") {
		t.Error("Expected prompt to contain task type")
	}

	if !contains(prompt, "REST API endpoint") {
		t.Error("Expected prompt to contain task description")
	}

	if !contains(prompt, "web_api") {
		t.Error("Expected prompt to contain project type")
	}

	// Test task-specific guidance
	guidance := generator.getTaskTypeSpecificGuidance(models.TaskTypeCodegen)
	if !contains(guidance, "CODEGEN AGENT REQUIREMENTS") {
		t.Error("Expected codegen-specific guidance")
	}
}

func TestMetaPromptGenerator_TaskTypeSpecificGuidance(t *testing.T) {
	generator := &MetaPromptGenerator{}

	testCases := []struct {
		taskType models.TaskType
		expected string
	}{
		{models.TaskTypeCodegen, "CODEGEN AGENT REQUIREMENTS"},
		{models.TaskTypeInfra, "INFRASTRUCTURE AGENT REQUIREMENTS"},
		{models.TaskTypeDoc, "DOCUMENTATION AGENT REQUIREMENTS"},
		{models.TaskTypeTest, "TESTING AGENT REQUIREMENTS"},
		{models.TaskTypeAnalyze, "ANALYSIS AGENT REQUIREMENTS"},
	}

	for _, tc := range testCases {
		guidance := generator.getTaskTypeSpecificGuidance(tc.taskType)
		if !contains(guidance, tc.expected) {
			t.Errorf("Expected guidance for %s to contain '%s'", tc.taskType, tc.expected)
		}
	}
}

func TestMetaPromptGenerator_FormatPreviousOutputs(t *testing.T) {
	generator := &MetaPromptGenerator{}

	// Test empty outputs
	emptyResult := generator.formatPreviousOutputs(map[string]string{})
	if emptyResult != "No previous task outputs available." {
		t.Error("Expected specific message for empty outputs")
	}

	// Test with outputs
	outputs := map[string]string{
		"task_1": "Generated API endpoint code",
		"task_2": "Created database schema",
	}

	result := generator.formatPreviousOutputs(outputs)
	if !contains(result, "task_1") || !contains(result, "task_2") {
		t.Error("Expected formatted output to contain all task IDs")
	}
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(str) < len(substr) {
		return false
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
