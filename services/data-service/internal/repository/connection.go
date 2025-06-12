package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"QLP/internal/logger"
)

// NewConnection creates a new database connection with optimized settings
func NewConnection(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.WithComponent("database").Info("Database connection established",
		zap.Int("max_open_conns", 25),
		zap.Int("max_idle_conns", 5))

	return db, nil
}

// CreateTables creates the required database tables for the data service
func CreateTables(db *sql.DB) error {
	queries := []string{
		// Enable UUID extension
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		
		// Enable pgvector extension
		`CREATE EXTENSION IF NOT EXISTS vector;`,
		
		// Create intents table
		`CREATE TABLE IF NOT EXISTS intents (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL,
			user_input TEXT NOT NULL,
			parsed_tasks JSONB,
			metadata JSONB DEFAULT '{}',
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			overall_score INTEGER DEFAULT 0,
			execution_time_ms INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE
		);`,
		
		// Create indexes for intents
		`CREATE INDEX IF NOT EXISTS idx_intents_tenant_id ON intents(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_intents_status ON intents(status);`,
		`CREATE INDEX IF NOT EXISTS idx_intents_created_at ON intents(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_intents_tenant_status ON intents(tenant_id, status);`,
		
		// Create intent_embeddings table for vector search
		`CREATE TABLE IF NOT EXISTS intent_embeddings (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			intent_id UUID NOT NULL REFERENCES intents(id) ON DELETE CASCADE,
			tenant_id UUID NOT NULL,
			embedding vector(1536),
			model_name VARCHAR(100) DEFAULT 'text-embedding-ada-002',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,
		
		// Create vector index for similarity search
		`CREATE INDEX IF NOT EXISTS idx_intent_embeddings_vector 
		 ON intent_embeddings USING ivfflat (embedding vector_cosine_ops) 
		 WITH (lists = 100);`,
		
		// Create indexes for embeddings
		`CREATE INDEX IF NOT EXISTS idx_intent_embeddings_tenant_id ON intent_embeddings(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_intent_embeddings_intent_id ON intent_embeddings(intent_id);`,
		
		// Create updated_at trigger function
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		 RETURNS TRIGGER AS $$
		 BEGIN
			 NEW.updated_at = NOW();
			 RETURN NEW;
		 END;
		 $$ language 'plpgsql';`,
		 
		// Create trigger for intents table
		`DROP TRIGGER IF EXISTS update_intents_updated_at ON intents;`,
		`CREATE TRIGGER update_intents_updated_at 
		 BEFORE UPDATE ON intents 
		 FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	logger.WithComponent("database").Info("Database tables created successfully")
	return nil
}