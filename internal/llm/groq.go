package llm

import (
	"context"
	"errors"

	"github.com/conneroisu/groq-go"
)

// GroqClient is a client for the Groq API.
type GroqClient struct {
	client *groq.Client
}

// NewGroqClient creates a new Groq client.
// It requires an API key to be provided.
func NewGroqClient(apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, errors.New("Groq API key is required")
	}
	client, err := groq.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	return &GroqClient{client: client}, nil
}

// Complete sends a completion request to the Groq API.
func (c *GroqClient) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.ChatCompletion(
		ctx,
		groq.ChatCompletionRequest{
			Model: groq.ModelLlama38B8192,
			Messages: []groq.ChatCompletionMessage{
				{
					Role:    groq.RoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices from Groq")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateEmbedding is not supported by the Groq client yet.
func (c *GroqClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("embedding generation is not supported by the Groq client")
}
