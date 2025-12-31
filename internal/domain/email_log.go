package domain

import "time"

// EmailLog represents a record of an email send attempt
type EmailLog struct {
	ID             int64     `json:"id" db:"id"`
	RecipientEmail string    `json:"recipient_email" db:"recipient_email"`
	EmailType      string    `json:"email_type" db:"email_type"` // "test", "password_reset", "verification", "notification"
	Subject        string    `json:"subject" db:"subject"`
	Success        bool      `json:"success" db:"success"`
	ErrorMessage   *string   `json:"error_message,omitempty" db:"error_message"`
	DebugInfo      *string   `json:"debug_info,omitempty" db:"debug_info"` // JSON of SMTPDebugInfo
	SentByUserID   *int64    `json:"sent_by_user_id,omitempty" db:"sent_by_user_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`

	// These fields are populated via JOIN queries and are not in the email_logs table
	SentByUserEmail *string `json:"sent_by_user_email,omitempty" db:"sent_by_user_email"`
}

// Email Types
const (
	EmailTypeTest          = "test"
	EmailTypePasswordReset = "password_reset"
	EmailTypeVerification  = "verification"
	EmailTypeNotification  = "notification"
)

// EmailLogRepository defines the interface for email log data access
type EmailLogRepository interface {
	// Create creates a new email log entry
	Create(log *EmailLog) error

	// GetByID retrieves a single email log by ID
	GetByID(id int64) (*EmailLog, error)

	// List retrieves email logs with pagination and optional filters
	List(filters EmailLogFilters, limit, offset int) ([]*EmailLog, error)

	// Count returns the total number of email logs matching the filters
	Count(filters EmailLogFilters) (int, error)

	// DeleteOlderThan deletes email logs older than the specified duration (for cleanup)
	DeleteOlderThan(before time.Time) (int64, error)
}

// EmailLogFilters represents filter options for querying email logs
type EmailLogFilters struct {
	EmailType      *string    // Filter by email type
	RecipientEmail *string    // Filter by recipient email (partial match)
	Success        *bool      // Filter by success status
	StartDate      *time.Time // Filter logs after this date
	EndDate        *time.Time // Filter logs before this date
	SentByUserID   *int64     // Filter by user who triggered the email
}
