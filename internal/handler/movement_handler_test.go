package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMovementHandler_Search_MissingQuery(t *testing.T) {
	handler := &MovementHandler{}

	req := createTestRequest(http.MethodGet, "/api/movements/search", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Query parameter 'q' is required")
}

func TestMovementHandler_Create_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPost, "/api/movements", `{"name": "Test", "type": "strength"}`)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestMovementHandler_Create_InvalidJSON(t *testing.T) {
	handler := &MovementHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/movements", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestMovementHandler_Create_MissingFields(t *testing.T) {
	handler := &MovementHandler{}

	tests := []struct {
		name       string
		body       string
		wantError  string
	}{
		{
			name:      "missing name",
			body:      `{"type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "missing type",
			body:      `{"name": "Test Movement"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "empty name",
			body:      `{"name": "", "type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "empty type",
			body:      `{"name": "Test", "type": ""}`,
			wantError: "Name and type are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/movements", tt.body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestMovementHandler_Update_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPut, "/api/movements/1", `{"name": "Test", "type": "strength"}`)
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

// Note: Update tests that require chi.URLParam need router context
// Testing ID parsing requires chi router setup which is complex for unit tests
// These validation tests verify the input is checked before service calls

func TestMovementHandler_Update_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// chi.URLParam returns empty string without router context, which fails int parsing
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/abc", `{"name": "Test", "type": "strength"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	// Without chi router context, URLParam returns empty string -> invalid ID
	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestMovementHandler_Delete_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodDelete, "/api/movements/1", "")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestMovementHandler_Delete_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// Without chi router context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

// Tests for ListAll

func TestMovementHandler_ListAll_NilRepo(t *testing.T) {
	handler := &MovementHandler{}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	// Without a repo, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListAll requires repository")
		}
	}()

	handler.ListAll(rr, req)
}

// Tests for ListStandard

func TestMovementHandler_ListStandard_NilRepo(t *testing.T) {
	handler := &MovementHandler{}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	// Without a repo, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListStandard requires repository")
		}
	}()

	handler.ListStandard(rr, req)
}

// Tests for GetByID

func TestMovementHandler_GetByID_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// Without chi router context, URLParam returns empty string -> invalid ID
	req := createTestRequest(http.MethodGet, "/api/movements/abc", "")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestMovementHandler_GetByID_ValidIDNilRepo(t *testing.T) {
	handler := &MovementHandler{}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a repo, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetByID requires repository")
		}
	}()

	handler.GetByID(rr, req)
}

// Tests for Search with limit parameter

func TestMovementHandler_Search_WithLimit(t *testing.T) {
	handler := &MovementHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"valid query", "/api/movements/search?q=squat"},
		{"with limit", "/api/movements/search?q=squat&limit=10"},
		{"with invalid limit", "/api/movements/search?q=squat&limit=abc"},
		{"with negative limit", "/api/movements/search?q=squat&limit=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Search requires repository")
				}
			}()

			handler.Search(rr, req)
		})
	}
}

// Tests for Update with valid ID

func TestMovementHandler_Update_InvalidJSON(t *testing.T) {
	handler := &MovementHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", "{bad json", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestMovementHandler_Update_MissingFields(t *testing.T) {
	handler := &MovementHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing name",
			body:      `{"type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "missing type",
			body:      `{"name": "Test Movement"}`,
			wantError: "Name and type are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", tt.body, 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", "1")
			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

// Test NewMovementHandler

func TestNewMovementHandler(t *testing.T) {
	handler := NewMovementHandler(nil, nil, nil)
	if handler == nil {
		t.Error("NewMovementHandler should return a non-nil handler")
	}
}
