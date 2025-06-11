package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/models"
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

	tasks, err := p.extractTasksFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tasks from LLM response: %w", err)
	}

	intent := &models.Intent{
		ID:          generateID(),
		UserInput:   userInput,
		ParsedTasks: tasks,
		Metadata:    p.extractMetadata(userInput),
		CreatedAt:   time.Now(),
		Status:      models.IntentStatusPending,
	}

	return intent, nil
}

func (p *IntentParser) buildParsingPrompt(userInput string) string {
	return fmt.Sprintf(`
You are an expert task decomposition agent. Break down the following natural language intent into atomic, executable tasks.

User Intent: %s

For each task, provide:
1. A unique identifier (task_id)
2. Task type (codegen, infra, doc, test, analyze)
3. Clear description of what needs to be done
4. Dependencies on other tasks (use task IDs)
5. Priority level (high, medium, low)

Return your response as a JSON array of tasks with this structure:
[
  {
    "id": "task_1",
    "type": "codegen",
    "description": "Create a REST API endpoint for user authentication",
    "dependencies": [],
    "priority": "high"
  }
]

Focus on creating tasks that are:
- Atomic and independently executable
- Have clear success criteria
- Include necessary context and requirements
- Form a logical dependency graph
`, userInput)
}

func (p *IntentParser) extractTasksFromResponse(response string) ([]models.Task, error) {
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var taskData []struct {
		ID           string   `json:"id"`
		Type         string   `json:"type"`
		Description  string   `json:"description"`
		Dependencies []string `json:"dependencies"`
		Priority     string   `json:"priority"`
	}

	if err := json.Unmarshal([]byte(response), &taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	tasks := make([]models.Task, len(taskData))
	now := time.Now()

	for i, td := range taskData {
		professionalID := p.generateProfessionalTaskID(td.Type, i+1)
		
		tasks[i] = models.Task{
			ID:           professionalID,
			Type:         models.TaskType(td.Type),
			Description:  td.Description,
			Dependencies: p.convertDependenciesToProfessionalIDs(td.Dependencies, taskData),
			Priority:     models.Priority(td.Priority),
			Status:       models.TaskStatusPending,
			CreatedAt:    now,
		}
	}

	return tasks, nil
}

func (p *IntentParser) extractMetadata(userInput string) map[string]string {
	metadata := make(map[string]string)

	metadata["original_length"] = fmt.Sprintf("%d", len(userInput))
	metadata["language"] = "en"

	if strings.Contains(strings.ToLower(userInput), "web") {
		metadata["domain"] = "web"
	} else if strings.Contains(strings.ToLower(userInput), "api") {
		metadata["domain"] = "api"
	} else if strings.Contains(strings.ToLower(userInput), "mobile") {
		metadata["domain"] = "mobile"
	}

	return metadata
}

func generateID() string {
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
		prefix = "GEN"
	}
	
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("QL-%s-%s-%03d", prefix, timestamp, sequence)
}

func (p *IntentParser) convertDependenciesToProfessionalIDs(dependencies []string, taskData []struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Priority     string   `json:"priority"`
}) []string {
	if len(dependencies) == 0 {
		return dependencies
	}
	
	professionalDeps := make([]string, len(dependencies))
	
	for i, dep := range dependencies {
		for j, task := range taskData {
			if task.ID == dep {
				professionalDeps[i] = p.generateProfessionalTaskID(task.Type, j+1)
				break
			}
		}
		if professionalDeps[i] == "" {
			professionalDeps[i] = dep
		}
	}
	
	return professionalDeps
}
