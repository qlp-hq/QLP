package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/services/intent-service/internal/models"
)

type IntentParser struct {
	llmClient llm.Client
}

func NewIntentParser(llmClient llm.Client) *IntentParser {
	return &IntentParser{
		llmClient: llmClient,
	}
}

func (p *IntentParser) ParseIntent(ctx context.Context, userInput string) (*models.Intent, error) {
	prompt := p.buildParsingPrompt(userInput)

	response, err := p.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent with LLM: %w", err)
	}

	intentID := generateID()
	tasks, language, err := p.extractTasksFromResponse(response, intentID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tasks from LLM response: %w", err)
	}

	intent := &models.Intent{
		ID:        intentID,
		UserInput: userInput,
		Tasks:     tasks,
		Language:  language,
		Metadata:  p.extractMetadata(userInput),
		Status:    models.IntentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return intent, nil
}

func (p *IntentParser) buildParsingPrompt(userInput string) string {
	return fmt.Sprintf(`
You are an expert task decomposition agent. Your first and most important job is to identify the primary programming language or technology from the user's intent.

User Intent: %s

Break down the natural language intent into atomic, executable tasks.

For each task, provide:
1. A unique identifier (task_id)
2. Task type (codegen, infra, doc, test, analyze)
3. The primary programming language (e.g., "python", "go", "typescript")
4. Clear description of what needs to be done
5. Dependencies on other tasks (use task IDs)
6. Priority level (high, medium, low)

Return your response as a single JSON object containing the language and a list of tasks. The language should be lowercase.
{
  "language": "python",
  "tasks": [
    {
      "id": "task_1",
      "type": "codegen",
      "description": "Create a function to sum two numbers",
      "dependencies": [],
      "priority": "high"
    }
  ]
}

Focus on creating tasks that are:
- Atomic and independently executable
- Have clear success criteria
- Include necessary context and requirements
- Form a logical dependency graph
`, userInput)
}

func (p *IntentParser) extractTasksFromResponse(response string, intentID string) ([]models.Task, string, error) {
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var responseData struct {
		Language string `json:"language"`
		Tasks    []struct {
			ID           string   `json:"id"`
			Type         string   `json:"type"`
			Description  string   `json:"description"`
			Dependencies []string `json:"dependencies"`
			Priority     string   `json:"priority"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal([]byte(response), &responseData); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	if responseData.Language == "" {
		// In a production system, we should not default. We should fail fast.
		// If the LLM can't identify a language, the request is ambiguous.
		return nil, "", fmt.Errorf("LLM failed to identify a programming language in the user request")
	}

	tasks := make([]models.Task, len(responseData.Tasks))
	now := time.Now()

	for i, td := range responseData.Tasks {
		professionalID := p.generateProfessionalTaskID(td.Type, i+1)

		tasks[i] = models.Task{
			ID:           professionalID,
			IntentID:     intentID,
			Type:         models.TaskType(td.Type),
			Description:  td.Description,
			Language:     responseData.Language,
			Dependencies: p.convertDependenciesToProfessionalIDs(td.Dependencies, responseData.Tasks),
			Priority:     models.Priority(td.Priority),
			Status:       models.TaskStatusPending,
			CreatedAt:    now,
		}
	}

	return tasks, responseData.Language, nil
}

func (p *IntentParser) extractMetadata(userInput string) map[string]string {
	metadata := make(map[string]string)

	metadata["original_length"] = fmt.Sprintf("%d", len(userInput))
	metadata["language_guess"] = "en" // Renamed to avoid confusion with programming language

	if strings.Contains(strings.ToLower(userInput), "web") {
		metadata["domain_guess"] = "web"
	} else if strings.Contains(strings.ToLower(userInput), "api") {
		metadata["domain_guess"] = "api"
	} else if strings.Contains(strings.ToLower(userInput), "mobile") {
		metadata["domain_guess"] = "mobile"
	}

	return metadata
}

func generateID() string {
	// Using a more robust unique ID format for production.
	return fmt.Sprintf("QLI-%d", time.Now().UnixNano())
}

func (p *IntentParser) generateProfessionalTaskID(taskType string, sequence int) string {
	typePrefix := map[string]string{
		"infra":   "INF",
		"codegen": "DEV",
		"test":    "TST",
		"doc":     "DOC",
		"analyze": "ANA",
	}

	prefix, exists := typePrefix[taskType]
	if !exists {
		prefix = "GEN" // Generic for unknown task types
	}

	timestamp := time.Now().Format("060102") // Shorter timestamp
	return fmt.Sprintf("T-%s-%s-%03d", prefix, timestamp, sequence)
}

func (p *IntentParser) convertDependenciesToProfessionalIDs(dependencies []string, taskData []struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Priority     string   `json:"priority"`
}) []string {
	if len(dependencies) == 0 {
		return nil // Return nil for no dependencies, it's cleaner JSON.
	}

	idMap := make(map[string]string)
	for i, td := range taskData {
		idMap[td.ID] = p.generateProfessionalTaskID(td.Type, i+1)
	}

	professionalDeps := make([]string, len(dependencies))
	for i, dep := range dependencies {
		if profID, ok := idMap[dep]; ok {
			professionalDeps[i] = profID
		} else {
			// This indicates a malformed response from the LLM. We should log this.
			// For now, we pass the original dependency ID.
			professionalDeps[i] = dep
		}
	}

	return professionalDeps
}
