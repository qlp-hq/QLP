package validators

import (
	"context"
	"regexp"
	"strings"

	"QLP/services/validation-service/pkg/contracts"
)

type SyntaxValidator interface {
	Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error)
	GetSupportedLanguages() []string
}

type SyntaxValidatorRegistry struct {
	validators map[string]SyntaxValidator
}

func NewSyntaxValidatorRegistry() *SyntaxValidatorRegistry {
	registry := &SyntaxValidatorRegistry{
		validators: make(map[string]SyntaxValidator),
	}
	
	// Register built-in validators
	registry.Register("go", NewGoSyntaxValidator())
	registry.Register("python", NewPythonSyntaxValidator())
	registry.Register("javascript", NewJavaScriptSyntaxValidator())
	registry.Register("typescript", NewTypeScriptSyntaxValidator())
	registry.Register("hcl", NewHCLSyntaxValidator())
	registry.Register("terraform", NewHCLSyntaxValidator()) // Alias
	registry.Register("markdown", NewMarkdownSyntaxValidator())
	registry.Register("yaml", NewYAMLSyntaxValidator())
	registry.Register("json", NewJSONSyntaxValidator())
	
	return registry
}

func (r *SyntaxValidatorRegistry) Register(language string, validator SyntaxValidator) {
	r.validators[strings.ToLower(language)] = validator
}

func (r *SyntaxValidatorRegistry) GetValidator(language string) SyntaxValidator {
	return r.validators[strings.ToLower(language)]
}

func (r *SyntaxValidatorRegistry) GetSupportedLanguages() []string {
	var languages []string
	for lang := range r.validators {
		languages = append(languages, lang)
	}
	return languages
}

// Go Syntax Validator
type GoSyntaxValidator struct{}

func NewGoSyntaxValidator() *GoSyntaxValidator {
	return &GoSyntaxValidator{}
}

func (v *GoSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	// Basic Go syntax checks
	issues := v.checkGoSyntax(content)
	result.Issues = append(result.Issues, issues...)
	
	// Deduct score for issues
	if len(issues) > 0 {
		result.Score -= len(issues) * 10
		result.Valid = false
	}
	
	// Add warnings for best practices
	warnings := v.checkGoBestPractices(content)
	result.Warnings = append(result.Warnings, warnings...)
	
	if result.Score < 0 {
		result.Score = 0
	}
	
	return result, nil
}

func (v *GoSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"go", "golang"}
}

func (v *GoSyntaxValidator) checkGoSyntax(content string) []contracts.SyntaxIssue {
	var issues []contracts.SyntaxIssue
	lines := strings.Split(content, "\n")
	
	hasPackage := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineNum := i + 1
		
		// Check for package declaration
		if strings.HasPrefix(trimmed, "package ") {
			hasPackage = true
		}
		
		// Check for unmatched braces (simple check)
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		if openBraces != closeBraces && (openBraces > 1 || closeBraces > 1) {
			issues = append(issues, contracts.SyntaxIssue{
				Type:     "syntax",
				Severity: contracts.SeverityError,
				Message:  "Potential unmatched braces",
				Line:     lineNum,
				Rule:     "go-braces",
			})
		}
		
		// Check for missing semicolons (Go doesn't require them, but check for C-style habits)
		if strings.Contains(trimmed, ";") && !strings.Contains(trimmed, "for") {
			issues = append(issues, contracts.SyntaxIssue{
				Type:     "style",
				Severity: contracts.SeverityWarning,
				Message:  "Unnecessary semicolon - Go doesn't require semicolons",
				Line:     lineNum,
				Rule:     "go-semicolon",
				Suggestion: "Remove the semicolon",
			})
		}
	}
	
	if !hasPackage && len(content) > 10 {
		issues = append(issues, contracts.SyntaxIssue{
			Type:     "syntax",
			Severity: contracts.SeverityError,
			Message:  "Missing package declaration",
			Line:     1,
			Rule:     "go-package",
			Suggestion: "Add 'package main' or appropriate package name",
		})
	}
	
	return issues
}

func (v *GoSyntaxValidator) checkGoBestPractices(content string) []contracts.SyntaxIssue {
	var warnings []contracts.SyntaxIssue
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineNum := i + 1
		
		// Check for exported functions without comments
		if matched, _ := regexp.MatchString(`^func\s+[A-Z]\w*`, trimmed); matched {
			// Check if previous line has a comment
			if i == 0 || !strings.HasPrefix(strings.TrimSpace(lines[i-1]), "//") {
				warnings = append(warnings, contracts.SyntaxIssue{
					Type:     "documentation",
					Severity: contracts.SeverityWarning,
					Message:  "Exported function should have a comment",
					Line:     lineNum,
					Rule:     "go-exported-comment",
					Suggestion: "Add a comment starting with the function name",
				})
			}
		}
		
		// Check for error handling
		if strings.Contains(trimmed, "err") && !strings.Contains(trimmed, "if err != nil") {
			if matched, _ := regexp.MatchString(`\w+,\s*err\s*:=`, trimmed); matched {
				warnings = append(warnings, contracts.SyntaxIssue{
					Type:     "error-handling",
					Severity: contracts.SeverityWarning,
					Message:  "Error returned but not checked",
					Line:     lineNum,
					Rule:     "go-error-check",
					Suggestion: "Add proper error handling",
				})
			}
		}
	}
	
	return warnings
}

// Python Syntax Validator
type PythonSyntaxValidator struct{}

func NewPythonSyntaxValidator() *PythonSyntaxValidator {
	return &PythonSyntaxValidator{}
}

func (v *PythonSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	// Basic Python syntax checks
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Check indentation
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		if spaces%4 != 0 && spaces > 0 {
			result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
				Type:     "style",
				Severity: contracts.SeverityWarning,
				Message:  "Inconsistent indentation - use 4 spaces",
				Line:     lineNum,
				Rule:     "python-indent",
			})
		}
		
		// Check for missing colons
		if matched, _ := regexp.MatchString(`(if|for|while|def|class|try|except|finally|with)\s+.*[^:]$`, trimmed); matched {
			result.Issues = append(result.Issues, contracts.SyntaxIssue{
				Type:     "syntax",
				Severity: contracts.SeverityError,
				Message:  "Missing colon after control statement",
				Line:     lineNum,
				Rule:     "python-colon",
				Suggestion: "Add ':' at the end of the line",
			})
			result.Score -= 10
			result.Valid = false
		}
	}
	
	if result.Score < 0 {
		result.Score = 0
	}
	
	return result, nil
}

func (v *PythonSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"python", "py"}
}

// JavaScript Syntax Validator
type JavaScriptSyntaxValidator struct{}

func NewJavaScriptSyntaxValidator() *JavaScriptSyntaxValidator {
	return &JavaScriptSyntaxValidator{}
}

func (v *JavaScriptSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	// Basic JavaScript syntax checks
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		// Check for var usage (prefer let/const)
		if strings.Contains(trimmed, "var ") {
			result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
				Type:     "style",
				Severity: contracts.SeverityWarning,
				Message:  "Consider using 'let' or 'const' instead of 'var'",
				Line:     lineNum,
				Rule:     "js-no-var",
				Suggestion: "Use 'let' for variables or 'const' for constants",
			})
		}
		
		// Check for missing semicolons
		if matched, _ := regexp.MatchString(`[^{};]\s*$`, trimmed); matched && 
			!strings.HasSuffix(trimmed, "{") && 
			!strings.HasSuffix(trimmed, "}") &&
			!strings.Contains(trimmed, "if") &&
			!strings.Contains(trimmed, "for") &&
			!strings.Contains(trimmed, "while") {
			result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
				Type:     "style",
				Severity: contracts.SeverityWarning,
				Message:  "Missing semicolon",
				Line:     lineNum,
				Rule:     "js-semicolon",
				Suggestion: "Add semicolon at the end of the statement",
			})
		}
	}
	
	return result, nil
}

func (v *JavaScriptSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"javascript", "js", "node"}
}

// TypeScript Syntax Validator
type TypeScriptSyntaxValidator struct{}

func NewTypeScriptSyntaxValidator() *TypeScriptSyntaxValidator {
	return &TypeScriptSyntaxValidator{}
}

func (v *TypeScriptSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	// Reuse JavaScript validator and add TypeScript-specific checks
	jsValidator := NewJavaScriptSyntaxValidator()
	result, err := jsValidator.Validate(ctx, content, language)
	if err != nil {
		return nil, err
	}
	
	result.Language = language
	
	// Add TypeScript-specific checks
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		// Check for missing type annotations on function parameters
		if matched, _ := regexp.MatchString(`function\s+\w+\s*\([^)]*\w+\s*[,)][^:]*\)`, trimmed); matched {
			result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
				Type:     "typescript",
				Severity: contracts.SeverityWarning,
				Message:  "Consider adding type annotations to function parameters",
				Line:     lineNum,
				Rule:     "ts-param-types",
				Suggestion: "Add type annotations for better type safety",
			})
		}
	}
	
	return result, nil
}

func (v *TypeScriptSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"typescript", "ts"}
}

// HCL/Terraform Syntax Validator
type HCLSyntaxValidator struct{}

func NewHCLSyntaxValidator() *HCLSyntaxValidator {
	return &HCLSyntaxValidator{}
}

func (v *HCLSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Check for resource/data blocks
		if matched, _ := regexp.MatchString(`^(resource|data)\s+"[^"]+"\s+"[^"]+"\s*{?`, trimmed); matched {
			if !strings.Contains(trimmed, "{") {
				result.Issues = append(result.Issues, contracts.SyntaxIssue{
					Type:     "syntax",
					Severity: contracts.SeverityError,
					Message:  "Missing opening brace for resource block",
					Line:     lineNum,
					Rule:     "hcl-block-brace",
					Suggestion: "Add '{' at the end of the resource declaration",
				})
				result.Score -= 10
				result.Valid = false
			}
		}
		
		// Check for unquoted strings in assignments
		if matched, _ := regexp.MatchString(`\w+\s*=\s*[^"'\[\{]\w+`, trimmed); matched {
			result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
				Type:     "style",
				Severity: contracts.SeverityWarning,
				Message:  "String values should be quoted",
				Line:     lineNum,
				Rule:     "hcl-quotes",
				Suggestion: "Wrap string values in double quotes",
			})
		}
	}
	
	if result.Score < 0 {
		result.Score = 0
	}
	
	return result, nil
}

func (v *HCLSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"hcl", "terraform"}
}

// Markdown Syntax Validator
type MarkdownSyntaxValidator struct{}

func NewMarkdownSyntaxValidator() *MarkdownSyntaxValidator {
	return &MarkdownSyntaxValidator{}
}

func (v *MarkdownSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		// Check for malformed links
		if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]") {
			if matched, _ := regexp.MatchString(`\[[^\]]*\]\([^)]*\)`, trimmed); !matched {
				if strings.Contains(trimmed, "](") {
					result.Issues = append(result.Issues, contracts.SyntaxIssue{
						Type:     "syntax",
						Severity: contracts.SeverityError,
						Message:  "Malformed markdown link",
						Line:     lineNum,
						Rule:     "md-link",
						Suggestion: "Check link syntax: [text](url)",
					})
					result.Score -= 5
				}
			}
		}
		
		// Check for inconsistent heading styles
		if strings.HasPrefix(trimmed, "#") {
			if matched, _ := regexp.MatchString(`^#+\s+`, trimmed); !matched {
				result.Warnings = append(result.Warnings, contracts.SyntaxIssue{
					Type:     "style",
					Severity: contracts.SeverityWarning,
					Message:  "Heading should have space after #",
					Line:     lineNum,
					Rule:     "md-heading-space",
					Suggestion: "Add space after # in headings",
				})
			}
		}
	}
	
	if result.Score < 0 {
		result.Score = 0
	}
	
	return result, nil
}

func (v *MarkdownSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"markdown", "md"}
}

// YAML Syntax Validator
type YAMLSyntaxValidator struct{}

func NewYAMLSyntaxValidator() *YAMLSyntaxValidator {
	return &YAMLSyntaxValidator{}
}

func (v *YAMLSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    95, // YAML is generally more forgiving
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	// Basic YAML validation would go here
	// For now, just check for tabs (YAML requires spaces)
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		lineNum := i + 1
		
		if strings.Contains(line, "\t") {
			result.Issues = append(result.Issues, contracts.SyntaxIssue{
				Type:     "syntax",
				Severity: contracts.SeverityError,
				Message:  "YAML does not allow tabs for indentation",
				Line:     lineNum,
				Rule:     "yaml-no-tabs",
				Suggestion: "Use spaces instead of tabs",
			})
			result.Score -= 10
			result.Valid = false
		}
	}
	
	return result, nil
}

func (v *YAMLSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"yaml", "yml"}
}

// JSON Syntax Validator
type JSONSyntaxValidator struct{}

func NewJSONSyntaxValidator() *JSONSyntaxValidator {
	return &JSONSyntaxValidator{}
}

func (v *JSONSyntaxValidator) Validate(ctx context.Context, content, language string) (*contracts.SyntaxResult, error) {
	result := &contracts.SyntaxResult{
		Score:    100,
		Valid:    true,
		Language: language,
		Issues:   []contracts.SyntaxIssue{},
		Warnings: []contracts.SyntaxIssue{},
	}
	
	// Basic JSON validation
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return result, nil
	}
	
	// Check basic JSON structure
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		result.Issues = append(result.Issues, contracts.SyntaxIssue{
			Type:     "syntax",
			Severity: contracts.SeverityError,
			Message:  "JSON must start with { or [",
			Line:     1,
			Rule:     "json-start",
		})
		result.Valid = false
		result.Score = 0
	}
	
	return result, nil
}

func (v *JSONSyntaxValidator) GetSupportedLanguages() []string {
	return []string{"json"}
}