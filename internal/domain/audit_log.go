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
	EventUserCreated        = "user_created"
	EventUserUpdated        = "user_updated"
	EventUserDeleted        = "user_deleted"
	EventRoleChanged        = "role_changed" // Admin promoted/demoted user
	EventProfileUpdated     = "profile_updated"
	EventUserSettingsUpdate = "user_settings_updated"

	// Organization Events
	EventOrganizationCreated = "organization_created"
	EventOrganizationUpdated = "organization_updated"
	EventOrganizationDeleted = "organization_deleted"

	// User-Organization Events
	EventUserAssignedToOrg  = "user_assigned_to_organization"
	EventUserRemovedFromOrg = "user_removed_from_organization"

	// Movement Events
	EventMovementCreated = "movement_created"
	EventMovementUpdated = "movement_updated"
	EventMovementDeleted = "movement_deleted"

	// WOD Events
	EventWODCreated = "wod_created"
	EventWODUpdated = "wod_updated"
	EventWODDeleted = "wod_deleted"

	// Workout Template Events
	EventWorkoutTemplateCreated = "workout_template_created"
	EventWorkoutTemplateUpdated = "workout_template_updated"
	EventWorkoutTemplateDeleted = "workout_template_deleted"

	// User Workout Events (logged workouts)
	EventUserWorkoutLogged  = "user_workout_logged"
	EventUserWorkoutUpdated = "user_workout_updated"
	EventUserWorkoutDeleted = "user_workout_deleted"

	// Rate Limiting Events
	EventRateLimitExceeded = "rate_limit_exceeded"

	// Subscription Events
	EventSubscriptionCreated       = "subscription_created"
	EventSubscriptionMarkedPaid    = "subscription_marked_paid"
	EventSubscriptionCancelled     = "subscription_cancelled"
	EventSubscriptionExpired       = "subscription_expired"
	EventSubscriptionUpdated       = "subscription_updated"
	EventOrgSubscriptionCreated    = "org_subscription_created"
	EventOrgSubscriptionMarkedPaid = "org_subscription_marked_paid"
	EventOrgSubscriptionCancelled  = "org_subscription_cancelled"
	EventOrgSubscriptionExpired    = "org_subscription_expired"
	EventOrgSubscriptionUpdated    = "org_subscription_updated"

	// User Import/Export Events
	EventUserImportPreview         = "user_import_preview"
	EventUserImportConfirm         = "user_import_confirm"
	EventUserImported              = "user_imported"
	EventUserExport                = "user_export"
	EventBatchPasswordResetRequest = "batch_password_reset_request"
	EventBatchPasswordResetSent    = "batch_password_reset_sent"

	// Class Scheduling Events - Locations
	EventGymLocationCreated = "gym_location_created"
	EventGymLocationUpdated = "gym_location_updated"
	EventGymLocationDeleted = "gym_location_deleted"

	// Class Scheduling Events - Templates
	EventClassTemplateCreated = "class_template_created"
	EventClassTemplateUpdated = "class_template_updated"
	EventClassTemplateDeleted = "class_template_deleted"

	// Class Scheduling Events - Schedule Slots
	EventScheduleSlotCreated = "schedule_slot_created"
	EventScheduleSlotUpdated = "schedule_slot_updated"
	EventScheduleSlotDeleted = "schedule_slot_deleted"

	// Class Scheduling Events - Sessions
	EventClassSessionCreated   = "class_session_created"
	EventClassSessionUpdated   = "class_session_updated"
	EventClassSessionCancelled = "class_session_cancelled"
	EventClassSessionCompleted = "class_session_completed"

	// Class Scheduling Events - Coach Assignments
	EventCoachAssigned   = "coach_assigned"
	EventCoachUnassigned = "coach_unassigned"

	// Class Scheduling Events - Template Coaches
	EventTemplateCoachAdded   = "template_coach_added"
	EventTemplateCoachRemoved = "template_coach_removed"

	// Class Scheduling Events - Batch Operations
	EventSessionWorkoutBatchUpdated = "session_workout_batch_updated"

	// Class Scheduling Events - Reservations
	EventReservationCreated   = "reservation_created"
	EventReservationCancelled = "reservation_cancelled"
	EventReservationCheckedIn = "reservation_checked_in"
	EventReservationNoShow    = "reservation_no_show"
	EventReservationAttended  = "reservation_attended"

	// Phase 4 Events - Documents
	EventDocumentCreated       = "document_created"
	EventDocumentUpdated       = "document_updated"
	EventDocumentDeleted       = "document_deleted"
	EventUserDocumentCompleted = "user_document_completed"

	// Phase 4 Events - Class Packages
	EventClassPackageCreated = "class_package_created"
	EventClassPackageUpdated = "class_package_updated"
	EventClassPackageDeleted = "class_package_deleted"

	// Phase 4 Events - Credits
	EventCreditsAdded = "credits_added"

	// Phase 4 Events - Waitlist
	EventWaitlistJoined   = "waitlist_joined"
	EventWaitlistCancelled = "waitlist_cancelled"
	EventWaitlistPromoted = "waitlist_promoted"
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
