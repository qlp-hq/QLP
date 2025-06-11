-- QuantumLayer Database Schema
-- PostgreSQL 15+ with JSON and UUID support

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Core intent tracking
CREATE TABLE IF NOT EXISTS intents (
    id VARCHAR(50) PRIMARY KEY, -- QLI-timestamp format
    user_input TEXT NOT NULL,
    parsed_tasks JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'pending',
    overall_score INTEGER DEFAULT 0,
    execution_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    embedding VECTOR(1536) -- OpenAI embedding dimension
);

-- Task management
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(50) PRIMARY KEY, -- QL-DEV-001 format
    intent_id VARCHAR(50) REFERENCES intents(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- codegen, test, doc, infra, analyze
    description TEXT NOT NULL,
    dependencies JSONB DEFAULT '[]',
    priority INTEGER DEFAULT 5,
    status VARCHAR(50) DEFAULT 'pending',
    agent_id VARCHAR(50),
    output TEXT,
    validation_score INTEGER DEFAULT 0,
    execution_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Agent tracking
CREATE TABLE IF NOT EXISTS agents (
    id VARCHAR(50) PRIMARY KEY, -- QLD-AGT-001 format
    type VARCHAR(50) NOT NULL,
    task_id VARCHAR(50) REFERENCES tasks(id),
    capabilities JSONB DEFAULT '{}',
    context JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'initializing',
    llm_tokens_used INTEGER DEFAULT 0,
    execution_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Validation results
CREATE TABLE IF NOT EXISTS validation_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id VARCHAR(50) REFERENCES tasks(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL,
    syntax_score INTEGER DEFAULT 0,
    security_score INTEGER DEFAULT 0,
    quality_score INTEGER DEFAULT 0,
    llm_critique_score INTEGER DEFAULT 0,
    validation_details JSONB DEFAULT '{}',
    issues_found JSONB DEFAULT '[]',
    passed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Security findings
CREATE TABLE IF NOT EXISTS security_findings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    validation_result_id UUID REFERENCES validation_results(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL, -- low, medium, high, critical
    description TEXT NOT NULL,
    location TEXT,
    mitigation TEXT,
    cwe_id VARCHAR(20),
    owasp_category VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- HITL decisions
CREATE TABLE IF NOT EXISTS hitl_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id VARCHAR(50) REFERENCES intents(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- approve, reject, modify, escalate
    confidence FLOAT NOT NULL,
    auto_approved BOOLEAN DEFAULT false,
    review_required BOOLEAN DEFAULT false,
    quality_gates JSONB DEFAULT '{}',
    recommendations JSONB DEFAULT '[]',
    decision_reason TEXT,
    decided_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    decided_by VARCHAR(100) DEFAULT 'system'
);

-- QuantumCapsule metadata
CREATE TABLE IF NOT EXISTS quantum_capsules (
    id VARCHAR(50) PRIMARY KEY, -- QL-CAP-xxx format
    intent_id VARCHAR(50) REFERENCES intents(id) ON DELETE CASCADE,
    metadata JSONB NOT NULL,
    artifacts JSONB DEFAULT '[]',
    unified_project_path TEXT,
    file_count INTEGER DEFAULT 0,
    size_bytes BIGINT DEFAULT 0,
    overall_score INTEGER DEFAULT 0,
    security_risk VARCHAR(20) DEFAULT 'unknown',
    quality_score INTEGER DEFAULT 0,
    enterprise_ready BOOLEAN DEFAULT false,
    compliance_status JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Performance metrics
CREATE TABLE IF NOT EXISTS performance_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id VARCHAR(50) REFERENCES intents(id),
    task_id VARCHAR(50) REFERENCES tasks(id),
    agent_id VARCHAR(50) REFERENCES agents(id),
    metric_type VARCHAR(50) NOT NULL, -- execution_time, memory_usage, llm_tokens
    metric_value FLOAT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- System events
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    payload JSONB DEFAULT '{}',
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_intents_status ON intents(status);
CREATE INDEX IF NOT EXISTS idx_intents_created_at ON intents(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_intent_id ON tasks(intent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_agents_task_id ON agents(task_id);
CREATE INDEX IF NOT EXISTS idx_validation_results_task_id ON validation_results(task_id);
CREATE INDEX IF NOT EXISTS idx_security_findings_validation_id ON security_findings(validation_result_id);
CREATE INDEX IF NOT EXISTS idx_hitl_decisions_intent_id ON hitl_decisions(intent_id);
CREATE INDEX IF NOT EXISTS idx_quantum_capsules_intent_id ON quantum_capsules(intent_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_timestamp ON performance_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);

-- Vector similarity search index (for intent embeddings)
CREATE INDEX IF NOT EXISTS idx_intents_embedding ON intents USING ivfflat (embedding vector_cosine_ops);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_intents_updated_at 
    BEFORE UPDATE ON intents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();