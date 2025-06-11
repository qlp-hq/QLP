package types

import (
	"time"
)

// Task represents a task in the QLP system
type Task struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Status      TaskStatus        `json:"status"`
	Priority    TaskPriority      `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the priority of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID       string                 `json:"task_id"`
	AgentID      string                 `json:"agent_id"`
	Status       TaskStatus             `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	Output       string                 `json:"output"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Attachments  map[string][]byte      `json:"attachments,omitempty"`
	Validation   *ValidationResult      `json:"validation,omitempty"`
}

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeCodegen   TaskType = "codegen"
	TaskTypeInfra     TaskType = "infra"
	TaskTypeTest      TaskType = "test"
	TaskTypeDoc       TaskType = "doc"
	TaskTypeAnalyze   TaskType = "analyze"
	TaskTypeValidate  TaskType = "validate"
	TaskTypeDeploy    TaskType = "deploy"
)