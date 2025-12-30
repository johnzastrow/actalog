package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// BackupHandler handles backup/restore endpoints
type BackupHandler struct {
	backupService domain.BackupService
	auditLogRepo  domain.AuditLogRepository
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService domain.BackupService, auditLogRepo domain.AuditLogRepository) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		auditLogRepo:  auditLogRepo,
	}
}

// CreateBackup creates a new database backup
// @Summary      Create backup (Admin)
// @Description  Create a new database backup
// @Tags         backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} map[string]interface{} "Backup created with filename and metadata"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups [post]
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Create backup
	filename, err := h.backupService.CreateBackup(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create backup: %v", err))
		return
	}

	// Get metadata for the created backup
	metadata, err := h.backupService.GetBackupMetadata(filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get backup metadata: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "Backup created successfully",
		"filename": filename,
		"metadata": metadata,
	})
}

// ListBackups returns a list of all available backups
// @Summary      List backups (Admin)
// @Description  Get a list of all available backup files
// @Tags         backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "List of backups with count"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups [get]
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	backups, err := h.backupService.ListBackups()
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list backups: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"backups": backups,
		"count":   len(backups),
	})
}

// GetBackupMetadata returns metadata for a specific backup
// @Summary      Get backup metadata (Admin)
// @Description  Get metadata information for a specific backup file
// @Tags         backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        filename path string true "Backup filename"
// @Success      200 {object} domain.BackupMetadata "Backup metadata"
// @Failure      400 {object} ErrorResponse "Invalid filename"
// @Failure      404 {object} ErrorResponse "Backup not found"
// @Router       /admin/backups/{filename}/metadata [get]
func (h *BackupHandler) GetBackupMetadata(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "Filename is required")
		return
	}

	// Validate filename (prevent directory traversal)
	if !isValidFilename(filename) {
		respondError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	metadata, err := h.backupService.GetBackupMetadata(filename)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Backup not found: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, metadata)
}

// DownloadBackup downloads a backup file
// @Summary      Download backup (Admin)
// @Description  Download a backup file
// @Tags         backups
// @Accept       json
// @Produce      application/zip
// @Security     BearerAuth
// @Param        filename path string true "Backup filename"
// @Success      200 {file} file "Backup ZIP file"
// @Failure      400 {object} ErrorResponse "Invalid filename"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Backup not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups/{filename} [get]
func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "Filename is required")
		return
	}

	// Validate filename (prevent directory traversal)
	if !isValidFilename(filename) {
		respondError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	filePath, err := h.backupService.DownloadBackup(filename)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Backup not found: %v", err))
		return
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get file info: %v", err))
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Open and serve the file
	file, err := os.Open(filePath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to open file: %v", err))
		return
	}
	defer file.Close()

	// Copy file contents to response
	http.ServeContent(w, r, filename, fileInfo.ModTime(), file)

	// Create audit log for download
	go func() {
		if err := h.auditLogRepo.Create(&domain.AuditLog{
			UserID:    &userID,
			EventType: "backup_downloaded",
			Details:   stringPtr(fmt.Sprintf("Downloaded backup: %s (size: %d bytes)", filename, fileInfo.Size())),
			CreatedAt: time.Now(),
		}); err != nil {
			fmt.Printf("Warning: failed to create audit log for backup download: %v\n", err)
		}
	}()
}

// DeleteBackup deletes a backup file
// @Summary      Delete backup (Admin)
// @Description  Delete a backup file
// @Tags         backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        filename path string true "Backup filename"
// @Success      200 {object} map[string]interface{} "Success message with filename"
// @Failure      400 {object} ErrorResponse "Invalid filename"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups/{filename} [delete]
func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "Filename is required")
		return
	}

	// Validate filename (prevent directory traversal)
	if !isValidFilename(filename) {
		respondError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete backup
	if err := h.backupService.DeleteBackup(filename, userID); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete backup: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Backup deleted successfully",
		"filename": filename,
	})
}

// UploadBackup uploads a backup ZIP file from another system
// @Summary      Upload backup (Admin)
// @Description  Upload a backup ZIP file from another system
// @Tags         backups
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Backup ZIP file"
// @Success      201 {object} map[string]interface{} "Upload success with filename and metadata"
// @Failure      400 {object} ErrorResponse "Invalid file or format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups/upload [post]
func (h *BackupHandler) UploadBackup(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to get uploaded file: %v", err))
		return
	}
	defer file.Close()

	// Validate file extension
	if filepath.Ext(header.Filename) != ".zip" {
		respondError(w, http.StatusBadRequest, "Only ZIP files are allowed")
		return
	}

	// Upload backup (this will save it to backups/ directory)
	filename, err := h.backupService.UploadBackup(file, header.Filename, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload backup: %v", err))
		return
	}

	// Get metadata for the uploaded backup
	metadata, err := h.backupService.GetBackupMetadata(filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get backup metadata: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "Backup uploaded successfully",
		"filename": filename,
		"metadata": metadata,
	})
}

// RestoreBackup restores database from a backup file
// @Summary      Restore backup (Admin)
// @Description  Restore database from a backup file (requires confirmation)
// @Tags         backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        filename path string true "Backup filename"
// @Param        request body object true "Restore options (confirm, mode)"
// @Success      200 {object} map[string]interface{} "Restore result"
// @Failure      400 {object} ErrorResponse "Invalid request or filename"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/backups/{filename}/restore [post]
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "Filename is required")
		return
	}

	// Validate filename (prevent directory traversal)
	if !isValidFilename(filename) {
		respondError(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body for confirmation and mode
	var req struct {
		Confirm bool   `json:"confirm"`
		Mode    string `json:"mode"` // "replace" (default), "merge", or "skip"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !req.Confirm {
		respondError(w, http.StatusBadRequest, "Restore confirmation required")
		return
	}

	// Validate and set restore mode
	mode := domain.RestoreModeReplace // Default
	switch req.Mode {
	case "", "replace":
		mode = domain.RestoreModeReplace
	case "merge":
		mode = domain.RestoreModeMerge
	case "skip":
		mode = domain.RestoreModeSkip
	default:
		respondError(w, http.StatusBadRequest, "Invalid restore mode. Use 'replace', 'merge', or 'skip'")
		return
	}

	// Restore backup
	result, err := h.backupService.RestoreBackup(filename, userID, mode)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restore backup: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Backup restored successfully",
		"filename": filename,
		"result":   result,
	})
}

// Helper function to validate filename (prevent directory traversal attacks)
func isValidFilename(filename string) bool {
	// Must end with .zip
	if filepath.Ext(filename) != ".zip" {
		return false
	}

	// Must not contain path separators
	if filepath.Base(filename) != filename {
		return false
	}

	// Must not contain ".."
	if filename == ".." || len(filename) == 0 {
		return false
	}

	return true
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
