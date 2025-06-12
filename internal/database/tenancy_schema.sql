-- Tenancy Database Schema for QLP Bridge Tenancy Model
-- This schema supports pooled, silo, and bridge tenancy models

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE,
    model VARCHAR(20) NOT NULL CHECK (model IN ('pooled', 'silo', 'bridge')),
    tier VARCHAR(20) NOT NULL CHECK (tier IN ('free', 'standard', 'premium', 'enterprise')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'suspended', 'pending', 'deleted')),
    
    -- JSON fields for flexible configuration
    settings JSONB NOT NULL DEFAULT '{}',
    resources JSONB NOT NULL DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for tenants table
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain) WHERE domain IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_model ON tenants(model);
CREATE INDEX IF NOT EXISTS idx_tenants_tier ON tenants(tier);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- GIN index for JSON fields for better query performance
CREATE INDEX IF NOT EXISTS idx_tenants_settings_gin ON tenants USING GIN(settings);
CREATE INDEX IF NOT EXISTS idx_tenants_resources_gin ON tenants USING GIN(resources);
CREATE INDEX IF NOT EXISTS idx_tenants_metadata_gin ON tenants USING GIN(metadata);

-- Tenant metrics table for real-time metrics tracking
CREATE TABLE IF NOT EXISTS tenant_metrics (
    tenant_id VARCHAR(255) PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    active_requests INTEGER NOT NULL DEFAULT 0,
    total_requests BIGINT NOT NULL DEFAULT 0,
    error_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    avg_response_time DECIMAL(10,3) NOT NULL DEFAULT 0.0,
    resource_utilization JSONB NOT NULL DEFAULT '{}',
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for tenant metrics
CREATE INDEX IF NOT EXISTS idx_tenant_metrics_last_activity ON tenant_metrics(last_activity);
CREATE INDEX IF NOT EXISTS idx_tenant_metrics_active_requests ON tenant_metrics(active_requests);
CREATE INDEX IF NOT EXISTS idx_tenant_metrics_error_rate ON tenant_metrics(error_rate);

-- Tenant events table for audit trail and lifecycle tracking
CREATE TABLE IF NOT EXISTS tenant_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    source VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for tenant events
CREATE INDEX IF NOT EXISTS idx_tenant_events_tenant_id ON tenant_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_events_type ON tenant_events(event_type);
CREATE INDEX IF NOT EXISTS idx_tenant_events_timestamp ON tenant_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_tenant_events_source ON tenant_events(source);

-- Database shards table for tracking shard assignments
CREATE TABLE IF NOT EXISTS database_shards (
    shard_name VARCHAR(100) PRIMARY KEY,
    shard_type VARCHAR(20) NOT NULL CHECK (shard_type IN ('pooled', 'silo', 'bridge')),
    connection_string TEXT NOT NULL,
    max_connections INTEGER NOT NULL DEFAULT 100,
    current_connections INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'maintenance', 'disabled')),
    region VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for database shards
CREATE INDEX IF NOT EXISTS idx_database_shards_type ON database_shards(shard_type);
CREATE INDEX IF NOT EXISTS idx_database_shards_status ON database_shards(status);
CREATE INDEX IF NOT EXISTS idx_database_shards_region ON database_shards(region);

-- Tenant shard mappings for tracking which tenants use which shards
CREATE TABLE IF NOT EXISTS tenant_shard_mappings (
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    shard_name VARCHAR(100) NOT NULL REFERENCES database_shards(shard_name),
    resource_type VARCHAR(50) NOT NULL, -- 'database', 'storage', 'compute'
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, shard_name, resource_type)
);

-- Indexes for tenant shard mappings
CREATE INDEX IF NOT EXISTS idx_tenant_shard_mappings_tenant ON tenant_shard_mappings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_shard_mappings_shard ON tenant_shard_mappings(shard_name);
CREATE INDEX IF NOT EXISTS idx_tenant_shard_mappings_resource_type ON tenant_shard_mappings(resource_type);

-- Update trigger for tenants updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at 
    BEFORE UPDATE ON tenants 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_metrics_updated_at 
    BEFORE UPDATE ON tenant_metrics 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_database_shards_updated_at 
    BEFORE UPDATE ON database_shards 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to create tenant with default settings
CREATE OR REPLACE FUNCTION create_tenant(
    p_tenant_id VARCHAR(255),
    p_name VARCHAR(255),
    p_domain VARCHAR(255),
    p_model VARCHAR(20),
    p_tier VARCHAR(20)
) RETURNS VOID AS $$
DECLARE
    default_settings JSONB;
    default_resources JSONB;
BEGIN
    -- Set default settings based on tier
    CASE p_tier
        WHEN 'free' THEN
            default_settings := '{
                "data_isolation": false,
                "compute_isolation": false,
                "network_isolation": false,
                "storage_isolation": false,
                "encryption_at_rest": false,
                "audit_logging": false,
                "max_concurrent_jobs": 2,
                "priority_boost": 0,
                "resource_quota": {
                    "max_cpu": 1.0,
                    "max_memory": 512,
                    "max_storage": 1,
                    "max_requests": 100,
                    "max_projects": 3,
                    "max_artifacts": 10
                },
                "enabled_features": [],
                "disabled_features": ["advanced_validation", "custom_rules"]
            }';
        WHEN 'standard' THEN
            default_settings := '{
                "data_isolation": false,
                "compute_isolation": false,
                "network_isolation": false,
                "storage_isolation": false,
                "encryption_at_rest": true,
                "audit_logging": true,
                "max_concurrent_jobs": 5,
                "priority_boost": 1,
                "resource_quota": {
                    "max_cpu": 2.0,
                    "max_memory": 2048,
                    "max_storage": 10,
                    "max_requests": 1000,
                    "max_projects": 10,
                    "max_artifacts": 100
                },
                "enabled_features": ["advanced_validation", "api_access"],
                "disabled_features": []
            }';
        WHEN 'premium' THEN
            default_settings := '{
                "data_isolation": true,
                "compute_isolation": false,
                "network_isolation": false,
                "storage_isolation": true,
                "encryption_at_rest": true,
                "audit_logging": true,
                "max_concurrent_jobs": 10,
                "priority_boost": 2,
                "resource_quota": {
                    "max_cpu": 4.0,
                    "max_memory": 8192,
                    "max_storage": 100,
                    "max_requests": 10000,
                    "max_projects": 50,
                    "max_artifacts": 1000
                },
                "enabled_features": ["advanced_validation", "api_access", "custom_rules", "priority_processing"],
                "disabled_features": []
            }';
        WHEN 'enterprise' THEN
            default_settings := '{
                "data_isolation": true,
                "compute_isolation": true,
                "network_isolation": true,
                "storage_isolation": true,
                "encryption_at_rest": true,
                "audit_logging": true,
                "max_concurrent_jobs": 50,
                "priority_boost": 5,
                "resource_quota": {
                    "max_cpu": 16.0,
                    "max_memory": 32768,
                    "max_storage": 1000,
                    "max_requests": 100000,
                    "max_projects": 500,
                    "max_artifacts": 10000
                },
                "enabled_features": ["advanced_validation", "api_access", "custom_rules", "priority_processing", "sso_integration", "compliance_reports"],
                "disabled_features": []
            }';
    END CASE;

    -- Set default resources
    default_resources := '{
        "current_cpu": 0,
        "current_memory": 0,
        "current_storage": 0,
        "current_requests": 0,
        "current_projects": 0,
        "current_artifacts": 0,
        "dedicated_nodes": [],
        "dedicated_dbs": [],
        "dedicated_storage": [],
        "network_segment": "",
        "last_updated": "' || NOW()::TEXT || '"
    }';

    -- Insert the tenant
    INSERT INTO tenants (
        id, name, domain, model, tier, status,
        settings, resources, metadata
    ) VALUES (
        p_tenant_id, p_name, p_domain, p_model, p_tier, 'pending',
        default_settings, default_resources, '{}'
    );

    -- Create initial metrics record
    INSERT INTO tenant_metrics (tenant_id) VALUES (p_tenant_id);

    -- Log creation event
    INSERT INTO tenant_events (tenant_id, event_type, source, data) 
    VALUES (
        p_tenant_id, 
        'tenant.created', 
        'system',
        jsonb_build_object(
            'tier', p_tier,
            'model', p_model,
            'domain', p_domain
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Function to update tenant resources
CREATE OR REPLACE FUNCTION update_tenant_resources(
    p_tenant_id VARCHAR(255),
    p_resource_updates JSONB
) RETURNS VOID AS $$
BEGIN
    UPDATE tenants 
    SET 
        resources = resources || p_resource_updates || jsonb_build_object('last_updated', NOW()::TEXT),
        updated_at = NOW()
    WHERE id = p_tenant_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tenant % not found', p_tenant_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Insert default database shards for pooled tenants
INSERT INTO database_shards (shard_name, shard_type, connection_string, region) VALUES
    ('pooled_shard_0', 'pooled', 'postgresql://pooled_user:password@pooled-db-0:5432/qlp_pooled_0', 'us-east-1'),
    ('pooled_shard_1', 'pooled', 'postgresql://pooled_user:password@pooled-db-1:5432/qlp_pooled_1', 'us-east-1'),
    ('pooled_shard_2', 'pooled', 'postgresql://pooled_user:password@pooled-db-2:5432/qlp_pooled_2', 'us-east-1'),
    ('pooled_shard_3', 'pooled', 'postgresql://pooled_user:password@pooled-db-3:5432/qlp_pooled_3', 'us-east-1')
ON CONFLICT (shard_name) DO NOTHING;

-- Create sample tenants for demonstration
SELECT create_tenant('tenant_demo_free', 'Demo Free Tenant', 'demo-free.qlp.dev', 'pooled', 'free');
SELECT create_tenant('tenant_demo_enterprise', 'Demo Enterprise Tenant', 'demo-enterprise.qlp.dev', 'silo', 'enterprise');
SELECT create_tenant('tenant_demo_bridge', 'Demo Bridge Tenant', 'demo-bridge.qlp.dev', 'bridge', 'premium');

-- Activate the demo tenants
UPDATE tenants SET status = 'active', activated_at = NOW() 
WHERE id IN ('tenant_demo_free', 'tenant_demo_enterprise', 'tenant_demo_bridge');

-- Add some sample metrics
UPDATE tenant_metrics SET 
    total_requests = 1250,
    avg_response_time = 145.5,
    resource_utilization = '{
        "cpu": 0.45,
        "memory": 0.67,
        "storage": 0.23,
        "network": 0.12
    }'
WHERE tenant_id = 'tenant_demo_enterprise';

-- Add indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tenants_composite_lookup 
    ON tenants(status, model, tier) 
    WHERE status = 'active';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tenant_events_recent 
    ON tenant_events(tenant_id, timestamp DESC) 
    WHERE timestamp > NOW() - INTERVAL '30 days';

-- Add comments for documentation
COMMENT ON TABLE tenants IS 'Core tenant information supporting bridge tenancy model';
COMMENT ON TABLE tenant_metrics IS 'Real-time tenant performance and usage metrics';
COMMENT ON TABLE tenant_events IS 'Audit trail for tenant lifecycle events';
COMMENT ON TABLE database_shards IS 'Database shard definitions for multi-tenant data isolation';
COMMENT ON TABLE tenant_shard_mappings IS 'Mapping of tenants to their assigned resource shards';

COMMENT ON COLUMN tenants.model IS 'Tenancy model: pooled (shared), silo (dedicated), bridge (hybrid)';
COMMENT ON COLUMN tenants.tier IS 'Service tier determining features and resource limits';
COMMENT ON COLUMN tenants.settings IS 'Tenant-specific configuration including isolation and security settings';
COMMENT ON COLUMN tenants.resources IS 'Current resource allocation and usage tracking';
