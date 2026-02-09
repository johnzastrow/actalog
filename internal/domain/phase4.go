// Package domain contains the Phase 4 business entities
package domain

import (
	"time"
)

// Document status constants
const (
	DocumentStatusPending   = "pending"
	DocumentStatusCompleted = "completed"
	DocumentStatusExpired   = "expired"
)

// Document types
const (
	DocumentTypeWaiver         = "waiver"
	DocumentTypeLiability      = "liability"
	DocumentTypeHealthForm     = "health_form"
	DocumentTypeEmergencyContact = "emergency_contact"
	DocumentTypeOther          = "other"
)

// Waitlist status constants
const (
	WaitlistStatusWaiting  = "waiting"
	WaitlistStatusPromoted = "promoted"
	WaitlistStatusExpired  = "expired"
	WaitlistStatusCancelled = "cancelled"
)

// Notification type constants
const (
	NotificationTypeClassReminder     = "class_reminder"
	NotificationTypeWaitlistPromotion = "waitlist_promotion"
	NotificationTypeClassCancelled    = "class_cancelled"
	NotificationTypeDocumentExpiring  = "document_expiring"
	NotificationTypeCreditsExpiring   = "credits_expiring"
)

// Notification status constants
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

// Document represents a document type that users may need to complete
type Document struct {
	ID               int64     `json:"id" db:"id"`
	OrganizationID   int64     `json:"organization_id" db:"organization_id"`
	Name             string    `json:"name" db:"name"`
	Description      *string   `json:"description,omitempty" db:"description"`
	DocumentType     string    `json:"document_type" db:"document_type"` // waiver, liability, health_form, etc.
	URL              *string   `json:"url,omitempty" db:"url"`           // Link to the document
	IsRequired       bool      `json:"is_required" db:"is_required"`
	ExpiresAfterDays *int      `json:"expires_after_days,omitempty" db:"expires_after_days"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UserDocument represents a user's completion status for a document
type UserDocument struct {
	ID               int64      `json:"id" db:"id"`
	UserID           int64      `json:"user_id" db:"user_id"`
	DocumentID       int64      `json:"document_id" db:"document_id"`
	Status           string     `json:"status" db:"status"` // pending, completed, expired
	CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	VerifiedByUserID *int64     `json:"verified_by_user_id,omitempty" db:"verified_by_user_id"`
	Notes            *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	DocumentName    *string `json:"document_name,omitempty"`
	DocumentType    *string `json:"document_type,omitempty"`
	UserName        *string `json:"user_name,omitempty"`
	UserEmail       *string `json:"user_email,omitempty"`
	VerifiedByName  *string `json:"verified_by_name,omitempty"`
	VerifiedByEmail *string `json:"verified_by_email,omitempty"`
}

// ClassPackage represents a credit package that can be purchased
type ClassPackage struct {
	ID             int64     `json:"id" db:"id"`
	OrganizationID int64     `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Credits        int       `json:"credits" db:"credits"`
	PriceCents     int       `json:"price_cents" db:"price_cents"`
	ValidityDays   *int      `json:"validity_days,omitempty" db:"validity_days"` // nil means never expires
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserClassCredits represents a user's purchased credit balance
type UserClassCredits struct {
	ID             int64      `json:"id" db:"id"`
	UserID         int64      `json:"user_id" db:"user_id"`
	OrganizationID int64      `json:"organization_id" db:"organization_id"`
	PackageID      *int64     `json:"package_id,omitempty" db:"package_id"`
	CreditsTotal   int        `json:"credits_total" db:"credits_total"`
	CreditsUsed    int        `json:"credits_used" db:"credits_used"`
	PurchasedAt    time.Time  `json:"purchased_at" db:"purchased_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// Computed fields
	CreditsRemaining int `json:"credits_remaining"`

	// Populated via JOINs
	PackageName *string `json:"package_name,omitempty"`
	UserName    *string `json:"user_name,omitempty"`
	UserEmail   *string `json:"user_email,omitempty"`
}

// WaitlistEntry represents a user's position on a session waitlist
type WaitlistEntry struct {
	ID         int64      `json:"id" db:"id"`
	SessionID  int64      `json:"session_id" db:"session_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Position   int        `json:"position" db:"position"`
	Status     string     `json:"status" db:"status"` // waiting, promoted, expired, cancelled
	JoinedAt   time.Time  `json:"joined_at" db:"joined_at"`
	PromotedAt *time.Time `json:"promoted_at,omitempty" db:"promoted_at"`
	ExpiredAt  *time.Time `json:"expired_at,omitempty" db:"expired_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	UserName    *string `json:"user_name,omitempty"`
	UserEmail   *string `json:"user_email,omitempty"`
	SessionName *string `json:"session_name,omitempty"`
}

// ClassNotification represents a scheduled notification for class-related events
type ClassNotification struct {
	ID              int64      `json:"id" db:"id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	SessionID       *int64     `json:"session_id,omitempty" db:"session_id"`
	ReservationID   *int64     `json:"reservation_id,omitempty" db:"reservation_id"`
	WaitlistEntryID *int64     `json:"waitlist_entry_id,omitempty" db:"waitlist_entry_id"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
	Status          string     `json:"status" db:"status"` // pending, sent, failed
	ScheduledFor    time.Time  `json:"scheduled_for" db:"scheduled_for"`
	SentAt          *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	ErrorMessage    *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	UserName    *string `json:"user_name,omitempty"`
	UserEmail   *string `json:"user_email,omitempty"`
	SessionName *string `json:"session_name,omitempty"`
}

// DocumentRepository defines the interface for document data access
type DocumentRepository interface {
	Create(document *Document) error
	GetByID(id int64) (*Document, error)
	GetByOrganizationID(orgID int64, includeInactive bool) ([]*Document, error)
	GetRequiredByOrganizationID(orgID int64) ([]*Document, error)
	Update(document *Document) error
	Delete(id int64) error
}

// UserDocumentRepository defines the interface for user document data access
type UserDocumentRepository interface {
	Create(userDoc *UserDocument) error
	GetByID(id int64) (*UserDocument, error)
	GetByUserID(userID int64) ([]*UserDocument, error)
	GetByUserAndDocument(userID, documentID int64) (*UserDocument, error)
	GetByUserAndOrganization(userID, orgID int64) ([]*UserDocument, error)
	GetPendingByUserAndOrganization(userID, orgID int64) ([]*UserDocument, error)
	GetExpiringSoon(daysAhead int) ([]*UserDocument, error)
	Update(userDoc *UserDocument) error
	MarkCompleted(id int64, verifiedByUserID *int64) error
	Delete(id int64) error
}

// ClassPackageRepository defines the interface for class package data access
type ClassPackageRepository interface {
	Create(pkg *ClassPackage) error
	GetByID(id int64) (*ClassPackage, error)
	GetByOrganizationID(orgID int64, includeInactive bool) ([]*ClassPackage, error)
	Update(pkg *ClassPackage) error
	Delete(id int64) error
}

// UserClassCreditsRepository defines the interface for user credits data access
type UserClassCreditsRepository interface {
	Create(credits *UserClassCredits) error
	GetByID(id int64) (*UserClassCredits, error)
	GetByUserID(userID int64) ([]*UserClassCredits, error)
	GetByUserAndOrganization(userID, orgID int64) ([]*UserClassCredits, error)
	GetAvailableCredits(userID, orgID int64) (int, error)
	UseCredit(userID, orgID int64) error    // Deducts 1 credit from oldest non-expired balance
	RefundCredit(userID, orgID int64) error // Restores 1 credit (opposite of UseCredit)
	GetExpiringSoon(daysAhead int) ([]*UserClassCredits, error)
	Update(credits *UserClassCredits) error
}

// WaitlistRepository defines the interface for waitlist data access
type WaitlistRepository interface {
	Create(entry *WaitlistEntry) error
	GetByID(id int64) (*WaitlistEntry, error)
	GetBySessionID(sessionID int64) ([]*WaitlistEntry, error)
	GetByUserID(userID int64) ([]*WaitlistEntry, error)
	GetBySessionAndUser(sessionID, userID int64) (*WaitlistEntry, error)
	GetNextInLine(sessionID int64) (*WaitlistEntry, error)
	GetPosition(sessionID, userID int64) (int, error)
	Promote(id int64) error
	Cancel(id int64) error
	ExpireOldEntries() (int, error)         // Expires entries for past sessions
	ReorderPositions(sessionID int64) error // Reorders positions after promotion/cancellation
	Delete(id int64) error
	// Bulk delete for template deletion
	DeleteBySessionIDs(sessionIDs []int64) error
}

// ClassNotificationRepository defines the interface for notification data access
type ClassNotificationRepository interface {
	Create(notification *ClassNotification) error
	GetByID(id int64) (*ClassNotification, error)
	GetPendingByScheduledTime(before time.Time) ([]*ClassNotification, error)
	GetByUserID(userID int64, limit int) ([]*ClassNotification, error)
	MarkSent(id int64) error
	MarkFailed(id int64, errorMessage string) error
	DeleteBySessionID(sessionID int64) error
	DeleteByReservationID(reservationID int64) error
	// Bulk delete for template deletion
	DeleteBySessionIDs(sessionIDs []int64) error
}
