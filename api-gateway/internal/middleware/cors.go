package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS middleware handler
func CORS(config *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			
			// Set Access-Control-Allow-Origin
			if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set Access-Control-Allow-Methods
			if len(config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			}

			// Set Access-Control-Allow-Headers
			if len(config.AllowedHeaders) == 1 && config.AllowedHeaders[0] == "*" {
				requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
				if requestedHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
				}
			} else if len(config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}

			// Set Access-Control-Expose-Headers
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			// Set Access-Control-Allow-Credentials
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set Access-Control-Max-Age
			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// Check for wildcard patterns
		if strings.Contains(allowed, "*") {
			if matchesPattern(origin, allowed) {
				return true
			}
		}
	}
	return false
}

// matchesPattern checks if origin matches a wildcard pattern
func matchesPattern(origin, pattern string) bool {
	// Simple wildcard matching - in production you might want more sophisticated pattern matching
	if pattern == "*" {
		return true
	}
	
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[2:] // Remove "*."
		return strings.HasSuffix(origin, "."+domain) || origin == domain
	}
	
	return origin == pattern
}