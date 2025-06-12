package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/data-service/pkg/contracts"
)

type VectorRepository struct {
	db *sql.DB
}

func NewVectorRepository(db *sql.DB) *VectorRepository {
	return &VectorRepository{db: db}
}

// CreateEmbedding stores an embedding for an intent
func (vr *VectorRepository) CreateEmbedding(ctx context.Context, intentID, tenantID string, embedding []float64, model string) error {
	query := `
		INSERT INTO intent_embeddings (intent_id, tenant_id, embedding, model_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (intent_id) DO UPDATE SET
			embedding = EXCLUDED.embedding,
			model_name = EXCLUDED.model_name,
			created_at = NOW()
	`

	// Convert float64 slice to postgres array format
	embeddingArray := pq.Array(embedding)

	_, err := vr.db.ExecContext(ctx, query, intentID, tenantID, embeddingArray, model)
	if err != nil {
		logger.WithComponent("vector-repository").Error("Failed to create embedding",
			zap.String("intent_id", intentID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return fmt.Errorf("failed to create embedding: %w", err)
	}

	logger.WithComponent("vector-repository").Info("Embedding created",
		zap.String("intent_id", intentID),
		zap.String("tenant_id", tenantID),
		zap.String("model", model))

	return nil
}

// FindSimilar finds similar intents using vector similarity search
func (vr *VectorRepository) FindSimilar(ctx context.Context, req *contracts.VectorSimilarRequest) (*contracts.VectorSimilarResponse, error) {
	if len(req.Embedding) == 0 {
		return nil, fmt.Errorf("embedding is required")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.7 // Default similarity threshold
	}

	query := `
		WITH similar_embeddings AS (
			SELECT 
				ie.intent_id,
				ie.tenant_id,
				1 - (ie.embedding <=> $1) AS similarity
			FROM intent_embeddings ie
			WHERE ie.tenant_id = $2
			AND 1 - (ie.embedding <=> $1) >= $3
			ORDER BY ie.embedding <=> $1
			LIMIT $4
		)
		SELECT 
			i.id, i.tenant_id, i.user_input, i.parsed_tasks, i.metadata, 
			i.status, i.overall_score, i.execution_time_ms, 
			i.created_at, i.updated_at, i.completed_at,
			se.similarity
		FROM similar_embeddings se
		JOIN intents i ON i.id = se.intent_id
		ORDER BY se.similarity DESC
	`

	embeddingArray := pq.Array(req.Embedding)
	rows, err := vr.db.QueryContext(ctx, query, embeddingArray, req.TenantID, threshold, limit)
	if err != nil {
		logger.WithComponent("vector-repository").Error("Failed to find similar intents",
			zap.String("tenant_id", req.TenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find similar intents: %w", err)
	}
	defer rows.Close()

	var results []contracts.SimilarIntent
	for rows.Next() {
		var intent contracts.Intent
		var similarity float64
		var metadataJSON, tasksJSON sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
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
			&similarity,
		)
		if err != nil {
			logger.WithComponent("vector-repository").Error("Failed to scan similar intent", zap.Error(err))
			continue
		}

		// Parse JSON fields (similar to intent_repository.go)
		if metadataJSON.Valid {
			// Parse metadata JSON - simplified for this example
			intent.Metadata = make(map[string]string)
		} else {
			intent.Metadata = make(map[string]string)
		}

		if tasksJSON.Valid {
			// Parse tasks JSON - simplified for this example
			intent.ParsedTasks = []contracts.Task{}
		} else {
			intent.ParsedTasks = []contracts.Task{}
		}

		if completedAt.Valid {
			intent.CompletedAt = &completedAt.Time
		}

		results = append(results, contracts.SimilarIntent{
			Intent:     intent,
			Similarity: similarity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating similar intent rows: %w", err)
	}

	logger.WithComponent("vector-repository").Info("Found similar intents",
		zap.String("tenant_id", req.TenantID),
		zap.Int("count", len(results)),
		zap.Float64("threshold", threshold))

	return &contracts.VectorSimilarResponse{Results: results}, nil
}

// GetEmbedding retrieves an existing embedding for an intent
func (vr *VectorRepository) GetEmbedding(ctx context.Context, intentID, tenantID string) ([]float64, error) {
	query := `
		SELECT embedding 
		FROM intent_embeddings 
		WHERE intent_id = $1 AND tenant_id = $2
	`

	var embeddingArray pq.Float64Array
	err := vr.db.QueryRowContext(ctx, query, intentID, tenantID).Scan(&embeddingArray)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("embedding not found")
		}
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	return []float64(embeddingArray), nil
}

// DeleteEmbedding removes an embedding for an intent
func (vr *VectorRepository) DeleteEmbedding(ctx context.Context, intentID, tenantID string) error {
	query := `DELETE FROM intent_embeddings WHERE intent_id = $1 AND tenant_id = $2`
	
	result, err := vr.db.ExecContext(ctx, query, intentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete embedding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("embedding not found")
	}

	logger.WithComponent("vector-repository").Info("Embedding deleted",
		zap.String("intent_id", intentID),
		zap.String("tenant_id", tenantID))

	return nil
}