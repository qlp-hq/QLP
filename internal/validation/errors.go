package validation

import (
	"errors"
	"fmt"
	"time"
)

// ValidationErrorCode represents standardized error codes for validation failures
type ValidationErrorCode string

const (
	// Build errors
	ErrorCodeBuildFailed         ValidationErrorCode = "BUILD_FAILED"
	ErrorCodeDependencyFailed    ValidationErrorCode = "DEPENDENCY_FAILED"
	ErrorCodeCompilationFailed   ValidationErrorCode = "COMPILATION_FAILED"
	
	// Test errors
	ErrorCodeTestGenerationFailed ValidationErrorCode = "TEST_GENERATION_FAILED"
	ErrorCodeTestExecutionFailed  ValidationErrorCode = "TEST_EXECUTION_FAILED"
	ErrorCodeTestTimeout         ValidationErrorCode = "TEST_TIMEOUT"
	
	// Service errors
	ErrorCodeServiceStartFailed  ValidationErrorCode = "SERVICE_START_FAILED"
	ErrorCodeServiceTimeout      ValidationErrorCode = "SERVICE_TIMEOUT"
	ErrorCodeHealthCheckFailed   ValidationErrorCode = "HEALTH_CHECK_FAILED"
	
	// Security errors
	ErrorCodeSecurityScanFailed  ValidationErrorCode = "SECURITY_SCAN_FAILED"
	ErrorCodeVulnerabilityFound  ValidationErrorCode = "VULNERABILITY_FOUND"
	ErrorCodeComplianceViolation ValidationErrorCode = "COMPLIANCE_VIOLATION"
	
	// Performance errors
	ErrorCodeLoadTestFailed      ValidationErrorCode = "LOAD_TEST_FAILED"
	ErrorCodePerformanceThreshold ValidationErrorCode = "PERFORMANCE_THRESHOLD_EXCEEDED"
	ErrorCodeResourceExhaustion  ValidationErrorCode = "RESOURCE_EXHAUSTION"
	
	// Infrastructure errors
	ErrorCodeExtractionFailed    ValidationErrorCode = "EXTRACTION_FAILED"
	ErrorCodeCleanupFailed       ValidationErrorCode = "CLEANUP_FAILED"
	ErrorCodePermissionDenied    ValidationErrorCode = "PERMISSION_DENIED"
	
	// LLM errors
	ErrorCodeLLMTimeout          ValidationErrorCode = "LLM_TIMEOUT"
	ErrorCodeLLMQuotaExceeded    ValidationErrorCode = "LLM_QUOTA_EXCEEDED"
	ErrorCodeLLMParsingFailed    ValidationErrorCode = "LLM_PARSING_FAILED"
	
	// Configuration errors
	ErrorCodeInvalidConfiguration ValidationErrorCode = "INVALID_CONFIGURATION"
	ErrorCodeMissingRequirement   ValidationErrorCode = "MISSING_REQUIREMENT"
	ErrorCodeUnsupportedFormat    ValidationErrorCode = "UNSUPPORTED_FORMAT"
)

// ValidationError represents a standardized validation error with context
type ValidationError struct {
	Code         ValidationErrorCode `json:"code"`
	Message      string             `json:"message"`
	Component    string             `json:"component"`
	Operation    string             `json:"operation"`
	Timestamp    time.Time          `json:"timestamp"`
	Details      map[string]string  `json:"details,omitempty"`
	Cause        error              `json:"-"`
	Retryable    bool               `json:"retryable"`
	UserFriendly string             `json:"user_friendly,omitempty"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", ve.Code, ve.Component, ve.Message, ve.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", ve.Code, ve.Component, ve.Message)
}

// Unwrap returns the underlying error for error chain support
func (ve *ValidationError) Unwrap() error {
	return ve.Cause
}

// Is supports error comparison with errors.Is
func (ve *ValidationError) Is(target error) bool {
	var targetErr *ValidationError
	if errors.As(target, &targetErr) {
		return ve.Code == targetErr.Code
	}
	return false
}

// NewValidationError creates a new standardized validation error
func NewValidationError(code ValidationErrorCode, component, operation, message string) *ValidationError {
	return &ValidationError{
		Code:      code,
		Message:   message,
		Component: component,
		Operation: operation,
		Timestamp: time.Now(),
		Details:   make(map[string]string),
		Retryable: isRetryableError(code),
	}
}

// WrapValidationError wraps an existing error with validation context
func WrapValidationError(err error, code ValidationErrorCode, component, operation string) *ValidationError {
	ve := NewValidationError(code, component, operation, err.Error())
	ve.Cause = err
	return ve
}

// WithDetail adds a detail field to the error
func (ve *ValidationError) WithDetail(key, value string) *ValidationError {
	ve.Details[key] = value
	return ve
}

// WithUserFriendlyMessage adds a user-friendly message
func (ve *ValidationError) WithUserFriendlyMessage(message string) *ValidationError {
	ve.UserFriendly = message
	return ve
}

// IsRetryable returns whether the error is retryable
func (ve *ValidationError) IsRetryable() bool {
	return ve.Retryable
}

// SetRetryable sets the retryable flag
func (ve *ValidationError) SetRetryable(retryable bool) *ValidationError {
	ve.Retryable = retryable
	return ve
}

// isRetryableError determines if an error code represents a retryable error
func isRetryableError(code ValidationErrorCode) bool {
	retryableCodes := map[ValidationErrorCode]bool{
		ErrorCodeServiceTimeout:      true,
		ErrorCodeTestTimeout:        true,
		ErrorCodeLLMTimeout:         true,
		ErrorCodeLLMQuotaExceeded:   true,
		ErrorCodeResourceExhaustion: true,
		ErrorCodeServiceStartFailed: true, // May succeed on retry
	}
	return retryableCodes[code]
}

// ErrorAggregator collects multiple validation errors
type ErrorAggregator struct {
	errors []*ValidationError
}

// NewErrorAggregator creates a new error aggregator
func NewErrorAggregator() *ErrorAggregator {
	return &ErrorAggregator{
		errors: make([]*ValidationError, 0),
	}
}

// Add adds an error to the aggregator
func (ea *ErrorAggregator) Add(err error) {
	if err == nil {
		return
	}
	
	var ve *ValidationError
	if errors.As(err, &ve) {
		ea.errors = append(ea.errors, ve)
	} else {
		// Convert regular error to ValidationError
		ve = NewValidationError(ErrorCodeMissingRequirement, "unknown", "unknown", err.Error())
		ve.Cause = err
		ea.errors = append(ea.errors, ve)
	}
}

// HasErrors returns true if any errors have been collected
func (ea *ErrorAggregator) HasErrors() bool {
	return len(ea.errors) > 0
}

// Errors returns all collected errors
func (ea *ErrorAggregator) Errors() []*ValidationError {
	return ea.errors
}

// Error implements the error interface, returning a summary of all errors
func (ea *ErrorAggregator) Error() string {
	if len(ea.errors) == 0 {
		return "no errors"
	}
	
	if len(ea.errors) == 1 {
		return ea.errors[0].Error()
	}
	
	return fmt.Sprintf("multiple validation errors (%d total): %s", len(ea.errors), ea.errors[0].Error())
}

// Critical returns only critical errors (non-retryable)
func (ea *ErrorAggregator) Critical() []*ValidationError {
	var critical []*ValidationError
	for _, err := range ea.errors {
		if !err.IsRetryable() {
			critical = append(critical, err)
		}
	}
	return critical
}

// Retryable returns only retryable errors
func (ea *ErrorAggregator) Retryable() []*ValidationError {
	var retryable []*ValidationError
	for _, err := range ea.errors {
		if err.IsRetryable() {
			retryable = append(retryable, err)
		}
	}
	return retryable
}

// ByComponent groups errors by component
func (ea *ErrorAggregator) ByComponent() map[string][]*ValidationError {
	byComponent := make(map[string][]*ValidationError)
	for _, err := range ea.errors {
		byComponent[err.Component] = append(byComponent[err.Component], err)
	}
	return byComponent
}

// Common error constructors for frequently used errors

// ErrBuildFailed creates a standardized build failure error
func ErrBuildFailed(component, details string, cause error) *ValidationError {
	return WrapValidationError(cause, ErrorCodeBuildFailed, component, "build").
		WithDetail("build_details", details).
		WithUserFriendlyMessage("The project failed to build. Please check your code for compilation errors.")
}

// ErrTestFailed creates a standardized test failure error
func ErrTestFailed(component, testName string, cause error) *ValidationError {
	return WrapValidationError(cause, ErrorCodeTestExecutionFailed, component, "test").
		WithDetail("test_name", testName).
		WithUserFriendlyMessage("Tests failed to execute. Please review your test cases.")
}

// ErrServiceStartFailed creates a standardized service start failure error
func ErrServiceStartFailed(component, serviceType string, cause error) *ValidationError {
	return WrapValidationError(cause, ErrorCodeServiceStartFailed, component, "service_start").
		WithDetail("service_type", serviceType).
		WithUserFriendlyMessage("The service failed to start. Please check your configuration and dependencies.")
}

// ErrSecurityViolation creates a standardized security violation error
func ErrSecurityViolation(component, violation string) *ValidationError {
	return NewValidationError(ErrorCodeVulnerabilityFound, component, "security_scan", violation).
		WithDetail("violation_type", violation).
		WithUserFriendlyMessage("Security vulnerabilities were found. Please address these issues before deployment.").
		SetRetryable(false)
}

// ErrLLMProcessing creates a standardized LLM processing error
func ErrLLMProcessing(component, operation string, cause error) *ValidationError {
	return WrapValidationError(cause, ErrorCodeLLMParsingFailed, component, operation).
		WithUserFriendlyMessage("AI analysis encountered an issue. This may be retried automatically.")
}