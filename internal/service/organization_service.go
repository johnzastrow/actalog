package service

import (
	"errors"
	"fmt"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrOrganizationNameExists = errors.New("organization name already exists")
	ErrOrganizationHasUsers   = errors.New("organization has users, cannot delete")
)

type OrganizationService struct {
	orgRepo  domain.OrganizationRepository
	userRepo domain.UserRepository
}

func NewOrganizationService(
	orgRepo domain.OrganizationRepository,
	userRepo domain.UserRepository,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

// Create creates a new organization
func (s *OrganizationService) Create(name string, description *string) (*domain.Organization, error) {
	// Validate input
	if name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Check if name already exists
	existing, err := s.orgRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing organization: %w", err)
	}
	if existing != nil {
		return nil, ErrOrganizationNameExists
	}

	org := &domain.Organization{
		Name:        name,
		Description: description,
	}

	if err := s.orgRepo.Create(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

// GetByID retrieves an organization by ID
func (s *OrganizationService) GetByID(id int64) (*domain.Organization, error) {
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	return org, nil
}

// List retrieves organizations with pagination
func (s *OrganizationService) List(limit, offset int) ([]*domain.Organization, int64, error) {
	// Validate pagination
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	orgs, count, err := s.orgRepo.List(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, count, nil
}

// Update updates an organization
func (s *OrganizationService) Update(id int64, name string, description *string) (*domain.Organization, error) {
	// Get existing organization
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	// Check if name is changing and if new name already exists
	if name != "" && name != org.Name {
		existing, err := s.orgRepo.GetByName(name)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing organization: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, ErrOrganizationNameExists
		}
		org.Name = name
	}

	org.Description = description

	if err := s.orgRepo.Update(org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

// Delete deletes an organization (only if no users are assigned)
func (s *OrganizationService) Delete(id int64) error {
	// Check if organization exists
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// TODO: Check if any users are assigned to this organization
	// This would require adding a method to UserRepository to count users by organization
	// For now, we'll allow deletion and let the database foreign key handle it (SET NULL)

	if err := s.orgRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// AssignUserToOrganization assigns a user to an organization
func (s *OrganizationService) AssignUserToOrganization(userID, orgID int64) error {
	// Verify organization exists
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update user's organization
	user.OrganizationID = &orgID

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// RemoveUserFromOrganization removes a user from their organization
func (s *OrganizationService) RemoveUserFromOrganization(userID int64) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Clear organization
	user.OrganizationID = nil

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
