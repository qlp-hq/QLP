package core

import (
	"context"
	"time"
)

// UniversalValidator defines the unified interface for all validation types
type UniversalValidator interface {
	Validate(ctx context.Context, input *ValidationInput) (*ValidationResult, error)
	GetValidatorType() ValidatorType
	GetSupportedLanguages() []string
	GetConfidenceThreshold() float64
}

// ValidationInput provides unified input structure for all validators
type ValidationInput struct {
	// Source content
	Content     map[string]string `json:"content"` // file_path -> content
	ProjectPath string            `json:"project_path"`
	Language    string            `json:"language"`
	Framework   string            `json:"framework"`

	// Validation scope
	ValidationTypes []ValidationType `json:"validation_types"`
	Requirements    *Requirements    `json:"requirements"`

	// Context
	UserContext     string           `json:"user_context"`
	ProjectMetadata *ProjectMetadata `json:"project_metadata"`
}

// ValidationResult provides unified result structure
type ValidationResult struct {
	// Overall assessment
	OverallScore int     `json:"overall_score"`
	Confidence   float64 `json:"confidence"`
	Passed       bool    `json:"passed"`

	// Component scores
	ComponentScores map[string]int `json:"component_scores"`

	// Detailed results
	Issues          []Issue          `json:"issues"`
	Warnings        []Warning        `json:"warnings"`
	Recommendations []Recommendation `json:"recommendations"`

	// Security specific
	SecurityFindings []SecurityFinding `json:"security_findings,omitempty"`
	ComplianceStatus map[string]bool   `json:"compliance_status,omitempty"`

	// Quality specific
	QualityMetrics *QualityMetrics `json:"quality_metrics,omitempty"`
	BestPractices  []BestPractice  `json:"best_practices,omitempty"`

	// Performance specific
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics,omitempty"`

	// Metadata
	ValidatorType    ValidatorType `json:"validator_type"`
	ValidationTime   time.Duration `json:"validation_time"`
	ValidatedAt      time.Time     `json:"validated_at"`
	ValidatorVersion string        `json:"validator_version"`
}

// ValidatorType enum for different validator implementations
type ValidatorType string

const (
	ValidatorTypeUniversal  ValidatorType = "universal"
	ValidatorTypeStatic     ValidatorType = "static"
	ValidatorTypeSecurity   ValidatorType = "security"
	ValidatorTypeDeployment ValidatorType = "deployment"
	ValidatorTypeEnterprise ValidatorType = "enterprise"
	ValidatorTypeSyntax     ValidatorType = "syntax"
)

// ValidationType enum for different validation aspects
type ValidationType string

const (
	ValidationTypeSyntax       ValidationType = "syntax"
	ValidationTypeSecurity     ValidationType = "security"
	ValidationTypeQuality      ValidationType = "quality"
	ValidationTypePerformance  ValidationType = "performance"
	ValidationTypeDeployment   ValidationType = "deployment"
	ValidationTypeCompliance   ValidationType = "compliance"
	ValidationTypeArchitecture ValidationType = "architecture"
)

// ScoringEngine provides unified scoring mechanisms
type ScoringEngine interface {
	CalculateOverallScore(componentScores map[string]int, weights map[string]float64) int
	CalculateComponentScore(metrics map[string]interface{}, rules []ScoringRule) int
	GetDefaultWeights(validatorType ValidatorType) map[string]float64
	ApplyPenalties(score int, penalties []Penalty) int
}

// PatternEngine provides unified pattern matching
type PatternEngine interface {
	MatchPatterns(content string, patterns []Pattern) []Match
	GetPatternsForLanguage(language string) []Pattern
	GetSecurityPatterns() []SecurityPattern
	GetQualityPatterns() []QualityPattern
}

// LLMIntegration provides unified LLM interaction
type LLMIntegration interface {
	GeneratePrompt(input *ValidationInput, promptType PromptType) (string, error)
	ParseResponse(response string, expectedType ResponseType) (interface{}, error)
	GetPromptTemplate(validatorType ValidatorType, promptType PromptType) string
	ValidateResponse(response interface{}) error
}

// Supporting types for unified validation framework

type Requirements struct {
	SecurityLevel       string         `json:"security_level"`
	ComplianceStandards []string       `json:"compliance_standards"`
	QualityThresholds   map[string]int `json:"quality_thresholds"`
	CustomRules         []CustomRule   `json:"custom_rules"`
}

type ProjectMetadata struct {
	ProjectType       string   `json:"project_type"`
	TechStack         []string `json:"tech_stack"`
	Dependencies      []string `json:"dependencies"`
	BuildTool         string   `json:"build_tool"`
	TestFramework     string   `json:"test_framework"`
	TargetEnvironment string   `json:"target_environment"`
}

type Issue struct {
	ID          string    `json:"id"`
	Type        IssueType `json:"type"`
	Severity    Severity  `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    *Location `json:"location,omitempty"`
	Suggestion  string    `json:"suggestion"`
	References  []string  `json:"references,omitempty"`
}

type Warning struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	Location   *Location `json:"location,omitempty"`
	Suggestion string    `json:"suggestion"`
}

type Recommendation struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Impact      Impact   `json:"impact"`
	Effort      Effort   `json:"effort"`
	Actions     []Action `json:"actions"`
}

type SecurityFinding struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    Severity  `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    *Location `json:"location,omitempty"`
	CWE         string    `json:"cwe,omitempty"`
	CVSS        float64   `json:"cvss,omitempty"`
	Remediation string    `json:"remediation"`
	References  []string  `json:"references,omitempty"`
	Confidence  float64   `json:"confidence"`
}

type QualityMetrics struct {
	Complexity      int     `json:"complexity"`
	Maintainability int     `json:"maintainability"`
	Readability     int     `json:"readability"`
	TestCoverage    float64 `json:"test_coverage"`
	Documentation   int     `json:"documentation"`
	CodeDuplication float64 `json:"code_duplication"`
	TechnicalDebt   string  `json:"technical_debt"`
}

type PerformanceMetrics struct {
	EstimatedMemoryUsage    int      `json:"estimated_memory_usage"`
	EstimatedCPUUsage       float64  `json:"estimated_cpu_usage"`
	AlgorithmComplexity     string   `json:"algorithm_complexity"`
	Bottlenecks             []string `json:"bottlenecks"`
	OptimizationSuggestions []string `json:"optimization_suggestions"`
}

type BestPractice struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Compliant   bool   `json:"compliant"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type Location struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Function  string `json:"function,omitempty"`
}

// Enums and constants

type IssueType string

const (
	IssueTypeSyntax      IssueType = "syntax"
	IssueTypeSecurity    IssueType = "security"
	IssueTypeQuality     IssueType = "quality"
	IssueTypePerformance IssueType = "performance"
	IssueTypeCompliance  IssueType = "compliance"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Impact string

const (
	ImpactLow    Impact = "low"
	ImpactMedium Impact = "medium"
	ImpactHigh   Impact = "high"
)

type Effort string

const (
	EffortLow    Effort = "low"
	EffortMedium Effort = "medium"
	EffortHigh   Effort = "high"
)

type Action struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
}

// Pattern matching types
type Pattern struct {
	ID          string      `json:"id"`
	Type        PatternType `json:"type"`
	Regex       string      `json:"regex"`
	Description string      `json:"description"`
	Severity    Severity    `json:"severity"`
	Category    string      `json:"category"`
}

type SecurityPattern struct {
	Pattern
	CWE        string   `json:"cwe,omitempty"`
	OWASP      string   `json:"owasp,omitempty"`
	References []string `json:"references,omitempty"`
}

type QualityPattern struct {
	Pattern
	BestPractice string `json:"best_practice"`
	Suggestion   string `json:"suggestion"`
}

type Match struct {
	Pattern    Pattern  `json:"pattern"`
	Location   Location `json:"location"`
	Confidence float64  `json:"confidence"`
	Context    string   `json:"context"`
}

type PatternType string

const (
	PatternTypeSecurity    PatternType = "security"
	PatternTypeQuality     PatternType = "quality"
	PatternTypePerformance PatternType = "performance"
	PatternTypeCompliance  PatternType = "compliance"
)

// LLM integration types
type PromptType string

const (
	PromptTypeAnalysis    PromptType = "analysis"
	PromptTypeValidation  PromptType = "validation"
	PromptTypeSuggestion  PromptType = "suggestion"
	PromptTypeExplanation PromptType = "explanation"
)

type ResponseType string

const (
	ResponseTypeJSON       ResponseType = "json"
	ResponseTypeStructured ResponseType = "structured"
	ResponseTypeText       ResponseType = "text"
)

// Scoring types
type ScoringRule struct {
	Name      string                 `json:"name"`
	Weight    float64                `json:"weight"`
	Condition map[string]interface{} `json:"condition"`
	Points    int                    `json:"points"`
}

type Penalty struct {
	Type       string  `json:"type"`
	Reason     string  `json:"reason"`
	Points     int     `json:"points"`
	Percentage float64 `json:"percentage,omitempty"`
}

type CustomRule struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      ValidationType         `json:"type"`
	Condition map[string]interface{} `json:"condition"`
	Action    string                 `json:"action"`
	Severity  Severity               `json:"severity"`
	Message   string                 `json:"message"`
}
