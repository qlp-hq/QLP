package contracts

import (
	"time"
)

// Intent represents the shared intent contract between services
type Intent struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	UserInput       string            `json:"user_input"`
	ParsedTasks     []Task            `json:"parsed_tasks,omitempty"`
	Metadata        map[string]string `json:"metadata"`
	Status          IntentStatus      `json:"status"`
	OverallScore    int               `json:"overall_score,omitempty"`
	ExecutionTimeMS int               `json:"execution_time_ms,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

type Task struct {
	ID           string            `json:"id"`
	Type         TaskType          `json:"type"`
	Description  string            `json:"description"`
	Dependencies []string          `json:"dependencies"`
	Priority     Priority          `json:"priority"`
	Metadata     map[string]string `json:"metadata"`
	Status       TaskStatus        `json:"status"`
	AgentID      string            `json:"agent_id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

type IntentStatus string

const (
	IntentStatusPending    IntentStatus = "pending"
	IntentStatusProcessing IntentStatus = "processing"
	IntentStatusCompleted  IntentStatus = "completed"
	IntentStatusFailed     IntentStatus = "failed"
)

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

// Vector search contracts
type VectorSimilarRequest struct {
	TenantID  string    `json:"tenant_id"`
	Query     string    `json:"query"`
	Embedding []float64 `json:"embedding,omitempty"`
	Limit     int       `json:"limit"`
	Threshold float64   `json:"threshold"`
}

type VectorSimilarResponse struct {
	Results []SimilarIntent `json:"results"`
}

type SimilarIntent struct {
	Intent     Intent  `json:"intent"`
	Similarity float64 `json:"similarity"`
}

type EmbeddingRequest struct {
	TenantID string `json:"tenant_id"`
	Text     string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// Tenant context for multi-tenancy
type TenantContext struct {
	TenantID    string            `json:"tenant_id"`
	UserID      string            `json:"user_id,omitempty"`
	OrgID       string            `json:"org_id,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// API request/response wrappers
type CreateIntentRequest struct {
	UserInput string            `json:"user_input"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type CreateIntentResponse struct {
	Intent Intent `json:"intent"`
}

type ListIntentsRequest struct {
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type ListIntentsResponse struct {
	Intents []Intent `json:"intents"`
	Total   int      `json:"total"`
}

type UpdateIntentRequest struct {
	Status          *IntentStatus     `json:"status,omitempty"`
	ParsedTasks     []Task            `json:"parsed_tasks,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	OverallScore    *int              `json:"overall_score,omitempty"`
	ExecutionTimeMS *int              `json:"execution_time_ms,omitempty"`
}