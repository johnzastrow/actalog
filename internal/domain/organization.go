// Package domain contains the core business entities
package domain

import (
	"time"
)

// Organization represents a gym or organization
type Organization struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	MemberCount int       `json:"member_count,omitempty" db:"-"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserOrganization represents the many-to-many relationship between users and organizations
type UserOrganization struct {
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	OrganizationID int64     `json:"organization_id" db:"organization_id"`
	Role           string    `json:"role" db:"role"` // member, admin, owner
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationRepository defines the interface for organization data access
type OrganizationRepository interface {
	Create(org *Organization) error
	GetByID(id int64) (*Organization, error)
	GetByName(name string) (*Organization, error)
	List(limit, offset int) ([]*Organization, int64, error)
	Update(org *Organization) error
	Delete(id int64) error
	Count() (int64, error)

	// Many-to-many relationship methods
	AddUserToOrganization(userID, orgID int64, role string) error
	RemoveUserFromOrganization(userID, orgID int64) error
	GetUserOrganizations(userID int64) ([]*Organization, error)
	GetOrganizationUsers(orgID int64) ([]*User, error)
	IsUserInOrganization(userID, orgID int64) (bool, error)
	GetUserOrganizationIDs(userID int64) ([]int64, error)
}
