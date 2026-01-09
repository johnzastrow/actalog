package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPRHandler(t *testing.T) {
	handler := NewPRHandler(nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewPRHandler() should not return nil")
	}
}

// Removed struct field assignment tests:
// - TestPersonalRecord_Struct, TestPersonalRecord_NilFields
// - TestMovementPRSummary_Struct, TestMovementPRSummary_NilFields
// These tests verified Go struct assignment works, not business logic.

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

func TestPRHandler_ToggleMovementPR_InvalidIDFormats(t *testing.T) {
	handler := &PRHandler{}

	testCases := []struct {
		name string
		id   string
	}{
		{"float", "1.5"},
		{"empty", ""},
		{"special chars", "1@#$"},
		{"overflow", "99999999999999999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/prs/toggle"
			if tc.id != "" {
				url += "?id=" + tc.id
			}
			req := createAuthenticatedRequest(http.MethodPut, url, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ToggleMovementPR(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
		})
	}
}

// Removed 12 panic-expectation tests:
// - TestPRHandler_GetPersonalRecords_WithLimitParam
// - TestPRHandler_GetPersonalRecords_WithInvalidLimitParam
// - TestPRHandler_GetPersonalRecords_WithNegativeLimitParam
// - TestPRHandler_GetPersonalRecords_WithHighLimitParam
// - TestPRHandler_GetPRMovements_WithLimitParam
// - TestPRHandler_GetPRMovements_WithInvalidLimitParam
// - TestPRHandler_GetPRMovements_WithHighLimitParam
// - TestPRHandler_ToggleMovementPR_NilDB
// - TestPRHandler_GetPersonalRecords_NilDB
// - TestPRHandler_GetPRMovements_NilDB
// - TestPRHandler_ToggleMovementPR_ZeroID
// - TestPRHandler_GetPersonalRecords_ZeroLimit
// - TestPRHandler_GetPRMovements_ZeroLimit
// These tests verified nil pointer panics, not business logic.
