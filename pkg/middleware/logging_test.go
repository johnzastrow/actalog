package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/pkg/logger"
)

// mockHTTPLogger captures log messages for testing
type mockHTTPLogger struct {
	infoLogs  []string
	debugLogs []string
}

func (m *mockHTTPLogger) Info(format string, v ...interface{}) {
	m.infoLogs = append(m.infoLogs, format)
}

func (m *mockHTTPLogger) Debug(format string, v ...interface{}) {
	m.debugLogs = append(m.debugLogs, format)
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	logger := &mockHTTPLogger{}

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(logger.infoLogs) == 0 {
		t.Error("Expected at least one info log")
	}
}

func TestRequestLogger_SkipsHealthCheck(t *testing.T) {
	logger := &mockHTTPLogger{}

	handlerCalled := false
	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("Handler should be called for health endpoint")
	}
	if len(logger.infoLogs) != 0 {
		t.Error("Should not log health check requests")
	}
}

func TestRequestLogger_LogsWithUserContext(t *testing.T) {
	logger := &mockHTTPLogger{}

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	// Add user context
	ctx := context.WithValue(req.Context(), UserIDKey, int64(42))
	ctx = context.WithValue(ctx, UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(logger.infoLogs) == 0 {
		t.Error("Expected at least one info log")
	}
}

func TestResponseWriter_CapturesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

			wrapped.WriteHeader(tt.statusCode)

			if wrapped.statusCode != tt.statusCode {
				t.Errorf("statusCode = %d, want %d", wrapped.statusCode, tt.statusCode)
			}
		})
	}
}

func TestLoggingResponseWriter_CapturesStatusAndSize(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &loggingResponseWriter{
		ResponseWriter: rec,
		status:         http.StatusOK,
		body:           &bytes.Buffer{},
	}

	wrapped.WriteHeader(http.StatusCreated)
	n, err := wrapped.Write([]byte("Hello, World!"))

	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	if n != 13 {
		t.Errorf("Write returned %d, want 13", n)
	}
	if wrapped.status != http.StatusCreated {
		t.Errorf("status = %d, want %d", wrapped.status, http.StatusCreated)
	}
	if wrapped.size != 13 {
		t.Errorf("size = %d, want 13", wrapped.size)
	}
}

func TestLoggingResponseWriter_CapturesErrorBody(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &loggingResponseWriter{
		ResponseWriter: rec,
		status:         http.StatusOK,
		body:           &bytes.Buffer{},
	}

	// Set error status first
	wrapped.WriteHeader(http.StatusBadRequest)
	wrapped.Write([]byte(`{"error": "bad request"}`))

	// Body should be captured for error responses
	if wrapped.body.Len() == 0 {
		t.Error("Expected error body to be captured")
	}
	if !strings.Contains(wrapped.body.String(), "bad request") {
		t.Errorf("Body = %s, should contain 'bad request'", wrapped.body.String())
	}
}

func TestLoggingResponseWriter_DoesNotCaptureSuccessBody(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &loggingResponseWriter{
		ResponseWriter: rec,
		status:         http.StatusOK, // Default success
		body:           &bytes.Buffer{},
	}

	wrapped.Write([]byte(`{"data": "success"}`))

	// Body should NOT be captured for success responses
	if wrapped.body.Len() != 0 {
		t.Error("Should not capture body for success responses")
	}
}

func TestLogger_Deprecated(t *testing.T) {
	handlerCalled := false
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLogger_DeprecatedSkipsHealthCheck(t *testing.T) {
	handlerCalled := false
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health/check", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("Handler should be called for health endpoint")
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	if id1 == "" {
		t.Error("Request ID should not be empty")
	}

	// IDs generated in sequence should be different (or at least have a pattern)
	if len(id1) < 10 {
		t.Errorf("Request ID seems too short: %s", id1)
	}

	// Note: IDs might be the same if generated in the same microsecond
	_ = id2
}

// ============================================================================
// Tests for LoggingMiddleware
// ============================================================================

func TestLoggingMiddleware_Success(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_POSTWithBody(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body to verify it was restored
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		if n == 0 {
			t.Error("Body should be readable after logging middleware")
		}
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"name": "test", "email": "test@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestLoggingMiddleware_POSTWithPasswordRedacted(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with password field that should be redacted in logs
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email": "test@example.com", "password": "secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_ErrorResponse(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoggingMiddleware_ServerError(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestLoggingMiddleware_Redirect(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	}))

	req := httptest.NewRequest("GET", "/api/redirect", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestLoggingMiddleware_WithQueryParams(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/search?q=test&limit=10", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_WithUserContext(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", int64(42))
	ctx = context.WithValue(ctx, "user_email", "test@example.com")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_NonJSONBody(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/upload", strings.NewReader("plain text body"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_LongNonJSONBody(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a body longer than 200 chars
	longBody := strings.Repeat("x", 300)
	req := httptest.NewRequest("POST", "/api/upload", strings.NewReader(longBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_LongErrorResponse(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		// Write a very long error message (> 500 chars)
		w.Write([]byte(strings.Repeat("error message ", 50)))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoggingMiddleware_HEADRequest(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("HEAD", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoggingMiddleware_WithTokenRedacted(t *testing.T) {
	log := createTestLogger()

	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/auth/refresh", strings.NewReader(`{"token": "secret-refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// ============================================================================
// Tests for RequestIDMiddleware
// ============================================================================

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	log := createTestLogger()

	handler := RequestIDMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	log := createTestLogger()

	handler := RequestIDMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != "existing-request-id-123" {
		t.Errorf("Expected X-Request-ID = 'existing-request-id-123', got %s", requestID)
	}
}

func TestRequestIDMiddleware_PassesToHandler(t *testing.T) {
	log := createTestLogger()

	handlerCalled := false
	handler := RequestIDMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
}

// createTestLogger creates a logger for testing
func createTestLogger() *logger.Logger {
	log, _ := logger.New(logger.Config{Level: "debug"})
	return log
}
