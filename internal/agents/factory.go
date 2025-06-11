package agents

import (
	"context"
	"fmt"
	"sync"

	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/models"
	"go.uber.org/zap"
)

type AgentFactory struct {
	llmClient      llm.Client
	eventBus       *events.EventBus
	activeAgents   map[string]*DynamicAgent
	agentOutputs   map[string]string
	mu             sync.RWMutex
	contextBuilder *ContextBuilder
}

func NewAgentFactory(llmClient llm.Client, eventBus *events.EventBus) *AgentFactory {
	return &AgentFactory{
		llmClient:      llmClient,
		eventBus:       eventBus,
		activeAgents:   make(map[string]*DynamicAgent),
		agentOutputs:   make(map[string]string),
		contextBuilder: NewContextBuilder(),
	}
}

func (af *AgentFactory) CreateAgent(ctx context.Context, task models.Task, projectContext ProjectContext) (*DynamicAgent, error) {
	logger.WithComponent("agents").Info("Creating dynamic agent",
		zap.String("task_id", task.ID),
		zap.String("task_type", string(task.Type)))

	agentContext := af.contextBuilder.BuildAgentContext(task, projectContext, af.agentOutputs)

	agent := NewDynamicAgent(task, af.llmClient, af.eventBus, agentContext)

	if err := agent.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize agent: %w", err)
	}

	af.mu.Lock()
	af.activeAgents[agent.ID] = agent
	af.mu.Unlock()

	logger.WithComponent("agents").Info("Agent created and initialized",
		zap.String("agent_id", agent.ID),
		zap.String("task_id", task.ID))
	return agent, nil
}

func (af *AgentFactory) ExecuteAgent(ctx context.Context, agent *DynamicAgent) error {
	if err := agent.Execute(ctx); err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	af.mu.Lock()
	af.agentOutputs[agent.Task.ID] = agent.GetOutput()
	af.mu.Unlock()

	return nil
}

func (af *AgentFactory) GetAgentOutput(taskID string) (string, bool) {
	af.mu.RLock()
	defer af.mu.RUnlock()

	output, exists := af.agentOutputs[taskID]
	return output, exists
}

func (af *AgentFactory) GetActiveAgents() map[string]*DynamicAgent {
	af.mu.RLock()
	defer af.mu.RUnlock()

	agents := make(map[string]*DynamicAgent)
	for id, agent := range af.activeAgents {
		agents[id] = agent
	}

	return agents
}

func (af *AgentFactory) CleanupAgent(agentID string) {
	af.mu.Lock()
	defer af.mu.Unlock()

	if agent, exists := af.activeAgents[agentID]; exists {
		logger.WithComponent("agents").Info("Cleaning up agent",
			zap.String("agent_id", agentID))

		af.eventBus.Publish(events.Event{
			ID:     fmt.Sprintf("agent_%s_cleanup", agentID),
			Type:   events.EventAgentStopped,
			Source: agentID,
			Payload: map[string]interface{}{
				"agent_id": agentID,
				"task_id":  agent.Task.ID,
				"status":   agent.GetStatus(),
			},
		})

		delete(af.activeAgents, agentID)
	}
}

type ProjectContext struct {
	ProjectType  string            `json:"project_type"`
	TechStack    []string          `json:"tech_stack"`
	Requirements []string          `json:"requirements"`
	Constraints  map[string]string `json:"constraints"`
	Architecture string            `json:"architecture"`
}

type ContextBuilder struct{}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{}
}

func (cb *ContextBuilder) BuildAgentContext(task models.Task, projectContext ProjectContext, previousOutputs map[string]string) AgentContext {
	dependencyOutputs := make(map[string]string)

	for _, depID := range task.Dependencies {
		if output, exists := previousOutputs[depID]; exists {
			dependencyOutputs[depID] = output
		}
	}

	outputRequirements := cb.inferOutputRequirements(task)
	constraints := cb.mergeConstraints(projectContext.Constraints, task)

	return AgentContext{
		ProjectType:        projectContext.ProjectType,
		TechStack:          projectContext.TechStack,
		Dependencies:       cb.getDependencyTasks(task),
		OutputRequirements: outputRequirements,
		Constraints:        constraints,
		PreviousOutputs:    dependencyOutputs,
	}
}

func (cb *ContextBuilder) inferOutputRequirements(task models.Task) []string {
	switch task.Type {
	case models.TaskTypeCodegen:
		return []string{
			"Complete, executable code",
			"Proper error handling",
			"Unit tests included",
			"Documentation comments",
			"Following best practices",
		}
	case models.TaskTypeInfra:
		return []string{
			"Infrastructure as Code files",
			"Deployment scripts",
			"Configuration files",
			"Security policies",
			"Monitoring setup",
		}
	case models.TaskTypeDoc:
		return []string{
			"Comprehensive documentation",
			"Code examples",
			"Setup instructions",
			"API references",
			"Troubleshooting guide",
		}
	case models.TaskTypeTest:
		return []string{
			"Test suite with good coverage",
			"Unit and integration tests",
			"Test data setup",
			"Assertion descriptions",
			"Performance benchmarks",
		}
	case models.TaskTypeAnalyze:
		return []string{
			"Detailed analysis report",
			"Data visualizations",
			"Actionable recommendations",
			"Risk assessment",
			"Implementation roadmap",
		}
	default:
		return []string{
			"High-quality deliverable",
			"Complete solution",
			"Proper documentation",
		}
	}
}

func (cb *ContextBuilder) mergeConstraints(projectConstraints map[string]string, task models.Task) map[string]string {
	constraints := make(map[string]string)

	for k, v := range projectConstraints {
		constraints[k] = v
	}

	constraints["task_priority"] = string(task.Priority)
	constraints["task_type"] = string(task.Type)

	return constraints
}

func (cb *ContextBuilder) getDependencyTasks(task models.Task) []models.Task {
	return []models.Task{}
}
