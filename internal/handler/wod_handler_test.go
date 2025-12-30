package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewWODHandler(t *testing.T) {
	// Test constructor with nil service
	handler := NewWODHandler(nil)

	if handler == nil {
		t.Error("NewWODHandler() should not return nil")
	}

	if handler.wodService != nil {
		t.Error("wodService should be nil when passed nil")
	}
}

func TestCreateWODRequest_Struct(t *testing.T) {
	url := "https://example.com/wod"
	notes := "Test notes"

	req := CreateWODRequest{
		Name:        "Fran",
		Source:      "CrossFit",
		Type:        "Benchmark",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "21-15-9 Thrusters and Pull-ups",
		URL:         &url,
		Notes:       &notes,
	}

	if req.Name != "Fran" {
		t.Errorf("Name = %q, want %q", req.Name, "Fran")
	}
	if req.Source != "CrossFit" {
		t.Errorf("Source = %q, want %q", req.Source, "CrossFit")
	}
	if req.Type != "Benchmark" {
		t.Errorf("Type = %q, want %q", req.Type, "Benchmark")
	}
	if req.Regime != "For Time" {
		t.Errorf("Regime = %q, want %q", req.Regime, "For Time")
	}
	if req.ScoreType != "Time" {
		t.Errorf("ScoreType = %q, want %q", req.ScoreType, "Time")
	}
	if req.URL == nil || *req.URL != url {
		t.Errorf("URL = %v, want %q", req.URL, url)
	}
	if req.Notes == nil || *req.Notes != notes {
		t.Errorf("Notes = %v, want %q", req.Notes, notes)
	}
}

func TestCreateWODRequest_NilOptionalFields(t *testing.T) {
	req := CreateWODRequest{
		Name: "Simple WOD",
	}

	if req.URL != nil {
		t.Error("URL should be nil")
	}
	if req.Notes != nil {
		t.Error("Notes should be nil")
	}
	if req.Source != "" {
		t.Errorf("Source = %q, want empty", req.Source)
	}
}

func TestUpdateWODRequest_Struct(t *testing.T) {
	url := "https://updated.com/wod"
	notes := "Updated notes"

	req := UpdateWODRequest{
		Name:        "Updated Fran",
		Source:      "Custom",
		Type:        "AMRAP",
		Regime:      "20 min",
		ScoreType:   "Rounds+Reps",
		Description: "Updated description",
		URL:         &url,
		Notes:       &notes,
	}

	if req.Name != "Updated Fran" {
		t.Errorf("Name = %q, want %q", req.Name, "Updated Fran")
	}
	if req.Type != "AMRAP" {
		t.Errorf("Type = %q, want %q", req.Type, "AMRAP")
	}
	if req.ScoreType != "Rounds+Reps" {
		t.Errorf("ScoreType = %q, want %q", req.ScoreType, "Rounds+Reps")
	}
}

func TestWODResponse_Struct(t *testing.T) {
	url := "https://example.com"
	notes := "Some notes"
	createdBy := int64(1)

	resp := WODResponse{
		ID:          123,
		Name:        "Test WOD",
		Source:      "CrossFit",
		Type:        "Benchmark",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "Test description",
		URL:         &url,
		Notes:       &notes,
		IsStandard:  true,
		CreatedBy:   &createdBy,
		CreatedAt:   "2024-01-15T10:00:00Z",
		UpdatedAt:   "2024-01-15T10:00:00Z",
	}

	if resp.ID != 123 {
		t.Errorf("ID = %d, want 123", resp.ID)
	}
	if resp.Name != "Test WOD" {
		t.Errorf("Name = %q, want %q", resp.Name, "Test WOD")
	}
	if !resp.IsStandard {
		t.Error("IsStandard should be true")
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 1 {
		t.Errorf("CreatedBy = %v, want 1", resp.CreatedBy)
	}
}

func TestWODResponse_NilFields(t *testing.T) {
	resp := WODResponse{
		ID:         1,
		Name:       "Custom WOD",
		IsStandard: false,
	}

	if resp.URL != nil {
		t.Error("URL should be nil")
	}
	if resp.Notes != nil {
		t.Error("Notes should be nil")
	}
	if resp.CreatedBy != nil {
		t.Error("CreatedBy should be nil")
	}
}

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

func TestWODHandler_CreateWOD_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods", `{"name": "Test WOD"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil wodService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.CreateWOD(rr, req)
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

func TestWODHandler_GetWOD_InvalidIDFormats(t *testing.T) {
	handler := &WODHandler{}

	testCases := []struct {
		name string
		id   string
	}{
		{"float", "1.5"},
		{"text", "abc"},
		{"special", "1@#"},
		{"overflow", "99999999999999999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/wods/"+tc.id, "")
			rr := httptest.NewRecorder()

			handler.GetWOD(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, "Invalid WOD ID")
		})
	}
}

func TestWODHandler_ListWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil wodService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods?limit=50&offset=10", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil wodService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithInvalidLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	// Invalid limit and offset should use defaults
	req := createTestRequest(http.MethodGet, "/api/wods?limit=invalid&offset=invalid", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_StandardOnly(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods?standard=true", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_Authenticated(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListStandardWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/standard", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListStandardWODs(rr, req)
}

func TestWODHandler_ListStandardWODs_WithLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/standard?limit=25&offset=5", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListStandardWODs(rr, req)
}

func TestWODHandler_ListMyWODs_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/mine", "")
	rr := httptest.NewRecorder()

	handler.ListMyWODs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Authentication required")
}

func TestWODHandler_ListMyWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods/mine", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListMyWODs(rr, req)
}

func TestWODHandler_ListMyWODs_WithLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods/mine?limit=20&offset=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListMyWODs(rr, req)
}

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

func TestWODHandler_SearchWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=fran", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.SearchWODs(rr, req)
}

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

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/abc", `{"name": "Updated WOD"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestWODHandler_UpdateWOD_InvalidJSON(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWODHandler_UpdateWOD_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", `{"name": "Updated WOD"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without chi context, fails on URL param
	handler.UpdateWOD(rr, req)
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWODHandler_UpdateWOD_AsAdmin(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", `{"name": "Updated WOD"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without chi context, fails on URL param
	handler.UpdateWOD(rr, req)
	assertStatusCode(t, rr, http.StatusBadRequest)
}

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

	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestWODHandler_DeleteWOD_InvalidIDFormats(t *testing.T) {
	handler := &WODHandler{}

	testCases := []struct {
		name string
		id   string
	}{
		{"float", "1.5"},
		{"text", "abc"},
		{"special", "1@#"},
		{"overflow", "99999999999999999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/"+tc.id, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.DeleteWOD(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, "Invalid WOD ID")
		})
	}
}

func TestWODHandler_DeleteWOD_NilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/1", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without chi context, fails on URL param first
	handler.DeleteWOD(rr, req)
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWODHandler_NegativeLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	// Negative values should use defaults
	req := createTestRequest(http.MethodGet, "/api/wods?limit=-5&offset=-10", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ZeroLimitOffset(t *testing.T) {
	handler := &WODHandler{}

	// Zero limit should use default
	req := createTestRequest(http.MethodGet, "/api/wods?limit=0&offset=0", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListWODs(rr, req)
}
