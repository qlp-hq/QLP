package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	anthropicModel  = "claude-3-opus-20240229"
)

// AnthropicClient is a client for the Anthropic API.
type AnthropicClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client.
func NewAnthropicClient(apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, errors.New("Anthropic API key is required")
	}
	return &AnthropicClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}, nil
}

type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Complete sends a completion request to the Anthropic API.
func (c *AnthropicClient) Complete(ctx context.Context, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     anthropicModel,
		MaxTokens: 1024,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.New("failed to get response from Anthropic: " + string(bodyBytes))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		return "", errors.New("no response choices from Anthropic")
	}

	return anthropicResp.Content[0].Text, nil
}

// GenerateEmbedding is not supported by the Anthropic client yet.
func (c *AnthropicClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("embedding generation is not supported by the Anthropic client")
}
