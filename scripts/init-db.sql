-- Initialize PostgreSQL with pgvector extension for QuantumLayer Platform
-- This script sets up the database with vector search capabilities

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Enable other useful extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Create vector storage for embeddings (1536 dimensions for OpenAI embeddings)
CREATE TABLE IF NOT EXISTS embeddings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(1536),
    metadata JSONB DEFAULT '{}',
    tenant_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS embeddings_embedding_idx ON embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create index for tenant isolation
CREATE INDEX IF NOT EXISTS embeddings_tenant_id_idx ON embeddings (tenant_id);

-- Create index for metadata search
CREATE INDEX IF NOT EXISTS embeddings_metadata_idx ON embeddings USING gin (metadata);

-- Create index for content search
CREATE INDEX IF NOT EXISTS embeddings_content_idx ON embeddings USING gin (content gin_trgm_ops);

-- Create function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_embeddings_updated_at 
    BEFORE UPDATE ON embeddings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create documents table for storing processed documents
CREATE TABLE IF NOT EXISTS documents (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(100) DEFAULT 'text/plain',
    size_bytes INTEGER DEFAULT 0,
    checksum VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    tenant_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for documents
CREATE INDEX IF NOT EXISTS documents_tenant_id_idx ON documents (tenant_id);
CREATE INDEX IF NOT EXISTS documents_name_idx ON documents (name);
CREATE INDEX IF NOT EXISTS documents_content_type_idx ON documents (content_type);
CREATE INDEX IF NOT EXISTS documents_metadata_idx ON documents USING gin (metadata);

-- Create trigger for documents updated_at
CREATE TRIGGER update_documents_updated_at 
    BEFORE UPDATE ON documents 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create vector collections table for organizing embeddings
CREATE TABLE IF NOT EXISTS vector_collections (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    dimension INTEGER DEFAULT 1536,
    distance_metric VARCHAR(50) DEFAULT 'cosine', -- cosine, euclidean, dot_product
    metadata JSONB DEFAULT '{}',
    tenant_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Create indexes for vector collections
CREATE INDEX IF NOT EXISTS vector_collections_tenant_id_idx ON vector_collections (tenant_id);
CREATE INDEX IF NOT EXISTS vector_collections_name_idx ON vector_collections (name);

-- Create trigger for vector collections updated_at
CREATE TRIGGER update_vector_collections_updated_at 
    BEFORE UPDATE ON vector_collections 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add collection_id to embeddings table
ALTER TABLE embeddings ADD COLUMN IF NOT EXISTS collection_id UUID;
CREATE INDEX IF NOT EXISTS embeddings_collection_id_idx ON embeddings (collection_id);

-- Create foreign key constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_embeddings_collection'
    ) THEN
        ALTER TABLE embeddings 
        ADD CONSTRAINT fk_embeddings_collection 
        FOREIGN KEY (collection_id) REFERENCES vector_collections(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Create function for vector similarity search
CREATE OR REPLACE FUNCTION search_similar_vectors(
    query_embedding vector(1536),
    collection_name text DEFAULT NULL,
    tenant_id_param text DEFAULT NULL,
    limit_param integer DEFAULT 10,
    threshold_param float DEFAULT 0.7
)
RETURNS TABLE (
    id UUID,
    content TEXT,
    similarity FLOAT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.id,
        e.content,
        1 - (e.embedding <=> query_embedding) as similarity,
        e.metadata,
        e.created_at
    FROM embeddings e
    LEFT JOIN vector_collections vc ON e.collection_id = vc.id
    WHERE 
        (tenant_id_param IS NULL OR e.tenant_id = tenant_id_param)
        AND (collection_name IS NULL OR vc.name = collection_name)
        AND (1 - (e.embedding <=> query_embedding)) >= threshold_param
    ORDER BY e.embedding <=> query_embedding
    LIMIT limit_param;
END;
$$ LANGUAGE plpgsql;

-- Create function for hybrid search (vector + text)
CREATE OR REPLACE FUNCTION hybrid_search(
    query_text text,
    query_embedding vector(1536),
    collection_name text DEFAULT NULL,
    tenant_id_param text DEFAULT NULL,
    limit_param integer DEFAULT 10,
    vector_weight float DEFAULT 0.7,
    text_weight float DEFAULT 0.3
)
RETURNS TABLE (
    id UUID,
    content TEXT,
    score FLOAT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.id,
        e.content,
        (vector_weight * (1 - (e.embedding <=> query_embedding))) + 
        (text_weight * similarity(e.content, query_text)) as score,
        e.metadata,
        e.created_at
    FROM embeddings e
    LEFT JOIN vector_collections vc ON e.collection_id = vc.id
    WHERE 
        (tenant_id_param IS NULL OR e.tenant_id = tenant_id_param)
        AND (collection_name IS NULL OR vc.name = collection_name)
    ORDER BY score DESC
    LIMIT limit_param;
END;
$$ LANGUAGE plpgsql;

-- Insert default vector collection
INSERT INTO vector_collections (name, description, tenant_id) 
VALUES ('default', 'Default vector collection for general purpose embeddings', 'system')
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Create sample data for testing (optional)
INSERT INTO embeddings (content, embedding, tenant_id, collection_id, metadata)
SELECT 
    'Sample document ' || i,
    array_fill(random()::float, ARRAY[1536])::vector,
    'demo-tenant',
    (SELECT id FROM vector_collections WHERE name = 'default' AND tenant_id = 'system'),
    json_build_object('type', 'sample', 'index', i)::jsonb
FROM generate_series(1, 5) as i
ON CONFLICT DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO qlp_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO qlp_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO qlp_user;

-- Log initialization completion
DO $$
BEGIN
    RAISE NOTICE 'QuantumLayer Platform database initialized successfully with pgvector support';
    RAISE NOTICE 'Vector database features:';
    RAISE NOTICE '  - pgvector extension for similarity search';
    RAISE NOTICE '  - 1536-dimensional embeddings (OpenAI compatible)';
    RAISE NOTICE '  - Hybrid search capabilities';
    RAISE NOTICE '  - Multi-tenant vector collections';
    RAISE NOTICE '  - Sample data loaded for testing';
END $$;