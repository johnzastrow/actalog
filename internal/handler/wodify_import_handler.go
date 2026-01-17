package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// WodifyImportHandler handles Wodify performance import endpoints
type WodifyImportHandler struct {
	wodifyImportService *service.WodifyImportService
}

// NewWodifyImportHandler creates a new Wodify import handler
func NewWodifyImportHandler(wodifyImportService *service.WodifyImportService) *WodifyImportHandler {
	return &WodifyImportHandler{
		wodifyImportService: wodifyImportService,
	}
}

// PreviewWodifyImport handles preview of Wodify performance import
// @Summary      Preview Wodify import
// @Description  Validate and preview Wodify performance data import before confirming
// @Tags         import-export
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Wodify export file"
// @Success      200 {object} map[string]interface{} "Import preview with validation results"
// @Failure      400 {object} ErrorResponse "Invalid file"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Preview failed"
// @Router       /import/wodify/preview [post]
func (h *WodifyImportHandler) PreviewWodifyImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Create reader for service
	reader := bytes.NewReader(fileBytes)

	// Preview import
	preview, err := h.wodifyImportService.PreviewImport(reader, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to preview import: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, preview)
}

// ConfirmWodifyImport handles confirmation of Wodify performance import
// @Summary      Confirm Wodify import
// @Description  Execute Wodify performance data import after preview
// @Tags         import-export
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Wodify export file"
// @Param        skip_duplicates formData bool false "Skip duplicate entries"
// @Param        update_duplicates formData bool false "Update existing duplicates"
// @Success      200 {object} map[string]interface{} "Import results with counts"
// @Failure      400 {object} ErrorResponse "Invalid file"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Import failed"
// @Router       /import/wodify/confirm [post]
func (h *WodifyImportHandler) ConfirmWodifyImport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Create reader for service
	reader := bytes.NewReader(fileBytes)

	// Parse options from form data
	skipDuplicates := r.FormValue("skip_duplicates") == "true"
	updateDuplicates := r.FormValue("update_duplicates") == "true"

	// Confirm import
	result, err := h.wodifyImportService.ConfirmImport(reader, userID, skipDuplicates, updateDuplicates)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}
