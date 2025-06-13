package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Prompt is a direct copy of the model from the prompt-service.
// In a real-world scenario, this might be a shared library or versioned contract.
type Prompt struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	TaskType   string    `json:"task_type"`
	PromptText string    `json:"prompt_text"`
	Version    int       `json:"version"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PromptServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *PromptServiceClient {
	return &PromptServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *PromptServiceClient) GetActivePromptByTaskType(ctx context.Context, taskType string) (*Prompt, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/prompts?task_type=%s&active=true&limit=1", c.baseURL, taskType), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prompt-service returned non-200 status: %d", resp.StatusCode)
	}

	var prompts []Prompt
	if err := json.NewDecoder(resp.Body).Decode(&prompts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no active prompt found for task type %s", taskType)
	}

	return &prompts[0], nil
}
