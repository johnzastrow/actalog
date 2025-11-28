package domain

import "time"

// AuditLog represents a security or system event that should be tracked
type AuditLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       *int64    `json:"user_id,omitempty" db:"user_id"`               // User who performed the action (NULL for system)
	TargetUserID *int64    `json:"target_user_id,omitempty" db:"target_user_id"` // User affected by the action
	EventType    string    `json:"event_type" db:"event_type"`
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	Details      *string   `json:"details,omitempty" db:"details"` // JSON string
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// These fields are populated via JOIN queries and are not in the audit_logs table
	UserEmail       *string `json:"user_email,omitempty" db:"user_email"`               // Email of user who performed action
	TargetUserEmail *string `json:"target_user_email,omitempty" db:"target_user_email"` // Email of affected user
}

// Audit Event Types
const (
	// Authentication Events
	EventLoginSuccess = "login_success"
	EventLoginFailed  = "login_failed"
	EventLogout       = "logout"
	EventTokenRefresh = "token_refresh"

	// Account Security Events
	EventAccountLockedAuto    = "account_locked_auto"    // System locked after failed attempts
	EventAccountUnlockedAdmin = "account_unlocked_admin" // Admin unlocked account
	EventAccountDisabled      = "account_disabled"       // Admin disabled account
	EventAccountEnabled       = "account_enabled"        // Admin enabled account

	// Password Events
	EventPasswordChanged = "password_changed"
	EventPasswordReset   = "password_reset"

	// Email Events
	EventEmailChanged          = "email_changed"
	EventEmailVerificationSent = "email_verification_sent"
	EventEmailVerified         = "email_verified"

	// User Management Events
	EventUserCreated = "user_created"
	EventUserDeleted = "user_deleted"
	EventRoleChanged = "role_changed" // Admin promoted/demoted user

	// Rate Limiting Events
	EventRateLimitExceeded = "rate_limit_exceeded"
)

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	// Create creates a new audit log entry
	Create(log *AuditLog) error

	// GetByID retrieves a single audit log by ID
	GetByID(id int64) (*AuditLog, error)

	// List retrieves audit logs with pagination and optional filters
	List(filters AuditLogFilters, limit, offset int) ([]*AuditLog, error)

	// Count returns the total number of audit logs matching the filters
	Count(filters AuditLogFilters) (int, error)

	// GetByUserID retrieves all audit logs for a specific user (actions performed BY the user)
	GetByUserID(userID int64, limit, offset int) ([]*AuditLog, error)

	// GetByTargetUserID retrieves all audit logs affecting a specific user
	GetByTargetUserID(targetUserID int64, limit, offset int) ([]*AuditLog, error)

	// DeleteOlderThan deletes audit logs older than the specified duration (for cleanup)
	DeleteOlderThan(before time.Time) (int, error)
}

// AuditLogFilters represents filter options for querying audit logs
type AuditLogFilters struct {
	UserID       *int64     // Filter by user who performed action
	TargetUserID *int64     // Filter by user affected by action
	EventType    *string    // Filter by event type
	IPAddress    *string    // Filter by IP address
	StartDate    *time.Time // Filter logs after this date
	EndDate      *time.Time // Filter logs before this date
}
