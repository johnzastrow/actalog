package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestOrganizationRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	desc := "A test organization"
	org := &domain.Organization{
		Name:        "Test Org",
		Description: &desc,
	}

	err = repo.Create(org)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if org.ID == 0 {
		t.Error("Create() should set ID")
	}
	if org.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestOrganizationRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Create org
	desc := "Test description"
	org := &domain.Organization{
		Name:        "Test Org",
		Description: &desc,
	}
	if err := repo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Get by ID
	got, err := repo.GetByID(org.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.Name != "Test Org" {
		t.Errorf("Name = %v, want 'Test Org'", got.Name)
	}
	if got.Description == nil || *got.Description != desc {
		t.Error("Description should match")
	}

	// Get non-existent
	got, err = repo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestOrganizationRepository_GetByName(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Create org
	org := &domain.Organization{
		Name: "Unique Org Name",
	}
	if err := repo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Get by name
	got, err := repo.GetByName("Unique Org Name")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByName() returned nil")
	}
	if got.ID != org.ID {
		t.Errorf("ID = %d, want %d", got.ID, org.ID)
	}

	// Get non-existent
	got, err = repo.GetByName("Non-existent")
	if err != nil {
		t.Fatalf("GetByName(non-existent) error = %v", err)
	}
	if got != nil {
		t.Error("GetByName(non-existent) should return nil")
	}
}

func TestOrganizationRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Create multiple orgs
	orgs := []string{"Alpha Org", "Beta Org", "Gamma Org", "Delta Org", "Epsilon Org"}
	for _, name := range orgs {
		org := &domain.Organization{Name: name}
		if err := repo.Create(org); err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}
	}

	// List all
	got, count, err := repo.List(100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 5 {
		t.Errorf("List() returned %d orgs, want 5", len(got))
	}
	if count != 5 {
		t.Errorf("Count = %d, want 5", count)
	}

	// Verify ordering by name (alphabetical)
	if got[0].Name != "Alpha Org" {
		t.Errorf("First org should be 'Alpha Org', got %v", got[0].Name)
	}

	// Test pagination
	got, _, err = repo.List(2, 0)
	if err != nil {
		t.Fatalf("List(2, 0) error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("List(2, 0) returned %d orgs, want 2", len(got))
	}

	got, _, err = repo.List(2, 2)
	if err != nil {
		t.Fatalf("List(2, 2) error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("List(2, 2) returned %d orgs, want 2", len(got))
	}
}

func TestOrganizationRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Create org
	org := &domain.Organization{Name: "Original Name"}
	if err := repo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Update
	newDesc := "Updated description"
	org.Name = "Updated Name"
	org.Description = &newDesc
	err = repo.Update(org)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, err := repo.GetByID(org.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Updated Name" {
		t.Errorf("Name = %v, want 'Updated Name'", got.Name)
	}
	if got.Description == nil || *got.Description != newDesc {
		t.Error("Description should be updated")
	}
}

func TestOrganizationRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Create org
	org := &domain.Organization{Name: "To Delete"}
	if err := repo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Delete
	err = repo.Delete(org.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, err := repo.GetByID(org.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Error("Org should be deleted")
	}
}

func TestOrganizationRepository_DeleteWithUsers(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create org
	org := &domain.Organization{Name: "Org with Users"}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Add user to org
	if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
		t.Fatalf("Failed to add user to org: %v", err)
	}

	// Try to delete (should fail due to RESTRICT)
	err = orgRepo.Delete(org.ID)
	if err == nil {
		t.Error("Delete() should fail when users are assigned")
	}
}

func TestOrganizationRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewOrganizationRepository(db)

	// Initially 0
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	// Create orgs
	for i := 0; i < 3; i++ {
		org := &domain.Organization{Name: "Org " + string(rune('A'+i))}
		if err := repo.Create(org); err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}
	}

	count, err = repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestOrganizationRepository_AddUserToOrganization(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create org
	org := &domain.Organization{Name: "Test Org"}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Add user to org
	err = orgRepo.AddUserToOrganization(user.ID, org.ID, "member")
	if err != nil {
		t.Fatalf("AddUserToOrganization() error = %v", err)
	}

	// Verify user is in org
	inOrg, err := orgRepo.IsUserInOrganization(user.ID, org.ID)
	if err != nil {
		t.Fatalf("IsUserInOrganization() error = %v", err)
	}
	if !inOrg {
		t.Error("User should be in organization")
	}
}

func TestOrganizationRepository_RemoveUserFromOrganization(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user and org
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	org := &domain.Organization{Name: "Test Org"}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Add user
	if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Remove user
	err = orgRepo.RemoveUserFromOrganization(user.ID, org.ID)
	if err != nil {
		t.Fatalf("RemoveUserFromOrganization() error = %v", err)
	}

	// Verify removed
	inOrg, err := orgRepo.IsUserInOrganization(user.ID, org.ID)
	if err != nil {
		t.Fatalf("IsUserInOrganization() error = %v", err)
	}
	if inOrg {
		t.Error("User should not be in organization")
	}
}

func TestOrganizationRepository_GetUserOrganizations(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create orgs
	orgNames := []string{"Org A", "Org B", "Org C"}
	for _, name := range orgNames {
		org := &domain.Organization{Name: name}
		if err := orgRepo.Create(org); err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}
		if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
			t.Fatalf("Failed to add user to org: %v", err)
		}
	}

	// Get user's orgs
	orgs, err := orgRepo.GetUserOrganizations(user.ID)
	if err != nil {
		t.Fatalf("GetUserOrganizations() error = %v", err)
	}
	if len(orgs) != 3 {
		t.Errorf("GetUserOrganizations() returned %d orgs, want 3", len(orgs))
	}
}

func TestOrganizationRepository_GetOrganizationUsers(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create org
	org := &domain.Organization{Name: "Test Org"}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Create users and add to org
	for i := 0; i < 3; i++ {
		user := &domain.User{
			Email:        "user" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hash",
			Name:         "User " + string(rune('0'+i)),
			Role:         "user",
		}
		if err := userRepo.Create(user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
			t.Fatalf("Failed to add user to org: %v", err)
		}
	}

	// Get org's users
	users, err := orgRepo.GetOrganizationUsers(org.ID)
	if err != nil {
		t.Fatalf("GetOrganizationUsers() error = %v", err)
	}
	if len(users) != 3 {
		t.Errorf("GetOrganizationUsers() returned %d users, want 3", len(users))
	}
}

func TestOrganizationRepository_IsUserInOrganization(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user and org
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	org := &domain.Organization{Name: "Test Org"}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Not in org initially
	inOrg, err := orgRepo.IsUserInOrganization(user.ID, org.ID)
	if err != nil {
		t.Fatalf("IsUserInOrganization() error = %v", err)
	}
	if inOrg {
		t.Error("User should not be in org initially")
	}

	// Add user
	if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Now in org
	inOrg, err = orgRepo.IsUserInOrganization(user.ID, org.ID)
	if err != nil {
		t.Fatalf("IsUserInOrganization() error = %v", err)
	}
	if !inOrg {
		t.Error("User should be in org")
	}
}

func TestOrganizationRepository_GetUserOrganizationIDs(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	// Create user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create orgs and add user
	var expectedIDs []int64
	for i := 0; i < 3; i++ {
		org := &domain.Organization{Name: "Org " + string(rune('A'+i))}
		if err := orgRepo.Create(org); err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}
		if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
			t.Fatalf("Failed to add user to org: %v", err)
		}
		expectedIDs = append(expectedIDs, org.ID)
	}

	// Get org IDs
	ids, err := orgRepo.GetUserOrganizationIDs(user.ID)
	if err != nil {
		t.Fatalf("GetUserOrganizationIDs() error = %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("GetUserOrganizationIDs() returned %d IDs, want 3", len(ids))
	}

	// User not in any org
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "user"}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	ids, err = orgRepo.GetUserOrganizationIDs(user2.ID)
	if err != nil {
		t.Fatalf("GetUserOrganizationIDs() error = %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("GetUserOrganizationIDs() returned %d IDs, want 0", len(ids))
	}
}
