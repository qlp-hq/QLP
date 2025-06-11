package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Go Test Validator
type GoTestValidator struct {
	goValidator *GoSyntaxValidator
}

func NewGoTestValidator() *GoTestValidator {
	return &GoTestValidator{
		goValidator: NewGoSyntaxValidator(),
	}
}

func (gtv *GoTestValidator) Validate(ctx context.Context, output string) (*SyntaxValidationResult, error) {
	// First run standard Go validation
	result, err := gtv.goValidator.Validate(ctx, output)
	if err != nil {
		return nil, err
	}

	// Additional test-specific validation
	testIssues := gtv.validateTestSpecifics(output)
	result.Issues = append(result.Issues, testIssues...)

	if len(testIssues) > 0 {
		result.Score -= len(testIssues) * 5
		if result.Score < 0 {
			result.Score = 0
		}
	}

	return result, nil
}

func (gtv *GoTestValidator) validateTestSpecifics(output string) []string {
	var issues []string

	// Check for test functions
	if !strings.Contains(output, "func Test") {
		issues = append(issues, "No test functions found")
	}

	// Check for testing package import
	if !strings.Contains(output, "testing") {
		issues = append(issues, "Missing testing package import")
	}

	// Check for proper test function signature
	testFuncs := regexp.MustCompile(`func\s+(Test\w+)\s*\(\s*(\w+)\s+\*testing\.T\s*\)`).FindAllStringSubmatch(output, -1)
	if len(testFuncs) == 0 && strings.Contains(output, "func Test") {
		issues = append(issues, "Test functions should have signature func TestXxx(t *testing.T)")
	}

	return issues
}

// Terraform Validator
type TerraformValidator struct{}

func NewTerraformValidator() *TerraformValidator {
	return &TerraformValidator{}
}

func (tv *TerraformValidator) Validate(ctx context.Context, output string) (*SyntaxValidationResult, error) {
	result := &SyntaxValidationResult{
		Score:    100,
		Valid:    true,
		Issues:   []string{},
		Warnings: []string{},
	}

	// Extract HCL/Terraform blocks
	terraformBlocks := extractTerraformBlocks(output)
	if len(terraformBlocks) == 0 {
		result.Score = 50
		result.Warnings = append(result.Warnings, "No Terraform configuration blocks found")
		return result, nil
	}

	for i, block := range terraformBlocks {
		issues := tv.validateTerraformSyntax(block)
		if len(issues) > 0 {
			result.Score -= 15
			result.Valid = false
			for _, issue := range issues {
				result.Issues = append(result.Issues, fmt.Sprintf("Block %d: %s", i+1, issue))
			}
		}

		warnings := tv.checkTerraformBestPractices(block)
		result.Warnings = append(result.Warnings, warnings...)
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}

func (tv *TerraformValidator) validateTerraformSyntax(block string) []string {
	var issues []string

	// Check for basic Terraform syntax
	if !regexp.MustCompile(`(resource|data|variable|output|provider|module)\s+`).MatchString(block) {
		issues = append(issues, "No valid Terraform blocks found")
	}

	// Check for unmatched braces
	openBraces := strings.Count(block, "{")
	closeBraces := strings.Count(block, "}")
	if openBraces != closeBraces {
		issues = append(issues, fmt.Sprintf("Unmatched braces: %d open, %d close", openBraces, closeBraces))
	}

	// Check for proper resource syntax
	resourceMatches := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"\s*{`).FindAllStringSubmatch(block, -1)
	for _, match := range resourceMatches {
		if len(match) >= 3 {
			resourceType := match[1]
			resourceName := match[2]
			
			if !regexp.MustCompile(`^[a-z_]+$`).MatchString(resourceType) {
				issues = append(issues, fmt.Sprintf("Invalid resource type format: %s", resourceType))
			}
			
			if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(resourceName) {
				issues = append(issues, fmt.Sprintf("Invalid resource name format: %s", resourceName))
			}
		}
	}

	return issues
}

func (tv *TerraformValidator) checkTerraformBestPractices(block string) []string {
	var warnings []string

	// Check for hardcoded values
	if strings.Contains(block, `"123456"`) || strings.Contains(block, `"password"`) {
		warnings = append(warnings, "Avoid hardcoded credentials or sensitive values")
	}

	// Check for variable usage
	if strings.Contains(block, "resource ") && !strings.Contains(block, "var.") {
		warnings = append(warnings, "Consider using variables for configurable values")
	}

	// Check for tags
	if strings.Contains(block, "resource ") && !strings.Contains(block, "tags") {
		warnings = append(warnings, "Consider adding tags for resource management")
	}

	return warnings
}

func extractTerraformBlocks(text string) []string {
	// Extract terraform/hcl code blocks
	blocks := extractCodeBlocks(text, "hcl")
	if len(blocks) == 0 {
		blocks = extractCodeBlocks(text, "terraform")
	}
	if len(blocks) == 0 {
		blocks = extractCodeBlocks(text, "")
	}
	return blocks
}

// Markdown Validator
type MarkdownValidator struct{}

func NewMarkdownValidator() *MarkdownValidator {
	return &MarkdownValidator{}
}

func (mv *MarkdownValidator) Validate(ctx context.Context, output string) (*SyntaxValidationResult, error) {
	result := &SyntaxValidationResult{
		Score:    100,
		Valid:    true,
		Issues:   []string{},
		Warnings: []string{},
	}

	issues := mv.validateMarkdownStructure(output)
	result.Issues = append(result.Issues, issues...)

	warnings := mv.checkMarkdownBestPractices(output)
	result.Warnings = append(result.Warnings, warnings...)

	if len(issues) > 0 {
		result.Score -= len(issues) * 10
		result.Valid = false
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}

func (mv *MarkdownValidator) validateMarkdownStructure(output string) []string {
	var issues []string

	// Check for basic markdown elements
	if !strings.Contains(output, "#") {
		issues = append(issues, "No headers found - documentation should have structure")
	}

	// Check for unmatched markdown syntax
	codeBlocks := strings.Count(output, "```")
	if codeBlocks%2 != 0 {
		issues = append(issues, "Unmatched code block markers")
	}

	// Check for broken links
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	links := linkPattern.FindAllStringSubmatch(output, -1)
	for _, link := range links {
		if len(link) >= 3 {
			linkText := link[1]
			linkURL := link[2]
			
			if strings.TrimSpace(linkText) == "" {
				issues = append(issues, "Empty link text found")
			}
			
			if strings.TrimSpace(linkURL) == "" {
				issues = append(issues, "Empty link URL found")
			}
		}
	}

	return issues
}

func (mv *MarkdownValidator) checkMarkdownBestPractices(output string) []string {
	var warnings []string

	// Check for proper heading hierarchy
	lines := strings.Split(output, "\n")
	prevHeaderLevel := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, char := range line {
				if char == '#' {
					level++
				} else {
					break
				}
			}
			
			if prevHeaderLevel > 0 && level > prevHeaderLevel+1 {
				warnings = append(warnings, fmt.Sprintf("Line %d: Skipping header levels (h%d to h%d)", i+1, prevHeaderLevel, level))
			}
			prevHeaderLevel = level
		}
	}

	// Check for table of contents
	if strings.Count(output, "#") > 5 && !strings.Contains(strings.ToLower(output), "table of contents") {
		warnings = append(warnings, "Consider adding a table of contents for long documents")
	}

	// Check for code examples
	if strings.Contains(strings.ToLower(output), "api") && !strings.Contains(output, "```") {
		warnings = append(warnings, "API documentation should include code examples")
	}

	return warnings
}

// Analysis Validator
type AnalysisValidator struct{}

func NewAnalysisValidator() *AnalysisValidator {
	return &AnalysisValidator{}
}

func (av *AnalysisValidator) Validate(ctx context.Context, output string) (*SyntaxValidationResult, error) {
	result := &SyntaxValidationResult{
		Score:    100,
		Valid:    true,
		Issues:   []string{},
		Warnings: []string{},
	}

	issues := av.validateAnalysisContent(output)
	result.Issues = append(result.Issues, issues...)

	warnings := av.checkAnalysisQuality(output)
	result.Warnings = append(result.Warnings, warnings...)

	if len(issues) > 0 {
		result.Score -= len(issues) * 8
		result.Valid = false
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}

func (av *AnalysisValidator) validateAnalysisContent(output string) []string {
	var issues []string

	requiredSections := []string{"analysis", "findings", "results"}
	foundSections := 0
	
	outputLower := strings.ToLower(output)
	for _, section := range requiredSections {
		if strings.Contains(outputLower, section) {
			foundSections++
		}
	}

	if foundSections == 0 {
		issues = append(issues, "Analysis should contain findings or results section")
	}

	// Check for recommendations
	recommendationKeywords := []string{"recommend", "suggest", "should", "improve"}
	hasRecommendations := false
	for _, keyword := range recommendationKeywords {
		if strings.Contains(outputLower, keyword) {
			hasRecommendations = true
			break
		}
	}

	if !hasRecommendations {
		issues = append(issues, "Analysis should include recommendations or suggestions")
	}

	return issues
}

func (av *AnalysisValidator) checkAnalysisQuality(output string) []string {
	var warnings []string

	// Check for data/metrics
	if !regexp.MustCompile(`\d+%|\d+\.\d+|metrics|data|statistics`).MatchString(strings.ToLower(output)) {
		warnings = append(warnings, "Consider including quantitative data or metrics")
	}

	// Check for structured format
	if !strings.Contains(output, "##") && !strings.Contains(output, "1.") && !strings.Contains(output, "-") {
		warnings = append(warnings, "Consider using structured formatting (headers, lists, etc.)")
	}

	// Check length - analysis should be comprehensive
	if len(output) < 500 {
		warnings = append(warnings, "Analysis might be too brief - consider more detailed examination")
	}

	return warnings
}