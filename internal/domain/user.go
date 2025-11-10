// Package domain contains the core business entities
package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                   int64      `json:"id" db:"id"`
	Email                string     `json:"email" db:"email"`
	PasswordHash         string     `json:"-" db:"password_hash"` // Never serialize password
	Name                 string     `json:"name" db:"name"`
	ProfileImage         string     `json:"profile_image,omitempty" db:"profile_image"`
	Role                 string     `json:"role" db:"role"` // user, admin
	ResetToken           *string    `json:"-" db:"reset_token"` // Never serialize reset token
	ResetTokenExpiresAt  *time.Time `json:"-" db:"reset_token_expires_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id int64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByResetToken(token string) (*User, error)
	Update(user *User) error
	Delete(id int64) error
	List(limit, offset int) ([]*User, error)
	Count() (int64, error)
}
