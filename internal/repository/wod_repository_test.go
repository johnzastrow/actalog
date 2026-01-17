package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestWODRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	tests := []struct {
		name    string
		wod     *domain.WOD
		wantErr bool
	}{
		{
			name: "create standard WOD",
			wod: &domain.WOD{
				Name:        "Fran",
				Source:      "CrossFit",
				Type:        "Girl",
				Regime:      "For Time",
				ScoreType:   "Time",
				Description: "21-15-9 Thrusters (95/65) and Pull-ups",
				IsStandard:  true,
			},
			wantErr: false,
		},
		{
			name: "create custom WOD",
			wod: &domain.WOD{
				Name:        "My Custom WOD",
				Source:      "Custom",
				Type:        "Custom",
				Regime:      "AMRAP",
				ScoreType:   "Rounds+Reps",
				Description: "10 min AMRAP of burpees and double unders",
				IsStandard:  false,
			},
			wantErr: false,
		},
		{
			name: "create hero WOD",
			wod: &domain.WOD{
				Name:        "Murph",
				Source:      "CrossFit",
				Type:        "Hero",
				Regime:      "For Time",
				ScoreType:   "Time",
				Description: "1 mile run, 100 pull-ups, 200 push-ups, 300 squats, 1 mile run",
				IsStandard:  true,
			},
			wantErr: false,
		},
		{
			name: "duplicate name should fail",
			wod: &domain.WOD{
				Name:        "Fran", // Already exists
				Source:      "Custom",
				Type:        "Custom",
				Regime:      "For Time",
				ScoreType:   "Time",
				Description: "Duplicate WOD",
				IsStandard:  false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.wod)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wod.ID == 0 {
				t.Error("Create() should set WOD ID")
			}
		})
	}
}

func TestWODRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a test WOD
	wod := &domain.WOD{
		Name:        "Helen",
		Source:      "CrossFit",
		Type:        "Girl",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "3 rounds: 400m run, 21 KB swings, 12 pull-ups",
		IsStandard:  true,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create test WOD: %v", err)
	}

	tests := []struct {
		name     string
		id       int64
		wantWOD  bool
		wantName string
	}{
		{
			name:     "existing WOD",
			id:       wod.ID,
			wantWOD:  true,
			wantName: "Helen",
		},
		{
			name:    "non-existent WOD",
			id:      99999,
			wantWOD: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() error = %v", err)
				return
			}
			if tt.wantWOD && got == nil {
				t.Error("GetByID() expected WOD but got nil")
				return
			}
			if !tt.wantWOD && got != nil {
				t.Errorf("GetByID() expected nil but got WOD: %+v", got)
				return
			}
			if tt.wantWOD && got.Name != tt.wantName {
				t.Errorf("GetByID() name = %v, want %v", got.Name, tt.wantName)
			}
		})
	}
}

func TestWODRepository_GetByName(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create test WOD
	wod := &domain.WOD{
		Name:        "Grace",
		Source:      "CrossFit",
		Type:        "Girl",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "30 Clean and Jerks (135/95)",
		IsStandard:  true,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create test WOD: %v", err)
	}

	tests := []struct {
		name    string
		wodName string
		wantWOD bool
	}{
		{
			name:    "existing WOD",
			wodName: "Grace",
			wantWOD: true,
		},
		{
			name:    "non-existent WOD",
			wodName: "NonExistent",
			wantWOD: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByName(tt.wodName)
			if err != nil {
				t.Errorf("GetByName() error = %v", err)
				return
			}
			if tt.wantWOD && got == nil {
				t.Error("GetByName() expected WOD but got nil")
			}
			if !tt.wantWOD && got != nil {
				t.Errorf("GetByName() expected nil but got WOD: %+v", got)
			}
		})
	}
}

func TestWODRepository_ListStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a mix of standard and custom WODs
	standardWODs := []string{"Standard WOD 1", "Standard WOD 2", "Standard WOD 3"}
	customWODs := []string{"Custom WOD 1", "Custom WOD 2"}

	for _, name := range standardWODs {
		wod := &domain.WOD{
			Name:        name,
			Source:      "CrossFit",
			Type:        "Benchmark",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Test WOD",
			IsStandard:  true,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create standard WOD: %v", err)
		}
	}

	for _, name := range customWODs {
		wod := &domain.WOD{
			Name:        name,
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "Test WOD",
			IsStandard:  false,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create custom WOD: %v", err)
		}
	}

	// ListStandard should only return standard WODs
	wods, err := repo.ListStandard(0, 0)
	if err != nil {
		t.Fatalf("ListStandard() error = %v", err)
	}

	if len(wods) != len(standardWODs) {
		t.Errorf("ListStandard() returned %d WODs, want %d", len(wods), len(standardWODs))
	}

	for _, w := range wods {
		if !w.IsStandard {
			t.Errorf("ListStandard() returned non-standard WOD: %s", w.Name)
		}
	}
}

func TestWODRepository_ListByUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create users first
	userRepo := NewSQLiteUserRepository(db)
	user1 := createTestUser(t, userRepo, "woduser1@example.com")
	user2 := createTestUser(t, userRepo, "woduser2@example.com")

	repo := NewWODRepository(db)

	// Create WODs for different users
	for i := 0; i < 3; i++ {
		wod := &domain.WOD{
			Name:        "User1 WOD " + string(rune('A'+i)),
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "Custom WOD",
			IsStandard:  false,
			CreatedBy:   &user1.ID,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create WOD for user1: %v", err)
		}
	}

	for i := 0; i < 2; i++ {
		wod := &domain.WOD{
			Name:        "User2 WOD " + string(rune('A'+i)),
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Custom WOD",
			IsStandard:  false,
			CreatedBy:   &user2.ID,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create WOD for user2: %v", err)
		}
	}

	// Test ListByUser for user1
	user1WODs, err := repo.ListByUser(user1.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(user1WODs) != 3 {
		t.Errorf("ListByUser(user1) returned %d WODs, want 3", len(user1WODs))
	}

	// Test ListByUser for user2
	user2WODs, err := repo.ListByUser(user2.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(user2WODs) != 2 {
		t.Errorf("ListByUser(user2) returned %d WODs, want 2", len(user2WODs))
	}
}

func TestWODRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a custom WOD (only custom WODs can be updated)
	wod := &domain.WOD{
		Name:        "Custom WOD",
		Source:      "Custom",
		Type:        "Custom",
		Regime:      "AMRAP",
		ScoreType:   "Rounds+Reps",
		Description: "Original description",
		IsStandard:  false,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create WOD: %v", err)
	}

	// Update the WOD
	wod.Name = "Updated WOD"
	wod.Description = "Updated description"
	wod.ScoreType = "Time"

	if err := repo.Update(wod); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(wod.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updated.Name != "Updated WOD" {
		t.Errorf("Update() name = %v, want 'Updated WOD'", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Update() description = %v, want 'Updated description'", updated.Description)
	}
	if updated.ScoreType != "Time" {
		t.Errorf("Update() score_type = %v, want 'Time'", updated.ScoreType)
	}
}

func TestWODRepository_Update_StandardWOD(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a standard WOD
	wod := &domain.WOD{
		Name:        "Standard WOD",
		Source:      "CrossFit",
		Type:        "Benchmark",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "Cannot be updated",
		IsStandard:  true,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create WOD: %v", err)
	}

	// Try to update the standard WOD
	wod.Name = "Should Not Update"
	err = repo.Update(wod)
	if err == nil {
		t.Error("Update() should fail for standard WODs")
	}
}

func TestWODRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a custom WOD
	wod := &domain.WOD{
		Name:        "To Delete",
		Source:      "Custom",
		Type:        "Custom",
		Regime:      "AMRAP",
		ScoreType:   "Rounds+Reps",
		Description: "This will be deleted",
		IsStandard:  false,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create WOD: %v", err)
	}

	// Delete the WOD
	if err := repo.Delete(wod.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's deleted
	deleted, err := repo.GetByID(wod.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if deleted != nil {
		t.Error("Delete() WOD still exists after deletion")
	}
}

func TestWODRepository_Delete_StandardWOD(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a standard WOD
	wod := &domain.WOD{
		Name:        "Standard To Delete",
		Source:      "CrossFit",
		Type:        "Benchmark",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "Cannot be deleted",
		IsStandard:  true,
	}
	if err := repo.Create(wod); err != nil {
		t.Fatalf("failed to create WOD: %v", err)
	}

	// Try to delete the standard WOD
	err = repo.Delete(wod.ID)
	if err == nil {
		t.Error("Delete() should fail for standard WODs")
	}
}

func TestWODRepository_Search(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create test WODs
	wods := []string{
		"Fran",
		"Diane",
		"Helen",
		"Annie",
		"Elizabeth",
		"Murph",
		"Michael",
	}

	for _, name := range wods {
		wod := &domain.WOD{
			Name:        name,
			Source:      "CrossFit",
			Type:        "Benchmark",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Test WOD",
			IsStandard:  true,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create WOD: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     string
		limit     int
		wantCount int
	}{
		{
			name:      "search for 'an' (Fran, Diane, Annie)",
			query:     "an",
			limit:     10,
			wantCount: 3, // Fran, Diane, Annie (all contain 'an')
		},
		{
			name:      "search for 'M'",
			query:     "M",
			limit:     10,
			wantCount: 2, // Murph, Michael
		},
		{
			name:      "search with limit",
			query:     "a",
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

func TestWODRepository_List_WithFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create WODs with different types and sources
	testWODs := []struct {
		name       string
		source     string
		wodType    string
		regime     string
		scoreType  string
		isStandard bool
	}{
		{"Fran", "CrossFit", "Girl", "For Time", "Time", true},
		{"Murph", "CrossFit", "Hero", "For Time", "Time", true},
		{"Cindy", "CrossFit", "Girl", "AMRAP", "Rounds+Reps", true},
		{"Custom1", "Custom", "Custom", "AMRAP", "Rounds+Reps", false},
		{"Custom2", "Custom", "Custom", "For Time", "Time", false},
	}

	for _, tw := range testWODs {
		wod := &domain.WOD{
			Name:        tw.name,
			Source:      tw.source,
			Type:        tw.wodType,
			Regime:      tw.regime,
			ScoreType:   tw.scoreType,
			Description: "Test WOD",
			IsStandard:  tw.isStandard,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create WOD: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   map[string]interface{}
		wantCount int
	}{
		{
			name:      "no filters",
			filters:   nil,
			wantCount: 5,
		},
		{
			name:      "filter by source",
			filters:   map[string]interface{}{"source": "CrossFit"},
			wantCount: 3,
		},
		{
			name:      "filter by type",
			filters:   map[string]interface{}{"type": "Girl"},
			wantCount: 2,
		},
		{
			name:      "filter by regime",
			filters:   map[string]interface{}{"regime": "AMRAP"},
			wantCount: 2,
		},
		{
			name:      "filter by is_standard",
			filters:   map[string]interface{}{"is_standard": true},
			wantCount: 3,
		},
		{
			name:      "filter by is_standard false",
			filters:   map[string]interface{}{"is_standard": false},
			wantCount: 2,
		},
		{
			name: "combined filters",
			filters: map[string]interface{}{
				"source":     "CrossFit",
				"score_type": "Time",
			},
			wantCount: 2, // Fran and Murph
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.List(tt.filters, 0, 0)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(results) != tt.wantCount {
				t.Errorf("List() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestWODRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "wodcount@example.com")

	repo := NewWODRepository(db)

	// Initial count should be 0
	count, err := repo.Count(nil)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() initial = %d, want 0", count)
	}

	// Create WODs
	for i := 0; i < 3; i++ {
		wod := &domain.WOD{
			Name:        "Count WOD " + string(rune('A'+i)),
			Source:      "CrossFit",
			Type:        "Benchmark",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Test WOD",
			IsStandard:  true,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create WOD: %v", err)
		}
	}

	// Create user-specific WODs
	for i := 0; i < 2; i++ {
		wod := &domain.WOD{
			Name:        "User Count WOD " + string(rune('A'+i)),
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "User WOD",
			IsStandard:  false,
			CreatedBy:   &user.ID,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create user WOD: %v", err)
		}
	}

	// Count all WODs
	count, err = repo.Count(nil)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 5 {
		t.Errorf("Count() all = %d, want 5", count)
	}

	// Count user-specific WODs
	count, err = repo.Count(&user.ID)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Count() user = %d, want 2", count)
	}
}

func TestWODRepository_CopyToStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "wodcreator@example.com")

	repo := NewWODRepository(db)

	// Create a custom WOD
	custom := &domain.WOD{
		Name:        "My Custom WOD",
		Source:      "Custom",
		Type:        "Custom",
		Regime:      "AMRAP",
		ScoreType:   "Rounds+Reps",
		Description: "A great custom WOD",
		IsStandard:  false,
		CreatedBy:   &user.ID,
	}
	if err := repo.Create(custom); err != nil {
		t.Fatalf("failed to create custom WOD: %v", err)
	}

	// Copy to standard with a new name
	standard, err := repo.CopyToStandard(custom.ID, "Official Custom WOD")
	if err != nil {
		t.Fatalf("CopyToStandard() error = %v", err)
	}

	if standard.Name != "Official Custom WOD" {
		t.Errorf("CopyToStandard() name = %v, want 'Official Custom WOD'", standard.Name)
	}
	if !standard.IsStandard {
		t.Error("CopyToStandard() new WOD should be standard")
	}
	if standard.CreatedBy != nil {
		t.Error("CopyToStandard() standard WOD should have no creator")
	}
	if standard.Description != custom.Description {
		t.Errorf("CopyToStandard() description = %v, want %v", standard.Description, custom.Description)
	}
}

func TestWODRepository_CopyToStandard_SameName(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "wodcreator2@example.com")

	repo := NewWODRepository(db)

	// Create a custom WOD
	custom := &domain.WOD{
		Name:        "Convert Me",
		Source:      "Custom",
		Type:        "Custom",
		Regime:      "For Time",
		ScoreType:   "Time",
		Description: "This will become standard",
		IsStandard:  false,
		CreatedBy:   &user.ID,
	}
	if err := repo.Create(custom); err != nil {
		t.Fatalf("failed to create custom WOD: %v", err)
	}

	// Copy to standard with the same name (should convert in place)
	standard, err := repo.CopyToStandard(custom.ID, "Convert Me")
	if err != nil {
		t.Fatalf("CopyToStandard() error = %v", err)
	}

	if standard.ID != custom.ID {
		t.Error("CopyToStandard() with same name should convert in place, keeping same ID")
	}
	if !standard.IsStandard {
		t.Error("CopyToStandard() WOD should be standard")
	}
	if standard.CreatedBy != nil {
		t.Error("CopyToStandard() standard WOD should have no creator")
	}
}

func TestWODRepository_CountAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a mix of WODs
	for i := 0; i < 3; i++ {
		wod := &domain.WOD{
			Name:        "Standard WOD Count " + string(rune('A'+i)),
			Source:      "CrossFit",
			Type:        "Benchmark",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Standard WOD",
			IsStandard:  true,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create standard WOD: %v", err)
		}
	}

	for i := 0; i < 5; i++ {
		wod := &domain.WOD{
			Name:        "Custom WOD Count " + string(rune('A'+i)),
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "Custom WOD",
			IsStandard:  false,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create custom WOD: %v", err)
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

func TestWODRepository_ListAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewWODRepository(db)

	// Create a mix of WODs
	standardCount := 2
	customCount := 4

	for i := 0; i < standardCount; i++ {
		wod := &domain.WOD{
			Name:        "Standard List " + string(rune('A'+i)),
			Source:      "CrossFit",
			Type:        "Benchmark",
			Regime:      "For Time",
			ScoreType:   "Time",
			Description: "Standard WOD",
			IsStandard:  true,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create standard WOD: %v", err)
		}
	}

	for i := 0; i < customCount; i++ {
		wod := &domain.WOD{
			Name:        "Custom List " + string(rune('A'+i)),
			Source:      "Custom",
			Type:        "Custom",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "Custom WOD",
			IsStandard:  false,
		}
		if err := repo.Create(wod); err != nil {
			t.Fatalf("failed to create custom WOD: %v", err)
		}
	}

	// ListAllUserCreated should only return custom WODs
	wods, err := repo.ListAllUserCreated(0, 0)
	if err != nil {
		t.Fatalf("ListAllUserCreated() error = %v", err)
	}

	if len(wods) != customCount {
		t.Errorf("ListAllUserCreated() returned %d WODs, want %d", len(wods), customCount)
	}

	for _, w := range wods {
		if w.IsStandard {
			t.Errorf("ListAllUserCreated() returned standard WOD: %s", w.Name)
		}
	}
}
