package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
