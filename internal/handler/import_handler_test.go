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

// Tests with multipart file upload but missing file field

func TestImportHandler_PreviewWODImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	// Create multipart request with wrong field name
	csvContent := []byte("name,type,score_type,description\nFran,For Time,time,21-15-9")
	req := createMultipartRequest(http.MethodPost, "/api/import/wods/preview",
		"wrongfield", "wods.csv", "text/csv", csvContent, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.PreviewWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

func TestImportHandler_ConfirmWODImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	csvContent := []byte("name,type,score_type,description\nFran,For Time,time,21-15-9")
	req := createMultipartRequest(http.MethodPost, "/api/import/wods/confirm",
		"wrongfield", "wods.csv", "text/csv", csvContent, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ConfirmWODImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

func TestImportHandler_PreviewMovementImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	csvContent := []byte("name,type,description\nBack Squat,weightlifting,Barbell back squat")
	req := createMultipartRequest(http.MethodPost, "/api/import/movements/preview",
		"wrongfield", "movements.csv", "text/csv", csvContent, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.PreviewMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

func TestImportHandler_ConfirmMovementImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	csvContent := []byte("name,type,description\nBack Squat,weightlifting,Barbell back squat")
	req := createMultipartRequest(http.MethodPost, "/api/import/movements/confirm",
		"wrongfield", "movements.csv", "text/csv", csvContent, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ConfirmMovementImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

func TestImportHandler_PreviewUserWorkoutImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	jsonContent := []byte(`{"workouts": []}`)
	req := createMultipartRequest(http.MethodPost, "/api/import/user-workouts/preview",
		"wrongfield", "workouts.json", "application/json", jsonContent, 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.PreviewUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

func TestImportHandler_ConfirmUserWorkoutImport_MissingFileField(t *testing.T) {
	handler := &ImportHandler{}

	jsonContent := []byte(`{"workouts": []}`)
	req := createMultipartRequest(http.MethodPost, "/api/import/user-workouts/confirm",
		"wrongfield", "workouts.json", "application/json", jsonContent, 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ConfirmUserWorkoutImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file uploaded")
}

// Removed 8 panic-expectation tests:
// - TestImportHandler_PreviewWODImport_WithFileNilService
// - TestImportHandler_ConfirmWODImport_WithFileNilService
// - TestImportHandler_PreviewMovementImport_WithFileNilService
// - TestImportHandler_ConfirmMovementImport_WithFileNilService
// - TestImportHandler_PreviewUserWorkoutImport_WithFileNilService
// - TestImportHandler_ConfirmUserWorkoutImport_WithFileNilService
// - TestImportHandler_PreviewWODImport_AsNonAdmin
// - TestImportHandler_PreviewMovementImport_AsNonAdmin
// These tests verified nil pointer panics, not business logic.
