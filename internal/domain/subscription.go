package domain

import "time"

// SubscriptionType represents subscription billing types
type SubscriptionType string

const (
	SubscriptionTypeFree    SubscriptionType = "free"
	SubscriptionTypeMonthly SubscriptionType = "monthly"
	SubscriptionTypeAnnual  SubscriptionType = "annual"
)

// SubscriptionStatus represents subscription states
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

// UserSubscription represents a user-level subscription
type UserSubscription struct {
	ID               int64              `json:"id" db:"id"`
	UserID           int64              `json:"user_id" db:"user_id"`
	UserEmail        string             `json:"user_email,omitempty" db:"user_email"` // Joined from users table
	UserName         string             `json:"user_name,omitempty" db:"user_name"`   // Joined from users table
	SubscriptionType SubscriptionType   `json:"subscription_type" db:"subscription_type"`
	Status           SubscriptionStatus `json:"status" db:"status"`
	IsPermanentFree  bool               `json:"is_permanent_free" db:"is_permanent_free"`
	StartDate        time.Time          `json:"start_date" db:"start_date"`
	EndDate          *time.Time         `json:"end_date,omitempty" db:"end_date"`
	LastPaymentDate  *time.Time         `json:"last_payment_date,omitempty" db:"last_payment_date"`
	NextBillingDate  *time.Time         `json:"next_billing_date,omitempty" db:"next_billing_date"`
	CancelledAt      *time.Time         `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason  *string            `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	Notes            *string            `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
	CreatedByUserID  *int64             `json:"created_by_user_id,omitempty" db:"created_by_user_id"`
}

// OrganizationSubscription represents an organization-level subscription
type OrganizationSubscription struct {
	ID               int64              `json:"id" db:"id"`
	OrganizationID   int64              `json:"organization_id" db:"organization_id"`
	OrganizationName string             `json:"organization_name,omitempty" db:"organization_name"` // Joined from organizations table
	SubscriptionType SubscriptionType   `json:"subscription_type" db:"subscription_type"`
	Status           SubscriptionStatus `json:"status" db:"status"`
	IsPermanentFree  bool               `json:"is_permanent_free" db:"is_permanent_free"`
	StartDate        time.Time          `json:"start_date" db:"start_date"`
	EndDate          *time.Time         `json:"end_date,omitempty" db:"end_date"`
	LastPaymentDate  *time.Time         `json:"last_payment_date,omitempty" db:"last_payment_date"`
	NextBillingDate  *time.Time         `json:"next_billing_date,omitempty" db:"next_billing_date"`
	CancelledAt      *time.Time         `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason  *string            `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	Notes            *string            `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
	CreatedByUserID  *int64             `json:"created_by_user_id,omitempty" db:"created_by_user_id"`
}

// SubscriptionAccessResult represents the result of checking user access
type SubscriptionAccessResult struct {
	HasAccess        bool                        `json:"has_access"`
	Source           string                      `json:"source"` // "user", "organization", "both", "none"
	UserSubscription *UserSubscription           `json:"user_subscription,omitempty"`
	OrgSubscriptions []*OrganizationSubscription `json:"org_subscriptions,omitempty"`
	ExpiresAt        *time.Time                  `json:"expires_at,omitempty"`
}

// UserSubscriptionRepository defines repository interface for user subscriptions
type UserSubscriptionRepository interface {
	Create(sub *UserSubscription) error
	GetByID(id int64) (*UserSubscription, error)
	GetActiveByUserID(userID int64) (*UserSubscription, error)
	GetByUserID(userID int64) ([]*UserSubscription, error)
	Update(sub *UserSubscription) error
	Delete(id int64) error
	ListAll() ([]*UserSubscription, error)

	// Admin operations
	MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64, durationDays *int) error
	MarkAsExpired(id int64) error
	Cancel(id int64, reason string, adminUserID int64) error
}

// OrganizationSubscriptionRepository defines repository interface for org subscriptions
type OrganizationSubscriptionRepository interface {
	Create(sub *OrganizationSubscription) error
	GetByID(id int64) (*OrganizationSubscription, error)
	GetActiveByOrganizationID(orgID int64) (*OrganizationSubscription, error)
	GetByOrganizationID(orgID int64) ([]*OrganizationSubscription, error)
	Update(sub *OrganizationSubscription) error
	Delete(id int64) error
	ListAll() ([]*OrganizationSubscription, error)

	// Admin operations
	MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64, durationDays *int) error
	MarkAsExpired(id int64) error
	Cancel(id int64, reason string, adminUserID int64) error
}

// SubscriptionAccessRepository provides fast access checking
type SubscriptionAccessRepository interface {
	// CheckUserAccess returns whether user has active subscription
	// Checks both user-level and org-level subscriptions
	CheckUserAccess(userID int64) (*SubscriptionAccessResult, error)
}
