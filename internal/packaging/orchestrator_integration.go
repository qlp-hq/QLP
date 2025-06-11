package packaging

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
)

// Integration with the main orchestrator for seamless capsule generation

type CapsuleOrchestrator struct {
	packager    *CapsulePackager
	outputDir   string
	autoExport  bool
	exportFormat string
}

func NewCapsuleOrchestrator(outputDir string) *CapsuleOrchestrator {
	return &CapsuleOrchestrator{
		packager:     NewCapsulePackager(outputDir),
		outputDir:    outputDir,
		autoExport:   true,
		exportFormat: "qlcapsule",
	}
}

func (co *CapsuleOrchestrator) ProcessIntentExecution(ctx context.Context, intent models.Intent, tasks []models.Task, executionResults map[string]*AgentExecutionResult) (*QLCapsule, error) {
	log.Printf("Processing intent execution for capsule generation: %s", intent.ID)

	// Convert agent execution results to task execution results
	taskResults := co.convertToTaskResults(tasks, executionResults)

	// Package the capsule
	capsule, err := co.packager.PackageCapsule(ctx, intent, taskResults)
	if err != nil {
		return nil, fmt.Errorf("failed to package capsule: %w", err)
	}

	// Validate capsule structure
	if err := co.packager.validateCapsuleStructure(capsule); err != nil {
		return nil, fmt.Errorf("capsule validation failed: %w", err)
	}

	// Auto-export if enabled
	if co.autoExport {
		if err := co.exportCapsuleToFile(ctx, capsule); err != nil {
			log.Printf("Warning: Failed to auto-export capsule: %v", err)
		}
	}

	log.Printf("Capsule generated successfully: %s", capsule.Metadata.CapsuleID)
	return capsule, nil
}

func (co *CapsuleOrchestrator) convertToTaskResults(tasks []models.Task, executionResults map[string]*AgentExecutionResult) []TaskExecutionResult {
	var taskResults []TaskExecutionResult

	for _, task := range tasks {
		result := TaskExecutionResult{
			Task:   task,
			Status: models.TaskStatusFailed, // Default to failed
		}

		// Find corresponding agent execution result
		if agentResult, exists := executionResults[task.ID]; exists {
			result.Status = co.mapAgentStatusToTaskStatus(agentResult.Status)
			result.Output = agentResult.Output
			result.AgentID = agentResult.AgentID
			result.ExecutionTime = agentResult.ExecutionTime
			result.SandboxResult = agentResult.SandboxResult
			result.ValidationResult = agentResult.ValidationResult
			result.Error = agentResult.Error
		}

		taskResults = append(taskResults, result)
	}

	return taskResults
}

func (co *CapsuleOrchestrator) mapAgentStatusToTaskStatus(agentStatus string) models.TaskStatus {
	switch agentStatus {
	case "completed":
		return models.TaskStatusCompleted
	case "failed":
		return models.TaskStatusFailed
	case "executing":
		return models.TaskStatusInProgress
	case "ready":
		return models.TaskStatusPending
	default:
		return models.TaskStatusFailed
	}
}

func (co *CapsuleOrchestrator) exportCapsuleToFile(ctx context.Context, capsule *QLCapsule) error {
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ql_capsule_%s_%s.%s", 
		capsule.Metadata.CapsuleID, 
		timestamp, 
		co.exportFormat)

	// Create full path
	fullPath := filepath.Join(co.outputDir, filename)

	// Export capsule
	data, err := co.packager.ExportCapsule(ctx, capsule, co.exportFormat)
	if err != nil {
		return fmt.Errorf("failed to export capsule: %w", err)
	}

	// Write real file to disk
	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write capsule file: %w", err)
	}
	
	log.Printf("Capsule exported to: %s (%d bytes)", fullPath, len(data))
	
	return nil
}

// Agent execution result structure for integration
type AgentExecutionResult struct {
	AgentID          string
	Status           string
	Output           string
	ExecutionTime    time.Duration
	SandboxResult    *sandbox.SandboxExecutionResult
	ValidationResult *types.ValidationResult
	Error            error
	StartTime        time.Time
	EndTime          time.Time
}

// Capsule query and management functions

func (co *CapsuleOrchestrator) QueryCapsuleMetrics(capsule *QLCapsule) CapsuleMetrics {
	metrics := CapsuleMetrics{
		CapsuleID:        capsule.Metadata.CapsuleID,
		TaskCount:        len(capsule.Tasks),
		SuccessRate:      float64(capsule.Metadata.SuccessfulTasks) / float64(capsule.Metadata.TotalTasks),
		OverallScore:     capsule.Metadata.OverallScore,
		SecurityRisk:     capsule.SecurityReport.OverallRiskLevel,
		QualityScore:     capsule.QualityReport.OverallQualityScore,
		ExecutionTime:    capsule.Metadata.Duration,
		ArtifactCount:    len(capsule.Artifacts),
		ValidationsPassed: co.countPassedValidations(capsule.ValidationResults),
	}

	return metrics
}

func (co *CapsuleOrchestrator) countPassedValidations(results []types.ValidationResult) int {
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	return passed
}

type CapsuleMetrics struct {
	CapsuleID         string                        `json:"capsule_id"`
	TaskCount         int                           `json:"task_count"`
	SuccessRate       float64                       `json:"success_rate"`
	OverallScore      int                           `json:"overall_score"`
	SecurityRisk      types.SecurityRiskLevel  `json:"security_risk"`
	QualityScore      int                           `json:"quality_score"`
	ExecutionTime     time.Duration                 `json:"execution_time"`
	ArtifactCount     int                           `json:"artifact_count"`
	ValidationsPassed int                           `json:"validations_passed"`
}

// Advanced capsule operations

func (co *CapsuleOrchestrator) CompareCapsules(capsule1, capsule2 *QLCapsule) CapsuleComparison {
	comparison := CapsuleComparison{
		Capsule1ID: capsule1.Metadata.CapsuleID,
		Capsule2ID: capsule2.Metadata.CapsuleID,
	}

	// Compare scores
	comparison.ScoreDifference = capsule2.Metadata.OverallScore - capsule1.Metadata.OverallScore
	comparison.QualityDifference = capsule2.QualityReport.OverallQualityScore - capsule1.QualityReport.OverallQualityScore
	comparison.SecurityDifference = capsule2.SecurityReport.SecurityScore - capsule1.SecurityReport.SecurityScore

	// Compare execution times
	comparison.ExecutionTimeDifference = capsule2.Metadata.Duration - capsule1.Metadata.Duration

	// Compare task counts
	comparison.TaskCountDifference = capsule2.Metadata.TotalTasks - capsule1.Metadata.TotalTasks

	// Generate insights
	comparison.Insights = co.generateComparisonInsights(comparison)

	return comparison
}

type CapsuleComparison struct {
	Capsule1ID              string        `json:"capsule1_id"`
	Capsule2ID              string        `json:"capsule2_id"`
	ScoreDifference         int           `json:"score_difference"`
	QualityDifference       int           `json:"quality_difference"`
	SecurityDifference      int           `json:"security_difference"`
	ExecutionTimeDifference time.Duration `json:"execution_time_difference"`
	TaskCountDifference     int           `json:"task_count_difference"`
	Insights                []string      `json:"insights"`
}

func (co *CapsuleOrchestrator) generateComparisonInsights(comparison CapsuleComparison) []string {
	var insights []string

	if comparison.ScoreDifference > 10 {
		insights = append(insights, "Significant improvement in overall score")
	} else if comparison.ScoreDifference < -10 {
		insights = append(insights, "Notable decline in overall score")
	}

	if comparison.QualityDifference > 15 {
		insights = append(insights, "Major quality improvements detected")
	} else if comparison.QualityDifference < -15 {
		insights = append(insights, "Quality regression identified")
	}

	if comparison.SecurityDifference > 10 {
		insights = append(insights, "Enhanced security posture")
	} else if comparison.SecurityDifference < -10 {
		insights = append(insights, "Security concerns increased")
	}

	if comparison.ExecutionTimeDifference > time.Minute {
		insights = append(insights, "Longer execution time - may indicate complexity increase")
	} else if comparison.ExecutionTimeDifference < -time.Minute {
		insights = append(insights, "Improved execution efficiency")
	}

	if len(insights) == 0 {
		insights = append(insights, "Similar performance characteristics")
	}

	return insights
}

// Capsule health assessment

func (co *CapsuleOrchestrator) AssessCapsuleHealth(capsule *QLCapsule) CapsuleHealthReport {
	health := CapsuleHealthReport{
		CapsuleID: capsule.Metadata.CapsuleID,
		Issues:    []HealthIssue{},
		Score:     100,
	}

	// Check success rate
	successRate := float64(capsule.Metadata.SuccessfulTasks) / float64(capsule.Metadata.TotalTasks)
	if successRate < 0.8 {
		health.Issues = append(health.Issues, HealthIssue{
			Type:        "execution",
			Severity:    "high",
			Description: fmt.Sprintf("Low task success rate: %.1f%%", successRate*100),
			Impact:      "Reduced reliability and potential workflow failures",
		})
		health.Score -= 20
	}

	// Check security
	if capsule.SecurityReport.OverallRiskLevel == types.SecurityRiskLevelHigh || 
	   capsule.SecurityReport.OverallRiskLevel == types.SecurityRiskLevelCritical {
		health.Issues = append(health.Issues, HealthIssue{
			Type:        "security",
			Severity:    "critical",
			Description: fmt.Sprintf("High security risk level: %s", capsule.SecurityReport.OverallRiskLevel),
			Impact:      "Potential security vulnerabilities and compliance issues",
		})
		health.Score -= 30
	}

	// Check quality
	if capsule.QualityReport.OverallQualityScore < 60 {
		health.Issues = append(health.Issues, HealthIssue{
			Type:        "quality",
			Severity:    "medium",
			Description: fmt.Sprintf("Low quality score: %d", capsule.QualityReport.OverallQualityScore),
			Impact:      "Reduced maintainability and potential technical debt",
		})
		health.Score -= 15
	}

	// Determine overall health status
	switch {
	case health.Score >= 90:
		health.Status = "excellent"
	case health.Score >= 70:
		health.Status = "good"
	case health.Score >= 50:
		health.Status = "fair"
	default:
		health.Status = "poor"
	}

	return health
}

type CapsuleHealthReport struct {
	CapsuleID string        `json:"capsule_id"`
	Status    string        `json:"status"`
	Score     int           `json:"score"`
	Issues    []HealthIssue `json:"issues"`
}

type HealthIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// Configuration options

func (co *CapsuleOrchestrator) SetAutoExport(enabled bool) {
	co.autoExport = enabled
}

func (co *CapsuleOrchestrator) SetExportFormat(format string) {
	co.exportFormat = format
}

func (co *CapsuleOrchestrator) SetOutputDirectory(dir string) {
	co.outputDir = dir
	co.packager.outputDir = dir
}