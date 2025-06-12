package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/data-service/pkg/contracts"
)

type IntentRepository struct {
	db *sql.DB
}

func NewIntentRepository(db *sql.DB) *IntentRepository {
	return &IntentRepository{db: db}
}

func (ir *IntentRepository) Create(ctx context.Context, req *contracts.CreateIntentRequest, tenantID string) (*contracts.Intent, error) {
	intent := &contracts.Intent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		UserInput: req.UserInput,
		Metadata:  req.Metadata,
		Status:    contracts.IntentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if intent.Metadata == nil {
		intent.Metadata = make(map[string]string)
	}

	metadataJSON, err := json.Marshal(intent.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO intents (
			id, tenant_id, user_input, metadata, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = ir.db.ExecContext(ctx, query,
		intent.ID,
		intent.TenantID,
		intent.UserInput,
		metadataJSON,
		intent.Status,
		intent.CreatedAt,
		intent.UpdatedAt,
	)

	if err != nil {
		logger.WithComponent("intent-repository").Error("Failed to create intent",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create intent: %w", err)
	}

	logger.WithComponent("intent-repository").Info("Intent created",
		zap.String("intent_id", intent.ID),
		zap.String("tenant_id", tenantID))

	return intent, nil
}

func (ir *IntentRepository) GetByID(ctx context.Context, intentID, tenantID string) (*contracts.Intent, error) {
	query := `
		SELECT 
			id, tenant_id, user_input, parsed_tasks, metadata, status, 
			overall_score, execution_time_ms, created_at, updated_at, completed_at
		FROM intents 
		WHERE id = $1 AND tenant_id = $2
	`

	row := ir.db.QueryRowContext(ctx, query, intentID, tenantID)
	return ir.scanIntent(row)
}

func (ir *IntentRepository) List(ctx context.Context, tenantID string, req *contracts.ListIntentsRequest) ([]contracts.Intent, int, error) {
	var intents []contracts.Intent
	var total int

	// Count total matching records
	countQuery := `SELECT COUNT(*) FROM intents WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if req.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, req.Status)
		argIndex++
	}

	err := ir.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count intents: %w", err)
	}

	// Build main query
	query := `
		SELECT 
			id, tenant_id, user_input, parsed_tasks, metadata, status,
			overall_score, execution_time_ms, created_at, updated_at, completed_at
		FROM intents 
		WHERE tenant_id = $1
	`

	if req.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
		argIndex++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, req.Offset)
	}

	rows, err := ir.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list intents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		intent, err := ir.scanIntent(rows)
		if err != nil {
			logger.WithComponent("intent-repository").Error("Failed to scan intent", zap.Error(err))
			continue
		}
		intents = append(intents, *intent)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating intent rows: %w", err)
	}

	return intents, total, nil
}

func (ir *IntentRepository) Update(ctx context.Context, intentID, tenantID string, req *contracts.UpdateIntentRequest) (*contracts.Intent, error) {

	// Build dynamic update query
	setParts := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++

		// Set completed_at if status is completed
		if *req.Status == contracts.IntentStatusCompleted {
			setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
			args = append(args, time.Now())
			argIndex++
		}
	}

	if req.ParsedTasks != nil {
		tasksJSON, err := json.Marshal(req.ParsedTasks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tasks: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("parsed_tasks = $%d", argIndex))
		args = append(args, tasksJSON)
		argIndex++
	}

	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if req.OverallScore != nil {
		setParts = append(setParts, fmt.Sprintf("overall_score = $%d", argIndex))
		args = append(args, *req.OverallScore)
		argIndex++
	}

	if req.ExecutionTimeMS != nil {
		setParts = append(setParts, fmt.Sprintf("execution_time_ms = $%d", argIndex))
		args = append(args, *req.ExecutionTimeMS)
		argIndex++
	}

	// Add WHERE clause args
	args = append(args, intentID, tenantID)
	whereClause := fmt.Sprintf("id = $%d AND tenant_id = $%d", argIndex, argIndex+1)

	// Build proper SET clause
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf("UPDATE intents SET %s WHERE %s", setClause, whereClause)

	_, err := ir.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update intent: %w", err)
	}

	// Return updated intent
	return ir.GetByID(ctx, intentID, tenantID)
}

func (ir *IntentRepository) Delete(ctx context.Context, intentID, tenantID string) error {
	query := `DELETE FROM intents WHERE id = $1 AND tenant_id = $2`
	
	result, err := ir.db.ExecContext(ctx, query, intentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete intent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("intent not found")
	}

	logger.WithComponent("intent-repository").Info("Intent deleted",
		zap.String("intent_id", intentID),
		zap.String("tenant_id", tenantID))

	return nil
}

// scanIntent is a helper to scan intent from database rows
func (ir *IntentRepository) scanIntent(scanner interface {
	Scan(dest ...interface{}) error
}) (*contracts.Intent, error) {
	var intent contracts.Intent
	var metadataJSON, tasksJSON sql.NullString
	var completedAt sql.NullTime

	err := scanner.Scan(
		&intent.ID,
		&intent.TenantID,
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("intent not found")
		}
		return nil, fmt.Errorf("failed to scan intent: %w", err)
	}

	// Parse metadata JSON
	if metadataJSON.Valid {
		err = json.Unmarshal([]byte(metadataJSON.String), &intent.Metadata)
		if err != nil {
			intent.Metadata = make(map[string]string)
		}
	} else {
		intent.Metadata = make(map[string]string)
	}

	// Parse tasks JSON
	if tasksJSON.Valid {
		err = json.Unmarshal([]byte(tasksJSON.String), &intent.ParsedTasks)
		if err != nil {
			intent.ParsedTasks = []contracts.Task{}
		}
	} else {
		intent.ParsedTasks = []contracts.Task{}
	}

	// Handle completed_at
	if completedAt.Valid {
		intent.CompletedAt = &completedAt.Time
	}

	return &intent, nil
}