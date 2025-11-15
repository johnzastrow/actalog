package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// AuditLogService handles audit logging business logic
type AuditLogService struct {
	repo domain.AuditLogRepository
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(repo domain.AuditLogRepository) *AuditLogService {
	return &AuditLogService{
		repo: repo,
	}
}

// Log creates a new audit log entry
func (s *AuditLogService) Log(log *domain.AuditLog) error {
	return s.repo.Create(log)
}

// LogEvent is a helper function to log an event with minimal parameters
func (s *AuditLogService) LogEvent(eventType string, userID *int64, targetUserID *int64, ipAddress *string, userAgent *string, details map[string]interface{}) error {
	log := &domain.AuditLog{
		UserID:       userID,
		TargetUserID: targetUserID,
		EventType:    eventType,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// Convert details map to JSON string
	if details != nil && len(details) > 0 {
		detailsJSON, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
		detailsStr := string(detailsJSON)
		log.Details = &detailsStr
	}

	return s.repo.Create(log)
}

// GetByID retrieves a single audit log by ID
func (s *AuditLogService) GetByID(id int64) (*domain.AuditLog, error) {
	return s.repo.GetByID(id)
}

// List retrieves audit logs with pagination and filters
func (s *AuditLogService) List(filters domain.AuditLogFilters, limit, offset int) ([]*domain.AuditLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(filters, limit, offset)
}

// Count returns the total number of audit logs matching the filters
func (s *AuditLogService) Count(filters domain.AuditLogFilters) (int, error) {
	return s.repo.Count(filters)
}

// GetByUserID retrieves all audit logs for a specific user (actions performed BY the user)
func (s *AuditLogService) GetByUserID(userID int64, limit, offset int) ([]*domain.AuditLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByUserID(userID, limit, offset)
}

// GetByTargetUserID retrieves all audit logs affecting a specific user
func (s *AuditLogService) GetByTargetUserID(targetUserID int64, limit, offset int) ([]*domain.AuditLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByTargetUserID(targetUserID, limit, offset)
}

// CleanupOldLogs deletes audit logs older than the specified duration
func (s *AuditLogService) CleanupOldLogs(retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, fmt.Errorf("retention days must be positive")
	}

	before := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOlderThan(before)
}

// Helper functions for common audit log scenarios

// LogLoginSuccess logs a successful login
func (s *AuditLogService) LogLoginSuccess(userID int64, ipAddress, userAgent string) error {
	return s.LogEvent(domain.EventLoginSuccess, &userID, nil, &ipAddress, &userAgent, nil)
}

// LogLoginFailed logs a failed login attempt
func (s *AuditLogService) LogLoginFailed(email, reason, ipAddress, userAgent string, attemptsRemaining int) error {
	details := map[string]interface{}{
		"email":              email,
		"reason":             reason,
		"attempts_remaining": attemptsRemaining,
	}
	return s.LogEvent(domain.EventLoginFailed, nil, nil, &ipAddress, &userAgent, details)
}

// LogAccountLocked logs when an account is automatically locked
func (s *AuditLogService) LogAccountLocked(targetUserID int64, email, ipAddress, userAgent string, failedAttempts int) error {
	details := map[string]interface{}{
		"email":           email,
		"failed_attempts": failedAttempts,
		"locked_by":       "system",
	}
	return s.LogEvent(domain.EventAccountLockedAuto, nil, &targetUserID, &ipAddress, &userAgent, details)
}

// LogAccountUnlocked logs when an admin unlocks an account
func (s *AuditLogService) LogAccountUnlocked(adminUserID, targetUserID int64, adminEmail, targetEmail string) error {
	details := map[string]interface{}{
		"target_email":       targetEmail,
		"unlocked_by_admin":  adminEmail,
		"unlocked_by_user_id": adminUserID,
	}
	return s.LogEvent(domain.EventAccountUnlockedAdmin, &adminUserID, &targetUserID, nil, nil, details)
}

// LogAccountDisabled logs when an admin disables an account
func (s *AuditLogService) LogAccountDisabled(adminUserID, targetUserID int64, adminEmail, targetEmail, reason string) error {
	details := map[string]interface{}{
		"target_email":      targetEmail,
		"disabled_by_admin": adminEmail,
		"disabled_by_user_id": adminUserID,
		"reason":            reason,
	}
	return s.LogEvent(domain.EventAccountDisabled, &adminUserID, &targetUserID, nil, nil, details)
}

// LogAccountEnabled logs when an admin enables an account
func (s *AuditLogService) LogAccountEnabled(adminUserID, targetUserID int64, adminEmail, targetEmail string) error {
	details := map[string]interface{}{
		"target_email":     targetEmail,
		"enabled_by_admin": adminEmail,
		"enabled_by_user_id": adminUserID,
	}
	return s.LogEvent(domain.EventAccountEnabled, &adminUserID, &targetUserID, nil, nil, details)
}

// LogPasswordChanged logs when a user changes their password
func (s *AuditLogService) LogPasswordChanged(userID int64, ipAddress, userAgent string) error {
	return s.LogEvent(domain.EventPasswordChanged, &userID, nil, &ipAddress, &userAgent, nil)
}

// LogPasswordReset logs when a user resets their password via email
func (s *AuditLogService) LogPasswordReset(userID int64, email, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"email": email,
	}
	return s.LogEvent(domain.EventPasswordReset, &userID, nil, &ipAddress, &userAgent, details)
}

// LogEmailChanged logs when a user changes their email
func (s *AuditLogService) LogEmailChanged(userID int64, oldEmail, newEmail, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"old_email": oldEmail,
		"new_email": newEmail,
	}
	return s.LogEvent(domain.EventEmailChanged, &userID, nil, &ipAddress, &userAgent, details)
}

// LogEmailVerified logs when a user verifies their email
func (s *AuditLogService) LogEmailVerified(userID int64, email string) error {
	details := map[string]interface{}{
		"email": email,
	}
	return s.LogEvent(domain.EventEmailVerified, &userID, nil, nil, nil, details)
}

// LogRoleChanged logs when an admin changes a user's role
func (s *AuditLogService) LogRoleChanged(adminUserID, targetUserID int64, adminEmail, targetEmail, oldRole, newRole string) error {
	details := map[string]interface{}{
		"target_email":    targetEmail,
		"changed_by_admin": adminEmail,
		"changed_by_user_id": adminUserID,
		"old_role":        oldRole,
		"new_role":        newRole,
	}
	return s.LogEvent(domain.EventRoleChanged, &adminUserID, &targetUserID, nil, nil, details)
}

// LogUserCreated logs when a new user is created
func (s *AuditLogService) LogUserCreated(userID int64, email, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"email": email,
	}
	return s.LogEvent(domain.EventUserCreated, &userID, nil, &ipAddress, &userAgent, details)
}

// LogRateLimitExceeded logs when a rate limit is exceeded
func (s *AuditLogService) LogRateLimitExceeded(endpoint, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"endpoint": endpoint,
	}
	return s.LogEvent(domain.EventRateLimitExceeded, nil, nil, &ipAddress, &userAgent, details)
}
