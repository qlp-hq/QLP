package core

import (
	"math"
	"sort"
)

// DefaultScoringEngine implements the ScoringEngine interface
type DefaultScoringEngine struct {
	defaultWeights map[ValidatorType]map[string]float64
	scoringRules   map[string][]ScoringRule
}

// NewDefaultScoringEngine creates a new scoring engine with default configurations
func NewDefaultScoringEngine() *DefaultScoringEngine {
	return &DefaultScoringEngine{
		defaultWeights: getDefaultWeights(),
		scoringRules:   getDefaultScoringRules(),
	}
}

// CalculateOverallScore computes weighted average of component scores
func (dse *DefaultScoringEngine) CalculateOverallScore(componentScores map[string]int, weights map[string]float64) int {
	if len(componentScores) == 0 {
		return 0
	}

	var totalScore float64
	var totalWeight float64

	for component, score := range componentScores {
		weight := weights[component]
		if weight == 0 {
			weight = 1.0 // Default weight if not specified
		}
		
		totalScore += float64(score) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	overallScore := int(math.Round(totalScore / totalWeight))
	
	// Ensure score is within valid range
	if overallScore < 0 {
		return 0
	}
	if overallScore > 100 {
		return 100
	}

	return overallScore
}

// CalculateComponentScore evaluates metrics against scoring rules
func (dse *DefaultScoringEngine) CalculateComponentScore(metrics map[string]interface{}, rules []ScoringRule) int {
	baseScore := 100
	var totalAdjustment float64

	for _, rule := range rules {
		if dse.evaluateCondition(metrics, rule.Condition) {
			adjustment := float64(rule.Points) * rule.Weight
			totalAdjustment += adjustment
		}
	}

	finalScore := int(math.Round(float64(baseScore) + totalAdjustment))
	
	// Clamp to valid range
	if finalScore < 0 {
		return 0
	}
	if finalScore > 100 {
		return 100
	}

	return finalScore
}

// GetDefaultWeights returns default scoring weights for validator types
func (dse *DefaultScoringEngine) GetDefaultWeights(validatorType ValidatorType) map[string]float64 {
	if weights, exists := dse.defaultWeights[validatorType]; exists {
		// Return a copy to prevent modification
		result := make(map[string]float64)
		for k, v := range weights {
			result[k] = v
		}
		return result
	}
	
	// Return balanced weights as fallback
	return map[string]float64{
		"syntax":      0.20,
		"security":    0.30,
		"quality":     0.25,
		"performance": 0.15,
		"compliance":  0.10,
	}
}

// ApplyPenalties reduces score based on penalty rules
func (dse *DefaultScoringEngine) ApplyPenalties(score int, penalties []Penalty) int {
	adjustedScore := float64(score)

	// Sort penalties by severity (apply most severe first)
	sortedPenalties := make([]Penalty, len(penalties))
	copy(sortedPenalties, penalties)
	sort.Slice(sortedPenalties, func(i, j int) bool {
		return sortedPenalties[i].Points > sortedPenalties[j].Points
	})

	for _, penalty := range sortedPenalties {
		if penalty.Percentage > 0 {
			// Percentage-based penalty
			adjustedScore *= (1.0 - penalty.Percentage)
		} else {
			// Point-based penalty
			adjustedScore -= float64(penalty.Points)
		}
		
		// Don't let score go below 0
		if adjustedScore < 0 {
			adjustedScore = 0
			break
		}
	}

	return int(math.Round(adjustedScore))
}

// CalculateConfidenceScore determines confidence based on multiple factors
func (dse *DefaultScoringEngine) CalculateConfidenceScore(factors map[string]float64) float64 {
	if len(factors) == 0 {
		return 0.0
	}

	// Weighted confidence calculation
	weights := map[string]float64{
		"llm_confidence":      0.35,
		"pattern_matches":     0.25,
		"static_analysis":     0.20,
		"completeness":        0.20,
	}

	var totalConfidence float64
	var totalWeight float64

	for factor, confidence := range factors {
		weight := weights[factor]
		if weight == 0 {
			weight = 0.1 // Small default weight for unknown factors
		}
		
		totalConfidence += confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	overallConfidence := totalConfidence / totalWeight
	
	// Clamp to valid range
	if overallConfidence < 0.0 {
		return 0.0
	}
	if overallConfidence > 1.0 {
		return 1.0
	}

	return overallConfidence
}

// CalculateRiskScore computes risk score based on security findings
func (dse *DefaultScoringEngine) CalculateRiskScore(findings []SecurityFinding) int {
	if len(findings) == 0 {
		return 0 // No risk if no findings
	}

	riskPoints := 0
	severityWeights := map[Severity]int{
		SeverityCritical: 25,
		SeverityHigh:     15,
		SeverityMedium:   8,
		SeverityLow:      3,
		SeverityInfo:     1,
	}

	for _, finding := range findings {
		weight := severityWeights[finding.Severity]
		// Apply confidence factor
		adjustedWeight := float64(weight) * finding.Confidence
		riskPoints += int(math.Round(adjustedWeight))
	}

	// Cap risk score at 100
	if riskPoints > 100 {
		return 100
	}

	return riskPoints
}

// Helper methods

func (dse *DefaultScoringEngine) evaluateCondition(metrics map[string]interface{}, condition map[string]interface{}) bool {
	for key, expectedValue := range condition {
		actualValue, exists := metrics[key]
		if !exists {
			return false
		}

		if !dse.compareValues(actualValue, expectedValue) {
			return false
		}
	}
	return true
}

func (dse *DefaultScoringEngine) compareValues(actual, expected interface{}) bool {
	switch expectedVal := expected.(type) {
	case string:
		if actualStr, ok := actual.(string); ok {
			return actualStr == expectedVal
		}
	case float64:
		if actualFloat, ok := actual.(float64); ok {
			return actualFloat >= expectedVal
		}
		if actualInt, ok := actual.(int); ok {
			return float64(actualInt) >= expectedVal
		}
	case int:
		if actualInt, ok := actual.(int); ok {
			return actualInt >= expectedVal
		}
		if actualFloat, ok := actual.(float64); ok {
			return actualFloat >= float64(expectedVal)
		}
	case bool:
		if actualBool, ok := actual.(bool); ok {
			return actualBool == expectedVal
		}
	case map[string]interface{}:
		// Handle complex conditions (e.g., ranges, operators)
		return dse.evaluateComplexCondition(actual, expectedVal)
	}
	return false
}

func (dse *DefaultScoringEngine) evaluateComplexCondition(actual interface{}, condition map[string]interface{}) bool {
	if operator, exists := condition["operator"]; exists {
		switch operator.(string) {
		case "gt":
			return dse.compareNumeric(actual, condition["value"], func(a, b float64) bool { return a > b })
		case "gte":
			return dse.compareNumeric(actual, condition["value"], func(a, b float64) bool { return a >= b })
		case "lt":
			return dse.compareNumeric(actual, condition["value"], func(a, b float64) bool { return a < b })
		case "lte":
			return dse.compareNumeric(actual, condition["value"], func(a, b float64) bool { return a <= b })
		case "eq":
			return dse.compareValues(actual, condition["value"])
		case "in":
			if values, ok := condition["values"].([]interface{}); ok {
				for _, value := range values {
					if dse.compareValues(actual, value) {
						return true
					}
				}
			}
			return false
		case "range":
			min := condition["min"]
			max := condition["max"]
			return dse.compareNumeric(actual, min, func(a, b float64) bool { return a >= b }) &&
				   dse.compareNumeric(actual, max, func(a, b float64) bool { return a <= b })
		}
	}
	return false
}

func (dse *DefaultScoringEngine) compareNumeric(actual, expected interface{}, compareFn func(float64, float64) bool) bool {
	actualFloat := dse.toFloat64(actual)
	expectedFloat := dse.toFloat64(expected)
	
	if actualFloat == nil || expectedFloat == nil {
		return false
	}
	
	return compareFn(*actualFloat, *expectedFloat)
}

func (dse *DefaultScoringEngine) toFloat64(value interface{}) *float64 {
	switch v := value.(type) {
	case float64:
		return &v
	case int:
		f := float64(v)
		return &f
	case string:
		// Could add string to float conversion if needed
		return nil
	default:
		return nil
	}
}

// Default configurations

func getDefaultWeights() map[ValidatorType]map[string]float64 {
	return map[ValidatorType]map[string]float64{
		ValidatorTypeUniversal: {
			"syntax":      0.15,
			"security":    0.30,
			"quality":     0.25,
			"performance": 0.20,
			"compliance":  0.10,
		},
		ValidatorTypeStatic: {
			"syntax":      0.40,
			"quality":     0.35,
			"performance": 0.25,
		},
		ValidatorTypeSecurity: {
			"security":    0.80,
			"compliance":  0.20,
		},
		ValidatorTypeDeployment: {
			"security":    0.25,
			"performance": 0.30,
			"reliability": 0.25,
			"scalability": 0.20,
		},
		ValidatorTypeEnterprise: {
			"security":    0.35,
			"compliance":  0.30,
			"quality":     0.20,
			"performance": 0.15,
		},
		ValidatorTypeSyntax: {
			"syntax":      0.100,
		},
	}
}

func getDefaultScoringRules() map[string][]ScoringRule {
	return map[string][]ScoringRule{
		"security": {
			{
				Name:   "high_severity_findings",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"severity": "high",
				},
				Points: -15,
			},
			{
				Name:   "critical_severity_findings",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"severity": "critical",
				},
				Points: -25,
			},
			{
				Name:   "low_confidence_findings",
				Weight: 0.5,
				Condition: map[string]interface{}{
					"confidence": map[string]interface{}{
						"operator": "lt",
						"value":    0.7,
					},
				},
				Points: -5,
			},
		},
		"quality": {
			{
				Name:   "high_complexity",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"complexity": map[string]interface{}{
						"operator": "gt",
						"value":    10,
					},
				},
				Points: -10,
			},
			{
				Name:   "low_test_coverage",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"test_coverage": map[string]interface{}{
						"operator": "lt",
						"value":    0.8,
					},
				},
				Points: -15,
			},
			{
				Name:   "high_duplication",
				Weight: 0.8,
				Condition: map[string]interface{}{
					"code_duplication": map[string]interface{}{
						"operator": "gt",
						"value":    0.1,
					},
				},
				Points: -8,
			},
		},
		"performance": {
			{
				Name:   "high_memory_usage",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"estimated_memory_usage": map[string]interface{}{
						"operator": "gt",
						"value":    1000, // MB
					},
				},
				Points: -12,
			},
			{
				Name:   "inefficient_algorithms",
				Weight: 1.0,
				Condition: map[string]interface{}{
					"algorithm_complexity": "exponential",
				},
				Points: -20,
			},
		},
	}
}