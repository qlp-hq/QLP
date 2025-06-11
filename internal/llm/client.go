package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Client interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type FallbackClient struct {
	clients []Client
}

func NewFallbackClient(clients ...Client) *FallbackClient {
	return &FallbackClient{
		clients: clients,
	}
}

func (f *FallbackClient) Complete(ctx context.Context, prompt string) (string, error) {
	var lastErr error

	for i, client := range f.clients {
		log.Printf("Trying LLM client %d", i+1)
		response, err := client.Complete(ctx, prompt)
		if err == nil {
			log.Printf("Successfully used LLM client %d", i+1)
			return response, nil
		}

		log.Printf("LLM client %d failed: %v", i+1, err)
		lastErr = err
	}

	return "", fmt.Errorf("all LLM clients failed, last error: %w", lastErr)
}

type AzureOpenAIClient struct {
	client   *openai.Client
	model    string
	endpoint string
}

func NewAzureOpenAIClient(apiKey, endpoint, model string) *AzureOpenAIClient {
	config := openai.DefaultAzureConfig(apiKey, endpoint)
	client := openai.NewClientWithConfig(config)

	if model == "" {
		model = "gpt-4"
	}

	return &AzureOpenAIClient{
		client:   client,
		model:    model,
		endpoint: endpoint,
	}
}

func (a *AzureOpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: a.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert task decomposition agent. Always respond with valid JSON arrays of tasks.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.1,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("Azure OpenAI completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://192.168.5.240:11434"
	}
	if model == "" {
		model = "llama3"
	}

	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (o *OllamaClient) Complete(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  o.model,
		Prompt: fmt.Sprintf("You are an expert task decomposition agent. Always respond with valid JSON arrays of tasks.\n\n%s", prompt),
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Complete(ctx context.Context, prompt string) (string, error) {
	return `[
  {
    "id": "task_1",
    "type": "codegen",
    "description": "Set up basic Go project structure with main.go and package organization",
    "dependencies": [],
    "priority": "high"
  },
  {
    "id": "task_2", 
    "type": "codegen",
    "description": "Implement HTTP server with basic routing",
    "dependencies": ["task_1"],
    "priority": "high"
  },
  {
    "id": "task_3",
    "type": "test",
    "description": "Write unit tests for the HTTP server",
    "dependencies": ["task_2"],
    "priority": "medium"
  },
  {
    "id": "task_4",
    "type": "doc",
    "description": "Create API documentation",
    "dependencies": ["task_2"],
    "priority": "low"
  }
]`, nil
}

func NewLLMClient() Client {
	var clients []Client

	// Try Azure OpenAI first (requires environment variables)
	azureAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")
	azureEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if azureAPIKey != "" && azureEndpoint != "" {
		azureClient := NewAzureOpenAIClient(
			azureAPIKey,
			azureEndpoint,
			"gpt-4",
		)
		clients = append(clients, azureClient)
	}

	// Fallback to Ollama
	ollamaClient := NewOllamaClient("http://192.168.5.240:11434", "llama3")
	clients = append(clients, ollamaClient)

	// Final fallback to mock
	mockClient := NewMockClient()
	clients = append(clients, mockClient)

	return NewFallbackClient(clients...)
}
