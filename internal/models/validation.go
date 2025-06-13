package models

import "time"

// Location specifies a position within a file.
type Location struct {
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// Issue represents a general problem or deviation from best practices.
type Issue struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // e.g., "critical", "high", "medium", "low"
	Location    *Location `json:"location,omitempty"`
	Suggestion  string    `json:"suggestion"`
}

// SecurityFinding represents a potential security vulnerability.
type SecurityFinding struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    *Location `json:"location,omitempty"`
	Confidence  float64   `json:"confidence"` // Confidence score from the scanner
	Remediation string    `json:"remediation"`
}

// ValidationResult holds the complete output of a validation process for an artifact.
type ValidationResult struct {
	Artifact         Artifact          `json:"artifact"`
	Passed           bool              `json:"passed"`
	OverallScore     int               `json:"overall_score"` // 0-100
	ComponentScores  map[string]int    `json:"component_scores"`
	Issues           []Issue           `json:"issues"`
	SecurityFindings []SecurityFinding `json:"security_findings"`
	ValidatedAt      time.Time         `json:"validated_at"`
}
