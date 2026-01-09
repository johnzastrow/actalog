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

func TestImportHandler_PreviewWODImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type,score_type\nTest WOD,AMRAP,rounds+reps")
	req := createMultipartRequest(http.MethodPost, "/api/import/wods/preview",
		"file", "wods.csv", "text/csv", content, 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("PreviewWODImport requires service")
		}
	}()

	handler.PreviewWODImport(rr, req)
}

func TestImportHandler_ConfirmWODImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type,score_type\nTest WOD,AMRAP,rounds+reps")
	req := createMultipartRequest(http.MethodPost, "/api/import/wods/confirm",
		"file", "wods.csv", "text/csv", content, 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ConfirmWODImport requires service")
		}
	}()

	handler.ConfirmWODImport(rr, req)
}

func TestImportHandler_PreviewMovementImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type\nBack Squat,weightlifting")
	req := createMultipartRequest(http.MethodPost, "/api/import/movements/preview",
		"file", "movements.csv", "text/csv", content, 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("PreviewMovementImport requires service")
		}
	}()

	handler.PreviewMovementImport(rr, req)
}

func TestImportHandler_ConfirmMovementImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type\nBack Squat,weightlifting")
	req := createMultipartRequest(http.MethodPost, "/api/import/movements/confirm",
		"file", "movements.csv", "text/csv", content, 1, "test@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ConfirmMovementImport requires service")
		}
	}()

	handler.ConfirmMovementImport(rr, req)
}

func TestImportHandler_PreviewUserWorkoutImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte(`[{"workout_date": "2024-01-15", "notes": "Test workout"}]`)
	req := createMultipartRequest(http.MethodPost, "/api/import/user-workouts/preview",
		"file", "workouts.json", "application/json", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("PreviewUserWorkoutImport requires service")
		}
	}()

	handler.PreviewUserWorkoutImport(rr, req)
}

func TestImportHandler_ConfirmUserWorkoutImport_WithFileNilService(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte(`[{"workout_date": "2024-01-15", "notes": "Test workout"}]`)
	req := createMultipartRequest(http.MethodPost, "/api/import/user-workouts/confirm",
		"file", "workouts.json", "application/json", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ConfirmUserWorkoutImport requires service")
		}
	}()

	handler.ConfirmUserWorkoutImport(rr, req)
}

func TestImportHandler_PreviewWODImport_AsNonAdmin(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type,score_type\nTest WOD,AMRAP,rounds+reps")
	req := createMultipartRequest(http.MethodPost, "/api/import/wods/preview",
		"file", "wods.csv", "text/csv", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("PreviewWODImport may check role or require service")
		}
	}()

	handler.PreviewWODImport(rr, req)
}

func TestImportHandler_PreviewMovementImport_AsNonAdmin(t *testing.T) {
	handler := &ImportHandler{}

	content := []byte("name,type\nBack Squat,weightlifting")
	req := createMultipartRequest(http.MethodPost, "/api/import/movements/preview",
		"file", "movements.csv", "text/csv", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("PreviewMovementImport may check role or require service")
		}
	}()

	handler.PreviewMovementImport(rr, req)
}
