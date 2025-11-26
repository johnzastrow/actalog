package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// DataChangeLogHandler handles HTTP requests for data change logs
type DataChangeLogHandler struct {
	service *service.DataChangeLogService
	logger  *logger.Logger
}

// NewDataChangeLogHandler creates a new data change log handler
func NewDataChangeLogHandler(service *service.DataChangeLogService, logger *logger.Logger) *DataChangeLogHandler {
	return &DataChangeLogHandler{
		service: service,
		logger:  logger,
	}
}

// GetDataChangeLog handles GET /api/data-change-logs/:id
func (h *DataChangeLogHandler) GetDataChangeLog(w http.ResponseWriter, r *http.Request) {
	// Only admins can access individual data change logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		if h.logger != nil {
			h.logger.Warn("Unauthorized access attempt to data change log")
		}
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	log, err := h.service.GetByID(id)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to get data change log: %v", err)
		}
		respondError(w, http.StatusNotFound, "Data change log not found")
		return
	}

	respondJSON(w, http.StatusOK, log)
}

// ListDataChangeLogs handles GET /api/data-change-logs
// Query params: entity_type, entity_id, operation, user_id, start_date, end_date, limit, offset
func (h *DataChangeLogHandler) ListDataChangeLogs(w http.ResponseWriter, r *http.Request) {
	// Only admins can list all data change logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		if h.logger != nil {
			h.logger.Warn("Unauthorized access attempt to data change logs list")
		}
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	// Parse query parameters
	filters := domain.DataChangeLogFilters{}

	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filters.EntityType = &entityType
	}

	if entityIDStr := r.URL.Query().Get("entity_id"); entityIDStr != "" {
		entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid entity_id")
			return
		}
		filters.EntityID = &entityID
	}

	if operation := r.URL.Query().Get("operation"); operation != "" {
		filters.Operation = &operation
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid user_id")
			return
		}
		filters.UserID = &userID
	}

	if userEmail := r.URL.Query().Get("user_email"); userEmail != "" {
		filters.UserEmail = &userEmail
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			// Try simpler date format
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid start_date format (use RFC3339 or YYYY-MM-DD)")
				return
			}
		}
		filters.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			// Try simpler date format
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid end_date format (use RFC3339 or YYYY-MM-DD)")
				return
			}
			// Set to end of day
			endDate = endDate.Add(24*time.Hour - time.Second)
		}
		filters.EndDate = &endDate
	}

	// Parse pagination
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get logs and count
	logs, err := h.service.List(filters, limit, offset)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to list data change logs: %v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve data change logs")
		return
	}

	count, err := h.service.Count(filters)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to count data change logs: %v", err)
		}
		// Continue anyway, just set count to 0
		count = 0
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetEntityHistory handles GET /api/data-change-logs/entity/:entity_type/:entity_id
// Returns the change history for a specific entity
func (h *DataChangeLogHandler) GetEntityHistory(w http.ResponseWriter, r *http.Request) {
	// Only admins can view entity history
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	entityType := chi.URLParam(r, "entity_type")
	entityIDStr := chi.URLParam(r, "entity_id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity_id")
		return
	}

	// Parse pagination
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	logs, err := h.service.GetByEntityID(entityType, entityID, limit, offset)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to get entity history: %v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve entity history")
		return
	}

	response := map[string]interface{}{
		"entity_type": entityType,
		"entity_id":   entityID,
		"logs":        logs,
	}

	respondJSON(w, http.StatusOK, response)
}

// CleanupOldLogs handles POST /api/admin/data-change-logs/cleanup
// Admin endpoint to delete old data change logs
func (h *DataChangeLogHandler) CleanupOldLogs(w http.ResponseWriter, r *http.Request) {
	// Only admins can cleanup data change logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var request struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.RetentionDays <= 0 {
		respondError(w, http.StatusBadRequest, "retention_days must be positive")
		return
	}

	deletedCount, err := h.service.CleanupOldLogs(request.RetentionDays)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to cleanup old data change logs: %v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to cleanup data change logs")
		return
	}

	if h.logger != nil {
		adminUserID, _ := middleware.GetUserID(r.Context())
		h.logger.Info("Cleaned up old data change logs: retention_days=%d, deleted_count=%d, admin_user_id=%d",
			request.RetentionDays, deletedCount, adminUserID)
	}

	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"message":       "Old data change logs deleted successfully",
	}

	respondJSON(w, http.StatusOK, response)
}
