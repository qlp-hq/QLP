package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"QLP/internal/deployment/azure"
	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/packaging"
	"QLP/internal/types"
	"go.uber.org/zap"
)

// DeploymentValidatorAgent performs real deployment validation using Azure
type DeploymentValidatorAgent struct {
	ID                string
	Type              string
	Status            AgentStatus
	Logger            logger.Interface
	deploymentManager *azure.DeploymentManager
	azureClient       *azure.AzureClient
	capsule           *packaging.QuantumDrop
	config            azure.DeploymentConfig
}

// DeploymentValidatorConfig configures the deployment validator agent
type DeploymentValidatorConfig struct {
	AzureConfig     azure.ClientConfig
	CostLimitUSD    float64
	TTL             time.Duration
	EnableHealthChecks bool
	EnableFunctionalTests bool
	CleanupPolicy   azure.CleanupPolicy
}

// NewDeploymentValidatorAgent creates a new deployment validator agent
func NewDeploymentValidatorAgent(
	agentID string,
	llmClient llm.Client,
	capsule *packaging.QuantumDrop,
	config DeploymentValidatorConfig,
) (*DeploymentValidatorAgent, error) {
	
	agentLogger := logger.GetDefaultLogger().WithComponent("deployment_validator_agent")
	
	// Create Azure client
	azureClient, err := azure.NewAzureClient(config.AzureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}
	
	// Create deployment manager
	deploymentManager := azure.NewDeploymentManager(azureClient, config.CostLimitUSD)
	
	// Generate deployment config
	deploymentConfig := azure.DeploymentConfig{
		CapsuleID:     capsule.ID,
		ResourceGroup: azure.GenerateResourceGroupName(capsule.ID),
		Location:      config.AzureConfig.Location,
		TTL:           config.TTL,
		CostLimitUSD:  config.CostLimitUSD,
		SecurityContext: azure.SecurityContext{
			ManagedIdentityOnly: true,
			NetworkIsolation:    true,
			AllowedOutbound:     []string{"https://api.github.com", "https://registry-1.docker.io"},
		},
	}
	
	agent := &DeploymentValidatorAgent{
		ID:                agentID,
		Type:              "deployment-validator",
		Status:            AgentStatusReady,
		Logger:            logger.GetDefaultLogger().WithComponent("deployment_validator_agent"),
		deploymentManager: deploymentManager,
		azureClient:       azureClient,
		capsule:           capsule,
		config:            deploymentConfig,
	}
	
	agentLogger.Info("Deployment validator agent created",
		zap.String("agent_id", agentID),
		zap.String("capsule_id", capsule.ID),
		zap.String("resource_group", deploymentConfig.ResourceGroup),
		zap.Float64("cost_limit", config.CostLimitUSD),
	)
	
	return agent, nil
}

// Execute performs the deployment validation
func (dva *DeploymentValidatorAgent) Execute(ctx context.Context, task types.Task) (*types.TaskResult, error) {
	dva.Logger.Info("Starting deployment validation",
		zap.String("task_id", task.ID),
		zap.String("capsule_id", dva.capsule.ID),
	)
	
	dva.Status = AgentStatusExecuting
	startTime := time.Now()
	
	// Create task result
	result := &types.TaskResult{
		TaskID:      task.ID,
		AgentID:     dva.ID,
		Status:      types.TaskStatusInProgress,
		StartTime:   startTime,
		Output:      "",
		Metadata:    make(map[string]interface{}),
		Attachments: make(map[string][]byte),
	}
	
	// Perform Azure deployment validation
	deploymentResult, err := dva.deploymentManager.Deploy(ctx, dva.capsule, dva.config)
	if err != nil {
		dva.Logger.Error("Deployment validation failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		
		result.Status = types.TaskStatusFailed
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		dva.Status = AgentStatusFailed
		
		return result, err
	}
	
	// Convert deployment result to task result
	dva.processDeploymentResult(deploymentResult, result)
	
	// Generate comprehensive validation report
	report, err := dva.generateValidationReport(deploymentResult)
	if err != nil {
		dva.Logger.Warn("Failed to generate validation report", zap.Error(err))
	} else {
		result.Output = report
	}
	
	// Store deployment artifacts
	dva.storeDeploymentArtifacts(deploymentResult, result)
	
	result.Status = types.TaskStatusCompleted
	result.EndTime = time.Now()
	dva.Status = AgentStatusCompleted
	
	dva.Logger.Info("Deployment validation completed",
		zap.String("task_id", task.ID),
		zap.String("capsule_id", dva.capsule.ID),
		zap.String("deployment_status", string(deploymentResult.Status)),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
		zap.Float64("cost_estimate", deploymentResult.CostEstimate.TotalUSD),
	)
	
	return result, nil
}

// processDeploymentResult converts Azure deployment result to task result metadata
func (dva *DeploymentValidatorAgent) processDeploymentResult(deploymentResult *azure.DeploymentResult, taskResult *types.TaskResult) {
	taskResult.Metadata["deployment_status"] = string(deploymentResult.Status)
	taskResult.Metadata["resource_group"] = deploymentResult.ResourceGroup
	taskResult.Metadata["deployment_duration"] = deploymentResult.Duration.String()
	taskResult.Metadata["cost_estimate_usd"] = deploymentResult.CostEstimate.TotalUSD
	taskResult.Metadata["health_checks_passed"] = dva.countPassedHealthChecks(deploymentResult.HealthChecks)
	taskResult.Metadata["total_health_checks"] = len(deploymentResult.HealthChecks)
	taskResult.Metadata["tests_passed"] = dva.countPassedTests(deploymentResult.TestResults)
	taskResult.Metadata["total_tests"] = len(deploymentResult.TestResults)
	taskResult.Metadata["logs_url"] = deploymentResult.LogsURL
	
	// Add deployment outputs
	if len(deploymentResult.DeploymentOutputs) > 0 {
		taskResult.Metadata["deployment_outputs"] = deploymentResult.DeploymentOutputs
	}
	
	// Set overall success based on deployment status
	switch deploymentResult.Status {
	case azure.StatusCompleted, azure.StatusHealthy:
		taskResult.Metadata["deployment_success"] = true
	case azure.StatusFailed, azure.StatusUnhealthy:
		taskResult.Metadata["deployment_success"] = false
		if deploymentResult.ErrorMessage != "" {
			taskResult.ErrorMessage = deploymentResult.ErrorMessage
		}
	default:
		taskResult.Metadata["deployment_success"] = false
	}
}

// generateValidationReport creates a comprehensive validation report
func (dva *DeploymentValidatorAgent) generateValidationReport(deploymentResult *azure.DeploymentResult) (string, error) {
	report := map[string]interface{}{
		"deployment_validation_report": map[string]interface{}{
			"capsule_id":       dva.capsule.ID,
			"resource_group":   deploymentResult.ResourceGroup,
			"status":           string(deploymentResult.Status),
			"start_time":       deploymentResult.StartTime.Format(time.RFC3339),
			"end_time":         deploymentResult.EndTime.Format(time.RFC3339),
			"duration_minutes": deploymentResult.Duration.Minutes(),
			"success":          deploymentResult.Status == azure.StatusCompleted || deploymentResult.Status == azure.StatusHealthy,
		},
		"cost_analysis": map[string]interface{}{
			"total_cost_usd":     deploymentResult.CostEstimate.TotalUSD,
			"resource_breakdown": deploymentResult.CostEstimate.ResourceBreakdown,
			"billing_period":     deploymentResult.CostEstimate.BillingPeriod,
			"cost_efficiency":    dva.calculateCostEfficiency(deploymentResult),
		},
		"health_checks": map[string]interface{}{
			"total":  len(deploymentResult.HealthChecks),
			"passed": dva.countPassedHealthChecks(deploymentResult.HealthChecks),
			"failed": len(deploymentResult.HealthChecks) - dva.countPassedHealthChecks(deploymentResult.HealthChecks),
			"details": deploymentResult.HealthChecks,
		},
		"functional_tests": map[string]interface{}{
			"total":   len(deploymentResult.TestResults),
			"passed":  dva.countPassedTests(deploymentResult.TestResults),
			"failed":  len(deploymentResult.TestResults) - dva.countPassedTests(deploymentResult.TestResults),
			"details": deploymentResult.TestResults,
		},
		"deployment_outputs": deploymentResult.DeploymentOutputs,
		"azure_resources": map[string]interface{}{
			"region":           dva.config.Location,
			"subscription_id":  dva.azureClient.GetSubscriptionID(),
			"cleanup_scheduled": deploymentResult.DestroyedAt != nil,
			"logs_url":         deploymentResult.LogsURL,
		},
	}
	
	// Add error information if deployment failed
	if deploymentResult.ErrorMessage != "" {
		report["error_analysis"] = map[string]interface{}{
			"error_message": deploymentResult.ErrorMessage,
			"failure_stage": dva.determineFailureStage(deploymentResult),
			"recommendations": dva.generateFailureRecommendations(deploymentResult),
		}
	}
	
	// Convert to JSON
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal validation report: %w", err)
	}
	
	return string(reportJSON), nil
}

// storeDeploymentArtifacts stores deployment-related files as task attachments
func (dva *DeploymentValidatorAgent) storeDeploymentArtifacts(deploymentResult *azure.DeploymentResult, taskResult *types.TaskResult) {
	// Store deployment result as JSON
	if resultJSON, err := json.MarshalIndent(deploymentResult, "", "  "); err == nil {
		taskResult.Attachments["deployment_result.json"] = resultJSON
	}
	
	// Store health check details
	if healthJSON, err := json.MarshalIndent(deploymentResult.HealthChecks, "", "  "); err == nil {
		taskResult.Attachments["health_checks.json"] = healthJSON
	}
	
	// Store test results
	if testJSON, err := json.MarshalIndent(deploymentResult.TestResults, "", "  "); err == nil {
		taskResult.Attachments["test_results.json"] = testJSON
	}
	
	// Store cost breakdown
	if costJSON, err := json.MarshalIndent(deploymentResult.CostEstimate, "", "  "); err == nil {
		taskResult.Attachments["cost_estimate.json"] = costJSON
	}
}

// Helper methods

func (dva *DeploymentValidatorAgent) countPassedHealthChecks(healthChecks []azure.HealthCheckResult) int {
	passed := 0
	for _, check := range healthChecks {
		if check.Status == "pass" {
			passed++
		}
	}
	return passed
}

func (dva *DeploymentValidatorAgent) countPassedTests(testResults map[string]azure.TestResult) int {
	passed := 0
	for _, test := range testResults {
		if test.Status == "pass" {
			passed++
		}
	}
	return passed
}

func (dva *DeploymentValidatorAgent) calculateCostEfficiency(deploymentResult *azure.DeploymentResult) string {
	// Simple cost efficiency calculation
	costPerMinute := deploymentResult.CostEstimate.TotalUSD / deploymentResult.Duration.Minutes()
	
	if costPerMinute < 0.01 {
		return "excellent"
	} else if costPerMinute < 0.05 {
		return "good"
	} else if costPerMinute < 0.10 {
		return "acceptable"
	} else {
		return "high"
	}
}

func (dva *DeploymentValidatorAgent) determineFailureStage(deploymentResult *azure.DeploymentResult) string {
	switch deploymentResult.Status {
	case azure.StatusFailed:
		if len(deploymentResult.HealthChecks) == 0 {
			return "deployment"
		} else if len(deploymentResult.TestResults) == 0 {
			return "health_checks"
		} else {
			return "testing"
		}
	case azure.StatusUnhealthy:
		return "health_checks"
	default:
		return "unknown"
	}
}

func (dva *DeploymentValidatorAgent) generateFailureRecommendations(deploymentResult *azure.DeploymentResult) []string {
	recommendations := []string{}
	
	// Analyze failure patterns and suggest fixes
	if deploymentResult.ErrorMessage != "" {
		if contains(deploymentResult.ErrorMessage, "timeout") {
			recommendations = append(recommendations, "Increase deployment timeout or optimize resource provisioning")
		}
		if contains(deploymentResult.ErrorMessage, "permission") || contains(deploymentResult.ErrorMessage, "unauthorized") {
			recommendations = append(recommendations, "Check Azure RBAC permissions and managed identity configuration")
		}
		if contains(deploymentResult.ErrorMessage, "quota") || contains(deploymentResult.ErrorMessage, "limit") {
			recommendations = append(recommendations, "Check Azure subscription quotas and resource limits")
		}
		if contains(deploymentResult.ErrorMessage, "network") {
			recommendations = append(recommendations, "Review network configuration and security group rules")
		}
	}
	
	// Analyze health check failures
	for _, check := range deploymentResult.HealthChecks {
		if check.Status == "fail" {
			if check.StatusCode >= 500 {
				recommendations = append(recommendations, fmt.Sprintf("Service %s is experiencing server errors - check application logs", check.Name))
			} else if check.StatusCode >= 400 {
				recommendations = append(recommendations, fmt.Sprintf("Service %s has client errors - verify API configuration", check.Name))
			}
		}
	}
	
	// Default recommendations if no specific patterns found
	if len(recommendations) == 0 {
		recommendations = append(recommendations, 
			"Review Azure Activity Log for detailed error information",
			"Check resource group deployment history",
			"Verify Terraform configuration and Azure provider versions",
		)
	}
	
	return recommendations
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   strings.Contains(strings.ToLower(s), strings.ToLower(substr))))
}

// Cleanup performs cleanup of Azure resources
func (dva *DeploymentValidatorAgent) Cleanup(ctx context.Context) error {
	dva.Logger.Info("Cleaning up deployment validator agent",
		zap.String("agent_id", dva.ID),
		zap.String("resource_group", dva.config.ResourceGroup),
	)
	
	// Force cleanup of the resource group
	cleanupManager := azure.NewCleanupManager(dva.azureClient)
	result := cleanupManager.ForceCleanup(ctx, dva.config.ResourceGroup)
	
	if result.Status != "success" {
		dva.Logger.Error("Failed to cleanup Azure resources",
			zap.String("resource_group", dva.config.ResourceGroup),
			zap.Strings("errors", result.ErrorsEncountered),
		)
		return fmt.Errorf("cleanup failed: %v", result.ErrorsEncountered)
	}
	
	dva.Logger.Info("Azure resources cleaned up successfully",
		zap.String("resource_group", dva.config.ResourceGroup),
		zap.Duration("cleanup_duration", result.Duration),
	)
	
	return nil
}

// GetStatus returns the current status of the deployment validator agent
func (dva *DeploymentValidatorAgent) GetStatus() AgentStatus {
	return dva.Status
}

// GetMetrics returns metrics for the deployment validator agent
func (dva *DeploymentValidatorAgent) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"agent_type":      "deployment-validator",
		"capsule_id":      dva.capsule.ID,
		"resource_group":  dva.config.ResourceGroup,
		"cost_limit_usd":  dva.config.CostLimitUSD,
		"azure_location":  dva.config.Location,
		"ttl_minutes":     dva.config.TTL.Minutes(),
	}
}