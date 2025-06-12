package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"QLP/services/llm-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// AzureOpenAIProvider implements the Provider interface for Azure OpenAI
type AzureOpenAIProvider struct {
	name     string
	client   *openai.Client
	model    string
	endpoint string
	enabled  bool
}

// NewAzureOpenAIProvider creates a new Azure OpenAI provider
func NewAzureOpenAIProvider(name, apiKey, endpoint, model string) *AzureOpenAIProvider {
	config := openai.DefaultAzureConfig(apiKey, endpoint)
	client := openai.NewClientWithConfig(config)

	if model == "" {
		model = "gpt-4"
	}

	return &AzureOpenAIProvider{
		name:     name,
		client:   client,
		model:    model,
		endpoint: endpoint,
		enabled:  true,
	}
}

// NewAzureOpenAIProviderFromEnv creates a new Azure OpenAI provider from environment variables
func NewAzureOpenAIProviderFromEnv() *AzureOpenAIProvider {
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	model := os.Getenv("AZURE_OPENAI_MODEL")

	if apiKey == "" || endpoint == "" {
		return nil
	}

	if model == "" {
		model = "gpt-4"
	}

	return NewAzureOpenAIProvider("azure-openai", apiKey, endpoint, model)
}

// Name returns the provider name
func (ap *AzureOpenAIProvider) Name() string {
	return ap.name
}

// Type returns the provider type
func (ap *AzureOpenAIProvider) Type() string {
	return "azure_openai"
}

// Complete performs text completion
func (ap *AzureOpenAIProvider) Complete(ctx context.Context, req *contracts.CompletionRequest) (*contracts.CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = ap.model
	}

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.1
	}

	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	logger.WithComponent("azure-openai-provider").Debug("Making completion request",
		zap.String("model", model),
		zap.Int("max_tokens", maxTokens),
		zap.Float32("temperature", temperature))

	resp, err := ap.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("Azure OpenAI completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	response := &contracts.CompletionResponse{
		Content: resp.Choices[0].Message.Content,
		Model:   resp.Model,
		Usage: contracts.UsageMetrics{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// GenerateEmbedding generates text embeddings
func (ap *AzureOpenAIProvider) GenerateEmbedding(ctx context.Context, req *contracts.EmbeddingRequest) (*contracts.EmbeddingResponse, error) {
	var embeddingModel openai.EmbeddingModel
	if req.Model == "" {
		embeddingModel = openai.AdaEmbeddingV2 // text-embedding-ada-002
	} else {
		// Map string model to OpenAI embedding model
		switch req.Model {
		case "text-embedding-ada-002":
			embeddingModel = openai.AdaEmbeddingV2
		default:
			embeddingModel = openai.AdaEmbeddingV2 // Default fallback
		}
	}

	embeddingReq := openai.EmbeddingRequest{
		Input: []string{req.Text},
		Model: embeddingModel,
	}

	logger.WithComponent("azure-openai-provider").Debug("Making embedding request",
		zap.String("model", req.Model),
		zap.Int("text_length", len(req.Text)))

	resp, err := ap.client.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("Azure OpenAI embedding failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	response := &contracts.EmbeddingResponse{
		Embedding: resp.Data[0].Embedding,
		Model:     req.Model,
		Usage: contracts.UsageMetrics{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// ChatCompletion performs chat completion
func (ap *AzureOpenAIProvider) ChatCompletion(ctx context.Context, req *contracts.ChatCompletionRequest) (*contracts.ChatCompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = ap.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.1
	}

	// Convert messages
	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		role := openai.ChatMessageRoleUser
		switch msg.Role {
		case "system":
			role = openai.ChatMessageRoleSystem
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		case "user":
			role = openai.ChatMessageRoleUser
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	chatReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      req.Stream,
	}

	logger.WithComponent("azure-openai-provider").Debug("Making chat completion request",
		zap.String("model", model),
		zap.Int("messages", len(messages)),
		zap.Int("max_tokens", maxTokens),
		zap.Float32("temperature", temperature))

	resp, err := ap.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("Azure OpenAI chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	response := &contracts.ChatCompletionResponse{
		Message: contracts.ChatMessage{
			Role:    "assistant",
			Content: resp.Choices[0].Message.Content,
		},
		Model: resp.Model,
		Usage: contracts.UsageMetrics{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Metadata: req.Metadata,
	}

	return response, nil
}

// HealthCheck performs a health check
func (ap *AzureOpenAIProvider) HealthCheck(ctx context.Context) error {
	// Simple health check with a minimal completion request
	req := &contracts.CompletionRequest{
		Prompt:    "Hello",
		MaxTokens: 5,
	}

	_, err := ap.Complete(ctx, req)
	return err
}

// GetModels returns available models
func (ap *AzureOpenAIProvider) GetModels() []contracts.ModelInfo {
	return []contracts.ModelInfo{
		{
			ID:           ap.model,
			Name:         "Azure OpenAI " + ap.model,
			Type:         "chat",
			MaxTokens:    8192,
			Capabilities: []string{"completion", "chat", "embedding"},
		},
		{
			ID:           "text-embedding-ada-002",
			Name:         "Azure OpenAI Embedding",
			Type:         "embedding",
			MaxTokens:    8191,
			Capabilities: []string{"embedding"},
		},
	}
}

// IsEnabled returns whether the provider is enabled
func (ap *AzureOpenAIProvider) IsEnabled() bool {
	return ap.enabled
}

// SetEnabled sets the provider enabled state
func (ap *AzureOpenAIProvider) SetEnabled(enabled bool) {
	ap.enabled = enabled
	
	logger.WithComponent("azure-openai-provider").Info("Provider enabled state changed",
		zap.String("provider", ap.name),
		zap.Bool("enabled", enabled))
}