package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Helper to create a test user
func createTestUserForWorkout(t *testing.T, repo *SQLiteUserRepository, email string) *domain.User {
	user := &domain.User{
		Email:        email,
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "user",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

// Helper to create a test movement
func createTestMovementForWorkout(t *testing.T, repo *MovementRepository, name string) *domain.Movement {
	movement := &domain.Movement{
		Name:       name,
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}
	if err := repo.Create(movement); err != nil {
		t.Fatalf("Failed to create test movement: %v", err)
	}
	return movement
}

func TestUserWorkoutRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create workout template
	now := time.Now()
	_, err = db.Exec(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"Test Workout", "Test notes", user.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to create workout template: %v", err)
	}

	workoutID := int64(1)
	workoutType := "strength"
	totalTime := 45
	notes := "Great workout!"

	tests := []struct {
		name        string
		userWorkout *domain.UserWorkout
		wantErr     bool
	}{
		{
			name: "Create template-based workout",
			userWorkout: &domain.UserWorkout{
				UserID:      user.ID,
				WorkoutID:   &workoutID,
				WorkoutDate: time.Now(),
				WorkoutType: &workoutType,
				TotalTime:   &totalTime,
				Notes:       &notes,
			},
			wantErr: false,
		},
		{
			name: "Create ad-hoc workout with name",
			userWorkout: &domain.UserWorkout{
				UserID:      user.ID,
				WorkoutName: ptrString("Ad-hoc Workout"),
				WorkoutDate: time.Now(),
				WorkoutType: &workoutType,
			},
			wantErr: false,
		},
		{
			name: "Create workout with minimal fields",
			userWorkout: &domain.UserWorkout{
				UserID:      user.ID,
				WorkoutDate: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userWorkoutRepo.Create(tt.userWorkout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.userWorkout.ID == 0 {
				t.Error("Create() should set ID")
			}
		})
	}
}

func TestUserWorkoutRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create a user workout
	workoutType := "strength"
	notes := "Test notes"
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutName: ptrString("My Workout"),
		WorkoutDate: time.Now(),
		WorkoutType: &workoutType,
		Notes:       &notes,
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantNil bool
	}{
		{
			name:    "Get existing workout",
			id:      userWorkout.ID,
			wantNil: false,
		},
		{
			name:    "Get non-existent workout",
			id:      999,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() error = %v", err)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("GetByID() = %v, want nil: %v", got, tt.wantNil)
			}
			if !tt.wantNil && got != nil {
				if got.ID != tt.id {
					t.Errorf("GetByID() ID = %v, want %v", got.ID, tt.id)
				}
				if *got.WorkoutName != "My Workout" {
					t.Errorf("GetByID() WorkoutName = %v, want My Workout", *got.WorkoutName)
				}
			}
		})
	}
}

func TestUserWorkoutRepository_ListByUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkout(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkout(t, userRepo, "user2@example.com")

	// Create workouts for user1
	for i := 0; i < 5; i++ {
		uw := &domain.UserWorkout{
			UserID:      user1.ID,
			WorkoutName: ptrString("Workout " + string(rune('A'+i))),
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	// Create workout for user2
	uw := &domain.UserWorkout{
		UserID:      user2.ID,
		WorkoutName: ptrString("User2 Workout"),
		WorkoutDate: time.Now(),
	}
	if err := userWorkoutRepo.Create(uw); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	tests := []struct {
		name      string
		userID    int64
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "List all user1 workouts",
			userID:    user1.ID,
			limit:     100,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "List with limit",
			userID:    user1.ID,
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "List with offset",
			userID:    user1.ID,
			limit:     100,
			offset:    3,
			wantCount: 2,
		},
		{
			name:      "List user2 workouts",
			userID:    user2.ID,
			limit:     100,
			offset:    0,
			wantCount: 1,
		},
		{
			name:      "List non-existent user",
			userID:    999,
			limit:     100,
			offset:    0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.ListByUser(tt.userID, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("ListByUser() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ListByUser() returned %d workouts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestUserWorkoutRepository_ListByUserAndDateRange(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create workouts on different dates
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -10),
		now.AddDate(0, 0, -5),
		now.AddDate(0, 0, -3),
		now.AddDate(0, 0, -1),
		now,
	}

	for i, date := range dates {
		uw := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutName: ptrString("Workout " + string(rune('A'+i))),
			WorkoutDate: date,
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		wantCount int
	}{
		{
			name:      "All dates",
			startDate: now.AddDate(0, 0, -15),
			endDate:   now.AddDate(0, 0, 1),
			wantCount: 5,
		},
		{
			name:      "Last 5 days",
			startDate: now.AddDate(0, 0, -5),
			endDate:   now.AddDate(0, 0, 1),
			wantCount: 4,
		},
		{
			name:      "Last 2 days",
			startDate: now.AddDate(0, 0, -2),
			endDate:   now.AddDate(0, 0, 1),
			wantCount: 2,
		},
		{
			name:      "Future range",
			startDate: now.AddDate(0, 0, 10),
			endDate:   now.AddDate(0, 0, 20),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.ListByUserAndDateRange(user.ID, tt.startDate, tt.endDate)
			if err != nil {
				t.Errorf("ListByUserAndDateRange() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ListByUserAndDateRange() returned %d workouts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestUserWorkoutRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkout(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkout(t, userRepo, "user2@example.com")

	// Create user workout
	workoutType := "strength"
	userWorkout := &domain.UserWorkout{
		UserID:      user1.ID,
		WorkoutName: ptrString("Original Name"),
		WorkoutDate: time.Now(),
		WorkoutType: &workoutType,
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	tests := []struct {
		name    string
		update  func(*domain.UserWorkout)
		userID  int64
		wantErr bool
	}{
		{
			name: "Update workout name",
			update: func(uw *domain.UserWorkout) {
				uw.WorkoutName = ptrString("Updated Name")
			},
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name: "Update notes",
			update: func(uw *domain.UserWorkout) {
				uw.Notes = ptrString("Updated notes")
			},
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name: "Update by wrong user fails",
			update: func(uw *domain.UserWorkout) {
				uw.UserID = user2.ID
				uw.Notes = ptrString("Hacker notes")
			},
			userID:  user2.ID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original state for each test
			userWorkout.UserID = tt.userID
			tt.update(userWorkout)

			err := userWorkoutRepo.Update(userWorkout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserWorkoutRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkout(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkout(t, userRepo, "user2@example.com")

	// Create user workout
	userWorkout := &domain.UserWorkout{
		UserID:      user1.ID,
		WorkoutName: ptrString("To Delete"),
		WorkoutDate: time.Now(),
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		userID  int64
		wantErr bool
	}{
		{
			name:    "Delete by wrong user fails",
			id:      userWorkout.ID,
			userID:  user2.ID,
			wantErr: true,
		},
		{
			name:    "Delete non-existent workout fails",
			id:      999,
			userID:  user1.ID,
			wantErr: true,
		},
		{
			name:    "Delete own workout succeeds",
			id:      userWorkout.ID,
			userID:  user1.ID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userWorkoutRepo.Delete(tt.id, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify deletion
	deleted, err := userWorkoutRepo.GetByID(userWorkout.ID)
	if err != nil {
		t.Errorf("GetByID after delete error = %v", err)
	}
	if deleted != nil {
		t.Error("Workout should be deleted")
	}
}

func TestUserWorkoutRepository_GetByUserWorkoutDate(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create workout template
	now := time.Now()
	_, err = db.Exec(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"Test Workout Template", "Notes", user.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to create workout template: %v", err)
	}
	workoutID := int64(1)

	// Create user workout with template
	workoutDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutID:   &workoutID,
		WorkoutDate: workoutDate,
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	tests := []struct {
		name      string
		userID    int64
		workoutID int64
		date      time.Time
		wantNil   bool
	}{
		{
			name:      "Find existing workout on date",
			userID:    user.ID,
			workoutID: workoutID,
			date:      workoutDate,
			wantNil:   false,
		},
		{
			name:      "Wrong date returns nil",
			userID:    user.ID,
			workoutID: workoutID,
			date:      time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC),
			wantNil:   true,
		},
		{
			name:      "Wrong workout ID returns nil",
			userID:    user.ID,
			workoutID: 999,
			date:      workoutDate,
			wantNil:   true,
		},
		{
			name:      "Wrong user ID returns nil",
			userID:    999,
			workoutID: workoutID,
			date:      workoutDate,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.GetByUserWorkoutDate(tt.userID, tt.workoutID, tt.date)
			if err != nil {
				t.Errorf("GetByUserWorkoutDate() error = %v", err)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("GetByUserWorkoutDate() = %v, want nil: %v", got, tt.wantNil)
			}
		})
	}
}

func TestUserWorkoutRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkout(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkout(t, userRepo, "user2@example.com")

	// Create workouts for user1
	for i := 0; i < 7; i++ {
		uw := &domain.UserWorkout{
			UserID:      user1.ID,
			WorkoutName: ptrString("Workout"),
			WorkoutDate: time.Now(),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	// Create workouts for user2
	for i := 0; i < 3; i++ {
		uw := &domain.UserWorkout{
			UserID:      user2.ID,
			WorkoutName: ptrString("Workout"),
			WorkoutDate: time.Now(),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    int64
		wantCount int64
	}{
		{
			name:      "Count user1 workouts",
			userID:    user1.ID,
			wantCount: 7,
		},
		{
			name:      "Count user2 workouts",
			userID:    user2.ID,
			wantCount: 3,
		},
		{
			name:      "Count non-existent user",
			userID:    999,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.Count(tt.userID)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("Count() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestUserWorkoutRepository_GetRecentForUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create workout templates
	now := time.Now()
	for i := 0; i < 5; i++ {
		_, err = db.Exec(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			"Workout Template "+string(rune('A'+i)), "Template notes", user.ID, now, now)
		if err != nil {
			t.Fatalf("Failed to create workout template: %v", err)
		}
	}

	// Create user workouts linked to templates
	for i := 1; i <= 5; i++ {
		workoutID := int64(i)
		uw := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutID:   &workoutID,
			WorkoutDate: now.AddDate(0, 0, -i),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    int64
		limit     int
		wantCount int
	}{
		{
			name:      "Get all recent",
			userID:    user.ID,
			limit:     10,
			wantCount: 5,
		},
		{
			name:      "Get limited recent",
			userID:    user.ID,
			limit:     3,
			wantCount: 3,
		},
		{
			name:      "Non-existent user",
			userID:    999,
			limit:     10,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.GetRecentForUser(tt.userID, tt.limit)
			if err != nil {
				t.Errorf("GetRecentForUser() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetRecentForUser() returned %d workouts, want %d", len(got), tt.wantCount)
			}
			// Verify workout names are populated
			for _, uw := range got {
				if uw.WorkoutName == "" {
					t.Error("GetRecentForUser() should populate WorkoutName")
				}
			}
		})
	}
}

func TestUserWorkoutRepository_GetByIDWithDetails_AdHoc(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create test movement
	movement := createTestMovementForWorkout(t, movementRepo, "Back Squat")

	// Create ad-hoc user workout
	workoutType := "strength"
	notes := "Great workout"
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutName: ptrString("Ad-hoc Squat Day"),
		WorkoutDate: time.Now(),
		WorkoutType: &workoutType,
		Notes:       &notes,
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Add performance movement
	now := time.Now()
	_, err = db.Exec(`INSERT INTO user_workout_movements (user_workout_id, movement_id, sets, reps, weight, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userWorkout.ID, movement.ID, 5, 5, 225.0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add performance movement: %v", err)
	}

	tests := []struct {
		name          string
		workoutID     int64
		userID        int64
		wantNil       bool
		wantErr       bool
		wantMovements int
	}{
		{
			name:          "Get ad-hoc workout with details",
			workoutID:     userWorkout.ID,
			userID:        user.ID,
			wantNil:       false,
			wantErr:       false,
			wantMovements: 1,
		},
		{
			name:      "Wrong user returns error",
			workoutID: userWorkout.ID,
			userID:    999,
			wantNil:   false,
			wantErr:   true,
		},
		{
			name:      "Non-existent workout",
			workoutID: 999,
			userID:    user.ID,
			wantNil:   true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userWorkoutRepo.GetByIDWithDetails(tt.workoutID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByIDWithDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("GetByIDWithDetails() = %v, want nil: %v", got, tt.wantNil)
				return
			}
			if !tt.wantNil && got != nil {
				if got.WorkoutName != "Ad-hoc Squat Day" {
					t.Errorf("GetByIDWithDetails() WorkoutName = %v, want Ad-hoc Squat Day", got.WorkoutName)
				}
				if len(got.PerformanceMovements) != tt.wantMovements {
					t.Errorf("GetByIDWithDetails() PerformanceMovements = %d, want %d", len(got.PerformanceMovements), tt.wantMovements)
				}
			}
		})
	}
}

func TestUserWorkoutRepository_GetByIDWithDetails_TemplateBased(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkout(t, userRepo, "test@example.com")

	// Create test movement
	movement := createTestMovementForWorkout(t, movementRepo, "Deadlift")

	// Create test WOD
	wod := &domain.WOD{
		Name:       "Fran",
		Type:       "benchmark",
		Regime:     "21-15-9",
		ScoreType:  "time",
		IsStandard: true,
	}
	if err := wodRepo.Create(wod); err != nil {
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create workout template
	now := time.Now()
	_, err = db.Exec(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"Strength + WOD", "Test template", user.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to create workout template: %v", err)
	}
	workoutTemplateID := int64(1)

	// Add movement to template
	_, err = db.Exec(`INSERT INTO workout_movements (workout_id, movement_id, sets, reps, weight, order_index, is_rx, is_pr, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		workoutTemplateID, movement.ID, 5, 5, 315.0, 0, 0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add template movement: %v", err)
	}

	// Add WOD to template
	_, err = db.Exec(`INSERT INTO workout_wods (workout_id, wod_id, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		workoutTemplateID, wod.ID, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add template WOD: %v", err)
	}

	// Create user workout from template
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutID:   &workoutTemplateID,
		WorkoutDate: now,
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Add actual performance data
	_, err = db.Exec(`INSERT INTO user_workout_movements (user_workout_id, movement_id, sets, reps, weight, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userWorkout.ID, movement.ID, 5, 5, 325.0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add performance movement: %v", err)
	}

	_, err = db.Exec(`INSERT INTO user_workout_wods (user_workout_id, wod_id, score_type, score_value, time_seconds, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userWorkout.ID, wod.ID, "time", "3:45", 225, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add performance WOD: %v", err)
	}

	// Test GetByIDWithDetails
	got, err := userWorkoutRepo.GetByIDWithDetails(userWorkout.ID, user.ID)
	if err != nil {
		t.Fatalf("GetByIDWithDetails() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByIDWithDetails() returned nil")
	}

	// Verify template data
	if got.WorkoutName != "Strength + WOD" {
		t.Errorf("WorkoutName = %v, want Strength + WOD", got.WorkoutName)
	}
	if len(got.Movements) != 1 {
		t.Errorf("Movements = %d, want 1", len(got.Movements))
	}
	if len(got.WODs) != 1 {
		t.Errorf("WODs = %d, want 1", len(got.WODs))
	}

	// Verify performance data
	if len(got.PerformanceMovements) != 1 {
		t.Errorf("PerformanceMovements = %d, want 1", len(got.PerformanceMovements))
	} else {
		if *got.PerformanceMovements[0].Weight != 325.0 {
			t.Errorf("PerformanceMovements[0].Weight = %v, want 325.0", *got.PerformanceMovements[0].Weight)
		}
	}
	if len(got.PerformanceWODs) != 1 {
		t.Errorf("PerformanceWODs = %d, want 1", len(got.PerformanceWODs))
	} else {
		if *got.PerformanceWODs[0].TimeSeconds != 225 {
			t.Errorf("PerformanceWODs[0].TimeSeconds = %v, want 225", *got.PerformanceWODs[0].TimeSeconds)
		}
	}
}

func TestUserWorkoutRepository_GetActiveUsersThisMonth(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkout(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkout(t, userRepo, "user2@example.com")
	user3 := createTestUserForWorkout(t, userRepo, "user3@example.com")
	user4 := createTestUserForWorkout(t, userRepo, "user4@example.com") // Different org

	// Create organizations
	now := time.Now()
	_, err = db.Exec(`INSERT INTO organizations (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"CrossFit Gym A", "Test gym A", now, now)
	if err != nil {
		t.Fatalf("Failed to create organization A: %v", err)
	}
	orgAID := int64(1)

	_, err = db.Exec(`INSERT INTO organizations (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"CrossFit Gym B", "Test gym B", now, now)
	if err != nil {
		t.Fatalf("Failed to create organization B: %v", err)
	}
	orgBID := int64(2)

	// Add users to organizations
	// user1, user2, user3 are in org A
	// user4 is in org B
	_, err = db.Exec(`INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		user1.ID, orgAID, "member", now, now, now)
	if err != nil {
		t.Fatalf("Failed to add user1 to org A: %v", err)
	}
	_, err = db.Exec(`INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		user2.ID, orgAID, "member", now, now, now)
	if err != nil {
		t.Fatalf("Failed to add user2 to org A: %v", err)
	}
	_, err = db.Exec(`INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		user3.ID, orgAID, "member", now, now, now)
	if err != nil {
		t.Fatalf("Failed to add user3 to org A: %v", err)
	}
	_, err = db.Exec(`INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		user4.ID, orgBID, "member", now, now, now)
	if err != nil {
		t.Fatalf("Failed to add user4 to org B: %v", err)
	}

	// Create workouts for this month
	// user1: 5 workouts
	for i := 0; i < 5; i++ {
		uw := &domain.UserWorkout{
			UserID:      user1.ID,
			WorkoutName: ptrString("Workout"),
			WorkoutDate: now.AddDate(0, 0, -i),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user1 workout: %v", err)
		}
	}

	// user2: 3 workouts
	for i := 0; i < 3; i++ {
		uw := &domain.UserWorkout{
			UserID:      user2.ID,
			WorkoutName: ptrString("Workout"),
			WorkoutDate: now.AddDate(0, 0, -i),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user2 workout: %v", err)
		}
	}

	// user3: 1 workout
	uw := &domain.UserWorkout{
		UserID:      user3.ID,
		WorkoutName: ptrString("Workout"),
		WorkoutDate: now,
	}
	if err := userWorkoutRepo.Create(uw); err != nil {
		t.Fatalf("Failed to create user3 workout: %v", err)
	}

	// user4: 10 workouts (different org, should not appear in user1's results)
	for i := 0; i < 10; i++ {
		uw := &domain.UserWorkout{
			UserID:      user4.ID,
			WorkoutName: ptrString("Workout"),
			WorkoutDate: now.AddDate(0, 0, -i),
		}
		if err := userWorkoutRepo.Create(uw); err != nil {
			t.Fatalf("Failed to create user4 workout: %v", err)
		}
	}

	// Test GetActiveUsersThisMonth for user1
	result, err := userWorkoutRepo.GetActiveUsersThisMonth(user1.ID)
	if err != nil {
		t.Fatalf("GetActiveUsersThisMonth() error = %v", err)
	}

	// Should have at least 1 result (current user)
	if len(result) == 0 {
		t.Fatal("GetActiveUsersThisMonth() returned no results")
	}

	// First result should be current user
	if result[0]["id"] != user1.ID {
		t.Errorf("First result should be current user, got ID = %v, want %v", result[0]["id"], user1.ID)
	}
	if result[0]["workout_count"] != 5 {
		t.Errorf("Current user workout_count = %v, want 5", result[0]["workout_count"])
	}
	if result[0]["is_current"] != true {
		t.Error("Current user should have is_current = true")
	}

	// Other results should be from the same organization
	for i := 1; i < len(result); i++ {
		userID := result[i]["id"].(int64)
		// Verify user is not from org B
		if userID == user4.ID {
			t.Errorf("User from different organization (user4) should not appear in results")
		}
	}

	// Test for user with no organization
	userNoOrg := createTestUserForWorkout(t, userRepo, "noorg@example.com")
	resultNoOrg, err := userWorkoutRepo.GetActiveUsersThisMonth(userNoOrg.ID)
	if err != nil {
		t.Fatalf("GetActiveUsersThisMonth() for user with no org error = %v", err)
	}
	// Should only return the current user
	if len(resultNoOrg) != 1 {
		t.Errorf("User with no org should only see themselves, got %d results", len(resultNoOrg))
	}
}

// Helper function to create string pointer
func ptrString(s string) *string {
	return &s
}
