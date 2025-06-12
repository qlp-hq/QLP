package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"QLP/api-gateway/pkg/config"
	"QLP/internal/logger"
)

// GatewayHandler handles gateway-specific endpoints
type GatewayHandler struct {
	config      *config.GatewayConfig
	proxyHandler *ProxyHandler
	startTime   time.Time
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(config *config.GatewayConfig, proxyHandler *ProxyHandler) *GatewayHandler {
	return &GatewayHandler{
		config:       config,
		proxyHandler: proxyHandler,
		startTime:    time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]interface{} `json:"services"`
}

// StatusResponse represents the gateway status response
type StatusResponse struct {
	Gateway   GatewayStatus          `json:"gateway"`
	Services  map[string]interface{} `json:"services"`
	Timestamp time.Time              `json:"timestamp"`
}

// GatewayStatus represents gateway-specific status
type GatewayStatus struct {
	Status      string        `json:"status"`
	Version     string        `json:"version"`
	Uptime      time.Duration `json:"uptime"`
	Environment string        `json:"environment"`
	Features    Features      `json:"features"`
}

// Features represents enabled gateway features
type Features struct {
	Authentication bool `json:"authentication"`
	RateLimit      bool `json:"rate_limit"`
	CircuitBreaker bool `json:"circuit_breaker"`
	CORS           bool `json:"cors"`
	HealthChecks   bool `json:"health_checks"`
}

// MetricsResponse represents gateway metrics
type MetricsResponse struct {
	Gateway   GatewayMetrics `json:"gateway"`
	Timestamp time.Time      `json:"timestamp"`
}

// GatewayMetrics represents gateway performance metrics
type GatewayMetrics struct {
	TotalRequests    int64   `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests   int64   `json:"failed_requests"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	ErrorRate        float64 `json:"error_rate_percent"`
}

// HealthCheck handles GET /health
func (gh *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	serviceHealth := gh.proxyHandler.GetServiceHealth()
	
	// Determine overall status
	status := "healthy"
	for _, health := range serviceHealth {
		if healthMap, ok := health.(map[string]interface{}); ok {
			if healthy, exists := healthMap["healthy"].(bool); exists && !healthy {
				status = "degraded"
				break
			}
		}
	}

	response := &HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(gh.startTime).String(),
		Services:  serviceHealth,
	}

	gh.writeJSON(w, http.StatusOK, response)
}

// GetStatus handles GET /api/v1/status
func (gh *GatewayHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	serviceHealth := gh.proxyHandler.GetServiceHealth()

	response := &StatusResponse{
		Gateway: GatewayStatus{
			Status:      "running",
			Version:     "1.0.0",
			Uptime:      time.Since(gh.startTime),
			Environment: gh.config.Environment,
			Features: Features{
				Authentication: gh.config.EnableAuth,
				RateLimit:      gh.config.EnableRateLimit,
				CircuitBreaker: gh.config.CircuitBreaker.Enabled,
				CORS:           gh.config.EnableCORS,
				HealthChecks:   gh.config.LoadBalancer.HealthCheck.Enabled,
			},
		},
		Services:  serviceHealth,
		Timestamp: time.Now(),
	}

	gh.writeJSON(w, http.StatusOK, response)
}

// GetMetrics handles GET /api/v1/metrics
func (gh *GatewayHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// In production, these would be real metrics from a metrics store
	response := &MetricsResponse{
		Gateway: GatewayMetrics{
			TotalRequests:       1000,
			SuccessfulRequests:  950,
			FailedRequests:      50,
			AverageResponseTime: 150.5,
			RequestsPerSecond:   10.2,
			ErrorRate:          5.0,
		},
		Timestamp: time.Now(),
	}

	gh.writeJSON(w, http.StatusOK, response)
}

// GetServices handles GET /api/v1/services
func (gh *GatewayHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]interface{})
	
	for name, service := range gh.config.Services {
		services[name] = map[string]interface{}{
			"name":            service.Name,
			"base_url":        service.BaseURL,
			"health_endpoint": service.HealthEndpoint,
			"timeout":         service.Timeout.String(),
			"max_retries":     service.MaxRetries,
			"circuit_breaker": service.CircuitBreaker,
			"load_balancing":  service.LoadBalancing,
			"routes":          len(service.Routes),
		}
	}

	response := map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now(),
	}

	gh.writeJSON(w, http.StatusOK, response)
}

// GetConfiguration handles GET /api/v1/config (for debugging/monitoring)
func (gh *GatewayHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	// Only expose non-sensitive configuration
	config := map[string]interface{}{
		"port":        gh.config.Port,
		"environment": gh.config.Environment,
		"features": Features{
			Authentication: gh.config.EnableAuth,
			RateLimit:      gh.config.EnableRateLimit,
			CircuitBreaker: gh.config.CircuitBreaker.Enabled,
			CORS:           gh.config.EnableCORS,
			HealthChecks:   gh.config.LoadBalancer.HealthCheck.Enabled,
		},
		"timeouts": map[string]interface{}{
			"read":    gh.config.Timeouts.Read.String(),
			"write":   gh.config.Timeouts.Write.String(),
			"idle":    gh.config.Timeouts.Idle.String(),
			"request": gh.config.Timeouts.Request.String(),
		},
		"rate_limit": map[string]interface{}{
			"enabled":             gh.config.RateLimit.Enabled,
			"requests_per_second": gh.config.RateLimit.RequestsPerSecond,
			"burst_size":          gh.config.RateLimit.BurstSize,
			"window_size":         gh.config.RateLimit.WindowSize.String(),
			"key_func":            gh.config.RateLimit.KeyFunc,
		},
		"circuit_breaker": map[string]interface{}{
			"enabled":              gh.config.CircuitBreaker.Enabled,
			"failure_threshold":    gh.config.CircuitBreaker.FailureThreshold,
			"recovery_timeout":     gh.config.CircuitBreaker.RecoveryTimeout.String(),
			"monitoring_period":    gh.config.CircuitBreaker.MonitoringPeriod.String(),
			"min_request_threshold": gh.config.CircuitBreaker.MinRequestThreshold,
		},
		"load_balancer": map[string]interface{}{
			"strategy": gh.config.LoadBalancer.Strategy,
			"health_check": map[string]interface{}{
				"enabled":  gh.config.LoadBalancer.HealthCheck.Enabled,
				"interval": gh.config.LoadBalancer.HealthCheck.Interval.String(),
				"timeout":  gh.config.LoadBalancer.HealthCheck.Timeout.String(),
				"path":     gh.config.LoadBalancer.HealthCheck.Path,
			},
		},
		"services": len(gh.config.Services),
		"timestamp": time.Now(),
	}

	gh.writeJSON(w, http.StatusOK, config)
}

// NotFound handles 404 errors
func (gh *GatewayHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	logger.WithComponent("gateway-handler").Warn("Route not found",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))

	gh.writeError(w, http.StatusNotFound, "route not found", "ROUTE_NOT_FOUND")
}

// MethodNotAllowed handles 405 errors
func (gh *GatewayHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	logger.WithComponent("gateway-handler").Warn("Method not allowed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))

	gh.writeError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
}

// Helper methods

func (gh *GatewayHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithComponent("gateway-handler").Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (gh *GatewayHandler) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	errorResponse := map[string]interface{}{
		"error":     message,
		"code":      code,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	gh.writeJSON(w, statusCode, errorResponse)
}