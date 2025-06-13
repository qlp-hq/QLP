package models

import "time"

// Artifact represents the output of a single Task execution by an Agent.
type Artifact struct {
	ID        string            `json:"id"`
	Task      Task              `json:"task"`
	Type      ArtifactType      `json:"type"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

// ArtifactType defines the type of content an artifact holds.
type ArtifactType string

const (
	ArtifactTypeSourceCode ArtifactType = "source_code"
	ArtifactTypeUnitTest   ArtifactType = "unit_test"
	ArtifactTypeDocument   ArtifactType = "document"
	ArtifactTypeInfraPlan  ArtifactType = "infra_plan"
	ArtifactTypeAnalysis   ArtifactType = "analysis_report"
)
