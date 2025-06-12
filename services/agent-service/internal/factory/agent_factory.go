package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"QLP/services/agent-service/pkg/contracts"
	"QLP/services/agent-service/internal/engines"
	"QLP/services/llm-service/pkg/client"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// AgentFactory manages agent creation and lifecycle
type AgentFactory struct {
	llmClient       *client.LLMClient
	agents          map[string]*engines.DynamicAgent
	deploymentAgents map[string]*engines.DeploymentValidatorAgent
	agentOutputs    map[string]string
	mu              sync.RWMutex
	metrics         *FactoryMetrics
	startTime       time.Time
}

// FactoryMetrics tracks factory statistics
type FactoryMetrics struct {
	TotalAgentsCreated    int64
	TotalAgentsExecuted   int64
	TotalAgentsCompleted  int64
	TotalAgentsFailed     int64
	AgentsByType          map[string]int64
	AgentsByStatus        map[string]int64
	mu                    sync.RWMutex
}

// NewAgentFactory creates a new agent factory
func NewAgentFactory(llmClient *client.LLMClient) *AgentFactory {
	return &AgentFactory{
		llmClient:        llmClient,
		agents:           make(map[string]*engines.DynamicAgent),
		deploymentAgents: make(map[string]*engines.DeploymentValidatorAgent),
		agentOutputs:     make(map[string]string),
		metrics: &FactoryMetrics{
			AgentsByType:   make(map[string]int64),
			AgentsByStatus: make(map[string]int64),
		},
		startTime: time.Now(),
	}
}

// CreateAgent creates a new dynamic agent
func (af *AgentFactory) CreateAgent(ctx context.Context, req *contracts.CreateAgentRequest) (*contracts.Agent, error) {
	af.mu.Lock()
	defer af.mu.Unlock()

	// Generate agent ID
	agentID := af.generateAgentID(req.TaskType)

	logger.WithComponent("agent-factory").Info("Creating agent",
		zap.String("agent_id", agentID),
		zap.String("task_id", req.TaskID),
		zap.String("task_type", req.TaskType))

	// Create dynamic agent
	dynamicAgent, err := engines.NewDynamicAgent(agentID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic agent: %w", err)
	}

	// Store agent
	af.agents[agentID] = dynamicAgent

	// Update metrics
	af.metrics.mu.Lock()
	af.metrics.TotalAgentsCreated++
	af.metrics.AgentsByType[req.TaskType]++
	af.metrics.AgentsByStatus[string(contracts.AgentStatusInitializing)]++
	af.metrics.mu.Unlock()

	// Convert to contract agent
	agent := af.convertToContractAgent(dynamicAgent, req)

	logger.WithComponent("agent-factory").Info("Agent created successfully",
		zap.String("agent_id", agentID),
		zap.String("status", string(agent.Status)))

	return agent, nil
}

// ExecuteAgent executes an agent
func (af *AgentFactory) ExecuteAgent(ctx context.Context, agentID string, req *contracts.ExecuteAgentRequest) error {
	af.mu.RLock()
	agent, exists := af.agents[agentID]
	af.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	logger.WithComponent("agent-factory").Info("Executing agent",
		zap.String("agent_id", agentID))

	// Update metrics
	af.metrics.mu.Lock()
	af.metrics.TotalAgentsExecuted++
	af.metrics.mu.Unlock()

	// Execute agent
	err := agent.Execute(ctx, af.llmClient)
	if err != nil {
		af.metrics.mu.Lock()
		af.metrics.TotalAgentsFailed++
		af.metrics.AgentsByStatus[string(contracts.AgentStatusFailed)]++
		af.metrics.mu.Unlock()

		logger.WithComponent("agent-factory").Error("Agent execution failed",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// Store output
	af.mu.Lock()
	af.agentOutputs[agent.GetTaskID()] = agent.GetOutput()
	af.mu.Unlock()

	// Update metrics
	af.metrics.mu.Lock()
	af.metrics.TotalAgentsCompleted++
	af.metrics.AgentsByStatus[string(contracts.AgentStatusCompleted)]++
	af.metrics.mu.Unlock()

	logger.WithComponent("agent-factory").Info("Agent execution completed",
		zap.String("agent_id", agentID))

	return nil
}

// GetAgent retrieves an agent by ID
func (af *AgentFactory) GetAgent(agentID string) (*contracts.Agent, error) {
	af.mu.RLock()
	defer af.mu.RUnlock()

	dynamicAgent, exists := af.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	// Convert to contract agent
	agent := af.convertDynamicAgentToContract(dynamicAgent)
	return agent, nil
}

// ListAgents lists agents with filtering
func (af *AgentFactory) ListAgents(page, pageSize int, status string) ([]*contracts.AgentSummary, int) {
	af.mu.RLock()
	defer af.mu.RUnlock()

	var filteredAgents []*engines.DynamicAgent
	for _, agent := range af.agents {
		if status == "" || string(agent.GetStatus()) == status {
			filteredAgents = append(filteredAgents, agent)
		}
	}

	total := len(filteredAgents)

	// Apply pagination
	start := page * pageSize
	end := start + pageSize
	if start >= total {
		return []*contracts.AgentSummary{}, total
	}
	if end > total {
		end = total
	}

	// Convert to summaries
	var summaries []*contracts.AgentSummary
	for i := start; i < end; i++ {
		agent := filteredAgents[i]
		summary := &contracts.AgentSummary{
			ID:              agent.GetID(),
			TaskID:          agent.GetTaskID(),
			TaskType:        agent.GetTaskType(),
			Status:          agent.GetStatus(),
			CreatedAt:       agent.GetCreatedAt(),
			CompletedAt:     agent.GetCompletedAt(),
			Duration:        agent.GetDuration(),
			ValidationScore: agent.GetValidationScore(),
		}
		summaries = append(summaries, summary)
	}

	return summaries, total
}

// CancelAgent cancels an agent execution
func (af *AgentFactory) CancelAgent(agentID string, reason string) error {
	af.mu.Lock()
	defer af.mu.Unlock()

	agent, exists := af.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	err := agent.Cancel(reason)
	if err != nil {
		return fmt.Errorf("failed to cancel agent: %w", err)
	}

	// Update metrics
	af.metrics.mu.Lock()
	af.metrics.AgentsByStatus[string(contracts.AgentStatusCancelled)]++
	af.metrics.mu.Unlock()

	logger.WithComponent("agent-factory").Info("Agent cancelled",
		zap.String("agent_id", agentID),
		zap.String("reason", reason))

	return nil
}

// CleanupAgent removes an agent from memory
func (af *AgentFactory) CleanupAgent(agentID string) {
	af.mu.Lock()
	defer af.mu.Unlock()

	if agent, exists := af.agents[agentID]; exists {
		logger.WithComponent("agent-factory").Info("Cleaning up agent",
			zap.String("agent_id", agentID),
			zap.String("task_id", agent.GetTaskID()))

		delete(af.agents, agentID)
	}
}

// CreateDeploymentValidator creates a deployment validator agent
func (af *AgentFactory) CreateDeploymentValidator(ctx context.Context, req *contracts.CreateDeploymentValidatorRequest) (*contracts.Agent, error) {
	af.mu.Lock()
	defer af.mu.Unlock()

	logger.WithComponent("agent-factory").Info("Creating deployment validator agent",
		zap.String("agent_id", req.AgentID),
		zap.String("capsule_id", req.CapsuleData.ID))

	// Create deployment validator agent
	deploymentAgent, err := engines.NewDeploymentValidatorAgent(req.AgentID, req.CapsuleData, req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment validator agent: %w", err)
	}

	// Store agent
	af.deploymentAgents[req.AgentID] = deploymentAgent

	// Update metrics
	af.metrics.mu.Lock()
	af.metrics.TotalAgentsCreated++
	af.metrics.AgentsByType["deployment_validator"]++
	af.metrics.AgentsByStatus[string(contracts.AgentStatusReady)]++
	af.metrics.mu.Unlock()

	// Convert to contract agent
	agent := af.convertDeploymentAgentToContract(deploymentAgent)

	logger.WithComponent("agent-factory").Info("Deployment validator agent created",
		zap.String("agent_id", req.AgentID))

	return agent, nil
}

// GetActiveAgents returns count of active agents
func (af *AgentFactory) GetActiveAgents() int {
	af.mu.RLock()
	defer af.mu.RUnlock()

	active := 0
	for _, agent := range af.agents {
		status := agent.GetStatus()
		if status == contracts.AgentStatusExecuting || status == contracts.AgentStatusReady {
			active++
		}
	}

	return active
}

// GetMetrics returns factory metrics
func (af *AgentFactory) GetMetrics() *contracts.AgentServiceMetrics {
	af.metrics.mu.RLock()
	defer af.metrics.mu.RUnlock()

	var averageExecutionTime time.Duration
	var totalValidationScore float64
	agentCount := int64(0)

	af.mu.RLock()
	for _, agent := range af.agents {
		if agent.GetStatus() == contracts.AgentStatusCompleted {
			averageExecutionTime += agent.GetDuration()
			totalValidationScore += float64(agent.GetValidationScore())
			agentCount++
		}
	}
	af.mu.RUnlock()

	if agentCount > 0 {
		averageExecutionTime = time.Duration(averageExecutionTime.Nanoseconds() / agentCount)
		totalValidationScore = totalValidationScore / float64(agentCount)
	}

	return &contracts.AgentServiceMetrics{
		TotalAgentsCreated:     af.metrics.TotalAgentsCreated,
		TotalAgentsExecuted:    af.metrics.TotalAgentsExecuted,
		TotalAgentsCompleted:   af.metrics.TotalAgentsCompleted,
		TotalAgentsFailed:      af.metrics.TotalAgentsFailed,
		AverageExecutionTime:   averageExecutionTime,
		AverageValidationScore: totalValidationScore,
		AgentsByType:           af.copyMap(af.metrics.AgentsByType),
		AgentsByStatus:         af.copyMap(af.metrics.AgentsByStatus),
		Uptime:                 time.Since(af.startTime),
	}
}

// Helper methods

func (af *AgentFactory) generateAgentID(taskType string) string {
	typePrefix := map[string]string{
		"infra":    "QLI",
		"codegen":  "QLD",
		"test":     "QLT",
		"doc":      "QLC",
		"analyze":  "QLA",
		"deploy":   "QLP",
		"validate": "QLV",
	}

	prefix, exists := typePrefix[taskType]
	if !exists {
		prefix = "QLG"
	}

	timestamp := time.Now().Format("150405")
	sequence := time.Now().UnixNano() % 1000
	return fmt.Sprintf("%s-AGT-%s-%03d", prefix, timestamp, sequence)
}

func (af *AgentFactory) convertToContractAgent(dynamicAgent *engines.DynamicAgent, req *contracts.CreateAgentRequest) *contracts.Agent {
	return &contracts.Agent{
		ID:              dynamicAgent.GetID(),
		TaskID:          req.TaskID,
		TaskType:        req.TaskType,
		TaskDescription: req.TaskDescription,
		Status:          dynamicAgent.GetStatus(),
		CreatedAt:       dynamicAgent.GetCreatedAt(),
		StartedAt:       dynamicAgent.GetStartedAt(),
		CompletedAt:     dynamicAgent.GetCompletedAt(),
		Duration:        dynamicAgent.GetDuration(),
		Output:          dynamicAgent.GetOutput(),
		Error:           dynamicAgent.GetErrorString(),
		Metrics:         dynamicAgent.GetMetrics(),
		Configuration:   req.Configuration,
		Metadata:        req.Metadata,
	}
}

func (af *AgentFactory) convertDynamicAgentToContract(dynamicAgent *engines.DynamicAgent) *contracts.Agent {
	return &contracts.Agent{
		ID:              dynamicAgent.GetID(),
		TaskID:          dynamicAgent.GetTaskID(),
		TaskType:        dynamicAgent.GetTaskType(),
		TaskDescription: dynamicAgent.GetTaskDescription(),
		Status:          dynamicAgent.GetStatus(),
		CreatedAt:       dynamicAgent.GetCreatedAt(),
		StartedAt:       dynamicAgent.GetStartedAt(),
		CompletedAt:     dynamicAgent.GetCompletedAt(),
		Duration:        dynamicAgent.GetDuration(),
		Output:          dynamicAgent.GetOutput(),
		Error:           dynamicAgent.GetErrorString(),
		Metrics:         dynamicAgent.GetMetrics(),
		Configuration:   dynamicAgent.GetConfiguration(),
		Metadata:        dynamicAgent.GetMetadata(),
	}
}

func (af *AgentFactory) convertDeploymentAgentToContract(deploymentAgent *engines.DeploymentValidatorAgent) *contracts.Agent {
	return &contracts.Agent{
		ID:              deploymentAgent.GetID(),
		TaskID:          deploymentAgent.GetCapsuleID(),
		TaskType:        "deployment_validator",
		TaskDescription: fmt.Sprintf("Deployment validation for capsule %s", deploymentAgent.GetCapsuleID()),
		Status:          deploymentAgent.GetStatus(),
		CreatedAt:       deploymentAgent.GetCreatedAt(),
		StartedAt:       deploymentAgent.GetStartedAt(),
		CompletedAt:     deploymentAgent.GetCompletedAt(),
		Duration:        deploymentAgent.GetDuration(),
		Output:          deploymentAgent.GetOutput(),
		Error:           deploymentAgent.GetErrorString(),
		Metrics:         deploymentAgent.GetMetrics(),
		Metadata:        deploymentAgent.GetMetadata(),
	}
}

func (af *AgentFactory) copyMap(original map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}