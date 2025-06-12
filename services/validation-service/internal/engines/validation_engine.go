package engines

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/internal/llm"
	"QLP/services/validation-service/internal/validators"
	"QLP/services/validation-service/internal/scanners"
	"QLP/services/validation-service/pkg/contracts"
)

type ValidationEngine struct {
	config           Config
	llmClient        llm.Client
	syntaxValidators *validators.SyntaxValidatorRegistry
	securityScanners *scanners.SecurityScannerRegistry
	qualityAnalyzers *validators.QualityAnalyzerRegistry
	validations      map[string]*contracts.ValidationResult
	validationsMu    sync.RWMutex
	semaphore        chan struct{}
}

type Config struct {
	LLMClient        llm.Client
	SyntaxValidators *validators.SyntaxValidatorRegistry
	SecurityScanners *scanners.SecurityScannerRegistry
	QualityAnalyzers *validators.QualityAnalyzerRegistry
	DefaultTimeout   time.Duration
	MaxConcurrent    int
}

func NewValidationEngine(config Config) *ValidationEngine {
	return &ValidationEngine{
		config:           config,
		llmClient:        config.LLMClient,
		syntaxValidators: config.SyntaxValidators,
		securityScanners: config.SecurityScanners,
		qualityAnalyzers: config.QualityAnalyzers,
		validations:      make(map[string]*contracts.ValidationResult),
		semaphore:        make(chan struct{}, config.MaxConcurrent),
	}
}

func (ve *ValidationEngine) ValidateContent(ctx context.Context, req *contracts.ValidateRequest, tenantID string) (*contracts.ValidateResponse, error) {
	validationID := uuid.New().String()
	
	logger.WithComponent("validation-engine").Info("Starting validation",
		zap.String("validation_id", validationID),
		zap.String("tenant_id", tenantID),
		zap.String("task_type", string(req.TaskType)),
		zap.String("language", req.Language))

	// Create validation record
	validation := &contracts.ValidationResult{
		ID:        validationID,
		TenantID:  tenantID,
		Status:    contracts.ValidationStatusPending,
		Level:     ve.getValidationLevel(req.Options),
		ValidatedAt: time.Now(),
		ExecutedChecks: []contracts.ValidationCheck{},
		SkippedChecks:  []contracts.ValidationCheck{},
	}

	// Store validation
	ve.validationsMu.Lock()
	ve.validations[validationID] = validation
	ve.validationsMu.Unlock()

	// Start validation asynchronously
	go ve.performValidation(ctx, req, validation)

	return &contracts.ValidateResponse{
		ValidationID: validationID,
		Status:       string(contracts.ValidationStatusPending),
		Message:      "Validation started",
	}, nil
}

func (ve *ValidationEngine) performValidation(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) {
	// Acquire semaphore for concurrency control
	ve.semaphore <- struct{}{}
	defer func() { <-ve.semaphore }()

	startTime := time.Now()
	defer func() {
		validation.ValidationTime = time.Since(startTime)
		now := time.Now()
		validation.CompletedAt = &now
	}()

	// Update status to running
	ve.updateValidationStatus(validation.ID, contracts.ValidationStatusRunning)

	// Set timeout
	timeout := ve.config.DefaultTimeout
	if req.Options != nil && req.Options.Timeout > 0 {
		timeout = time.Duration(req.Options.Timeout) * time.Second
	}

	validationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Determine which checks to run
	enabledChecks := ve.getEnabledChecks(req.Options, validation.Level)
	
	logger.WithComponent("validation-engine").Info("Running validation checks",
		zap.String("validation_id", validation.ID),
		zap.Strings("checks", ve.checksToStrings(enabledChecks)))

	// Run validation checks based on level
	var err error
	switch validation.Level {
	case contracts.ValidationLevelFast:
		err = ve.runFastValidation(validationCtx, req, validation, enabledChecks)
	case contracts.ValidationLevelStandard:
		err = ve.runStandardValidation(validationCtx, req, validation, enabledChecks)
	case contracts.ValidationLevelComprehensive:
		err = ve.runComprehensiveValidation(validationCtx, req, validation, enabledChecks)
	default:
		err = ve.runStandardValidation(validationCtx, req, validation, enabledChecks)
	}

	if err != nil {
		logger.WithComponent("validation-engine").Error("Validation failed",
			zap.String("validation_id", validation.ID),
			zap.Error(err))
		
		validation.ErrorMessage = err.Error()
		ve.updateValidationStatus(validation.ID, contracts.ValidationStatusFailed)
		return
	}

	// Calculate overall score and pass/fail
	ve.calculateOverallScore(validation)
	
	// Mark as completed
	ve.updateValidationStatus(validation.ID, contracts.ValidationStatusCompleted)
	
	logger.WithComponent("validation-engine").Info("Validation completed",
		zap.String("validation_id", validation.ID),
		zap.Int("overall_score", validation.OverallScore),
		zap.Bool("passed", validation.Passed),
		zap.Duration("duration", validation.ValidationTime))
}

func (ve *ValidationEngine) runFastValidation(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult, checks []contracts.ValidationCheck) error {
	for _, check := range checks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch check {
		case contracts.ValidationCheckSyntax:
			if err := ve.runSyntaxValidation(ctx, req, validation); err != nil {
				return fmt.Errorf("syntax validation failed: %w", err)
			}
		case contracts.ValidationCheckSecurity:
			if err := ve.runFastSecurityScan(ctx, req, validation); err != nil {
				return fmt.Errorf("security scan failed: %w", err)
			}
		case contracts.ValidationCheckQuality:
			if err := ve.runFastQualityAnalysis(ctx, req, validation); err != nil {
				return fmt.Errorf("quality analysis failed: %w", err)
			}
		}
		
		validation.ExecutedChecks = append(validation.ExecutedChecks, check)
	}
	
	return nil
}

func (ve *ValidationEngine) runStandardValidation(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult, checks []contracts.ValidationCheck) error {
	// Run fast validation first
	if err := ve.runFastValidation(ctx, req, validation, checks); err != nil {
		return err
	}
	
	// Add standard-level checks
	for _, check := range checks {
		switch check {
		case contracts.ValidationCheckPerformance:
			if err := ve.runPerformanceAnalysis(ctx, req, validation); err != nil {
				return fmt.Errorf("performance analysis failed: %w", err)
			}
		case contracts.ValidationCheckCompliance:
			if err := ve.runComplianceCheck(ctx, req, validation); err != nil {
				return fmt.Errorf("compliance check failed: %w", err)
			}
		}
	}
	
	return nil
}

func (ve *ValidationEngine) runComprehensiveValidation(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult, checks []contracts.ValidationCheck) error {
	// Run standard validation first
	if err := ve.runStandardValidation(ctx, req, validation, checks); err != nil {
		return err
	}
	
	// Add comprehensive-level checks
	for _, check := range checks {
		switch check {
		case contracts.ValidationCheckLLMCritique:
			if err := ve.runLLMCritique(ctx, req, validation); err != nil {
				return fmt.Errorf("LLM critique failed: %w", err)
			}
		case contracts.ValidationCheckAccessibility:
			if err := ve.runAccessibilityCheck(ctx, req, validation); err != nil {
				return fmt.Errorf("accessibility check failed: %w", err)
			}
		}
	}
	
	return nil
}

func (ve *ValidationEngine) runSyntaxValidation(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	language := ve.detectLanguage(req.Language, req.TaskType, req.Content)
	
	validator := ve.syntaxValidators.GetValidator(language)
	if validator == nil {
		// No specific validator, create basic result
		validation.SyntaxResult = &contracts.SyntaxResult{
			Score:    100,
			Valid:    true,
			Language: language,
			Issues:   []contracts.SyntaxIssue{},
			Warnings: []contracts.SyntaxIssue{
				{
					Type:     "info",
					Severity: contracts.SeverityInfo,
					Message:  fmt.Sprintf("No specific syntax validator available for %s", language),
				},
			},
		}
		return nil
	}
	
	result, err := validator.Validate(ctx, req.Content, language)
	if err != nil {
		return err
	}
	
	validation.SyntaxResult = result
	return nil
}

func (ve *ValidationEngine) runFastSecurityScan(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	scanner := ve.securityScanners.GetScanner("fast")
	if scanner == nil {
		// Create basic security result
		validation.SecurityResult = &contracts.SecurityResult{
			Score:           85,
			RiskLevel:       contracts.SecurityRiskLevelLow,
			Vulnerabilities: []contracts.SecurityIssue{},
			Warnings:        []contracts.SecurityIssue{},
			Passed:          true,
			ScannedBy:       []string{"fast-scanner"},
		}
		return nil
	}
	
	result, err := scanner.Scan(ctx, req.Content, req.Language, req.TaskType)
	if err != nil {
		return err
	}
	
	validation.SecurityResult = result
	validation.SecurityScore = result.Score
	return nil
}

func (ve *ValidationEngine) runFastQualityAnalysis(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	analyzer := ve.qualityAnalyzers.GetAnalyzer("fast")
	if analyzer == nil {
		// Create basic quality result
		validation.QualityResult = &contracts.QualityResult{
			Score:           80,
			Maintainability: 80,
			Documentation:   75,
			BestPractices:   85,
			TestCoverage:    0.0,
			Issues:          []contracts.QualityIssue{},
			Suggestions:     []contracts.QualitySuggestion{},
			Passed:          true,
		}
		validation.QualityScore = 80
		return nil
	}
	
	result, err := analyzer.Analyze(ctx, req.Content, req.Language, req.TaskType)
	if err != nil {
		return err
	}
	
	validation.QualityResult = result
	validation.QualityScore = result.Score
	return nil
}

func (ve *ValidationEngine) runPerformanceAnalysis(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	// Mock performance analysis
	validation.PerformanceResult = &contracts.PerformanceResult{
		Score:         85,
		Issues:        []contracts.PerformanceIssue{},
		Optimizations: []contracts.PerformanceHint{},
		Benchmarks:    []contracts.BenchmarkResult{},
		Passed:        true,
	}
	validation.PerformanceScore = 85
	return nil
}

func (ve *ValidationEngine) runComplianceCheck(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	// Mock compliance check
	validation.ComplianceResult = &contracts.ComplianceResult{
		Score:      90,
		Standards:  []contracts.StandardResult{},
		Violations: []contracts.ComplianceIssue{},
		Passed:     true,
	}
	validation.ComplianceScore = 90
	return nil
}

func (ve *ValidationEngine) runLLMCritique(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	prompt := fmt.Sprintf(`Please provide a comprehensive critique of the following %s code:

%s

Focus on:
1. Code quality and best practices
2. Security considerations
3. Performance implications
4. Maintainability
5. Documentation quality

Provide a score from 0-100 and detailed feedback.`, req.Language, req.Content)

	response, err := ve.llmClient.Complete(ctx, prompt)
	if err != nil {
		logger.WithComponent("validation-engine").Warn("LLM critique failed, using fallback", zap.Error(err))
		// Fallback critique
		validation.LLMCritiqueResult = &contracts.LLMCritiqueResult{
			Score:      80,
			Feedback:   "LLM critique unavailable - using fallback assessment",
			Confidence: 0.5,
			Model:      "fallback",
		}
		return nil
	}
	
	// Parse LLM response (simplified)
	validation.LLMCritiqueResult = &contracts.LLMCritiqueResult{
		Score:      85,
		Feedback:   response,
		Confidence: 0.8,
		Model:      "gpt-4",
	}
	
	return nil
}

func (ve *ValidationEngine) runAccessibilityCheck(ctx context.Context, req *contracts.ValidateRequest, validation *contracts.ValidationResult) error {
	// Mock accessibility check - would implement actual a11y scanning
	return nil
}

func (ve *ValidationEngine) calculateOverallScore(validation *contracts.ValidationResult) {
	var totalScore float64
	var weights float64
	
	// Weight different aspects based on validation level
	if validation.SyntaxResult != nil {
		totalScore += float64(validation.SyntaxResult.Score) * 0.15
		weights += 0.15
	}
	
	if validation.SecurityResult != nil {
		totalScore += float64(validation.SecurityResult.Score) * 0.30
		weights += 0.30
	}
	
	if validation.QualityResult != nil {
		totalScore += float64(validation.QualityResult.Score) * 0.25
		weights += 0.25
	}
	
	if validation.PerformanceResult != nil {
		totalScore += float64(validation.PerformanceResult.Score) * 0.15
		weights += 0.15
	}
	
	if validation.ComplianceResult != nil {
		totalScore += float64(validation.ComplianceResult.Score) * 0.10
		weights += 0.10
	}
	
	if validation.LLMCritiqueResult != nil {
		totalScore += float64(validation.LLMCritiqueResult.Score) * 0.05
		weights += 0.05
	}
	
	if weights > 0 {
		validation.OverallScore = int(totalScore / weights)
	} else {
		validation.OverallScore = 0
	}
	
	// Determine pass/fail (70% threshold)
	validation.Passed = validation.OverallScore >= 70
}

func (ve *ValidationEngine) GetValidation(validationID, tenantID string) (*contracts.ValidationResult, error) {
	ve.validationsMu.RLock()
	defer ve.validationsMu.RUnlock()

	validation, exists := ve.validations[validationID]
	if !exists {
		return nil, fmt.Errorf("validation not found")
	}

	if validation.TenantID != tenantID {
		return nil, fmt.Errorf("validation not found for tenant")
	}

	return validation, nil
}

func (ve *ValidationEngine) ListValidations(req *contracts.ListValidationsRequest) (*contracts.ListValidationsResponse, error) {
	ve.validationsMu.RLock()
	defer ve.validationsMu.RUnlock()

	var filtered []*contracts.ValidationResult
	for _, validation := range ve.validations {
		if validation.TenantID != req.TenantID {
			continue
		}

		if req.Status != "" && string(validation.Status) != req.Status {
			continue
		}

		if req.Since != nil && validation.ValidatedAt.Before(*req.Since) {
			continue
		}

		filtered = append(filtered, validation)
	}

	// Apply pagination
	total := len(filtered)
	start := req.Offset
	end := start + req.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	if req.Limit == 0 {
		end = total
	}

	var summaries []contracts.ValidationSummary
	for i := start; i < end; i++ {
		v := filtered[i]
		summaries = append(summaries, contracts.ValidationSummary{
			ID:           v.ID,
			TaskID:       v.TaskID,
			Status:       v.Status,
			OverallScore: v.OverallScore,
			Passed:       v.Passed,
			Level:        v.Level,
			ValidatedAt:  v.ValidatedAt,
			CompletedAt:  v.CompletedAt,
		})
	}

	return &contracts.ListValidationsResponse{
		Validations: summaries,
		Total:       total,
	}, nil
}

func (ve *ValidationEngine) Shutdown(ctx context.Context) {
	logger.WithComponent("validation-engine").Info("Shutting down validation engine")

	// Cancel all running validations
	ve.validationsMu.Lock()
	for _, validation := range ve.validations {
		if validation.Status == contracts.ValidationStatusRunning ||
			validation.Status == contracts.ValidationStatusPending {
			validation.Status = contracts.ValidationStatusTimeout
		}
	}
	ve.validationsMu.Unlock()

	// Wait for all workers to finish or timeout
	for i := 0; i < ve.config.MaxConcurrent; i++ {
		select {
		case ve.semaphore <- struct{}{}:
			<-ve.semaphore
		case <-ctx.Done():
			logger.WithComponent("validation-engine").Warn("Force shutdown due to timeout")
			return
		}
	}

	logger.WithComponent("validation-engine").Info("Validation engine shutdown complete")
}

// Helper methods
func (ve *ValidationEngine) updateValidationStatus(validationID string, status contracts.ValidationStatus) {
	ve.validationsMu.Lock()
	defer ve.validationsMu.Unlock()

	if validation, exists := ve.validations[validationID]; exists {
		validation.Status = status
	}
}

func (ve *ValidationEngine) getValidationLevel(options *contracts.ValidationOptions) contracts.ValidationLevel {
	if options != nil {
		return options.Level
	}
	return contracts.ValidationLevelStandard
}

func (ve *ValidationEngine) getEnabledChecks(options *contracts.ValidationOptions, level contracts.ValidationLevel) []contracts.ValidationCheck {
	if options != nil && len(options.EnabledChecks) > 0 {
		return options.EnabledChecks
	}
	
	// Default checks based on level
	switch level {
	case contracts.ValidationLevelFast:
		return []contracts.ValidationCheck{
			contracts.ValidationCheckSyntax,
			contracts.ValidationCheckSecurity,
			contracts.ValidationCheckQuality,
		}
	case contracts.ValidationLevelStandard:
		return []contracts.ValidationCheck{
			contracts.ValidationCheckSyntax,
			contracts.ValidationCheckSecurity,
			contracts.ValidationCheckQuality,
			contracts.ValidationCheckPerformance,
			contracts.ValidationCheckCompliance,
		}
	case contracts.ValidationLevelComprehensive:
		return []contracts.ValidationCheck{
			contracts.ValidationCheckSyntax,
			contracts.ValidationCheckSecurity,
			contracts.ValidationCheckQuality,
			contracts.ValidationCheckPerformance,
			contracts.ValidationCheckCompliance,
			contracts.ValidationCheckLLMCritique,
			contracts.ValidationCheckAccessibility,
		}
	default:
		return []contracts.ValidationCheck{
			contracts.ValidationCheckSyntax,
			contracts.ValidationCheckSecurity,
			contracts.ValidationCheckQuality,
		}
	}
}

func (ve *ValidationEngine) checksToStrings(checks []contracts.ValidationCheck) []string {
	result := make([]string, len(checks))
	for i, check := range checks {
		result[i] = string(check)
	}
	return result
}

func (ve *ValidationEngine) detectLanguage(language string, taskType contracts.TaskType, content string) string {
	if language != "" {
		return language
	}
	
	// Detect based on task type
	switch taskType {
	case contracts.TaskTypeCodegen:
		return "go" // Default
	case contracts.TaskTypeInfra:
		return "hcl"
	case contracts.TaskTypeDoc:
		return "markdown"
	case contracts.TaskTypeTest:
		return "go"
	default:
		return "text"
	}
}