package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPRHandler_GetPersonalRecords_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs", "")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_GetPRMovements_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs/movements", "")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_ToggleMovementPR_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodPut, "/api/prs/toggle?id=1", "")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_ToggleMovementPR_MissingID(t *testing.T) {
	handler := &PRHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Missing movement ID")
}

func TestPRHandler_ToggleMovementPR_InvalidID(t *testing.T) {
	handler := &PRHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}
