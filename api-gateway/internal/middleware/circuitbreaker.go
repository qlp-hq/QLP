package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"QLP/api-gateway/pkg/config"
	"QLP/internal/logger"
)

// CircuitBreakerMiddleware provides circuit breaker functionality
type CircuitBreakerMiddleware struct {
	config   *config.CircuitBreakerConfig
	breakers sync.Map // map[string]*CircuitBreaker
}

// CircuitBreaker represents a circuit breaker for a service
type CircuitBreaker struct {
	name            string
	state           CircuitBreakerState
	failures        int
	requests        int
	lastFailureTime time.Time
	lastRequestTime time.Time
	config          *config.CircuitBreakerConfig
	mutex           sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware
func NewCircuitBreakerMiddleware(config *config.CircuitBreakerConfig) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		config: config,
	}
}

// Protect wraps a handler with circuit breaker protection
func (cbm *CircuitBreakerMiddleware) Protect(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cbm.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			cb := cbm.getOrCreateCircuitBreaker(serviceName)

			// Check if circuit breaker allows the request
			if !cb.CanExecute() {
				logger.WithComponent("circuitbreaker-middleware").Warn("Circuit breaker open",
					zap.String("service", serviceName),
					zap.String("path", r.URL.Path),
					zap.String("state", cb.GetState().String()))

				cbm.writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable", "CIRCUIT_BREAKER_OPEN")
				return
			}

			// Create a response recorder to capture the response
			recorder := &ResponseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Execute the request
			start := time.Now()
			next.ServeHTTP(recorder, r)
			duration := time.Since(start)

			// Report the result to circuit breaker
			if recorder.statusCode >= 500 {
				cb.RecordFailure()
				logger.WithComponent("circuitbreaker-middleware").Warn("Request failed",
					zap.String("service", serviceName),
					zap.Int("status_code", recorder.statusCode),
					zap.Duration("duration", duration))
			} else {
				cb.RecordSuccess()
			}
		})
	}
}

// getOrCreateCircuitBreaker gets existing circuit breaker or creates new one
func (cbm *CircuitBreakerMiddleware) getOrCreateCircuitBreaker(serviceName string) *CircuitBreaker {
	if cb, ok := cbm.breakers.Load(serviceName); ok {
		return cb.(*CircuitBreaker)
	}

	// Create new circuit breaker
	cb := &CircuitBreaker{
		name:            serviceName,
		state:           StateClosed,
		config:          cbm.config,
		lastRequestTime: time.Now(),
	}

	// Store and return circuit breaker
	if existing, loaded := cbm.breakers.LoadOrStore(serviceName, cb); loaded {
		return existing.(*CircuitBreaker)
	}

	return cb
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if recovery timeout has passed
		if now.Sub(cb.lastFailureTime) >= cb.config.RecoveryTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			// Double-check after acquiring write lock
			if cb.state == StateOpen && now.Sub(cb.lastFailureTime) >= cb.config.RecoveryTimeout {
				cb.state = StateHalfOpen
				cb.requests = 0
				cb.failures = 0
			}
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return cb.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.lastRequestTime = time.Now()
	cb.requests++

	if cb.state == StateHalfOpen {
		// If we're in half-open state and getting successes, close the circuit
		cb.state = StateClosed
		cb.failures = 0
		cb.requests = 0
		
		logger.WithComponent("circuitbreaker").Info("Circuit breaker closed",
			zap.String("service", cb.name))
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.lastRequestTime = time.Now()
	cb.lastFailureTime = time.Now()
	cb.requests++
	cb.failures++

	// Check if we should open the circuit
	if cb.shouldOpen() {
		cb.state = StateOpen
		
		logger.WithComponent("circuitbreaker").Warn("Circuit breaker opened",
			zap.String("service", cb.name),
			zap.Int("failures", cb.failures),
			zap.Int("requests", cb.requests))
	}
}

// shouldOpen determines if the circuit breaker should open
func (cb *CircuitBreaker) shouldOpen() bool {
	// Only consider opening if we have enough requests
	if cb.requests < cb.config.MinRequestThreshold {
		return false
	}

	// Check if failure rate exceeds threshold
	failureRate := float64(cb.failures) / float64(cb.requests)
	threshold := float64(cb.config.FailureThreshold) / 100.0

	return failureRate >= threshold
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.state.String(),
		"failures":          cb.failures,
		"requests":          cb.requests,
		"last_failure_time": cb.lastFailureTime,
		"last_request_time": cb.lastRequestTime,
	}
}

// ResponseRecorder captures response details for circuit breaker evaluation
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures status code is set if not already set
func (rr *ResponseRecorder) Write(data []byte) (int, error) {
	if rr.statusCode == 0 {
		rr.statusCode = http.StatusOK
	}
	return rr.ResponseWriter.Write(data)
}

// GetServiceStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerMiddleware) GetServiceStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	cbm.breakers.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		cb := value.(*CircuitBreaker)
		stats[serviceName] = cb.GetStats()
		return true
	})
	
	return stats
}

// Cleanup removes stale circuit breakers to prevent memory leaks
func (cbm *CircuitBreakerMiddleware) Cleanup() {
	ticker := time.NewTicker(cbm.config.MonitoringPeriod)
	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				cbm.breakers.Range(func(key, value interface{}) bool {
					cb := value.(*CircuitBreaker)
					cb.mutex.RLock()
					
					// Remove circuit breakers that haven't been used for a long time
					if now.Sub(cb.lastRequestTime) > cbm.config.MonitoringPeriod*10 {
						cb.mutex.RUnlock()
						cbm.breakers.Delete(key)
						return true
					}
					
					// Reset counters periodically for closed circuit breakers
					if cb.state == StateClosed && now.Sub(cb.lastRequestTime) > cbm.config.MonitoringPeriod {
						cb.mutex.RUnlock()
						cb.mutex.Lock()
						cb.requests = 0
						cb.failures = 0
						cb.mutex.Unlock()
						return true
					}
					
					cb.mutex.RUnlock()
					return true
				})
			}
		}
	}()
}

func (cbm *CircuitBreakerMiddleware) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorMsg := fmt.Sprintf(`{"error":"%s","code":"%s","timestamp":"%s"}`, 
		message, code, time.Now().Format(time.RFC3339))
	w.Write([]byte(errorMsg))
}