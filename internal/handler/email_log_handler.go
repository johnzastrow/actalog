package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// EmailLogHandler handles email log endpoints
type EmailLogHandler struct {
	emailLogService *service.EmailLogService
	logger          *logger.Logger
}

// NewEmailLogHandler creates a new email log handler
func NewEmailLogHandler(emailLogService *service.EmailLogService, logger *logger.Logger) *EmailLogHandler {
	return &EmailLogHandler{
		emailLogService: emailLogService,
		logger:          logger,
	}
}

// ListEmailLogs handles GET /api/admin/email-logs
// Returns a paginated list of email logs with optional filters
func (h *EmailLogHandler) ListEmailLogs(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Pagination
	limit := 50
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := domain.EmailLogFilters{}

	if emailType := query.Get("type"); emailType != "" {
		filters.EmailType = &emailType
	}

	if recipient := query.Get("recipient"); recipient != "" {
		filters.RecipientEmail = &recipient
	}

	if successStr := query.Get("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filters.Success = &success
		}
	}

	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Add one day to include the end date
			endDate = endDate.Add(24 * time.Hour)
			filters.EndDate = &endDate
		}
	}

	// Get logs
	logs, err := h.emailLogService.List(filters, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch email logs")
		return
	}

	// Get total count
	count, err := h.emailLogService.Count(filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count email logs")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":   logs,
		"count":  count,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEmailLog handles GET /api/admin/email-logs/{id}
// Returns a single email log by ID with full debug info
func (h *EmailLogHandler) GetEmailLog(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid email log ID")
		return
	}

	// Get log
	log, err := h.emailLogService.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Email log not found")
		return
	}

	respondJSON(w, http.StatusOK, log)
}

// GetEmailStats handles GET /api/admin/email-logs/stats
// Returns statistics about email logs
func (h *EmailLogHandler) GetEmailStats(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.emailLogService.GetEmailStats()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get email statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// CleanupEmailLogs handles POST /api/admin/email-logs/cleanup
// Removes email logs older than the retention period
func (h *EmailLogHandler) CleanupEmailLogs(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Use default retention period (90 days)
	deleted, err := h.emailLogService.CleanupOldLogs(service.EmailLogRetentionDays)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to cleanup email logs")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Email logs cleaned up successfully",
		"deleted_count":  deleted,
		"retention_days": service.EmailLogRetentionDays,
	})
}

// GetRecentFailures handles GET /api/admin/email-logs/failures
// Returns recent failed email attempts
func (h *EmailLogHandler) GetRecentFailures(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse limit
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	failures, err := h.emailLogService.GetRecentFailures(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch recent failures")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"failures": failures,
		"count":    len(failures),
	})
}
