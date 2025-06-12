package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"QLP/services/llm-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// MockProvider implements the Provider interface for testing and development
type MockProvider struct {
	name    string
	enabled bool
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	if name == "" {
		name = "mock"
	}

	return &MockProvider{
		name:    name,
		enabled: true,
	}
}

// Name returns the provider name
func (mp *MockProvider) Name() string {
	return mp.name
}

// Type returns the provider type
func (mp *MockProvider) Type() string {
	return "mock"
}

// Complete performs mock text completion
func (mp *MockProvider) Complete(ctx context.Context, req *contracts.CompletionRequest) (*contracts.CompletionResponse, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Generate mock response based on prompt content
	content := mp.generateMockCompletion(req.Prompt)

	response := &contracts.CompletionResponse{
		Content: content,
		Model:   req.Model,
		Usage: contracts.UsageMetrics{
			PromptTokens:     len(req.Prompt) / 4, // Rough token estimate
			CompletionTokens: len(content) / 4,
			TotalTokens:      (len(req.Prompt) + len(content)) / 4,
		},
		ResponseTime: 100 * time.Millisecond,
		Provider:     mp.name,
		RequestID:    fmt.Sprintf("mock-completion-%d", time.Now().UnixNano()),
	}

	logger.WithComponent("mock-provider").Info("Generated completion",
		zap.String("provider", mp.name),
		zap.String("model", req.Model),
		zap.Int("prompt_tokens", response.Usage.PromptTokens),
		zap.Int("completion_tokens", response.Usage.CompletionTokens))

	return response, nil
}

// GenerateEmbedding generates mock embeddings
func (mp *MockProvider) GenerateEmbedding(ctx context.Context, req *contracts.EmbeddingRequest) (*contracts.EmbeddingResponse, error) {
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// Generate mock embedding
	embedding := mp.generateMockEmbedding(req.Text)

	response := &contracts.EmbeddingResponse{
		Embedding:    embedding,
		Model:        req.Model,
		Dimensions:   len(embedding),
		Usage: contracts.UsageMetrics{
			PromptTokens: len(req.Text) / 4,
			TotalTokens:  len(req.Text) / 4,
		},
		ResponseTime: 50 * time.Millisecond,
		Provider:     mp.name,
		RequestID:    fmt.Sprintf("mock-embedding-%d", time.Now().UnixNano()),
	}

	logger.WithComponent("mock-provider").Info("Generated embedding",
		zap.String("provider", mp.name),
		zap.String("model", req.Model),
		zap.Int("text_length", len(req.Text)),
		zap.Int("embedding_dim", len(embedding)))

	return response, nil
}

// ChatCompletion performs mock chat completion
func (mp *MockProvider) ChatCompletion(ctx context.Context, req *contracts.ChatCompletionRequest) (*contracts.ChatCompletionResponse, error) {
	// Simulate processing time
	time.Sleep(150 * time.Millisecond)

	// Get the last user message
	var lastMessage string
	if len(req.Messages) > 0 {
		lastMessage = req.Messages[len(req.Messages)-1].Content
	}

	// Generate mock response
	content := mp.generateMockChatResponse(lastMessage, req.Messages)

	response := &contracts.ChatCompletionResponse{
		Message: contracts.ChatMessage{
			Role:    "assistant",
			Content: content,
		},
		Model: req.Model,
		Usage: contracts.UsageMetrics{
			PromptTokens:     mp.countTokensInMessages(req.Messages),
			CompletionTokens: len(content) / 4,
			TotalTokens:      mp.countTokensInMessages(req.Messages) + len(content)/4,
		},
		ResponseTime: 150 * time.Millisecond,
		Provider:     mp.name,
		RequestID:    fmt.Sprintf("mock-chat-%d", time.Now().UnixNano()),
	}

	logger.WithComponent("mock-provider").Info("Generated chat completion",
		zap.String("provider", mp.name),
		zap.String("model", req.Model),
		zap.Int("message_count", len(req.Messages)),
		zap.Int("completion_tokens", response.Usage.CompletionTokens))

	return response, nil
}

// HealthCheck performs a mock health check
func (mp *MockProvider) HealthCheck(ctx context.Context) error {
	// Mock providers are always healthy
	return nil
}

// GetModels returns available mock models
func (mp *MockProvider) GetModels() []contracts.ModelInfo {
	return []contracts.ModelInfo{
		{
			ID:   "mock-gpt-3.5-turbo",
			Name: "Mock GPT-3.5 Turbo",
			Type: "chat",
			Capabilities: []string{
				"text-generation",
				"chat-completion",
				"function-calling",
			},
			MaxTokens: 4096,
		},
		{
			ID:   "mock-text-embedding-ada-002",
			Name: "Mock Text Embedding Ada 002",
			Type: "embedding",
			Capabilities: []string{
				"text-embedding",
			},
			MaxTokens: 8191,
		},
	}
}

// IsEnabled returns whether the provider is enabled
func (mp *MockProvider) IsEnabled() bool {
	return mp.enabled
}

// SetEnabled sets the provider enabled state
func (mp *MockProvider) SetEnabled(enabled bool) {
	mp.enabled = enabled
	logger.WithComponent("mock-provider").Info("Provider enabled state changed",
		zap.String("provider", mp.name),
		zap.Bool("enabled", enabled))
}

// Helper methods for generating mock responses

func (mp *MockProvider) generateMockCompletion(prompt string) string {
	prompt = strings.ToLower(prompt)
	
	// Generate contextual responses based on prompt content
	if strings.Contains(prompt, "task") && strings.Contains(prompt, "json") {
		return mp.generateTaskJSON()
	}
	
	if strings.Contains(prompt, "code") || strings.Contains(prompt, "function") {
		return mp.generateCodeResponse(prompt)
	}
	
	if strings.Contains(prompt, "documentation") || strings.Contains(prompt, "readme") {
		return mp.generateDocumentationResponse()
	}
	
	if strings.Contains(prompt, "test") {
		return mp.generateTestResponse()
	}
	
	// Default response
	return "This is a mock response generated for development and testing purposes. " +
		"The mock provider has analyzed your request and provided this simulated output."
}

func (mp *MockProvider) generateTaskJSON() string {
	return `{
	"task_id": "mock-task-001",
	"description": "This is a mock task generated for testing",
	"status": "pending",
	"priority": "medium",
	"steps": [
		{
			"id": 1,
			"description": "Initialize project structure",
			"status": "pending"
		},
		{
			"id": 2,
			"description": "Implement core functionality",
			"status": "pending"
		},
		{
			"id": 3,
			"description": "Add tests and documentation",
			"status": "pending"
		}
	],
	"metadata": {
		"created_at": "2024-01-01T12:00:00Z",
		"estimated_duration": "2 hours"
	}
}`
}

func (mp *MockProvider) generateCodeResponse(prompt string) string {
	if strings.Contains(prompt, "go") || strings.Contains(prompt, "golang") {
		return `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/health", healthHandler)
	
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World! This is a mock Go server.")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"healthy\"}")
}`
	}
	
	return "// This is mock code generated for demonstration purposes\nfunction mockFunction() {\n    console.log('Mock implementation');\n    return 'success';\n}"
}

func (mp *MockProvider) generateDocumentationResponse() string {
	return `# Project Documentation

## Overview
This is mock documentation generated for development and testing purposes.

## Features
- Feature 1: Basic functionality
- Feature 2: Advanced operations
- Feature 3: Integration capabilities

## Installation
` + "```bash" + `
npm install mock-package
` + "```" + `

## Usage
` + "```javascript" + `
const mock = require('mock-package');
mock.initialize();
` + "```" + `

## API Reference
See the API documentation for detailed information about available endpoints and methods.`
}

func (mp *MockProvider) generateTestResponse() string {
	return `package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "Hello, World! This is a mock Go server."
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}`
}

func (mp *MockProvider) generateMockChatResponse(lastMessage string, messages []contracts.ChatMessage) string {
	lastMessage = strings.ToLower(lastMessage)
	
	// Check context from conversation
	hasSystemMessage := false
	systemContent := ""
	for _, msg := range messages {
		if msg.Role == "system" {
			hasSystemMessage = true
			systemContent = strings.ToLower(msg.Content)
			break
		}
	}
	
	// Generate contextual responses
	if hasSystemMessage && strings.Contains(systemContent, "task") {
		return "I understand you need help with task decomposition. Based on your request, I can break this down into manageable steps with clear dependencies and priorities."
	}
	
	if strings.Contains(lastMessage, "hello") || strings.Contains(lastMessage, "hi") {
		return "Hello! I'm a mock AI assistant. I'm here to help you with your development and testing needs."
	}
	
	if strings.Contains(lastMessage, "code") {
		return "I can help you with code generation. What programming language and specific functionality are you looking for?"
	}
	
	if strings.Contains(lastMessage, "explain") {
		return "I'd be happy to explain that concept. This is a mock explanation that demonstrates how the AI would provide detailed information about the topic you're asking about."
	}
	
	// Default conversational response
	return "Thank you for your message. This is a mock response from the AI assistant. In a real implementation, I would provide a more specific and helpful response based on your exact needs."
}

func (mp *MockProvider) generateMockEmbedding(text string) []float32 {
	// Generate deterministic embedding based on text content
	embedding := make([]float32, 1536) // Standard embedding size
	
	// Simple hash-like approach for consistent embeddings
	for i, char := range text {
		if i >= len(embedding) {
			break
		}
		embedding[i%len(embedding)] += float32(char) / 1000.0
	}
	
	// Add some variance based on text length and content
	textLen := float32(len(text))
	for i := range embedding {
		embedding[i] += textLen / 10000.0
		if i%2 == 0 {
			embedding[i] *= 1.1
		} else {
			embedding[i] *= 0.9
		}
	}
	
	// Normalize the embedding
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

func (mp *MockProvider) countTokensInMessages(messages []contracts.ChatMessage) int {
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += len(msg.Content) / 4 // Rough token estimate
	}
	return totalTokens
}

