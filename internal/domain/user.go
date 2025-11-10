// Package domain contains the core business entities
package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
<<<<<<< HEAD
	ID                          int64      `json:"id" db:"id"`
	Email                       string     `json:"email" db:"email"`
	PasswordHash                string     `json:"-" db:"password_hash"` // Never serialize password
	Name                        string     `json:"name" db:"name"`
	ProfileImage                *string    `json:"profile_image,omitempty" db:"profile_image"`
	Role                        string     `json:"role" db:"role"` // user, admin
	EmailVerified               bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt             *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	VerificationToken           *string    `json:"-" db:"verification_token"` // Never serialize verification token
	VerificationTokenExpiresAt  *time.Time `json:"-" db:"verification_token_expires_at"`
	ResetToken                  *string    `json:"-" db:"reset_token"` // Never serialize reset token
	ResetTokenExpiresAt         *time.Time `json:"-" db:"reset_token_expires_at"`
	CreatedAt                   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt                 *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
=======
	ID                   int64      `json:"id" db:"id"`
	Email                string     `json:"email" db:"email"`
	PasswordHash         string     `json:"-" db:"password_hash"` // Never serialize password
	Name                 string     `json:"name" db:"name"`
	ProfileImage         *string    `json:"profile_image,omitempty" db:"profile_image"`
	Role                 string     `json:"role" db:"role"` // user, admin
	ResetToken           *string    `json:"-" db:"reset_token"` // Never serialize reset token
	ResetTokenExpiresAt  *time.Time `json:"-" db:"reset_token_expires_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
>>>>>>> 6401243 (feat: Add Remember Me functionality (v0.3.2-beta))
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
	Delete(id int64) error
	List(limit, offset int) ([]*User, error)
	Count() (int64, error)
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
