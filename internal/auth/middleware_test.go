package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAPIKey(t *testing.T) {
	secret := "test-secret-123"
	
	// Create a test handler that will be protected
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized"))
	})
	
	// Wrap it with auth middleware
	protectedHandler := RequireAPIKey(secret)(nextHandler)
	
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer test-secret-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "authorized",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer wrong-secret",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid API key"}`,
		},
		{
			name:           "missing bearer prefix",
			authHeader:     "test-secret-123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing or invalid Authorization header"}`,
		},
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing or invalid Authorization header"}`,
		},
		{
			name:           "bearer with no token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid API key"}`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			
			rr := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			
			body := rr.Body.String()
			if body != tt.expectedBody && body != tt.expectedBody+"\n" {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}
