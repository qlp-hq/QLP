package contracts

import (
	"time"
)

// ValidationRequest represents a request to validate code/output
type ValidationRequest struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	TaskID      string            `json:"task_id,omitempty"`
	TaskType    TaskType          `json:"task_type"`
	Content     string            `json:"content"`
	Language    string            `json:"language,omitempty"`
	Context     *ValidationContext `json:"context,omitempty"`
	Options     *ValidationOptions `json:"options,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type TaskType string

const (
	TaskTypeCodegen   TaskType = "codegen"
	TaskTypeInfra     TaskType = "infra"
	TaskTypeDoc       TaskType = "doc"
	TaskTypeTest      TaskType = "test"
	TaskTypeAnalyze   TaskType = "analyze"
	TaskTypeValidate  TaskType = "validate"
	TaskTypePackage   TaskType = "package"
)

type ValidationContext struct {
	ProjectType     string            `json:"project_type,omitempty"`
	TechStack       []string          `json:"tech_stack,omitempty"`
	Requirements    []string          `json:"requirements,omitempty"`
	Constraints     map[string]string `json:"constraints,omitempty"`
	PreviousResults []ValidationSummary `json:"previous_results,omitempty"`
}

type ValidationOptions struct {
	Level            ValidationLevel   `json:"level"`
	EnabledChecks    []ValidationCheck `json:"enabled_checks,omitempty"`
	DisabledChecks   []ValidationCheck `json:"disabled_checks,omitempty"`
	CustomRules      []CustomRule      `json:"custom_rules,omitempty"`
	FailFast         bool              `json:"fail_fast,omitempty"`
	Timeout          int               `json:"timeout_seconds,omitempty"`
}

type ValidationLevel string

const (
	ValidationLevelFast         ValidationLevel = "fast"
	ValidationLevelStandard     ValidationLevel = "standard"
	ValidationLevelComprehensive ValidationLevel = "comprehensive"
	ValidationLevelCustom       ValidationLevel = "custom"
)

type ValidationCheck string

const (
	ValidationCheckSyntax        ValidationCheck = "syntax"
	ValidationCheckSecurity      ValidationCheck = "security"
	ValidationCheckQuality       ValidationCheck = "quality"
	ValidationCheckPerformance   ValidationCheck = "performance"
	ValidationCheckCompliance    ValidationCheck = "compliance"
	ValidationCheckAccessibility ValidationCheck = "accessibility"
	ValidationCheckLLMCritique   ValidationCheck = "llm_critique"
)

type CustomRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Pattern     string            `json:"pattern,omitempty"`
	Script      string            `json:"script,omitempty"`
	Severity    Severity          `json:"severity"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ValidationResult represents the complete validation outcome
type ValidationResult struct {
	ID               string              `json:"id"`
	RequestID        string              `json:"request_id"`
	TenantID         string              `json:"tenant_id"`
	TaskID           string              `json:"task_id,omitempty"`
	Status           ValidationStatus    `json:"status"`
	OverallScore     int                 `json:"overall_score"`
	SecurityScore    int                 `json:"security_score"`
	QualityScore     int                 `json:"quality_score"`
	PerformanceScore int                 `json:"performance_score,omitempty"`
	ComplianceScore  int                 `json:"compliance_score,omitempty"`
	Passed           bool                `json:"passed"`
	Level            ValidationLevel     `json:"level"`
	
	// Detailed results
	SyntaxResult      *SyntaxResult      `json:"syntax_result,omitempty"`
	SecurityResult    *SecurityResult    `json:"security_result,omitempty"`
	QualityResult     *QualityResult     `json:"quality_result,omitempty"`
	PerformanceResult *PerformanceResult `json:"performance_result,omitempty"`
	ComplianceResult  *ComplianceResult  `json:"compliance_result,omitempty"`
	LLMCritiqueResult *LLMCritiqueResult `json:"llm_critique_result,omitempty"`
	
	// Execution metadata
	ValidationTime   time.Duration       `json:"validation_time"`
	ExecutedChecks   []ValidationCheck   `json:"executed_checks"`
	SkippedChecks    []ValidationCheck   `json:"skipped_checks"`
	ErrorMessage     string              `json:"error_message,omitempty"`
	ValidatedAt      time.Time           `json:"validated_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
}

type ValidationStatus string

const (
	ValidationStatusPending    ValidationStatus = "pending"
	ValidationStatusRunning    ValidationStatus = "running"
	ValidationStatusCompleted  ValidationStatus = "completed"
	ValidationStatusFailed     ValidationStatus = "failed"
	ValidationStatusTimeout    ValidationStatus = "timeout"
)

// Individual validation result types
type SyntaxResult struct {
	Score       int           `json:"score"`
	Valid       bool          `json:"valid"`
	Issues      []SyntaxIssue `json:"issues,omitempty"`
	Warnings    []SyntaxIssue `json:"warnings,omitempty"`
	LintResults []LintResult  `json:"lint_results,omitempty"`
	Language    string        `json:"language"`
}

type SyntaxIssue struct {
	Type        string   `json:"type"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Line        int      `json:"line,omitempty"`
	Column      int      `json:"column,omitempty"`
	Rule        string   `json:"rule,omitempty"`
	Suggestion  string   `json:"suggestion,omitempty"`
}

type LintResult struct {
	Tool     string        `json:"tool"`
	Version  string        `json:"version"`
	Issues   []SyntaxIssue `json:"issues"`
	Passed   bool          `json:"passed"`
	ExitCode int           `json:"exit_code"`
}

type SecurityResult struct {
	Score           int               `json:"score"`
	RiskLevel       SecurityRiskLevel `json:"risk_level"`
	Vulnerabilities []SecurityIssue   `json:"vulnerabilities,omitempty"`
	Warnings        []SecurityIssue   `json:"warnings,omitempty"`
	Passed          bool              `json:"passed"`
	ScannedBy       []string          `json:"scanned_by,omitempty"`
}

type SecurityRiskLevel string

const (
	SecurityRiskLevelNone     SecurityRiskLevel = "none"
	SecurityRiskLevelLow      SecurityRiskLevel = "low"
	SecurityRiskLevelMedium   SecurityRiskLevel = "medium"
	SecurityRiskLevelHigh     SecurityRiskLevel = "high"
	SecurityRiskLevelCritical SecurityRiskLevel = "critical"
)

type SecurityIssue struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Severity    Severity          `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Line        int               `json:"line,omitempty"`
	Column      int               `json:"column,omitempty"`
	CWE         string            `json:"cwe,omitempty"`
	CVSS        float64           `json:"cvss,omitempty"`
	References  []string          `json:"references,omitempty"`
	Remediation string            `json:"remediation,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type QualityResult struct {
	Score           int               `json:"score"`
	Maintainability int               `json:"maintainability"`
	Documentation   int               `json:"documentation"`
	BestPractices   int               `json:"best_practices"`
	TestCoverage    float64           `json:"test_coverage"`
	Complexity      *ComplexityMetrics `json:"complexity,omitempty"`
	Issues          []QualityIssue    `json:"issues,omitempty"`
	Suggestions     []QualitySuggestion `json:"suggestions,omitempty"`
	Passed          bool              `json:"passed"`
}

type ComplexityMetrics struct {
	Cyclomatic    int     `json:"cyclomatic"`
	Cognitive     int     `json:"cognitive"`
	Halstead      float64 `json:"halstead,omitempty"`
	LinesOfCode   int     `json:"lines_of_code"`
	Maintainability float64 `json:"maintainability_index,omitempty"`
}

type QualityIssue struct {
	Type        string   `json:"type"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Line        int      `json:"line,omitempty"`
	Column      int      `json:"column,omitempty"`
	Rule        string   `json:"rule,omitempty"`
	Category    string   `json:"category,omitempty"`
}

type QualitySuggestion struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Example     string `json:"example,omitempty"`
}

type PerformanceResult struct {
	Score           int                   `json:"score"`
	Issues          []PerformanceIssue    `json:"issues,omitempty"`
	Optimizations   []PerformanceHint     `json:"optimizations,omitempty"`
	Benchmarks      []BenchmarkResult     `json:"benchmarks,omitempty"`
	ResourceUsage   *ResourceAnalysis     `json:"resource_usage,omitempty"`
	Passed          bool                  `json:"passed"`
}

type PerformanceIssue struct {
	Type        string   `json:"type"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Line        int      `json:"line,omitempty"`
	Impact      string   `json:"impact"`
	Suggestion  string   `json:"suggestion"`
}

type PerformanceHint struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Example     string `json:"example,omitempty"`
}

type BenchmarkResult struct {
	Name        string        `json:"name"`
	Operations  int64         `json:"operations"`
	Duration    time.Duration `json:"duration"`
	MemoryUsage int64         `json:"memory_usage"`
	AllocsPerOp int64         `json:"allocs_per_op"`
}

type ResourceAnalysis struct {
	CPUComplexity    string `json:"cpu_complexity"`
	MemoryComplexity string `json:"memory_complexity"`
	IOComplexity     string `json:"io_complexity"`
	NetworkUsage     string `json:"network_usage"`
}

type ComplianceResult struct {
	Score       int                `json:"score"`
	Standards   []StandardResult   `json:"standards,omitempty"`
	Violations  []ComplianceIssue  `json:"violations,omitempty"`
	Passed      bool               `json:"passed"`
}

type StandardResult struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Score       int               `json:"score"`
	Passed      bool              `json:"passed"`
	Issues      []ComplianceIssue `json:"issues,omitempty"`
}

type ComplianceIssue struct {
	Standard    string   `json:"standard"`
	Rule        string   `json:"rule"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Line        int      `json:"line,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
}

type LLMCritiqueResult struct {
	Score        int      `json:"score"`
	Feedback     string   `json:"feedback"`
	Strengths    []string `json:"strengths,omitempty"`
	Weaknesses   []string `json:"weaknesses,omitempty"`
	Suggestions  []string `json:"suggestions,omitempty"`
	Improvements []string `json:"improvements,omitempty"`
	Confidence   float64  `json:"confidence"`
	Model        string   `json:"model,omitempty"`
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// API Request/Response types
type ValidateRequest struct {
	Content  string             `json:"content"`
	TaskType TaskType           `json:"task_type"`
	Language string             `json:"language,omitempty"`
	Context  *ValidationContext `json:"context,omitempty"`
	Options  *ValidationOptions `json:"options,omitempty"`
}

type ValidateResponse struct {
	ValidationID string `json:"validation_id"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
}

type GetValidationResponse struct {
	Validation ValidationResult `json:"validation"`
}

type ListValidationsRequest struct {
	TenantID   string     `json:"tenant_id"`
	Status     string     `json:"status,omitempty"`
	TaskType   TaskType   `json:"task_type,omitempty"`
	Level      ValidationLevel  `json:"level,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
	Since      *time.Time `json:"since,omitempty"`
}

type ListValidationsResponse struct {
	Validations []ValidationSummary `json:"validations"`
	Total       int                 `json:"total"`
}

type ValidationSummary struct {
	ID           string          `json:"id"`
	TaskID       string          `json:"task_id,omitempty"`
	Status       ValidationStatus `json:"status"`
	OverallScore int             `json:"overall_score"`
	Passed       bool            `json:"passed"`
	Level        ValidationLevel `json:"level"`
	ValidatedAt  time.Time       `json:"validated_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

// Batch validation types
type BatchValidationRequest struct {
	Requests []ValidateRequest `json:"requests"`
	Options  *BatchOptions     `json:"options,omitempty"`
}

type BatchOptions struct {
	Parallel    bool `json:"parallel"`
	MaxWorkers  int  `json:"max_workers,omitempty"`
	FailFast    bool `json:"fail_fast,omitempty"`
	Timeout     int  `json:"timeout_seconds,omitempty"`
}

type BatchValidationResponse struct {
	BatchID   string   `json:"batch_id"`
	Status    string   `json:"status"`
	RequestIDs []string `json:"request_ids"`
	Message   string   `json:"message,omitempty"`
}

// Handler-specific batch types (simpler than the above)
type BatchValidateRequest struct {
	Items []ValidateRequest `json:"items"`
}

type BatchValidateResponse struct {
	BatchID     string             `json:"batch_id"`
	TotalItems  int                `json:"total_items"`
	Validations []ValidateResponse `json:"validations"`
	Status      string             `json:"status"`
}

// Streaming validation types for real-time updates
type ValidationUpdate struct {
	ValidationID string          `json:"validation_id"`
	Status       ValidationStatus `json:"status"`
	Progress     int             `json:"progress_pct"`
	CurrentCheck string          `json:"current_check,omitempty"`
	PartialResult *ValidationResult `json:"partial_result,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	Error        string          `json:"error,omitempty"`
}