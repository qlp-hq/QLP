package parser

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// ValidationResponseValidator validates validation-specific responses
type ValidationResponseValidator struct{}

func (vrv *ValidationResponseValidator) Validate(ctx context.Context, response interface{}) error {
	validationResp, ok := response.(*ValidationResponse)
	if !ok {
		return fmt.Errorf("expected *ValidationResponse, got %T", response)
	}

	// Validate overall score
	if validationResp.OverallScore < 0 || validationResp.OverallScore > 100 {
		return fmt.Errorf("overall_score must be between 0 and 100, got %d", validationResp.OverallScore)
	}

	// Validate security score
	if validationResp.SecurityScore < 0 || validationResp.SecurityScore > 100 {
		return fmt.Errorf("security_score must be between 0 and 100, got %d", validationResp.SecurityScore)
	}

	// Validate quality score
	if validationResp.QualityScore < 0 || validationResp.QualityScore > 100 {
		return fmt.Errorf("quality_score must be between 0 and 100, got %d", validationResp.QualityScore)
	}

	// Validate performance score
	if validationResp.PerformanceScore < 0 || validationResp.PerformanceScore > 100 {
		return fmt.Errorf("performance_score must be between 0 and 100, got %d", validationResp.PerformanceScore)
	}

	// Validate confidence
	if validationResp.Confidence < 0.0 || validationResp.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", validationResp.Confidence)
	}

	// Validate issues
	for i, issue := range validationResp.Issues {
		if err := vrv.validateIssue(issue, i); err != nil {
			return fmt.Errorf("issue %d validation failed: %w", i, err)
		}
	}

	// Validate recommendations
	for i, rec := range validationResp.Recommendations {
		if err := vrv.validateRecommendation(rec, i); err != nil {
			return fmt.Errorf("recommendation %d validation failed: %w", i, err)
		}
	}

	// Validate timestamp if present
	if validationResp.Timestamp != "" {
		if _, err := time.Parse(time.RFC3339, validationResp.Timestamp); err != nil {
			return fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	return nil
}

func (vrv *ValidationResponseValidator) validateIssue(issue ValidationIssue, index int) error {
	if issue.Type == "" {
		return fmt.Errorf("issue type is required")
	}

	validSeverities := []string{"info", "low", "medium", "high", "critical"}
	if !vrv.contains(validSeverities, issue.Severity) {
		return fmt.Errorf("invalid severity '%s', must be one of: %v", issue.Severity, validSeverities)
	}

	if issue.Description == "" {
		return fmt.Errorf("issue description is required")
	}

	return nil
}

func (vrv *ValidationResponseValidator) validateRecommendation(rec ValidationRecommendation, index int) error {
	validPriorities := []string{"low", "medium", "high", "critical"}
	if rec.Priority != "" && !vrv.contains(validPriorities, rec.Priority) {
		return fmt.Errorf("invalid priority '%s', must be one of: %v", rec.Priority, validPriorities)
	}

	if rec.Description == "" {
		return fmt.Errorf("recommendation description is required")
	}

	validImpacts := []string{"low", "medium", "high"}
	if rec.Impact != "" && !vrv.contains(validImpacts, rec.Impact) {
		return fmt.Errorf("invalid impact '%s', must be one of: %v", rec.Impact, validImpacts)
	}

	validEfforts := []string{"low", "medium", "high"}
	if rec.Effort != "" && !vrv.contains(validEfforts, rec.Effort) {
		return fmt.Errorf("invalid effort '%s', must be one of: %v", rec.Effort, validEfforts)
	}

	return nil
}

func (vrv *ValidationResponseValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (vrv *ValidationResponseValidator) GetType() ResponseType {
	return ResponseTypeValidation
}

// AnalysisResponseValidator validates analysis-specific responses
type AnalysisResponseValidator struct{}

func (arv *AnalysisResponseValidator) Validate(ctx context.Context, response interface{}) error {
	analysisResp, ok := response.(*AnalysisResponse)
	if !ok {
		return fmt.Errorf("expected *AnalysisResponse, got %T", response)
	}

	// Validate analysis type
	if analysisResp.AnalysisType == "" {
		return fmt.Errorf("analysis_type is required")
	}

	validAnalysisTypes := []string{
		"security", "performance", "quality", "architecture", 
		"compliance", "general", "code_review", "vulnerability",
	}
	if !arv.contains(validAnalysisTypes, analysisResp.AnalysisType) {
		return fmt.Errorf("invalid analysis_type '%s', must be one of: %v", 
			analysisResp.AnalysisType, validAnalysisTypes)
	}

	// Validate confidence
	if analysisResp.Confidence < 0.0 || analysisResp.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", analysisResp.Confidence)
	}

	// Validate findings
	for i, finding := range analysisResp.Findings {
		if err := arv.validateFinding(finding, i); err != nil {
			return fmt.Errorf("finding %d validation failed: %w", i, err)
		}
	}

	// Validate recommendations
	for i, rec := range analysisResp.Recommendations {
		if err := arv.validateAnalysisRecommendation(rec, i); err != nil {
			return fmt.Errorf("recommendation %d validation failed: %w", i, err)
		}
	}

	// Validate timestamp if present
	if analysisResp.Timestamp != "" {
		if _, err := time.Parse(time.RFC3339, analysisResp.Timestamp); err != nil {
			return fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	return nil
}

func (arv *AnalysisResponseValidator) validateFinding(finding AnalysisFinding, index int) error {
	if finding.Category == "" {
		return fmt.Errorf("finding category is required")
	}

	validCategories := []string{
		"security", "performance", "quality", "maintainability",
		"scalability", "reliability", "usability", "compatibility",
	}
	if !arv.contains(validCategories, finding.Category) {
		return fmt.Errorf("invalid category '%s', must be one of: %v", 
			finding.Category, validCategories)
	}

	if finding.Description == "" {
		return fmt.Errorf("finding description is required")
	}

	if finding.Confidence < 0.0 || finding.Confidence > 1.0 {
		return fmt.Errorf("finding confidence must be between 0.0 and 1.0, got %f", finding.Confidence)
	}

	return nil
}

func (arv *AnalysisResponseValidator) validateAnalysisRecommendation(rec AnalysisRecommendation, index int) error {
	validTypes := []string{
		"fix", "improve", "optimize", "refactor", "upgrade", 
		"replace", "monitor", "investigate", "document",
	}
	if rec.Type != "" && !arv.contains(validTypes, rec.Type) {
		return fmt.Errorf("invalid recommendation type '%s', must be one of: %v", rec.Type, validTypes)
	}

	if rec.Description == "" {
		return fmt.Errorf("recommendation description is required")
	}

	validPriorities := []string{"low", "medium", "high", "critical"}
	if rec.Priority != "" && !arv.contains(validPriorities, rec.Priority) {
		return fmt.Errorf("invalid priority '%s', must be one of: %v", rec.Priority, validPriorities)
	}

	validImpacts := []string{"low", "medium", "high"}
	if rec.Impact != "" && !arv.contains(validImpacts, rec.Impact) {
		return fmt.Errorf("invalid impact '%s', must be one of: %v", rec.Impact, validImpacts)
	}

	return nil
}

func (arv *AnalysisResponseValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (arv *AnalysisResponseValidator) GetType() ResponseType {
	return ResponseTypeAnalysis
}

// JSONResponseValidator validates generic JSON responses
type JSONResponseValidator struct{}

func (jrv *JSONResponseValidator) Validate(ctx context.Context, response interface{}) error {
	// For generic JSON, just check that it's a valid map or slice
	v := reflect.ValueOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// Valid JSON object
		return nil
	case reflect.Slice:
		// Valid JSON array
		return nil
	case reflect.Struct:
		// Valid structured response
		return nil
	default:
		return fmt.Errorf("invalid JSON response type: %T", response)
	}
}

func (jrv *JSONResponseValidator) GetType() ResponseType {
	return ResponseTypeJSON
}

// StructuredResponseValidator validates structured text responses
type StructuredResponseValidator struct{}

func (srv *StructuredResponseValidator) Validate(ctx context.Context, response interface{}) error {
	// For structured responses, ensure it's a map with some content
	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", response)
	}

	if len(responseMap) == 0 {
		return fmt.Errorf("structured response cannot be empty")
	}

	return nil
}

func (srv *StructuredResponseValidator) GetType() ResponseType {
	return ResponseTypeStructured
}

// TextResponseValidator validates plain text responses
type TextResponseValidator struct{}

func (trv *TextResponseValidator) Validate(ctx context.Context, response interface{}) error {
	genericResp, ok := response.(*GenericResponse)
	if !ok {
		return fmt.Errorf("expected *GenericResponse, got %T", response)
	}

	if genericResp.Content == "" {
		return fmt.Errorf("text response content cannot be empty")
	}

	if len(genericResp.Content) < 10 {
		return fmt.Errorf("text response content too short: %d characters", len(genericResp.Content))
	}

	return nil
}

func (trv *TextResponseValidator) GetType() ResponseType {
	return ResponseTypeText
}

// CompositeValidator combines multiple validators for comprehensive validation
type CompositeValidator struct {
	validators []ResponseValidator
}

func NewCompositeValidator(validators ...ResponseValidator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

func (cv *CompositeValidator) Validate(ctx context.Context, response interface{}) error {
	var errors []error

	for _, validator := range cv.validators {
		if err := validator.Validate(ctx, response); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed with %d errors: %v", len(errors), errors)
	}

	return nil
}

func (cv *CompositeValidator) GetType() ResponseType {
	return ResponseTypeJSON // Generic type for composite
}

// BusinessLogicValidator validates business-specific rules
type BusinessLogicValidator struct{}

func (blv *BusinessLogicValidator) Validate(ctx context.Context, response interface{}) error {
	// Business logic validation (can be extended based on domain requirements)
	
	// Check for validation responses
	if validationResp, ok := response.(*ValidationResponse); ok {
		return blv.validateBusinessRules(validationResp)
	}

	// Check for analysis responses
	if analysisResp, ok := response.(*AnalysisResponse); ok {
		return blv.validateAnalysisBusinessRules(analysisResp)
	}

	return nil // No business rules for other response types
}

func (blv *BusinessLogicValidator) validateBusinessRules(resp *ValidationResponse) error {
	// Business rule: Security score should correlate with overall score
	if resp.SecurityScore < 50 && resp.OverallScore > 80 {
		return fmt.Errorf("business rule violation: low security score (%d) with high overall score (%d)", 
			resp.SecurityScore, resp.OverallScore)
	}

	// Business rule: Critical issues should impact overall score significantly
	criticalIssues := 0
	for _, issue := range resp.Issues {
		if issue.Severity == "critical" {
			criticalIssues++
		}
	}

	if criticalIssues > 0 && resp.OverallScore > 70 {
		return fmt.Errorf("business rule violation: %d critical issues found but overall score is %d (should be ≤ 70)", 
			criticalIssues, resp.OverallScore)
	}

	return nil
}

func (blv *BusinessLogicValidator) validateAnalysisBusinessRules(resp *AnalysisResponse) error {
	// Business rule: Analysis should have findings if confidence is high
	if resp.Confidence > 0.8 && len(resp.Findings) == 0 {
		return fmt.Errorf("business rule violation: high confidence (%.2f) analysis should have findings", 
			resp.Confidence)
	}

	// Business rule: Security analysis should have security-specific recommendations
	if resp.AnalysisType == "security" {
		securityRecs := 0
		for _, rec := range resp.Recommendations {
			if rec.Type == "fix" || rec.Priority == "high" || rec.Priority == "critical" {
				securityRecs++
			}
		}

		if len(resp.Findings) > 3 && securityRecs == 0 {
			return fmt.Errorf("business rule violation: security analysis with %d findings should have actionable recommendations", 
				len(resp.Findings))
		}
	}

	return nil
}

func (blv *BusinessLogicValidator) GetType() ResponseType {
	return ResponseTypeJSON // Generic type for business logic
}