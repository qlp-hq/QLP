package repository

import (
	"QLP/internal/database"
	"QLP/services/prompt-service/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PromptRepository struct {
	db *database.Database
}

func NewPromptRepository(db *database.Database) *PromptRepository {
	return &PromptRepository{db: db}
}

func (r *PromptRepository) Create(prompt *models.Prompt) error {
	prompt.ID = uuid.New()
	prompt.CreatedAt = time.Now()
	prompt.UpdatedAt = time.Now()
	prompt.Version = 1
	prompt.IsActive = true

	query := `
		INSERT INTO prompt_registry (id, name, task_type, prompt_text, version, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.GetConnection().Exec(query,
		prompt.ID,
		prompt.Name,
		prompt.TaskType,
		prompt.PromptText,
		prompt.Version,
		prompt.IsActive,
		prompt.CreatedAt,
		prompt.UpdatedAt,
	)
	return err
}

func (r *PromptRepository) GetByID(id uuid.UUID) (*models.Prompt, error) {
	query := `SELECT id, name, task_type, prompt_text, version, is_active, created_at, updated_at FROM prompt_registry WHERE id = $1`
	row := r.db.GetConnection().QueryRow(query, id)
	return r.scanPrompt(row)
}

func (r *PromptRepository) GetByName(name string) (*models.Prompt, error) {
	query := `SELECT id, name, task_type, prompt_text, version, is_active, created_at, updated_at FROM prompt_registry WHERE name = $1 AND is_active = true ORDER BY version DESC LIMIT 1`
	row := r.db.GetConnection().QueryRow(query, name)
	return r.scanPrompt(row)
}

func (r *PromptRepository) GetByTaskType(taskType string) ([]*models.Prompt, error) {
	query := `
		SELECT id, name, task_type, prompt_text, version, is_active, created_at, updated_at
		FROM prompt_registry
		WHERE task_type = $1 AND is_active = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.GetConnection().Query(query, taskType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []*models.Prompt
	for rows.Next() {
		prompt, err := r.scanPrompt(rows)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

func (r *PromptRepository) Update(id uuid.UUID, newPromptText string) (*models.Prompt, error) {
	tx, err := r.db.GetConnection().Begin()
	if err != nil {
		return nil, err
	}

	var current models.Prompt
	err = tx.QueryRow(`SELECT version, name, task_type FROM prompt_registry WHERE id = $1 FOR UPDATE`, id).Scan(&current.Version, &current.Name, &current.TaskType)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newVersion := current.Version + 1
	updatedAt := time.Now()

	_, err = tx.Exec(`UPDATE prompt_registry SET is_active = false WHERE id = $1`, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newID := uuid.New()
	query := `
		INSERT INTO prompt_registry (id, name, task_type, prompt_text, version, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.Exec(query, newID, current.Name, current.TaskType, newPromptText, newVersion, true, updatedAt, updatedAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(newID)
}

func (r *PromptRepository) Deactivate(id uuid.UUID) error {
	query := `UPDATE prompt_registry SET is_active = false, updated_at = $2 WHERE id = $1`
	_, err := r.db.GetConnection().Exec(query, id, time.Now())
	return err
}

func (r *PromptRepository) List(limit int, offset int) ([]*models.Prompt, error) {
	query := `
		SELECT id, name, task_type, prompt_text, version, is_active, created_at, updated_at
		FROM prompt_registry
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.GetConnection().Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []*models.Prompt
	for rows.Next() {
		prompt, err := r.scanPrompt(rows)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *PromptRepository) scanPrompt(s scanner) (*models.Prompt, error) {
	var p models.Prompt
	err := s.Scan(
		&p.ID,
		&p.Name,
		&p.TaskType,
		&p.PromptText,
		&p.Version,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to scan prompt: %w", err)
	}
	return &p, nil
}
