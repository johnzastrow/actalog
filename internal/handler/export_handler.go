package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// ExportHandler handles data export endpoints
type ExportHandler struct {
	exportService *service.ExportService
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// ExportWODs exports WODs to CSV or JSON format
// GET /api/export/wods?include_standard=true&include_custom=true&format=csv
func (h *ExportHandler) ExportWODs(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	includeStandard := parseBoolParam(r.URL.Query().Get("include_standard"), true)
	includeCustom := parseBoolParam(r.URL.Query().Get("include_custom"), true)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv" // default to CSV for backwards compatibility
	}

	var data []byte
	var err error
	var contentType, filename string

	// Generate export based on format
	if format == "json" {
		data, err = h.exportService.ExportWODsToJSON(userID, isAdmin, includeStandard, includeCustom)
		contentType = "application/json"
		filename = "wods_export.json"
	} else {
		data, err = h.exportService.ExportWODsToCSV(userID, isAdmin, includeStandard, includeCustom)
		contentType = "text/csv"
		filename = "wods_export.csv"
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export WODs: %v", err))
		return
	}

	// Set response headers for download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// Write data
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ExportMovements exports movements to CSV or JSON format
// GET /api/export/movements?include_standard=true&include_custom=true&format=csv
func (h *ExportHandler) ExportMovements(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	includeStandard := parseBoolParam(r.URL.Query().Get("include_standard"), true)
	includeCustom := parseBoolParam(r.URL.Query().Get("include_custom"), true)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv" // default to CSV for backwards compatibility
	}

	var data []byte
	var err error
	var contentType, filename string

	// Generate export based on format
	if format == "json" {
		data, err = h.exportService.ExportMovementsToJSON(userID, isAdmin, includeStandard, includeCustom)
		contentType = "application/json"
		filename = "movements_export.json"
	} else {
		data, err = h.exportService.ExportMovementsToCSV(userID, isAdmin, includeStandard, includeCustom)
		contentType = "text/csv"
		filename = "movements_export.csv"
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export movements: %v", err))
		return
	}

	// Set response headers for download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// Write data
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Helper function to parse boolean query parameters
func parseBoolParam(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

// ExportUserWorkouts exports user workouts to CSV or JSON format
// GET /api/export/user-workouts?start_date=2024-01-01&end_date=2024-12-31&format=json
func (h *ExportHandler) ExportUserWorkouts(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // default to JSON for backwards compatibility
	}

	var startDate, endDate *time.Time

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
		startDate = &parsed
	}

	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
		endDate = &parsed
	}

	// Both start and end date must be provided together, or neither
	if (startDate != nil && endDate == nil) || (startDate == nil && endDate != nil) {
		respondError(w, http.StatusBadRequest, "Both start_date and end_date must be provided together")
		return
	}

	var data []byte
	var err error
	var contentType, filename, fileExtension string

	// Generate export based on format
	if format == "csv" {
		data, err = h.exportService.ExportUserWorkoutsToCSV(userID, startDate, endDate)
		contentType = "text/csv"
		fileExtension = "csv"
	} else {
		data, err = h.exportService.ExportUserWorkoutsToJSON(userID, startDate, endDate)
		contentType = "application/json"
		fileExtension = "json"
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export user workouts: %v", err))
		return
	}

	// Determine filename based on date range
	if startDate != nil && endDate != nil {
		filename = fmt.Sprintf("user_workouts_%s_to_%s.%s",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			fileExtension)
	} else {
		filename = fmt.Sprintf("user_workouts_export.%s", fileExtension)
	}

	// Set response headers for download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// Write data
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
