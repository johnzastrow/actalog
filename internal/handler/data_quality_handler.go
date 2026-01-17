package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// DataQualityHandler handles data quality scanning and duplicate cleanup
type DataQualityHandler struct {
	dataQualityService *service.DataQualityService
	logger             *logger.Logger
}

// NewDataQualityHandler creates a new data quality handler
func NewDataQualityHandler(
	dataQualityService *service.DataQualityService,
	logger *logger.Logger,
) *DataQualityHandler {
	return &DataQualityHandler{
		dataQualityService: dataQualityService,
		logger:             logger,
	}
}

// ScanAllDuplicates scans all entity types for duplicates
// @Summary      Scan all entity types for duplicates (Admin)
// @Description  Scans movements, WODs, user workouts, users, and workout templates for duplicates
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Duplicate scan results for all entity types"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/duplicates [get]
func (h *DataQualityHandler) ScanAllDuplicates(w http.ResponseWriter, r *http.Request) {
	results, err := h.dataQualityService.ScanAllDuplicates()
	if err != nil {
		h.logger.Error("Failed to scan for duplicates: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan for duplicates"})
		return
	}

	// Convert map[EntityType] to map[string] for JSON
	response := make(map[string]interface{})
	for et, result := range results {
		response[string(et)] = result
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ScanDuplicatesByEntity scans a specific entity type for duplicates
// @Summary      Scan specific entity type for duplicates (Admin)
// @Description  Scans a single entity type (movements, wods, user_workouts, users, workouts) for duplicates
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity path string true "Entity type" Enums(movements, wods, user_workouts, users, workouts)
// @Success      200 {object} service.DuplicateScanResult "Duplicate scan results"
// @Failure      400 {object} ErrorResponse "Invalid entity type"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/duplicates/{entity} [get]
func (h *DataQualityHandler) ScanDuplicatesByEntity(w http.ResponseWriter, r *http.Request) {
	entityStr := chi.URLParam(r, "entity")
	entityType := service.EntityType(entityStr)

	// Validate entity type
	validTypes := map[service.EntityType]bool{
		service.EntityTypeMovement:    true,
		service.EntityTypeWOD:         true,
		service.EntityTypeUserWorkout: true,
		service.EntityTypeUser:        true,
		service.EntityTypeWorkout:     true,
	}

	if !validTypes[entityType] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid entity type. Must be one of: movements, wods, user_workouts, users, workouts"})
		return
	}

	result, err := h.dataQualityService.ScanForDuplicates(entityType)
	if err != nil {
		h.logger.Error("Failed to scan %s for duplicates: %v", entityType, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan for duplicates"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MergePreviewRequest is the request body for merge preview
type MergePreviewRequest struct {
	EntityType string  `json:"entity_type"`
	KeepID     int64   `json:"keep_id"`
	DeleteIDs  []int64 `json:"delete_ids"`
}

// PreviewMerge shows what will happen when merging duplicates
// @Summary      Preview merge operation (Admin)
// @Description  Shows FK references that will be updated and any warnings before executing a merge
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body MergePreviewRequest true "Merge preview request"
// @Success      200 {object} service.MergePreview "Merge preview"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/duplicates/merge/preview [post]
func (h *DataQualityHandler) PreviewMerge(w http.ResponseWriter, r *http.Request) {
	var req MergePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.EntityType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "entity_type is required"})
		return
	}
	if req.KeepID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "keep_id is required"})
		return
	}
	if len(req.DeleteIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "delete_ids is required"})
		return
	}

	entityType := service.EntityType(req.EntityType)
	preview, err := h.dataQualityService.PreviewMerge(entityType, req.KeepID, req.DeleteIDs)
	if err != nil {
		h.logger.Error("Failed to preview merge: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// MergeConfirmRequest is the request body for confirming a merge
type MergeConfirmRequest struct {
	EntityType string  `json:"entity_type"`
	KeepID     int64   `json:"keep_id"`
	DeleteIDs  []int64 `json:"delete_ids"`
}

// ConfirmMerge executes the merge operation
// @Summary      Execute merge operation (Admin)
// @Description  Merges duplicate records by updating FK references and deleting duplicates
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body MergeConfirmRequest true "Merge confirm request"
// @Success      200 {object} service.MergeResult "Merge result"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/duplicates/merge/confirm [post]
func (h *DataQualityHandler) ConfirmMerge(w http.ResponseWriter, r *http.Request) {
	var req MergeConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.EntityType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "entity_type is required"})
		return
	}
	if req.KeepID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "keep_id is required"})
		return
	}
	if len(req.DeleteIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "delete_ids is required"})
		return
	}

	// Get admin user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	entityType := service.EntityType(req.EntityType)
	result, err := h.dataQualityService.ExecuteMerge(entityType, req.KeepID, req.DeleteIDs, userID)
	if err != nil {
		h.logger.Error("Failed to execute merge: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ScanDataQuality performs a comprehensive data quality scan
// @Summary      Scan for data quality issues (Admin)
// @Description  Checks for orphaned FK references, empty required fields, future dates, invalid emails, etc.
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} service.DataQualityScanResult "Data quality scan results"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/issues [get]
func (h *DataQualityHandler) ScanDataQuality(w http.ResponseWriter, r *http.Request) {
	result, err := h.dataQualityService.ScanDataQuality()
	if err != nil {
		h.logger.Error("Failed to scan data quality: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan data quality"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// FullDataQualityScan runs both duplicate detection and data quality checks
// @Summary      Run full data quality scan (Admin)
// @Description  Runs comprehensive scan including duplicate detection and data quality issues
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Full scan results"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/full-scan [get]
func (h *DataQualityHandler) FullDataQualityScan(w http.ResponseWriter, r *http.Request) {
	// Run duplicate scan
	duplicates, err := h.dataQualityService.ScanAllDuplicates()
	if err != nil {
		h.logger.Error("Failed to scan for duplicates: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan for duplicates"})
		return
	}

	// Run data quality scan
	issues, err := h.dataQualityService.ScanDataQuality()
	if err != nil {
		h.logger.Error("Failed to scan data quality: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan data quality"})
		return
	}

	// Convert duplicates map for JSON
	duplicatesMap := make(map[string]interface{})
	totalDuplicates := 0
	for et, result := range duplicates {
		duplicatesMap[string(et)] = result
		totalDuplicates += result.TotalDups
	}

	response := map[string]interface{}{
		"duplicates": duplicatesMap,
		"issues":     issues,
		"summary": map[string]int{
			"total_duplicate_records":   totalDuplicates,
			"total_data_quality_issues": issues.TotalCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDuplicateSummary returns a summary of duplicates across all entity types
// @Summary      Get duplicate summary (Admin)
// @Description  Returns counts of duplicates for each entity type
// @Tags         admin,data-quality
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Duplicate summary"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-quality/duplicates/summary [get]
func (h *DataQualityHandler) GetDuplicateSummary(w http.ResponseWriter, r *http.Request) {
	results, err := h.dataQualityService.ScanAllDuplicates()
	if err != nil {
		h.logger.Error("Failed to scan for duplicates: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to scan for duplicates"})
		return
	}

	summary := make(map[string]interface{})
	totalDuplicates := 0
	totalGroups := 0

	for et, result := range results {
		summary[string(et)] = map[string]interface{}{
			"duplicate_count": result.TotalDups,
			"group_count":     len(result.Groups),
			"scan_time":       result.ScanTime,
		}
		totalDuplicates += result.TotalDups
		totalGroups += len(result.Groups)
	}

	response := map[string]interface{}{
		"by_entity":        summary,
		"total_duplicates": totalDuplicates,
		"total_groups":     totalGroups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to parse int64 slice from query parameter
func parseInt64Slice(param string) ([]int64, error) {
	if param == "" {
		return nil, nil
	}
	parts := strings.Split(param, ",")
	result := make([]int64, len(parts))
	for i, p := range parts {
		val, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}
