package validation

import (
	"context"
	"regexp"
	"strings"

	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
)

type QualityAnalyzer struct {
	metrics map[models.TaskType]QualityMetrics
}

type QualityMetrics struct {
	CompletenessChecks    []QualityCheck
	MaintainabilityChecks []QualityCheck
	PerformanceChecks     []QualityCheck
	BestPracticesChecks   []QualityCheck
	DocumentationChecks   []QualityCheck
}

type QualityCheck struct {
	Name        string
	Description string
	Check       func(string) (bool, int) // returns (passed, score_impact)
	Weight      float64
}

func NewQualityAnalyzer() *QualityAnalyzer {
	return &QualityAnalyzer{
		metrics: initializeQualityMetrics(),
	}
}

func (qa *QualityAnalyzer) AnalyzeOutput(ctx context.Context, task models.Task, output string, sandboxResult *sandbox.SandboxExecutionResult) (*types.QualityResult, error) {
	result := &types.QualityResult{
		Score:           100,
		Maintainability: 100,
		Documentation:   100,
		BestPractices:   100,
		TestCoverage:    0.0,
		Passed:          true,
	}

	// Local variables for detailed analysis
	completeness := 100
	performance := 100

	metrics, exists := qa.metrics[task.Type]
	if !exists {
		// Use generic metrics for unknown task types
		metrics = qa.metrics[models.TaskTypeCodegen]
	}

	// Analyze completeness
	completeness = qa.analyzeCompleteness(output, metrics.CompletenessChecks)

	// Analyze maintainability
	result.Maintainability = qa.analyzeMaintainability(output, metrics.MaintainabilityChecks)

	// Analyze performance
	performance = qa.analyzePerformance(output, sandboxResult, metrics.PerformanceChecks)

	// Analyze best practices
	result.BestPractices = qa.analyzeBestPractices(output, metrics.BestPracticesChecks)

	// Analyze documentation
	result.Documentation = qa.analyzeDocumentation(output, metrics.DocumentationChecks)

	// Calculate test coverage for code tasks
	if task.Type == models.TaskTypeCodegen || task.Type == models.TaskTypeTest {
		result.TestCoverage = qa.calculateTestCoverage(output)
	}

	// Calculate overall score with weights
	result.Score = qa.calculateOverallQualityScore(completeness, performance, result)
	result.Passed = result.Score >= 70

	return result, nil
}

func (qa *QualityAnalyzer) analyzeCompleteness(output string, checks []QualityCheck) int {
	return qa.runQualityChecks(output, checks)
}

func (qa *QualityAnalyzer) analyzeMaintainability(output string, checks []QualityCheck) int {
	return qa.runQualityChecks(output, checks)
}

func (qa *QualityAnalyzer) analyzePerformance(output string, sandboxResult *sandbox.SandboxExecutionResult, checks []QualityCheck) int {
	score := qa.runQualityChecks(output, checks)

	// Factor in sandbox performance metrics
	if sandboxResult != nil {
		for _, result := range sandboxResult.Results {
			if result.Metrics != nil {
				// Deduct points for high resource usage
				if result.Metrics.CPUUsagePercent > 80 {
					score -= 10
				}
				if result.Metrics.MemoryUsageBytes > 1024*1024*1024 { // > 1GB
					score -= 10
				}
				if result.Duration.Seconds() > 60 { // > 1 minute
					score -= 5
				}
			}
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (qa *QualityAnalyzer) analyzeBestPractices(output string, checks []QualityCheck) int {
	return qa.runQualityChecks(output, checks)
}

func (qa *QualityAnalyzer) analyzeDocumentation(output string, checks []QualityCheck) int {
	return qa.runQualityChecks(output, checks)
}

func (qa *QualityAnalyzer) runQualityChecks(output string, checks []QualityCheck) int {
	totalScore := 100.0
	totalWeight := 0.0

	for _, check := range checks {
		passed, impact := check.Check(output)
		totalWeight += check.Weight

		if !passed {
			totalScore -= float64(impact) * check.Weight
		}
	}

	if totalWeight > 0 {
		score := int(totalScore * (1.0 / totalWeight))
		if score < 0 {
			score = 0
		}
		return score
	}

	return 100
}

func (qa *QualityAnalyzer) calculateTestCoverage(output string) float64 {
	// Simple heuristic: count test functions vs regular functions
	testFunctions := len(regexp.MustCompile(`func\s+Test\w+`).FindAllString(output, -1))
	allFunctions := len(regexp.MustCompile(`func\s+\w+`).FindAllString(output, -1))

	if allFunctions == 0 {
		return 0.0
	}

	coverage := float64(testFunctions) / float64(allFunctions)
	if coverage > 1.0 {
		coverage = 1.0
	}

	return coverage
}

func (qa *QualityAnalyzer) calculateOverallQualityScore(completeness, performance int, result *types.QualityResult) int {
	weights := map[string]float64{
		"completeness":    0.25,
		"maintainability": 0.20,
		"performance":     0.15,
		"best_practices":  0.25,
		"documentation":   0.15,
	}

	score := float64(completeness)*weights["completeness"] +
		float64(result.Maintainability)*weights["maintainability"] +
		float64(performance)*weights["performance"] +
		float64(result.BestPractices)*weights["best_practices"] +
		float64(result.Documentation)*weights["documentation"]

	// Bonus for good test coverage
	if result.TestCoverage > 0.8 {
		score += 5
	} else if result.TestCoverage > 0.6 {
		score += 3
	}

	return int(score)
}

func initializeQualityMetrics() map[models.TaskType]QualityMetrics {
	return map[models.TaskType]QualityMetrics{
		models.TaskTypeCodegen: {
			CompletenessChecks: []QualityCheck{
				{
					Name:        "Has Main Function",
					Description: "Code should have a main function or entry point",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "func main") || strings.Contains(output, "package main"), 15
					},
					Weight: 1.0,
				},
				{
					Name:        "Has Error Handling",
					Description: "Code should include proper error handling",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "if err != nil") || strings.Contains(output, "error"), 10
					},
					Weight: 0.8,
				},
				{
					Name:        "Has Input Validation",
					Description: "Code should validate inputs",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "validate") || strings.Contains(output, "check"), 8
					},
					Weight: 0.6,
				},
			},
			MaintainabilityChecks: []QualityCheck{
				{
					Name:        "Function Length",
					Description: "Functions should be reasonably sized",
					Check: func(output string) (bool, int) {
						functions := regexp.MustCompile(`func\s+\w+[^}]*{[^}]*}`).FindAllString(output, -1)
						for _, fn := range functions {
							if strings.Count(fn, "\n") > 50 {
								return false, 10
							}
						}
						return true, 0
					},
					Weight: 0.7,
				},
				{
					Name:        "Variable Naming",
					Description: "Variables should have descriptive names",
					Check: func(output string) (bool, int) {
						badNames := regexp.MustCompile(`\b(a|b|c|x|y|z|foo|bar)\s*:=`).FindAllString(output, -1)
						return len(badNames) < 3, len(badNames) * 5
					},
					Weight: 0.5,
				},
			},
			PerformanceChecks: []QualityCheck{
				{
					Name:        "Efficient Data Structures",
					Description: "Should use appropriate data structures",
					Check: func(output string) (bool, int) {
						hasSlices := strings.Contains(output, "[]")
						hasMaps := strings.Contains(output, "map[")
						return hasSlices || hasMaps, 8
					},
					Weight: 0.6,
				},
				{
					Name:        "Memory Management",
					Description: "Should avoid obvious memory leaks",
					Check: func(output string) (bool, int) {
						hasDefer := strings.Contains(output, "defer")
						hasClose := strings.Contains(output, ".Close()")
						return hasDefer || hasClose, 5
					},
					Weight: 0.4,
				},
			},
			BestPracticesChecks: []QualityCheck{
				{
					Name:        "Package Declaration",
					Description: "Should have proper package declaration",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "package "), 15
					},
					Weight: 1.0,
				},
				{
					Name:        "Imports Organization",
					Description: "Should organize imports properly",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "import"), 5
					},
					Weight: 0.3,
				},
				{
					Name:        "Constants Usage",
					Description: "Should use constants for magic numbers",
					Check: func(output string) (bool, int) {
						magicNumbers := regexp.MustCompile(`\b\d{2,}\b`).FindAllString(output, -1)
						return len(magicNumbers) < 5, len(magicNumbers) * 2
					},
					Weight: 0.4,
				},
			},
			DocumentationChecks: []QualityCheck{
				{
					Name:        "Function Comments",
					Description: "Functions should have comments",
					Check: func(output string) (bool, int) {
						functions := strings.Count(output, "func ")
						comments := strings.Count(output, "//")
						if functions == 0 {
							return true, 0
						}
						ratio := float64(comments) / float64(functions)
						return ratio >= 0.5, int((0.5-ratio)*20)
					},
					Weight: 0.8,
				},
				{
					Name:        "Package Documentation", 
					Description: "Package should have documentation",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "// Package"), 10
					},
					Weight: 0.5,
				},
			},
		},
		models.TaskTypeTest: {
			CompletenessChecks: []QualityCheck{
				{
					Name:        "Has Test Functions",
					Description: "Should contain test functions",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "func Test"), 20
					},
					Weight: 1.0,
				},
				{
					Name:        "Test Coverage",
					Description: "Should test main functionality",
					Check: func(output string) (bool, int) {
						testCount := strings.Count(output, "func Test")
						return testCount >= 3, (3-testCount)*5
					},
					Weight: 0.8,
				},
			},
			BestPracticesChecks: []QualityCheck{
				{
					Name:        "Table Driven Tests",
					Description: "Should use table-driven tests where appropriate",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "for _, tc := range") || strings.Contains(output, "for _, test := range"), 5
					},
					Weight: 0.6,
				},
				{
					Name:        "Test Assertions",
					Description: "Should use proper test assertions",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "t.Error") || strings.Contains(output, "t.Fatal"), 8
					},
					Weight: 0.8,
				},
			},
		},
		models.TaskTypeInfra: {
			CompletenessChecks: []QualityCheck{
				{
					Name:        "Resource Definitions",
					Description: "Should define infrastructure resources",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "resource ") || strings.Contains(output, "data "), 15
					},
					Weight: 1.0,
				},
				{
					Name:        "Provider Configuration",
					Description: "Should configure providers",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "provider "), 10
					},
					Weight: 0.8,
				},
			},
			BestPracticesChecks: []QualityCheck{
				{
					Name:        "Variable Usage",
					Description: "Should use variables for configuration",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "variable "), 8
					},
					Weight: 0.7,
				},
				{
					Name:        "Output Values",
					Description: "Should define outputs",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "output "), 5
					},
					Weight: 0.5,
				},
			},
		},
		models.TaskTypeDoc: {
			CompletenessChecks: []QualityCheck{
				{
					Name:        "Has Headers",
					Description: "Documentation should have clear structure",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "#"), 10
					},
					Weight: 0.8,
				},
				{
					Name:        "Has Examples",
					Description: "Should include examples",
					Check: func(output string) (bool, int) {
						return strings.Contains(output, "```") || strings.Contains(output, "example"), 8
					},
					Weight: 0.7,
				},
			},
			DocumentationChecks: []QualityCheck{
				{
					Name:        "Clear Instructions",
					Description: "Should provide clear instructions",
					Check: func(output string) (bool, int) {
						instructionWords := []string{"install", "setup", "configure", "run", "usage"}
						for _, word := range instructionWords {
							if strings.Contains(strings.ToLower(output), word) {
								return true, 0
							}
						}
						return false, 10
					},
					Weight: 1.0,
				},
			},
		},
		models.TaskTypeAnalyze: {
			CompletenessChecks: []QualityCheck{
				{
					Name:        "Has Analysis Results",
					Description: "Should provide analysis findings",
					Check: func(output string) (bool, int) {
						keywords := []string{"analysis", "findings", "results", "metrics", "performance"}
						for _, keyword := range keywords {
							if strings.Contains(strings.ToLower(output), keyword) {
								return true, 0
							}
						}
						return false, 15
					},
					Weight: 1.0,
				},
				{
					Name:        "Has Recommendations",
					Description: "Should provide actionable recommendations",
					Check: func(output string) (bool, int) {
						keywords := []string{"recommend", "suggest", "improve", "optimize"}
						for _, keyword := range keywords {
							if strings.Contains(strings.ToLower(output), keyword) {
								return true, 0
							}
						}
						return false, 10
					},
					Weight: 0.8,
				},
			},
		},
	}
}