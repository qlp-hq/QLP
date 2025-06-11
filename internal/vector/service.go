package vector

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"QLP/internal/database"
	"QLP/internal/llm"
	"QLP/internal/models"
)

type VectorService struct {
	db        *database.Database
	llmClient llm.Client
}

type SimilarIntent struct {
	Intent     *models.Intent `json:"intent"`
	Similarity float64        `json:"similarity"`
}

type IntentEmbedding struct {
	IntentID  string    `json:"intent_id"`
	Embedding []float32 `json:"embedding"`
}

func NewVectorService(db *database.Database, llmClient llm.Client) *VectorService {
	return &VectorService{
		db:        db,
		llmClient: llmClient,
	}
}

// GenerateEmbedding creates a vector embedding for the given text
func (vs *VectorService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if !vs.db.IsConnected() {
		// Fallback: return a simple hash-based embedding for development
		return vs.generateSimpleEmbedding(text), nil
	}

	// Use OpenAI's embedding API
	embedding, err := vs.llmClient.GenerateEmbedding(ctx, text)
	if err != nil {
		log.Printf("⚠️  Failed to generate LLM embedding: %v", err)
		// Fallback to simple embedding
		return vs.generateSimpleEmbedding(text), nil
	}

	return embedding, nil
}

// StoreIntentEmbedding stores an intent's embedding in the database
func (vs *VectorService) StoreIntentEmbedding(ctx context.Context, intentID string, userInput string) error {
	if !vs.db.IsConnected() {
		log.Printf("📝 Database not connected, skipping embedding storage")
		return nil
	}

	embedding, err := vs.GenerateEmbedding(ctx, userInput)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert embedding to PostgreSQL vector format
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `
		UPDATE intents 
		SET embedding = $1::vector 
		WHERE id = $2
	`

	_, err = vs.db.GetConnection().ExecContext(ctx, query, string(embeddingJSON), intentID)
	if err != nil {
		log.Printf("⚠️  Failed to store embedding (pgvector not available): %v", err)
		// Don't fail the whole operation if vector storage fails
		return nil
	}

	log.Printf("🔍 Stored embedding for intent: %s", intentID)
	return nil
}

// FindSimilarIntents finds intents similar to the given text
func (vs *VectorService) FindSimilarIntents(ctx context.Context, userInput string, limit int) ([]SimilarIntent, error) {
	if !vs.db.IsConnected() {
		return []SimilarIntent{}, nil
	}

	// Generate embedding for the query text
	queryEmbedding, err := vs.GenerateEmbedding(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to JSON for query
	embeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Query for similar intents using cosine similarity
	query := `
		SELECT id, user_input, parsed_tasks, metadata, status, overall_score,
		       execution_time_ms, created_at, updated_at, completed_at,
		       1 - (embedding <=> $1::vector) as similarity
		FROM intents 
		WHERE embedding IS NOT NULL 
		  AND id != $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	rows, err := vs.db.GetConnection().QueryContext(ctx, query, string(embeddingJSON), "", limit)
	if err != nil {
		// If pgvector queries fail, return empty results
		log.Printf("⚠️  Vector similarity search not available: %v", err)
		return []SimilarIntent{}, nil
	}
	defer rows.Close()

	var results []SimilarIntent
	for rows.Next() {
		var intent models.Intent
		var tasksJSON, metadataJSON []byte
		var completedAt sql.NullTime
		var similarity float64

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
			&similarity,
		)

		if err != nil {
			continue // Skip malformed records
		}

		if completedAt.Valid {
			intent.CompletedAt = &completedAt.Time
		}

		if err := json.Unmarshal(tasksJSON, &intent.Tasks); err != nil {
			continue
		}

		if err := json.Unmarshal(metadataJSON, &intent.Metadata); err != nil {
			continue
		}

		results = append(results, SimilarIntent{
			Intent:     &intent,
			Similarity: similarity,
		})
	}

	return results, nil
}

// GetIntentSuggestions provides suggestions based on similar intents
func (vs *VectorService) GetIntentSuggestions(ctx context.Context, userInput string) ([]string, error) {
	similarIntents, err := vs.FindSimilarIntents(ctx, userInput, 3)
	if err != nil {
		return nil, err
	}

	var suggestions []string

	for _, similar := range similarIntents {
		if similar.Similarity > 0.8 { // High similarity threshold
			suggestion := fmt.Sprintf("Similar to '%s' (%.1f%% match) - achieved %d/100 score",
				similar.Intent.UserInput,
				similar.Similarity*100,
				similar.Intent.OverallScore)
			suggestions = append(suggestions, suggestion)
		}
	}

	if len(suggestions) == 0 && len(similarIntents) > 0 {
		// Provide a general suggestion
		best := similarIntents[0]
		suggestion := fmt.Sprintf("You might also consider: '%s' (achieved %d/100 score)",
			best.Intent.UserInput,
			best.Intent.OverallScore)
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// generateSimpleEmbedding creates a basic embedding from text for fallback
func (vs *VectorService) generateSimpleEmbedding(text string) []float32 {
	// Simple character-based embedding for development/fallback
	embedding := make([]float32, 384) // Standard embedding size
	
	// Basic hash-like distribution
	for i, char := range text {
		if i >= len(embedding) {
			break
		}
		embedding[i%len(embedding)] += float32(char) / 1000.0
	}
	
	// Normalize
	var norm float32
	for _, val := range embedding {
		norm += val * val
	}
	if norm > 0 {
		norm = 1.0 / sqrt(norm)
		for i := range embedding {
			embedding[i] *= norm
		}
	}
	
	return embedding
}

// Simple square root implementation
func sqrt(x float32) float32 {
	if x <= 0 {
		return 0
	}
	
	guess := x / 2
	for i := 0; i < 10; i++ { // Newton's method iterations
		guess = (guess + x/guess) / 2
	}
	return guess
}