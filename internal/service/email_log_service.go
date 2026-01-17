package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/email"
)

// EmailLogRetentionDays is the default retention period for email logs (90 days)
const EmailLogRetentionDays = 90

// EmailLogService handles email log business logic
type EmailLogService struct {
	repo domain.EmailLogRepository
}

// NewEmailLogService creates a new email log service
func NewEmailLogService(repo domain.EmailLogRepository) *EmailLogService {
	return &EmailLogService{
		repo: repo,
	}
}

// LogEmailAttempt creates a new email log entry
func (s *EmailLogService) LogEmailAttempt(
	recipientEmail, emailType, subject string,
	success bool, errorMessage *string,
	debugInfo *email.SMTPDebugInfo,
	sentByUserID *int64,
) error {
	log := &domain.EmailLog{
		RecipientEmail: recipientEmail,
		EmailType:      emailType,
		Subject:        subject,
		Success:        success,
		ErrorMessage:   errorMessage,
		SentByUserID:   sentByUserID,
	}

	// Convert debug info to JSON string if present
	if debugInfo != nil {
		debugJSON, err := json.Marshal(debugInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal debug info: %w", err)
		}
		debugStr := string(debugJSON)
		log.DebugInfo = &debugStr
	}

	return s.repo.Create(log)
}

// LogTestEmail logs a test email attempt
func (s *EmailLogService) LogTestEmail(recipientEmail, subject string, result email.SendResult, sentByUserID int64) error {
	var errorMsg *string
	if result.Error != nil {
		errStr := result.Error.Error()
		errorMsg = &errStr
	}

	return s.LogEmailAttempt(
		recipientEmail,
		domain.EmailTypeTest,
		subject,
		result.Success,
		errorMsg,
		result.DebugInfo,
		&sentByUserID,
	)
}

// LogPasswordResetEmail logs a password reset email attempt
func (s *EmailLogService) LogPasswordResetEmail(recipientEmail, subject string, result email.SendResult) error {
	var errorMsg *string
	if result.Error != nil {
		errStr := result.Error.Error()
		errorMsg = &errStr
	}

	return s.LogEmailAttempt(
		recipientEmail,
		domain.EmailTypePasswordReset,
		subject,
		result.Success,
		errorMsg,
		result.DebugInfo,
		nil,
	)
}

// LogVerificationEmail logs an email verification email attempt
func (s *EmailLogService) LogVerificationEmail(recipientEmail, subject string, result email.SendResult) error {
	var errorMsg *string
	if result.Error != nil {
		errStr := result.Error.Error()
		errorMsg = &errStr
	}

	return s.LogEmailAttempt(
		recipientEmail,
		domain.EmailTypeVerification,
		subject,
		result.Success,
		errorMsg,
		result.DebugInfo,
		nil,
	)
}

// LogNotificationEmail logs a notification email attempt
func (s *EmailLogService) LogNotificationEmail(recipientEmail, subject string, result email.SendResult) error {
	var errorMsg *string
	if result.Error != nil {
		errStr := result.Error.Error()
		errorMsg = &errStr
	}

	return s.LogEmailAttempt(
		recipientEmail,
		domain.EmailTypeNotification,
		subject,
		result.Success,
		errorMsg,
		result.DebugInfo,
		nil,
	)
}

// GetByID retrieves a single email log by ID
func (s *EmailLogService) GetByID(id int64) (*domain.EmailLog, error) {
	return s.repo.GetByID(id)
}

// List retrieves email logs with pagination and filters
func (s *EmailLogService) List(filters domain.EmailLogFilters, limit, offset int) ([]*domain.EmailLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(filters, limit, offset)
}

// Count returns the total number of email logs matching the filters
func (s *EmailLogService) Count(filters domain.EmailLogFilters) (int, error) {
	return s.repo.Count(filters)
}

// CleanupOldLogs deletes email logs older than the specified duration
func (s *EmailLogService) CleanupOldLogs(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = EmailLogRetentionDays
	}

	before := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOlderThan(before)
}

// GetRecentFailures retrieves recent failed email attempts
func (s *EmailLogService) GetRecentFailures(limit int) ([]*domain.EmailLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	failed := false
	filters := domain.EmailLogFilters{
		Success: &failed,
	}

	return s.repo.List(filters, limit, 0)
}

// GetEmailStats returns statistics about email logs
func (s *EmailLogService) GetEmailStats() (*EmailStats, error) {
	stats := &EmailStats{
		CountByType: make(map[string]int),
	}

	// Count total emails
	total, err := s.repo.Count(domain.EmailLogFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}
	stats.TotalCount = total

	// Count successful emails
	success := true
	successCount, err := s.repo.Count(domain.EmailLogFilters{Success: &success})
	if err != nil {
		return nil, fmt.Errorf("failed to count successful emails: %w", err)
	}
	stats.SuccessCount = successCount

	// Count failed emails
	failed := false
	failedCount, err := s.repo.Count(domain.EmailLogFilters{Success: &failed})
	if err != nil {
		return nil, fmt.Errorf("failed to count failed emails: %w", err)
	}
	stats.FailedCount = failedCount

	// Count by type
	for _, emailType := range []string{
		domain.EmailTypeTest,
		domain.EmailTypePasswordReset,
		domain.EmailTypeVerification,
		domain.EmailTypeNotification,
	} {
		typeStr := emailType
		count, err := s.repo.Count(domain.EmailLogFilters{EmailType: &typeStr})
		if err != nil {
			return nil, fmt.Errorf("failed to count %s emails: %w", emailType, err)
		}
		stats.CountByType[emailType] = count
	}

	// Calculate success rate
	if stats.TotalCount > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalCount) * 100
	}

	return stats, nil
}

// EmailStats holds statistics about email logs
type EmailStats struct {
	TotalCount   int            `json:"total_count"`
	SuccessCount int            `json:"success_count"`
	FailedCount  int            `json:"failed_count"`
	SuccessRate  float64        `json:"success_rate"`
	CountByType  map[string]int `json:"count_by_type"`
}

// NewEmailStats creates a new EmailStats with initialized map
func NewEmailStats() *EmailStats {
	return &EmailStats{
		CountByType: make(map[string]int),
	}
}
