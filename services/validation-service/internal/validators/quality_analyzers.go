package validators

import (
	"context"
	"strings"

	"QLP/services/validation-service/pkg/contracts"
)

type QualityAnalyzer interface {
	Analyze(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.QualityResult, error)
	GetSupportedLanguages() []string
}

type QualityAnalyzerRegistry struct {
	analyzers map[string]QualityAnalyzer
}

func NewQualityAnalyzerRegistry() *QualityAnalyzerRegistry {
	registry := &QualityAnalyzerRegistry{
		analyzers: make(map[string]QualityAnalyzer),
	}
	
	// Register built-in analyzers
	registry.Register("fast", NewFastQualityAnalyzer())
	registry.Register("standard", NewStandardQualityAnalyzer())
	registry.Register("comprehensive", NewComprehensiveQualityAnalyzer())
	
	return registry
}

func (r *QualityAnalyzerRegistry) Register(name string, analyzer QualityAnalyzer) {
	r.analyzers[strings.ToLower(name)] = analyzer
}

func (r *QualityAnalyzerRegistry) GetAnalyzer(name string) QualityAnalyzer {
	return r.analyzers[strings.ToLower(name)]
}

// Fast Quality Analyzer (from existing validation logic)
type FastQualityAnalyzer struct{}

func NewFastQualityAnalyzer() *FastQualityAnalyzer {
	return &FastQualityAnalyzer{}
}

func (a *FastQualityAnalyzer) Analyze(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.QualityResult, error) {
	result := &contracts.QualityResult{
		Score:           85, // Start with good baseline
		Maintainability: 85,
		Documentation:   85,
		BestPractices:   85,
		TestCoverage:    0.0,
		Issues:          []contracts.QualityIssue{},
		Suggestions:     []contracts.QualitySuggestion{},
		Passed:          true,
	}

	// Apply task-type specific analysis
	switch taskType {
	case contracts.TaskTypeCodegen:
		a.analyzeCodeQuality(content, language, result)
	case contracts.TaskTypeTest:
		a.analyzeTestQuality(content, language, result)
	case contracts.TaskTypeInfra:
		a.analyzeInfraQuality(content, language, result)
	case contracts.TaskTypeDoc:
		a.analyzeDocQuality(content, language, result)
	case contracts.TaskTypeAnalyze:
		a.analyzeAnalysisQuality(content, language, result)
	default:
		a.analyzeGenericQuality(content, language, result)
	}

	result.Passed = result.Score >= 70
	return result, nil
}

func (a *FastQualityAnalyzer) GetSupportedLanguages() []string {
	return []string{"go", "python", "javascript", "typescript", "hcl", "markdown", "yaml", "json"}
}

func (a *FastQualityAnalyzer) analyzeCodeQuality(content, language string, result *contracts.QualityResult) {
	lines := strings.Split(content, "\n")
	
	switch language {
	case "go":
		a.analyzeGoCode(content, lines, result)
	case "python":
		a.analyzePythonCode(content, lines, result)
	case "javascript", "typescript":
		a.analyzeJSCode(content, lines, result)
	default:
		a.analyzeGenericCode(content, lines, result)
	}
}

func (a *FastQualityAnalyzer) analyzeGoCode(content string, lines []string, result *contracts.QualityResult) {
	// Check for basic Go structure
	if !strings.Contains(content, "package ") {
		result.Score -= 15
		result.BestPractices -= 20
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "structure",
			Severity: contracts.SeverityError,
			Message:  "Missing package declaration",
			Rule:     "go-package-required",
			Category: "structure",
		})
	}
	
	if !strings.Contains(content, "func ") {
		result.Score -= 20
		result.Maintainability -= 25
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "structure",
			Severity: contracts.SeverityWarning,
			Message:  "No functions defined",
			Rule:     "go-functions",
			Category: "structure",
		})
	}
	
	// Check for error handling
	if strings.Contains(content, "if err != nil") || strings.Contains(content, "error") {
		result.Score += 5
		result.BestPractices += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "best-practice",
			Message: "Good error handling practices detected",
			Impact:  "positive",
			Effort:  "none",
		})
	} else if strings.Contains(content, "err") {
		result.Score -= 10
		result.BestPractices -= 15
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "error-handling",
			Severity: contracts.SeverityWarning,
			Message:  "Error variables found but not properly handled",
			Rule:     "go-error-check",
			Category: "error-handling",
		})
	}
	
	// Check for comments
	commentCount := strings.Count(content, "//")
	functionCount := strings.Count(content, "func ")
	if functionCount > 0 {
		commentRatio := float64(commentCount) / float64(functionCount)
		if commentRatio < 0.3 {
			result.Documentation -= 20
			result.Score -= 10
			result.Issues = append(result.Issues, contracts.QualityIssue{
				Type:     "documentation",
				Severity: contracts.SeverityWarning,
				Message:  "Insufficient code documentation",
				Rule:     "go-comments",
				Category: "documentation",
			})
		} else if commentRatio > 0.7 {
			result.Documentation += 10
			result.Score += 5
		}
	}
	
	// Check for imports (suggests proper organization)
	if strings.Contains(content, "import") {
		result.BestPractices += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "structure",
			Message: "Good import organization",
			Impact:  "positive",
			Effort:  "none",
		})
	}
	
	// Calculate complexity metrics
	result.Complexity = &contracts.ComplexityMetrics{
		Cyclomatic:      a.calculateCyclomaticComplexity(content),
		LinesOfCode:     len(strings.Split(content, "\n")),
		Maintainability: float64(result.Maintainability),
	}
}

func (a *FastQualityAnalyzer) analyzePythonCode(content string, lines []string, result *contracts.QualityResult) {
	// Check for proper structure
	hasMain := strings.Contains(content, "def main(") || strings.Contains(content, "if __name__ == \"__main__\"")
	if !hasMain && len(lines) > 20 {
		result.Score -= 10
		result.BestPractices -= 15
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "structure",
			Severity: contracts.SeverityWarning,
			Message:  "Consider adding a main function for better structure",
			Rule:     "python-main",
			Category: "structure",
		})
	}
	
	// Check for docstrings
	docstringCount := strings.Count(content, `"""`) + strings.Count(content, `'''`)
	functionCount := strings.Count(content, "def ")
	if functionCount > 0 && docstringCount < functionCount {
		result.Documentation -= 15
		result.Score -= 8
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "documentation",
			Severity: contracts.SeverityWarning,
			Message:  "Functions missing docstrings",
			Rule:     "python-docstrings",
			Category: "documentation",
		})
	}
	
	// Check for proper imports
	if strings.Contains(content, "import") {
		result.BestPractices += 5
	}
}

func (a *FastQualityAnalyzer) analyzeJSCode(content string, lines []string, result *contracts.QualityResult) {
	// Check for modern JavaScript practices
	if strings.Contains(content, "const ") || strings.Contains(content, "let ") {
		result.BestPractices += 10
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "modern-js",
			Message: "Good use of modern variable declarations",
			Impact:  "positive",
			Effort:  "none",
		})
	}
	
	if strings.Contains(content, "var ") {
		result.BestPractices -= 10
		result.Score -= 5
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "style",
			Severity: contracts.SeverityWarning,
			Message:  "Consider using 'let' or 'const' instead of 'var'",
			Rule:     "js-no-var",
			Category: "modern-practices",
		})
	}
	
	// Check for arrow functions
	if strings.Contains(content, "=>") {
		result.BestPractices += 5
	}
	
	// Check for proper error handling
	if strings.Contains(content, "try") && strings.Contains(content, "catch") {
		result.BestPractices += 10
	}
}

func (a *FastQualityAnalyzer) analyzeGenericCode(content string, lines []string, result *contracts.QualityResult) {
	// Basic checks for any code
	if len(content) < 50 {
		result.Score -= 20
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "content",
			Severity: contracts.SeverityWarning,
			Message:  "Content appears too short for meaningful analysis",
			Rule:     "min-content",
			Category: "completeness",
		})
	}
	
	// Check for structured content
	if strings.Contains(content, "\n") && len(lines) > 3 {
		result.Score += 5
		result.Maintainability += 5
	}
}

func (a *FastQualityAnalyzer) analyzeTestQuality(content, language string, result *contracts.QualityResult) {
	// Test-specific quality checks
	testCount := strings.Count(content, "func Test") + strings.Count(content, "test_") + strings.Count(content, "it(") + strings.Count(content, "describe(")
	
	if testCount == 0 {
		result.Score -= 30
		result.BestPractices -= 40
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "test-coverage",
			Severity: contracts.SeverityError,
			Message:  "No test functions found",
			Rule:     "test-required",
			Category: "testing",
		})
	} else if testCount >= 3 {
		result.Score += 10
		result.BestPractices += 15
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "testing",
			Message: "Good test coverage with multiple test cases",
			Impact:  "positive",
			Effort:  "none",
		})
	}
	
	// Check for assertions
	assertionKeywords := []string{"assert", "expect", "should", "t.Error", "t.Fatal"}
	hasAssertions := false
	for _, keyword := range assertionKeywords {
		if strings.Contains(content, keyword) {
			hasAssertions = true
			break
		}
	}
	
	if hasAssertions {
		result.BestPractices += 10
		result.Score += 5
	} else if testCount > 0 {
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "test-quality",
			Severity: contracts.SeverityWarning,
			Message:  "Tests found but no assertions detected",
			Rule:     "test-assertions",
			Category: "testing",
		})
	}
	
	// Check for table-driven tests (Go specific)
	if strings.Contains(content, "for _, tc := range") || strings.Contains(content, "for _, test := range") {
		result.BestPractices += 15
		result.Score += 10
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "testing",
			Message: "Excellent use of table-driven tests",
			Impact:  "high",
			Effort:  "none",
		})
	}
	
	// Calculate test coverage estimate
	allFunctions := strings.Count(content, "func ")
	if allFunctions > 0 {
		result.TestCoverage = float64(testCount) / float64(allFunctions)
		if result.TestCoverage > 1.0 {
			result.TestCoverage = 1.0
		}
	}
}

func (a *FastQualityAnalyzer) analyzeInfraQuality(content, language string, result *contracts.QualityResult) {
	// Infrastructure code quality checks
	if strings.Contains(content, "resource ") || strings.Contains(content, "data ") {
		result.BestPractices += 10
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "infrastructure",
			Message: "Good use of Terraform resources",
			Impact:  "positive",
			Effort:  "none",
		})
	} else {
		result.Score -= 20
		result.BestPractices -= 25
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "structure",
			Severity: contracts.SeverityError,
			Message:  "No infrastructure resources defined",
			Rule:     "infra-resources",
			Category: "infrastructure",
		})
	}
	
	// Check for provider configuration
	if strings.Contains(content, "provider ") {
		result.BestPractices += 5
	}
	
	// Check for variables
	if strings.Contains(content, "variable ") {
		result.BestPractices += 8
		result.Score += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "infrastructure",
			Message: "Good use of variables for parameterization",
			Impact:  "medium",
			Effort:  "none",
		})
	}
	
	// Check for outputs
	if strings.Contains(content, "output ") {
		result.BestPractices += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "infrastructure",
			Message: "Good practice defining outputs",
			Impact:  "medium",
			Effort:  "none",
		})
	}
}

func (a *FastQualityAnalyzer) analyzeDocQuality(content, language string, result *contracts.QualityResult) {
	// Documentation quality checks
	headerCount := strings.Count(content, "#")
	if headerCount == 0 {
		result.Score -= 15
		result.Documentation -= 20
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "structure",
			Severity: contracts.SeverityWarning,
			Message:  "No headers found - consider adding structure",
			Rule:     "doc-headers",
			Category: "documentation",
		})
	} else if headerCount >= 3 {
		result.Documentation += 10
		result.Score += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "documentation",
			Message: "Good document structure with multiple headers",
			Impact:  "positive",
			Effort:  "none",
		})
	}
	
	// Check for examples
	if strings.Contains(content, "```") || strings.Contains(content, "example") {
		result.Documentation += 10
		result.Score += 8
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "documentation",
			Message: "Good use of examples in documentation",
			Impact:  "high",
			Effort:  "none",
		})
	}
	
	// Check for clear instructions
	instructionWords := []string{"install", "setup", "configure", "run", "usage", "how to"}
	hasInstructions := false
	contentLower := strings.ToLower(content)
	for _, word := range instructionWords {
		if strings.Contains(contentLower, word) {
			hasInstructions = true
			break
		}
	}
	
	if hasInstructions {
		result.Documentation += 15
		result.Score += 10
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "documentation",
			Message: "Good instructional content",
			Impact:  "high",
			Effort:  "none",
		})
	} else {
		result.Documentation -= 15
		result.Score -= 10
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "completeness",
			Severity: contracts.SeverityWarning,
			Message:  "Missing clear usage instructions",
			Rule:     "doc-instructions",
			Category: "documentation",
		})
	}
}

func (a *FastQualityAnalyzer) analyzeAnalysisQuality(content, language string, result *contracts.QualityResult) {
	// Analysis quality checks
	analysisKeywords := []string{"analysis", "findings", "results", "metrics", "performance", "summary"}
	hasAnalysis := false
	contentLower := strings.ToLower(content)
	for _, keyword := range analysisKeywords {
		if strings.Contains(contentLower, keyword) {
			hasAnalysis = true
			break
		}
	}
	
	if !hasAnalysis {
		result.Score -= 20
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "content",
			Severity: contracts.SeverityWarning,
			Message:  "Missing analysis keywords",
			Rule:     "analysis-content",
			Category: "analysis",
		})
	}
	
	// Check for recommendations
	recKeywords := []string{"recommend", "suggest", "improve", "optimize", "should", "consider"}
	hasRecommendations := false
	for _, keyword := range recKeywords {
		if strings.Contains(contentLower, keyword) {
			hasRecommendations = true
			break
		}
	}
	
	if hasRecommendations {
		result.Score += 10
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "analysis",
			Message: "Good inclusion of recommendations",
			Impact:  "high",
			Effort:  "none",
		})
	} else {
		result.Score -= 10
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "completeness",
			Severity: contracts.SeverityWarning,
			Message:  "Missing recommendations or suggestions",
			Rule:     "analysis-recommendations",
			Category: "analysis",
		})
	}
}

func (a *FastQualityAnalyzer) analyzeGenericQuality(content, language string, result *contracts.QualityResult) {
	// Generic quality checks for unknown content types
	if len(content) < 50 {
		result.Score -= 20
		result.Issues = append(result.Issues, contracts.QualityIssue{
			Type:     "completeness",
			Severity: contracts.SeverityWarning,
			Message:  "Content appears incomplete",
			Rule:     "min-content",
			Category: "completeness",
		})
	} else if len(content) > 1000 {
		result.Score += 5
		result.Suggestions = append(result.Suggestions, contracts.QualitySuggestion{
			Type:    "completeness",
			Message: "Good content length",
			Impact:  "positive",
			Effort:  "none",
		})
	}
	
	// Check for structured content
	if strings.Contains(content, "\n") {
		result.Score += 5
		result.Maintainability += 5
	}
}

func (a *FastQualityAnalyzer) calculateCyclomaticComplexity(content string) int {
	// Simple cyclomatic complexity calculation
	complexity := 1 // Base complexity
	
	// Count decision points
	decisionKeywords := []string{"if ", "else", "for ", "while ", "switch ", "case ", "&&", "||", "?"}
	for _, keyword := range decisionKeywords {
		complexity += strings.Count(content, keyword)
	}
	
	return complexity
}

// Standard Quality Analyzer (placeholder for more comprehensive analysis)
type StandardQualityAnalyzer struct {
	fastAnalyzer *FastQualityAnalyzer
}

func NewStandardQualityAnalyzer() *StandardQualityAnalyzer {
	return &StandardQualityAnalyzer{
		fastAnalyzer: NewFastQualityAnalyzer(),
	}
}

func (a *StandardQualityAnalyzer) Analyze(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.QualityResult, error) {
	// Start with fast analysis
	result, err := a.fastAnalyzer.Analyze(ctx, content, language, taskType)
	if err != nil {
		return nil, err
	}
	
	// Add standard-level checks
	// TODO: Implement more sophisticated analysis
	
	return result, nil
}

func (a *StandardQualityAnalyzer) GetSupportedLanguages() []string {
	return a.fastAnalyzer.GetSupportedLanguages()
}

// Comprehensive Quality Analyzer (placeholder for advanced analysis)
type ComprehensiveQualityAnalyzer struct {
	standardAnalyzer *StandardQualityAnalyzer
}

func NewComprehensiveQualityAnalyzer() *ComprehensiveQualityAnalyzer {
	return &ComprehensiveQualityAnalyzer{
		standardAnalyzer: NewStandardQualityAnalyzer(),
	}
}

func (a *ComprehensiveQualityAnalyzer) Analyze(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.QualityResult, error) {
	// Start with standard analysis
	result, err := a.standardAnalyzer.Analyze(ctx, content, language, taskType)
	if err != nil {
		return nil, err
	}
	
	// Add comprehensive-level checks
	// TODO: Implement advanced analysis (AST parsing, deep metrics, etc.)
	
	return result, nil
}

func (a *ComprehensiveQualityAnalyzer) GetSupportedLanguages() []string {
	return a.standardAnalyzer.GetSupportedLanguages()
}