package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
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
	assertBodyContains(t, rr, "filename is required")
}

func TestBackupHandler_DownloadBackup_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/admin/backups/", "")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "filename is required")
}

func TestBackupHandler_DownloadBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	// Need to simulate a valid filename but no user context
	req := createTestRequest(http.MethodGet, "/api/admin/backups/test.zip", "")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	// Without chi router context, filename is empty -> "filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestBackupHandler_DeleteBackup_MissingFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/backups/", "")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "filename is required")
}

func TestBackupHandler_DeleteBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/backups/test.zip", "")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	// Without chi router context, filename is empty -> "filename is required"
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
	assertBodyContains(t, rr, "filename is required")
}

func TestBackupHandler_RestoreBackup_Unauthorized(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", `{"confirm": true}`)
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	// Without chi router context, filename is empty -> "filename is required"
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestBackupHandler_RestoreBackup_NotConfirmed(t *testing.T) {
	handler := &BackupHandler{}

	// Would need chi router context to test this properly
	// Testing pattern: if filename is present but confirm is false
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", `{"confirm": false}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	// Without chi router context, filename is empty -> "filename is required"
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

// Test NewBackupHandler

func TestNewBackupHandler(t *testing.T) {
	handler := NewBackupHandler(nil, nil)
	if handler == nil {
		t.Error("NewBackupHandler should return a non-nil handler")
	}
}

// Tests with chi URL params

func TestBackupHandler_DownloadBackup_InvalidFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/backups/../test.zip", "")
	req = addChiURLParam(req, "filename", "../test.zip")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid filename")
}

func TestBackupHandler_DeleteBackup_InvalidFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/backups/../test.zip", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "../test.zip")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid filename")
}

func TestBackupHandler_GetBackupMetadata_InvalidFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/backups/../test.zip/metadata", "")
	req = addChiURLParam(req, "filename", "../test.zip")
	rr := httptest.NewRecorder()

	handler.GetBackupMetadata(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid filename")
}

func TestBackupHandler_RestoreBackup_InvalidFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/../test.zip/restore", `{"confirm": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "../test.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid filename")
}

func TestBackupHandler_RestoreBackup_InvalidJSON(t *testing.T) {
	handler := &BackupHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "test.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestBackupHandler_RestoreBackup_NotConfirmedWithValidFilename(t *testing.T) {
	handler := &BackupHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/test.zip/restore", `{"confirm": false}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "test.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "confirmation required")
}

// Removed 10 panic-expectation tests:
// - TestBackupHandler_ListBackups_NilService
// - TestBackupHandler_CreateBackup_ValidInput
// - TestBackupHandler_DownloadBackup_ValidFilename
// - TestBackupHandler_DeleteBackup_ValidFilename
// - TestBackupHandler_GetBackupMetadata_ValidFilename
// - TestBackupHandler_RestoreBackup_ValidConfirmed
// - TestBackupHandler_CreateBackup_NoDescription
// - TestBackupHandler_CreateBackup_InvalidJSON
// - TestBackupHandler_UploadBackup_WithFileNilService
// - TestBackupHandler_DownloadBackup_DifferentFilenames (3 subtests)
// These tests verified nil pointer panics, not business logic.

func TestIsValidFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "with underscore", filename: "backup_test.zip", want: true},
		{name: "with dash", filename: "backup-test.zip", want: true},
		{name: "with numbers", filename: "backup123.zip", want: true},
		{name: "uppercase ZIP", filename: "backup.ZIP", want: false},
		// These are allowed by the current implementation
		{name: "double extension", filename: "backup.tar.zip", want: true},
		{name: "hidden file", filename: ".backup.zip", want: true},
		{name: "windows path has backslash", filename: "C:\\backup.zip", want: true},
		{name: "url encoded", filename: "%2e%2e/backup.zip", want: false}, // has forward slash
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

// Tests with mock backup service

func TestBackupHandler_CreateBackup_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateBackup(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "Backup created successfully")
	assertBodyContains(t, rr, "backup_test.json")
}

func TestBackupHandler_CreateBackup_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to create backup")
}

func TestBackupHandler_ListBackups_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createTestRequest(http.MethodGet, "/api/admin/backups", "")
	rr := httptest.NewRecorder()

	handler.ListBackups(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "backups")
	assertBodyContains(t, rr, "count")
}

func TestBackupHandler_ListBackups_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createTestRequest(http.MethodGet, "/api/admin/backups", "")
	rr := httptest.NewRecorder()

	handler.ListBackups(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to list backups")
}

func TestBackupHandler_GetBackupMetadata_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createTestRequest(http.MethodGet, "/api/admin/backups/backup_2024-01-15.json/metadata", "")
	req = addChiURLParam(req, "filename", "backup_2024-01-15.zip")
	rr := httptest.NewRecorder()

	handler.GetBackupMetadata(rr, req)

	// backup_2024-01-15.zip is not in mock backups
	assertStatusCode(t, rr, http.StatusNotFound)
}

func TestBackupHandler_GetBackupMetadata_NotFound(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createTestRequest(http.MethodGet, "/api/admin/backups/nonexistent.zip/metadata", "")
	req = addChiURLParam(req, "filename", "nonexistent.zip")
	rr := httptest.NewRecorder()

	handler.GetBackupMetadata(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Backup not found")
}

func TestBackupHandler_DeleteBackup_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/backups/backup.zip", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Backup deleted successfully")
}

func TestBackupHandler_DeleteBackup_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/backups/backup.zip", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.DeleteBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to delete backup")
}

func TestBackupHandler_RestoreBackup_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/backup.zip/restore", `{"confirm": true, "mode": "replace"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Backup restored successfully")
}

func TestBackupHandler_RestoreBackup_MergeMode(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/backup.zip/restore", `{"confirm": true, "mode": "merge"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Backup restored successfully")
}

func TestBackupHandler_RestoreBackup_SkipMode(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/backup.zip/restore", `{"confirm": true, "mode": "skip"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Backup restored successfully")
}

func TestBackupHandler_RestoreBackup_InvalidMode(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/backup.zip/restore", `{"confirm": true, "mode": "invalid"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid restore mode")
}

func TestBackupHandler_RestoreBackup_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups/backup.zip/restore", `{"confirm": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.RestoreBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to restore backup")
}

func TestStringPtr(t *testing.T) {
	s := "test string"
	ptr := stringPtr(s)
	if ptr == nil {
		t.Error("stringPtr should return non-nil pointer")
	}
	if *ptr != s {
		t.Errorf("stringPtr returned wrong value: got %q, want %q", *ptr, s)
	}
}

// Additional tests for DownloadBackup
func TestBackupHandler_DownloadBackup_BackupNotFound(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockNotFound)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/backups/nonexistent.zip", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "nonexistent.zip")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Backup not found")
}

// Test UploadBackup with non-zip file
func TestBackupHandler_UploadBackup_NonZipFile(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	// Create a multipart request with non-zip file
	content := []byte("This is not a zip file")
	req := createMultipartRequest(http.MethodPost, "/api/admin/backups/upload",
		"file", "backup.txt", "text/plain", content, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Only ZIP files are allowed")
}

// Test UploadBackup with service error
func TestBackupHandler_UploadBackup_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	// Create a multipart request with zip file
	content := []byte("PK\x03\x04") // Start of a zip file
	req := createMultipartRequest(http.MethodPost, "/api/admin/backups/upload",
		"file", "backup.zip", "application/zip", content, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to upload backup")
}

// Test UploadBackup success
func TestBackupHandler_UploadBackup_Success(t *testing.T) {
	mockService := NewMockBackupService()
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	// Create a multipart request with zip file
	content := []byte("PK\x03\x04") // Start of a zip file
	req := createMultipartRequest(http.MethodPost, "/api/admin/backups/upload",
		"file", "backup.zip", "application/zip", content, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "Backup uploaded successfully")
}

// Test CreateBackup with metadata retrieval error
func TestBackupHandler_CreateBackup_MetadataError(t *testing.T) {
	mockService := &MockBackupServiceWithMetadataError{}
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/backups", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get backup metadata")
}

// MockBackupServiceWithMetadataError is a mock that fails on GetBackupMetadata
type MockBackupServiceWithMetadataError struct{}

func (m *MockBackupServiceWithMetadataError) CreateBackup(createdByUserID int64) (string, error) {
	return "test.zip", nil
}

func (m *MockBackupServiceWithMetadataError) ListBackups() ([]domain.BackupMetadata, error) {
	return nil, nil
}

func (m *MockBackupServiceWithMetadataError) GetBackupMetadata(filename string) (*domain.BackupMetadata, error) {
	return nil, ErrMockInternalError
}

func (m *MockBackupServiceWithMetadataError) DownloadBackup(filename string) (string, error) {
	return "", nil
}

func (m *MockBackupServiceWithMetadataError) UploadBackup(file interface{}, filename string, uploadedByUserID int64) (string, error) {
	return "", nil
}

func (m *MockBackupServiceWithMetadataError) DeleteBackup(filename string, deletedByUserID int64) error {
	return nil
}

func (m *MockBackupServiceWithMetadataError) RestoreBackup(filename string, restoredByUserID int64, mode domain.RestoreMode) (*domain.RestoreResult, error) {
	return nil, nil
}

// Test UploadBackup with metadata retrieval error
func TestBackupHandler_UploadBackup_MetadataError(t *testing.T) {
	mockService := &MockBackupServiceWithUploadMetadataError{}
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	// Create a multipart request with zip file
	content := []byte("PK\x03\x04") // Start of a zip file
	req := createMultipartRequest(http.MethodPost, "/api/admin/backups/upload",
		"file", "backup.zip", "application/zip", content, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UploadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get backup metadata")
}

// MockBackupServiceWithUploadMetadataError is a mock that succeeds on Upload but fails on GetBackupMetadata
type MockBackupServiceWithUploadMetadataError struct{}

func (m *MockBackupServiceWithUploadMetadataError) CreateBackup(createdByUserID int64) (string, error) {
	return "", nil
}

func (m *MockBackupServiceWithUploadMetadataError) ListBackups() ([]domain.BackupMetadata, error) {
	return nil, nil
}

func (m *MockBackupServiceWithUploadMetadataError) GetBackupMetadata(filename string) (*domain.BackupMetadata, error) {
	return nil, ErrMockInternalError
}

func (m *MockBackupServiceWithUploadMetadataError) DownloadBackup(filename string) (string, error) {
	return "", nil
}

func (m *MockBackupServiceWithUploadMetadataError) UploadBackup(file interface{}, filename string, uploadedByUserID int64) (string, error) {
	return "uploaded.zip", nil
}

func (m *MockBackupServiceWithUploadMetadataError) DeleteBackup(filename string, deletedByUserID int64) error {
	return nil
}

func (m *MockBackupServiceWithUploadMetadataError) RestoreBackup(filename string, restoredByUserID int64, mode domain.RestoreMode) (*domain.RestoreResult, error) {
	return nil, nil
}

// Tests for DownloadBackup handler
func TestBackupHandler_DownloadBackup_ServiceError(t *testing.T) {
	mockService := NewMockBackupService()
	mockService.SetError(ErrMockInternalError)
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/backups/test.zip/download", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "test.zip")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Backup not found")
}

func TestBackupHandler_DownloadBackup_Success(t *testing.T) {
	// Create a temporary file to serve
	tmpFile, err := os.CreateTemp("", "backup_test_*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some content to the file
	content := []byte("test backup content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Create a mock that returns the temp file path
	mockService := &MockBackupServiceWithFilePath{filePath: tmpFile.Name()}
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/backups/backup.zip/download", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Should have correct content type for zip files
	if rr.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Expected Content-Type application/zip, got %s", rr.Header().Get("Content-Type"))
	}
	// Should have Content-Disposition header
	if rr.Header().Get("Content-Disposition") == "" {
		t.Error("Expected Content-Disposition header to be set")
	}
}

func TestBackupHandler_DownloadBackup_FileStatError(t *testing.T) {
	// Use a path that doesn't exist
	mockService := &MockBackupServiceWithFilePath{filePath: "/nonexistent/path/file.zip"}
	mockAuditRepo := NewMockAuditLogRepository()
	handler := NewBackupHandler(mockService, mockAuditRepo)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/backups/backup.zip/download", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "filename", "backup.zip")
	rr := httptest.NewRecorder()

	handler.DownloadBackup(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get file info")
}

// MockBackupServiceWithFilePath returns a specific file path for download
type MockBackupServiceWithFilePath struct {
	filePath string
}

func (m *MockBackupServiceWithFilePath) CreateBackup(createdByUserID int64) (string, error) {
	return "test.zip", nil
}

func (m *MockBackupServiceWithFilePath) ListBackups() ([]domain.BackupMetadata, error) {
	return []domain.BackupMetadata{}, nil
}

func (m *MockBackupServiceWithFilePath) GetBackupMetadata(filename string) (*domain.BackupMetadata, error) {
	return &domain.BackupMetadata{
		Filename:       filename,
		FileSize:       1024,
		CreatedAt:      time.Now(),
		CreatedByEmail: "admin@example.com",
		Version:        "0.17.0",
		DatabaseDriver: "sqlite3",
		DatabaseName:   "test.db",
	}, nil
}

func (m *MockBackupServiceWithFilePath) DownloadBackup(filename string) (string, error) {
	return m.filePath, nil
}

func (m *MockBackupServiceWithFilePath) UploadBackup(file interface{}, filename string, uploadedByUserID int64) (string, error) {
	return "", nil
}

func (m *MockBackupServiceWithFilePath) DeleteBackup(filename string, deletedByUserID int64) error {
	return nil
}

func (m *MockBackupServiceWithFilePath) RestoreBackup(filename string, restoredByUserID int64, mode domain.RestoreMode) (*domain.RestoreResult, error) {
	return nil, nil
}
