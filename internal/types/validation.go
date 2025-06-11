package types

import "time"

// ValidationResult represents common validation result structure to avoid import cycles
type ValidationResult struct {
	OverallScore     int                    `json:"overall_score"`
	SecurityScore    int                    `json:"security_score"`
	QualityScore     int                    `json:"quality_score"`
	Passed           bool                   `json:"passed"`
	ValidationTime   time.Duration          `json:"validation_time"`
	ValidatedAt      time.Time              `json:"validated_at"`
	SecurityResult   *SecurityResult        `json:"security_result,omitempty"`
	QualityResult    *QualityResult         `json:"quality_result,omitempty"`
}

// SecurityResult represents security validation results
type SecurityResult struct {
	Score              int                `json:"score"`
	RiskLevel          SecurityRiskLevel  `json:"risk_level"`
	Vulnerabilities    []SecurityIssue    `json:"vulnerabilities"`
	SandboxViolations  []string           `json:"sandbox_violations"`
	Passed             bool               `json:"passed"`
}

// QualityResult represents quality validation results  
type QualityResult struct {
	Score           int     `json:"score"`
	Coverage        float64 `json:"coverage"`
	Maintainability int     `json:"maintainability"`
	Documentation   int     `json:"documentation"`
	BestPractices   int     `json:"best_practices"`
	TestCoverage    float64 `json:"test_coverage"`
	Passed          bool    `json:"passed"`
}

// Common enum types to avoid import cycles
type SecurityRiskLevel string

const (
	SecurityRiskLevelNone     SecurityRiskLevel = "none"
	SecurityRiskLevelLow      SecurityRiskLevel = "low"
	SecurityRiskLevelMedium   SecurityRiskLevel = "medium"
	SecurityRiskLevelHigh     SecurityRiskLevel = "high"
	SecurityRiskLevelCritical SecurityRiskLevel = "critical"
	
	// String constants for severity comparisons
	SecurityRiskNoneStr     = "none"
	SecurityRiskLowStr      = "low"
	SecurityRiskMediumStr   = "medium"
	SecurityRiskHighStr     = "high"
	SecurityRiskCriticalStr = "critical"
)

type SecurityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// Test related types to avoid import cycles
type TestSuite struct {
	Name  string     `json:"name"`
	Tests []TestCase `json:"tests"`
}

type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Endpoint    string `json:"endpoint"`
}

// SecurityFinding represents a security vulnerability or issue
type SecurityFinding struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	Recommendation string `json:"recommendation"`
	CWE            string `json:"cwe,omitempty"`
	OWASP          string `json:"owasp,omitempty"`
}

// NewTestSuite creates a new test suite
func NewTestSuite() *TestSuite {
	return &TestSuite{
		Name:  "Generated Test Suite",
		Tests: []TestCase{},
	}
}

// QuantumCapsule basic definition to avoid cycles
type QuantumCapsule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Drops       []QuantumDrop `json:"drops"`
}

type QuantumDrop struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Files map[string]string `json:"files"`
}