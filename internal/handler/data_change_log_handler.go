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
// @Summary      Get data change log (Admin)
// @Description  Retrieve a specific data change log entry by ID
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Data Change Log ID"
// @Success      200 {object} domain.DataChangeLog "Data change log entry"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      404 {object} ErrorResponse "Log not found"
// @Router       /data-change-logs/{id} [get]
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
// @Summary      List data change logs (Admin)
// @Description  Get a paginated list of data change logs with optional filters
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_type query string false "Filter by entity type"
// @Param        entity_id query int false "Filter by entity ID"
// @Param        operation query string false "Filter by operation (create, update, delete)"
// @Param        user_id query int false "Filter by user ID"
// @Param        user_email query string false "Filter by user email"
// @Param        start_date query string false "Filter by start date (RFC3339 or YYYY-MM-DD)"
// @Param        end_date query string false "Filter by end date (RFC3339 or YYYY-MM-DD)"
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "Data change logs with total count"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /data-change-logs [get]
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
// @Summary      Get entity history (Admin)
// @Description  Get the change history for a specific entity
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_type path string true "Entity type (e.g., user, workout, wod)"
// @Param        entity_id path int true "Entity ID"
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "Entity change history"
// @Failure      400 {object} ErrorResponse "Invalid entity_id"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /data-change-logs/entity/{entity_type}/{entity_id} [get]
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
// @Summary      Cleanup old data change logs (Admin)
// @Description  Delete data change logs older than specified retention period
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Retention days (must be positive)"
// @Success      200 {object} map[string]interface{} "Deleted count and success message"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/data-change-logs/cleanup [post]
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
