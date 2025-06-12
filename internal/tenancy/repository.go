package tenancy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"QLP/internal/logger"
)

// PostgresTenantRepository implements TenantRepository using PostgreSQL
type PostgresTenantRepository struct {
	db *sql.DB
}

// NewPostgresTenantRepository creates a new PostgreSQL tenant repository
func NewPostgresTenantRepository(db *sql.DB) *PostgresTenantRepository {
	return &PostgresTenantRepository{
		db: db,
	}
}

// GetTenant retrieves a tenant by ID
func (r *PostgresTenantRepository) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	query := `
		SELECT 
			id, name, domain, model, tier, status,
			settings, resources, metadata,
			created_at, updated_at, activated_at
		FROM tenants 
		WHERE id = $1 AND status != 'deleted'
	`

	var tenant Tenant
	var settingsJSON, resourcesJSON, metadataJSON []byte
	var activatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Domain,
		&tenant.Model,
		&tenant.Tier,
		&tenant.Status,
		&settingsJSON,
		&resourcesJSON,
		&metadataJSON,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&activatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant %s not found", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
		return nil, fmt.Errorf("failed to parse tenant settings: %w", err)
	}

	if err := json.Unmarshal(resourcesJSON, &tenant.Resources); err != nil {
		return nil, fmt.Errorf("failed to parse tenant resources: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse tenant metadata: %w", err)
		}
	}

	if activatedAt.Valid {
		tenant.ActivatedAt = &activatedAt.Time
	}

	return &tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain name
func (r *PostgresTenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	query := `
		SELECT 
			id, name, domain, model, tier, status,
			settings, resources, metadata,
			created_at, updated_at, activated_at
		FROM tenants 
		WHERE domain = $1 AND status = 'active'
	`

	var tenant Tenant
	var settingsJSON, resourcesJSON, metadataJSON []byte
	var activatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Domain,
		&tenant.Model,
		&tenant.Tier,
		&tenant.Status,
		&settingsJSON,
		&resourcesJSON,
		&metadataJSON,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&activatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found for domain %s", domain)
		}
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
		return nil, fmt.Errorf("failed to parse tenant settings: %w", err)
	}

	if err := json.Unmarshal(resourcesJSON, &tenant.Resources); err != nil {
		return nil, fmt.Errorf("failed to parse tenant resources: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse tenant metadata: %w", err)
		}
	}

	if activatedAt.Valid {
		tenant.ActivatedAt = &activatedAt.Time
	}

	return &tenant, nil
}

// ListTenants retrieves tenants with filtering
func (r *PostgresTenantRepository) ListTenants(ctx context.Context, filters TenantFilters) ([]*Tenant, error) {
	query := `
		SELECT 
			id, name, domain, model, tier, status,
			settings, resources, metadata,
			created_at, updated_at, activated_at
		FROM tenants 
		WHERE status != 'deleted'
	`
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if filters.Model != "" {
		query += fmt.Sprintf(" AND model = $%d", argIndex)
		args = append(args, filters.Model)
		argIndex++
	}

	if filters.Tier != "" {
		query += fmt.Sprintf(" AND tier = $%d", argIndex)
		args = append(args, filters.Tier)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		var settingsJSON, resourcesJSON, metadataJSON []byte
		var activatedAt sql.NullTime

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Domain,
			&tenant.Model,
			&tenant.Tier,
			&tenant.Status,
			&settingsJSON,
			&resourcesJSON,
			&metadataJSON,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&activatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
			logger.WithComponent("tenant-repository").Warn("Failed to parse tenant settings",
				zap.String("tenant_id", tenant.ID),
				zap.Error(err))
			continue
		}

		if err := json.Unmarshal(resourcesJSON, &tenant.Resources); err != nil {
			logger.WithComponent("tenant-repository").Warn("Failed to parse tenant resources",
				zap.String("tenant_id", tenant.ID),
				zap.Error(err))
			continue
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &tenant.Metadata); err != nil {
				logger.WithComponent("tenant-repository").Warn("Failed to parse tenant metadata",
					zap.String("tenant_id", tenant.ID),
					zap.Error(err))
			}
		}

		if activatedAt.Valid {
			tenant.ActivatedAt = &activatedAt.Time
		}

		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// CreateTenant creates a new tenant
func (r *PostgresTenantRepository) CreateTenant(ctx context.Context, tenant *Tenant) error {
	query := `
		INSERT INTO tenants (
			id, name, domain, model, tier, status,
			settings, resources, metadata,
			created_at, updated_at, activated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	// Marshal JSON fields
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant settings: %w", err)
	}

	resourcesJSON, err := json.Marshal(tenant.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant resources: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Domain,
		tenant.Model,
		tenant.Tier,
		tenant.Status,
		settingsJSON,
		resourcesJSON,
		metadataJSON,
		tenant.CreatedAt,
		tenant.UpdatedAt,
		tenant.ActivatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("tenant with ID %s already exists", tenant.ID)
			}
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// UpdateTenant updates an existing tenant
func (r *PostgresTenantRepository) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	query := `
		UPDATE tenants SET
			name = $2,
			domain = $3,
			model = $4,
			tier = $5,
			status = $6,
			settings = $7,
			resources = $8,
			metadata = $9,
			updated_at = $10,
			activated_at = $11
		WHERE id = $1
	`

	// Marshal JSON fields
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant settings: %w", err)
	}

	resourcesJSON, err := json.Marshal(tenant.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant resources: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant metadata: %w", err)
	}

	tenant.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Domain,
		tenant.Model,
		tenant.Tier,
		tenant.Status,
		settingsJSON,
		resourcesJSON,
		metadataJSON,
		tenant.UpdatedAt,
		tenant.ActivatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant %s not found", tenant.ID)
	}

	return nil
}

// DeleteTenant soft deletes a tenant
func (r *PostgresTenantRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	query := `
		UPDATE tenants SET
			status = 'deleted',
			updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant %s not found", tenantID)
	}

	return nil
}

// GetTenantMetrics retrieves current metrics for a tenant
func (r *PostgresTenantRepository) GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error) {
	query := `
		SELECT 
			active_requests,
			total_requests,
			error_rate,
			avg_response_time,
			resource_utilization,
			last_activity
		FROM tenant_metrics 
		WHERE tenant_id = $1
	`

	var metrics TenantMetrics
	var resourceUtilizationJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&metrics.ActiveRequests,
		&metrics.TotalRequests,
		&metrics.ErrorRate,
		&metrics.AvgResponseTime,
		&resourceUtilizationJSON,
		&metrics.LastActivity,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default metrics if none exist
			return &TenantMetrics{
				ActiveRequests:      0,
				TotalRequests:       0,
				ErrorRate:           0.0,
				AvgResponseTime:     0.0,
				ResourceUtilization: make(map[string]float64),
				LastActivity:        time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get tenant metrics: %w", err)
	}

	// Parse resource utilization JSON
	if len(resourceUtilizationJSON) > 0 {
		if err := json.Unmarshal(resourceUtilizationJSON, &metrics.ResourceUtilization); err != nil {
			logger.WithComponent("tenant-repository").Warn("Failed to parse resource utilization",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			metrics.ResourceUtilization = make(map[string]float64)
		}
	} else {
		metrics.ResourceUtilization = make(map[string]float64)
	}

	return &metrics, nil
}

// UpdateTenantMetrics updates tenant metrics
func (r *PostgresTenantRepository) UpdateTenantMetrics(ctx context.Context, tenantID string, metrics *TenantMetrics) error {
	query := `
		INSERT INTO tenant_metrics (
			tenant_id, active_requests, total_requests, error_rate,
			avg_response_time, resource_utilization, last_activity
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			active_requests = EXCLUDED.active_requests,
			total_requests = EXCLUDED.total_requests,
			error_rate = EXCLUDED.error_rate,
			avg_response_time = EXCLUDED.avg_response_time,
			resource_utilization = EXCLUDED.resource_utilization,
			last_activity = EXCLUDED.last_activity
	`

	resourceUtilizationJSON, err := json.Marshal(metrics.ResourceUtilization)
	if err != nil {
		return fmt.Errorf("failed to marshal resource utilization: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tenantID,
		metrics.ActiveRequests,
		metrics.TotalRequests,
		metrics.ErrorRate,
		metrics.AvgResponseTime,
		resourceUtilizationJSON,
		metrics.LastActivity,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant metrics: %w", err)
	}

	return nil
}