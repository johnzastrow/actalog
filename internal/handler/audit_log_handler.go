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

// AuditLogHandler handles HTTP requests for audit logs
type AuditLogHandler struct {
	service *service.AuditLogService
	logger  *logger.Logger
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(service *service.AuditLogService, logger *logger.Logger) *AuditLogHandler {
	return &AuditLogHandler{
		service: service,
		logger:  logger,
	}
}

// GetAuditLog handles GET /api/audit-logs/:id
// @Summary      Get audit log entry (Admin)
// @Description  Retrieve a specific audit log entry by ID
// @Tags         audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Audit Log ID"
// @Success      200 {object} domain.AuditLog "Audit log entry"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      404 {object} ErrorResponse "Audit log not found"
// @Router       /audit-logs/{id} [get]
func (h *AuditLogHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	// Only admins can access individual audit logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		h.logger.Warn("Unauthorized access attempt to audit log: user_role=%s path=%s", userRole, r.URL.Path)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	log, err := h.service.GetByID(id)
	if err != nil {
		h.logger.Error("Failed to get audit log: id=%d error=%v", id, err)
		http.Error(w, "Audit log not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

// ListAuditLogs handles GET /api/audit-logs
// @Summary      List audit logs (Admin)
// @Description  Get a paginated list of audit logs with optional filters
// @Tags         audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id query int false "Filter by user ID"
// @Param        target_user_id query int false "Filter by target user ID"
// @Param        event_type query string false "Filter by event type"
// @Param        ip_address query string false "Filter by IP address"
// @Param        start_date query string false "Filter by start date (RFC3339)"
// @Param        end_date query string false "Filter by end date (RFC3339)"
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "Audit logs list with total count"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /audit-logs [get]
func (h *AuditLogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Only admins can list all audit logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		h.logger.Warn("Unauthorized access attempt to audit logs list: user_role=%s", userRole)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse query parameters
	filters := domain.AuditLogFilters{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		filters.UserID = &userID
	}

	if targetUserIDStr := r.URL.Query().Get("target_user_id"); targetUserIDStr != "" {
		targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid target_user_id", http.StatusBadRequest)
			return
		}
		filters.TargetUserID = &targetUserID
	}

	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		filters.EventType = &eventType
	}

	if ipAddress := r.URL.Query().Get("ip_address"); ipAddress != "" {
		filters.IPAddress = &ipAddress
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use RFC3339)", http.StatusBadRequest)
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use RFC3339)", http.StatusBadRequest)
			return
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
		h.logger.Error("Failed to list audit logs: %v", err)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	count, err := h.service.Count(filters)
	if err != nil {
		h.logger.Error("Failed to count audit logs: %v", err)
		// Continue anyway, just set count to 0
		count = 0
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMyAuditLogs handles GET /api/users/me/audit-logs
// @Summary      Get my audit logs
// @Description  Get audit logs for the current user only
// @Tags         audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "User's audit logs with total count"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /users/me/audit-logs [get]
func (h *AuditLogHandler) GetMyAuditLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	// Get logs for current user
	logs, err := h.service.GetByUserID(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user audit logs: user_id=%d error=%v", userID, err)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	// Count total logs for this user
	filters := domain.AuditLogFilters{
		UserID: &userID,
	}
	count, err := h.service.Count(filters)
	if err != nil {
		h.logger.Error("Failed to count user audit logs: user_id=%d error=%v", userID, err)
		count = 0
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CleanupOldLogs handles POST /api/admin/audit-logs/cleanup
// @Summary      Cleanup old audit logs (Admin)
// @Description  Delete audit logs older than specified retention period
// @Tags         audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Retention days (must be positive)"
// @Success      200 {object} map[string]interface{} "Deleted count and success message"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/audit-logs/cleanup [post]
func (h *AuditLogHandler) CleanupOldLogs(w http.ResponseWriter, r *http.Request) {
	// Only admins can cleanup audit logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var request struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.RetentionDays <= 0 {
		http.Error(w, "retention_days must be positive", http.StatusBadRequest)
		return
	}

	deletedCount, err := h.service.CleanupOldLogs(request.RetentionDays)
	if err != nil {
		h.logger.Error("Failed to cleanup old audit logs: retention_days=%d error=%v", request.RetentionDays, err)
		http.Error(w, "Failed to cleanup audit logs", http.StatusInternalServerError)
		return
	}

	adminUserID, _ := middleware.GetUserID(r.Context())
	h.logger.Info("Cleaned up old audit logs: retention_days=%d deleted_count=%d admin_user_id=%d", request.RetentionDays, deletedCount, adminUserID)

	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"message":       "Old audit logs deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
