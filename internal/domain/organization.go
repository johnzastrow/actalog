// Package domain contains the core business entities
package domain

import (
	"time"
)

// Organization represents a gym or organization
type Organization struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
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
}
