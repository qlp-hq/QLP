package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"QLP/api-gateway/pkg/config"
	"QLP/internal/logger"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	config *config.AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
	}
}

// Authenticate validates JWT tokens and adds user context
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled or for skip paths
		if !am.config.Enabled || am.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.writeError(w, http.StatusUnauthorized, "missing authorization header", "MISSING_AUTH_HEADER")
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			am.writeError(w, http.StatusUnauthorized, "invalid authorization header format", "INVALID_AUTH_FORMAT")
			return
		}

		tokenString := parts[1]

		// Validate token based on auth type
		var userContext *UserContext
		var err error

		switch am.config.Type {
		case "jwt":
			userContext, err = am.validateJWT(tokenString)
		case "api_key":
			userContext, err = am.validateAPIKey(tokenString)
		case "basic":
			userContext, err = am.validateBasicAuth(authHeader)
		default:
			am.writeError(w, http.StatusInternalServerError, "unsupported auth type", "UNSUPPORTED_AUTH_TYPE")
			return
		}

		if err != nil {
			logger.WithComponent("auth-middleware").Warn("Authentication failed",
				zap.String("path", r.URL.Path),
				zap.String("auth_type", am.config.Type),
				zap.Error(err))
			am.writeError(w, http.StatusUnauthorized, "authentication failed", "AUTH_FAILED")
			return
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), UserContextKey, userContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateJWT validates JWT tokens
func (am *AuthMiddleware) validateJWT(tokenString string) (*UserContext, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Extract user information from claims
	userContext := &UserContext{
		UserID:   getStringClaim(claims, "sub"),
		Username: getStringClaim(claims, "username"),
		Email:    getStringClaim(claims, "email"),
		TenantID: getStringClaim(claims, "tenant_id"),
		Scopes:   getStringSliceClaim(claims, "scopes"),
		Issuer:   getStringClaim(claims, "iss"),
	}

	// Validate required scopes
	if !am.hasRequiredScopes(userContext.Scopes) {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Validate issuer if specified
	if am.config.JWTIssuer != "" && userContext.Issuer != am.config.JWTIssuer {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return userContext, nil
}

// validateAPIKey validates API key authentication
func (am *AuthMiddleware) validateAPIKey(apiKey string) (*UserContext, error) {
	// In production, this would lookup the API key in a database or external service
	// For now, we'll use a simple validation
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	// Mock validation - in production this would be a proper lookup
	userContext := &UserContext{
		UserID:   "api-user",
		Username: "api-user",
		TenantID: "default",
		Scopes:   []string{"read", "write"},
		AuthType: "api_key",
	}

	return userContext, nil
}

// validateBasicAuth validates basic authentication
func (am *AuthMiddleware) validateBasicAuth(authHeader string) (*UserContext, error) {
	// Extract username and password from Basic auth header
	// In production, this would validate against a user store
	if !strings.HasPrefix(authHeader, "Basic ") {
		return nil, ErrInvalidBasicAuth
	}

	// Mock validation - in production this would be proper credential validation
	userContext := &UserContext{
		UserID:   "basic-user",
		Username: "basic-user",
		TenantID: "default",
		Scopes:   []string{"read"},
		AuthType: "basic",
	}

	return userContext, nil
}

// hasRequiredScopes checks if user has required scopes
func (am *AuthMiddleware) hasRequiredScopes(userScopes []string) bool {
	if len(am.config.RequiredScopes) == 0 {
		return true
	}

	scopeMap := make(map[string]bool)
	for _, scope := range userScopes {
		scopeMap[scope] = true
	}

	for _, requiredScope := range am.config.RequiredScopes {
		if !scopeMap[requiredScope] {
			return false
		}
	}

	return true
}

// shouldSkipPath checks if the path should skip authentication
func (am *AuthMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range am.config.SkipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// Helper functions

func getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func getStringSliceClaim(claims jwt.MapClaims, key string) []string {
	if value, ok := claims[key].([]interface{}); ok {
		var result []string
		for _, v := range value {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}

func (am *AuthMiddleware) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `","code":"` + code + `"}`))
}

// UserContext represents authenticated user information
type UserContext struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	TenantID string   `json:"tenant_id"`
	Scopes   []string `json:"scopes"`
	Issuer   string   `json:"issuer"`
	AuthType string   `json:"auth_type"`
}

// Context key for user context
type contextKey string

const UserContextKey contextKey = "user_context"

// GetUserFromContext extracts user context from request context
func GetUserFromContext(ctx context.Context) *UserContext {
	if user, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return user
	}
	return nil
}

// Custom errors
var (
	ErrInvalidAPIKey    = fmt.Errorf("invalid API key")
	ErrInvalidBasicAuth = fmt.Errorf("invalid basic auth")
)