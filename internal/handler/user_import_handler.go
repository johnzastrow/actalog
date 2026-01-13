package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// UserImportHandler handles user import/export and batch operations
type UserImportHandler struct {
	userImportService *service.UserImportService
	logger            *logger.Logger
}

// NewUserImportHandler creates a new user import handler
func NewUserImportHandler(userImportService *service.UserImportService, logger *logger.Logger) *UserImportHandler {
	return &UserImportHandler{
		userImportService: userImportService,
		logger:            logger,
	}
}

// PreviewUserImport validates and previews user CSV import
// @Summary      Preview user import
// @Description  Validate and preview user CSV import before confirming
// @Tags         admin-users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "CSV file to import"
// @Success      200 {object} service.UserImportResult "Import preview with validation results"
// @Failure      400 {object} ErrorResponse "Invalid file or validation failed"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /admin/users/import/preview [post]
func (h *UserImportHandler) PreviewUserImport(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.logger.Error("action=preview_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("action=preview_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Preview import
	result, err := h.userImportService.PreviewUserImport(file)
	if err != nil {
		h.logger.Error("action=preview_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Import preview failed: %v", err))
		return
	}

	h.logger.Info("action=preview_user_import outcome=success total_rows=%d valid_rows=%d duplicate_rows=%d",
		result.TotalRows, result.ValidRows, result.DuplicateRows)
	respondJSON(w, http.StatusOK, result)
}

// ConfirmUserImport confirms and executes user CSV import
// @Summary      Confirm user import
// @Description  Execute user CSV import after preview
// @Tags         admin-users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "CSV file to import"
// @Param        skip_duplicates formData bool false "Skip duplicate entries"
// @Success      200 {object} service.UserImportResult "Import results with counts"
// @Failure      400 {object} ErrorResponse "Invalid file"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Import failed"
// @Router       /admin/users/import/confirm [post]
func (h *UserImportHandler) ConfirmUserImport(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.logger.Error("action=confirm_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("action=confirm_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Parse options from form data
	skipDuplicates := r.FormValue("skip_duplicates") == "true"

	// Confirm import
	result, err := h.userImportService.ConfirmUserImport(file, skipDuplicates)
	if err != nil {
		h.logger.Error("action=confirm_user_import outcome=failure error=%v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	h.logger.Info("action=confirm_user_import outcome=success created=%d skipped=%d",
		result.CreatedCount, result.SkippedCount)
	respondJSON(w, http.StatusOK, result)
}

// ExportUsers exports users to CSV format
// @Summary      Export users
// @Description  Export users to CSV format (email, name)
// @Tags         admin-users
// @Produce      text/csv
// @Security     BearerAuth
// @Success      200 {file} file "Exported CSV file"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/export [get]
func (h *UserImportHandler) ExportUsers(w http.ResponseWriter, r *http.Request) {
	data, err := h.userImportService.ExportUsersToCSV()
	if err != nil {
		h.logger.Error("action=export_users outcome=failure error=%v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export users: %v", err))
		return
	}

	// Set response headers for download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=users_export.csv")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// Write data
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)

	h.logger.Info("action=export_users outcome=success")
}

// ListUsersWithFilter lists users with filter criteria
// @Summary      List users with filter
// @Description  Get paginated list of users with optional search and date filters
// @Tags         admin-users
// @Produce      json
// @Security     BearerAuth
// @Param        search query string false "Search in name and email"
// @Param        created_after query string false "Filter users created after this date (YYYY-MM-DD)"
// @Param        created_before query string false "Filter users created before this date (YYYY-MM-DD)"
// @Param        limit query int false "Number of results per page (default 50)"
// @Param        offset query int false "Offset for pagination"
// @Success      200 {object} map[string]interface{} "Users list with total count"
// @Failure      400 {object} ErrorResponse "Invalid parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/filter [get]
func (h *UserImportHandler) ListUsersWithFilter(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := domain.UserListFilter{
		Search: r.URL.Query().Get("search"),
	}

	// Parse date filters
	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		parsed, err := time.Parse("2006-01-02", createdAfter)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid created_after format. Use YYYY-MM-DD")
			return
		}
		filter.CreatedAfter = &parsed
	}

	if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
		parsed, err := time.Parse("2006-01-02", createdBefore)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid created_before format. Use YYYY-MM-DD")
			return
		}
		// Add 1 day to include the entire day
		parsed = parsed.Add(24 * time.Hour)
		filter.CreatedBefore = &parsed
	}

	// Parse pagination
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 1000 {
				limit = 1000 // Cap at 1000
			}
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	users, count, err := h.userImportService.ListUsersWithFilter(filter, limit, offset)
	if err != nil {
		h.logger.Error("action=list_users_with_filter outcome=failure error=%v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list users: %v", err))
		return
	}

	// Clear password hashes before returning
	for _, user := range users {
		user.PasswordHash = ""
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": count,
	})
}

// BatchPasswordResetRequest represents the request body for batch password reset
type BatchPasswordResetRequest struct {
	UserIDs []int64 `json:"user_ids"`
}

// SendBatchPasswordResetEmails sends password reset emails to selected users
// @Summary      Send batch password reset emails
// @Description  Send password reset emails to multiple users at once
// @Tags         admin-users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body BatchPasswordResetRequest true "User IDs to send reset emails to"
// @Success      200 {object} service.BatchPasswordResetResult "Result with success/failure counts"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/batch-password-reset [post]
func (h *UserImportHandler) SendBatchPasswordResetEmails(w http.ResponseWriter, r *http.Request) {
	var req BatchPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("action=batch_password_reset outcome=failure error=%v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.UserIDs) == 0 {
		respondError(w, http.StatusBadRequest, "No user IDs provided")
		return
	}

	if len(req.UserIDs) > 100 {
		respondError(w, http.StatusBadRequest, "Maximum 100 users can be selected at once")
		return
	}

	result, err := h.userImportService.SendBatchPasswordResetEmails(req.UserIDs)
	if err != nil {
		h.logger.Error("action=batch_password_reset outcome=failure error=%v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send emails: %v", err))
		return
	}

	h.logger.Info("action=batch_password_reset outcome=success sent=%d failed=%d",
		result.SuccessfullySent, result.FailedCount)
	respondJSON(w, http.StatusOK, result)
}
