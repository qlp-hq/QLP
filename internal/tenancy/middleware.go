package tenancy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/internal/logger"
)

// TenantContextKey is the context key for tenant information
type TenantContextKey string

const (
	TenantContextKeyTenant   TenantContextKey = "tenant"
	TenantContextKeyContext  TenantContextKey = "tenant_context"
	TenantContextKeyRequestID TenantContextKey = "request_id"
	TenantContextKeyUserID   TenantContextKey = "user_id"
	TenantContextKeySessionID TenantContextKey = "session_id"
)

// TenantMiddleware provides tenant resolution middleware for HTTP requests
type TenantMiddleware struct {
	resolver *TenantResolver
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(resolver *TenantResolver) *TenantMiddleware {
	return &TenantMiddleware{
		resolver: resolver,
	}
}

// ResolveTenantFromURL middleware extracts tenant ID from URL path and resolves tenant context
func (tm *TenantMiddleware) ResolveTenantFromURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "tenantId")
		if tenantID == "" {
			logger.WithComponent("tenant-middleware").Warn("No tenant ID in URL",
				zap.String("path", r.URL.Path))
			http.Error(w, "Tenant ID is required", http.StatusBadRequest)
			return
		}

		// Add request context information
		ctx := tm.enrichContext(r.Context(), r)

		// Resolve tenant
		tenantCtx, err := tm.resolver.ResolveTenant(ctx, tenantID)
		if err != nil {
			logger.WithComponent("tenant-middleware").Error("Failed to resolve tenant",
				zap.String("tenant_id", tenantID),
				zap.Error(err))

			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Tenant not found", http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not active") {
				http.Error(w, "Tenant is not active", http.StatusForbidden)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Add tenant context to request context
		ctx = context.WithValue(ctx, TenantContextKeyTenant, tenantCtx.Tenant)
		ctx = context.WithValue(ctx, TenantContextKeyContext, tenantCtx)

		// Add tenant-specific headers for downstream services
		w.Header().Set("X-Tenant-ID", tenantCtx.Tenant.ID)
		w.Header().Set("X-Tenant-Model", string(tenantCtx.Tenant.Model))
		w.Header().Set("X-Tenant-Tier", string(tenantCtx.Tenant.Tier))
		w.Header().Set("X-Database-Shard", tenantCtx.Isolation.DatabaseShard)
		w.Header().Set("X-Storage-Bucket", tenantCtx.Isolation.StorageBucket)

		// Log tenant resolution
		logger.WithComponent("tenant-middleware").Debug("Tenant resolved",
			zap.String("tenant_id", tenantCtx.Tenant.ID),
			zap.String("model", string(tenantCtx.Tenant.Model)),
			zap.String("tier", string(tenantCtx.Tenant.Tier)),
			zap.String("database_shard", tenantCtx.Isolation.DatabaseShard))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ResolveTenantFromDomain middleware extracts tenant from domain name
func (tm *TenantMiddleware) ResolveTenantFromDomain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := tm.extractDomain(r)
		if domain == "" {
			logger.WithComponent("tenant-middleware").Warn("No domain in request",
				zap.String("host", r.Host))
			http.Error(w, "Domain is required", http.StatusBadRequest)
			return
		}

		// Add request context information
		ctx := tm.enrichContext(r.Context(), r)

		// Resolve tenant by domain
		tenantCtx, err := tm.resolver.ResolveTenantByDomain(ctx, domain)
		if err != nil {
			logger.WithComponent("tenant-middleware").Error("Failed to resolve tenant by domain",
				zap.String("domain", domain),
				zap.Error(err))

			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Tenant not found for domain", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Add tenant context to request context
		ctx = context.WithValue(ctx, TenantContextKeyTenant, tenantCtx.Tenant)
		ctx = context.WithValue(ctx, TenantContextKeyContext, tenantCtx)

		// Add tenant-specific headers
		w.Header().Set("X-Tenant-ID", tenantCtx.Tenant.ID)
		w.Header().Set("X-Tenant-Domain", tenantCtx.Tenant.Domain)
		w.Header().Set("X-Tenant-Model", string(tenantCtx.Tenant.Model))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware checks if tenant has required permission
func (tm *TenantMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantCtx := GetTenantContextFromRequest(r)
			if tenantCtx == nil {
				logger.WithComponent("tenant-middleware").Error("No tenant context in permission check")
				http.Error(w, "Tenant context required", http.StatusInternalServerError)
				return
			}

			if !tenantCtx.Permissions[permission] {
				logger.WithComponent("tenant-middleware").Warn("Permission denied",
					zap.String("tenant_id", tenantCtx.Tenant.ID),
					zap.String("permission", permission),
					zap.String("tier", string(tenantCtx.Tenant.Tier)))

				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnforceResourceQuota middleware checks resource quotas
func (tm *TenantMiddleware) EnforceResourceQuota(resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantCtx := GetTenantContextFromRequest(r)
			if tenantCtx == nil {
				http.Error(w, "Tenant context required", http.StatusInternalServerError)
				return
			}

			// Check resource quota based on type
			exceeded, err := tm.checkResourceQuota(tenantCtx, resourceType)
			if err != nil {
				logger.WithComponent("tenant-middleware").Error("Error checking resource quota",
					zap.String("tenant_id", tenantCtx.Tenant.ID),
					zap.String("resource_type", resourceType),
					zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if exceeded {
				logger.WithComponent("tenant-middleware").Warn("Resource quota exceeded",
					zap.String("tenant_id", tenantCtx.Tenant.ID),
					zap.String("resource_type", resourceType))

				http.Error(w, "Resource quota exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// enrichContext adds request-specific information to context
func (tm *TenantMiddleware) enrichContext(ctx context.Context, r *http.Request) context.Context {
	// Extract request ID from header or generate one
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}
	ctx = context.WithValue(ctx, TenantContextKeyRequestID, requestID)

	// Extract user ID from header (would come from auth middleware)
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		ctx = context.WithValue(ctx, TenantContextKeyUserID, userID)
	}

	// Extract session ID from header
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID != "" {
		ctx = context.WithValue(ctx, TenantContextKeySessionID, sessionID)
	}

	return ctx
}

// extractDomain extracts domain from request host
func (tm *TenantMiddleware) extractDomain(r *http.Request) string {
	host := r.Host
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

// checkResourceQuota checks if resource quota is exceeded
func (tm *TenantMiddleware) checkResourceQuota(tenantCtx *TenantContext, resourceType string) (bool, error) {
	quota := tenantCtx.Tenant.Settings.ResourceQuota
	resources := tenantCtx.Tenant.Resources

	switch resourceType {
	case "requests":
		return resources.CurrentRequests >= quota.MaxRequests, nil
	case "projects":
		return resources.CurrentProjects >= quota.MaxProjects, nil
	case "artifacts":
		return resources.CurrentArtifacts >= quota.MaxArtifacts, nil
	case "cpu":
		return resources.CurrentCPU >= quota.MaxCPU, nil
	case "memory":
		return resources.CurrentMemory >= quota.MaxMemory, nil
	case "storage":
		return resources.CurrentStorage >= quota.MaxStorage, nil
	default:
		return false, nil // Unknown resource type, allow by default
	}
}

// Helper functions

// GetTenantFromRequest extracts tenant from request context
func GetTenantFromRequest(r *http.Request) *Tenant {
	if tenant, ok := r.Context().Value(TenantContextKeyTenant).(*Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantContextFromRequest extracts full tenant context from request
func GetTenantContextFromRequest(r *http.Request) *TenantContext {
	if tenantCtx, ok := r.Context().Value(TenantContextKeyContext).(*TenantContext); ok {
		return tenantCtx
	}
	return nil
}

// GetTenantIDFromRequest extracts tenant ID from request context
func GetTenantIDFromRequest(r *http.Request) string {
	if tenant := GetTenantFromRequest(r); tenant != nil {
		return tenant.ID
	}
	return ""
}

// CheckPermission checks if current tenant has permission
func CheckPermission(r *http.Request, permission string) bool {
	if tenantCtx := GetTenantContextFromRequest(r); tenantCtx != nil {
		return tenantCtx.Permissions[permission]
	}
	return false
}

// GetIsolationContext gets isolation context from request
func GetIsolationContext(r *http.Request) *IsolationContext {
	if tenantCtx := GetTenantContextFromRequest(r); tenantCtx != nil {
		return &tenantCtx.Isolation
	}
	return nil
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - in production, use UUID or similar
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// TenantAwareHandler wraps handler with tenant context validation
func TenantAwareHandler(handler func(w http.ResponseWriter, r *http.Request, tenantCtx *TenantContext)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx := GetTenantContextFromRequest(r)
		if tenantCtx == nil {
			logger.WithComponent("tenant-middleware").Error("No tenant context in tenant-aware handler")
			http.Error(w, "Tenant context required", http.StatusInternalServerError)
			return
		}

		handler(w, r, tenantCtx)
	})
}

// TenantMetricsHandler returns tenant metrics as JSON
func (tm *TenantMiddleware) TenantMetricsHandler() http.Handler {
	return TenantAwareHandler(func(w http.ResponseWriter, r *http.Request, tenantCtx *TenantContext) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tenant_id":  tenantCtx.Tenant.ID,
			"model":      tenantCtx.Tenant.Model,
			"tier":       tenantCtx.Tenant.Tier,
			"status":     tenantCtx.Tenant.Status,
			"isolation":  tenantCtx.Isolation,
			"resources":  tenantCtx.Tenant.Resources,
			"metrics":    tenantCtx.Metrics,
		})
	})
}