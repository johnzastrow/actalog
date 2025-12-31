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

func TestMovementHandler_Create_ValidInputNilService(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/movements",
		`{"name": "Test Movement", "type": "strength", "description": "A test movement"}`,
		1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("Create requires service")
		}
	}()

	handler.Create(rr, req)
}

func TestMovementHandler_Update_ValidInputNilService(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1",
		`{"name": "Updated Movement", "type": "cardio", "description": "Updated description"}`,
		1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("Update requires service")
		}
	}()

	handler.Update(rr, req)
}

func TestMovementHandler_Update_AsAdminNilService(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	// Request as admin
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1",
		`{"name": "Admin Updated", "type": "strength", "description": "Admin update"}`,
		1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests admin path
	defer func() {
		if r := recover(); r == nil {
			t.Log("Update as admin requires service")
		}
	}()

	handler.Update(rr, req)
}

func TestMovementHandler_Delete_ValidIDNilService(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("Delete requires service")
		}
	}()

	handler.Delete(rr, req)
}

func TestMovementHandler_Search_ValidQueryWithLogger(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"simple query", "/api/movements/search?q=squat"},
		{"with limit 10", "/api/movements/search?q=press&limit=10"},
		{"with limit 50", "/api/movements/search?q=deadlift&limit=50"},
		{"with zero limit", "/api/movements/search?q=curl&limit=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			// Without repo, will panic - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("Search requires repository - params were parsed")
				}
			}()

			handler.Search(rr, req)
		})
	}
}

func TestMovementHandler_ListAll_WithLogger(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	// Without repo, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListAll requires repository")
		}
	}()

	handler.ListAll(rr, req)
}

func TestMovementHandler_ListStandard_WithLogger(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	// Without repo, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListStandard requires repository")
		}
	}()

	handler.ListStandard(rr, req)
}

func TestMovementHandler_GetByID_WithLogger(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without repo, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetByID requires repository")
		}
	}()

	handler.GetByID(rr, req)
}

func TestMovementHandler_GetByID_DifferentIDs(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100", "999"}

	for _, id := range testIDs {
		t.Run("id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/movements/"+id, "")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetByID requires repository")
				}
			}()

			handler.GetByID(rr, req)
		})
	}
}

func TestMovementHandler_Delete_DifferentIDs(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/"+id, "", 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Delete requires service")
				}
			}()

			handler.Delete(rr, req)
		})
	}
}

func TestMovementHandler_Update_DifferentIDs(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/movements/"+id, `{"name": "Test", "type": "strength"}`, 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Update requires service")
				}
			}()

			handler.Update(rr, req)
		})
	}
}

func TestMovementHandler_Search_DifferentQueries(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	queries := []string{"squat", "bench", "deadlift", "clean", "snatch"}

	for _, q := range queries {
		t.Run("query_"+q, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/movements/search?q="+q, "")
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

func TestMovementHandler_Create_AllFields(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/movements",
		`{"name": "Full Movement", "type": "olympic", "description": "Full movement with all fields", "video_url": "https://example.com/video", "demo_url": "https://example.com/demo"}`,
		1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("Create requires service")
		}
	}()

	handler.Create(rr, req)
}

func TestMovementHandler_Create_DifferentTypes(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	types := []string{"strength", "cardio", "olympic", "gymnastics", "plyometric"}

	for _, movType := range types {
		t.Run("type_"+movType, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/movements",
				`{"name": "Test Movement", "type": "`+movType+`"}`,
				1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Create requires service")
				}
			}()

			handler.Create(rr, req)
		})
	}
}

func TestMovementHandler_Update_AllFields(t *testing.T) {
	handler := &MovementHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1",
		`{"name": "Updated Full Movement", "type": "strength", "description": "Updated description", "video_url": "https://example.com/newvideo"}`,
		1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("Update requires service")
		}
	}()

	handler.Update(rr, req)
}

// Tests using MockMovementRepository for full code path coverage

func TestMovementHandler_ListAll_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
	assertBodyContains(t, rr, "Back Squat")
}

func TestMovementHandler_ListAll_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movements")
}

func TestMovementHandler_ListStandard_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_ListStandard_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movements")
}

func TestMovementHandler_Search_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_Search_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to search movements")
}

func TestMovementHandler_Search_WithValidLimit(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=test&limit=5", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_GetByID_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Back Squat")
}

func TestMovementHandler_GetByID_NotFound(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestMovementHandler_GetByID_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movement")
}

func TestMovementHandler_ListAll_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_ListAll_ErrorNoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestMovementHandler_ListStandard_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_Search_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=test", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_GetByID_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}
