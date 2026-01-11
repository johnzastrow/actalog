package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
)

// createTestBenchmarkService creates a BenchmarkService with in-memory database for testing
func createTestBenchmarkService(t *testing.T) (*service.BenchmarkService, func()) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create test user for benchmarks
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "benchtest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		cleanup()
		t.Fatalf("Failed to create test user: %v", err)
	}

	benchRepo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := service.NewBenchmarkService(benchRepo)

	return svc, cleanup
}

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

// Success path tests with real service

func TestRunBenchmark_Success(t *testing.T) {
	svc, cleanup := createTestBenchmarkService(t)
	defer cleanup()

	handler := NewBenchmarkHandler(svc, createTestLogger())

	// Request with valid user context and small record count
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=10&cleanup=true", "", 1, "benchtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RunBenchmark(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	// Verify response contains expected fields
	body := rr.Body.String()
	if !strings.Contains(body, "overall") {
		t.Error("Response should contain 'overall' field")
	}
	if !strings.Contains(body, "database") {
		t.Error("Response should contain 'database' field")
	}
	if !strings.Contains(body, "serialization") {
		t.Error("Response should contain 'serialization' field")
	}
	if !strings.Contains(body, "business_logic") {
		t.Error("Response should contain 'business_logic' field")
	}
	if !strings.Contains(body, "total_operations") {
		t.Error("Response should contain 'total_operations' field")
	}
}

func TestRunBenchmark_WithConcurrent(t *testing.T) {
	svc, cleanup := createTestBenchmarkService(t)
	defer cleanup()

	handler := NewBenchmarkHandler(svc, createTestLogger())

	// Request with concurrent=true
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=5&concurrent=true&cleanup=true", "", 1, "benchtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RunBenchmark(rr, req)

	assertStatusCode(t, rr, http.StatusOK)

	// When concurrent=true, response should contain concurrent field
	body := rr.Body.String()
	if !strings.Contains(body, "concurrent") {
		t.Error("Response should contain 'concurrent' field when concurrent=true")
	}
}

func TestRunBenchmark_WithoutCleanup(t *testing.T) {
	svc, cleanup := createTestBenchmarkService(t)
	defer cleanup()

	handler := NewBenchmarkHandler(svc, createTestLogger())

	// Request with cleanup=false
	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=5&cleanup=false", "", 1, "benchtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RunBenchmark(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

// NOTE: Tests with default or large record counts (0, -1, invalid strings, or > max)
// are skipped because they would use the default record count (1000) which is too slow
// for unit tests. These edge cases are handled by the service layer and tested there.

func TestGetBenchmarkStatus_Success(t *testing.T) {
	svc, cleanup := createTestBenchmarkService(t)
	defer cleanup()

	handler := NewBenchmarkHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/benchmark/status", "", 1, "benchtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetBenchmarkStatus(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	// Verify response contains expected fields
	body := rr.Body.String()
	if !strings.Contains(body, "version") {
		t.Error("Response should contain 'version' field")
	}
	if !strings.Contains(body, "database_driver") {
		t.Error("Response should contain 'database_driver' field")
	}
	if !strings.Contains(body, "record_count") {
		t.Error("Response should contain 'record_count' field")
	}
	if !strings.Contains(body, "status") {
		t.Error("Response should contain 'status' field")
	}
}

func TestCleanupBenchmarkData_Success(t *testing.T) {
	svc, cleanup := createTestBenchmarkService(t)
	defer cleanup()

	handler := NewBenchmarkHandler(svc, createTestLogger())

	// Admin request
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/benchmark/data", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupBenchmarkData(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, "success")
	assertBodyContains(t, rr, "deleted")
}

// Test nil service scenarios
func TestRunBenchmark_NilService(t *testing.T) {
	handler := NewBenchmarkHandler(nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/benchmark?records=10", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil service - we're testing that the handler properly
	// handles the authenticated path up to the service call
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil service")
		}
	}()

	handler.RunBenchmark(rr, req)
}

func TestGetBenchmarkStatus_NilService(t *testing.T) {
	handler := NewBenchmarkHandler(nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/benchmark/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil service
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil service")
		}
	}()

	handler.GetBenchmarkStatus(rr, req)
}

func TestCleanupBenchmarkData_NilService(t *testing.T) {
	handler := NewBenchmarkHandler(nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/benchmark/data", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// This will panic due to nil service
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil service")
		}
	}()

	handler.CleanupBenchmarkData(rr, req)
}
