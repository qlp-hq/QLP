package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"QLP/internal/llm"
)

// DefaultLLMIntegration implements the LLMIntegration interface
type DefaultLLMIntegration struct {
	llmClient llm.Client
	templates map[string]*template.Template
	parsers   map[ResponseType]ResponseParser
}

// ResponseParser defines interface for parsing different response types
type ResponseParser interface {
	Parse(response string, target interface{}) error
	Validate(response string) error
}

// NewDefaultLLMIntegration creates a new LLM integration layer
func NewDefaultLLMIntegration(llmClient llm.Client) *DefaultLLMIntegration {
	integration := &DefaultLLMIntegration{
		llmClient: llmClient,
		templates: make(map[string]*template.Template),
		parsers: map[ResponseType]ResponseParser{
			ResponseTypeJSON:       &JSONResponseParser{},
			ResponseTypeStructured: &StructuredResponseParser{},
			ResponseTypeText:       &TextResponseParser{},
		},
	}
	
	integration.initializeTemplates()
	return integration
}

// GeneratePrompt creates a prompt based on input and type
func (dli *DefaultLLMIntegration) GeneratePrompt(input *ValidationInput, promptType PromptType) string {
	templateKey := fmt.Sprintf("%s_%s", input.Language, promptType)
	
	// Try specific language template first
	if tmpl, exists := dli.templates[templateKey]; exists {
		return dli.executeTemplate(tmpl, input)
	}
	
	// Fall back to generic template
	genericKey := fmt.Sprintf("generic_%s", promptType)
	if tmpl, exists := dli.templates[genericKey]; exists {
		return dli.executeTemplate(tmpl, input)
	}
	
	// Ultimate fallback to basic prompt
	return dli.generateBasicPrompt(input, promptType)
}

// ParseResponse parses LLM response based on expected type
func (dli *DefaultLLMIntegration) ParseResponse(response string, expectedType ResponseType) (interface{}, error) {
	parser, exists := dli.parsers[expectedType]
	if !exists {
		return nil, fmt.Errorf("unsupported response type: %s", expectedType)
	}
	
	var result interface{}
	switch expectedType {
	case ResponseTypeJSON:
		result = &ValidationResult{}
	case ResponseTypeStructured:
		result = &StructuredAnalysis{}
	case ResponseTypeText:
		result = &TextAnalysis{}
	}
	
	if err := parser.Parse(response, result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return result, nil
}

// GetPromptTemplate returns the template for a specific validator and prompt type
func (dli *DefaultLLMIntegration) GetPromptTemplate(validatorType ValidatorType, promptType PromptType) string {
	templateKey := fmt.Sprintf("%s_%s", validatorType, promptType)
	
	if tmpl, exists := dli.templates[templateKey]; exists {
		return tmpl.Root.String()
	}
	
	return ""
}

// ValidateResponse checks if the response meets quality standards
func (dli *DefaultLLMIntegration) ValidateResponse(response interface{}) error {
	switch r := response.(type) {
	case *ValidationResult:
		return dli.validateValidationResult(r)
	case *StructuredAnalysis:
		return dli.validateStructuredAnalysis(r)
	case *TextAnalysis:
		return dli.validateTextAnalysis(r)
	default:
		return fmt.Errorf("unsupported response type for validation")
	}
}

// Helper methods

func (dli *DefaultLLMIntegration) executeTemplate(tmpl *template.Template, input *ValidationInput) string {
	var buf strings.Builder
	
	templateData := map[string]interface{}{
		"Input":           input,
		"Content":         input.Content,
		"Language":        input.Language,
		"Framework":       input.Framework,
		"ProjectPath":     input.ProjectPath,
		"ValidationTypes": input.ValidationTypes,
		"Requirements":    input.Requirements,
		"UserContext":     input.UserContext,
		"ProjectMetadata": input.ProjectMetadata,
	}
	
	if err := tmpl.Execute(&buf, templateData); err != nil {
		// Fallback to basic prompt if template execution fails
		return dli.generateBasicPrompt(input, PromptTypeAnalysis)
	}
	
	return buf.String()
}

func (dli *DefaultLLMIntegration) generateBasicPrompt(input *ValidationInput, promptType PromptType) string {
	switch promptType {
	case PromptTypeAnalysis:
		return dli.generateAnalysisPrompt(input)
	case PromptTypeValidation:
		return dli.generateValidationPrompt(input)
	case PromptTypeSuggestion:
		return dli.generateSuggestionPrompt(input)
	case PromptTypeExplanation:
		return dli.generateExplanationPrompt(input)
	default:
		return dli.generateAnalysisPrompt(input)
	}
}

func (dli *DefaultLLMIntegration) generateAnalysisPrompt(input *ValidationInput) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf(`You are a senior software architect analyzing a %s project for comprehensive validation.

PROJECT CONTEXT:
Language: %s
Framework: %s
Project Type: %s

VALIDATION SCOPE:
`, input.Language, input.Framework, input.ProjectMetadata.ProjectType))
	
	for _, vType := range input.ValidationTypes {
		builder.WriteString(fmt.Sprintf("- %s validation\n", vType))
	}
	
	builder.WriteString("\nCODE TO ANALYZE:\n")
	for filePath, content := range input.Content {
		builder.WriteString(fmt.Sprintf("\n=== %s ===\n", filePath))
		if len(content) > 2000 {
			builder.WriteString(content[:2000])
			builder.WriteString("\n... [truncated] ...\n")
		} else {
			builder.WriteString(content)
		}
	}
	
	builder.WriteString(`

ANALYSIS REQUIREMENTS:
1. Provide comprehensive assessment with scores (0-100)
2. Identify security vulnerabilities with CVSS scores
3. Assess code quality and maintainability
4. Evaluate performance characteristics
5. Check compliance with best practices
6. Provide actionable recommendations

RESPOND WITH JSON:
{
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
}`)
	
	return builder.String()
}

func (dli *DefaultLLMIntegration) generateValidationPrompt(input *ValidationInput) string {
	return fmt.Sprintf(`Validate the following %s code for production readiness:

%s

Focus on:
- Security vulnerabilities
- Code quality issues
- Performance bottlenecks
- Best practice violations

Provide detailed validation report with specific recommendations.`, 
		input.Language, dli.formatContent(input.Content))
}

func (dli *DefaultLLMIntegration) generateSuggestionPrompt(input *ValidationInput) string {
	return fmt.Sprintf(`Analyze this %s code and provide improvement suggestions:

%s

Provide specific, actionable suggestions for:
1. Code quality improvements
2. Performance optimizations
3. Security enhancements
4. Best practice adoption

Format as prioritized list with implementation guidance.`, 
		input.Language, dli.formatContent(input.Content))
}

func (dli *DefaultLLMIntegration) generateExplanationPrompt(input *ValidationInput) string {
	return fmt.Sprintf(`Explain the architecture and implementation of this %s project:

%s

Provide comprehensive explanation covering:
1. Overall architecture and design patterns
2. Key components and their responsibilities
3. Data flow and interactions
4. Technology choices and rationale
5. Potential areas for improvement

Make it understandable for both technical and non-technical stakeholders.`, 
		input.Language, dli.formatContent(input.Content))
}

func (dli *DefaultLLMIntegration) formatContent(content map[string]string) string {
	var builder strings.Builder
	
	for filePath, fileContent := range content {
		builder.WriteString(fmt.Sprintf("\n=== %s ===\n", filePath))
		if len(fileContent) > 1500 {
			builder.WriteString(fileContent[:1500])
			builder.WriteString("\n... [truncated] ...\n")
		} else {
			builder.WriteString(fileContent)
		}
	}
	
	return builder.String()
}

func (dli *DefaultLLMIntegration) initializeTemplates() {
	// Initialize prompt templates for different combinations
	templateConfigs := map[string]string{
		"generic_analysis": `You are a {{.Language}} expert analyzing code for validation.

PROJECT: {{.ProjectMetadata.ProjectType}}
FRAMEWORK: {{.Framework}}

{{range $path, $content := .Content}}
=== {{$path}} ===
{{$content}}

{{end}}

Provide comprehensive analysis with JSON response.`,
		
		"security_validation": `You are a cybersecurity expert reviewing {{.Language}} code.

SECURITY ASSESSMENT for {{.ProjectMetadata.ProjectType}}:

{{range $path, $content := .Content}}
=== {{$path}} ===
{{$content}}

{{end}}

Focus on OWASP Top 10, CWE vulnerabilities, and security best practices.
Provide detailed security findings with CVSS scores.`,

		"quality_analysis": `You are a senior code reviewer analyzing {{.Language}} code quality.

QUALITY ASSESSMENT:

{{range $path, $content := .Content}}
=== {{$path}} ===
{{$content}}

{{end}}

Evaluate:
- Code complexity and maintainability
- Design patterns and architecture
- Test coverage and quality
- Documentation completeness
- Best practices adherence

Provide quality metrics and improvement recommendations.`,
	}
	
	for name, templateStr := range templateConfigs {
		tmpl, err := template.New(name).Parse(templateStr)
		if err == nil {
			dli.templates[name] = tmpl
		}
	}
}

func (dli *DefaultLLMIntegration) validateValidationResult(result *ValidationResult) error {
	if result.OverallScore < 0 || result.OverallScore > 100 {
		return fmt.Errorf("invalid overall score: %d (must be 0-100)", result.OverallScore)
	}
	
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence: %f (must be 0.0-1.0)", result.Confidence)
	}
	
	for component, score := range result.ComponentScores {
		if score < 0 || score > 100 {
			return fmt.Errorf("invalid component score for %s: %d (must be 0-100)", component, score)
		}
	}
	
	return nil
}

func (dli *DefaultLLMIntegration) validateStructuredAnalysis(analysis *StructuredAnalysis) error {
	if analysis.Language == "" {
		return fmt.Errorf("language is required in structured analysis")
	}
	
	if analysis.Confidence < 0.0 || analysis.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence: %f (must be 0.0-1.0)", analysis.Confidence)
	}
	
	return nil
}

func (dli *DefaultLLMIntegration) validateTextAnalysis(analysis *TextAnalysis) error {
	if len(analysis.Content) < 10 {
		return fmt.Errorf("text analysis content too short: %d characters", len(analysis.Content))
	}
	
	return nil
}

// Response parser implementations

type JSONResponseParser struct{}

func (jrp *JSONResponseParser) Parse(response string, target interface{}) error {
	// Extract JSON from response (handle cases where LLM adds extra text)
	jsonStr := extractJSON(response)
	return json.Unmarshal([]byte(jsonStr), target)
}

func (jrp *JSONResponseParser) Validate(response string) error {
	jsonStr := extractJSON(response)
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}

type StructuredResponseParser struct{}

func (srp *StructuredResponseParser) Parse(response string, target interface{}) error {
	// Parse structured non-JSON response
	analysis := &StructuredAnalysis{
		Language:   extractField(response, "Language:"),
		Framework:  extractField(response, "Framework:"),
		Issues:     extractList(response, "Issues:"),
		Suggestions: extractList(response, "Suggestions:"),
		Confidence: extractFloat(response, "Confidence:"),
	}
	
	if targetAnalysis, ok := target.(*StructuredAnalysis); ok {
		*targetAnalysis = *analysis
		return nil
	}
	
	return fmt.Errorf("target is not *StructuredAnalysis")
}

func (srp *StructuredResponseParser) Validate(response string) error {
	if len(response) < 50 {
		return fmt.Errorf("structured response too short")
	}
	return nil
}

type TextResponseParser struct{}

func (trp *TextResponseParser) Parse(response string, target interface{}) error {
	analysis := &TextAnalysis{
		Content:   response,
		WordCount: len(strings.Fields(response)),
	}
	
	if targetAnalysis, ok := target.(*TextAnalysis); ok {
		*targetAnalysis = *analysis
		return nil
	}
	
	return fmt.Errorf("target is not *TextAnalysis")
}

func (trp *TextResponseParser) Validate(response string) error {
	if len(response) < 10 {
		return fmt.Errorf("text response too short")
	}
	return nil
}

// Supporting types for different response formats

type StructuredAnalysis struct {
	Language    string    `json:"language"`
	Framework   string    `json:"framework"`
	Issues      []string  `json:"issues"`
	Suggestions []string  `json:"suggestions"`
	Confidence  float64   `json:"confidence"`
}

type TextAnalysis struct {
	Content   string `json:"content"`
	WordCount int    `json:"word_count"`
}

// Utility functions

func extractJSON(response string) string {
	// Find JSON block in response
	start := strings.Index(response, "{")
	if start == -1 {
		return "{}"
	}
	
	end := strings.LastIndex(response, "}")
	if end == -1 || end <= start {
		return "{}"
	}
	
	return response[start : end+1]
}

func extractField(response, fieldName string) string {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.Contains(line, fieldName) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func extractList(response, fieldName string) []string {
	var items []string
	lines := strings.Split(response, "\n")
	inList := false
	
	for _, line := range lines {
		if strings.Contains(line, fieldName) {
			inList = true
			continue
		}
		
		if inList {
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				item := strings.TrimSpace(line[1:])
				items = append(items, item)
			}
		}
	}
	
	return items
}

func extractFloat(response, fieldName string) float64 {
	field := extractField(response, fieldName)
	if field == "" {
		return 0.0
	}
	
	// Simple float parsing
	var result float64
	fmt.Sscanf(field, "%f", &result)
	return result
}