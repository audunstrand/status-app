package auth

import (
	"net/http"
	"strings"
)

// RequireAPIKey is a middleware that validates the API secret in the Authorization header
func RequireAPIKey(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			
			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"Missing or invalid Authorization header"}`, http.StatusUnauthorized)
				return
			}
			
			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Validate token matches secret
			if token != secret {
				http.Error(w, `{"error":"Invalid API key"}`, http.StatusUnauthorized)
				return
			}
			
			// Token is valid, continue
			next.ServeHTTP(w, r)
		})
	}
}
