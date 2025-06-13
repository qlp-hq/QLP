package llm

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient is a client for the OpenAI API.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client.
// It requires an API key to be provided.
func NewOpenAIClient(apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIClient{client: client}, nil
}

// Complete sends a completion request to the OpenAI API.
func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4TurboPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateEmbedding creates a vector embedding for the given text.
func (c *OpenAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}

	res, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, errors.New("no embedding data returned")
	}

	return res.Data[0].Embedding, nil
}
