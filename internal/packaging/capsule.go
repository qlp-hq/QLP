package packaging

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
)

type CapsulePackager struct {
	outputDir     string
	fileGenerator *FileGenerator
	projectMerger *ProjectMerger
}

type QLCapsule struct {
	Metadata      CapsuleMetadata      `json:"metadata"`
	Tasks         []TaskArtifact       `json:"tasks"`
	ValidationResults []types.ValidationResult `json:"validation_results"`
	ExecutionSummary ExecutionSummary   `json:"execution_summary"`
	SecurityReport SecurityReport      `json:"security_report"`
	QualityReport  QualityReport       `json:"quality_report"`
	Artifacts     []ArtifactReference  `json:"artifacts"`
	Manifest      CapsuleManifest      `json:"manifest"`
	UnifiedProject *UnifiedProject     `json:"unified_project,omitempty"`
	ValidationReport *DeploymentValidationReport `json:"validation_report,omitempty"`
}

type CapsuleMetadata struct {
	CapsuleID       string                 `json:"capsule_id"`
	Version         string                 `json:"version"`
	IntentID        string                 `json:"intent_id"`
	IntentText      string                 `json:"intent_text"`
	CreatedAt       time.Time              `json:"created_at"`
	CompletedAt     time.Time              `json:"completed_at"`
	Duration        time.Duration          `json:"duration"`
	TotalTasks      int                    `json:"total_tasks"`
	SuccessfulTasks int                    `json:"successful_tasks"`
	FailedTasks     int                    `json:"failed_tasks"`
	OverallScore    int                    `json:"overall_score"`
	QualityScore    int                    `json:"quality_score"`
	Tags            []string               `json:"tags"`
	Environment     map[string]interface{} `json:"environment"`
}

type TaskArtifact struct {
	TaskID           string                        `json:"task_id"`
	Type             models.TaskType               `json:"type"`
	Description      string                        `json:"description"`
	Status           models.TaskStatus             `json:"status"`
	Output           string                        `json:"output"`
	AgentID          string                        `json:"agent_id"`
	ExecutionTime    time.Duration                 `json:"execution_time"`
	SandboxResult    *sandbox.SandboxExecutionResult `json:"sandbox_result,omitempty"`
	ValidationResult *types.ValidationResult   `json:"validation_result,omitempty"`
	Dependencies     []string                      `json:"dependencies"`
	Artifacts        []string                      `json:"artifacts"`
}

type ExecutionSummary struct {
	TotalExecutionTime time.Duration            `json:"total_execution_time"`
	TaskBreakdown      map[models.TaskType]int  `json:"task_breakdown"`
	AgentUtilization   map[string]time.Duration `json:"agent_utilization"`
	ResourceUsage      ResourceUsageSummary     `json:"resource_usage"`
	ErrorSummary       []ErrorSummary           `json:"error_summary"`
	PerformanceMetrics PerformanceMetrics       `json:"performance_metrics"`
}

type SecurityReport struct {
	OverallRiskLevel      types.SecurityRiskLevel `json:"overall_risk_level"`
	SecurityScore         int                          `json:"security_score"`
	VulnerabilitiesFound  int                          `json:"vulnerabilities_found"`
	CriticalIssues        []types.SecurityIssue   `json:"critical_issues"`
	ComplianceScore       int                          `json:"compliance_score"`
	SandboxViolations     []string                     `json:"sandbox_violations"`
	RecommendedActions    []string                     `json:"recommended_actions"`
}

type QualityReport struct {
	OverallQualityScore int                    `json:"overall_quality_score"`
	CodeQualityMetrics  CodeQualityMetrics     `json:"code_quality_metrics"`
	TestCoverageData    TestCoverageData       `json:"test_coverage_data"`
	DocumentationScore  int                    `json:"documentation_score"`
	BestPracticesScore  int                    `json:"best_practices_score"`
	Recommendations     []QualityRecommendation `json:"recommendations"`
}

type ResourceUsageSummary struct {
	PeakCPUUsage    float64       `json:"peak_cpu_usage"`
	PeakMemoryUsage int64         `json:"peak_memory_usage"`
	TotalDiskIO     int64         `json:"total_disk_io"`
	NetworkUsage    int64         `json:"network_usage"`
	ExecutionTime   time.Duration `json:"execution_time"`
}

type ErrorSummary struct {
	TaskID      string `json:"task_id"`
	ErrorType   string `json:"error_type"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Recoverable bool   `json:"recoverable"`
}

type PerformanceMetrics struct {
	TasksPerSecond       float64 `json:"tasks_per_second"`
	AverageTaskDuration  float64 `json:"average_task_duration"`
	ConcurrentAgents     int     `json:"concurrent_agents"`
	SandboxOverhead      float64 `json:"sandbox_overhead"`
	ValidationOverhead   float64 `json:"validation_overhead"`
}

type CodeQualityMetrics struct {
	LinesOfCode           int     `json:"lines_of_code"`
	CyclomaticComplexity  int     `json:"cyclomatic_complexity"`
	MaintainabilityIndex  int     `json:"maintainability_index"`
	TechnicalDebt         float64 `json:"technical_debt"`
	CodeDuplication       float64 `json:"code_duplication"`
}

type TestCoverageData struct {
	OverallCoverage   float64            `json:"overall_coverage"`
	LineCoverage      float64            `json:"line_coverage"`
	BranchCoverage    float64            `json:"branch_coverage"`
	FunctionCoverage  float64            `json:"function_coverage"`
	CoverageByFile    map[string]float64 `json:"coverage_by_file"`
	UncoveredLines    []string           `json:"uncovered_lines"`
}

type QualityRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
}

type ArtifactReference struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Path        string            `json:"path"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	MimeType    string            `json:"mime_type"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

type CapsuleManifest struct {
	SchemaVersion string                 `json:"schema_version"`
	CapsuleFormat string                 `json:"capsule_format"`
	Compatibility []string               `json:"compatibility"`
	FileStructure map[string]string      `json:"file_structure"`
	Dependencies  []DependencyInfo       `json:"dependencies"`
	Runtime       RuntimeRequirements    `json:"runtime"`
	Documentation DocumentationManifest `json:"documentation"`
}

type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Source  string `json:"source"`
}

type RuntimeRequirements struct {
	GoVersion      string   `json:"go_version"`
	Platforms      []string `json:"platforms"`
	MinMemory      string   `json:"min_memory"`
	MinCPU         string   `json:"min_cpu"`
	ContainerImage string   `json:"container_image"`
}

type DocumentationManifest struct {
	README       string   `json:"readme"`
	API          string   `json:"api"`
	Examples     []string `json:"examples"`
	Changelog    string   `json:"changelog"`
	Architecture string   `json:"architecture"`
}

// DeploymentValidationReport contains real Azure deployment validation results
type DeploymentValidationReport struct {
	CapsuleID           string    `json:"capsule_id"`
	ResourceGroup       string    `json:"resource_group"`
	DeploymentSuccess   bool      `json:"deployment_success"`
	Status              string    `json:"status"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	CostEstimateUSD     float64   `json:"cost_estimate_usd"`
	HealthChecksPassed  int       `json:"health_checks_passed"`
	TotalHealthChecks   int       `json:"total_health_checks"`
	TestsPassed         int       `json:"tests_passed"`
	TotalTests          int       `json:"total_tests"`
	AzureLocation       string    `json:"azure_location"`
	ValidationDetails   map[string]interface{} `json:"validation_details"`
	ErrorMessage        string    `json:"error_message,omitempty"`
}

func NewCapsulePackager(outputDir string) *CapsulePackager {
	return &CapsulePackager{
		outputDir:     outputDir,
		fileGenerator: NewFileGenerator(),
		projectMerger: NewProjectMerger(),
	}
}

func (cp *CapsulePackager) PackageCapsule(ctx context.Context, intent models.Intent, taskResults []TaskExecutionResult) (*QLCapsule, error) {
	log.Printf("Starting capsule packaging for intent %s", intent.ID)

	capsuleID := generateCapsuleID(intent)
	
	// Create unified project from all tasks
	unifiedProject, err := cp.projectMerger.MergeTasksIntoProject(intent, taskResults)
	if err != nil {
		log.Printf("Warning: Failed to merge tasks into unified project: %v", err)
		unifiedProject = nil
	}
	
	// Build quality report first so we can use its score in metadata
	qualityReport := cp.buildQualityReport(taskResults)
	
	capsule := &QLCapsule{
		Metadata: cp.buildMetadata(intent, taskResults, capsuleID, qualityReport.OverallQualityScore),
		Tasks:    cp.buildTaskArtifacts(taskResults),
		ValidationResults: cp.extractValidationResults(taskResults),
		ExecutionSummary: cp.buildExecutionSummary(taskResults),
		SecurityReport: cp.buildSecurityReport(taskResults),
		QualityReport: qualityReport,
		Artifacts: cp.collectArtifacts(taskResults),
		Manifest: cp.buildManifest(),
		UnifiedProject: unifiedProject,
	}

	return capsule, nil
}

func (cp *CapsulePackager) ExportCapsule(ctx context.Context, capsule *QLCapsule, format string) ([]byte, error) {
	switch format {
	case "qlcapsule", "zip":
		return cp.exportAsZip(capsule)
	case "json":
		return cp.exportAsJSON(capsule)
	case "tar.gz":
		return cp.exportAsTarGz(capsule)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (cp *CapsulePackager) exportAsZip(capsule *QLCapsule) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add manifest.json
	manifestData, err := json.MarshalIndent(capsule.Manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	manifestWriter, err := zipWriter.Create("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest file: %w", err)
	}
	
	if _, err := manifestWriter.Write(manifestData); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Add metadata.json
	metadataData, err := json.MarshalIndent(capsule.Metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	metadataWriter, err := zipWriter.Create("metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata file: %w", err)
	}
	
	if _, err := metadataWriter.Write(metadataData); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// Add tasks/
	for _, task := range capsule.Tasks {
		taskPath := fmt.Sprintf("tasks/%s.json", task.TaskID)
		taskData, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task %s: %w", task.TaskID, err)
		}
		
		taskWriter, err := zipWriter.Create(taskPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create task file %s: %w", taskPath, err)
		}
		
		if _, err := taskWriter.Write(taskData); err != nil {
			return nil, fmt.Errorf("failed to write task %s: %w", task.TaskID, err)
		}

		// Skip individual task file generation - we'll create unified project below
	}

	// Add unified project if available from capsule
	if capsule.UnifiedProject != nil {
		// Add unified project files
		err = cp.addUnifiedProject(zipWriter, capsule.UnifiedProject)
		if err != nil {
			return nil, fmt.Errorf("failed to add unified project: %w", err)
		}
	}

	// Add reports/
	reportsData := map[string]interface{}{
		"execution_summary": capsule.ExecutionSummary,
		"security_report":   capsule.SecurityReport,
		"quality_report":    capsule.QualityReport,
		"validation_results": capsule.ValidationResults,
	}

	for name, data := range reportsData {
		reportData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s: %w", name, err)
		}
		
		reportWriter, err := zipWriter.Create(fmt.Sprintf("reports/%s.json", name))
		if err != nil {
			return nil, fmt.Errorf("failed to create report file %s: %w", name, err)
		}
		
		if _, err := reportWriter.Write(reportData); err != nil {
			return nil, fmt.Errorf("failed to write report %s: %w", name, err)
		}
	}

	// Add README.md
	readme := cp.generateREADME(capsule)
	readmeWriter, err := zipWriter.Create("README.md")
	if err != nil {
		return nil, fmt.Errorf("failed to create README: %w", err)
	}
	
	if _, err := readmeWriter.Write([]byte(readme)); err != nil {
		return nil, fmt.Errorf("failed to write README: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (cp *CapsulePackager) exportAsJSON(capsule *QLCapsule) ([]byte, error) {
	return json.MarshalIndent(capsule, "", "  ")
}

func (cp *CapsulePackager) exportAsTarGz(capsule *QLCapsule) ([]byte, error) {
	// For production implementation, would use tar+gzip
	// For now, return zip format
	return cp.exportAsZip(capsule)
}

func (cp *CapsulePackager) buildMetadata(intent models.Intent, results []TaskExecutionResult, capsuleID string, qualityScore int) CapsuleMetadata {
	successfulTasks := 0
	failedTasks := 0
	totalScore := 0
	validationCount := 0

	for _, result := range results {
		if result.Status == models.TaskStatusCompleted {
			successfulTasks++
		} else if result.Status == models.TaskStatusFailed {
			failedTasks++
		}

		if result.ValidationResult != nil {
			totalScore += result.ValidationResult.OverallScore
			validationCount++
		}
	}

	overallScore := 0
	if validationCount > 0 {
		overallScore = totalScore / validationCount
	}

	return CapsuleMetadata{
		CapsuleID:       capsuleID,
		Version:         "1.0.0",
		IntentID:        intent.ID,
		IntentText:      intent.UserInput,
		CreatedAt:       intent.CreatedAt,
		CompletedAt:     time.Now(),
		Duration:        time.Since(intent.CreatedAt),
		TotalTasks:      len(results),
		SuccessfulTasks: successfulTasks,
		FailedTasks:     failedTasks,
		OverallScore:    overallScore,
		QualityScore:    qualityScore,
		Tags:            cp.generateTags(intent, results),
		Environment:     cp.captureEnvironment(),
	}
}

func (cp *CapsulePackager) buildTaskArtifacts(results []TaskExecutionResult) []TaskArtifact {
	var artifacts []TaskArtifact

	for _, result := range results {
		artifact := TaskArtifact{
			TaskID:           result.Task.ID,
			Type:             result.Task.Type,
			Description:      result.Task.Description,
			Status:           result.Status,
			Output:           result.Output,
			AgentID:          result.AgentID,
			ExecutionTime:    result.ExecutionTime,
			SandboxResult:    result.SandboxResult,
			ValidationResult: result.ValidationResult,
			Dependencies:     result.Task.Dependencies,
			Artifacts:        cp.extractTaskArtifacts(result),
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

func generateCapsuleID(intent models.Intent) string {
	hash := sha256.Sum256([]byte(intent.ID + intent.UserInput + time.Now().String()))
	return fmt.Sprintf("QL-CAP-%x", hash[:8])
}

func (cp *CapsulePackager) generateTags(intent models.Intent, results []TaskExecutionResult) []string {
	tags := []string{"ql-capsule", "production"}
	
	// Add task type tags
	taskTypes := make(map[models.TaskType]bool)
	for _, result := range results {
		taskTypes[result.Task.Type] = true
	}
	
	for taskType := range taskTypes {
		tags = append(tags, string(taskType))
	}

	// Add complexity tag
	if len(results) > 10 {
		tags = append(tags, "complex")
	} else if len(results) > 5 {
		tags = append(tags, "medium")
	} else {
		tags = append(tags, "simple")
	}

	return tags
}

func (cp *CapsulePackager) captureEnvironment() map[string]interface{} {
	return map[string]interface{}{
		"go_version":     "1.21+",
		"platform":      "linux/amd64",
		"quantum_layer": "v1.0.0",
		"packaged_at":   time.Now().UTC(),
		"packager":      "QuantumLayer Capsule Packager v1.0",
	}
}

// addTaskProjectFiles processes the LLM output and adds structured project files to the capsule
func (cp *CapsulePackager) addTaskProjectFiles(zipWriter *zip.Writer, task TaskArtifact) error {
	// Extract LLM output from the combined output
	llmOutput := cp.extractLLMOutput(task.Output)
	
	// Parse the LLM output using the file generator
	projectStruct, err := cp.fileGenerator.ParseLLMOutput(task.TaskID, string(task.Type), llmOutput)
	if err != nil {
		return fmt.Errorf("failed to parse LLM output for task %s: %w", task.TaskID, err)
	}
	
	// Generate file structure
	fileMap := cp.fileGenerator.GenerateFileStructure(projectStruct)
	
	// Create project directory for this task
	projectDir := fmt.Sprintf("projects/%s", task.TaskID)
	
	// Add each file to the zip
	for filePath, content := range fileMap {
		fullPath := fmt.Sprintf("%s/%s", projectDir, filePath)
		
		fileWriter, err := zipWriter.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
		
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
	}
	
	return nil
}

// extractLLMOutput extracts the raw LLM output from the combined agent output
func (cp *CapsulePackager) extractLLMOutput(agentOutput string) string {
	// The agent output format is:
	// === LLM OUTPUT ===
	// <content>
	// === SANDBOX EXECUTION ===
	// ...
	
	lines := strings.Split(agentOutput, "\n")
	var llmOutput []string
	inLLMSection := false
	
	for _, line := range lines {
		if strings.Contains(line, "=== LLM OUTPUT ===") {
			inLLMSection = true
			continue
		}
		if strings.Contains(line, "=== SANDBOX EXECUTION ===") {
			break
		}
		if inLLMSection {
			llmOutput = append(llmOutput, line)
		}
	}
	
	return strings.Join(llmOutput, "\n")
}

// addUnifiedProject adds the merged project structure to the capsule
func (cp *CapsulePackager) addUnifiedProject(zipWriter *zip.Writer, project *UnifiedProject) error {
	log.Printf("Adding unified project '%s' with %d files", project.Name, len(project.Files))
	
	// Create the project directory
	projectDir := fmt.Sprintf("project/%s", project.Name)
	
	// Add each file to the zip
	for filePath, content := range project.Files {
		fullPath := fmt.Sprintf("%s/%s", projectDir, filePath)
		
		fileWriter, err := zipWriter.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
		
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
	}
	
	// Add project metadata
	projectMetadata := map[string]interface{}{
		"name":        project.Name,
		"type":        project.Type,
		"description": project.Description,
		"structure":   project.Structure,
		"file_count":  len(project.Files),
	}
	
	metadataJSON, _ := json.MarshalIndent(projectMetadata, "", "  ")
	metadataPath := fmt.Sprintf("%s/project.json", projectDir)
	
	metadataWriter, err := zipWriter.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create project metadata: %w", err)
	}
	
	if _, err := metadataWriter.Write(metadataJSON); err != nil {
		return fmt.Errorf("failed to write project metadata: %w", err)
	}
	
	log.Printf("Successfully added unified project with %d files", len(project.Files))
	return nil
}

// Helper types for task execution results
type TaskExecutionResult struct {
	Task             models.Task
	Status           models.TaskStatus
	Output           string
	AgentID          string
	ExecutionTime    time.Duration
	SandboxResult    *sandbox.SandboxExecutionResult
	ValidationResult *types.ValidationResult
	Error            error
}