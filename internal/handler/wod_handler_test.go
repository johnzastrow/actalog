package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWODHandler_CreateWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPost, "/api/wods", `{"name": "Test WOD"}`)
	rr := httptest.NewRecorder()

	handler.CreateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_CreateWOD_InvalidJSON(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWODHandler_CreateWOD_MissingName(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "empty name",
			body:      `{"name": ""}`,
			wantError: "WOD name is required",
		},
		{
			name:      "missing name field",
			body:      `{"source": "CrossFit"}`,
			wantError: "WOD name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/wods", tt.body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.CreateWOD(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestWODHandler_GetWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi router context, URLParam returns empty string
	req := createTestRequest(http.MethodGet, "/api/wods/abc", "")
	rr := httptest.NewRecorder()

	handler.GetWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}
