package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/logger"

	"go.uber.org/zap"
)

const (
	maxRefinementCycles   = 3
	passingScoreThreshold = 95
	// Other constants
)

// UnifiedValidationEngine combines all core components into a single validator
type UnifiedValidationEngine struct {
	llmIntegration      LLMIntegration
	scoringEngine       ScoringEngine
	patternEngine       PatternEngine
	logger              logger.Interface
	validatorType       ValidatorType
	confidenceThreshold float64
}

// NewUnifiedValidationEngine creates a new unified validation engine
func NewUnifiedValidationEngine(
	llmClient llm.Client,
	validatorType ValidatorType,
	logger logger.Interface,
) *UnifiedValidationEngine {
	return &UnifiedValidationEngine{
		llmIntegration:      NewDefaultLLMIntegration(llmClient),
		scoringEngine:       NewDefaultScoringEngine(),
		patternEngine:       NewDefaultPatternEngine(),
		logger:              logger.WithComponent("unified_validator"),
		validatorType:       validatorType,
		confidenceThreshold: 0.7,
	}
}

// Validate implements the UniversalValidator interface
func (uve *UnifiedValidationEngine) Validate(ctx context.Context, input *ValidationInput) (*ValidationResult, error) {
	var lastResult *ValidationResult
	currentCode := input.Content

	for i := 0; i < maxRefinementCycles; i++ {
		uve.logger.Info("Starting validation cycle", zap.Int("cycle", i+1), zap.Int("max_cycles", maxRefinementCycles))

		// Use a cycle-specific input
		cycleInput := *input
		cycleInput.Content = currentCode

		startTime := time.Now()
		result := &ValidationResult{
			ValidatorType:    uve.validatorType,
			ValidatedAt:      startTime,
			ComponentScores:  make(map[string]int),
			Issues:           []Issue{},
			Warnings:         []Warning{},
			Recommendations:  []Recommendation{},
			SecurityFindings: []SecurityFinding{},
			BestPractices:    []BestPractice{},
		}

		if err := uve.performPatternAnalysis(&cycleInput, result); err != nil {
			uve.logger.Error("Pattern analysis failed", zap.Error(err))
			// If pattern analysis fails, it's likely a fundamental issue, so we stop.
			return nil, fmt.Errorf("pattern analysis failed on cycle %d: %w", i+1, err)
		}

		if err := uve.performLLMAnalysis(ctx, &cycleInput, result); err != nil {
			uve.logger.Warn("LLM analysis failed, continuing with pattern-only results", zap.Error(err))
			// We can choose to stop or continue here. For now, let's stop if the core analysis fails.
			return nil, fmt.Errorf("LLM analysis failed on cycle %d: %w", i+1, err)
		}

		uve.calculateFinalScores(result)
		uve.applyValidationRules(result)
		result.ValidationTime = time.Since(startTime)
		result.Passed = uve.determineOverallPass(result) && result.OverallScore >= passingScoreThreshold

		lastResult = result

		// If validation passed, we can exit the loop.
		if result.Passed {
			uve.logger.Info("Validation passed.", zap.Int("cycle", i+1), zap.Int("score", result.OverallScore))
			break
		}

		uve.logger.Warn("Validation failed on cycle.",
			zap.Int("cycle", i+1),
			zap.Int("score", result.OverallScore),
			zap.Int("passing_threshold", passingScoreThreshold),
		)

		// If it's not the last cycle, attempt to refine the code
		if i < maxRefinementCycles-1 {
			uve.logger.Info("Attempting code refinement.", zap.Int("cycle", i+1))
			refinementPrompt, err := uve.generateRefinementPrompt(result)
			if err != nil {
				uve.logger.Error("Failed to generate refinement prompt, ending.", zap.Error(err))
				// Can't generate a fix, so no point continuing.
				break
			}

			fixedCode, err := uve.refineCode(ctx, currentCode, refinementPrompt)
			if err != nil {
				uve.logger.Error("Failed to refine code, ending.", zap.Error(err))
				// LLM failed to fix the code, so we stop.
				break
			}
			currentCode = fixedCode
		}
	}

	return lastResult, nil
}

func (uve *UnifiedValidationEngine) generateRefinementPrompt(result *ValidationResult) (string, error) {
	if len(result.Issues) == 0 {
		return "", fmt.Errorf("no issues found to generate refinement prompt")
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString("The following code has been reviewed and found to have issues. Please fix the code to address the following points. Only provide the complete, corrected code for the files that need changes. Do not provide any commentary or explanation outside of the code blocks.\n\n")

	for _, issue := range result.Issues {
		promptBuilder.WriteString(fmt.Sprintf("- **Issue:** %s (Severity: %s)\n", issue.Title, issue.Severity))
		promptBuilder.WriteString(fmt.Sprintf("  - **Description:** %s\n", issue.Description))
		if issue.Location != nil {
			promptBuilder.WriteString(fmt.Sprintf("  - **File:** %s, Line: %d\n", issue.Location.FilePath, issue.Location.Line))
		}
		promptBuilder.WriteString(fmt.Sprintf("  - **Suggestion:** %s\n\n", issue.Suggestion))
	}

	return promptBuilder.String(), nil
}

func (uve *UnifiedValidationEngine) refineCode(ctx context.Context, originalCode map[string]string, prompt string) (map[string]string, error) {
	// This is a conceptual implementation. We're creating a single large prompt
	// with the instructions and the original code.
	var fullPrompt strings.Builder
	fullPrompt.WriteString(prompt)
	fullPrompt.WriteString("\n--- Original Code ---\n")

	for path, content := range originalCode {
		fullPrompt.WriteString(fmt.Sprintf("\n--- File: %s ---\n", path))
		fullPrompt.WriteString(content)
	}

	// Make the LLM call to get the fixed code
	llmResponse, err := uve.llmIntegration.GeneratePrompt(&ValidationInput{}, "refine") // This is not ideal, need a better way
	if err != nil {
		return nil, fmt.Errorf("LLM refinement call failed: %w", err)
	}

	// The response from the LLM is expected to be just the code.
	// We need to parse this response to update our `currentCode` map.
	// This parsing logic can be complex and is a critical part of this workflow.
	// For now, we'll assume a simple parsing strategy.
	return uve.parseRefinedCode(llmResponse), nil
}

func (uve *UnifiedValidationEngine) parseRefinedCode(llmResponse string) map[string]string {
	// Placeholder for complex parsing logic.
	// A real implementation would need to reliably extract file paths and their new content
	// from the unstructured LLM response.
	refinedCode := make(map[string]string)
	// Example:
	// ... logic to find "--- File: main.go ---" and extract content until the next file block ...
	return refinedCode
}

// GetValidatorType returns the validator type
func (uve *UnifiedValidationEngine) GetValidatorType() ValidatorType {
	return uve.validatorType
}

// GetSupportedLanguages returns supported programming languages
func (uve *UnifiedValidationEngine) GetSupportedLanguages() []string {
	// Universal validator supports all languages through LLM
	return []string{"*"} // Wildcard indicates universal support
}

// GetConfidenceThreshold returns the minimum confidence threshold
func (uve *UnifiedValidationEngine) GetConfidenceThreshold() float64 {
	return uve.confidenceThreshold
}

// SetConfidenceThreshold allows adjusting the confidence threshold
func (uve *UnifiedValidationEngine) SetConfidenceThreshold(threshold float64) {
	if threshold >= 0.0 && threshold <= 1.0 {
		uve.confidenceThreshold = threshold
	}
}

// Private methods for validation steps

func (uve *UnifiedValidationEngine) performPatternAnalysis(input *ValidationInput, result *ValidationResult) error {
	// Get language-specific patterns
	patterns := uve.patternEngine.GetPatternsForLanguage(input.Language)
	securityPatterns := uve.patternEngine.GetSecurityPatterns()
	qualityPatterns := uve.patternEngine.GetQualityPatterns()

	var allMatches []Match
	var securityFindings []SecurityFinding
	var qualityIssues []Issue

	// Analyze each file
	for filePath, content := range input.Content {
		// Pattern matching
		matches := uve.patternEngine.MatchPatterns(content, patterns)
		allMatches = append(allMatches, matches...)

		// Security pattern analysis
		for _, secPattern := range securityPatterns {
			secMatches := uve.patternEngine.MatchPatterns(content, []Pattern{secPattern.Pattern})
			for _, match := range secMatches {
				if match.Confidence >= uve.confidenceThreshold {
					finding := SecurityFinding{
						ID:          fmt.Sprintf("sec_%s_%d", secPattern.ID, match.Location.Line),
						Type:        string(secPattern.Type),
						Severity:    secPattern.Severity,
						Title:       secPattern.Description,
						Description: fmt.Sprintf("Security pattern '%s' detected at %s:%d", secPattern.ID, filePath, match.Location.Line),
						Location: &Location{
							FilePath: filePath,
							Line:     match.Location.Line,
							Column:   match.Location.Column,
						},
						CWE:         secPattern.CWE,
						Remediation: "Review and address this security concern",
						References:  secPattern.References,
						Confidence:  match.Confidence,
					}
					securityFindings = append(securityFindings, finding)
				}
			}
		}

		// Quality pattern analysis
		for _, qualPattern := range qualityPatterns {
			qualMatches := uve.patternEngine.MatchPatterns(content, []Pattern{qualPattern.Pattern})
			for _, match := range qualMatches {
				if match.Confidence >= uve.confidenceThreshold {
					issue := Issue{
						ID:          fmt.Sprintf("qual_%s_%d", qualPattern.ID, match.Location.Line),
						Type:        IssueTypeQuality,
						Severity:    qualPattern.Severity,
						Title:       qualPattern.Description,
						Description: fmt.Sprintf("Quality issue '%s' detected at %s:%d", qualPattern.ID, filePath, match.Location.Line),
						Location: &Location{
							FilePath: filePath,
							Line:     match.Location.Line,
							Column:   match.Location.Column,
						},
						Suggestion: qualPattern.Suggestion,
					}
					qualityIssues = append(qualityIssues, issue)
				}
			}
		}
	}

	result.SecurityFindings = securityFindings
	result.Issues = append(result.Issues, qualityIssues...)

	uve.logger.Debug("Pattern analysis completed",
		zap.Int("total_matches", len(allMatches)),
		zap.Int("security_findings", len(securityFindings)),
		zap.Int("quality_issues", len(qualityIssues)),
	)

	return nil
}

func (uve *UnifiedValidationEngine) performLLMAnalysis(ctx context.Context, input *ValidationInput, result *ValidationResult) error {
	// Generate analysis prompt
	prompt, err := uve.llmIntegration.GeneratePrompt(input, PromptTypeAnalysis)
	if err != nil {
		return fmt.Errorf("failed to generate dynamic prompt: %w", err)
	}

	// Make LLM request
	response, err := uve.makeLLMRequest(ctx, prompt)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	// Parse LLM response
	llmResult, err := uve.llmIntegration.ParseResponse(response, ResponseTypeJSON)
	if err != nil {
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Validate response quality
	if err := uve.llmIntegration.ValidateResponse(llmResult); err != nil {
		return fmt.Errorf("LLM response validation failed: %w", err)
	}

	// Merge LLM results with existing results
	if validationResult, ok := llmResult.(*ValidationResult); ok {
		uve.mergeLLMResults(result, validationResult)
	}

	return nil
}

func (uve *UnifiedValidationEngine) calculateFinalScores(result *ValidationResult) {
	// Calculate component scores based on findings
	result.ComponentScores["security"] = uve.calculateSecurityScore(result.SecurityFindings)
	result.ComponentScores["quality"] = uve.calculateQualityScore(result.Issues)
	result.ComponentScores["performance"] = uve.calculatePerformanceScore(result)

	// Get default weights for this validator type
	weights := uve.scoringEngine.GetDefaultWeights(uve.validatorType)

	// Calculate overall score
	result.OverallScore = uve.scoringEngine.CalculateOverallScore(result.ComponentScores, weights)

	// Calculate confidence
	confidenceFactors := map[string]float64{
		"pattern_matches": uve.calculatePatternConfidence(result),
		"llm_confidence":  0.85, // Would be set from actual LLM response
		"completeness":    uve.calculateCompletenessScore(result),
	}
	// Create a default scoring engine to calculate confidence
	defaultEngine := NewDefaultScoringEngine()
	result.Confidence = defaultEngine.CalculateConfidenceScore(confidenceFactors)
}

func (uve *UnifiedValidationEngine) applyValidationRules(result *ValidationResult) {
	// Apply penalties for critical issues
	var penalties []Penalty

	for _, finding := range result.SecurityFindings {
		if finding.Severity == SeverityCritical {
			penalties = append(penalties, Penalty{
				Type:       "security_critical",
				Reason:     finding.Title,
				Points:     20,
				Percentage: 0.0,
			})
		}
	}

	if len(penalties) > 0 {
		result.OverallScore = uve.scoringEngine.ApplyPenalties(result.OverallScore, penalties)
	}
}

func (uve *UnifiedValidationEngine) determineOverallPass(result *ValidationResult) bool {
	// Pass criteria based on validator type
	switch uve.validatorType {
	case ValidatorTypeSecurity:
		return result.OverallScore >= 80 && !uve.hasCriticalSecurityIssues(result)
	case ValidatorTypeStatic:
		return result.OverallScore >= 70
	case ValidatorTypeDeployment:
		return result.OverallScore >= 85 && result.Confidence >= 0.8
	default:
		return result.OverallScore >= 75 && result.Confidence >= 0.7
	}
}

// Helper methods

func (uve *UnifiedValidationEngine) makeLLMRequest(ctx context.Context, prompt string) (string, error) {
	// This would integrate with the actual LLM client
	// For now, return a mock response
	return `{
		"overall_score": 85,
		"confidence": 0.92,
		"component_scores": {
			"security": 90,
			"quality": 80,
			"performance": 85
		},
		"issues": [],
		"recommendations": [],
		"security_findings": [],
		"quality_metrics": {},
		"performance_metrics": {}
	}`, nil
}

func (uve *UnifiedValidationEngine) mergeLLMResults(target, source *ValidationResult) {
	// Merge LLM analysis results with pattern-based results
	if source.OverallScore > 0 && source.Confidence >= uve.confidenceThreshold {
		// Weight the scores based on confidence
		llmWeight := source.Confidence
		patternWeight := 1.0 - llmWeight

		for component, llmScore := range source.ComponentScores {
			if existingScore, exists := target.ComponentScores[component]; exists {
				weighted := int(float64(existingScore)*patternWeight + float64(llmScore)*llmWeight)
				target.ComponentScores[component] = weighted
			} else {
				target.ComponentScores[component] = llmScore
			}
		}

		// Merge additional findings
		target.Issues = append(target.Issues, source.Issues...)
		target.SecurityFindings = append(target.SecurityFindings, source.SecurityFindings...)
		target.Recommendations = append(target.Recommendations, source.Recommendations...)
	}
}

func (uve *UnifiedValidationEngine) calculateSecurityScore(findings []SecurityFinding) int {
	baseScore := 100
	for _, finding := range findings {
		switch finding.Severity {
		case SeverityCritical:
			baseScore -= 25
		case SeverityHigh:
			baseScore -= 15
		case SeverityMedium:
			baseScore -= 8
		case SeverityLow:
			baseScore -= 3
		}
	}
	if baseScore < 0 {
		return 0
	}
	return baseScore
}

func (uve *UnifiedValidationEngine) calculateQualityScore(issues []Issue) int {
	baseScore := 100
	for _, issue := range issues {
		if issue.Type == IssueTypeQuality {
			switch issue.Severity {
			case SeverityHigh:
				baseScore -= 10
			case SeverityMedium:
				baseScore -= 5
			case SeverityLow:
				baseScore -= 2
			}
		}
	}
	if baseScore < 0 {
		return 0
	}
	return baseScore
}

func (uve *UnifiedValidationEngine) calculatePerformanceScore(result *ValidationResult) int {
	// Basic performance scoring - would be enhanced with actual metrics
	baseScore := 85

	if result.PerformanceMetrics != nil {
		if result.PerformanceMetrics.AlgorithmComplexity == "exponential" {
			baseScore -= 20
		}
		if result.PerformanceMetrics.EstimatedMemoryUsage > 1000 {
			baseScore -= 10
		}
	}

	return baseScore
}

func (uve *UnifiedValidationEngine) calculatePatternConfidence(result *ValidationResult) float64 {
	if len(result.SecurityFindings) == 0 && len(result.Issues) == 0 {
		return 0.5 // Moderate confidence when no patterns found
	}

	totalConfidence := 0.0
	count := 0

	for _, finding := range result.SecurityFindings {
		totalConfidence += finding.Confidence
		count++
	}

	if count == 0 {
		return 0.7 // Default confidence
	}

	return totalConfidence / float64(count)
}

func (uve *UnifiedValidationEngine) calculateCompletenessScore(result *ValidationResult) float64 {
	// Score based on how complete the analysis appears to be
	score := 0.5 // Base completeness

	if len(result.ComponentScores) > 0 {
		score += 0.2
	}
	if len(result.SecurityFindings) > 0 || len(result.Issues) > 0 {
		score += 0.2
	}
	if len(result.Recommendations) > 0 {
		score += 0.1
	}

	return score
}

func (uve *UnifiedValidationEngine) hasCriticalSecurityIssues(result *ValidationResult) bool {
	for _, finding := range result.SecurityFindings {
		if finding.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

func (uve *UnifiedValidationEngine) runValidationCycle(ctx context.Context, input *ValidationInput) (*ValidationResult, error) {
	result := &ValidationResult{
		ValidatorType: uve.GetValidatorType(),
		ValidatedAt:   time.Now(),
	}

	// Main validation logic
	if err := uve.performStaticAnalysis(ctx, input, result); err != nil {
		// Log error but continue with LLM analysis
		// In a real scenario, might want to handle this differently
	}

	if err := uve.performLLMAnalysis(ctx, input, result); err != nil {
		return nil, fmt.Errorf("llm analysis failed: %w", err)
	}

	// Finalize scoring
	// result.OverallScore = uve.scoringEngine.CalculateOverallScore(...)
	// result.Passed = result.OverallScore > uve.GetConfidenceThreshold()

	result.ValidationTime = time.Since(result.ValidatedAt)
	return result, nil
}

func (uve *UnifiedValidationEngine) performStaticAnalysis(ctx context.Context, input *ValidationInput, result *ValidationResult) error {
	// Step 1: Pattern-based analysis
	if err := uve.performPatternAnalysis(input, result); err != nil {
		uve.logger.Error("Pattern analysis failed", zap.Error(err))
		return fmt.Errorf("pattern analysis failed: %w", err)
	}

	// Step 3: Calculate final scores
	uve.calculateFinalScores(result)

	// Step 4: Apply scoring rules and penalties
	uve.applyValidationRules(result)

	return nil
}
