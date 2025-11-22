package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// ImportHandler handles data import endpoints
type ImportHandler struct {
	importService *service.ImportService
}

// NewImportHandler creates a new import handler
func NewImportHandler(importService *service.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// ImportWODsPreviewRequest represents the request body for confirming WOD import
type ImportWODsConfirmRequest struct {
	SkipDuplicates   bool   `json:"skip_duplicates"`
	UpdateDuplicates bool   `json:"update_duplicates"`
	CSVData          string `json:"csv_data"` // Base64 encoded or raw CSV string
}

// ImportMovementsConfirmRequest represents the request body for confirming movement import
type ImportMovementsConfirmRequest struct {
	SkipDuplicates   bool   `json:"skip_duplicates"`
	UpdateDuplicates bool   `json:"update_duplicates"`
	CSVData          string `json:"csv_data"` // Base64 encoded or raw CSV string
}

const maxUploadSize = 10 * 1024 * 1024 // 10MB

// PreviewWODImport validates and previews WOD CSV import
// POST /api/import/wods/preview (multipart/form-data with "file" field)
func (h *ImportHandler) PreviewWODImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	isAdmin := role == "admin"

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Preview import
	result, err := h.importService.PreviewWODImport(file, userID, isAdmin)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Import preview failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ConfirmWODImport confirms and executes WOD CSV import
// POST /api/import/wods/confirm (multipart/form-data with "file", "skip_duplicates", "update_duplicates" fields)
func (h *ImportHandler) ConfirmWODImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	isAdmin := role == "admin"

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Parse options from form data
	skipDuplicates := r.FormValue("skip_duplicates") == "true"
	updateDuplicates := r.FormValue("update_duplicates") == "true"

	// Confirm import
	result, err := h.importService.ConfirmWODImport(file, userID, isAdmin, skipDuplicates, updateDuplicates)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// PreviewMovementImport validates and previews movement CSV import
// POST /api/import/movements/preview (multipart/form-data with "file" field)
func (h *ImportHandler) PreviewMovementImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	isAdmin := role == "admin"

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Preview import
	result, err := h.importService.PreviewMovementImport(file, userID, isAdmin)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Import preview failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ConfirmMovementImport confirms and executes movement CSV import
// POST /api/import/movements/confirm (multipart/form-data with "file", "skip_duplicates", "update_duplicates" fields)
func (h *ImportHandler) ConfirmMovementImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is admin
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	isAdmin := role == "admin"

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Parse options from form data
	skipDuplicates := r.FormValue("skip_duplicates") == "true"
	updateDuplicates := r.FormValue("update_duplicates") == "true"

	// Confirm import
	result, err := h.importService.ConfirmMovementImport(file, userID, isAdmin, skipDuplicates, updateDuplicates)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// PreviewUserWorkoutImport validates and previews user workout JSON import
// POST /api/import/user-workouts/preview (multipart/form-data with "file" field)
func (h *ImportHandler) PreviewUserWorkoutImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Read file content
	jsonData, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Preview import
	result, err := h.importService.PreviewUserWorkoutImport(jsonData, userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Import preview failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ConfirmUserWorkoutImport confirms and executes user workout JSON import
// POST /api/import/user-workouts/confirm (multipart/form-data with "file", "skip_duplicates" fields)
func (h *ImportHandler) ConfirmUserWorkoutImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Read file content
	jsonData, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Parse options from form data
	skipDuplicates := r.FormValue("skip_duplicates") == "true"

	// Confirm import
	result, err := h.importService.ConfirmUserWorkoutImport(jsonData, userID, skipDuplicates)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}
