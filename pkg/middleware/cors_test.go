package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_SetsHeaders(t *testing.T) {
	handler := CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check CORS headers
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods should be set")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers should be set")
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "300" {
		t.Errorf("Access-Control-Max-Age = %q, want %q", got, "300")
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	handlerCalled := false
	handler := CORS([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS request status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if handlerCalled {
		t.Error("Handler should not be called for OPTIONS preflight request")
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	handler := CORS([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://any-origin.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://any-origin.com")
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No Origin header set
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should still work but not set Access-Control-Allow-Origin
	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"https://example.com",
	}

	handler := CORS(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		origin string
	}{
		{"http://localhost:3000"},
		{"http://localhost:8080"},
		{"https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.origin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.origin)
			}
		})
	}
}

func TestCORS_NonAllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://allowed.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://notallowed.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	// A disallowed origin must not receive Access-Control-Allow-Origin.
	// The downstream handler still runs (CORS is a browser-side enforcement),
	// but the missing header causes the browser to discard the response.
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q for disallowed origin, want empty", got)
	}
}
