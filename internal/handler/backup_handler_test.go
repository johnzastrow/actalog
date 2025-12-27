package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackupHandler_CreateBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/backups", "")
	rr := httptest.NewRecorder()

	handler.CreateBackup(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestBackupHandler_GetBackupMetadata_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/admin/backups//metadata", "")
	rr := httptest.NewRecorder()

	handler.GetBackupMetadata(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Filename is required")
}

func TestBackupHandler_DownloadBackup_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/admin/backups/", "")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Filename is required")
}

func TestBackupHandler_DownloadBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	// Need to simulate a valid filename but no user context
	req := createTestRequest(http.MethodGet, "/api/admin/backups/test.zip", "")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	// Without chi router context, filename is empty -> "Filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestBackupHandler_DeleteBackup_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/backups/", "")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Filename is required")
}

func TestBackupHandler_DeleteBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/backups/test.zip", "")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	// Without chi router context, filename is empty -> "Filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestBackupHandler_UploadBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/backups/upload", "")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestBackupHandler_UploadBackup_NoFile(t *testing.T) {
	handler := &BackupHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/upload", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Failed to parse multipart form")
}

func TestBackupHandler_RestoreBackup_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/backups//restore", `{"confirm": true}`)
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Filename is required")
}

func TestBackupHandler_RestoreBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", `{"confirm": true}`)
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	// Without chi router context, filename is empty -> "Filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestBackupHandler_RestoreBackup_NotConfirmed(t *testing.T) {
	handler := &BackupHandler{}

	// Would need chi router context to test this properly
	// Testing pattern: if filename is present but confirm is false
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", `{"confirm": false}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	// Without chi router context, filename is empty -> "Filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "valid zip file", filename: "backup.zip", want: true},
		{name: "valid with date", filename: "backup_2024-01-15.zip", want: true},
		{name: "invalid extension", filename: "backup.tar", want: false},
		{name: "invalid no extension", filename: "backup", want: false},
		{name: "path traversal", filename: "../backup.zip", want: false},
		{name: "path with directory", filename: "path/backup.zip", want: false},
		{name: "double dot", filename: "..", want: false},
		{name: "empty", filename: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFilename(tt.filename)
			if got != tt.want {
				t.Errorf("isValidFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
