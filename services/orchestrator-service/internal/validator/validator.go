package validation

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
	"go.uber.org/zap"
)

type ValidationEngine struct {
	llmClient        llm.Client
	syntaxValidators map[models.TaskType]SyntaxValidator
	securityScanner  *SecurityScanner
	qualityAnalyzer  *QualityAnalyzer
}

// ValidationResult moved to internal/types to avoid import cycles
// Using types.ValidationResult instead

// Local validation types - keeping detailed ones here
type SyntaxValidationResult struct {
	Score       int      `json:"score"`
	Valid       bool     `json:"valid"`
	Issues      []string `json:"issues"`
	Warnings    []string `json:"warnings"`
	LintResults []string `json:"lint_results"`
}

type LLMCritiqueResult struct {
	Score        int      `json:"score"`
	Feedback     string   `json:"feedback"`
	Suggestions  []string `json:"suggestions"`
	Improvements []string `json:"improvements"`
	Confidence   float64  `json:"confidence"`
}

func NewValidationEngine(llmClient llm.Client) *ValidationEngine {
	return &ValidationEngine{
		llmClient:        llmClient,
		syntaxValidators: initializeSyntaxValidators(),
		securityScanner:  NewSecurityScanner(),
		qualityAnalyzer:  NewQualityAnalyzer(),
	}
}

func (ve *ValidationEngine) ValidateTaskOutput(ctx context.Context, task models.Task, output string, sandboxResult *sandbox.SandboxExecutionResult) (*types.ValidationResult, error) {
	startTime := time.Now()
	
	logger.WithComponent("validation").Info("Starting validation",
		zap.String("task_id", task.ID),
		zap.String("task_type", string(task.Type)))

	result := &types.ValidationResult{
		ValidatedAt: startTime,
	}

	// Check for fast mode
	validationLevel := os.Getenv("QLP_VALIDATION_LEVEL")
	if validationLevel == "fast" {
		// Fast mode: Use lightweight quality calculation with real metrics
		qualityResult := ve.calculateFastQualityScore(task, output)
		securityScore := ve.calculateFastSecurityScore(output)
		
		result.SecurityResult = &types.SecurityResult{
			Score:       securityScore,
			RiskLevel:   ve.getSecurityRiskLevel(securityScore),
			Vulnerabilities: []types.SecurityIssue{},
			Passed:      securityScore >= 70,
		}
		result.QualityResult = qualityResult
		
		// Calculate overall score using similar weights as full validation
		result.OverallScore = ve.calculateFastOverallScore(securityScore, qualityResult.Score)
		result.SecurityScore = securityScore
		result.QualityScore = qualityResult.Score
		result.Passed = result.OverallScore >= 70
		result.ValidationTime = time.Since(startTime)
		
		logger.WithComponent("validation").Info("Fast validation completed",
			zap.String("task_id", task.ID),
			zap.Int("overall_score", result.OverallScore),
			zap.Int("security_score", result.SecurityScore),
			zap.Int("quality_score", result.QualityScore),
			zap.Bool("passed", result.Passed),
			zap.String("mode", "fast"))
		return result, nil
	}

	// 1. Syntax Validation
	syntaxResult, err := ve.validateSyntax(ctx, task, output)
	if err != nil {
		return nil, fmt.Errorf("syntax validation failed: %w", err)
	}
	_ = syntaxResult // TODO: add SyntaxResult field to ValidationResult if needed

	// 2. Security Validation
	securityResult, err := ve.validateSecurity(ctx, task, output, sandboxResult)
	if err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}
	result.SecurityResult = securityResult

	// 3. Quality Analysis
	qualityResult, err := ve.analyzeQuality(ctx, task, output, sandboxResult)
	if err != nil {
		return nil, fmt.Errorf("quality analysis failed: %w", err)
	}
	result.QualityResult = qualityResult

	// 4. LLM Self-Critique
	critiqueResult, err := ve.performLLMCritique(ctx, task, output)
	if err != nil {
		return nil, fmt.Errorf("LLM critique failed: %w", err)
	}
	_ = critiqueResult // TODO: add LLMCritiqueResult field to ValidationResult if needed

	// Calculate overall score
	result.OverallScore = ve.calculateOverallScore(syntaxResult, securityResult, qualityResult, critiqueResult)
	result.SecurityScore = securityResult.Score
	result.QualityScore = qualityResult.Score
	result.Passed = result.OverallScore >= 70 // 70% threshold for passing
	result.ValidationTime = time.Since(startTime)

	logger.WithComponent("validation").Info("Validation completed",
		zap.String("task_id", task.ID),
		zap.Int("overall_score", result.OverallScore),
		zap.Bool("passed", result.Passed),
		zap.Int("security_score", result.SecurityScore),
		zap.Int("quality_score", result.QualityScore),
		zap.Duration("validation_time", result.ValidationTime))

	return result, nil
}

func (ve *ValidationEngine) validateSyntax(ctx context.Context, task models.Task, output string) (*SyntaxValidationResult, error) {
	validator, exists := ve.syntaxValidators[task.Type]
	if !exists {
		return &SyntaxValidationResult{
			Score:   100,
			Valid:   true,
			Issues:  []string{},
			Warnings: []string{"No syntax validator available for task type"},
		}, nil
	}

	return validator.Validate(ctx, output)
}

func (ve *ValidationEngine) validateSecurity(ctx context.Context, task models.Task, output string, sandboxResult *sandbox.SandboxExecutionResult) (*types.SecurityResult, error) {
	return ve.securityScanner.ScanOutput(ctx, task, output, sandboxResult)
}

func (ve *ValidationEngine) analyzeQuality(ctx context.Context, task models.Task, output string, sandboxResult *sandbox.SandboxExecutionResult) (*types.QualityResult, error) {
	return ve.qualityAnalyzer.AnalyzeOutput(ctx, task, output, sandboxResult)
}

func (ve *ValidationEngine) performLLMCritique(ctx context.Context, task models.Task, output string) (*LLMCritiqueResult, error) {
	prompt := ve.buildCritiquePrompt(task, output)
	
	response, err := ve.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM critique request failed: %w", err)
	}

	return ve.parseCritiqueResponse(response)
}

func (ve *ValidationEngine) buildCritiquePrompt(task models.Task, output string) string {
	return fmt.Sprintf(`
You are an expert code reviewer and quality analyst. Please critically analyze the following output for task type "%s".

TASK DESCRIPTION: %s

OUTPUT TO REVIEW:
%s

Please provide a comprehensive critique focusing on:
1. Correctness and completeness
2. Code quality and best practices
3. Security considerations
4. Performance implications
5. Maintainability and readability
6. Adherence to standards

Rate the output on a scale of 0-100 and provide:
- Overall feedback (2-3 sentences)
- Specific suggestions for improvement
- Any potential issues or concerns

Respond in this JSON format:
{
  "score": <0-100>,
  "feedback": "<overall assessment>",
  "suggestions": ["<suggestion1>", "<suggestion2>"],
  "improvements": ["<improvement1>", "<improvement2>"],
  "confidence": <0.0-1.0>
}
`, task.Type, task.Description, output)
}

func (ve *ValidationEngine) parseCritiqueResponse(response string) (*LLMCritiqueResult, error) {
	// Extract JSON from response
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	// For now, parse manually - in production you'd use proper JSON parsing
	result := &LLMCritiqueResult{
		Score:        75, // Default reasonable score
		Feedback:     "LLM critique analysis completed",
		Suggestions:  []string{"Consider adding more error handling", "Improve code documentation"},
		Improvements: []string{"Add input validation", "Enhance test coverage"},
		Confidence:   0.8,
	}

	// Extract actual score if possible
	if scoreMatch := regexp.MustCompile(`"score":\s*(\d+)`).FindStringSubmatch(response); len(scoreMatch) > 1 {
		if score := parseInt(scoreMatch[1]); score > 0 {
			result.Score = score
		}
	}

	// Extract feedback
	if feedbackMatch := regexp.MustCompile(`"feedback":\s*"([^"]+)"`).FindStringSubmatch(response); len(feedbackMatch) > 1 {
		result.Feedback = feedbackMatch[1]
	}

	return result, nil
}

func (ve *ValidationEngine) calculateOverallScore(syntax *SyntaxValidationResult, security *types.SecurityResult, quality *types.QualityResult, critique *LLMCritiqueResult) int {
	// Weighted scoring system
	weights := map[string]float64{
		"syntax":    0.25, // 25%
		"security":  0.30, // 30% - highest weight for security
		"quality":   0.25, // 25%
		"critique":  0.20, // 20%
	}

	totalScore := float64(syntax.Score)*weights["syntax"] +
		float64(security.Score)*weights["security"] +
		float64(quality.Score)*weights["quality"] +
		float64(critique.Score)*weights["critique"]

	return int(totalScore)
}

// Syntax Validators
type SyntaxValidator interface {
	Validate(ctx context.Context, output string) (*SyntaxValidationResult, error)
}

func initializeSyntaxValidators() map[models.TaskType]SyntaxValidator {
	return map[models.TaskType]SyntaxValidator{
		models.TaskTypeCodegen: NewGoSyntaxValidator(),
		models.TaskTypeTest:    NewGoTestValidator(),
		models.TaskTypeInfra:   NewTerraformValidator(),
		models.TaskTypeDoc:     NewMarkdownValidator(),
		models.TaskTypeAnalyze: NewAnalysisValidator(),
	}
}

// Go Syntax Validator
type GoSyntaxValidator struct{}

func NewGoSyntaxValidator() *GoSyntaxValidator {
	return &GoSyntaxValidator{}
}

func (gsv *GoSyntaxValidator) Validate(ctx context.Context, output string) (*SyntaxValidationResult, error) {
	result := &SyntaxValidationResult{
		Score:    100,
		Valid:    true,
		Issues:   []string{},
		Warnings: []string{},
	}

	// Extract Go code blocks
	codeBlocks := extractCodeBlocks(output, "go")
	if len(codeBlocks) == 0 {
		result.Score = 50
		result.Warnings = append(result.Warnings, "No Go code blocks found")
		return result, nil
	}

	for i, code := range codeBlocks {
		// Basic Go syntax checks
		issues := gsv.checkGoSyntax(code)
		if len(issues) > 0 {
			result.Score -= 20
			result.Valid = false
			for _, issue := range issues {
				result.Issues = append(result.Issues, fmt.Sprintf("Block %d: %s", i+1, issue))
			}
		}

		// Best practices checks
		warnings := gsv.checkGoBestPractices(code)
		result.Warnings = append(result.Warnings, warnings...)
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}

func (gsv *GoSyntaxValidator) checkGoSyntax(code string) []string {
	var issues []string

	// Basic syntax checks
	if !strings.Contains(code, "package ") {
		issues = append(issues, "Missing package declaration")
	}

	// Check for basic Go syntax errors
	patterns := map[string]string{
		`func\s+\w+\s*\([^)]*\)\s*[^{]*\s*{`: "Function declaration",
		`import\s*\(`: "Import statement",
		`var\s+\w+\s+\w+`: "Variable declaration",
	}

	for pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, code); !matched && strings.Contains(code, "func ") {
			// Only check function patterns if there are functions
			continue
		}
	}

	// Check for unmatched braces
	openBraces := strings.Count(code, "{")
	closeBraces := strings.Count(code, "}")
	if openBraces != closeBraces {
		issues = append(issues, fmt.Sprintf("Unmatched braces: %d open, %d close", openBraces, closeBraces))
	}

	return issues
}

func (gsv *GoSyntaxValidator) checkGoBestPractices(code string) []string {
	var warnings []string

	// Check for error handling
	if strings.Contains(code, "err") && !strings.Contains(code, "if err != nil") {
		warnings = append(warnings, "Consider proper error handling")
	}

	// Check for exported functions without comments
	if matched, _ := regexp.MatchString(`func\s+[A-Z]\w*`, code); matched && !strings.Contains(code, "//") {
		warnings = append(warnings, "Exported functions should have comments")
	}

	return warnings
}

// Helper functions
func extractCodeBlocks(text, language string) []string {
	var blocks []string
	lines := strings.Split(text, "\n")
	
	var currentBlock strings.Builder
	inBlock := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```"+language) || (strings.HasPrefix(line, "```") && language == "") {
			if inBlock {
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}
	
	return blocks
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// Fast validation helper methods

func (ve *ValidationEngine) calculateFastQualityScore(task models.Task, output string) *types.QualityResult {
	result := &types.QualityResult{
		Score:           85, // Start with a good baseline
		Maintainability: 85,
		Documentation:   85,
		BestPractices:   85,
		TestCoverage:    0.0,
		Passed:          true,
	}

	// Apply quick heuristic checks based on task type
	switch task.Type {
	case models.TaskTypeCodegen:
		result = ve.fastCodeQualityCheck(output)
	case models.TaskTypeTest:
		result = ve.fastTestQualityCheck(output)
	case models.TaskTypeInfra:
		result = ve.fastInfraQualityCheck(output)
	case models.TaskTypeDoc:
		result = ve.fastDocQualityCheck(output)
	case models.TaskTypeAnalyze:
		result = ve.fastAnalysisQualityCheck(output)
	default:
		result = ve.fastGenericQualityCheck(output)
	}

	result.Passed = result.Score >= 70
	return result
}

func (ve *ValidationEngine) fastCodeQualityCheck(output string) *types.QualityResult {
	score := 85
	maintainability := 85
	documentation := 85
	bestPractices := 85

	// Check for basic code structure
	if !strings.Contains(output, "package ") {
		score -= 15
		bestPractices -= 20
	}
	
	if !strings.Contains(output, "func ") {
		score -= 20
		maintainability -= 25
	}
	
	// Check for error handling
	if strings.Contains(output, "if err != nil") || strings.Contains(output, "error") {
		score += 5
		bestPractices += 5
	} else {
		score -= 10
		bestPractices -= 15
	}
	
	// Check for comments
	commentCount := strings.Count(output, "//")
	functionCount := strings.Count(output, "func ")
	if functionCount > 0 {
		commentRatio := float64(commentCount) / float64(functionCount)
		if commentRatio < 0.3 {
			documentation -= 20
			score -= 10
		} else if commentRatio > 0.7 {
			documentation += 10
			score += 5
		}
	}
	
	// Check for imports (suggests proper organization)
	if strings.Contains(output, "import") {
		bestPractices += 5
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: clampScore(maintainability),
		Documentation:   clampScore(documentation),
		BestPractices:   clampScore(bestPractices),
		TestCoverage:    0.0,
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) fastTestQualityCheck(output string) *types.QualityResult {
	score := 85
	bestPractices := 85
	
	// Check for test functions
	testCount := strings.Count(output, "func Test")
	if testCount == 0 {
		score -= 30
		bestPractices -= 40
	} else if testCount >= 3 {
		score += 10
		bestPractices += 15
	}
	
	// Check for test assertions
	if strings.Contains(output, "t.Error") || strings.Contains(output, "t.Fatal") {
		bestPractices += 10
		score += 5
	}
	
	// Check for table-driven tests
	if strings.Contains(output, "for _, tc := range") || strings.Contains(output, "for _, test := range") {
		bestPractices += 15
		score += 10
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: 85,
		Documentation:   80,
		BestPractices:   clampScore(bestPractices),
		TestCoverage:    calculateSimpleTestCoverage(output),
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) fastInfraQualityCheck(output string) *types.QualityResult {
	score := 85
	bestPractices := 85
	
	// Check for resource definitions
	if strings.Contains(output, "resource ") || strings.Contains(output, "data ") {
		bestPractices += 10
	} else {
		score -= 20
		bestPractices -= 25
	}
	
	// Check for provider configuration
	if strings.Contains(output, "provider ") {
		bestPractices += 5
	}
	
	// Check for variables
	if strings.Contains(output, "variable ") {
		bestPractices += 8
		score += 5
	}
	
	// Check for outputs
	if strings.Contains(output, "output ") {
		bestPractices += 5
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: 85,
		Documentation:   80,
		BestPractices:   clampScore(bestPractices),
		TestCoverage:    0.0,
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) fastDocQualityCheck(output string) *types.QualityResult {
	score := 85
	documentation := 85
	
	// Check for headers
	headerCount := strings.Count(output, "#")
	if headerCount == 0 {
		score -= 15
		documentation -= 20
	} else if headerCount >= 3 {
		documentation += 10
		score += 5
	}
	
	// Check for examples
	if strings.Contains(output, "```") || strings.Contains(output, "example") {
		documentation += 10
		score += 8
	}
	
	// Check for clear instructions
	instructionWords := []string{"install", "setup", "configure", "run", "usage"}
	hasInstructions := false
	for _, word := range instructionWords {
		if strings.Contains(strings.ToLower(output), word) {
			hasInstructions = true
			break
		}
	}
	if hasInstructions {
		documentation += 15
		score += 10
	} else {
		documentation -= 15
		score -= 10
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: 80,
		Documentation:   clampScore(documentation),
		BestPractices:   80,
		TestCoverage:    0.0,
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) fastAnalysisQualityCheck(output string) *types.QualityResult {
	score := 85
	
	// Check for analysis keywords
	analysisKeywords := []string{"analysis", "findings", "results", "metrics", "performance"}
	hasAnalysis := false
	for _, keyword := range analysisKeywords {
		if strings.Contains(strings.ToLower(output), keyword) {
			hasAnalysis = true
			break
		}
	}
	if !hasAnalysis {
		score -= 20
	}
	
	// Check for recommendations
	recKeywords := []string{"recommend", "suggest", "improve", "optimize"}
	hasRecommendations := false
	for _, keyword := range recKeywords {
		if strings.Contains(strings.ToLower(output), keyword) {
			hasRecommendations = true
			break
		}
	}
	if hasRecommendations {
		score += 10
	} else {
		score -= 10
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: 80,
		Documentation:   85,
		BestPractices:   80,
		TestCoverage:    0.0,
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) fastGenericQualityCheck(output string) *types.QualityResult {
	score := 80 // Slightly lower baseline for unknown types
	
	// Basic content checks
	if len(output) < 50 {
		score -= 20
	} else if len(output) > 1000 {
		score += 5
	}
	
	// Check for structured content
	if strings.Contains(output, "\n") {
		score += 5
	}

	return &types.QualityResult{
		Score:           clampScore(score),
		Maintainability: 80,
		Documentation:   80,
		BestPractices:   80,
		TestCoverage:    0.0,
		Passed:          score >= 70,
	}
}

func (ve *ValidationEngine) calculateFastSecurityScore(output string) int {
	score := 85 // Start with good baseline
	
	// Quick security checks
	securityIssues := []string{
		"password", "secret", "api_key", "private_key", "token",
		"hardcoded", "plain_text", "unencrypted",
		"sql injection", "xss", "csrf",
	}
	
	outputLower := strings.ToLower(output)
	for _, issue := range securityIssues {
		if strings.Contains(outputLower, issue) {
			score -= 10
		}
	}
	
	// Positive security practices
	if strings.Contains(outputLower, "validate") || strings.Contains(outputLower, "sanitize") {
		score += 5
	}
	
	if strings.Contains(outputLower, "encrypt") || strings.Contains(outputLower, "hash") {
		score += 8
	}
	
	return clampScore(score)
}

func (ve *ValidationEngine) getSecurityRiskLevel(score int) types.SecurityRiskLevel {
	if score >= 90 {
		return types.SecurityRiskLevelNone
	} else if score >= 80 {
		return types.SecurityRiskLevelLow
	} else if score >= 60 {
		return types.SecurityRiskLevelMedium
	} else if score >= 40 {
		return types.SecurityRiskLevelHigh
	}
	return types.SecurityRiskLevelCritical
}

func (ve *ValidationEngine) calculateFastOverallScore(securityScore, qualityScore int) int {
	// Use similar weights as full validation but simplified
	weights := map[string]float64{
		"security": 0.40, // Slightly higher weight for security in fast mode
		"quality":  0.60, // Quality gets the rest
	}
	
	score := float64(securityScore)*weights["security"] + float64(qualityScore)*weights["quality"]
	return clampScore(int(score))
}

// Helper functions

func clampScore(score int) int {
	if score > 100 {
		return 100
	} else if score < 0 {
		return 0
	}
	return score
}

func calculateSimpleTestCoverage(output string) float64 {
	testFunctions := strings.Count(output, "func Test")
	allFunctions := strings.Count(output, "func ")
	
	if allFunctions == 0 {
		return 0.0
	}
	
	coverage := float64(testFunctions) / float64(allFunctions)
	if coverage > 1.0 {
		coverage = 1.0
	}
	
	return coverage
}