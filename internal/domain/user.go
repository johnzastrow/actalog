// Package domain contains the core business entities
package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                         int64      `json:"id" db:"id"`
	Email                      string     `json:"email" db:"email"`
	PasswordHash               string     `json:"-" db:"password_hash"` // Never serialize password
	Name                       string     `json:"name" db:"name"`
	ProfileImage               *string    `json:"profile_image,omitempty" db:"profile_image"`
	Birthday                   *time.Time `json:"birthday,omitempty" db:"birthday"`
	Role                       string     `json:"role" db:"role"` // user, admin
	EmailVerified              bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt            *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	VerificationToken          *string    `json:"-" db:"verification_token"` // Never serialize verification token
	VerificationTokenExpiresAt *time.Time `json:"-" db:"verification_token_expires_at"`
	ResetToken                 *string    `json:"-" db:"reset_token"` // Never serialize reset token
	ResetTokenExpiresAt        *time.Time `json:"-" db:"reset_token_expires_at"`

	// Security fields (account lockout and disable)
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`           // Don't expose to client
	LockedAt            *time.Time `json:"-" db:"locked_at"`                       // Don't expose to client
	LockedUntil         *time.Time `json:"-" db:"locked_until"`                    // Don't expose to client
	AccountDisabled     bool       `json:"account_disabled" db:"account_disabled"` // Expose so client knows
	DisabledAt          *time.Time `json:"disabled_at,omitempty" db:"disabled_at"`
	DisabledByUserID    *int64     `json:"disabled_by_user_id,omitempty" db:"disabled_by_user_id"`
	DisableReason       *string    `json:"disable_reason,omitempty" db:"disable_reason"` // Reason for account disable

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// RefreshToken represents a refresh token for "Remember Me" functionality
type RefreshToken struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Token      string     `json:"token" db:"token"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	DeviceInfo string     `json:"device_info,omitempty" db:"device_info"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id int64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByResetToken(token string) (*User, error)
	GetByVerificationToken(token string) (*User, error)
	Update(user *User) error
	UpdatePassword(userID int64, hashedPassword string) error
	Delete(id int64) error
	List(limit, offset int) ([]*User, error)
	Count() (int64, error)

	// Account security methods
	IncrementFailedAttempts(userID int64) error
	ResetFailedAttempts(userID int64) error
	LockAccount(userID int64, lockDuration time.Duration) error
	UnlockAccount(userID int64) error
	IsAccountLocked(userID int64) (bool, *time.Time, error) // Returns locked status and unlock time
	DisableAccount(userID int64, disabledBy int64, reason string) error
	EnableAccount(userID int64) error
}

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByToken(token string) (*RefreshToken, error)
	GetByUserID(userID int64) ([]*RefreshToken, error)
	Revoke(tokenID int64) error
	RevokeAllForUser(userID int64) error
	DeleteExpired() error
	Delete(tokenID int64) error
}
