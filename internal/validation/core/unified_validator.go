package core

import (
	"context"
	"fmt"
	"time"

	"QLP/internal/llm"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// UnifiedValidationEngine combines all core components into a single validator
type UnifiedValidationEngine struct {
	llmIntegration  LLMIntegration
	scoringEngine   ScoringEngine
	patternEngine   PatternEngine
	logger          logger.Interface
	validatorType   ValidatorType
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
	startTime := time.Now()
	
	uve.logger.Info("Starting unified validation",
		zap.String("validator_type", string(uve.validatorType)),
		zap.String("language", input.Language),
		zap.Int("files_count", len(input.Content)),
	)

	result := &ValidationResult{
		ValidatorType:      uve.validatorType,
		ValidatedAt:        startTime,
		ComponentScores:    make(map[string]int),
		Issues:             []Issue{},
		Warnings:           []Warning{},
		Recommendations:    []Recommendation{},
		SecurityFindings:   []SecurityFinding{},
		BestPractices:      []BestPractice{},
	}

	// Step 1: Pattern-based analysis
	if err := uve.performPatternAnalysis(input, result); err != nil {
		uve.logger.Error("Pattern analysis failed", zap.Error(err))
		return nil, fmt.Errorf("pattern analysis failed: %w", err)
	}

	// Step 2: LLM-powered analysis
	if err := uve.performLLMAnalysis(ctx, input, result); err != nil {
		uve.logger.Warn("LLM analysis failed, continuing with pattern-only results", zap.Error(err))
	}

	// Step 3: Calculate final scores
	uve.calculateFinalScores(result)

	// Step 4: Apply scoring rules and penalties
	uve.applyValidationRules(result)

	result.ValidationTime = time.Since(startTime)
	result.Passed = uve.determineOverallPass(result)

	uve.logger.Info("Unified validation completed",
		zap.Int("overall_score", result.OverallScore),
		zap.Float64("confidence", result.Confidence),
		zap.Bool("passed", result.Passed),
		zap.Duration("duration", result.ValidationTime),
	)

	return result, nil
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
							FilePath:  filePath,
							Line:      match.Location.Line,
							Column:    match.Location.Column,
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
							FilePath:  filePath,
							Line:      match.Location.Line,
							Column:    match.Location.Column,
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
	prompt := uve.llmIntegration.GeneratePrompt(input, PromptTypeAnalysis)

	// Make LLM request (this would interface with actual LLM client)
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