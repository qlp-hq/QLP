package contracts

import (
	"fmt"
	"time"
)

// CreateAgentRequest represents a request to create a new agent
type CreateAgentRequest struct {
	TaskID          string            `json:"task_id"`
	TaskType        string            `json:"task_type"`
	TaskDescription string            `json:"task_description"`
	Priority        string            `json:"priority"`
	Dependencies    []string          `json:"dependencies"`
	ProjectContext  ProjectContext    `json:"project_context"`
	Configuration   AgentConfig       `json:"configuration"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// ExecuteAgentRequest represents a request to execute an agent
type ExecuteAgentRequest struct {
	AgentID   string            `json:"agent_id"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AgentConfig represents configuration for agent behavior
type AgentConfig struct {
	EnableSandbox    bool          `json:"enable_sandbox"`
	EnableValidation bool          `json:"enable_validation"`
	Timeout          time.Duration `json:"timeout"`
	MaxRetries       int           `json:"max_retries"`
	LLMProvider      string        `json:"llm_provider,omitempty"`
	SandboxConfig    SandboxConfig `json:"sandbox_config"`
	ValidationConfig ValidationConfig `json:"validation_config"`
}

// SandboxConfig represents sandbox execution configuration
type SandboxConfig struct {
	Enabled       bool          `json:"enabled"`
	TimeLimit     time.Duration `json:"time_limit"`
	MemoryLimit   int64         `json:"memory_limit_mb"`
	NetworkAccess bool          `json:"network_access"`
	AllowedPorts  []int         `json:"allowed_ports"`
}

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	Enabled        bool     `json:"enabled"`
	RequiredChecks []string `json:"required_checks"`
	MinScore       int      `json:"min_score"`
	SecurityLevel  string   `json:"security_level"`
	QualityLevel   string   `json:"quality_level"`
}

// ProjectContext provides context for agent execution
type ProjectContext struct {
	ProjectType     string            `json:"project_type"`
	TechStack       []string          `json:"tech_stack"`
	Requirements    []string          `json:"requirements"`
	Constraints     map[string]string `json:"constraints"`
	Architecture    string            `json:"architecture"`
	PreviousOutputs map[string]string `json:"previous_outputs,omitempty"`
}

// Agent represents an agent instance
type Agent struct {
	ID              string            `json:"id"`
	TaskID          string            `json:"task_id"`
	TaskType        string            `json:"task_type"`
	TaskDescription string            `json:"task_description"`
	Status          AgentStatus       `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	StartedAt       *time.Time        `json:"started_at,omitempty"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
	Duration        time.Duration     `json:"duration"`
	Output          string            `json:"output"`
	Error           string            `json:"error,omitempty"`
	Metrics         AgentMetrics      `json:"metrics"`
	Configuration   AgentConfig       `json:"configuration"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusInitializing AgentStatus = "initializing"
	AgentStatusReady        AgentStatus = "ready"
	AgentStatusExecuting    AgentStatus = "executing"
	AgentStatusCompleted    AgentStatus = "completed"
	AgentStatusFailed       AgentStatus = "failed"
	AgentStatusCancelled    AgentStatus = "cancelled"
	AgentStatusTimeout      AgentStatus = "timeout"
)

// AgentMetrics represents metrics for an agent execution
type AgentMetrics struct {
	LLMTokensUsed       int           `json:"llm_tokens_used"`
	LLMResponseTime     time.Duration `json:"llm_response_time"`
	SandboxExecutionTime time.Duration `json:"sandbox_execution_time"`
	ValidationTime      time.Duration `json:"validation_time"`
	TotalExecutionTime  time.Duration `json:"total_execution_time"`
	ValidationScore     int           `json:"validation_score"`
	SecurityScore       int           `json:"security_score"`
	QualityScore        int           `json:"quality_score"`
	MemoryUsed          int64         `json:"memory_used_mb"`
	CPUTime             time.Duration `json:"cpu_time"`
}

// Response types

// CreateAgentResponse represents the response from creating an agent
type CreateAgentResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Agent   *Agent `json:"agent,omitempty"`
}

// ExecuteAgentResponse represents the response from executing an agent
type ExecuteAgentResponse struct {
	AgentID     string       `json:"agent_id"`
	Status      string       `json:"status"`
	Message     string       `json:"message"`
	ExecutionID string       `json:"execution_id"`
	Agent       *Agent       `json:"agent,omitempty"`
}

// GetAgentResponse represents the response from getting an agent
type GetAgentResponse struct {
	Agent *Agent `json:"agent"`
}

// ListAgentsResponse represents the response from listing agents
type ListAgentsResponse struct {
	Agents   []AgentSummary `json:"agents"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// AgentSummary represents a summary of an agent
type AgentSummary struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"task_id"`
	TaskType        string        `json:"task_type"`
	Status          AgentStatus   `json:"status"`
	CreatedAt       time.Time     `json:"created_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	Duration        time.Duration `json:"duration"`
	ValidationScore int           `json:"validation_score"`
}

// Control operations

// CancelAgentRequest represents a request to cancel an agent
type CancelAgentRequest struct {
	Reason string `json:"reason,omitempty"`
	Force  bool   `json:"force"`
}

// RetryAgentRequest represents a request to retry an agent
type RetryAgentRequest struct {
	Reason string `json:"reason,omitempty"`
}

// UpdateAgentConfigRequest represents a request to update agent configuration
type UpdateAgentConfigRequest struct {
	Configuration AgentConfig `json:"configuration"`
}

// Agent Factory operations

// CreateDeploymentValidatorRequest represents a request to create a deployment validator agent
type CreateDeploymentValidatorRequest struct {
	AgentID     string                   `json:"agent_id"`
	CapsuleData DeploymentCapsuleData    `json:"capsule_data"`
	Config      DeploymentValidatorConfig `json:"config"`
	Metadata    map[string]string        `json:"metadata,omitempty"`
}

// DeploymentCapsuleData represents data for deployment validation
type DeploymentCapsuleData struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Files       map[string]string `json:"files"`
	Technologies []string         `json:"technologies"`
	Environment string            `json:"environment"`
}

// DeploymentValidatorConfig represents configuration for deployment validation
type DeploymentValidatorConfig struct {
	AzureConfig           AzureConfig   `json:"azure_config"`
	CostLimitUSD          float64       `json:"cost_limit_usd"`
	TTL                   time.Duration `json:"ttl"`
	EnableHealthChecks    bool          `json:"enable_health_checks"`
	EnableFunctionalTests bool          `json:"enable_functional_tests"`
	CleanupPolicy         string        `json:"cleanup_policy"`
}

// AzureConfig represents Azure-specific configuration
type AzureConfig struct {
	SubscriptionID string `json:"subscription_id"`
	Location       string `json:"location"`
	ResourceGroup  string `json:"resource_group,omitempty"`
}

// DeploymentValidationResult represents the result of deployment validation
type DeploymentValidationResult struct {
	CapsuleID         string                 `json:"capsule_id"`
	ResourceGroup     string                 `json:"resource_group"`
	DeploymentSuccess bool                   `json:"deployment_success"`
	Status            string                 `json:"status"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	Duration          time.Duration          `json:"duration"`
	CostEstimateUSD   float64                `json:"cost_estimate_usd"`
	HealthChecksPassed int                   `json:"health_checks_passed"`
	TotalHealthChecks int                    `json:"total_health_checks"`
	TestsPassed       int                    `json:"tests_passed"`
	TotalTests        int                    `json:"total_tests"`
	AzureLocation     string                 `json:"azure_location"`
	ValidationDetails map[string]interface{} `json:"validation_details"`
}

// Batch operations

// BatchCreateAgentsRequest represents a request to create multiple agents
type BatchCreateAgentsRequest struct {
	Agents   []CreateAgentRequest `json:"agents"`
	Metadata map[string]string    `json:"metadata,omitempty"`
}

// BatchCreateAgentsResponse represents the response from batch agent creation
type BatchCreateAgentsResponse struct {
	BatchID      string                 `json:"batch_id"`
	TotalCount   int                    `json:"total_count"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Results      []CreateAgentResponse  `json:"results"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

// Service status and health

// ServiceStatus represents the status of the agent service
type ServiceStatus struct {
	Status        string            `json:"status"`
	Timestamp     time.Time         `json:"timestamp"`
	ActiveAgents  int               `json:"active_agents"`
	TotalAgents   int               `json:"total_agents"`
	Version       string            `json:"version"`
	Uptime        time.Duration     `json:"uptime"`
	ResourceUsage ResourceUsage     `json:"resource_usage"`
	Dependencies  []DependencyStatus `json:"dependencies"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsedMB  int64   `json:"memory_used_mb"`
	MemoryTotalMB int64   `json:"memory_total_mb"`
	DiskUsedMB    int64   `json:"disk_used_mb"`
	ActiveThreads int     `json:"active_threads"`
}

// DependencyStatus represents the status of a service dependency
type DependencyStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Healthy   bool          `json:"healthy"`
	LastCheck time.Time     `json:"last_check"`
	Response  time.Duration `json:"response_time"`
}

// Metrics and monitoring

// AgentServiceMetrics represents metrics for the agent service
type AgentServiceMetrics struct {
	TotalAgentsCreated    int64             `json:"total_agents_created"`
	TotalAgentsExecuted   int64             `json:"total_agents_executed"`
	TotalAgentsCompleted  int64             `json:"total_agents_completed"`
	TotalAgentsFailed     int64             `json:"total_agents_failed"`
	AverageExecutionTime  time.Duration     `json:"average_execution_time"`
	AverageValidationScore float64          `json:"average_validation_score"`
	AgentsByType          map[string]int64  `json:"agents_by_type"`
	AgentsByStatus        map[string]int64  `json:"agents_by_status"`
	ResourceMetrics       ResourceUsage     `json:"resource_metrics"`
	Uptime                time.Duration     `json:"uptime"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Service     string            `json:"service"`
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version"`
	Checks      map[string]string `json:"checks"`
	ActiveAgents int              `json:"active_agents"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Details   string    `json:"details,omitempty"`
	AgentID   string    `json:"agent_id,omitempty"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Events and notifications

// AgentEvent represents an event related to agent execution
type AgentEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	AgentID   string                 `json:"agent_id"`
	TaskID    string                 `json:"task_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

const (
	EventAgentCreated    = "agent.created"
	EventAgentStarted    = "agent.started"
	EventAgentCompleted  = "agent.completed"
	EventAgentFailed     = "agent.failed"
	EventAgentCancelled  = "agent.cancelled"
	EventAgentTimeout    = "agent.timeout"
	EventValidationDone  = "agent.validation.completed"
	EventSandboxDone     = "agent.sandbox.completed"
)

// Validation

// Validate validates a create agent request
func (r *CreateAgentRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if r.TaskType == "" {
		return fmt.Errorf("task_type is required")
	}
	if r.TaskDescription == "" {
		return fmt.Errorf("task_description is required")
	}
	
	// Validate task type
	validTypes := []string{"codegen", "infra", "test", "doc", "analyze", "deploy", "validate"}
	validType := false
	for _, vt := range validTypes {
		if r.TaskType == vt {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid task_type: %s", r.TaskType)
	}
	
	// Validate priority
	if r.Priority != "" {
		validPriorities := []string{"low", "medium", "high"}
		validPriority := false
		for _, vp := range validPriorities {
			if r.Priority == vp {
				validPriority = true
				break
			}
		}
		if !validPriority {
			return fmt.Errorf("invalid priority: %s", r.Priority)
		}
	}
	
	return nil
}

// Validate validates an execute agent request
func (r *ExecuteAgentRequest) Validate() error {
	if r.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	return nil
}

// Validate validates a batch create agents request
func (r *BatchCreateAgentsRequest) Validate() error {
	if len(r.Agents) == 0 {
		return fmt.Errorf("at least one agent is required")
	}
	if len(r.Agents) > 100 {
		return fmt.Errorf("batch size exceeds limit of 100 agents")
	}
	
	for i, agent := range r.Agents {
		if err := agent.Validate(); err != nil {
			return fmt.Errorf("agent %d: %w", i, err)
		}
	}
	
	return nil
}