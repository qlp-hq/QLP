package contracts

import (
	"time"
)

// ExecuteWorkflowRequest represents a request to execute a workflow
type ExecuteWorkflowRequest struct {
	IntentID      string                 `json:"intent_id"`
	IntentText    string                 `json:"intent_text"`
	Tasks         []Task                 `json:"tasks"`
	Dependencies  []TaskDependency       `json:"dependencies"`
	Configuration WorkflowConfiguration  `json:"configuration"`
	Context       ProjectContext         `json:"context"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Task represents a single task in the workflow
type Task struct {
	ID           string            `json:"id"`
	Type         TaskType          `json:"type"`
	Description  string            `json:"description"`
	Dependencies []string          `json:"dependencies"`
	Priority     Priority          `json:"priority"`
	AgentType    string            `json:"agent_type"`
	Parameters   map[string]interface{} `json:"parameters"`
	Timeout      time.Duration     `json:"timeout"`
	Retries      int              `json:"retries"`
	Metadata     map[string]string `json:"metadata"`
	Status       TaskStatus        `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

type TaskType string

const (
	TaskTypeCodegen TaskType = "codegen"
	TaskTypeInfra   TaskType = "infra"
	TaskTypeDoc     TaskType = "doc"
	TaskTypeTest    TaskType = "test"
	TaskTypeAnalyze TaskType = "analyze"
	TaskTypeDeploy  TaskType = "deploy"
	TaskTypeValidate TaskType = "validate"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusRetrying   TaskStatus = "retrying"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "prerequisite", "data", "resource"
}

// WorkflowConfiguration contains workflow execution settings
type WorkflowConfiguration struct {
	MaxConcurrency     int           `json:"max_concurrency"`
	Timeout            time.Duration `json:"timeout"`
	FailurePolicy      string        `json:"failure_policy"` // "abort", "continue", "retry"
	RetryPolicy        RetryPolicy   `json:"retry_policy"`
	NotificationPolicy string        `json:"notification_policy"`
	EnableHITL         bool          `json:"enable_hitl"`
	ValidationLevel    string        `json:"validation_level"` // "none", "basic", "strict"
}

type RetryPolicy struct {
	MaxRetries   int           `json:"max_retries"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	BackoffFactor float64      `json:"backoff_factor"`
}

// ProjectContext provides context for workflow execution
type ProjectContext struct {
	ProjectType   string            `json:"project_type"`
	TechStack     []string          `json:"tech_stack"`
	Requirements  []string          `json:"requirements"`
	Constraints   map[string]string `json:"constraints"`
	Environment   string            `json:"environment"`
	Region        string            `json:"region"`
	Budget        *Budget           `json:"budget,omitempty"`
}

type Budget struct {
	MaxCost  float64 `json:"max_cost"`
	Currency string  `json:"currency"`
	Period   string  `json:"period"`
}

// WorkflowExecution represents the execution of a workflow
type WorkflowExecution struct {
	ID            string                     `json:"id"`
	IntentID      string                     `json:"intent_id"`
	Status        WorkflowStatus             `json:"status"`
	Tasks         []TaskExecution            `json:"tasks"`
	Dependencies  []TaskDependency           `json:"dependencies"`
	Configuration WorkflowConfiguration      `json:"configuration"`
	Context       ProjectContext             `json:"context"`
	Progress      WorkflowProgress           `json:"progress"`
	Results       map[string]TaskResult      `json:"results"`
	Errors        []WorkflowError            `json:"errors"`
	CreatedAt     time.Time                  `json:"created_at"`
	StartedAt     *time.Time                 `json:"started_at,omitempty"`
	CompletedAt   *time.Time                 `json:"completed_at,omitempty"`
	Duration      time.Duration              `json:"duration"`
	Metadata      map[string]interface{}     `json:"metadata"`
}

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusPaused    WorkflowStatus = "paused"
)

// TaskExecution represents the execution state of a task
type TaskExecution struct {
	Task        Task               `json:"task"`
	Status      TaskStatus         `json:"status"`
	AgentID     string            `json:"agent_id"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Duration    time.Duration     `json:"duration"`
	Attempts    int               `json:"attempts"`
	Result      *TaskResult       `json:"result,omitempty"`
	Error       *TaskError        `json:"error,omitempty"`
	Logs        []string          `json:"logs"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Output           string                 `json:"output"`
	Files            map[string]string      `json:"files"`
	Artifacts        []string               `json:"artifacts"`
	Metrics          map[string]interface{} `json:"metrics"`
	ValidationResult *ValidationResult      `json:"validation_result,omitempty"`
	SandboxResult    *SandboxResult         `json:"sandbox_result,omitempty"`
}

type ValidationResult struct {
	Status   string                 `json:"status"`
	Score    int                    `json:"score"`
	Issues   []ValidationIssue      `json:"issues"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type SandboxResult struct {
	ExitCode      int               `json:"exit_code"`
	Stdout        string            `json:"stdout"`
	Stderr        string            `json:"stderr"`
	ExecutionTime time.Duration     `json:"execution_time"`
	ResourceUsage ResourceUsage     `json:"resource_usage"`
	Artifacts     []string          `json:"artifacts"`
}

type ResourceUsage struct {
	CPUTime    time.Duration `json:"cpu_time"`
	MemoryUsed int64         `json:"memory_used"`
	DiskUsed   int64         `json:"disk_used"`
	NetworkIO  int64         `json:"network_io"`
}

// TaskError represents an error during task execution
type TaskError struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Code        string    `json:"code"`
	Details     string    `json:"details"`
	Recoverable bool      `json:"recoverable"`
	Timestamp   time.Time `json:"timestamp"`
}

// WorkflowProgress represents the progress of workflow execution
type WorkflowProgress struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	FailedTasks     int     `json:"failed_tasks"`
	RunningTasks    int     `json:"running_tasks"`
	PendingTasks    int     `json:"pending_tasks"`
	PercentComplete float64 `json:"percent_complete"`
	EstimatedTimeLeft time.Duration `json:"estimated_time_left"`
}

// WorkflowError represents an error during workflow execution
type WorkflowError struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	TaskID    string    `json:"task_id,omitempty"`
	Code      string    `json:"code"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// Response types
type ExecuteWorkflowResponse struct {
	WorkflowID string           `json:"workflow_id"`
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	Execution  *WorkflowExecution `json:"execution,omitempty"`
}

type GetWorkflowResponse struct {
	WorkflowID string             `json:"workflow_id"`
	Execution  *WorkflowExecution `json:"execution"`
}

type ListWorkflowsResponse struct {
	Workflows []WorkflowSummary `json:"workflows"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
}

type WorkflowSummary struct {
	ID          string         `json:"id"`
	IntentID    string         `json:"intent_id"`
	Status      WorkflowStatus `json:"status"`
	Progress    WorkflowProgress `json:"progress"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Duration    time.Duration  `json:"duration"`
}

// Control operations
type PauseWorkflowRequest struct {
	Reason string `json:"reason,omitempty"`
}

type ResumeWorkflowRequest struct {
	Reason string `json:"reason,omitempty"`
}

type CancelWorkflowRequest struct {
	Reason string `json:"reason,omitempty"`
	Force  bool   `json:"force"`
}

type RetryTaskRequest struct {
	TaskID string `json:"task_id"`
	Reason string `json:"reason,omitempty"`
}

// Event types for workflow execution
type WorkflowEvent struct {
	Type        string                 `json:"type"`
	WorkflowID  string                 `json:"workflow_id"`
	TaskID      string                 `json:"task_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

const (
	EventWorkflowStarted   = "workflow.started"
	EventWorkflowCompleted = "workflow.completed"
	EventWorkflowFailed    = "workflow.failed"
	EventWorkflowPaused    = "workflow.paused"
	EventWorkflowResumed   = "workflow.resumed"
	EventWorkflowCancelled = "workflow.cancelled"
	EventTaskStarted       = "task.started"
	EventTaskCompleted     = "task.completed"
	EventTaskFailed        = "task.failed"
	EventTaskRetrying      = "task.retrying"
)

// DAG (Directed Acyclic Graph) operations
type DAGValidationRequest struct {
	Tasks        []Task           `json:"tasks"`
	Dependencies []TaskDependency `json:"dependencies"`
}

type DAGValidationResponse struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	ExecutionOrder []string `json:"execution_order,omitempty"`
}

// Metrics and monitoring
type WorkflowMetrics struct {
	WorkflowID       string            `json:"workflow_id"`
	ExecutionTime    time.Duration     `json:"execution_time"`
	TaskMetrics      []TaskMetrics     `json:"task_metrics"`
	ResourceUsage    ResourceUsage     `json:"resource_usage"`
	ErrorRate        float64           `json:"error_rate"`
	ThroughputTasks  float64           `json:"throughput_tasks_per_hour"`
	CustomMetrics    map[string]interface{} `json:"custom_metrics"`
}

type TaskMetrics struct {
	TaskID        string            `json:"task_id"`
	ExecutionTime time.Duration     `json:"execution_time"`
	ResourceUsage ResourceUsage     `json:"resource_usage"`
	RetryCount    int               `json:"retry_count"`
	CustomMetrics map[string]interface{} `json:"custom_metrics"`
}