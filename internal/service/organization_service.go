package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrOrganizationNameExists = errors.New("organization name already exists")
	ErrOrganizationHasUsers   = errors.New("organization has users, cannot delete")
)

type OrganizationService struct {
	orgRepo      domain.OrganizationRepository
	userRepo     domain.UserRepository
	auditLogRepo domain.AuditLogRepository
}

func NewOrganizationService(
	orgRepo domain.OrganizationRepository,
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:      orgRepo,
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

// Create creates a new organization
func (s *OrganizationService) Create(adminUserID int64, name string, description *string) (*domain.Organization, error) {
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

	// Audit log
	details, _ := json.Marshal(map[string]interface{}{
		"organization_id":   org.ID,
		"organization_name": org.Name,
		"description":       description,
	})
	detailsStr := string(details)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrganizationCreated,
		Details:   &detailsStr,
		CreatedAt: time.Now(),
	})

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
func (s *OrganizationService) Update(adminUserID int64, id int64, name string, description *string) (*domain.Organization, error) {
	// Get existing organization
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	// Track changes for audit log
	changes := make(map[string]interface{})
	oldName := org.Name
	oldDesc := org.Description

	// Check if name is changing and if new name already exists
	if name != "" && name != org.Name {
		existing, err := s.orgRepo.GetByName(name)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing organization: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, ErrOrganizationNameExists
		}
		changes["name_old"] = oldName
		changes["name_new"] = name
		org.Name = name
	}

	if (description == nil && oldDesc != nil) || (description != nil && oldDesc == nil) || (description != nil && oldDesc != nil && *description != *oldDesc) {
		changes["description_old"] = oldDesc
		changes["description_new"] = description
	}
	org.Description = description

	if err := s.orgRepo.Update(org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	// Audit log
	if len(changes) > 0 {
		changes["organization_id"] = org.ID
		changes["organization_name"] = org.Name
		details, _ := json.Marshal(changes)
		detailsStr := string(details)
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:    &adminUserID,
			EventType: domain.EventOrganizationUpdated,
			Details:   &detailsStr,
			CreatedAt: time.Now(),
		})
	}

	return org, nil
}

// Delete deletes an organization (only if no users are assigned)
func (s *OrganizationService) Delete(adminUserID int64, id int64) error {
	// Check if organization exists
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Save org details for audit log before deletion
	orgName := org.Name
	orgDesc := org.Description

	// Repository Delete method now checks for users and returns error if any exist (RESTRICT behavior)
	if err := s.orgRepo.Delete(id); err != nil {
		// Check if error is due to users being assigned
		if err.Error() == "cannot delete organization: users are still assigned" {
			return ErrOrganizationHasUsers
		}
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Audit log
	details, _ := json.Marshal(map[string]interface{}{
		"organization_id":   id,
		"organization_name": orgName,
		"description":       orgDesc,
	})
	detailsStr := string(details)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrganizationDeleted,
		Details:   &detailsStr,
		CreatedAt: time.Now(),
	})

	return nil
}

// AssignUserToOrganization adds a user to an organization (many-to-many)
func (s *OrganizationService) AssignUserToOrganization(adminUserID, userID, orgID int64) error {
	// Verify organization exists
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check if user is already a member
	isMember, err := s.orgRepo.IsUserInOrganization(userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user is already a member of this organization")
	}

	// Add user to organization with 'member' role
	if err := s.orgRepo.AddUserToOrganization(userID, orgID, "member"); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Audit log
	details, _ := json.Marshal(map[string]interface{}{
		"user_id":           userID,
		"user_email":        user.Email,
		"organization_id":   orgID,
		"organization_name": org.Name,
		"role":              "member",
	})
	detailsStr := string(details)
	targetUserID := userID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventUserAssignedToOrg,
		Details:      &detailsStr,
		CreatedAt:    time.Now(),
	})

	return nil
}

// RemoveUserFromOrganization removes a user from a specific organization
func (s *OrganizationService) RemoveUserFromOrganization(adminUserID, userID, orgID int64) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify organization exists
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Remove user from organization
	if err := s.orgRepo.RemoveUserFromOrganization(userID, orgID); err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}

	// Audit log
	details, _ := json.Marshal(map[string]interface{}{
		"user_id":           userID,
		"user_email":        user.Email,
		"organization_id":   orgID,
		"organization_name": org.Name,
	})
	detailsStr := string(details)
	targetUserID := userID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventUserRemovedFromOrg,
		Details:      &detailsStr,
		CreatedAt:    time.Now(),
	})

	return nil
}

// GetUserOrganizations returns all organizations for a user
func (s *OrganizationService) GetUserOrganizations(userID int64) ([]*domain.Organization, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	orgs, err := s.orgRepo.GetUserOrganizations(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	return orgs, nil
}

// GetOrganizationUsers returns all users in an organization
func (s *OrganizationService) GetOrganizationUsers(orgID int64) ([]*domain.User, error) {
	// Verify organization exists
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	users, err := s.orgRepo.GetOrganizationUsers(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization users: %w", err)
	}

	return users, nil
}
