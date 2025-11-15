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
func (h *AuditLogHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	// Only admins can access individual audit logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		h.logger.Warn("Unauthorized access attempt to audit log",
			"user_role", userRole,
			"path", r.URL.Path,
		)
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
		h.logger.Error("Failed to get audit log",
			"id", id,
			"error", err.Error(),
		)
		http.Error(w, "Audit log not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

// ListAuditLogs handles GET /api/audit-logs
// Query params: user_id, target_user_id, event_type, ip_address, start_date, end_date, limit, offset
func (h *AuditLogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Only admins can list all audit logs
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		h.logger.Warn("Unauthorized access attempt to audit logs list",
			"user_role", userRole,
		)
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
		h.logger.Error("Failed to list audit logs",
			"error", err.Error(),
		)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	count, err := h.service.Count(filters)
	if err != nil {
		h.logger.Error("Failed to count audit logs",
			"error", err.Error(),
		)
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
// Returns audit logs for the current user only
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
		h.logger.Error("Failed to get user audit logs",
			"user_id", userID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	// Count total logs for this user
	filters := domain.AuditLogFilters{
		UserID: &userID,
	}
	count, err := h.service.Count(filters)
	if err != nil {
		h.logger.Error("Failed to count user audit logs",
			"user_id", userID,
			"error", err.Error(),
		)
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
// Admin endpoint to delete old audit logs
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
		h.logger.Error("Failed to cleanup old audit logs",
			"retention_days", request.RetentionDays,
			"error", err.Error(),
		)
		http.Error(w, "Failed to cleanup audit logs", http.StatusInternalServerError)
		return
	}

	adminUserID, _ := middleware.GetUserID(r.Context())
	h.logger.Info("Cleaned up old audit logs",
		"retention_days", request.RetentionDays,
		"deleted_count", deletedCount,
		"admin_user_id", adminUserID,
	)

	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"message":       "Old audit logs deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
