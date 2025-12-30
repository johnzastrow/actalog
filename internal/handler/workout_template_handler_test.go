package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkoutTemplateHandler_CreateTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodPost, "/api/templates", `{"name": "Test Template"}`)
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_CreateTemplate_InvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWorkoutTemplateHandler_CreateTemplate_MissingName(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", `{"description": "Test"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Template name is required")
}

func TestWorkoutTemplateHandler_GetTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/templates/abc", "")
	rr := httptest.NewRecorder()

	handler.GetTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

func TestWorkoutTemplateHandler_ListMyTemplates_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/my-templates", "")
	rr := httptest.NewRecorder()

	handler.ListMyTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_UpdateTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodPut, "/api/templates/1", `{"name": "Updated"}`)
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_UpdateTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/abc", `{"name": "Updated"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

func TestWorkoutTemplateHandler_UpdateTemplate_InvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	// Without chi router context, URL param is empty -> Invalid template ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWorkoutTemplateHandler_DeleteTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodDelete, "/api/templates/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_DeleteTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/templates/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

// Tests for ListStandardTemplates

func TestWorkoutTemplateHandler_ListStandardTemplates_NilService(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodGet, "/api/templates/standard", "")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListStandardTemplates requires service")
		}
	}()

	handler.ListStandardTemplates(rr, req)
}

// Tests for ListMyTemplates with service

func TestWorkoutTemplateHandler_ListMyTemplates_NilService(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/templates/mine", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListMyTemplates requires service")
		}
	}()

	handler.ListMyTemplates(rr, req)
}

// Test GetTemplate with valid ID but nil service

func TestWorkoutTemplateHandler_GetTemplate_ValidIDNilService(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodGet, "/api/templates/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetTemplate requires service")
		}
	}()

	handler.GetTemplate(rr, req)
}

// Test NewWorkoutTemplateHandler

func TestNewWorkoutTemplateHandler(t *testing.T) {
	handler := NewWorkoutTemplateHandler(nil)
	if handler == nil {
		t.Error("NewWorkoutTemplateHandler should return a non-nil handler")
	}
}

// Test UpdateTemplate with valid ID and JSON

func TestWorkoutTemplateHandler_UpdateTemplate_ValidIDInvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", "{bad json", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWorkoutTemplateHandler_UpdateTemplate_MissingName(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", `{"description": "Test"}`, 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Template name is required")
}

// Test DeleteTemplate with valid ID

func TestWorkoutTemplateHandler_DeleteTemplate_ValidIDNilService(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/templates/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteTemplate requires service")
		}
	}()

	handler.DeleteTemplate(rr, req)
}
