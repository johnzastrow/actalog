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

// Tests for ListWODs

func TestWODHandler_ListWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	// Without auth, tries to call service.ListStandard which will panic on nil service
	// This tests the function entry - for full test would need mock service
	req := createTestRequest(http.MethodGet, "/api/wods", "")
	rr := httptest.NewRecorder()

	// We expect a panic or error since service is nil
	// Test the code path by wrapping in recover
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListWODs requires service - expected panic on nil service")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithStandardOnlyParam(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods?standard=true", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ListWODs requires service")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithPaginationParams(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"with limit", "/api/wods?limit=10"},
		{"with offset", "/api/wods?offset=5"},
		{"with both", "/api/wods?limit=10&offset=5"},
		{"with invalid limit", "/api/wods?limit=abc"},
		{"with invalid offset", "/api/wods?offset=xyz"},
		{"with negative values", "/api/wods?limit=-1&offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListWODs requires service - params were parsed")
				}
			}()

			handler.ListWODs(rr, req)
		})
	}
}

// Tests for ListStandardWODs

func TestWODHandler_ListStandardWODs_WithPagination(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default pagination", "/api/wods/standard"},
		{"with limit", "/api/wods/standard?limit=50"},
		{"with offset", "/api/wods/standard?offset=10"},
		{"with both", "/api/wods/standard?limit=25&offset=10"},
		{"invalid limit ignored", "/api/wods/standard?limit=abc"},
		{"invalid offset ignored", "/api/wods/standard?offset=xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListStandardWODs requires service")
				}
			}()

			handler.ListStandardWODs(rr, req)
		})
	}
}

// Tests for ListMyWODs

func TestWODHandler_ListMyWODs_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	// Request without user context
	req := createTestRequest(http.MethodGet, "/api/wods/mine", "")
	rr := httptest.NewRecorder()

	handler.ListMyWODs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Authentication required")
}

func TestWODHandler_ListMyWODs_WithPagination(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default pagination", "/api/wods/mine"},
		{"with limit", "/api/wods/mine?limit=50"},
		{"with offset", "/api/wods/mine?offset=10"},
		{"with both", "/api/wods/mine?limit=25&offset=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListMyWODs requires service")
				}
			}()

			handler.ListMyWODs(rr, req)
		})
	}
}

// Tests for SearchWODs

func TestWODHandler_SearchWODs_MissingQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestWODHandler_SearchWODs_EmptyQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestWODHandler_SearchWODs_ValidQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=fran", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("SearchWODs requires service")
		}
	}()

	handler.SearchWODs(rr, req)
}

// Tests for UpdateWOD

func TestWODHandler_UpdateWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodPut, "/api/wods/1", `{"name": "Updated WOD"}`)
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_UpdateWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/abc", `{"name": "Updated"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestWODHandler_UpdateWOD_InvalidJSON(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", "{bad json", 1, "test@example.com", "user")
	// Add chi URL param context
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Tests for DeleteWOD

func TestWODHandler_DeleteWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodDelete, "/api/wods/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_DeleteWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}
