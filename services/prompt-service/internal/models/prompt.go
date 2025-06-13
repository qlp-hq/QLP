package models

import (
	"time"

	"github.com/google/uuid"
)

type Prompt struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	TaskType   string    `json:"task_type"`
	PromptText string    `json:"prompt_text"`
	Version    int       `json:"version"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
