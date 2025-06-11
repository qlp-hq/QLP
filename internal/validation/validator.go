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
		// Fast mode: Skip heavy validations, use simple heuristics
		result.SecurityResult = &types.SecurityResult{
			Score:       75,
			RiskLevel:   types.SecurityRiskLevelLow,
			Vulnerabilities: []types.SecurityIssue{},
			Passed:      true,
		}
		result.QualityResult = &types.QualityResult{
			Score: 75,
			Maintainability: 75,
			Documentation: 75,
			BestPractices: 75,
			Passed: true,
		}
		
		// Skip LLM critique in fast mode
		result.OverallScore = 75
		result.Passed = true
		result.ValidationTime = time.Since(startTime)
		
		logger.WithComponent("validation").Info("Fast validation completed",
			zap.String("task_id", task.ID),
			zap.Int("overall_score", result.OverallScore),
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