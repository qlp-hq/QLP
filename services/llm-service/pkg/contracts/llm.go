package contracts

import (
	"fmt"
	"time"
)

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt      string            `json:"prompt"`
	Model       string            `json:"model,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CompletionResponse represents the response from text completion
type CompletionResponse struct {
	Content       string            `json:"content"`
	Model         string            `json:"model"`
	Usage         UsageMetrics      `json:"usage"`
	ResponseTime  time.Duration     `json:"response_time"`
	Provider      string            `json:"provider"`
	RequestID     string            `json:"request_id"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// EmbeddingRequest represents a request for text embedding
type EmbeddingRequest struct {
	Text     string            `json:"text"`
	Model    string            `json:"model,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EmbeddingResponse represents the response from text embedding
type EmbeddingResponse struct {
	Embedding    []float32         `json:"embedding"`
	Model        string            `json:"model"`
	Dimensions   int               `json:"dimensions"`
	Usage        UsageMetrics      `json:"usage"`
	ResponseTime time.Duration     `json:"response_time"`
	Provider     string            `json:"provider"`
	RequestID    string            `json:"request_id"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Messages    []ChatMessage     `json:"messages"`
	Model       string            `json:"model,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ChatMessage represents a single message in a chat
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from chat completion
type ChatCompletionResponse struct {
	Message      ChatMessage       `json:"message"`
	Model        string            `json:"model"`
	Usage        UsageMetrics      `json:"usage"`
	ResponseTime time.Duration     `json:"response_time"`
	Provider     string            `json:"provider"`
	RequestID    string            `json:"request_id"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UsageMetrics represents token usage information
type UsageMetrics struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderStatus represents the status of an LLM provider
type ProviderStatus struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // "azure_openai", "ollama", "mock"
	Available    bool              `json:"available"`
	Healthy      bool              `json:"healthy"`
	ResponseTime time.Duration     `json:"response_time"`
	LastCheck    time.Time         `json:"last_check"`
	ErrorCount   int               `json:"error_count"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ServiceStatus represents the overall service status
type ServiceStatus struct {
	Status      string           `json:"status"`
	Timestamp   time.Time        `json:"timestamp"`
	Providers   []ProviderStatus `json:"providers"`
	ActiveModel string           `json:"active_model"`
	Version     string           `json:"version"`
}

// ProviderConfig represents configuration for an LLM provider
type ProviderConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Enabled     bool              `json:"enabled"`
	Priority    int               `json:"priority"`
	Config      map[string]string `json:"config"`
	Models      []ModelInfo       `json:"models"`
	Limits      ProviderLimits    `json:"limits"`
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"` // "completion", "embedding", "chat"
	MaxTokens    int      `json:"max_tokens"`
	Capabilities []string `json:"capabilities"`
}

// ProviderLimits represents rate limits and quotas for a provider
type ProviderLimits struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`
	MaxRequestSize    int `json:"max_request_size"`
}

// ListProvidersResponse represents the response for listing providers
type ListProvidersResponse struct {
	Providers []ProviderStatus `json:"providers"`
	Total     int              `json:"total"`
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Service   string            `json:"service"`
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// MetricsResponse represents metrics for the LLM service
type MetricsResponse struct {
	TotalRequests     int64             `json:"total_requests"`
	TotalTokens       int64             `json:"total_tokens"`
	AverageLatency    time.Duration     `json:"average_latency"`
	ErrorRate         float64           `json:"error_rate"`
	ActiveProviders   int               `json:"active_providers"`
	RequestsByModel   map[string]int64  `json:"requests_by_model"`
	RequestsByProvider map[string]int64 `json:"requests_by_provider"`
	Uptime            time.Duration     `json:"uptime"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Batch processing structures

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Requests []interface{}     `json:"requests"` // Array of CompletionRequest, EmbeddingRequest, etc.
	Type     string            `json:"type"`     // "completion", "embedding", "chat"
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BatchResponse represents a batch processing response
type BatchResponse struct {
	Responses   []interface{}     `json:"responses"` // Array of corresponding response types
	BatchID     string            `json:"batch_id"`
	TotalCount  int               `json:"total_count"`
	SuccessCount int              `json:"success_count"`
	FailureCount int              `json:"failure_count"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Stream processing structures

// StreamRequest represents a streaming request
type StreamRequest struct {
	Prompt      string            `json:"prompt"`
	Model       string            `json:"model,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// StreamChunk represents a single chunk in a streaming response
type StreamChunk struct {
	Content   string `json:"content"`
	Delta     string `json:"delta"`
	Finished  bool   `json:"finished"`
	RequestID string `json:"request_id"`
	Index     int    `json:"index"`
}

// Provider-specific configurations

// AzureOpenAIConfig represents Azure OpenAI specific configuration
type AzureOpenAIConfig struct {
	APIKey     string `json:"api_key"`
	Endpoint   string `json:"endpoint"`
	APIVersion string `json:"api_version"`
	Model      string `json:"model"`
}

// OllamaConfig represents Ollama specific configuration
type OllamaConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	Timeout int    `json:"timeout"` // seconds
}

// OpenAIConfig represents OpenAI specific configuration
type OpenAIConfig struct {
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url,omitempty"`
	Model       string `json:"model"`
	Organization string `json:"organization,omitempty"`
}

// Request/Response validation

// Validate validates a completion request
func (r *CompletionRequest) Validate() error {
	if r.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if r.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}
	if r.Temperature < 0 || r.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	return nil
}

// Validate validates an embedding request
func (r *EmbeddingRequest) Validate() error {
	if r.Text == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

// Validate validates a chat completion request
func (r *ChatCompletionRequest) Validate() error {
	if len(r.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	for i, msg := range r.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return fmt.Errorf("message %d: invalid role %s", i, msg.Role)
		}
	}
	if r.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}
	if r.Temperature < 0 || r.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	return nil
}