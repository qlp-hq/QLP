package tenancy

import (
	"time"
)

// TenantModel represents the tenancy model for a tenant
type TenantModel string

const (
	// TenantModelPooled - shared infrastructure (default for most tenants)
	TenantModelPooled TenantModel = "pooled"
	// TenantModelSilo - dedicated infrastructure (enterprise tenants)
	TenantModelSilo TenantModel = "silo"
	// TenantModelBridge - hybrid model with selective isolation
	TenantModelBridge TenantModel = "bridge"
)

// TenantTier represents the service tier level
type TenantTier string

const (
	TenantTierFree       TenantTier = "free"
	TenantTierStandard   TenantTier = "standard"
	TenantTierPremium    TenantTier = "premium"
	TenantTierEnterprise TenantTier = "enterprise"
)

// TenantStatus represents the current status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Domain      string            `json:"domain,omitempty"`
	Model       TenantModel       `json:"model"`
	Tier        TenantTier        `json:"tier"`
	Status      TenantStatus      `json:"status"`
	Settings    TenantSettings    `json:"settings"`
	Resources   TenantResources   `json:"resources"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ActivatedAt *time.Time        `json:"activated_at,omitempty"`
}

// TenantSettings contains tenant-specific configuration
type TenantSettings struct {
	// Isolation settings
	DataIsolation     bool `json:"data_isolation"`
	ComputeIsolation  bool `json:"compute_isolation"`
	NetworkIsolation  bool `json:"network_isolation"`
	StorageIsolation  bool `json:"storage_isolation"`
	
	// Security settings
	EncryptionAtRest  bool   `json:"encryption_at_rest"`
	CustomKMSKey      string `json:"custom_kms_key,omitempty"`
	AuditLogging      bool   `json:"audit_logging"`
	ComplianceMode    string `json:"compliance_mode,omitempty"` // HIPAA, SOC2, etc.
	
	// Performance settings
	MaxConcurrentJobs int `json:"max_concurrent_jobs"`
	PriorityBoost     int `json:"priority_boost"`
	ResourceQuota     ResourceQuota `json:"resource_quota"`
	
	// Feature flags
	EnabledFeatures   []string `json:"enabled_features,omitempty"`
	DisabledFeatures  []string `json:"disabled_features,omitempty"`
}

// ResourceQuota defines resource limits for a tenant
type ResourceQuota struct {
	MaxCPU       float64 `json:"max_cpu"`        // CPU cores
	MaxMemory    int64   `json:"max_memory"`     // Memory in MB
	MaxStorage   int64   `json:"max_storage"`    // Storage in GB
	MaxRequests  int     `json:"max_requests"`   // Requests per minute
	MaxProjects  int     `json:"max_projects"`   // Number of projects
	MaxArtifacts int     `json:"max_artifacts"`  // Number of artifacts
}

// TenantResources tracks current resource usage
type TenantResources struct {
	// Current usage
	CurrentCPU       float64 `json:"current_cpu"`
	CurrentMemory    int64   `json:"current_memory"`
	CurrentStorage   int64   `json:"current_storage"`
	CurrentRequests  int     `json:"current_requests"`
	CurrentProjects  int     `json:"current_projects"`
	CurrentArtifacts int     `json:"current_artifacts"`
	
	// Infrastructure assignments (for silo/bridge tenants)
	DedicatedNodes   []string `json:"dedicated_nodes,omitempty"`
	DedicatedDBs     []string `json:"dedicated_dbs,omitempty"`
	DedicatedStorage []string `json:"dedicated_storage,omitempty"`
	NetworkSegment   string   `json:"network_segment,omitempty"`
	
	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// TenantContext provides request-scoped tenant information
type TenantContext struct {
	Tenant      *Tenant              `json:"tenant"`
	Isolation   IsolationContext     `json:"isolation"`
	Permissions map[string]bool      `json:"permissions"`
	Metrics     *TenantMetrics       `json:"metrics,omitempty"`
	RequestID   string               `json:"request_id"`
	UserID      string               `json:"user_id,omitempty"`
	SessionID   string               `json:"session_id,omitempty"`
}

// IsolationContext contains isolation-specific routing information
type IsolationContext struct {
	Model           TenantModel       `json:"model"`
	DatabaseShard   string            `json:"database_shard"`
	StorageBucket   string            `json:"storage_bucket"`
	ComputeNodes    []string          `json:"compute_nodes,omitempty"`
	NetworkSegment  string            `json:"network_segment,omitempty"`
	EncryptionKey   string            `json:"encryption_key,omitempty"`
	RoutingTags     map[string]string `json:"routing_tags,omitempty"`
}

// TenantMetrics contains real-time tenant metrics
type TenantMetrics struct {
	ActiveRequests   int       `json:"active_requests"`
	TotalRequests    int64     `json:"total_requests"`
	ErrorRate        float64   `json:"error_rate"`
	AvgResponseTime  float64   `json:"avg_response_time_ms"`
	ResourceUtilization map[string]float64 `json:"resource_utilization"`
	LastActivity     time.Time `json:"last_activity"`
}

// TenantEvent represents tenant lifecycle events
type TenantEvent struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	Type      TenantEventType   `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
}

type TenantEventType string

const (
	TenantEventCreated         TenantEventType = "tenant.created"
	TenantEventUpdated         TenantEventType = "tenant.updated"
	TenantEventActivated       TenantEventType = "tenant.activated"
	TenantEventSuspended       TenantEventType = "tenant.suspended"
	TenantEventDeleted         TenantEventType = "tenant.deleted"
	TenantEventTierChanged     TenantEventType = "tenant.tier_changed"
	TenantEventModelChanged    TenantEventType = "tenant.model_changed"
	TenantEventQuotaExceeded   TenantEventType = "tenant.quota_exceeded"
	TenantEventResourceAllocated TenantEventType = "tenant.resource_allocated"
)