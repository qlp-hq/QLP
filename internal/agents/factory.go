package agents

import (
	"fmt"

	"QLP/internal/llm"
	"QLP/internal/models"
	promptclient "QLP/services/prompt-service/pkg/client"
)

// AgentFactory is responsible for creating agent instances.
type AgentFactory struct {
	llmClients    map[string]llm.Client
	defaultClient llm.Client
	promptClient  *promptclient.PromptServiceClient
}

// NewAgentFactory creates a new agent factory.
func NewAgentFactory(clients map[string]llm.Client, defaultClient llm.Client, promptClient *promptclient.PromptServiceClient) (*AgentFactory, error) {
	if len(clients) == 0 {
		return nil, fmt.Errorf("at least one LLM client is required")
	}
	if defaultClient == nil {
		return nil, fmt.Errorf("a default LLM client is required")
	}
	return &AgentFactory{
		llmClients:    clients,
		defaultClient: defaultClient,
		promptClient:  promptClient,
	}, nil
}

// GetAgent returns the appropriate agent for a given task.
func (f *AgentFactory) GetAgent(task models.Task) (Agent, error) {
	client := f.getClientForTask(task)

	switch task.Type {
	case models.TaskTypeCodegen:
		return NewCodeGenAgent(client), nil
	default:
		// The DynamicAgent can handle any task type by fetching the right prompt
		// from the prompt-service. We just provide an empty context for now.
		return NewDynamicAgent(task, client, nil, AgentContext{}, f.promptClient), nil
	}
}

func (f *AgentFactory) getClientForTask(task models.Task) llm.Client {
	if task.Model != "" {
		if client, ok := f.llmClients[task.Model]; ok {
			return client
		}
	}
	return f.defaultClient
}
