package contracts

import (
	"time"
)

// CreateCapsuleRequest represents a request to create a QL capsule
type CreateCapsuleRequest struct {
	IntentID     string                 `json:"intent_id"`
	IntentText   string                 `json:"intent_text"`
	Tasks        []Task                 `json:"tasks"`
	Validations  []ValidationResult     `json:"validations"`
	Artifacts    []ArtifactReference    `json:"artifacts"`
	ProjectFiles map[string]string      `json:"project_files"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CreateQuantumDropRequest represents a request to create quantum drops
type CreateQuantumDropRequest struct {
	IntentID     string            `json:"intent_id"`
	Tasks        []Task            `json:"tasks"`
	ProjectFiles map[string]string `json:"project_files"`
	DropTypes    []DropType        `json:"drop_types,omitempty"` // If empty, create all types
}

// Task represents a task in the packaging context
type Task struct {
	ID           string            `json:"id"`
	Type         TaskType          `json:"type"`
	Description  string            `json:"description"`
	Status       TaskStatus        `json:"status"`
	Files        map[string]string `json:"files"`
	Metadata     map[string]string `json:"metadata"`
	AgentID      string            `json:"agent_id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

type TaskType string

const (
	TaskTypeCodegen TaskType = "codegen"
	TaskTypeInfra   TaskType = "infra"
	TaskTypeDoc     TaskType = "doc"
	TaskTypeTest    TaskType = "test"
	TaskTypeAnalyze TaskType = "analyze"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
)

// ValidationResult represents validation outcome
type ValidationResult struct {
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Score       int                    `json:"score"`
	Issues      []ValidationIssue      `json:"issues"`
	Metadata    map[string]interface{} `json:"metadata"`
	ValidatedAt time.Time              `json:"validated_at"`
}

type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ArtifactReference represents a reference to a generated artifact
type ArtifactReference struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Path        string                 `json:"path"`
	Size        int64                  `json:"size"`
	Checksum    string                 `json:"checksum"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// QLCapsule represents the complete packaged result
type QLCapsule struct {
	Metadata         CapsuleMetadata       `json:"metadata"`
	Tasks            []Task                `json:"tasks"`
	ValidationResults []ValidationResult   `json:"validation_results"`
	ExecutionSummary ExecutionSummary     `json:"execution_summary"`
	SecurityReport   SecurityReport       `json:"security_report"`
	QualityReport    QualityReport        `json:"quality_report"`
	Artifacts        []ArtifactReference  `json:"artifacts"`
	Manifest         CapsuleManifest      `json:"manifest"`
	UnifiedProject   *UnifiedProject      `json:"unified_project,omitempty"`
	ValidationReport *ValidationReport    `json:"validation_report,omitempty"`
}

type CapsuleMetadata struct {
	CapsuleID       string        `json:"capsule_id"`
	Version         string        `json:"version"`
	IntentID        string        `json:"intent_id"`
	IntentText      string        `json:"intent_text"`
	CreatedAt       time.Time     `json:"created_at"`
	CompletedAt     time.Time     `json:"completed_at"`
	Duration        time.Duration `json:"duration"`
	TotalTasks      int           `json:"total_tasks"`
	SuccessfulTasks int           `json:"successful_tasks"`
	FailedTasks     int           `json:"failed_tasks"`
	OverallScore    int           `json:"overall_score"`
	QualityScore    int           `json:"quality_score"`
	TenantID        string        `json:"tenant_id"`
	Environment     string        `json:"environment"`
}

type ExecutionSummary struct {
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	TotalDuration    time.Duration `json:"total_duration"`
	TasksExecuted    int           `json:"tasks_executed"`
	TasksSuccessful  int           `json:"tasks_successful"`
	TasksFailed      int           `json:"tasks_failed"`
	FilesGenerated   int           `json:"files_generated"`
	LinesOfCode      int           `json:"lines_of_code"`
	ValidationsPassed int          `json:"validations_passed"`
	ValidationsFailed int          `json:"validations_failed"`
}

type SecurityReport struct {
	OverallRisk      string              `json:"overall_risk"`
	VulnerabilityCount int               `json:"vulnerability_count"`
	Issues           []SecurityIssue     `json:"issues"`
	ScanResults      []SecurityScanResult `json:"scan_results"`
	Recommendations  []string            `json:"recommendations"`
	ComplianceStatus map[string]string   `json:"compliance_status"`
}

type SecurityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Remediation string `json:"remediation"`
}

type SecurityScanResult struct {
	Scanner     string    `json:"scanner"`
	Status      string    `json:"status"`
	IssuesFound int       `json:"issues_found"`
	ScanTime    time.Time `json:"scan_time"`
}

type QualityReport struct {
	OverallScore     int                   `json:"overall_score"`
	CodeQuality      int                   `json:"code_quality"`
	TestCoverage     float64              `json:"test_coverage"`
	Documentation    int                   `json:"documentation"`
	Maintainability  int                   `json:"maintainability"`
	Performance      int                   `json:"performance"`
	Metrics          map[string]interface{} `json:"metrics"`
	Recommendations  []string              `json:"recommendations"`
}

type CapsuleManifest struct {
	Version      string                 `json:"version"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Files        []ManifestFile         `json:"files"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ManifestFile struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
	Description string `json:"description,omitempty"`
}

type UnifiedProject struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Structure    ProjectStructure  `json:"structure"`
	Files        map[string]string `json:"files"`
	Dependencies []string          `json:"dependencies"`
	BuildConfig  BuildConfig       `json:"build_config"`
}

type ProjectStructure struct {
	Root        string              `json:"root"`
	Directories []string            `json:"directories"`
	FileTypes   map[string][]string `json:"file_types"`
}

type BuildConfig struct {
	Language    string            `json:"language"`
	Version     string            `json:"version"`
	BuildTool   string            `json:"build_tool"`
	Scripts     map[string]string `json:"scripts"`
	Environment map[string]string `json:"environment"`
}

type ValidationReport struct {
	OverallStatus   string                    `json:"overall_status"`
	ValidationSuite []ValidationSuiteResult   `json:"validation_suite"`
	Summary         ValidationSummary         `json:"summary"`
	GeneratedAt     time.Time                 `json:"generated_at"`
}

type ValidationSuiteResult struct {
	Name        string              `json:"name"`
	Status      string              `json:"status"`
	Score       int                 `json:"score"`
	Results     []ValidationResult  `json:"results"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ValidationSummary struct {
	TotalChecks   int `json:"total_checks"`
	PassedChecks  int `json:"passed_checks"`
	FailedChecks  int `json:"failed_checks"`
	WarningChecks int `json:"warning_checks"`
	OverallScore  int `json:"overall_score"`
}

// QuantumDrop represents a categorized output package
type QuantumDrop struct {
	ID          string                 `json:"id"`
	Type        DropType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Files       map[string]string      `json:"files"`
	Structure   map[string][]string    `json:"structure"`
	Metadata    DropMetadata           `json:"metadata"`
	Status      DropStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	Tasks       []string               `json:"tasks"`
}

type DropType string

const (
	DropTypeInfrastructure DropType = "infrastructure"
	DropTypeCodebase       DropType = "codebase"
	DropTypeDocumentation  DropType = "documentation"
	DropTypeTesting        DropType = "testing"
	DropTypeAnalysis       DropType = "analysis"
)

type DropStatus string

const (
	DropStatusPending     DropStatus = "pending"
	DropStatusReady       DropStatus = "ready"
	DropStatusApproved    DropStatus = "approved"
	DropStatusRejected    DropStatus = "rejected"
	DropStatusModified    DropStatus = "modified"
)

type DropMetadata struct {
	FileCount       int                    `json:"file_count"`
	TotalLines      int                    `json:"total_lines"`
	Technologies    []string               `json:"technologies"`
	EstimatedEffort string                 `json:"estimated_effort"`
	Complexity      string                 `json:"complexity"`
	Dependencies    []string               `json:"dependencies"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

// Response types
type CreateCapsuleResponse struct {
	CapsuleID  string     `json:"capsule_id"`
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	Capsule    *QLCapsule `json:"capsule,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
}

type CreateQuantumDropResponse struct {
	DropID      string         `json:"drop_id"`
	Status      string         `json:"status"`
	Message     string         `json:"message"`
	Drops       []QuantumDrop  `json:"drops"`
	DownloadURL string         `json:"download_url,omitempty"`
}

type GetCapsuleResponse struct {
	CapsuleID string     `json:"capsule_id"`
	Capsule   *QLCapsule `json:"capsule"`
	Status    string     `json:"status"`
}

type ListCapsulesResponse struct {
	Capsules []CapsuleMetadata `json:"capsules"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}