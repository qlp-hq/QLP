package contracts

import (
	"time"
)

// WorkerTask represents a task that can be executed by the worker runtime
type WorkerTask struct {
	ID              string            `json:"id"`
	Type            TaskType          `json:"type"`
	Description     string            `json:"description"`
	Code            string            `json:"code,omitempty"`
	Language        string            `json:"language,omitempty"`
	Dependencies    []string          `json:"dependencies"`
	Priority        Priority          `json:"priority"`
	Metadata        map[string]string `json:"metadata"`
	ResourceLimits  *ResourceLimits   `json:"resource_limits,omitempty"`
	TenantID        string            `json:"tenant_id"`
	CreatedAt       time.Time         `json:"created_at"`
	TimeoutSeconds  int               `json:"timeout_seconds,omitempty"`
}

type TaskType string

const (
	TaskTypeCodegen   TaskType = "codegen"
	TaskTypeInfra     TaskType = "infra"
	TaskTypeDoc       TaskType = "doc"
	TaskTypeTest      TaskType = "test"
	TaskTypeAnalyze   TaskType = "analyze"
	TaskTypeValidate  TaskType = "validate"
	TaskTypePackage   TaskType = "package"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type ResourceLimits struct {
	CPUMillicores int64  `json:"cpu_millicores"`    // 1000 = 1 CPU
	MemoryMB      int64  `json:"memory_mb"`         // Memory limit in MB
	TimeoutSec    int    `json:"timeout_sec"`       // Execution timeout
	NetworkAccess bool   `json:"network_access"`    // Allow network access
	FileSystemRW  bool   `json:"filesystem_rw"`     // Allow file system writes
}

// WorkerExecution represents the execution of a task
type WorkerExecution struct {
	ID                string              `json:"id"`
	TaskID            string              `json:"task_id"`
	TenantID          string              `json:"tenant_id"`
	AgentID           string              `json:"agent_id,omitempty"`
	Status            ExecutionStatus     `json:"status"`
	Output            string              `json:"output,omitempty"`
	Error             string              `json:"error,omitempty"`
	StartTime         time.Time           `json:"start_time"`
	EndTime           *time.Time          `json:"end_time,omitempty"`
	ExecutionTime     time.Duration       `json:"execution_time"`
	SandboxResult     *SandboxResult      `json:"sandbox_result,omitempty"`
	ValidationResult  *ValidationResult   `json:"validation_result,omitempty"`
	ResourceUsage     *ResourceUsage      `json:"resource_usage,omitempty"`
}

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
	ExecutionStatusCanceled  ExecutionStatus = "canceled"
)

// SandboxResult contains the results of sandbox execution
type SandboxResult struct {
	ExitCode        int               `json:"exit_code"`
	Stdout          string            `json:"stdout"`
	Stderr          string            `json:"stderr"`
	ExecutionTime   time.Duration     `json:"execution_time"`
	ResourceUsage   *ResourceUsage    `json:"resource_usage"`
	Files           []FileOutput      `json:"files,omitempty"`
	NetworkCalls    []NetworkCall     `json:"network_calls,omitempty"`
	SecurityViolations []SecurityViolation `json:"security_violations,omitempty"`
}

type ResourceUsage struct {
	CPUTimeMs    int64 `json:"cpu_time_ms"`
	MemoryPeakMB int64 `json:"memory_peak_mb"`
	DiskReadMB   int64 `json:"disk_read_mb"`
	DiskWriteMB  int64 `json:"disk_write_mb"`
	NetworkInMB  int64 `json:"network_in_mb"`
	NetworkOutMB int64 `json:"network_out_mb"`
}

type FileOutput struct {
	Path        string `json:"path"`
	Content     string `json:"content,omitempty"`
	Size        int64  `json:"size"`
	Hash        string `json:"hash"`
	IsDirectory bool   `json:"is_directory"`
}

type NetworkCall struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	StatusCode int               `json:"status_code"`
	Duration   time.Duration     `json:"duration"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type SecurityViolation struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// ValidationResult represents validation outcome
type ValidationResult struct {
	OverallScore    int                   `json:"overall_score"`
	SecurityScore   int                   `json:"security_score"`
	QualityScore    int                   `json:"quality_score"`
	Passed          bool                  `json:"passed"`
	Issues          []ValidationIssue     `json:"issues,omitempty"`
	Warnings        []ValidationWarning   `json:"warnings,omitempty"`
	ValidationTime  time.Duration         `json:"validation_time"`
}

type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type ValidationWarning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// API Request/Response types
type ExecuteTaskRequest struct {
	Task            WorkerTask      `json:"task"`
	ValidateOutput  bool            `json:"validate_output"`
	ReturnFiles     bool            `json:"return_files"`
	StreamOutput    bool            `json:"stream_output,omitempty"`
}

type ExecuteTaskResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

type GetExecutionRequest struct {
	ExecutionID string `json:"execution_id"`
	TenantID    string `json:"tenant_id"`
}

type GetExecutionResponse struct {
	Execution WorkerExecution `json:"execution"`
}

type ListExecutionsRequest struct {
	TenantID   string          `json:"tenant_id"`
	Status     ExecutionStatus `json:"status,omitempty"`
	TaskType   TaskType        `json:"task_type,omitempty"`
	Limit      int             `json:"limit,omitempty"`
	Offset     int             `json:"offset,omitempty"`
	Since      *time.Time      `json:"since,omitempty"`
}

type ListExecutionsResponse struct {
	Executions []WorkerExecution `json:"executions"`
	Total      int               `json:"total"`
}

type CancelExecutionRequest struct {
	ExecutionID string `json:"execution_id"`
	TenantID    string `json:"tenant_id"`
	Reason      string `json:"reason,omitempty"`
}

type CancelExecutionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Agent context for task execution
type AgentContext struct {
	ProjectType     string            `json:"project_type"`
	TechStack       []string          `json:"tech_stack"`
	Requirements    []string          `json:"requirements"`
	Constraints     map[string]string `json:"constraints"`
	Architecture    string            `json:"architecture"`
	PreviousOutputs []PreviousOutput  `json:"previous_outputs,omitempty"`
}

type PreviousOutput struct {
	TaskID      string    `json:"task_id"`
	TaskType    TaskType  `json:"task_type"`
	Output      string    `json:"output"`
	CompletedAt time.Time `json:"completed_at"`
}

// Streaming response for real-time updates
type ExecutionUpdate struct {
	ExecutionID   string          `json:"execution_id"`
	Status        ExecutionStatus `json:"status"`
	Output        string          `json:"output,omitempty"`
	Error         string          `json:"error,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
	ProgressPct   int             `json:"progress_pct,omitempty"`
}