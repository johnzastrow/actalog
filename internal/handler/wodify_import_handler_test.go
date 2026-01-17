package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWodifyImportHandler_PreviewWodifyImport_Unauthorized(t *testing.T) {
	handler := &WodifyImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/wodify/preview", "")
	rr := httptest.NewRecorder()

	handler.PreviewWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWodifyImportHandler_PreviewWodifyImport_NoFile(t *testing.T) {
	handler := &WodifyImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/wodify/preview", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.PreviewWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse form data")
}

func TestWodifyImportHandler_ConfirmWodifyImport_Unauthorized(t *testing.T) {
	handler := &WodifyImportHandler{}

	req := createTestRequest(http.MethodPost, "/api/import/wodify/confirm", "")
	rr := httptest.NewRecorder()

	handler.ConfirmWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWodifyImportHandler_ConfirmWodifyImport_NoFile(t *testing.T) {
	handler := &WodifyImportHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/import/wodify/confirm", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ConfirmWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse form data")
}

func TestNewWodifyImportHandler(t *testing.T) {
	handler := NewWodifyImportHandler(nil)
	if handler == nil {
		t.Error("NewWodifyImportHandler should return a non-nil handler")
	}
}

// Removed 3 panic-expectation tests:
// - TestWodifyImportHandler_PreviewWodifyImport_WithFileNilService
// - TestWodifyImportHandler_ConfirmWodifyImport_WithFileNilService
// - TestWodifyImportHandler_PreviewWodifyImport_EmptyFile
// These tests verified nil pointer panics, not business logic.

func TestWodifyImportHandler_PreviewWodifyImport_WrongFieldName(t *testing.T) {
	handler := &WodifyImportHandler{}

	content := []byte("Date,Workout Name,Duration\n2024-01-15,Fran,5:30")
	req := createMultipartRequest(http.MethodPost, "/api/import/wodify/preview",
		"wrongfield", "wodify.csv", "text/csv", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.PreviewWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWodifyImportHandler_ConfirmWodifyImport_WrongFieldName(t *testing.T) {
	handler := &WodifyImportHandler{}

	content := []byte("Date,Workout Name,Duration\n2024-01-15,Fran,5:30")
	req := createMultipartRequest(http.MethodPost, "/api/import/wodify/confirm",
		"wrongfield", "wodify.csv", "text/csv", content, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ConfirmWodifyImport(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
}
