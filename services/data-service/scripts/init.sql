-- Initialize database for QLP Data Service
-- This script runs when the container starts for the first time

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- Create intents table
CREATE TABLE IF NOT EXISTS intents (
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
);

-- Create indexes for intents
CREATE INDEX IF NOT EXISTS idx_intents_tenant_id ON intents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_intents_status ON intents(status);
CREATE INDEX IF NOT EXISTS idx_intents_created_at ON intents(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intents_tenant_status ON intents(tenant_id, status);

-- Create intent_embeddings table for vector search
CREATE TABLE IF NOT EXISTS intent_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id UUID NOT NULL REFERENCES intents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    embedding vector(1536),
    model_name VARCHAR(100) DEFAULT 'text-embedding-ada-002',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(intent_id)
);

-- Create vector index for similarity search
CREATE INDEX IF NOT EXISTS idx_intent_embeddings_vector 
ON intent_embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

-- Create indexes for embeddings
CREATE INDEX IF NOT EXISTS idx_intent_embeddings_tenant_id ON intent_embeddings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_intent_embeddings_intent_id ON intent_embeddings(intent_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for intents table
DROP TRIGGER IF EXISTS update_intents_updated_at ON intents;
CREATE TRIGGER update_intents_updated_at 
BEFORE UPDATE ON intents 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Insert sample data for testing
INSERT INTO intents (tenant_id, user_input, status) VALUES 
('11111111-1111-1111-1111-111111111111', 'Create a REST API for user management', 'pending'),
('11111111-1111-1111-1111-111111111111', 'Deploy a web application to Azure', 'completed'),
('22222222-2222-2222-2222-222222222222', 'Analyze code quality and security', 'processing');

COMMIT;