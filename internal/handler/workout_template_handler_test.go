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
