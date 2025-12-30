package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImportHandler_PreviewWODImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/wods/preview", "")
	rr := httptest.NewRecorder()

	handler.PreviewWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_PreviewWODImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	// Authenticated but no multipart form
	req := createAuthenticatedRequest(http.MethodPost, "/api/import/wods/preview", "", 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.PreviewWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestImportHandler_ConfirmWODImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/wods/confirm", "")
	rr := httptest.NewRecorder()

	handler.ConfirmWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_ConfirmWODImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/wods/confirm", "", 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ConfirmWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestImportHandler_PreviewMovementImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/movements/preview", "")
	rr := httptest.NewRecorder()

	handler.PreviewMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_PreviewMovementImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/movements/preview", "", 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.PreviewMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestImportHandler_ConfirmMovementImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/movements/confirm", "")
	rr := httptest.NewRecorder()

	handler.ConfirmMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_ConfirmMovementImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/movements/confirm", "", 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ConfirmMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestImportHandler_PreviewUserWorkoutImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/user-workouts/preview", "")
	rr := httptest.NewRecorder()

	handler.PreviewUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_PreviewUserWorkoutImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/user-workouts/preview", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.PreviewUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestImportHandler_ConfirmUserWorkoutImport_Unauthorized(t *testing.T) {
	handler := &ImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/user-workouts/confirm", "")
	rr := httptest.NewRecorder()

	handler.ConfirmUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestImportHandler_ConfirmUserWorkoutImport_NoFile(t *testing.T) {
	handler := &ImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/user-workouts/confirm", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ConfirmUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestNewImportHandler(t *testing.T) {
	handler := NewImportHandler(nil)
	if handler == nil {
		t.Error("NewImportHandler should return a non-nil handler")
	}
}
