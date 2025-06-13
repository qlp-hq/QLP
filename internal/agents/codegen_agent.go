package agents

import (
	"context"
	"fmt"
	"time"

	"QLP/internal/llm"
	"QLP/internal/models"
)

// CodeGenAgent is a specialized agent for generating source code.
type CodeGenAgent struct {
	llmClient llm.Client
}

// NewCodeGenAgent creates a new code generation agent.
func NewCodeGenAgent(llmClient llm.Client) *CodeGenAgent {
	return &CodeGenAgent{
		llmClient: llmClient,
	}
}

// Execute performs the code generation task.
func (a *CodeGenAgent) Execute(ctx context.Context, task models.Task) (*models.Artifact, error) {
	prompt := a.buildPrompt(task)

	code, err := a.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	artifact := &models.Artifact{
		ID:      fmt.Sprintf("ART-%d", time.Now().UnixNano()),
		Task:    task,
		Type:    models.ArtifactTypeSourceCode,
		Content: code,
		Metadata: map[string]string{
			"language": task.Language,
			"model":    task.Model,
		},
		CreatedAt: time.Now(),
	}

	return artifact, nil
}

func (a *CodeGenAgent) buildPrompt(task models.Task) string {
	return fmt.Sprintf(
		"You are a world-class software engineer specializing in %s.\n"+
			"Your task is to write clean, efficient, and production-ready code for the following requirement:\n\n"+
			"Requirement: %s\n\n"+
			"Please provide only the raw source code for the solution. Do not include any explanations, markdown formatting, or any text other than the code itself.",
		task.Language,
		task.Description,
	)
}
