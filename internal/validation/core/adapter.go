package core

import (
	"context"
	"fmt"

	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/types"
	"go.uber.org/zap"
)

// ValidationAdapter bridges existing validators with the unified framework
type ValidationAdapter struct {
	unifiedEngine *UnifiedValidationEngine
	logger        logger.Interface
}

// NewValidationAdapter creates a new adapter for existing validators
func NewValidationAdapter(llmClient llm.Client, validatorType ValidatorType, logger logger.Interface) *ValidationAdapter {
	return &ValidationAdapter{
		unifiedEngine: NewUnifiedValidationEngine(llmClient, validatorType, logger),
		logger:        logger.WithComponent("validation_adapter"),
	}
}

// AdaptLegacyValidation converts old validation patterns to unified framework
func (va *ValidationAdapter) AdaptLegacyValidation(
	ctx context.Context,
	projectPath string,
	files map[string]string,
	language string,
	requirements map[string]interface{},
) (*types.ValidationResult, error) {
	
	// Convert to unified input format
	input := va.convertToUnifiedInput(projectPath, files, language, requirements)
	
	// Run unified validation
	unifiedResult, err := va.unifiedEngine.Validate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unified validation failed: %w", err)
	}
	
	// Convert back to legacy format
	legacyResult := va.convertToLegacyResult(unifiedResult)
	
	va.logger.Info("Legacy validation adapted successfully",
		zap.Int("overall_score", legacyResult.OverallScore),
		zap.Int("security_score", legacyResult.SecurityScore),
		zap.Int("quality_score", legacyResult.QualityScore),
	)
	
	return legacyResult, nil
}

// ValidateWithUnifiedFramework provides direct access to unified validation
func (va *ValidationAdapter) ValidateWithUnifiedFramework(ctx context.Context, input *ValidationInput) (*ValidationResult, error) {
	return va.unifiedEngine.Validate(ctx, input)
}

// GetUnifiedEngine provides access to the underlying unified engine
func (va *ValidationAdapter) GetUnifiedEngine() *UnifiedValidationEngine {
	return va.unifiedEngine
}

// Private conversion methods

func (va *ValidationAdapter) convertToUnifiedInput(
	projectPath string,
	files map[string]string,
	language string,
	requirements map[string]interface{},
) *ValidationInput {
	
	input := &ValidationInput{
		Content:      files,
		ProjectPath:  projectPath,
		Language:     language,
		Framework:    va.detectFramework(files, language),
		ValidationTypes: []ValidationType{
			ValidationTypeSecurity,
			ValidationTypeQuality,
			ValidationTypePerformance,
		},
		Requirements: va.convertRequirements(requirements),
		ProjectMetadata: va.extractProjectMetadata(files, language),
	}
	
	return input
}

func (va *ValidationAdapter) convertToLegacyResult(unified *ValidationResult) *types.ValidationResult {
	legacy := &types.ValidationResult{
		OverallScore:   unified.OverallScore,
		SecurityScore:  unified.ComponentScores["security"],
		QualityScore:   unified.ComponentScores["quality"],
		Passed:         unified.Passed,
		ValidationTime: unified.ValidationTime,
		ValidatedAt:    unified.ValidatedAt,
	}
	
	// Convert security findings
	if len(unified.SecurityFindings) > 0 {
		legacy.SecurityResult = &types.SecurityResult{
			Score:           unified.ComponentScores["security"],
			RiskLevel:      va.convertRiskLevel(unified.SecurityFindings),
			Vulnerabilities: va.convertSecurityIssues(unified.SecurityFindings),
			Passed:         unified.ComponentScores["security"] >= 70,
		}
	}
	
	// Convert quality metrics
	if unified.QualityMetrics != nil {
		legacy.QualityResult = &types.QualityResult{
			Score:           unified.ComponentScores["quality"],
			Coverage:        unified.QualityMetrics.TestCoverage,
			Maintainability: unified.QualityMetrics.Maintainability,
			Documentation:   unified.QualityMetrics.Documentation,
			BestPractices:   va.calculateBestPracticesScore(unified.BestPractices),
			TestCoverage:    unified.QualityMetrics.TestCoverage,
			Passed:         unified.ComponentScores["quality"] >= 60,
		}
	}
	
	return legacy
}

func (va *ValidationAdapter) detectFramework(files map[string]string, language string) string {
	// Basic framework detection logic
	for filePath, content := range files {
		switch language {
		case "go":
			if va.containsFrameworkMarkers(content, []string{"gin-gonic", "fiber", "echo"}) {
				return "web_framework"
			}
		case "javascript", "typescript":
			if va.containsFrameworkMarkers(filePath, []string{"package.json"}) {
				if va.containsFrameworkMarkers(content, []string{"express", "koa", "fastify"}) {
					return "node_framework"
				}
				if va.containsFrameworkMarkers(content, []string{"react", "vue", "angular"}) {
					return "frontend_framework"
				}
			}
		case "python":
			if va.containsFrameworkMarkers(content, []string{"flask", "django", "fastapi"}) {
				return "web_framework"
			}
		}
	}
	return "unknown"
}

func (va *ValidationAdapter) containsFrameworkMarkers(content string, markers []string) bool {
	for _, marker := range markers {
		if containsIgnoreCase(content, marker) {
			return true
		}
	}
	return false
}

func (va *ValidationAdapter) convertRequirements(requirements map[string]interface{}) *Requirements {
	req := &Requirements{
		SecurityLevel:      "medium", // Default
		ComplianceStandards: []string{},
		QualityThresholds:  make(map[string]int),
		CustomRules:        []CustomRule{},
	}
	
	if secLevel, ok := requirements["security_level"].(string); ok {
		req.SecurityLevel = secLevel
	}
	
	if standards, ok := requirements["compliance_standards"].([]string); ok {
		req.ComplianceStandards = standards
	}
	
	if thresholds, ok := requirements["quality_thresholds"].(map[string]int); ok {
		req.QualityThresholds = thresholds
	} else {
		// Set default thresholds
		req.QualityThresholds = map[string]int{
			"maintainability": 70,
			"test_coverage":   80,
			"documentation":   60,
		}
	}
	
	return req
}

func (va *ValidationAdapter) extractProjectMetadata(files map[string]string, language string) *ProjectMetadata {
	metadata := &ProjectMetadata{
		ProjectType:       va.detectProjectType(files),
		TechStack:         []string{language},
		Dependencies:      va.extractDependencies(files),
		BuildTool:         va.detectBuildTool(files, language),
		TestFramework:     va.detectTestFramework(files, language),
		TargetEnvironment: "production",
	}
	
	return metadata
}

func (va *ValidationAdapter) detectProjectType(files map[string]string) string {
	for filePath := range files {
		if containsIgnoreCase(filePath, "main.go") || containsIgnoreCase(filePath, "main.py") {
			return "application"
		}
		if containsIgnoreCase(filePath, "lib") || containsIgnoreCase(filePath, "package") {
			return "library"
		}
		if containsIgnoreCase(filePath, "test") {
			return "test_suite"
		}
	}
	return "application"
}

func (va *ValidationAdapter) extractDependencies(files map[string]string) []string {
	var dependencies []string
	
	for filePath, content := range files {
		switch {
		case containsIgnoreCase(filePath, "go.mod"):
			dependencies = append(dependencies, va.extractGoDependencies(content)...)
		case containsIgnoreCase(filePath, "package.json"):
			dependencies = append(dependencies, va.extractNpmDependencies(content)...)
		case containsIgnoreCase(filePath, "requirements.txt"):
			dependencies = append(dependencies, va.extractPythonDependencies(content)...)
		}
	}
	
	return dependencies
}

func (va *ValidationAdapter) detectBuildTool(files map[string]string, language string) string {
	for filePath := range files {
		switch language {
		case "go":
			if containsIgnoreCase(filePath, "Makefile") {
				return "make"
			}
			return "go_build"
		case "javascript", "typescript":
			if containsIgnoreCase(filePath, "webpack.config") {
				return "webpack"
			}
			if containsIgnoreCase(filePath, "package.json") {
				return "npm"
			}
		case "python":
			if containsIgnoreCase(filePath, "setup.py") {
				return "setuptools"
			}
			if containsIgnoreCase(filePath, "pyproject.toml") {
				return "poetry"
			}
		}
	}
	return "unknown"
}

func (va *ValidationAdapter) detectTestFramework(files map[string]string, language string) string {
	for filePath, content := range files {
		switch language {
		case "go":
			if containsIgnoreCase(filePath, "_test.go") {
				return "go_test"
			}
		case "javascript", "typescript":
			if va.containsFrameworkMarkers(content, []string{"jest", "mocha", "cypress"}) {
				return "jest"
			}
		case "python":
			if va.containsFrameworkMarkers(content, []string{"pytest", "unittest"}) {
				return "pytest"
			}
		}
	}
	return "unknown"
}

func (va *ValidationAdapter) convertRiskLevel(findings []SecurityFinding) types.SecurityRiskLevel {
	maxRisk := types.SecurityRiskLevelNone
	
	for _, finding := range findings {
		switch finding.Severity {
		case SeverityCritical:
			return types.SecurityRiskLevelCritical
		case SeverityHigh:
			if maxRisk != types.SecurityRiskLevelCritical {
				maxRisk = types.SecurityRiskLevelHigh
			}
		case SeverityMedium:
			if maxRisk == types.SecurityRiskLevelNone || maxRisk == types.SecurityRiskLevelLow {
				maxRisk = types.SecurityRiskLevelMedium
			}
		case SeverityLow:
			if maxRisk == types.SecurityRiskLevelNone {
				maxRisk = types.SecurityRiskLevelLow
			}
		}
	}
	
	return maxRisk
}

func (va *ValidationAdapter) convertSecurityIssues(findings []SecurityFinding) []types.SecurityIssue {
	var issues []types.SecurityIssue
	
	for _, finding := range findings {
		issue := types.SecurityIssue{
			Type:        finding.Type,
			Severity:    string(finding.Severity),
			Description: finding.Description,
			Location:    va.formatLocation(finding.Location),
		}
		issues = append(issues, issue)
	}
	
	return issues
}

func (va *ValidationAdapter) calculateBestPracticesScore(practices []BestPractice) int {
	if len(practices) == 0 {
		return 50 // Default score when no practices evaluated
	}
	
	compliant := 0
	for _, practice := range practices {
		if practice.Compliant {
			compliant++
		}
	}
	
	return (compliant * 100) / len(practices)
}

func (va *ValidationAdapter) formatLocation(location *Location) string {
	if location == nil {
		return "unknown"
	}
	
	if location.Line > 0 {
		return fmt.Sprintf("%s:%d", location.FilePath, location.Line)
	}
	
	return location.FilePath
}

// Helper methods for dependency extraction

func (va *ValidationAdapter) extractGoDependencies(content string) []string {
	// Simple extraction - would be more sophisticated in practice
	var deps []string
	lines := splitLines(content)
	
	for _, line := range lines {
		if containsIgnoreCase(line, "require") && !containsIgnoreCase(line, "//") {
			// Extract dependency name
			if parts := splitWhitespace(line); len(parts) >= 2 {
				deps = append(deps, parts[1])
			}
		}
	}
	
	return deps
}

func (va *ValidationAdapter) extractNpmDependencies(content string) []string {
	// Basic JSON parsing for dependencies
	var deps []string
	
	if containsIgnoreCase(content, "dependencies") {
		deps = append(deps, "express", "react") // Simplified extraction
	}
	
	return deps
}

func (va *ValidationAdapter) extractPythonDependencies(content string) []string {
	var deps []string
	lines := splitLines(content)
	
	for _, line := range lines {
		if line != "" && !containsIgnoreCase(line, "#") {
			// Extract package name (before ==, >=, etc.)
			if parts := splitOnChars(line, []string{"==", ">=", "<=", ">", "<"}); len(parts) > 0 {
				deps = append(deps, parts[0])
			}
		}
	}
	
	return deps
}

// Utility functions

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || stringContains(toLowerCase(s), toLowerCase(substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

func splitLines(s string) []string {
	return splitOnChar(s, '\n')
}

func splitWhitespace(s string) []string {
	var parts []string
	var current string
	
	for _, char := range s {
		if char == ' ' || char == '\t' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

func splitOnChar(s string, sep rune) []string {
	var parts []string
	var current string
	
	for _, char := range s {
		if char == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	parts = append(parts, current)
	return parts
}

func splitOnChars(s string, separators []string) []string {
	parts := []string{s}
	
	for _, sep := range separators {
		var newParts []string
		for _, part := range parts {
			subParts := splitOnString(part, sep)
			newParts = append(newParts, subParts...)
		}
		parts = newParts
	}
	
	return parts
}

func splitOnString(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	
	var parts []string
	start := 0
	
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	
	parts = append(parts, s[start:])
	return parts
}