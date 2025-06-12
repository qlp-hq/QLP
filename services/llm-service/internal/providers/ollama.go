package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"QLP/services/llm-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	name    string
	baseURL string
	model   string
	client  *http.Client
	enabled bool
}

// OllamaRequest represents an Ollama API request
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaChatRequest represents an Ollama chat API request
type OllamaChatRequest struct {
	Model    string        `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// OllamaMessage represents a message in Ollama chat
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse represents an Ollama API response
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// OllamaChatResponse represents an Ollama chat API response
type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(name, baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}

	return &OllamaProvider{
		name:    name,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		enabled: true,
	}
}

// NewOllamaProviderFromEnv creates a new Ollama provider from environment variables
func NewOllamaProviderFromEnv() *OllamaProvider {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	model := os.Getenv("OLLAMA_MODEL")

	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}

	return NewOllamaProvider("ollama", baseURL, model)
}

// Name returns the provider name
func (op *OllamaProvider) Name() string {
	return op.name
}

// Type returns the provider type
func (op *OllamaProvider) Type() string {
	return "ollama"
}

// Complete performs text completion
func (op *OllamaProvider) Complete(ctx context.Context, req *contracts.CompletionRequest) (*contracts.CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = op.model
	}

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	// Combine system prompt with user prompt
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, req.Prompt)

	ollamaReq := OllamaRequest{
		Model:  model,
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.WithComponent("ollama-provider").Debug("Making completion request",
		zap.String("model", model),
		zap.String("url", op.baseURL+"/api/generate"))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", op.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := op.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	response := &contracts.CompletionResponse{
		Content: strings.TrimSpace(ollamaResp.Response),
		Model:   model,
		Usage: contracts.UsageMetrics{
			// Ollama doesn't provide token counts, so we estimate
			PromptTokens:     len(strings.Fields(req.Prompt)),
			CompletionTokens: len(strings.Fields(ollamaResp.Response)),
			TotalTokens:      len(strings.Fields(req.Prompt)) + len(strings.Fields(ollamaResp.Response)),
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// GenerateEmbedding generates text embeddings (simplified implementation)
func (op *OllamaProvider) GenerateEmbedding(ctx context.Context, req *contracts.EmbeddingRequest) (*contracts.EmbeddingResponse, error) {
	// Ollama doesn't have a standard embedding API, so we'll use a simple hash-based approach
	// In a real implementation, you'd use a dedicated embedding model
	embedding := op.generateSimpleEmbedding(req.Text)

	response := &contracts.EmbeddingResponse{
		Embedding: embedding,
		Model:     "simple-hash",
		Usage: contracts.UsageMetrics{
			PromptTokens: len(strings.Fields(req.Text)),
			TotalTokens:  len(strings.Fields(req.Text)),
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// ChatCompletion performs chat completion
func (op *OllamaProvider) ChatCompletion(ctx context.Context, req *contracts.ChatCompletionRequest) (*contracts.ChatCompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = op.model
	}

	// Convert messages
	var messages []OllamaMessage
	for _, msg := range req.Messages {
		messages = append(messages, OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	ollamaReq := OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.WithComponent("ollama-provider").Debug("Making chat completion request",
		zap.String("model", model),
		zap.Int("messages", len(messages)),
		zap.String("url", op.baseURL+"/api/chat"))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", op.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := op.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Calculate token estimates
	var totalPromptTokens int
	for _, msg := range req.Messages {
		totalPromptTokens += len(strings.Fields(msg.Content))
	}
	completionTokens := len(strings.Fields(ollamaResp.Message.Content))

	response := &contracts.ChatCompletionResponse{
		Message: contracts.ChatMessage{
			Role:    ollamaResp.Message.Role,
			Content: strings.TrimSpace(ollamaResp.Message.Content),
		},
		Model: model,
		Usage: contracts.UsageMetrics{
			PromptTokens:     totalPromptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalPromptTokens + completionTokens,
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// HealthCheck performs a health check
func (op *OllamaProvider) HealthCheck(ctx context.Context) error {
	// Check if Ollama is running by making a simple request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", op.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := op.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GetModels returns available models
func (op *OllamaProvider) GetModels() []contracts.ModelInfo {
	return []contracts.ModelInfo{
		{
			ID:           op.model,
			Name:         "Ollama " + op.model,
			Type:         "chat",
			MaxTokens:    4096, // Typical for most Ollama models
			Capabilities: []string{"completion", "chat"},
		},
	}
}

// IsEnabled returns whether the provider is enabled
func (op *OllamaProvider) IsEnabled() bool {
	return op.enabled
}

// SetEnabled sets the provider enabled state
func (op *OllamaProvider) SetEnabled(enabled bool) {
	op.enabled = enabled
	
	logger.WithComponent("ollama-provider").Info("Provider enabled state changed",
		zap.String("provider", op.name),
		zap.Bool("enabled", enabled))
}

// generateSimpleEmbedding creates a basic embedding from text for fallback
func (op *OllamaProvider) generateSimpleEmbedding(text string) []float32 {
	// Simple character-based embedding for development/fallback
	embedding := make([]float32, 1536) // Match OpenAI text-embedding-ada-002 dimensions
	
	// Basic hash-like distribution
	for i, char := range text {
		if i >= len(embedding) {
			break
		}
		embedding[i%len(embedding)] += float32(char) / 1000.0
	}
	
	// Normalize
	var norm float32
	for _, val := range embedding {
		norm += val * val
	}
	if norm > 0 {
		norm = 1.0 / sqrt(norm)
		for i := range embedding {
			embedding[i] *= norm
		}
	}
	
	return embedding
}

// Simple square root implementation
func sqrt(x float32) float32 {
	if x <= 0 {
		return 0
	}
	
	guess := x / 2
	for i := 0; i < 10; i++ { // Newton's method iterations
		guess = (guess + x/guess) / 2
	}
	return guess
}