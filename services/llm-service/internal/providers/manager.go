package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"QLP/services/llm-service/pkg/contracts"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// Provider interface defines the contract for LLM providers
type Provider interface {
	Name() string
	Type() string
	Complete(ctx context.Context, req *contracts.CompletionRequest) (*contracts.CompletionResponse, error)
	GenerateEmbedding(ctx context.Context, req *contracts.EmbeddingRequest) (*contracts.EmbeddingResponse, error)
	ChatCompletion(ctx context.Context, req *contracts.ChatCompletionRequest) (*contracts.ChatCompletionResponse, error)
	HealthCheck(ctx context.Context) error
	GetModels() []contracts.ModelInfo
	IsEnabled() bool
	SetEnabled(enabled bool)
}

// ProviderManager manages multiple LLM providers with fallback support
type ProviderManager struct {
	providers       []Provider
	providerStats   map[string]*ProviderStats
	mu              sync.RWMutex
	healthCheckInternal time.Duration
	startTime       time.Time
}

// ProviderStats tracks statistics for a provider
type ProviderStats struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalTokens     int64
	TotalLatency    time.Duration
	LastError       error
	LastErrorTime   time.Time
	LastSuccessTime time.Time
	Enabled         bool
	Healthy         bool
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers:           []Provider{},
		providerStats:       make(map[string]*ProviderStats),
		healthCheckInternal: 1 * time.Minute,
		startTime:           time.Now(),
	}
}

// RegisterProvider registers a new LLM provider
func (pm *ProviderManager) RegisterProvider(provider Provider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.providers = append(pm.providers, provider)
	pm.providerStats[provider.Name()] = &ProviderStats{
		Enabled: provider.IsEnabled(),
		Healthy: true,
	}

	logger.WithComponent("llm-provider-manager").Info("Provider registered",
		zap.String("name", provider.Name()),
		zap.String("type", provider.Type()),
		zap.Bool("enabled", provider.IsEnabled()))
}

// StartHealthChecks starts periodic health checks for all providers
func (pm *ProviderManager) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(pm.healthCheckInternal)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.performHealthChecks(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// Complete performs text completion with provider fallback
func (pm *ProviderManager) Complete(ctx context.Context, req *contracts.CompletionRequest) (*contracts.CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pm.mu.RLock()
	providers := pm.getEnabledProviders()
	pm.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no enabled providers available")
	}

	var lastErr error
	for _, provider := range providers {
		startTime := time.Now()
		
		response, err := provider.Complete(ctx, req)
		
		pm.recordProviderMetrics(provider.Name(), time.Since(startTime), err)
		
		if err == nil {
			// Add provider information to response
			response.Provider = provider.Name()
			response.ResponseTime = time.Since(startTime)
			response.RequestID = fmt.Sprintf("comp_%d", time.Now().UnixNano())
			
			logger.WithComponent("llm-provider-manager").Info("Completion successful",
				zap.String("provider", provider.Name()),
				zap.Duration("response_time", response.ResponseTime),
				zap.Int("tokens", response.Usage.TotalTokens))
			
			return response, nil
		}

		logger.WithComponent("llm-provider-manager").Warn("Provider completion failed",
			zap.String("provider", provider.Name()),
			zap.Error(err))
		
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// GenerateEmbedding generates embeddings with provider fallback
func (pm *ProviderManager) GenerateEmbedding(ctx context.Context, req *contracts.EmbeddingRequest) (*contracts.EmbeddingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pm.mu.RLock()
	providers := pm.getEnabledProviders()
	pm.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no enabled providers available")
	}

	var lastErr error
	for _, provider := range providers {
		startTime := time.Now()
		
		response, err := provider.GenerateEmbedding(ctx, req)
		
		pm.recordProviderMetrics(provider.Name(), time.Since(startTime), err)
		
		if err == nil {
			// Add provider information to response
			response.Provider = provider.Name()
			response.ResponseTime = time.Since(startTime)
			response.RequestID = fmt.Sprintf("emb_%d", time.Now().UnixNano())
			response.Dimensions = len(response.Embedding)
			
			logger.WithComponent("llm-provider-manager").Info("Embedding successful",
				zap.String("provider", provider.Name()),
				zap.Duration("response_time", response.ResponseTime),
				zap.Int("dimensions", response.Dimensions))
			
			return response, nil
		}

		logger.WithComponent("llm-provider-manager").Warn("Provider embedding failed",
			zap.String("provider", provider.Name()),
			zap.Error(err))
		
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// ChatCompletion performs chat completion with provider fallback
func (pm *ProviderManager) ChatCompletion(ctx context.Context, req *contracts.ChatCompletionRequest) (*contracts.ChatCompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pm.mu.RLock()
	providers := pm.getEnabledProviders()
	pm.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no enabled providers available")
	}

	var lastErr error
	for _, provider := range providers {
		startTime := time.Now()
		
		response, err := provider.ChatCompletion(ctx, req)
		
		pm.recordProviderMetrics(provider.Name(), time.Since(startTime), err)
		
		if err == nil {
			// Add provider information to response
			response.Provider = provider.Name()
			response.ResponseTime = time.Since(startTime)
			response.RequestID = fmt.Sprintf("chat_%d", time.Now().UnixNano())
			
			logger.WithComponent("llm-provider-manager").Info("Chat completion successful",
				zap.String("provider", provider.Name()),
				zap.Duration("response_time", response.ResponseTime),
				zap.Int("tokens", response.Usage.TotalTokens))
			
			return response, nil
		}

		logger.WithComponent("llm-provider-manager").Warn("Provider chat completion failed",
			zap.String("provider", provider.Name()),
			zap.Error(err))
		
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// GetProviderStatus returns the status of all providers
func (pm *ProviderManager) GetProviderStatus() []contracts.ProviderStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var statuses []contracts.ProviderStatus
	for _, provider := range pm.providers {
		stats := pm.providerStats[provider.Name()]
		
		var responseTime time.Duration
		if stats.SuccessRequests > 0 {
			responseTime = time.Duration(stats.TotalLatency.Nanoseconds() / stats.SuccessRequests)
		}

		status := contracts.ProviderStatus{
			Name:         provider.Name(),
			Type:         provider.Type(),
			Available:    provider.IsEnabled(),
			Healthy:      stats.Healthy,
			ResponseTime: responseTime,
			LastCheck:    time.Now(),
			ErrorCount:   int(stats.FailedRequests),
			Metadata: map[string]string{
				"total_requests":   fmt.Sprintf("%d", stats.TotalRequests),
				"success_requests": fmt.Sprintf("%d", stats.SuccessRequests),
				"total_tokens":     fmt.Sprintf("%d", stats.TotalTokens),
			},
		}
		
		statuses = append(statuses, status)
	}

	return statuses
}

// GetMetrics returns aggregated metrics for all providers
func (pm *ProviderManager) GetMetrics() *contracts.MetricsResponse {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var totalRequests, totalTokens, successRequests, failedRequests int64
	var totalLatency time.Duration
	requestsByProvider := make(map[string]int64)
	requestsByModel := make(map[string]int64)

	activeProviders := 0
	for _, provider := range pm.providers {
		if provider.IsEnabled() {
			activeProviders++
		}
		
		stats := pm.providerStats[provider.Name()]
		totalRequests += stats.TotalRequests
		totalTokens += stats.TotalTokens
		successRequests += stats.SuccessRequests
		failedRequests += stats.FailedRequests
		totalLatency += stats.TotalLatency
		
		requestsByProvider[provider.Name()] = stats.TotalRequests
		
		// Add model requests (simplified)
		for _, model := range provider.GetModels() {
			requestsByModel[model.ID] += stats.TotalRequests / int64(len(provider.GetModels()))
		}
	}

	var averageLatency time.Duration
	if successRequests > 0 {
		averageLatency = time.Duration(totalLatency.Nanoseconds() / successRequests)
	}

	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(failedRequests) / float64(totalRequests)
	}

	return &contracts.MetricsResponse{
		TotalRequests:      totalRequests,
		TotalTokens:        totalTokens,
		AverageLatency:     averageLatency,
		ErrorRate:          errorRate,
		ActiveProviders:    activeProviders,
		RequestsByModel:    requestsByModel,
		RequestsByProvider: requestsByProvider,
		Uptime:             time.Since(pm.startTime),
	}
}

// GetActiveModel returns the name of the currently active model
func (pm *ProviderManager) GetActiveModel() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, provider := range pm.providers {
		if provider.IsEnabled() {
			models := provider.GetModels()
			if len(models) > 0 {
				return models[0].ID
			}
		}
	}
	
	return "none"
}

// Helper methods

func (pm *ProviderManager) getEnabledProviders() []Provider {
	var enabled []Provider
	for _, provider := range pm.providers {
		if provider.IsEnabled() && pm.providerStats[provider.Name()].Healthy {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

func (pm *ProviderManager) recordProviderMetrics(providerName string, duration time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stats := pm.providerStats[providerName]
	stats.TotalRequests++
	stats.TotalLatency += duration
	
	if err != nil {
		stats.FailedRequests++
		stats.LastError = err
		stats.LastErrorTime = time.Now()
		
		// Mark as unhealthy after 3 consecutive failures
		if stats.FailedRequests-stats.SuccessRequests >= 3 {
			stats.Healthy = false
		}
	} else {
		stats.SuccessRequests++
		stats.LastSuccessTime = time.Now()
		stats.Healthy = true
	}
}

func (pm *ProviderManager) performHealthChecks(ctx context.Context) {
	pm.mu.Lock()
	providers := make([]Provider, len(pm.providers))
	copy(providers, pm.providers)
	pm.mu.Unlock()

	for _, provider := range providers {
		go func(p Provider) {
			err := p.HealthCheck(ctx)
			
			pm.mu.Lock()
			defer pm.mu.Unlock()
			
			stats := pm.providerStats[p.Name()]
			if err != nil {
				stats.Healthy = false
				stats.LastError = err
				stats.LastErrorTime = time.Now()
				
				logger.WithComponent("llm-provider-manager").Warn("Provider health check failed",
					zap.String("provider", p.Name()),
					zap.Error(err))
			} else {
				stats.Healthy = true
				stats.LastSuccessTime = time.Now()
				
				logger.WithComponent("llm-provider-manager").Debug("Provider health check passed",
					zap.String("provider", p.Name()))
			}
		}(provider)
	}
}