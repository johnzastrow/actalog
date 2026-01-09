package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewBenchmarkHandler tests that the constructor returns a non-nil handler
func TestNewBenchmarkHandler(t *testing.T) {
	handler := NewBenchmarkHandler(nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewBenchmarkHandler should return a non-nil handler")
	}
}

// Authorization tests - these test the handler's authentication/authorization logic

func TestRunBenchmark_Unauthorized(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request without user context (unauthenticated)
	req := createTestRequest(http.MethodPost, "/api/benchmark", "")
	rr := httptest.NewRecorder()

	handler.RunBenchmark(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestGetBenchmarkStatus_Unauthorized(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request without user context (unauthenticated)
	req := createTestRequest(http.MethodGet, "/api/benchmark/status", "")
	rr := httptest.NewRecorder()

	handler.GetBenchmarkStatus(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestCleanupBenchmarkData_Forbidden_NoRole(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request without role in context
	req := createTestRequest(http.MethodDelete, "/api/admin/benchmark/data", "")
	rr := httptest.NewRecorder()

	handler.CleanupBenchmarkData(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestCleanupBenchmarkData_Forbidden_NonAdmin(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with non-admin role
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/benchmark/data", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CleanupBenchmarkData(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

// Note: Success and error path tests for RunBenchmark, GetBenchmarkStatus, and
// CleanupBenchmarkData require a mock BenchmarkService. Currently, the handler
// uses a concrete *service.BenchmarkService type, which prevents interface-based
// mocking. To enable comprehensive unit testing:
//
// 1. Define a BenchmarkServiceInterface in the handler package
// 2. Update BenchmarkHandler to accept the interface
// 3. Create MockBenchmarkService implementing the interface
// 4. Add success/error path tests using the mock
//
// For now, integration tests should be used to verify the full request paths.
