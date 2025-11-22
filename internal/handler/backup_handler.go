package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// BackupHandler handles backup/restore endpoints
type BackupHandler struct {
	backupService domain.BackupService
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService domain.BackupService) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

// CreateBackup creates a new database backup
// POST /api/admin/backups
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
// GET /api/admin/backups
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
// GET /api/admin/backups/{filename}/metadata
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
// GET /api/admin/backups/{filename}
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
}

// DeleteBackup deletes a backup file
// DELETE /api/admin/backups/{filename}
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

// RestoreBackup restores database from a backup file
// POST /api/admin/backups/{filename}/restore
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

	// Parse request body for confirmation
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !req.Confirm {
		respondError(w, http.StatusBadRequest, "Restore confirmation required")
		return
	}

	// Restore backup
	if err := h.backupService.RestoreBackup(filename, userID); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restore backup: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Backup restored successfully",
		"filename": filename,
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
