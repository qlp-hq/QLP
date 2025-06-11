package validation

import (
	"context"
	"errors"
	"math"
	"time"

	"QLP/internal/logger"
	"go.uber.org/zap"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []ValidationErrorCode `json:"retryable_errors"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []ValidationErrorCode{
			ErrorCodeServiceTimeout,
			ErrorCodeTestTimeout,
			ErrorCodeLLMTimeout,
			ErrorCodeLLMQuotaExceeded,
			ErrorCodeResourceExhaustion,
			ErrorCodeServiceStartFailed,
		},
	}
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func(ctx context.Context, attempt int) error

// Retry executes an operation with exponential backoff retry logic
func Retry(ctx context.Context, config *RetryConfig, operation RetryableOperation, component, operationName string) error {
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		logger.WithComponent("validation").Info("Executing operation",
			zap.String("component", component),
			zap.String("operation", operationName),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", config.MaxAttempts))

		err := operation(ctx, attempt)
		if err == nil {
			if attempt > 1 {
				logger.WithComponent("validation").Info("Operation succeeded after retry",
					zap.String("component", component),
					zap.String("operation", operationName),
					zap.Int("attempt", attempt))
			}
			return nil
		}

		lastErr = err

		// Check if the error is retryable
		if !isRetryableValidationError(err, config) {
			logger.WithComponent("validation").Warn("Operation failed with non-retryable error",
				zap.String("component", component),
				zap.String("operation", operationName),
				zap.Int("attempt", attempt),
				zap.Error(err))
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateBackoffDelay(attempt, config)
		
		logger.WithComponent("validation").Warn("Operation failed, retrying",
			zap.String("component", component),
			zap.String("operation", operationName),
			zap.Int("attempt", attempt),
			zap.Duration("retry_delay", delay),
			zap.Error(err))

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	logger.WithComponent("validation").Error("Operation failed after all retry attempts",
		zap.String("component", component),
		zap.String("operation", operationName),
		zap.Int("max_attempts", config.MaxAttempts),
		zap.Error(lastErr))

	return lastErr
}

// isRetryableValidationError determines if an error should be retried
func isRetryableValidationError(err error, config *RetryConfig) bool {
	var ve *ValidationError
	if !errors.As(err, &ve) {
		// For non-ValidationError types, assume not retryable
		return false
	}

	// Check if the error is marked as retryable
	if !ve.IsRetryable() {
		return false
	}

	// Check if the error code is in the retryable list
	for _, code := range config.RetryableErrors {
		if ve.Code == code {
			return true
		}
	}

	return false
}

// calculateBackoffDelay calculates the delay for exponential backoff
func calculateBackoffDelay(attempt int, config *RetryConfig) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))
	
	// Cap at max delay
	if time.Duration(delay) > config.MaxDelay {
		delay = float64(config.MaxDelay)
	}
	
	return time.Duration(delay)
}

// CircuitBreaker implements the circuit breaker pattern for validation operations
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation RetryableOperation, component, operationName string) error {
	// Check circuit state
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// Transition to half-open
			cb.state = CircuitHalfOpen
			logger.WithComponent("validation").Info("Circuit breaker transitioning to half-open",
				zap.String("component", component),
				zap.String("operation", operationName))
		} else {
			return NewValidationError(ErrorCodeServiceTimeout, component, operationName, 
				"circuit breaker is open - too many recent failures").
				WithUserFriendlyMessage("Service is temporarily unavailable due to recent failures. Please try again later.")
		}
	}

	err := operation(ctx, 1)
	
	if err != nil {
		cb.onFailure(component, operationName)
		return err
	}
	
	cb.onSuccess(component, operationName)
	return nil
}

// onFailure handles operation failure
func (cb *CircuitBreaker) onFailure(component, operationName string) {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
		logger.WithComponent("validation").Warn("Circuit breaker opened",
			zap.String("component", component),
			zap.String("operation", operationName),
			zap.Int("failure_count", cb.failureCount),
			zap.Int("max_failures", cb.maxFailures))
	}
}

// onSuccess handles operation success
func (cb *CircuitBreaker) onSuccess(component, operationName string) {
	if cb.state == CircuitHalfOpen {
		// Reset to closed
		cb.state = CircuitClosed
		cb.failureCount = 0
		logger.WithComponent("validation").Info("Circuit breaker closed",
			zap.String("component", component),
			zap.String("operation", operationName))
	}
}

// Timeout wraps an operation with a timeout
func Timeout(ctx context.Context, timeout time.Duration, operation RetryableOperation, component, operationName string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	
	go func() {
		done <- operation(ctx, 1)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return NewValidationError(ErrorCodeTestTimeout, component, operationName, 
				"operation timed out").
				WithDetail("timeout", timeout.String()).
				WithUserFriendlyMessage("The operation took longer than expected and was cancelled.")
		}
		return ctx.Err()
	}
}

// RetryWithCircuitBreaker combines retry logic with circuit breaker pattern
func RetryWithCircuitBreaker(ctx context.Context, retryConfig *RetryConfig, cb *CircuitBreaker, 
	operation RetryableOperation, component, operationName string) error {
	
	return cb.Execute(ctx, func(ctx context.Context, attempt int) error {
		return Retry(ctx, retryConfig, operation, component, operationName)
	}, component, operationName)
}