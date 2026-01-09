package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestMovementRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	tests := []struct {
		name     string
		movement *domain.Movement
		wantErr  bool
	}{
		{
			name: "create standard movement",
			movement: &domain.Movement{
				Name:        "Back Squat",
				Description: "A compound lower body exercise",
				Type:        domain.MovementTypeWeightlifting,
				IsStandard:  true,
			},
			wantErr: false,
		},
		{
			name: "create custom movement",
			movement: &domain.Movement{
				Name:        "Custom Press",
				Description: "User-defined exercise",
				Type:        domain.MovementTypeWeightlifting,
				IsStandard:  false,
			},
			wantErr: false,
		},
		{
			name: "create bodyweight movement",
			movement: &domain.Movement{
				Name:        "Pull-ups",
				Description: "Upper body pulling movement",
				Type:        domain.MovementTypeBodyweight,
				IsStandard:  true,
			},
			wantErr: false,
		},
		{
			name: "create cardio movement",
			movement: &domain.Movement{
				Name:        "Running",
				Description: "Cardiovascular exercise",
				Type:        domain.MovementTypeCardio,
				IsStandard:  true,
			},
			wantErr: false,
		},
		{
			name: "duplicate name should fail",
			movement: &domain.Movement{
				Name:        "Back Squat", // Already exists
				Description: "Duplicate movement",
				Type:        domain.MovementTypeWeightlifting,
				IsStandard:  false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.movement)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.movement.ID == 0 {
				t.Error("Create() should set movement ID")
			}
		})
	}
}

func TestMovementRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a test movement
	movement := &domain.Movement{
		Name:        "Deadlift",
		Description: "Hip hinge movement",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  true,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create test movement: %v", err)
	}

	tests := []struct {
		name         string
		id           int64
		wantMovement bool
		wantName     string
	}{
		{
			name:         "existing movement",
			id:           movement.ID,
			wantMovement: true,
			wantName:     "Deadlift",
		},
		{
			name:         "non-existent movement",
			id:           99999,
			wantMovement: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() error = %v", err)
				return
			}
			if tt.wantMovement && got == nil {
				t.Error("GetByID() expected movement but got nil")
				return
			}
			if !tt.wantMovement && got != nil {
				t.Errorf("GetByID() expected nil but got movement: %+v", got)
				return
			}
			if tt.wantMovement && got.Name != tt.wantName {
				t.Errorf("GetByID() name = %v, want %v", got.Name, tt.wantName)
			}
		})
	}
}

func TestMovementRepository_GetByName(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create test movement
	movement := &domain.Movement{
		Name:        "Bench Press",
		Description: "Chest press movement",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  true,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create test movement: %v", err)
	}

	tests := []struct {
		name         string
		movementName string
		wantMovement bool
	}{
		{
			name:         "existing movement",
			movementName: "Bench Press",
			wantMovement: true,
		},
		{
			name:         "non-existent movement",
			movementName: "Non-Existent",
			wantMovement: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByName(tt.movementName)
			if err != nil {
				t.Errorf("GetByName() error = %v", err)
				return
			}
			if tt.wantMovement && got == nil {
				t.Error("GetByName() expected movement but got nil")
			}
			if !tt.wantMovement && got != nil {
				t.Errorf("GetByName() expected nil but got movement: %+v", got)
			}
		})
	}
}

func TestMovementRepository_ListStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a mix of standard and custom movements
	standardMovements := []string{"Standard Move 1", "Standard Move 2", "Standard Move 3"}
	customMovements := []string{"Custom Move 1", "Custom Move 2"}

	for _, name := range standardMovements {
		m := &domain.Movement{
			Name:       name,
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: true,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create standard movement: %v", err)
		}
	}

	for _, name := range customMovements {
		m := &domain.Movement{
			Name:       name,
			Type:       domain.MovementTypeBodyweight,
			IsStandard: false,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create custom movement: %v", err)
		}
	}

	// ListStandard should only return standard movements
	movements, err := repo.ListStandard()
	if err != nil {
		t.Fatalf("ListStandard() error = %v", err)
	}

	if len(movements) != len(standardMovements) {
		t.Errorf("ListStandard() returned %d movements, want %d", len(movements), len(standardMovements))
	}

	for _, m := range movements {
		if !m.IsStandard {
			t.Errorf("ListStandard() returned non-standard movement: %s", m.Name)
		}
	}
}

func TestMovementRepository_ListAll(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a mix of movements
	movements := []struct {
		name       string
		isStandard bool
	}{
		{"All Move 1", true},
		{"All Move 2", false},
		{"All Move 3", true},
		{"All Move 4", false},
	}

	for _, m := range movements {
		movement := &domain.Movement{
			Name:       m.name,
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: m.isStandard,
		}
		if err := repo.Create(movement); err != nil {
			t.Fatalf("failed to create movement: %v", err)
		}
	}

	// ListAll should return all movements
	result, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}

	if len(result) != len(movements) {
		t.Errorf("ListAll() returned %d movements, want %d", len(result), len(movements))
	}
}

func TestMovementRepository_ListByUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create users first
	userRepo := NewSQLiteUserRepository(db)
	user1 := createTestUser(t, userRepo, "user1@example.com")
	user2 := createTestUser(t, userRepo, "user2@example.com")

	repo := NewMovementRepository(db)

	// Create movements for different users
	for i := 0; i < 3; i++ {
		m := &domain.Movement{
			Name:       "User1 Movement " + string(rune('A'+i)),
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: false,
			CreatedBy:  &user1.ID,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create movement for user1: %v", err)
		}
	}

	for i := 0; i < 2; i++ {
		m := &domain.Movement{
			Name:       "User2 Movement " + string(rune('A'+i)),
			Type:       domain.MovementTypeBodyweight,
			IsStandard: false,
			CreatedBy:  &user2.ID,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create movement for user2: %v", err)
		}
	}

	// Test ListByUser for user1
	user1Movements, err := repo.ListByUser(user1.ID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(user1Movements) != 3 {
		t.Errorf("ListByUser(user1) returned %d movements, want 3", len(user1Movements))
	}

	// Test ListByUser for user2
	user2Movements, err := repo.ListByUser(user2.ID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(user2Movements) != 2 {
		t.Errorf("ListByUser(user2) returned %d movements, want 2", len(user2Movements))
	}
}

func TestMovementRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a custom movement (only custom movements can be updated)
	movement := &domain.Movement{
		Name:        "Custom Movement",
		Description: "Original description",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  false,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create movement: %v", err)
	}

	// Update the movement
	movement.Name = "Updated Movement"
	movement.Description = "Updated description"
	movement.Type = domain.MovementTypeBodyweight

	if err := repo.Update(movement); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(movement.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updated.Name != "Updated Movement" {
		t.Errorf("Update() name = %v, want 'Updated Movement'", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Update() description = %v, want 'Updated description'", updated.Description)
	}
	if updated.Type != domain.MovementTypeBodyweight {
		t.Errorf("Update() type = %v, want bodyweight", updated.Type)
	}
}

func TestMovementRepository_Update_StandardMovement(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a standard movement
	movement := &domain.Movement{
		Name:        "Standard Movement",
		Description: "Cannot be updated",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  true,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create movement: %v", err)
	}

	// Try to update the standard movement
	movement.Name = "Should Not Update"
	err = repo.Update(movement)
	if err == nil {
		t.Error("Update() should fail for standard movements")
	}
}

func TestMovementRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a custom movement
	movement := &domain.Movement{
		Name:        "To Delete",
		Description: "This will be deleted",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  false,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create movement: %v", err)
	}

	// Delete the movement
	if err := repo.Delete(movement.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's deleted
	deleted, err := repo.GetByID(movement.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if deleted != nil {
		t.Error("Delete() movement still exists after deletion")
	}
}

func TestMovementRepository_Delete_StandardMovement(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a standard movement
	movement := &domain.Movement{
		Name:        "Standard To Delete",
		Description: "Cannot be deleted",
		Type:        domain.MovementTypeWeightlifting,
		IsStandard:  true,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("failed to create movement: %v", err)
	}

	// Try to delete the standard movement
	err = repo.Delete(movement.ID)
	if err == nil {
		t.Error("Delete() should fail for standard movements")
	}
}

func TestMovementRepository_Search(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create test movements
	movements := []string{
		"Back Squat",
		"Front Squat",
		"Overhead Squat",
		"Squat Clean",
		"Deadlift",
		"Sumo Deadlift",
	}

	for _, name := range movements {
		m := &domain.Movement{
			Name:       name,
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: true,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create movement: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     string
		limit     int
		wantCount int
	}{
		{
			name:      "search for squat",
			query:     "Squat",
			limit:     10,
			wantCount: 4, // Back Squat, Front Squat, Overhead Squat, Squat Clean
		},
		{
			name:      "search for deadlift",
			query:     "Deadlift",
			limit:     10,
			wantCount: 2,
		},
		{
			name:      "search with limit",
			query:     "Squat",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "no results",
			query:     "NonExistent",
			limit:     10,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.Search(tt.query, tt.limit)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}
			if len(results) != tt.wantCount {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestMovementRepository_CopyToStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "creator@example.com")

	repo := NewMovementRepository(db)

	// Create a custom movement
	custom := &domain.Movement{
		Name:        "My Custom Exercise",
		Description: "A great custom exercise",
		Type:        domain.MovementTypeGymnastics,
		IsStandard:  false,
		CreatedBy:   &user.ID,
	}
	if err := repo.Create(custom); err != nil {
		t.Fatalf("failed to create custom movement: %v", err)
	}

	// Copy to standard
	standard, err := repo.CopyToStandard(custom.ID, "Official Custom Exercise")
	if err != nil {
		t.Fatalf("CopyToStandard() error = %v", err)
	}

	if standard.Name != "Official Custom Exercise" {
		t.Errorf("CopyToStandard() name = %v, want 'Official Custom Exercise'", standard.Name)
	}
	if !standard.IsStandard {
		t.Error("CopyToStandard() new movement should be standard")
	}
	if standard.CreatedBy != nil {
		t.Error("CopyToStandard() standard movement should have no creator")
	}
	if standard.Description != custom.Description {
		t.Errorf("CopyToStandard() description = %v, want %v", standard.Description, custom.Description)
	}
	if standard.Type != custom.Type {
		t.Errorf("CopyToStandard() type = %v, want %v", standard.Type, custom.Type)
	}
}

func TestMovementRepository_ListAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create a mix of movements
	standardCount := 3
	customCount := 4

	for i := 0; i < standardCount; i++ {
		m := &domain.Movement{
			Name:       "Standard " + string(rune('A'+i)),
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: true,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create standard movement: %v", err)
		}
	}

	for i := 0; i < customCount; i++ {
		m := &domain.Movement{
			Name:       "Custom " + string(rune('A'+i)),
			Type:       domain.MovementTypeBodyweight,
			IsStandard: false,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create custom movement: %v", err)
		}
	}

	// ListAllUserCreated should only return custom movements
	movements, err := repo.ListAllUserCreated()
	if err != nil {
		t.Fatalf("ListAllUserCreated() error = %v", err)
	}

	if len(movements) != customCount {
		t.Errorf("ListAllUserCreated() returned %d movements, want %d", len(movements), customCount)
	}

	for _, m := range movements {
		if m.IsStandard {
			t.Errorf("ListAllUserCreated() returned standard movement: %s", m.Name)
		}
	}
}

func TestMovementRepository_CountAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewMovementRepository(db)

	// Create movements
	for i := 0; i < 2; i++ {
		m := &domain.Movement{
			Name:       "Standard Count " + string(rune('A'+i)),
			Type:       domain.MovementTypeWeightlifting,
			IsStandard: true,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create standard movement: %v", err)
		}
	}

	for i := 0; i < 5; i++ {
		m := &domain.Movement{
			Name:       "Custom Count " + string(rune('A'+i)),
			Type:       domain.MovementTypeBodyweight,
			IsStandard: false,
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("failed to create custom movement: %v", err)
		}
	}

	count, err := repo.CountAllUserCreated()
	if err != nil {
		t.Fatalf("CountAllUserCreated() error = %v", err)
	}

	if count != 5 {
		t.Errorf("CountAllUserCreated() = %d, want 5", count)
	}
}

// Helper function to create a test user
func createTestUser(t *testing.T, repo *SQLiteUserRepository, email string) *domain.User {
	t.Helper()
	user := &domain.User{
		Email:        email,
		PasswordHash: "hash123",
		Name:         "Test User",
		Role:         "user",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}
