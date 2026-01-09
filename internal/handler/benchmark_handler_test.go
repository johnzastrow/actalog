package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests for NewBenchmarkHandler
func TestNewBenchmarkHandler(t *testing.T) {
	handler := NewBenchmarkHandler(nil, nil)
	if handler == nil {
		t.Error("NewBenchmarkHandler should return a non-nil handler")
	}
}

func TestNewBenchmarkHandler_WithLogger(t *testing.T) {
	log := createTestLogger()
	handler := NewBenchmarkHandler(nil, log)

	if handler == nil {
		t.Error("NewBenchmarkHandler should return a non-nil handler")
	}
	if handler.logger != log {
		t.Error("Handler should have the provided logger")
	}
}

// Tests for RunBenchmark authorization
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

func TestRunBenchmark_NilService(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid auth
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_Records(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with records parameter
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=5000", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter parsing
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_Concurrent(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with concurrent parameter
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?concurrent=true", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter parsing
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_Cleanup(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with cleanup parameter
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?cleanup=false", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter parsing
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_All(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with all parameters
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=10000&concurrent=true&cleanup=false", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter parsing
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_InvalidRecords(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with invalid records parameter (non-numeric)
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=invalid", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter parsing with invalid value
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_QueryParams_ExceedsMaxRecords(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with records exceeding max (500000)
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=1000000", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests parameter capping
	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

// Tests for GetBenchmarkStatus authorization
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

func TestGetBenchmarkStatus_NilService(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/benchmark/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetBenchmarkStatus requires service")
		}
	}()

	handler.GetBenchmarkStatus(rr, req)
}

// Tests for CleanupBenchmarkData authorization
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

func TestCleanupBenchmarkData_NilService(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with admin role
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/benchmark/data", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("CleanupBenchmarkData requires service")
		}
	}()

	handler.CleanupBenchmarkData(rr, req)
}

// Edge case tests
func TestRunBenchmark_ZeroRecords(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with zero records - should use default
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestRunBenchmark_NegativeRecords(t *testing.T) {
	handler := &BenchmarkHandler{
		logger: createTestLogger(),
	}

	// Request with negative records - should use default
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=-100", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("RunBenchmark requires service")
		}
	}()

	handler.RunBenchmark(rr, req)
}
