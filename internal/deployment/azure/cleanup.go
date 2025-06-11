package azure

import (
	"context"
	"fmt"
	"time"

	"QLP/internal/logger"
	"go.uber.org/zap"
)

// CleanupManager handles automated cleanup of Azure resources
type CleanupManager struct {
	logger      logger.Interface
	azureClient *AzureClient
}

// CleanupPolicy defines how resources should be cleaned up
type CleanupPolicy struct {
	MaxAge              time.Duration // Maximum age before forced cleanup
	CheckInterval       time.Duration // How often to check for expired resources
	GracePeriod         time.Duration // Grace period after TTL expiration
	CostThreshold       float64       // Cleanup if cost exceeds threshold
	RetryAttempts       int           // Number of retry attempts for failed cleanups
	RetryDelay          time.Duration // Delay between retry attempts
	DryRun              bool          // If true, log actions but don't execute
	PreserveOnError     bool          // If true, don't cleanup resources with errors
	NotificationWebhook string        // Webhook URL for cleanup notifications
}

// CleanupResult tracks the results of cleanup operations
type CleanupResult struct {
	ResourceGroup     string        `json:"resource_group"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
	Duration          time.Duration `json:"duration"`
	Status            string        `json:"status"` // "success", "failed", "partial"
	ResourcesDeleted  int           `json:"resources_deleted"`
	ErrorsEncountered []string      `json:"errors_encountered"`
	CostSaved         float64       `json:"cost_saved_usd"`
	RetryAttempts     int           `json:"retry_attempts"`
}

// ExpiredResource represents a resource that should be cleaned up
type ExpiredResource struct {
	ResourceGroup   string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	CapsuleID       string
	CostEstimate    float64
	LastActivity    time.Time
	HasErrors       bool
	ErrorMessages   []string
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(azureClient *AzureClient) *CleanupManager {
	return &CleanupManager{
		logger:      logger.GetDefaultLogger().WithComponent("azure_cleanup"),
		azureClient: azureClient,
	}
}

// StartCleanupScheduler starts the automatic cleanup scheduler
func (cm *CleanupManager) StartCleanupScheduler(ctx context.Context, policy CleanupPolicy) {
	cm.logger.Info("Starting cleanup scheduler",
		zap.Duration("check_interval", policy.CheckInterval),
		zap.Duration("max_age", policy.MaxAge),
		zap.Bool("dry_run", policy.DryRun),
	)

	ticker := time.NewTicker(policy.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Cleanup scheduler stopped")
			return
		case <-ticker.C:
			cm.performScheduledCleanup(ctx, policy)
		}
	}
}

// performScheduledCleanup performs a scheduled cleanup check
func (cm *CleanupManager) performScheduledCleanup(ctx context.Context, policy CleanupPolicy) {
	cm.logger.Debug("Performing scheduled cleanup check")

	expiredResources, err := cm.findExpiredResources(ctx, policy)
	if err != nil {
		cm.logger.Error("Failed to find expired resources", zap.Error(err))
		return
	}

	if len(expiredResources) == 0 {
		cm.logger.Debug("No expired resources found")
		return
	}

	cm.logger.Info("Found expired resources for cleanup",
		zap.Int("count", len(expiredResources)),
	)

	for _, resource := range expiredResources {
		result := cm.cleanupResource(ctx, resource, policy)
		cm.logCleanupResult(resource, result)
		
		// Send notification if webhook is configured
		if policy.NotificationWebhook != "" {
			cm.sendCleanupNotification(policy.NotificationWebhook, resource, result)
		}
	}
}

// findExpiredResources identifies resources that should be cleaned up
func (cm *CleanupManager) findExpiredResources(ctx context.Context, policy CleanupPolicy) ([]ExpiredResource, error) {
	resourceGroups, err := cm.azureClient.ListResourceGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %w", err)
	}

	var expiredResources []ExpiredResource
	now := time.Now()

	for _, rg := range resourceGroups {
		if rg.Tags == nil {
			continue
		}

		// Check if this is a QuantumLayer resource group
		createdBy, exists := rg.Tags["created-by"]
		if !exists || *createdBy != "quantumlayer" {
			continue
		}

		resource := ExpiredResource{
			ResourceGroup: *rg.Name,
		}

		// Parse creation time and expiration
		if expirationStr, exists := rg.Tags["auto-delete-after"]; exists {
			if expiresAt, err := time.Parse(time.RFC3339, *expirationStr); err == nil {
				resource.ExpiresAt = expiresAt
			}
		}

		// Parse capsule ID
		if capsuleID, exists := rg.Tags["capsule-id"]; exists {
			resource.CapsuleID = *capsuleID
		}

		// Check if resource is expired
		isExpired := false
		
		// Check TTL expiration
		if !resource.ExpiresAt.IsZero() && now.After(resource.ExpiresAt.Add(policy.GracePeriod)) {
			isExpired = true
		}
		
		// Check max age
		if !resource.CreatedAt.IsZero() && now.Sub(resource.CreatedAt) > policy.MaxAge {
			isExpired = true
		}

		// Check cost threshold (if available)
		// TODO: Implement cost checking via Azure Cost Management API

		// Skip if preserve on error is enabled and resource has errors
		if policy.PreserveOnError && resource.HasErrors {
			cm.logger.Info("Skipping cleanup of resource with errors",
				zap.String("resource_group", resource.ResourceGroup),
				zap.Strings("errors", resource.ErrorMessages),
			)
			continue
		}

		if isExpired {
			expiredResources = append(expiredResources, resource)
		}
	}

	return expiredResources, nil
}

// cleanupResource performs cleanup for a single expired resource
func (cm *CleanupManager) cleanupResource(ctx context.Context, resource ExpiredResource, policy CleanupPolicy) CleanupResult {
	result := CleanupResult{
		ResourceGroup: resource.ResourceGroup,
		StartTime:     time.Now(),
		Status:        "failed",
	}

	cm.logger.Info("Starting cleanup of expired resource",
		zap.String("resource_group", resource.ResourceGroup),
		zap.String("capsule_id", resource.CapsuleID),
		zap.Time("expires_at", resource.ExpiresAt),
		zap.Bool("dry_run", policy.DryRun),
	)

	if policy.DryRun {
		cm.logger.Info("DRY RUN: Would delete resource group",
			zap.String("resource_group", resource.ResourceGroup),
		)
		result.Status = "dry_run"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Attempt cleanup with retries
	var lastErr error
	for attempt := 0; attempt <= policy.RetryAttempts; attempt++ {
		if attempt > 0 {
			cm.logger.Info("Retrying cleanup",
				zap.String("resource_group", resource.ResourceGroup),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", policy.RetryAttempts),
			)
			time.Sleep(policy.RetryDelay)
		}

		err := cm.azureClient.DeleteResourceGroup(ctx, resource.ResourceGroup)
		if err == nil {
			result.Status = "success"
			result.ResourcesDeleted = 1 // TODO: Count actual resources
			break
		}

		lastErr = err
		result.RetryAttempts = attempt + 1
		result.ErrorsEncountered = append(result.ErrorsEncountered, err.Error())

		cm.logger.Warn("Cleanup attempt failed",
			zap.String("resource_group", resource.ResourceGroup),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
	}

	if lastErr != nil {
		cm.logger.Error("All cleanup attempts failed",
			zap.String("resource_group", resource.ResourceGroup),
			zap.Int("attempts", result.RetryAttempts),
			zap.Error(lastErr),
		)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.CostSaved = resource.CostEstimate // Estimated cost savings

	return result
}

// logCleanupResult logs the result of a cleanup operation
func (cm *CleanupManager) logCleanupResult(resource ExpiredResource, result CleanupResult) {
	fields := []zap.Field{
		zap.String("resource_group", result.ResourceGroup),
		zap.String("status", result.Status),
		zap.Duration("duration", result.Duration),
		zap.Int("retry_attempts", result.RetryAttempts),
	}

	if result.Status == "success" {
		fields = append(fields,
			zap.Int("resources_deleted", result.ResourcesDeleted),
			zap.Float64("cost_saved_usd", result.CostSaved),
		)
		cm.logger.Info("Cleanup completed successfully", fields...)
	} else {
		fields = append(fields, zap.Strings("errors", result.ErrorsEncountered))
		cm.logger.Error("Cleanup failed", fields...)
	}
}

// sendCleanupNotification sends a notification about cleanup results
func (cm *CleanupManager) sendCleanupNotification(webhookURL string, resource ExpiredResource, result CleanupResult) {
	// TODO: Implement webhook notification
	// This could send to Slack, Teams, or custom webhook endpoints
	cm.logger.Debug("Sending cleanup notification",
		zap.String("webhook_url", webhookURL),
		zap.String("resource_group", result.ResourceGroup),
		zap.String("status", result.Status),
	)
}

// CleanupByTag performs cleanup of resources matching specific tags
func (cm *CleanupManager) CleanupByTag(ctx context.Context, tagKey, tagValue string, policy CleanupPolicy) ([]CleanupResult, error) {
	cm.logger.Info("Starting cleanup by tag",
		zap.String("tag_key", tagKey),
		zap.String("tag_value", tagValue),
	)

	resourceGroups, err := cm.azureClient.ListResourceGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list resource groups: %w", err)
	}

	var results []CleanupResult

	for _, rg := range resourceGroups {
		if rg.Tags == nil {
			continue
		}

		// Check if resource group has the specified tag
		if value, exists := rg.Tags[tagKey]; exists && *value == tagValue {
			resource := ExpiredResource{
				ResourceGroup: *rg.Name,
				HasErrors:     false, // Assume no errors for tag-based cleanup
			}

			// Parse capsule ID if available
			if capsuleID, exists := rg.Tags["capsule-id"]; exists {
				resource.CapsuleID = *capsuleID
			}

			result := cm.cleanupResource(ctx, resource, policy)
			results = append(results, result)
		}
	}

	return results, nil
}

// ForceCleanup performs immediate cleanup of a specific resource group
func (cm *CleanupManager) ForceCleanup(ctx context.Context, resourceGroup string) CleanupResult {
	cm.logger.Info("Performing force cleanup",
		zap.String("resource_group", resourceGroup),
	)

	resource := ExpiredResource{
		ResourceGroup: resourceGroup,
		ExpiresAt:     time.Now(), // Mark as expired now
	}

	policy := CleanupPolicy{
		RetryAttempts: 3,
		RetryDelay:    30 * time.Second,
		DryRun:        false,
	}

	return cm.cleanupResource(ctx, resource, policy)
}

// GetCleanupStats returns statistics about cleanup operations
func (cm *CleanupManager) GetCleanupStats(ctx context.Context) (map[string]interface{}, error) {
	resourceGroups, err := cm.azureClient.ListResourceGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleanup stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_resource_groups": len(resourceGroups),
		"expired_count":         0,
		"cost_estimate":         0.0,
		"oldest_resource":       "",
		"newest_resource":       "",
	}

	now := time.Now()
	var oldestTime, newestTime time.Time

	for _, rg := range resourceGroups {
		if rg.Tags == nil {
			continue
		}

		// Check TTL expiration
		if expirationStr, exists := rg.Tags["auto-delete-after"]; exists {
			if expiresAt, err := time.Parse(time.RFC3339, *expirationStr); err == nil {
				if now.After(expiresAt) {
					stats["expired_count"] = stats["expired_count"].(int) + 1
				}

				// Track oldest and newest
				if oldestTime.IsZero() || expiresAt.Before(oldestTime) {
					oldestTime = expiresAt
					stats["oldest_resource"] = *rg.Name
				}
				if newestTime.IsZero() || expiresAt.After(newestTime) {
					newestTime = expiresAt
					stats["newest_resource"] = *rg.Name
				}
			}
		}
	}

	return stats, nil
}

// DefaultCleanupPolicy returns a sensible default cleanup policy
func DefaultCleanupPolicy() CleanupPolicy {
	return CleanupPolicy{
		MaxAge:              24 * time.Hour,      // Max 24 hours
		CheckInterval:       15 * time.Minute,    // Check every 15 minutes
		GracePeriod:         5 * time.Minute,     // 5 minute grace period
		CostThreshold:       10.0,                // $10 USD threshold
		RetryAttempts:       3,                   // 3 retry attempts
		RetryDelay:          30 * time.Second,    // 30 second delay between retries
		DryRun:              false,               // Actually perform cleanup
		PreserveOnError:     true,                // Preserve resources with errors
		NotificationWebhook: "",                  // No notifications by default
	}
}