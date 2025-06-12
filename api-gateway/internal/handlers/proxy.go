package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"QLP/api-gateway/pkg/config"
	"QLP/internal/logger"
)

// ProxyHandler handles HTTP proxy requests to backend services
type ProxyHandler struct {
	config   *config.GatewayConfig
	services map[string]*ServiceProxy
}

// ServiceProxy represents a proxy to a specific service
type ServiceProxy struct {
	service *config.Service
	proxy   *httputil.ReverseProxy
	health  *HealthChecker
}

// HealthChecker monitors service health
type HealthChecker struct {
	service     *config.Service
	healthy     bool
	lastCheck   time.Time
	lastError   error
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(config *config.GatewayConfig) *ProxyHandler {
	ph := &ProxyHandler{
		config:   config,
		services: make(map[string]*ServiceProxy),
	}

	// Initialize service proxies
	for name, service := range config.Services {
		ph.services[name] = ph.createServiceProxy(&service)
	}

	// Start health checking
	ph.startHealthChecking()

	return ph
}

// createServiceProxy creates a reverse proxy for a service
func (ph *ProxyHandler) createServiceProxy(service *config.Service) *ServiceProxy {
	// Parse backend URL
	backendURL, err := url.Parse(service.BaseURL)
	if err != nil {
		logger.WithComponent("proxy-handler").Error("Invalid backend URL",
			zap.String("service", service.Name),
			zap.String("url", service.BaseURL),
			zap.Error(err))
		return nil
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Customize the director function
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		ph.modifyRequest(req, service)
	}

	// Add error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		ph.handleProxyError(w, r, service, err)
	}

	// Create health checker
	healthChecker := &HealthChecker{
		service:   service,
		healthy:   true,
		lastCheck: time.Now(),
	}

	return &ServiceProxy{
		service: service,
		proxy:   proxy,
		health:  healthChecker,
	}
}

// ProxyRequest proxies a request to the appropriate backend service
func (ph *ProxyHandler) ProxyRequest(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceProxy, exists := ph.services[serviceName]
		if !exists {
			ph.writeError(w, http.StatusNotFound, "service not found", "SERVICE_NOT_FOUND")
			return
		}

		// Check service health
		if !serviceProxy.health.healthy {
			ph.writeError(w, http.StatusServiceUnavailable, "service unavailable", "SERVICE_UNAVAILABLE")
			return
		}

		// Find matching route
		route := ph.findMatchingRoute(serviceProxy.service, r)
		if route == nil {
			ph.writeError(w, http.StatusNotFound, "route not found", "ROUTE_NOT_FOUND")
			return
		}

		// Apply route-specific modifications
		ph.applyRouteConfig(w, r, route)

		// Add request tracking
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
			r.Header.Set("X-Request-ID", requestID)
		}

		logger.WithComponent("proxy-handler").Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", requestID))

		// Proxy the request
		serviceProxy.proxy.ServeHTTP(w, r)

		// Log completion
		duration := time.Since(start)
		logger.WithComponent("proxy-handler").Info("Request completed",
			zap.String("service", serviceName),
			zap.String("request_id", requestID),
			zap.Duration("duration", duration))
	}
}

// findMatchingRoute finds a route that matches the request
func (ph *ProxyHandler) findMatchingRoute(service *config.Service, r *http.Request) *config.Route {
	for _, route := range service.Routes {
		if ph.routeMatches(&route, r) {
			return &route
		}
	}
	return nil
}

// routeMatches checks if a route matches the request
func (ph *ProxyHandler) routeMatches(route *config.Route, r *http.Request) bool {
	// Check method
	methodMatches := len(route.Methods) == 0
	for _, method := range route.Methods {
		if method == r.Method {
			methodMatches = true
			break
		}
	}
	if !methodMatches {
		return false
	}

	// The route path should match the full request path since that's what we're comparing
	requestPath := r.URL.Path
	routePath := route.Path
	
	// Handle path parameters by doing a more sophisticated match
	// Convert {param} to regex patterns for matching
	regexPath := strings.ReplaceAll(routePath, "{tenantId}", "[^/]+")
	regexPath = strings.ReplaceAll(regexPath, "{agentId}", "[^/]+")
	
	// For now, use a simple approach: check if the paths match structurally
	// by replacing path parameters with wildcards and doing pattern matching
	if strings.Contains(routePath, "{") {
		// Split both paths into segments
		routeSegments := strings.Split(strings.Trim(routePath, "/"), "/")
		requestSegments := strings.Split(strings.Trim(requestPath, "/"), "/")
		
		// Must have same number of segments
		if len(routeSegments) != len(requestSegments) {
			return false
		}
		
		// Check each segment
		for i, routeSegment := range routeSegments {
			if strings.HasPrefix(routeSegment, "{") && strings.HasSuffix(routeSegment, "}") {
				// This is a path parameter, it matches any non-empty value
				if requestSegments[i] == "" {
					return false
				}
			} else {
				// This is a literal segment, must match exactly
				if routeSegment != requestSegments[i] {
					return false
				}
			}
		}
		return true
	}
	
	// Simple prefix match for routes without parameters
	return strings.HasPrefix(requestPath, routePath)
}

// modifyRequest modifies the request before proxying
func (ph *ProxyHandler) modifyRequest(req *http.Request, service *config.Service) {
	// Add service-specific headers
	req.Header.Set("X-Forwarded-Service", service.Name)
	req.Header.Set("X-Gateway-Version", "1.0.0")
	
	// Add timestamp
	req.Header.Set("X-Gateway-Timestamp", time.Now().Format(time.RFC3339))
	
	// Preserve original host
	req.Header.Set("X-Original-Host", req.Host)
}

// applyRouteConfig applies route-specific configuration
func (ph *ProxyHandler) applyRouteConfig(w http.ResponseWriter, r *http.Request, route *config.Route) {
	// Strip prefix if configured
	if route.StripPrefix && strings.HasPrefix(r.URL.Path, route.Path) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Path)
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
	}

	// Add headers
	for key, value := range route.AddHeaders {
		r.Header.Set(key, value)
	}

	// Remove headers
	for _, header := range route.RemoveHeaders {
		r.Header.Del(header)
	}
}

// handleProxyError handles errors from the reverse proxy
func (ph *ProxyHandler) handleProxyError(w http.ResponseWriter, r *http.Request, service *config.Service, err error) {
	logger.WithComponent("proxy-handler").Error("Proxy error",
		zap.String("service", service.Name),
		zap.String("path", r.URL.Path),
		zap.Error(err))

	// Mark service as potentially unhealthy
	if serviceProxy, exists := ph.services[service.Name]; exists {
		serviceProxy.health.lastError = err
	}

	// Return appropriate error response
	ph.writeError(w, http.StatusBadGateway, "backend service error", "BACKEND_ERROR")
}

// startHealthChecking starts background health checking for all services
func (ph *ProxyHandler) startHealthChecking() {
	if !ph.config.LoadBalancer.HealthCheck.Enabled {
		return
	}

	ticker := time.NewTicker(ph.config.LoadBalancer.HealthCheck.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ph.performHealthChecks()
			}
		}
	}()
}

// performHealthChecks checks health of all services
func (ph *ProxyHandler) performHealthChecks() {
	for name, serviceProxy := range ph.services {
		go ph.checkServiceHealth(name, serviceProxy)
	}
}

// checkServiceHealth checks the health of a specific service
func (ph *ProxyHandler) checkServiceHealth(name string, serviceProxy *ServiceProxy) {
	service := serviceProxy.service
	healthURL := service.BaseURL + service.HealthEndpoint

	// Create health check request
	ctx, cancel := context.WithTimeout(context.Background(), ph.config.LoadBalancer.HealthCheck.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		ph.markServiceUnhealthy(serviceProxy, err)
		return
	}

	// Perform health check
	client := &http.Client{
		Timeout: ph.config.LoadBalancer.HealthCheck.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		ph.markServiceUnhealthy(serviceProxy, err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ph.markServiceHealthy(serviceProxy)
	} else {
		ph.markServiceUnhealthy(serviceProxy, fmt.Errorf("health check failed with status %d", resp.StatusCode))
	}
}

// markServiceHealthy marks a service as healthy
func (ph *ProxyHandler) markServiceHealthy(serviceProxy *ServiceProxy) {
	wasUnhealthy := !serviceProxy.health.healthy
	serviceProxy.health.healthy = true
	serviceProxy.health.lastCheck = time.Now()
	serviceProxy.health.lastError = nil

	if wasUnhealthy {
		logger.WithComponent("proxy-handler").Info("Service recovered",
			zap.String("service", serviceProxy.service.Name))
	}
}

// markServiceUnhealthy marks a service as unhealthy
func (ph *ProxyHandler) markServiceUnhealthy(serviceProxy *ServiceProxy, err error) {
	wasHealthy := serviceProxy.health.healthy
	serviceProxy.health.healthy = false
	serviceProxy.health.lastCheck = time.Now()
	serviceProxy.health.lastError = err

	if wasHealthy {
		logger.WithComponent("proxy-handler").Warn("Service became unhealthy",
			zap.String("service", serviceProxy.service.Name),
			zap.Error(err))
	}
}

// GetServiceHealth returns health status of all services
func (ph *ProxyHandler) GetServiceHealth() map[string]interface{} {
	health := make(map[string]interface{})
	
	for name, serviceProxy := range ph.services {
		health[name] = map[string]interface{}{
			"healthy":    serviceProxy.health.healthy,
			"last_check": serviceProxy.health.lastCheck,
			"last_error": serviceProxy.health.lastError,
		}
	}
	
	return health
}

// SetupRoutes sets up all proxy routes
func (ph *ProxyHandler) SetupRoutes(router chi.Router) {
	// Data Service routes
	router.Route("/api/v1/tenants/{tenantId}/intents", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("data")))
	})

	// Worker Runtime Service routes
	router.Route("/api/v1/tenants/{tenantId}/runtime", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("worker")))
	})

	// Packaging Service routes
	router.Route("/api/v1/tenants/{tenantId}/capsules", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("packaging")))
	})
	router.Route("/api/v1/tenants/{tenantId}/quantum-drops", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("packaging")))
	})

	// Orchestrator Service routes
	router.Route("/api/v1/tenants/{tenantId}/workflows", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("orchestrator")))
	})
	router.Route("/api/v1/dag", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("orchestrator")))
	})

	// LLM Service routes
	router.Route("/api/v1/tenants/{tenantId}/completion", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("llm")))
	})
	router.Route("/api/v1/tenants/{tenantId}/embedding", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("llm")))
	})
	router.Route("/api/v1/tenants/{tenantId}/chat", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("llm")))
	})
	router.Route("/api/v1/providers", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("llm")))
	})

	// Agent Service routes
	router.Route("/api/v1/tenants/{tenantId}/agents", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("agent")))
	})

	// Validation Service routes
	router.Route("/api/v1/tenants/{tenantId}/validate", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(ph.ProxyRequest("validation")))
	})
}

func (ph *ProxyHandler) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorMsg := fmt.Sprintf(`{"error":"%s","code":"%s","timestamp":"%s"}`, 
		message, code, time.Now().Format(time.RFC3339))
	w.Write([]byte(errorMsg))
}