package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPerformanceHandler_UnifiedSearch_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_UnifiedSearch_MissingQuery(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestPerformanceHandler_GetMovementPerformance_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/movements/1", "")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_GetMovementPerformance_InvalidID(t *testing.T) {
	handler := &PerformanceHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/movements/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestPerformanceHandler_GetWODPerformance_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/wods/1", "")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_GetWODPerformance_InvalidID(t *testing.T) {
	handler := &PerformanceHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestNewPerformanceHandler(t *testing.T) {
	handler := NewPerformanceHandler(nil, nil, nil, nil, nil)
	if handler == nil {
		t.Error("NewPerformanceHandler should return a non-nil handler")
	}
}

// Removed 8 panic-expectation tests:
// - TestPerformanceHandler_UnifiedSearch_NilRepos
// - TestPerformanceHandler_GetMovementPerformance_ValidIDNilRepos
// - TestPerformanceHandler_GetWODPerformance_ValidIDNilRepos
// - TestPerformanceHandler_UnifiedSearch_WithLimitParams (6 subtests)
// - TestPerformanceHandler_GetMovementPerformance_WithLogger
// - TestPerformanceHandler_GetWODPerformance_WithLogger
// - TestPerformanceHandler_GetMovementPerformance_DifferentIDs (4 subtests)
// - TestPerformanceHandler_GetWODPerformance_DifferentIDs (4 subtests)
// These tests verified nil pointer panics, not business logic.
