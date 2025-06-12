package tenancy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"QLP/internal/logger"
)

// TenantResolver provides tenant resolution and routing logic for bridge tenancy
type TenantResolver struct {
	repository   TenantRepository
	cache        map[string]*Tenant
	cacheMutex   sync.RWMutex
	cacheTTL     time.Duration
	lastRefresh  map[string]time.Time
	refreshMutex sync.RWMutex
}

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	ListTenants(ctx context.Context, filters TenantFilters) ([]*Tenant, error)
	CreateTenant(ctx context.Context, tenant *Tenant) error
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error)
}

// TenantFilters for querying tenants
type TenantFilters struct {
	Model    TenantModel  `json:"model,omitempty"`
	Tier     TenantTier   `json:"tier,omitempty"`
	Status   TenantStatus `json:"status,omitempty"`
	Limit    int          `json:"limit,omitempty"`
	Offset   int          `json:"offset,omitempty"`
}

// NewTenantResolver creates a new tenant resolver
func NewTenantResolver(repository TenantRepository) *TenantResolver {
	return &TenantResolver{
		repository:   repository,
		cache:        make(map[string]*Tenant),
		cacheTTL:     5 * time.Minute, // Cache tenants for 5 minutes
		lastRefresh:  make(map[string]time.Time),
	}
}

// ResolveTenant resolves tenant information and determines isolation context
func (tr *TenantResolver) ResolveTenant(ctx context.Context, tenantID string) (*TenantContext, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	// Get tenant from cache or repository
	tenant, err := tr.getTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant %s: %w", tenantID, err)
	}

	if tenant.Status != TenantStatusActive {
		return nil, fmt.Errorf("tenant %s is not active (status: %s)", tenantID, tenant.Status)
	}

	// Build isolation context based on tenant model
	isolationCtx, err := tr.buildIsolationContext(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to build isolation context: %w", err)
	}

	// Get tenant metrics
	metrics, err := tr.repository.GetTenantMetrics(ctx, tenantID)
	if err != nil {
		logger.WithComponent("tenant-resolver").Warn("Failed to get tenant metrics",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
	}

	// Build permissions based on tier and settings
	permissions := tr.buildPermissions(tenant)

	tc := &TenantContext{
		Tenant:      tenant,
		Isolation:   *isolationCtx,
		Permissions: permissions,
		Metrics:     metrics,
		RequestID:   extractRequestID(ctx),
		UserID:      extractUserID(ctx),
		SessionID:   extractSessionID(ctx),
	}

	logger.WithComponent("tenant-resolver").Debug("Tenant resolved",
		zap.String("tenant_id", tenantID),
		zap.String("model", string(tenant.Model)),
		zap.String("tier", string(tenant.Tier)),
		zap.String("database_shard", isolationCtx.DatabaseShard))

	return tc, nil
}

// ResolveTenantByDomain resolves tenant by domain name
func (tr *TenantResolver) ResolveTenantByDomain(ctx context.Context, domain string) (*TenantContext, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	tenant, err := tr.repository.GetTenantByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant by domain %s: %w", domain, err)
	}

	return tr.ResolveTenant(ctx, tenant.ID)
}

// getTenant retrieves tenant from cache or repository
func (tr *TenantResolver) getTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	// Check cache first
	tr.cacheMutex.RLock()
	cachedTenant, exists := tr.cache[tenantID]
	tr.cacheMutex.RUnlock()

	if exists {
		// Check if cache entry is still valid
		tr.refreshMutex.RLock()
		lastRefresh, hasRefreshTime := tr.lastRefresh[tenantID]
		tr.refreshMutex.RUnlock()

		if hasRefreshTime && time.Since(lastRefresh) < tr.cacheTTL {
			return cachedTenant, nil
		}
	}

	// Fetch from repository
	tenant, err := tr.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Update cache
	tr.cacheMutex.Lock()
	tr.cache[tenantID] = tenant
	tr.cacheMutex.Unlock()

	tr.refreshMutex.Lock()
	tr.lastRefresh[tenantID] = time.Now()
	tr.refreshMutex.Unlock()

	return tenant, nil
}

// buildIsolationContext creates the appropriate isolation context based on tenant model
func (tr *TenantResolver) buildIsolationContext(tenant *Tenant) (*IsolationContext, error) {
	isolationCtx := &IsolationContext{
		Model:        tenant.Model,
		RoutingTags:  make(map[string]string),
	}

	switch tenant.Model {
	case TenantModelPooled:
		// Pooled model: shared infrastructure with logical separation
		isolationCtx.DatabaseShard = tr.selectPooledDatabaseShard(tenant)
		isolationCtx.StorageBucket = tr.selectPooledStorageBucket(tenant)
		isolationCtx.ComputeNodes = tr.selectPooledComputeNodes(tenant)
		isolationCtx.RoutingTags["isolation"] = "pooled"
		isolationCtx.RoutingTags["tier"] = string(tenant.Tier)

	case TenantModelSilo:
		// Silo model: dedicated infrastructure
		isolationCtx.DatabaseShard = fmt.Sprintf("silo_%s", tenant.ID)
		isolationCtx.StorageBucket = fmt.Sprintf("silo-%s-storage", tenant.ID)
		isolationCtx.ComputeNodes = tenant.Resources.DedicatedNodes
		isolationCtx.NetworkSegment = tenant.Resources.NetworkSegment
		isolationCtx.EncryptionKey = tenant.Settings.CustomKMSKey
		isolationCtx.RoutingTags["isolation"] = "silo"
		isolationCtx.RoutingTags["tenant_id"] = tenant.ID

	case TenantModelBridge:
		// Bridge model: hybrid approach based on tenant settings
		if tenant.Settings.DataIsolation {
			isolationCtx.DatabaseShard = fmt.Sprintf("bridge_%s", tenant.ID)
		} else {
			isolationCtx.DatabaseShard = tr.selectPooledDatabaseShard(tenant)
		}

		if tenant.Settings.StorageIsolation {
			isolationCtx.StorageBucket = fmt.Sprintf("bridge-%s-storage", tenant.ID)
		} else {
			isolationCtx.StorageBucket = tr.selectPooledStorageBucket(tenant)
		}

		if tenant.Settings.ComputeIsolation {
			isolationCtx.ComputeNodes = tenant.Resources.DedicatedNodes
		} else {
			isolationCtx.ComputeNodes = tr.selectPooledComputeNodes(tenant)
		}

		if tenant.Settings.NetworkIsolation {
			isolationCtx.NetworkSegment = tenant.Resources.NetworkSegment
		}

		if tenant.Settings.EncryptionAtRest {
			isolationCtx.EncryptionKey = tenant.Settings.CustomKMSKey
		}

		isolationCtx.RoutingTags["isolation"] = "bridge"
		isolationCtx.RoutingTags["tier"] = string(tenant.Tier)
		isolationCtx.RoutingTags["data_isolation"] = fmt.Sprintf("%t", tenant.Settings.DataIsolation)

	default:
		return nil, fmt.Errorf("unsupported tenant model: %s", tenant.Model)
	}

	return isolationCtx, nil
}

// buildPermissions creates permission map based on tenant tier and settings
func (tr *TenantResolver) buildPermissions(tenant *Tenant) map[string]bool {
	permissions := make(map[string]bool)

	// Base permissions for all tenants
	permissions["read_projects"] = true
	permissions["create_projects"] = true
	permissions["validate_code"] = true

	// Tier-based permissions
	switch tenant.Tier {
	case TenantTierFree:
		permissions["max_projects"] = false // Limited by quota
		permissions["advanced_validation"] = false
		permissions["custom_rules"] = false
		permissions["priority_support"] = false

	case TenantTierStandard:
		permissions["advanced_validation"] = true
		permissions["batch_operations"] = true
		permissions["api_access"] = true
		permissions["custom_rules"] = false

	case TenantTierPremium:
		permissions["advanced_validation"] = true
		permissions["batch_operations"] = true
		permissions["api_access"] = true
		permissions["custom_rules"] = true
		permissions["priority_processing"] = true
		permissions["audit_logs"] = true

	case TenantTierEnterprise:
		permissions["advanced_validation"] = true
		permissions["batch_operations"] = true
		permissions["api_access"] = true
		permissions["custom_rules"] = true
		permissions["priority_processing"] = true
		permissions["audit_logs"] = true
		permissions["sso_integration"] = true
		permissions["compliance_reports"] = true
		permissions["dedicated_support"] = true
		permissions["custom_integrations"] = true
	}

	// Feature-specific permissions
	for _, feature := range tenant.Settings.EnabledFeatures {
		permissions[feature] = true
	}

	for _, feature := range tenant.Settings.DisabledFeatures {
		permissions[feature] = false
	}

	return permissions
}

// Resource selection helpers for pooled tenants
func (tr *TenantResolver) selectPooledDatabaseShard(tenant *Tenant) string {
	// Simple hash-based sharding for now
	// In production, this would use more sophisticated load balancing
	hashValue := tr.simpleHash(tenant.ID)
	shardCount := 4 // Number of database shards
	shardIndex := hashValue % shardCount
	return fmt.Sprintf("pooled_shard_%d", shardIndex)
}

func (tr *TenantResolver) selectPooledStorageBucket(tenant *Tenant) string {
	// Storage bucket selection based on tier and region
	switch tenant.Tier {
	case TenantTierFree, TenantTierStandard:
		return "pooled-standard-storage"
	case TenantTierPremium, TenantTierEnterprise:
		return "pooled-premium-storage"
	default:
		return "pooled-default-storage"
	}
}

func (tr *TenantResolver) selectPooledComputeNodes(tenant *Tenant) []string {
	// Compute node selection based on tier
	switch tenant.Tier {
	case TenantTierFree:
		return []string{"pooled-basic-node-1"}
	case TenantTierStandard:
		return []string{"pooled-standard-node-1", "pooled-standard-node-2"}
	case TenantTierPremium:
		return []string{"pooled-premium-node-1", "pooled-premium-node-2", "pooled-premium-node-3"}
	case TenantTierEnterprise:
		return []string{"pooled-enterprise-node-1", "pooled-enterprise-node-2", "pooled-enterprise-node-3", "pooled-enterprise-node-4"}
	default:
		return []string{"pooled-default-node-1"}
	}
}

// Simple hash function for tenant distribution
func (tr *TenantResolver) simpleHash(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// Context extraction helpers
func extractRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func extractUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func extractSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

// InvalidateTenantCache removes a tenant from cache (useful when tenant is updated)
func (tr *TenantResolver) InvalidateTenantCache(tenantID string) {
	tr.cacheMutex.Lock()
	delete(tr.cache, tenantID)
	tr.cacheMutex.Unlock()

	tr.refreshMutex.Lock()
	delete(tr.lastRefresh, tenantID)
	tr.refreshMutex.Unlock()

	logger.WithComponent("tenant-resolver").Debug("Tenant cache invalidated",
		zap.String("tenant_id", tenantID))
}

// ClearCache clears all cached tenants
func (tr *TenantResolver) ClearCache() {
	tr.cacheMutex.Lock()
	tr.cache = make(map[string]*Tenant)
	tr.cacheMutex.Unlock()

	tr.refreshMutex.Lock()
	tr.lastRefresh = make(map[string]time.Time)
	tr.refreshMutex.Unlock()

	logger.WithComponent("tenant-resolver").Info("Tenant cache cleared")
}

// GetCacheStats returns cache statistics
func (tr *TenantResolver) GetCacheStats() map[string]interface{} {
	tr.cacheMutex.RLock()
	cacheSize := len(tr.cache)
	tr.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_size": cacheSize,
		"cache_ttl_seconds": int(tr.cacheTTL.Seconds()),
	}
}