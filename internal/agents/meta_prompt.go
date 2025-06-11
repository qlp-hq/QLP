package agents

import (
	"context"
	"fmt"
	"strings"

	"QLP/internal/llm"
	"QLP/internal/models"
)

type MetaPromptGenerator struct {
	llmClient llm.Client
}

func NewMetaPromptGenerator(llmClient llm.Client) *MetaPromptGenerator {
	return &MetaPromptGenerator{
		llmClient: llmClient,
	}
}

func (m *MetaPromptGenerator) GenerateAgentPrompt(ctx context.Context, task models.Task, context AgentContext) (string, error) {
	metaPrompt := m.buildMetaPrompt(task, context)

	agentPrompt, err := m.llmClient.Complete(ctx, metaPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate agent prompt: %w", err)
	}

	return strings.TrimSpace(agentPrompt), nil
}

type AgentContext struct {
	ProjectType        string            `json:"project_type"`
	TechStack          []string          `json:"tech_stack"`
	Dependencies       []models.Task     `json:"dependencies"`
	OutputRequirements []string          `json:"output_requirements"`
	Constraints        map[string]string `json:"constraints"`
	PreviousOutputs    map[string]string `json:"previous_outputs"`
}

func (m *MetaPromptGenerator) buildMetaPrompt(task models.Task, context AgentContext) string {
	basePrompt := fmt.Sprintf(`
You are an Expert %s Agent. Your job is to DIRECTLY EXECUTE the following task and provide the complete, ready-to-use output.

TASK TO EXECUTE:
- ID: %s
- Type: %s
- Description: %s
- Priority: %s
- Dependencies: %v

PROJECT CONTEXT:
- Project Type: %s
- Tech Stack: %v
- Output Requirements: %v
- Constraints: %v

PREVIOUS TASK OUTPUTS:
%s

EXECUTE THIS TASK NOW and provide ONLY the final deliverable output - no explanations, no steps, no process descriptions.
`,
		task.Type,
		task.ID,
		task.Type,
		task.Description,
		task.Priority,
		task.Dependencies,
		context.ProjectType,
		context.TechStack,
		context.OutputRequirements,
		context.Constraints,
		m.formatPreviousOutputs(context.PreviousOutputs),
	)

	return basePrompt + m.getTaskTypeSpecificGuidance(task.Type)
}

func (m *MetaPromptGenerator) getTaskTypeSpecificGuidance(taskType models.TaskType) string {
	switch taskType {
	case models.TaskTypeCodegen:
		return `
REQUIRED OUTPUT: Complete, executable Go code including:
- Package declaration and all necessary imports
- Full function implementations with error handling
- Comments explaining key functionality
- Code that compiles and runs without modification

EXAMPLE OUTPUT FORMAT:
package main

import (
    "fmt"
    "net/http"
    "log"
)

func main() {
    http.HandleFunc("/health", healthHandler)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "OK")
}
`
	case models.TaskTypeInfra:
		return `
REQUIRED OUTPUT: Complete infrastructure configuration (YAML/HCL) including:
- All resource definitions
- Working configuration ready for deployment
- No placeholder values

EXAMPLE OUTPUT FORMAT:
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 80
    targetPort: 8080
`
	case models.TaskTypeDoc:
		return `
REQUIRED OUTPUT: Complete documentation in Markdown format including:
- Proper heading structure
- Code examples and usage instructions
- Complete sentences and paragraphs

EXAMPLE OUTPUT FORMAT:
# API Documentation

## Overview
This API provides user management functionality.

## Endpoints

### GET /health
Returns the health status of the service.
`
	case models.TaskTypeTest:
		return `
REQUIRED OUTPUT: Complete test code including:
- Test functions with proper naming
- All necessary imports
- Test assertions and data

EXAMPLE OUTPUT FORMAT:
package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    rr := httptest.NewRecorder()
    healthHandler(rr, req)
    
    if rr.Code != http.StatusOK {
        t.Errorf("Expected status OK, got %d", rr.Code)
    }
}
`
	case models.TaskTypeAnalyze:
		return `
REQUIRED OUTPUT: Complete analysis report including:
- Executive summary
- Detailed findings with data
- Specific recommendations

EXAMPLE OUTPUT FORMAT:
# Performance Analysis Report

## Executive Summary
The application shows peak memory usage of 256MB with average response times of 45ms.

## Key Findings
1. Database queries lack proper indexing
2. Memory usage spikes during peak traffic

## Recommendations
1. Add database indexes on frequently queried columns
2. Implement connection pooling
`
	default:
		return `
REQUIRED OUTPUT: Complete, production-ready deliverable for the specified task type.
Provide only the final working result - no steps, no explanations.
`
	}
}

func (m *MetaPromptGenerator) formatPreviousOutputs(outputs map[string]string) string {
	if len(outputs) == 0 {
		return "No previous task outputs available."
	}

	var formatted strings.Builder
	for taskID, output := range outputs {
		formatted.WriteString(fmt.Sprintf("Task %s Output:\n%s\n\n", taskID, output))
	}

	return formatted.String()
}
