package models

import (
	"time"
)

type Intent struct {
	ID              string            `json:"id"`
	UserInput       string            `json:"user_input"`
	Tasks           []Task            `json:"tasks"`
	Language        string            `json:"language"`
	Metadata        map[string]string `json:"metadata"`
	Status          IntentStatus      `json:"status"`
	OverallScore    int               `json:"overall_score"`
	ExecutionTimeMS int               `json:"execution_time_ms"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

type IntentStatus string

const (
	IntentStatusPending    IntentStatus = "pending"
	IntentStatusProcessing IntentStatus = "processing"
	IntentStatusCompleted  IntentStatus = "completed"
	IntentStatusFailed     IntentStatus = "failed"
)

type Task struct {
	ID           string            `json:"id"`
	IntentID     string            `json:"intent_id"`
	Type         TaskType          `json:"type"`
	Description  string            `json:"description"`
	Language     string            `json:"language"`
	Model        string            `json:"model,omitempty"`
	Dependencies []string          `json:"dependencies"`
	Priority     Priority          `json:"priority"`
	Metadata     map[string]string `json:"metadata"`
	Status       TaskStatus        `json:"status"`
	AgentID      string            `json:"agent_id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

type TaskType string

const (
	TaskTypeCodegen TaskType = "codegen"
	TaskTypeInfra   TaskType = "infra"
	TaskTypeDoc     TaskType = "doc"
	TaskTypeTest    TaskType = "test"
	TaskTypeAnalyze TaskType = "analyze"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)
