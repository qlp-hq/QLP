package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// GatewayConfig represents the API Gateway configuration
type GatewayConfig struct {
	Port            string            `json:"port"`
	Environment     string            `json:"environment"`
	EnableCORS      bool              `json:"enable_cors"`
	EnableRateLimit bool              `json:"enable_rate_limit"`
	EnableAuth      bool              `json:"enable_auth"`
	Services        map[string]Service `json:"services"`
	RateLimit       RateLimitConfig   `json:"rate_limit"`
	Auth            AuthConfig        `json:"auth"`
	Timeouts        TimeoutConfig     `json:"timeouts"`
	CircuitBreaker  CircuitBreakerConfig `json:"circuit_breaker"`
	LoadBalancer    LoadBalancerConfig `json:"load_balancer"`
}

// Service represents a backend service configuration
type Service struct {
	Name            string          `json:"name"`
	BaseURL         string          `json:"base_url"`
	HealthEndpoint  string          `json:"health_endpoint"`
	Timeout         time.Duration   `json:"timeout"`
	MaxRetries      int             `json:"max_retries"`
	RetryDelay      time.Duration   `json:"retry_delay"`
	CircuitBreaker  bool            `json:"circuit_breaker"`
	LoadBalancing   bool            `json:"load_balancing"`
	Instances       []ServiceInstance `json:"instances"`
	Routes          []Route         `json:"routes"`
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Weight  int    `json:"weight"`
	Healthy bool   `json:"healthy"`
}

// Route represents a route configuration
type Route struct {
	Path         string            `json:"path"`
	Methods      []string          `json:"methods"`
	StripPrefix  bool              `json:"strip_prefix"`
	AddHeaders   map[string]string `json:"add_headers"`
	RemoveHeaders []string         `json:"remove_headers"`
	RateLimit    *RateLimitConfig  `json:"rate_limit,omitempty"`
	Auth         *AuthConfig       `json:"auth,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	RequestsPerSecond int     `json:"requests_per_second"`
	BurstSize   int           `json:"burst_size"`
	WindowSize  time.Duration `json:"window_size"`
	KeyFunc     string        `json:"key_func"` // "ip", "tenant", "user"
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Enabled    bool     `json:"enabled"`
	Type       string   `json:"type"` // "jwt", "api_key", "basic"
	JWTSecret  string   `json:"jwt_secret"`
	JWTIssuer  string   `json:"jwt_issuer"`
	RequiredScopes []string `json:"required_scopes"`
	SkipPaths  []string `json:"skip_paths"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	Read    time.Duration `json:"read"`
	Write   time.Duration `json:"write"`
	Idle    time.Duration `json:"idle"`
	Request time.Duration `json:"request"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled           bool          `json:"enabled"`
	FailureThreshold  int           `json:"failure_threshold"`
	RecoveryTimeout   time.Duration `json:"recovery_timeout"`
	MonitoringPeriod  time.Duration `json:"monitoring_period"`
	MinRequestThreshold int         `json:"min_request_threshold"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Strategy string `json:"strategy"` // "round_robin", "weighted", "least_connections"
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled  bool          `json:"enabled"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Path     string        `json:"path"`
}

// LoadGatewayConfig loads configuration from environment variables
func LoadGatewayConfig() *GatewayConfig {
	config := &GatewayConfig{
		Port:            getEnvOrDefault("GATEWAY_PORT", "8080"),
		Environment:     getEnvOrDefault("ENVIRONMENT", "development"),
		EnableCORS:      getBoolEnvOrDefault("ENABLE_CORS", true),
		EnableRateLimit: getBoolEnvOrDefault("ENABLE_RATE_LIMIT", true),
		EnableAuth:      getBoolEnvOrDefault("ENABLE_AUTH", false),
		Services:        loadServicesConfig(),
		RateLimit:       loadRateLimitConfig(),
		Auth:            loadAuthConfig(),
		Timeouts:        loadTimeoutConfig(),
		CircuitBreaker:  loadCircuitBreakerConfig(),
		LoadBalancer:    loadLoadBalancerConfig(),
	}

	return config
}

// loadServicesConfig loads service configurations
func loadServicesConfig() map[string]Service {
	services := make(map[string]Service)

	// Data Service
	services["data"] = Service{
		Name:           "data-service",
		BaseURL:        getEnvOrDefault("DATA_SERVICE_URL", "http://localhost:8081"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("DATA_SERVICE_TIMEOUT", 30*time.Second),
		MaxRetries:     getIntEnvOrDefault("DATA_SERVICE_MAX_RETRIES", 3),
		RetryDelay:     getDurationEnvOrDefault("DATA_SERVICE_RETRY_DELAY", 1*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("DATA_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/intents",
				Methods:     []string{"GET", "POST", "PUT", "DELETE"},
				StripPrefix: false,
			},
		},
	}

	// Worker Runtime Service
	services["worker"] = Service{
		Name:           "worker-runtime-service",
		BaseURL:        getEnvOrDefault("WORKER_SERVICE_URL", "http://localhost:8082"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("WORKER_SERVICE_TIMEOUT", 60*time.Second),
		MaxRetries:     getIntEnvOrDefault("WORKER_SERVICE_MAX_RETRIES", 2),
		RetryDelay:     getDurationEnvOrDefault("WORKER_SERVICE_RETRY_DELAY", 2*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("WORKER_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/runtime",
				Methods:     []string{"GET", "POST"},
				StripPrefix: false,
			},
		},
	}

	// Packaging Service
	services["packaging"] = Service{
		Name:           "packaging-service",
		BaseURL:        getEnvOrDefault("PACKAGING_SERVICE_URL", "http://localhost:8083"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("PACKAGING_SERVICE_TIMEOUT", 120*time.Second),
		MaxRetries:     getIntEnvOrDefault("PACKAGING_SERVICE_MAX_RETRIES", 2),
		RetryDelay:     getDurationEnvOrDefault("PACKAGING_SERVICE_RETRY_DELAY", 3*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("PACKAGING_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/capsules",
				Methods:     []string{"GET", "POST"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/tenants/{tenantId}/quantum-drops",
				Methods:     []string{"GET", "POST"},
				StripPrefix: false,
			},
		},
	}

	// Orchestrator Service
	services["orchestrator"] = Service{
		Name:           "orchestrator-service",
		BaseURL:        getEnvOrDefault("ORCHESTRATOR_SERVICE_URL", "http://localhost:8084"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("ORCHESTRATOR_SERVICE_TIMEOUT", 300*time.Second),
		MaxRetries:     getIntEnvOrDefault("ORCHESTRATOR_SERVICE_MAX_RETRIES", 1),
		RetryDelay:     getDurationEnvOrDefault("ORCHESTRATOR_SERVICE_RETRY_DELAY", 5*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("ORCHESTRATOR_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/workflows",
				Methods:     []string{"GET", "POST", "PUT", "DELETE"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/dag",
				Methods:     []string{"POST"},
				StripPrefix: false,
			},
		},
	}

	// LLM Service
	services["llm"] = Service{
		Name:           "llm-service",
		BaseURL:        getEnvOrDefault("LLM_SERVICE_URL", "http://localhost:8085"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("LLM_SERVICE_TIMEOUT", 180*time.Second),
		MaxRetries:     getIntEnvOrDefault("LLM_SERVICE_MAX_RETRIES", 2),
		RetryDelay:     getDurationEnvOrDefault("LLM_SERVICE_RETRY_DELAY", 2*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("LLM_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/completion",
				Methods:     []string{"POST"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/tenants/{tenantId}/embedding",
				Methods:     []string{"POST"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/tenants/{tenantId}/chat/completion",
				Methods:     []string{"POST"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/providers",
				Methods:     []string{"GET"},
				StripPrefix: false,
			},
		},
	}

	// Agent Service
	services["agent"] = Service{
		Name:           "agent-service",
		BaseURL:        getEnvOrDefault("AGENT_SERVICE_URL", "http://localhost:8086"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("AGENT_SERVICE_TIMEOUT", 600*time.Second),
		MaxRetries:     getIntEnvOrDefault("AGENT_SERVICE_MAX_RETRIES", 1),
		RetryDelay:     getDurationEnvOrDefault("AGENT_SERVICE_RETRY_DELAY", 5*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("AGENT_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/agents",
				Methods:     []string{"GET", "POST"},
				StripPrefix: false,
			},
			{
				Path:        "/api/v1/tenants/{tenantId}/agents/{agentId}",
				Methods:     []string{"GET", "POST", "PUT", "DELETE"},
				StripPrefix: false,
			},
		},
	}

	// Validation Service
	services["validation"] = Service{
		Name:           "validation-service",
		BaseURL:        getEnvOrDefault("VALIDATION_SERVICE_URL", "http://localhost:8087"),
		HealthEndpoint: "/health",
		Timeout:        getDurationEnvOrDefault("VALIDATION_SERVICE_TIMEOUT", 60*time.Second),
		MaxRetries:     getIntEnvOrDefault("VALIDATION_SERVICE_MAX_RETRIES", 2),
		RetryDelay:     getDurationEnvOrDefault("VALIDATION_SERVICE_RETRY_DELAY", 2*time.Second),
		CircuitBreaker: getBoolEnvOrDefault("VALIDATION_SERVICE_CIRCUIT_BREAKER", true),
		LoadBalancing:  false,
		Routes: []Route{
			{
				Path:        "/api/v1/tenants/{tenantId}/validate",
				Methods:     []string{"POST"},
				StripPrefix: false,
			},
		},
	}

	return services
}

// loadRateLimitConfig loads rate limiting configuration
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:           getBoolEnvOrDefault("RATE_LIMIT_ENABLED", true),
		RequestsPerSecond: getIntEnvOrDefault("RATE_LIMIT_RPS", 100),
		BurstSize:         getIntEnvOrDefault("RATE_LIMIT_BURST", 200),
		WindowSize:        getDurationEnvOrDefault("RATE_LIMIT_WINDOW", 1*time.Minute),
		KeyFunc:           getEnvOrDefault("RATE_LIMIT_KEY_FUNC", "tenant"),
	}
}

// loadAuthConfig loads authentication configuration
func loadAuthConfig() AuthConfig {
	return AuthConfig{
		Enabled:        getBoolEnvOrDefault("AUTH_ENABLED", false),
		Type:           getEnvOrDefault("AUTH_TYPE", "jwt"),
		JWTSecret:      getEnvOrDefault("JWT_SECRET", ""),
		JWTIssuer:      getEnvOrDefault("JWT_ISSUER", "qlp-gateway"),
		RequiredScopes: []string{"read", "write"},
		SkipPaths:      []string{"/health", "/metrics", "/api/v1/status"},
	}
}

// loadTimeoutConfig loads timeout configuration
func loadTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Read:    getDurationEnvOrDefault("TIMEOUT_READ", 30*time.Second),
		Write:   getDurationEnvOrDefault("TIMEOUT_WRITE", 30*time.Second),
		Idle:    getDurationEnvOrDefault("TIMEOUT_IDLE", 60*time.Second),
		Request: getDurationEnvOrDefault("TIMEOUT_REQUEST", 300*time.Second),
	}
}

// loadCircuitBreakerConfig loads circuit breaker configuration
func loadCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:             getBoolEnvOrDefault("CIRCUIT_BREAKER_ENABLED", true),
		FailureThreshold:    getIntEnvOrDefault("CIRCUIT_BREAKER_FAILURE_THRESHOLD", 5),
		RecoveryTimeout:     getDurationEnvOrDefault("CIRCUIT_BREAKER_RECOVERY_TIMEOUT", 30*time.Second),
		MonitoringPeriod:    getDurationEnvOrDefault("CIRCUIT_BREAKER_MONITORING_PERIOD", 10*time.Second),
		MinRequestThreshold: getIntEnvOrDefault("CIRCUIT_BREAKER_MIN_REQUESTS", 3),
	}
}

// loadLoadBalancerConfig loads load balancer configuration
func loadLoadBalancerConfig() LoadBalancerConfig {
	return LoadBalancerConfig{
		Strategy: getEnvOrDefault("LOAD_BALANCER_STRATEGY", "round_robin"),
		HealthCheck: HealthCheckConfig{
			Enabled:  getBoolEnvOrDefault("HEALTH_CHECK_ENABLED", true),
			Interval: getDurationEnvOrDefault("HEALTH_CHECK_INTERVAL", 30*time.Second),
			Timeout:  getDurationEnvOrDefault("HEALTH_CHECK_TIMEOUT", 5*time.Second),
			Path:     getEnvOrDefault("HEALTH_CHECK_PATH", "/health"),
		},
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Validate validates the gateway configuration
func (gc *GatewayConfig) Validate() error {
	if gc.Port == "" {
		return fmt.Errorf("port is required")
	}

	if len(gc.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	for name, service := range gc.Services {
		if service.Name == "" {
			return fmt.Errorf("service %s: name is required", name)
		}
		if service.BaseURL == "" {
			return fmt.Errorf("service %s: base_url is required", name)
		}
		if service.Timeout == 0 {
			return fmt.Errorf("service %s: timeout must be greater than 0", name)
		}
	}

	if gc.EnableAuth && gc.Auth.JWTSecret == "" && gc.Auth.Type == "jwt" {
		return fmt.Errorf("JWT secret is required when JWT authentication is enabled")
	}

	return nil
}