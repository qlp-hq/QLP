package packaging

import (
	"fmt"
	"strings"
	"time"

	"QLP/internal/models"
	"QLP/internal/types"
)

func (cp *CapsulePackager) extractValidationResults(results []TaskExecutionResult) []types.ValidationResult {
	var validationResults []types.ValidationResult
	
	for _, result := range results {
		if result.ValidationResult != nil {
			validationResults = append(validationResults, *result.ValidationResult)
		}
	}
	
	return validationResults
}

func (cp *CapsulePackager) buildExecutionSummary(results []TaskExecutionResult) ExecutionSummary {
	summary := ExecutionSummary{
		TaskBreakdown:    make(map[models.TaskType]int),
		AgentUtilization: make(map[string]time.Duration),
		ErrorSummary:     []ErrorSummary{},
	}

	var totalExecutionTime time.Duration
	var totalCPU, totalMemory, totalDiskIO, totalNetwork int64
	var peakCPU float64
	var peakMemory int64
	
	for _, result := range results {
		// Task breakdown
		summary.TaskBreakdown[result.Task.Type]++
		
		// Agent utilization
		summary.AgentUtilization[result.AgentID] += result.ExecutionTime
		
		// Total execution time
		totalExecutionTime += result.ExecutionTime
		
		// Resource usage from sandbox
		if result.SandboxResult != nil {
			for _, sandboxRes := range result.SandboxResult.Results {
				if sandboxRes.Metrics != nil {
					if sandboxRes.Metrics.CPUUsagePercent > peakCPU {
						peakCPU = sandboxRes.Metrics.CPUUsagePercent
					}
					if sandboxRes.Metrics.MemoryUsageBytes > peakMemory {
						peakMemory = sandboxRes.Metrics.MemoryUsageBytes
					}
					totalCPU += int64(sandboxRes.Metrics.CPUUsagePercent)
					totalMemory += sandboxRes.Metrics.MemoryUsageBytes
					totalDiskIO += sandboxRes.Metrics.DiskUsageBytes
					totalNetwork += sandboxRes.Metrics.NetworkRxBytes + sandboxRes.Metrics.NetworkTxBytes
				}
			}
		}
		
		// Error summary
		if result.Error != nil {
			errorSummary := ErrorSummary{
				TaskID:      result.Task.ID,
				ErrorType:   "execution_error",
				Message:     result.Error.Error(),
				Severity:    determineSeverity(result.Error),
				Recoverable: isRecoverableError(result.Error),
			}
			summary.ErrorSummary = append(summary.ErrorSummary, errorSummary)
		}
	}

	summary.TotalExecutionTime = totalExecutionTime
	summary.ResourceUsage = ResourceUsageSummary{
		PeakCPUUsage:    peakCPU,
		PeakMemoryUsage: peakMemory,
		TotalDiskIO:     totalDiskIO,
		NetworkUsage:    totalNetwork,
		ExecutionTime:   totalExecutionTime,
	}

	// Performance metrics
	totalTasks := len(results)
	avgTaskDuration := float64(totalExecutionTime.Milliseconds()) / float64(totalTasks)
	tasksPerSecond := float64(totalTasks) / totalExecutionTime.Seconds()
	
	summary.PerformanceMetrics = PerformanceMetrics{
		TasksPerSecond:      tasksPerSecond,
		AverageTaskDuration: avgTaskDuration,
		ConcurrentAgents:    len(summary.AgentUtilization),
		SandboxOverhead:     cp.calculateSandboxOverhead(results),
		ValidationOverhead:  cp.calculateValidationOverhead(results),
	}

	return summary
}

func (cp *CapsulePackager) buildSecurityReport(results []TaskExecutionResult) SecurityReport {
	report := SecurityReport{
		OverallRiskLevel:   types.SecurityRiskLevelNone,
		SecurityScore:      100,
		CriticalIssues:     []types.SecurityIssue{},
		SandboxViolations:  []string{},
		RecommendedActions: []string{},
	}

	var totalSecurityScore, validationCount int
	var allVulnerabilities []types.SecurityIssue
	var highestRiskLevel types.SecurityRiskLevel

	for _, result := range results {
		if result.ValidationResult != nil && result.ValidationResult.SecurityResult != nil {
			secResult := result.ValidationResult.SecurityResult
			
			totalSecurityScore += secResult.Score
			validationCount++
			
			// Collect vulnerabilities
			allVulnerabilities = append(allVulnerabilities, secResult.Vulnerabilities...)
			
			// Track highest risk level
			if secResult.RiskLevel > highestRiskLevel {
				highestRiskLevel = secResult.RiskLevel
			}
			
			// Collect sandbox violations
			report.SandboxViolations = append(report.SandboxViolations, secResult.SandboxViolations...)
		}
	}

	// Calculate averages
	if validationCount > 0 {
		report.SecurityScore = totalSecurityScore / validationCount
	}
	
	report.OverallRiskLevel = highestRiskLevel
	report.VulnerabilitiesFound = len(allVulnerabilities)

	// Extract critical issues
	for _, vuln := range allVulnerabilities {
		if vuln.Severity == types.SecurityRiskCriticalStr || vuln.Severity == types.SecurityRiskHighStr {
			report.CriticalIssues = append(report.CriticalIssues, vuln)
		}
	}

	// Generate recommendations
	report.RecommendedActions = cp.generateSecurityRecommendations(allVulnerabilities, report.SandboxViolations)

	return report
}

func (cp *CapsulePackager) buildQualityReport(results []TaskExecutionResult) QualityReport {
	report := QualityReport{
		CodeQualityMetrics: CodeQualityMetrics{},
		TestCoverageData:   TestCoverageData{},
		Recommendations:    []QualityRecommendation{},
	}

	var totalQualityScore, totalDocScore, totalBestPracticesScore int
	var validationCount int
	var totalTestCoverage float64
	var totalLOC, totalComplexity int

	for _, result := range results {
		if result.ValidationResult != nil {
			// Use the QualityScore from types.ValidationResult directly
			totalQualityScore += result.ValidationResult.QualityScore
			// Set reasonable defaults for other quality metrics
			totalDocScore += result.ValidationResult.QualityScore
			totalBestPracticesScore += result.ValidationResult.QualityScore
			totalTestCoverage += float64(result.ValidationResult.QualityScore)
			validationCount++
			
			// Estimate code metrics from output
			if result.Output != "" {
				lines := strings.Split(result.Output, "\n")
				totalLOC += len(lines)
				totalComplexity += cp.estimateComplexity(result.Output)
			}
		}
	}

	// Calculate averages
	if validationCount > 0 {
		report.OverallQualityScore = totalQualityScore / validationCount
		report.DocumentationScore = totalDocScore / validationCount
		report.BestPracticesScore = totalBestPracticesScore / validationCount
		
		report.TestCoverageData.OverallCoverage = totalTestCoverage / float64(validationCount)
		report.TestCoverageData.LineCoverage = report.TestCoverageData.OverallCoverage
		report.TestCoverageData.BranchCoverage = report.TestCoverageData.OverallCoverage * 0.8
		report.TestCoverageData.FunctionCoverage = report.TestCoverageData.OverallCoverage * 0.9
	}

	report.CodeQualityMetrics = CodeQualityMetrics{
		LinesOfCode:          totalLOC,
		CyclomaticComplexity: totalComplexity,
		MaintainabilityIndex: cp.calculateMaintainabilityIndex(totalLOC, totalComplexity),
		TechnicalDebt:        cp.estimateTechnicalDebt(results),
		CodeDuplication:      cp.estimateCodeDuplication(results),
	}

	// Generate quality recommendations
	report.Recommendations = cp.generateQualityRecommendations(report)

	return report
}

func (cp *CapsulePackager) collectArtifacts(results []TaskExecutionResult) []ArtifactReference {
	var artifacts []ArtifactReference
	
	for _, result := range results {
		// Main task output artifact
		if result.Output != "" {
			artifact := ArtifactReference{
				Name:      fmt.Sprintf("%s_output", result.Task.ID),
				Type:      "task_output",
				Path:      fmt.Sprintf("outputs/%s.txt", result.Task.ID),
				Size:      int64(len(result.Output)),
				Checksum:  cp.calculateChecksum([]byte(result.Output)),
				MimeType:  cp.determineMimeType(result.Task.Type),
				CreatedAt: time.Now(),
				Metadata: map[string]string{
					"task_id":   result.Task.ID,
					"task_type": string(result.Task.Type),
					"agent_id":  result.AgentID,
				},
			}
			artifacts = append(artifacts, artifact)
		}

		// Sandbox artifacts
		if result.SandboxResult != nil {
			sandboxArtifact := ArtifactReference{
				Name:      fmt.Sprintf("%s_sandbox", result.Task.ID),
				Type:      "sandbox_result",
				Path:      fmt.Sprintf("sandbox/%s.json", result.Task.ID),
				Size:      int64(len(fmt.Sprintf("%+v", result.SandboxResult))),
				MimeType:  "application/json",
				CreatedAt: time.Now(),
				Metadata: map[string]string{
					"task_id":        result.Task.ID,
					"success":        fmt.Sprintf("%t", result.SandboxResult.Success),
					"security_score": fmt.Sprintf("%d", result.SandboxResult.SecurityScore),
				},
			}
			artifacts = append(artifacts, sandboxArtifact)
		}

		// Validation artifacts
		if result.ValidationResult != nil {
			validationArtifact := ArtifactReference{
				Name:      fmt.Sprintf("%s_validation", result.Task.ID),
				Type:      "validation_result",
				Path:      fmt.Sprintf("validation/%s.json", result.Task.ID),
				Size:      int64(len(fmt.Sprintf("%+v", result.ValidationResult))),
				MimeType:  "application/json",
				CreatedAt: time.Now(),
				Metadata: map[string]string{
					"task_id":       result.Task.ID,
					"overall_score": fmt.Sprintf("%d", result.ValidationResult.OverallScore),
					"passed":        fmt.Sprintf("%t", result.ValidationResult.Passed),
				},
			}
			artifacts = append(artifacts, validationArtifact)
		}
	}

	return artifacts
}

func (cp *CapsulePackager) buildManifest() CapsuleManifest {
	return CapsuleManifest{
		SchemaVersion: "1.0.0",
		CapsuleFormat: "qlcapsule",
		Compatibility: []string{"ql-runtime-v1", "docker", "kubernetes"},
		FileStructure: map[string]string{
			"manifest.json":      "Capsule manifest and metadata",
			"metadata.json":      "Intent and execution metadata",
			"tasks/":            "Individual task artifacts",
			"outputs/":          "Task execution outputs",
			"reports/":          "Validation and analysis reports",
			"sandbox/":          "Sandbox execution results",
			"validation/":       "Validation results per task",
			"README.md":         "Human-readable documentation",
		},
		Dependencies: []DependencyInfo{
			{
				Name:    "go",
				Version: "1.21+",
				Type:    "runtime",
				Source:  "golang.org",
			},
			{
				Name:    "docker",
				Version: "20.10+",
				Type:    "container",
				Source:  "docker.com",
			},
		},
		Runtime: RuntimeRequirements{
			GoVersion:      "1.21+",
			Platforms:      []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"},
			MinMemory:      "512MB",
			MinCPU:         "1 core",
			ContainerImage: "quantumlayer/runtime:latest",
		},
		Documentation: DocumentationManifest{
			README:       "README.md",
			API:          "docs/api.md",
			Examples:     []string{"examples/"},
			Changelog:    "CHANGELOG.md",
			Architecture: "docs/architecture.md",
		},
	}
}

// Helper functions

func determineSeverity(err error) string {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "critical") || strings.Contains(errStr, "fatal"):
		return "critical"
	case strings.Contains(errStr, "error") || strings.Contains(errStr, "fail"):
		return "high"
	case strings.Contains(errStr, "warn"):
		return "medium"
	default:
		return "low"
	}
}

func isRecoverableError(err error) bool {
	errStr := strings.ToLower(err.Error())
	unrecoverableKeywords := []string{"fatal", "panic", "segmentation", "stack overflow"}
	
	for _, keyword := range unrecoverableKeywords {
		if strings.Contains(errStr, keyword) {
			return false
		}
	}
	return true
}

func (cp *CapsulePackager) calculateSandboxOverhead(results []TaskExecutionResult) float64 {
	totalSandboxTime := time.Duration(0)
	totalExecutionTime := time.Duration(0)
	
	for _, result := range results {
		totalExecutionTime += result.ExecutionTime
		if result.SandboxResult != nil {
			totalSandboxTime += result.SandboxResult.ExecutionTime
		}
	}
	
	if totalExecutionTime > 0 {
		return (float64(totalSandboxTime) / float64(totalExecutionTime)) * 100
	}
	return 0
}

func (cp *CapsulePackager) calculateValidationOverhead(results []TaskExecutionResult) float64 {
	totalValidationTime := time.Duration(0)
	totalExecutionTime := time.Duration(0)
	
	for _, result := range results {
		totalExecutionTime += result.ExecutionTime
		if result.ValidationResult != nil {
			totalValidationTime += result.ValidationResult.ValidationTime
		}
	}
	
	if totalExecutionTime > 0 {
		return (float64(totalValidationTime) / float64(totalExecutionTime)) * 100
	}
	return 0
}

func (cp *CapsulePackager) estimateComplexity(output string) int {
	complexity := 0
	
	// Count control flow statements
	controlFlowKeywords := []string{"if", "for", "while", "switch", "case", "else"}
	for _, keyword := range controlFlowKeywords {
		complexity += strings.Count(strings.ToLower(output), keyword)
	}
	
	// Count function definitions
	complexity += strings.Count(output, "func ")
	
	return complexity
}

func (cp *CapsulePackager) calculateMaintainabilityIndex(loc, complexity int) int {
	if loc == 0 {
		return 100
	}
	
	// Simplified maintainability index calculation
	// Real implementation would use more sophisticated metrics
	score := 100 - (complexity*2) - (loc/100)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

func (cp *CapsulePackager) estimateTechnicalDebt(results []TaskExecutionResult) float64 {
	debtScore := 0.0
	
	for _, result := range results {
		if result.ValidationResult != nil && result.ValidationResult.QualityResult != nil {
			qualResult := result.ValidationResult.QualityResult
			// Technical debt increases with lower quality scores
			debtScore += float64(100 - qualResult.Score) * 0.1
		}
	}
	
	return debtScore
}

func (cp *CapsulePackager) estimateCodeDuplication(results []TaskExecutionResult) float64 {
	// Simplified code duplication estimation
	// Real implementation would use more sophisticated analysis
	return 5.0 // Default 5% duplication estimate
}

func (cp *CapsulePackager) generateSecurityRecommendations(vulnerabilities []types.SecurityIssue, violations []string) []string {
	recommendations := []string{}
	
	if len(vulnerabilities) > 0 {
		recommendations = append(recommendations, "Review and address identified security vulnerabilities")
	}
	
	if len(violations) > 0 {
		recommendations = append(recommendations, "Investigate sandbox violations and strengthen security policies")
	}
	
	// Add specific recommendations based on vulnerability types
	vulnTypes := make(map[string]bool)
	for _, vuln := range vulnerabilities {
		vulnTypes[vuln.Type] = true
	}
	
	if vulnTypes["Injection"] {
		recommendations = append(recommendations, "Implement input validation and parameterized queries")
	}
	
	if vulnTypes["Secrets"] {
		recommendations = append(recommendations, "Use secure credential storage and environment variables")
	}
	
	if vulnTypes["Cryptography"] {
		recommendations = append(recommendations, "Upgrade to strong cryptographic algorithms")
	}
	
	return recommendations
}

func (cp *CapsulePackager) generateQualityRecommendations(report QualityReport) []QualityRecommendation {
	recommendations := []QualityRecommendation{}
	
	if report.TestCoverageData.OverallCoverage < 0.8 {
		recommendations = append(recommendations, QualityRecommendation{
			Category:    "Testing",
			Priority:    "high",
			Description: "Increase test coverage to at least 80%",
			Impact:      "Improves code reliability and reduces bugs",
			Effort:      "medium",
		})
	}
	
	if report.DocumentationScore < 70 {
		recommendations = append(recommendations, QualityRecommendation{
			Category:    "Documentation",
			Priority:    "medium",
			Description: "Improve code documentation and comments",
			Impact:      "Enhances maintainability and team collaboration",
			Effort:      "low",
		})
	}
	
	if report.CodeQualityMetrics.CyclomaticComplexity > 10 {
		recommendations = append(recommendations, QualityRecommendation{
			Category:    "Complexity",
			Priority:    "medium",
			Description: "Reduce cyclomatic complexity by refactoring complex functions",
			Impact:      "Improves code readability and maintainability",
			Effort:      "high",
		})
	}
	
	return recommendations
}

func (cp *CapsulePackager) extractTaskArtifacts(result TaskExecutionResult) []string {
	artifacts := []string{}
	
	if result.Output != "" {
		artifacts = append(artifacts, fmt.Sprintf("outputs/%s.txt", result.Task.ID))
	}
	
	if result.SandboxResult != nil {
		artifacts = append(artifacts, fmt.Sprintf("sandbox/%s.json", result.Task.ID))
	}
	
	if result.ValidationResult != nil {
		artifacts = append(artifacts, fmt.Sprintf("validation/%s.json", result.Task.ID))
	}
	
	return artifacts
}