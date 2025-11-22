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

// ExportWODs exports WODs to CSV format
// GET /api/export/wods?include_standard=true&include_custom=true
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

	// Generate CSV
	csvData, err := h.exportService.ExportWODsToCSV(userID, isAdmin, includeStandard, includeCustom)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export WODs: %v", err))
		return
	}

	// Set response headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=wods_export.csv")
	w.Header().Set("Content-Length", strconv.Itoa(len(csvData)))

	// Write CSV data
	w.WriteHeader(http.StatusOK)
	w.Write(csvData)
}

// ExportMovements exports movements to CSV format
// GET /api/export/movements?include_standard=true&include_custom=true
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

	// Generate CSV
	csvData, err := h.exportService.ExportMovementsToCSV(userID, isAdmin, includeStandard, includeCustom)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export movements: %v", err))
		return
	}

	// Set response headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=movements_export.csv")
	w.Header().Set("Content-Length", strconv.Itoa(len(csvData)))

	// Write CSV data
	w.WriteHeader(http.StatusOK)
	w.Write(csvData)
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

// ExportUserWorkouts exports user workouts with full nested data to JSON format
// GET /api/export/user-workouts?start_date=2024-01-01&end_date=2024-12-31
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

	// Generate JSON
	jsonData, err := h.exportService.ExportUserWorkoutsToJSON(userID, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export user workouts: %v", err))
		return
	}

	// Determine filename based on date range
	filename := "user_workouts_export.json"
	if startDate != nil && endDate != nil {
		filename = fmt.Sprintf("user_workouts_%s_to_%s.json", 
			startDate.Format("2006-01-02"), 
			endDate.Format("2006-01-02"))
	}

	// Set response headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Write JSON data
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
