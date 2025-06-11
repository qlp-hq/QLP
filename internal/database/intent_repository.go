package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"QLP/internal/models"
)

type IntentRepository struct {
	db *Database
}

func NewIntentRepository(db *Database) *IntentRepository {
	return &IntentRepository{db: db}
}

func (r *IntentRepository) Create(intent *models.Intent) error {
	if !r.db.IsConnected() {
		// Fallback to file-based storage
		return r.createFileBased(intent)
	}

	tasksJSON, err := json.Marshal(intent.Tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	metadataJSON, err := json.Marshal(intent.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO intents (id, user_input, parsed_tasks, metadata, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err = r.db.conn.Exec(query, 
		intent.ID, 
		intent.UserInput, 
		tasksJSON, 
		metadataJSON,
		intent.Status,
		intent.CreatedAt,
	)
	
	return err
}

func (r *IntentRepository) GetByID(id string) (*models.Intent, error) {
	if !r.db.IsConnected() {
		return r.getByIDFileBased(id)
	}

	query := `
		SELECT id, user_input, parsed_tasks, metadata, status, overall_score, 
		       execution_time_ms, created_at, updated_at, completed_at
		FROM intents WHERE id = $1
	`
	
	row := r.db.conn.QueryRow(query, id)
	
	var intent models.Intent
	var tasksJSON, metadataJSON []byte
	var completedAt sql.NullTime
	
	err := row.Scan(
		&intent.ID,
		&intent.UserInput,
		&tasksJSON,
		&metadataJSON,
		&intent.Status,
		&intent.OverallScore,
		&intent.ExecutionTimeMS,
		&intent.CreatedAt,
		&intent.UpdatedAt,
		&completedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if completedAt.Valid {
		intent.CompletedAt = &completedAt.Time
	}
	
	if err := json.Unmarshal(tasksJSON, &intent.Tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}
	
	if err := json.Unmarshal(metadataJSON, &intent.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &intent, nil
}

func (r *IntentRepository) Update(intent *models.Intent) error {
	if !r.db.IsConnected() {
		return r.updateFileBased(intent)
	}

	tasksJSON, err := json.Marshal(intent.Tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	metadataJSON, err := json.Marshal(intent.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE intents 
		SET parsed_tasks = $2, metadata = $3, status = $4, overall_score = $5,
		    execution_time_ms = $6, completed_at = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	var completedAt interface{}
	if intent.CompletedAt != nil {
		completedAt = *intent.CompletedAt
	}
	
	_, err = r.db.conn.Exec(query,
		intent.ID,
		tasksJSON,
		metadataJSON,
		intent.Status,
		intent.OverallScore,
		intent.ExecutionTimeMS,
		completedAt,
	)
	
	return err
}

func (r *IntentRepository) List(limit int, offset int) ([]*models.Intent, error) {
	if !r.db.IsConnected() {
		return r.listFileBased(limit, offset)
	}

	query := `
		SELECT id, user_input, parsed_tasks, metadata, status, overall_score,
		       execution_time_ms, created_at, updated_at, completed_at
		FROM intents 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var intents []*models.Intent
	
	for rows.Next() {
		var intent models.Intent
		var tasksJSON, metadataJSON []byte
		var completedAt sql.NullTime
		
		err := rows.Scan(
			&intent.ID,
			&intent.UserInput,
			&tasksJSON,
			&metadataJSON,
			&intent.Status,
			&intent.OverallScore,
			&intent.ExecutionTimeMS,
			&intent.CreatedAt,
			&intent.UpdatedAt,
			&completedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if completedAt.Valid {
			intent.CompletedAt = &completedAt.Time
		}
		
		if err := json.Unmarshal(tasksJSON, &intent.Tasks); err != nil {
			continue // Skip malformed records
		}
		
		if err := json.Unmarshal(metadataJSON, &intent.Metadata); err != nil {
			continue // Skip malformed records
		}
		
		intents = append(intents, &intent)
	}
	
	return intents, nil
}

// File-based fallback methods
func (r *IntentRepository) createFileBased(intent *models.Intent) error {
	// TODO: Implement file-based storage as fallback
	return nil
}

func (r *IntentRepository) getByIDFileBased(id string) (*models.Intent, error) {
	// TODO: Implement file-based retrieval as fallback
	return nil, fmt.Errorf("intent not found in file storage")
}

func (r *IntentRepository) updateFileBased(intent *models.Intent) error {
	// TODO: Implement file-based update as fallback
	return nil
}

func (r *IntentRepository) listFileBased(limit, offset int) ([]*models.Intent, error) {
	// TODO: Implement file-based listing as fallback
	return []*models.Intent{}, nil
}