package engines

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/packaging-service/pkg/contracts"
)

// CapsuleEngine handles QL capsule creation and management
type CapsuleEngine struct {
	outputDir string
}

// NewCapsuleEngine creates a new capsule engine
func NewCapsuleEngine(outputDir string) *CapsuleEngine {
	return &CapsuleEngine{
		outputDir: outputDir,
	}
}

// CreateCapsule creates a new QL capsule from the provided data
func (ce *CapsuleEngine) CreateCapsule(ctx context.Context, tenantID string, req *contracts.CreateCapsuleRequest) (*contracts.QLCapsule, error) {
	logger.WithComponent("capsule-engine").Info("Creating QL capsule",
		zap.String("tenant_id", tenantID),
		zap.String("intent_id", req.IntentID))

	// Generate capsule ID and metadata
	capsuleID := fmt.Sprintf("QL-CAP-%s", generateShortID())
	startTime := time.Now()

	// Build capsule metadata
	metadata := ce.buildCapsuleMetadata(capsuleID, req, tenantID, startTime)

	// Process tasks and calculate summaries
	executionSummary := ce.buildExecutionSummary(req.Tasks, startTime)

	// Generate security report
	securityReport := ce.generateSecurityReport(req.Tasks, req.ProjectFiles)

	// Generate quality report
	qualityReport := ce.generateQualityReport(req.Tasks, req.ProjectFiles, req.Validations)

	// Build capsule manifest
	manifest := ce.buildCapsuleManifest(capsuleID, req)

	// Create unified project if we have project files
	var unifiedProject *contracts.UnifiedProject
	if len(req.ProjectFiles) > 0 {
		unifiedProject = ce.buildUnifiedProject(req)
	}

	// Build validation report
	validationReport := ce.buildValidationReport(req.Validations)

	// Create the complete capsule
	capsule := &contracts.QLCapsule{
		Metadata:          metadata,
		Tasks:             req.Tasks,
		ValidationResults: req.Validations,
		ExecutionSummary:  executionSummary,
		SecurityReport:    securityReport,
		QualityReport:     qualityReport,
		Artifacts:         req.Artifacts,
		Manifest:          manifest,
		UnifiedProject:    unifiedProject,
		ValidationReport:  validationReport,
	}

	logger.WithComponent("capsule-engine").Info("QL capsule created successfully",
		zap.String("capsule_id", capsuleID),
		zap.Int("task_count", len(req.Tasks)),
		zap.Int("file_count", len(req.ProjectFiles)))

	return capsule, nil
}

// PackageCapsule packages a capsule into a downloadable ZIP file
func (ce *CapsuleEngine) PackageCapsule(ctx context.Context, capsule *contracts.QLCapsule) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add capsule metadata
	if err := ce.addJSONToZip(zipWriter, "capsule.json", capsule); err != nil {
		return nil, fmt.Errorf("failed to add capsule metadata: %w", err)
	}

	// Add unified project files if available
	if capsule.UnifiedProject != nil {
		for filePath, content := range capsule.UnifiedProject.Files {
			if err := ce.addFileToZip(zipWriter, filepath.Join("project", filePath), content); err != nil {
				return nil, fmt.Errorf("failed to add project file %s: %w", filePath, err)
			}
		}
	}

	// Add task-specific files
	for _, task := range capsule.Tasks {
		taskDir := fmt.Sprintf("tasks/%s", task.ID)
		for fileName, content := range task.Files {
			if err := ce.addFileToZip(zipWriter, filepath.Join(taskDir, fileName), content); err != nil {
				return nil, fmt.Errorf("failed to add task file %s: %w", fileName, err)
			}
		}
	}

	// Add reports
	if err := ce.addJSONToZip(zipWriter, "reports/security.json", capsule.SecurityReport); err != nil {
		return nil, fmt.Errorf("failed to add security report: %w", err)
	}

	if err := ce.addJSONToZip(zipWriter, "reports/quality.json", capsule.QualityReport); err != nil {
		return nil, fmt.Errorf("failed to add quality report: %w", err)
	}

	if err := ce.addJSONToZip(zipWriter, "reports/validation.json", capsule.ValidationReport); err != nil {
		return nil, fmt.Errorf("failed to add validation report: %w", err)
	}

	// Add manifest
	if err := ce.addJSONToZip(zipWriter, "manifest.json", capsule.Manifest); err != nil {
		return nil, fmt.Errorf("failed to add manifest: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper methods

func (ce *CapsuleEngine) buildCapsuleMetadata(capsuleID string, req *contracts.CreateCapsuleRequest, tenantID string, startTime time.Time) contracts.CapsuleMetadata {
	successfulTasks := 0
	failedTasks := 0
	overallScore := 0

	for _, task := range req.Tasks {
		switch task.Status {
		case contracts.TaskStatusCompleted:
			successfulTasks++
		case contracts.TaskStatusFailed:
			failedTasks++
		}
	}

	// Calculate overall score based on task success rate and validation results
	if len(req.Tasks) > 0 {
		successRate := float64(successfulTasks) / float64(len(req.Tasks))
		overallScore = int(successRate * 100)
	}

	// Factor in validation scores
	if len(req.Validations) > 0 {
		validationScore := 0
		for _, validation := range req.Validations {
			validationScore += validation.Score
		}
		avgValidationScore := validationScore / len(req.Validations)
		overallScore = (overallScore + avgValidationScore) / 2
	}

	return contracts.CapsuleMetadata{
		CapsuleID:       capsuleID,
		Version:         "1.0.0",
		IntentID:        req.IntentID,
		IntentText:      req.IntentText,
		CreatedAt:       startTime,
		CompletedAt:     time.Now(),
		Duration:        time.Since(startTime),
		TotalTasks:      len(req.Tasks),
		SuccessfulTasks: successfulTasks,
		FailedTasks:     failedTasks,
		OverallScore:    overallScore,
		QualityScore:    overallScore, // Will be refined by quality analysis
		TenantID:        tenantID,
		Environment:     "production",
	}
}

func (ce *CapsuleEngine) buildExecutionSummary(tasks []contracts.Task, startTime time.Time) contracts.ExecutionSummary {
	tasksSuccessful := 0
	tasksFailed := 0
	filesGenerated := 0
	linesOfCode := 0

	var earliestStart, latestEnd time.Time
	earliestStart = time.Now()

	for _, task := range tasks {
		if task.CreatedAt.Before(earliestStart) {
			earliestStart = task.CreatedAt
		}

		if task.CompletedAt != nil && task.CompletedAt.After(latestEnd) {
			latestEnd = *task.CompletedAt
		}

		switch task.Status {
		case contracts.TaskStatusCompleted:
			tasksSuccessful++
		case contracts.TaskStatusFailed:
			tasksFailed++
		}

		filesGenerated += len(task.Files)

		// Estimate lines of code
		for _, content := range task.Files {
			linesOfCode += len(strings.Split(content, "\n"))
		}
	}

	if latestEnd.IsZero() {
		latestEnd = time.Now()
	}

	return contracts.ExecutionSummary{
		StartTime:        earliestStart,
		EndTime:          latestEnd,
		TotalDuration:    latestEnd.Sub(earliestStart),
		TasksExecuted:    len(tasks),
		TasksSuccessful:  tasksSuccessful,
		TasksFailed:      tasksFailed,
		FilesGenerated:   filesGenerated,
		LinesOfCode:      linesOfCode,
		ValidationsPassed: 0, // Will be calculated from validations
		ValidationsFailed: 0,
	}
}

func (ce *CapsuleEngine) generateSecurityReport(tasks []contracts.Task, projectFiles map[string]string) contracts.SecurityReport {
	issues := []contracts.SecurityIssue{}
	scanResults := []contracts.SecurityScanResult{}
	
	// Basic security analysis
	vulnerabilityCount := 0
	overallRisk := "low"

	// Scan for common security issues
	for _, task := range tasks {
		for fileName, content := range task.Files {
			if ce.containsSecurityIssues(content) {
				issues = append(issues, contracts.SecurityIssue{
					Type:        "potential_vulnerability",
					Severity:    "medium",
					Description: "Potential security vulnerability detected",
					File:        fileName,
					Remediation: "Review code for security best practices",
				})
				vulnerabilityCount++
			}
		}
	}

	if vulnerabilityCount > 0 {
		overallRisk = "medium"
	}
	if vulnerabilityCount > 5 {
		overallRisk = "high"
	}

	scanResults = append(scanResults, contracts.SecurityScanResult{
		Scanner:     "basic_scanner",
		Status:      "completed",
		IssuesFound: vulnerabilityCount,
		ScanTime:    time.Now(),
	})

	return contracts.SecurityReport{
		OverallRisk:        overallRisk,
		VulnerabilityCount: vulnerabilityCount,
		Issues:             issues,
		ScanResults:        scanResults,
		Recommendations:    []string{"Review security best practices", "Implement input validation"},
		ComplianceStatus:   map[string]string{"OWASP": "partial"},
	}
}

func (ce *CapsuleEngine) generateQualityReport(tasks []contracts.Task, projectFiles map[string]string, validations []contracts.ValidationResult) contracts.QualityReport {
	totalScore := 0
	validationCount := 0

	for _, validation := range validations {
		totalScore += validation.Score
		validationCount++
	}

	overallScore := 80 // Default score
	if validationCount > 0 {
		overallScore = totalScore / validationCount
	}

	return contracts.QualityReport{
		OverallScore:     overallScore,
		CodeQuality:      overallScore,
		TestCoverage:     0.85, // Placeholder
		Documentation:    overallScore,
		Maintainability:  overallScore,
		Performance:      overallScore,
		Metrics:          map[string]interface{}{"complexity": "low", "debt": "minimal"},
		Recommendations:  []string{"Add more unit tests", "Improve documentation"},
	}
}

func (ce *CapsuleEngine) buildCapsuleManifest(capsuleID string, req *contracts.CreateCapsuleRequest) contracts.CapsuleManifest {
	files := []contracts.ManifestFile{}

	// Add project files
	for filePath := range req.ProjectFiles {
		files = append(files, contracts.ManifestFile{
			Path:        filePath,
			Type:        ce.getFileType(filePath),
			Size:        int64(len(req.ProjectFiles[filePath])),
			Checksum:    ce.calculateChecksum(req.ProjectFiles[filePath]),
			Description: "Generated project file",
		})
	}

	return contracts.CapsuleManifest{
		Version:      "1.0.0",
		Name:         capsuleID,
		Description:  fmt.Sprintf("QL Capsule for intent: %s", req.IntentText),
		Files:        files,
		Dependencies: []string{},
		Metadata:     map[string]interface{}{"intent_id": req.IntentID},
	}
}

func (ce *CapsuleEngine) buildUnifiedProject(req *contracts.CreateCapsuleRequest) *contracts.UnifiedProject {
	// Analyze project structure
	directories := []string{}
	fileTypes := make(map[string][]string)

	for filePath := range req.ProjectFiles {
		dir := filepath.Dir(filePath)
		if dir != "." {
			directories = append(directories, dir)
		}

		ext := filepath.Ext(filePath)
		if ext != "" {
			fileTypes[ext] = append(fileTypes[ext], filePath)
		}
	}

	return &contracts.UnifiedProject{
		Name:        "generated-project",
		Description: "Auto-generated project from QL capsule",
		Structure: contracts.ProjectStructure{
			Root:        "/",
			Directories: directories,
			FileTypes:   fileTypes,
		},
		Files:        req.ProjectFiles,
		Dependencies: []string{},
		BuildConfig: contracts.BuildConfig{
			Language:    ce.detectLanguage(req.ProjectFiles),
			Version:     "latest",
			BuildTool:   ce.detectBuildTool(req.ProjectFiles),
			Scripts:     map[string]string{"build": "echo 'Build script'"},
			Environment: map[string]string{},
		},
	}
}

func (ce *CapsuleEngine) buildValidationReport(validations []contracts.ValidationResult) *contracts.ValidationReport {
	passed := 0
	failed := 0
	warnings := 0
	totalScore := 0

	for _, validation := range validations {
		switch validation.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "warning":
			warnings++
		}
		totalScore += validation.Score
	}

	overallScore := 0
	if len(validations) > 0 {
		overallScore = totalScore / len(validations)
	}

	overallStatus := "passed"
	if failed > 0 {
		overallStatus = "failed"
	} else if warnings > 0 {
		overallStatus = "warning"
	}

	return &contracts.ValidationReport{
		OverallStatus: overallStatus,
		ValidationSuite: []contracts.ValidationSuiteResult{
			{
				Name:     "comprehensive_validation",
				Status:   overallStatus,
				Score:    overallScore,
				Results:  validations,
				Metadata: map[string]interface{}{"suite_version": "1.0"},
			},
		},
		Summary: contracts.ValidationSummary{
			TotalChecks:   len(validations),
			PassedChecks:  passed,
			FailedChecks:  failed,
			WarningChecks: warnings,
			OverallScore:  overallScore,
		},
		GeneratedAt: time.Now(),
	}
}

// Utility methods

func (ce *CapsuleEngine) addJSONToZip(zipWriter *zip.Writer, path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ce.addFileToZip(zipWriter, path, string(jsonData))
}

func (ce *CapsuleEngine) addFileToZip(zipWriter *zip.Writer, path, content string) error {
	writer, err := zipWriter.Create(path)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(content))
	return err
}

func (ce *CapsuleEngine) containsSecurityIssues(content string) bool {
	// Basic security pattern detection
	securityPatterns := []string{
		"eval(",
		"exec(",
		"system(",
		"shell_exec",
		"password",
		"secret",
		"api_key",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range securityPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

func (ce *CapsuleEngine) getFileType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "source"
	case ".js", ".ts":
		return "source"
	case ".py":
		return "source"
	case ".java":
		return "source"
	case ".md":
		return "documentation"
	case ".json", ".yaml", ".yml":
		return "config"
	default:
		return "file"
	}
}

func (ce *CapsuleEngine) calculateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (ce *CapsuleEngine) detectLanguage(files map[string]string) string {
	for filePath := range files {
		ext := filepath.Ext(filePath)
		switch ext {
		case ".go":
			return "go"
		case ".js", ".ts":
			return "javascript"
		case ".py":
			return "python"
		case ".java":
			return "java"
		case ".rb":
			return "ruby"
		}
	}
	return "unknown"
}

func (ce *CapsuleEngine) detectBuildTool(files map[string]string) string {
	for filePath := range files {
		switch filepath.Base(filePath) {
		case "go.mod":
			return "go"
		case "package.json":
			return "npm"
		case "requirements.txt", "setup.py":
			return "pip"
		case "pom.xml":
			return "maven"
		case "build.gradle":
			return "gradle"
		case "Gemfile":
			return "bundler"
		}
	}
	return "unknown"
}

func generateShortID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
}