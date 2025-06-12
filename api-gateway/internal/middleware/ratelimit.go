package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"QLP/api-gateway/pkg/config"
	"QLP/internal/logger"
)

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	config  *config.RateLimitConfig
	buckets sync.Map // map[string]*TokenBucket
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	capacity    int
	tokens      int
	refillRate  int
	lastRefill  time.Time
	mutex       sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config *config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		config: config,
	}
}

// Limit applies rate limiting based on configuration
func (rlm *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rlm.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get rate limit key based on configuration
		key := rlm.getRateLimitKey(r)
		if key == "" {
			// If we can't determine the key, allow the request
			next.ServeHTTP(w, r)
			return
		}

		// Get or create token bucket for this key
		bucket := rlm.getOrCreateBucket(key)

		// Check if request is allowed
		allowed, remaining, resetTime := bucket.Allow()
		if !allowed {
			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(resetTime).Seconds())))

			logger.WithComponent("ratelimit-middleware").Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))

			rlm.writeError(w, http.StatusTooManyRequests, "rate limit exceeded", "RATE_LIMIT_EXCEEDED")
			return
		}

		// Add rate limit headers for successful requests
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerSecond))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		next.ServeHTTP(w, r)
	})
}

// getRateLimitKey determines the rate limit key based on configuration
func (rlm *RateLimitMiddleware) getRateLimitKey(r *http.Request) string {
	switch rlm.config.KeyFunc {
	case "ip":
		return rlm.getClientIP(r)
	case "tenant":
		return rlm.getTenantID(r)
	case "user":
		return rlm.getUserID(r)
	case "api_key":
		return rlm.getAPIKey(r)
	default:
		return rlm.getClientIP(r) // Default to IP-based rate limiting
	}
}

// getClientIP extracts client IP from request
func (rlm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getTenantID extracts tenant ID from request
func (rlm *RateLimitMiddleware) getTenantID(r *http.Request) string {
	// Try to get tenant ID from URL path
	if tenantID := extractTenantFromPath(r.URL.Path); tenantID != "" {
		return tenantID
	}
	
	// Try to get from user context if available
	if userCtx := GetUserFromContext(r.Context()); userCtx != nil {
		return userCtx.TenantID
	}
	
	// Fall back to IP-based rate limiting
	return rlm.getClientIP(r)
}

// getUserID extracts user ID from request context
func (rlm *RateLimitMiddleware) getUserID(r *http.Request) string {
	if userCtx := GetUserFromContext(r.Context()); userCtx != nil {
		return userCtx.UserID
	}
	
	// Fall back to IP-based rate limiting
	return rlm.getClientIP(r)
}

// getAPIKey extracts API key from request
func (rlm *RateLimitMiddleware) getAPIKey(r *http.Request) string {
	// Check Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth
	}
	
	// Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}
	
	// Fall back to IP-based rate limiting
	return rlm.getClientIP(r)
}

// getOrCreateBucket gets existing bucket or creates new one
func (rlm *RateLimitMiddleware) getOrCreateBucket(key string) *TokenBucket {
	if bucket, ok := rlm.buckets.Load(key); ok {
		return bucket.(*TokenBucket)
	}
	
	// Create new bucket
	bucket := &TokenBucket{
		capacity:   rlm.config.BurstSize,
		tokens:     rlm.config.BurstSize,
		refillRate: rlm.config.RequestsPerSecond,
		lastRefill: time.Now(),
	}
	
	// Store and return bucket
	if existing, loaded := rlm.buckets.LoadOrStore(key, bucket); loaded {
		return existing.(*TokenBucket)
	}
	
	return bucket
}

// Allow checks if request is allowed and returns remaining tokens
func (tb *TokenBucket) Allow() (allowed bool, remaining int, resetTime time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	now := time.Now()
	
	// Refill tokens based on time elapsed
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
	
	// Calculate reset time (next refill)
	resetTime = tb.lastRefill.Add(time.Second)
	
	// Check if request is allowed
	if tb.tokens > 0 {
		tb.tokens--
		return true, tb.tokens, resetTime
	}
	
	return false, 0, resetTime
}

// extractTenantFromPath extracts tenant ID from URL path
func extractTenantFromPath(path string) string {
	// Extract tenant ID from paths like /api/v1/tenants/{tenantId}/...
	if len(path) > 20 && path[:20] == "/api/v1/tenants/" {
		parts := strings.Split(path[20:], "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

// Cleanup removes expired buckets to prevent memory leaks
func (rlm *RateLimitMiddleware) Cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				rlm.buckets.Range(func(key, value interface{}) bool {
					bucket := value.(*TokenBucket)
					bucket.mutex.Lock()
					
					// Remove buckets that haven't been used for more than window size
					if now.Sub(bucket.lastRefill) > rlm.config.WindowSize*2 {
						rlm.buckets.Delete(key)
					}
					
					bucket.mutex.Unlock()
					return true
				})
			}
		}
	}()
}

func (rlm *RateLimitMiddleware) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorMsg := fmt.Sprintf(`{"error":"%s","code":"%s","timestamp":"%s"}`, 
		message, code, time.Now().Format(time.RFC3339))
	w.Write([]byte(errorMsg))
}