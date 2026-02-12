package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

// =============================================================================
// Organization Service Tests
// =============================================================================

func TestOrganizationService_Create(t *testing.T) {
	tests := []struct {
		name        string
		orgName     string
		description *string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid organization",
			orgName: "Test Gym",
			wantErr: false,
		},
		{
			name:        "with description",
			orgName:     "CrossFit Box",
			description: stringPtr("A local CrossFit gym"),
			wantErr:     false,
		},
		{
			name:        "empty name",
			orgName:     "",
			wantErr:     true,
			errContains: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgRepo := newMockOrganizationRepo()
			userRepo := newMockUserRepo()
			auditRepo := newMockAuditLogRepo()
			service := NewOrganizationService(orgRepo, userRepo, auditRepo)

			org, err := service.Create(1, tt.orgName, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Create() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			if org.Name != tt.orgName {
				t.Errorf("Create() name = %v, want %v", org.Name, tt.orgName)
			}
			if (tt.description == nil && org.Description != nil) ||
				(tt.description != nil && org.Description == nil) ||
				(tt.description != nil && org.Description != nil && *tt.description != *org.Description) {
				t.Errorf("Create() description mismatch")
			}
		})
	}
}

func TestOrganizationService_Create_DuplicateName(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create first organization
	_, err := service.Create(1, "Test Gym", nil)
	if err != nil {
		t.Fatalf("Create() first org error: %v", err)
	}

	// Try to create duplicate
	_, err = service.Create(1, "Test Gym", nil)
	if err == nil {
		t.Fatal("Create() expected error for duplicate name")
	}
	if !errors.Is(err, ErrOrganizationNameExists) {
		t.Errorf("Create() error = %v, want ErrOrganizationNameExists", err)
	}
}

func TestOrganizationService_GetByID(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create an organization
	created, _ := service.Create(1, "Test Gym", nil)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType error
	}{
		{
			name:    "existing organization",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent organization",
			id:      999,
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := service.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("GetByID() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByID() unexpected error: %v", err)
			}
			if org.ID != tt.id {
				t.Errorf("GetByID() ID = %v, want %v", org.ID, tt.id)
			}
		})
	}
}

func TestOrganizationService_List(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create some organizations
	service.Create(1, "Gym A", nil)
	service.Create(1, "Gym B", nil)
	service.Create(1, "Gym C", nil)

	// Test basic listing
	orgs, total, err := service.List(10, 0)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(orgs) != 3 {
		t.Errorf("List() count = %v, want 3", len(orgs))
	}
	if total != 3 {
		t.Errorf("List() total = %v, want 3", total)
	}

	// Test with invalid limit (defaults to 50)
	orgs, total, err = service.List(0, 0)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(orgs) != 3 {
		t.Errorf("List() with invalid limit count = %v, want 3", len(orgs))
	}

	// Test with negative offset (defaults to 0)
	orgs, total, err = service.List(10, -5)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(orgs) != 3 {
		t.Errorf("List() with negative offset count = %v, want 3", len(orgs))
	}
}

func TestOrganizationService_Update(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization
	created, _ := service.Create(1, "Old Name", nil)

	tests := []struct {
		name        string
		id          int64
		newName     string
		description *string
		wantErr     bool
		errType     error
	}{
		{
			name:        "update name",
			id:          created.ID,
			newName:     "New Name",
			description: nil,
			wantErr:     false,
		},
		{
			name:    "non-existent organization",
			id:      999,
			newName: "Test",
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := service.Update(1, tt.id, tt.newName, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Update() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
			if org.Name != tt.newName {
				t.Errorf("Update() name = %v, want %v", org.Name, tt.newName)
			}
		})
	}
}

func TestOrganizationService_Update_DuplicateName(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create two organizations
	org1, _ := service.Create(1, "Gym A", nil)
	service.Create(1, "Gym B", nil)

	// Try to rename org1 to Gym B (duplicate)
	_, err := service.Update(1, org1.ID, "Gym B", nil)
	if err == nil {
		t.Fatal("Update() expected error for duplicate name")
	}
	if !errors.Is(err, ErrOrganizationNameExists) {
		t.Errorf("Update() error = %v, want ErrOrganizationNameExists", err)
	}
}

func TestOrganizationService_Delete(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization
	org, _ := service.Create(1, "Test Gym", nil)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType error
	}{
		{
			name:    "delete existing organization",
			id:      org.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent organization",
			id:      999,
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(1, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Delete() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("Delete() unexpected error: %v", err)
			}

			// Verify deleted
			_, getErr := service.GetByID(tt.id)
			if !errors.Is(getErr, ErrOrganizationNotFound) {
				t.Errorf("GetByID() after delete should return ErrOrganizationNotFound")
			}
		})
	}
}

func TestOrganizationService_Delete_WithUsers(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization
	org, _ := service.Create(1, "Test Gym", nil)

	// Create and assign user
	user := &domain.User{
		Email:        "user@test.com",
		Name:         "Test User",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
	}
	userRepo.Create(user)
	orgRepo.AddUserToOrganization(user.ID, org.ID, "member")

	// Try to delete organization with users
	err := service.Delete(1, org.ID)
	if err == nil {
		t.Fatal("Delete() expected error when organization has users")
	}
	if !errors.Is(err, ErrOrganizationHasUsers) {
		t.Errorf("Delete() error = %v, want ErrOrganizationHasUsers", err)
	}
}

func TestOrganizationService_AssignUserToOrganization(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization and user
	org, _ := service.Create(1, "Test Gym", nil)
	user := &domain.User{
		Email:        "user@test.com",
		Name:         "Test User",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
	}
	userRepo.Create(user)

	tests := []struct {
		name        string
		userID      int64
		orgID       int64
		wantErr     bool
		errType     error
		errContains string
	}{
		{
			name:    "assign user to organization",
			userID:  user.ID,
			orgID:   org.ID,
			wantErr: false,
		},
		{
			name:    "assign to non-existent organization",
			userID:  user.ID,
			orgID:   999,
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
		{
			name:    "assign non-existent user",
			userID:  999,
			orgID:   org.ID,
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AssignUserToOrganization(1, tt.userID, tt.orgID)

			if tt.wantErr {
				if err == nil {
					t.Error("AssignUserToOrganization() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("AssignUserToOrganization() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("AssignUserToOrganization() unexpected error: %v", err)
			}

			// Verify membership
			isMember, _ := orgRepo.IsUserInOrganization(tt.userID, tt.orgID)
			if !isMember {
				t.Error("User should be a member after assignment")
			}
		})
	}
}

func TestOrganizationService_AssignUserToOrganization_AlreadyMember(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization and user
	org, _ := service.Create(1, "Test Gym", nil)
	user := &domain.User{
		Email:        "user@test.com",
		Name:         "Test User",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
	}
	userRepo.Create(user)

	// First assignment
	err := service.AssignUserToOrganization(1, user.ID, org.ID)
	if err != nil {
		t.Fatalf("First assignment error: %v", err)
	}

	// Try to assign again
	err = service.AssignUserToOrganization(1, user.ID, org.ID)
	if err == nil {
		t.Fatal("AssignUserToOrganization() expected error for duplicate assignment")
	}
	if !strings.Contains(err.Error(), "already a member") {
		t.Errorf("Error should mention already a member, got: %v", err)
	}
}

func TestOrganizationService_RemoveUserFromOrganization(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization and user
	org, _ := service.Create(1, "Test Gym", nil)
	user := &domain.User{
		Email:        "user@test.com",
		Name:         "Test User",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
	}
	userRepo.Create(user)
	service.AssignUserToOrganization(1, user.ID, org.ID)

	tests := []struct {
		name    string
		userID  int64
		orgID   int64
		wantErr bool
		errType error
	}{
		{
			name:    "remove user from organization",
			userID:  user.ID,
			orgID:   org.ID,
			wantErr: false,
		},
		{
			name:    "remove from non-existent organization",
			userID:  user.ID,
			orgID:   999,
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
		{
			name:    "remove non-existent user",
			userID:  999,
			orgID:   org.ID,
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RemoveUserFromOrganization(1, tt.userID, tt.orgID)

			if tt.wantErr {
				if err == nil {
					t.Error("RemoveUserFromOrganization() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("RemoveUserFromOrganization() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("RemoveUserFromOrganization() unexpected error: %v", err)
			}

			// Verify removal
			isMember, _ := orgRepo.IsUserInOrganization(tt.userID, tt.orgID)
			if isMember {
				t.Error("User should not be a member after removal")
			}
		})
	}
}

func TestOrganizationService_GetUserOrganizations(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create user and organizations
	user := &domain.User{
		Email:        "user@test.com",
		Name:         "Test User",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
	}
	userRepo.Create(user)
	org1, _ := service.Create(1, "Gym A", nil)
	org2, _ := service.Create(1, "Gym B", nil)
	service.Create(1, "Gym C", nil) // Not assigned to user

	// Assign user to some organizations
	service.AssignUserToOrganization(1, user.ID, org1.ID)
	service.AssignUserToOrganization(1, user.ID, org2.ID)

	tests := []struct {
		name      string
		userID    int64
		wantCount int
		wantErr   bool
		errType   error
	}{
		{
			name:      "get user organizations",
			userID:    user.ID,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "non-existent user",
			userID:  999,
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgs, err := service.GetUserOrganizations(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetUserOrganizations() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("GetUserOrganizations() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetUserOrganizations() unexpected error: %v", err)
			}
			if len(orgs) != tt.wantCount {
				t.Errorf("GetUserOrganizations() count = %v, want %v", len(orgs), tt.wantCount)
			}
		})
	}
}

func TestOrganizationService_GetOrganizationUsers(t *testing.T) {
	orgRepo := newMockOrganizationRepo()
	userRepo := newMockUserRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewOrganizationService(orgRepo, userRepo, auditRepo)

	// Create organization
	org, _ := service.Create(1, "Test Gym", nil)

	tests := []struct {
		name    string
		orgID   int64
		wantErr bool
		errType error
	}{
		{
			name:    "get organization users",
			orgID:   org.ID,
			wantErr: false,
		},
		{
			name:    "non-existent organization",
			orgID:   999,
			wantErr: true,
			errType: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetOrganizationUsers(tt.orgID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetOrganizationUsers() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("GetOrganizationUsers() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetOrganizationUsers() unexpected error: %v", err)
			}
		})
	}
}
