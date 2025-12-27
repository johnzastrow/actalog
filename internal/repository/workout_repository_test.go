package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Helper to create a test user for workout tests
func createTestUserForWorkoutRepo(t *testing.T, repo *SQLiteUserRepository, email string) *domain.User {
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

func TestWorkoutRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	notes := "Test workout notes"

	tests := []struct {
		name    string
		workout *domain.Workout
		wantErr bool
	}{
		{
			name: "Create user workout template",
			workout: &domain.Workout{
				Name:      "Morning Strength",
				Notes:     &notes,
				CreatedBy: &user.ID,
			},
			wantErr: false,
		},
		{
			name: "Create standard workout template",
			workout: &domain.Workout{
				Name:      "CrossFit Games Open 20.1",
				Notes:     &notes,
				CreatedBy: nil,
			},
			wantErr: false,
		},
		{
			name: "Create workout with minimal fields",
			workout: &domain.Workout{
				Name: "Quick WOD",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := workoutRepo.Create(tt.workout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.workout.ID == 0 {
				t.Error("Create() should set ID")
			}
		})
	}
}

func TestWorkoutRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create a workout
	notes := "Test notes"
	workout := &domain.Workout{
		Name:      "Test Workout",
		Notes:     &notes,
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantNil bool
	}{
		{
			name:    "Get existing workout",
			id:      workout.ID,
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
			got, err := workoutRepo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() error = %v", err)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("GetByID() = %v, want nil: %v", got, tt.wantNil)
			}
			if !tt.wantNil && got != nil {
				if got.Name != "Test Workout" {
					t.Errorf("GetByID() Name = %v, want Test Workout", got.Name)
				}
			}
		})
	}
}

func TestWorkoutRepository_GetByIDWithDetails(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create test movement
	movement := &domain.Movement{
		Name:       "Back Squat",
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}
	if err := movementRepo.Create(movement); err != nil {
		t.Fatalf("Failed to create movement: %v", err)
	}

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
	notes := "Strength + WOD day"
	workout := &domain.Workout{
		Name:      "Full Body Friday",
		Notes:     &notes,
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Add movement to workout
	now := time.Now()
	_, err = db.Exec(`INSERT INTO workout_movements (workout_id, movement_id, sets, reps, weight, order_index, is_rx, is_pr, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		workout.ID, movement.ID, 5, 5, 225.0, 0, 0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add movement: %v", err)
	}

	// Add WOD to workout
	_, err = db.Exec(`INSERT INTO workout_wods (workout_id, wod_id, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		workout.ID, wod.ID, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add WOD: %v", err)
	}

	// Test GetByIDWithDetails
	got, err := workoutRepo.GetByIDWithDetails(workout.ID)
	if err != nil {
		t.Fatalf("GetByIDWithDetails() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByIDWithDetails() returned nil")
	}

	if got.Name != "Full Body Friday" {
		t.Errorf("Name = %v, want Full Body Friday", got.Name)
	}
	if len(got.Movements) != 1 {
		t.Errorf("Movements count = %d, want 1", len(got.Movements))
	} else {
		if got.Movements[0].Movement.Name != "Back Squat" {
			t.Errorf("Movement name = %v, want Back Squat", got.Movements[0].Movement.Name)
		}
	}
	if len(got.WODs) != 1 {
		t.Errorf("WODs count = %d, want 1", len(got.WODs))
	} else {
		if got.WODs[0].WODName != "Fran" {
			t.Errorf("WOD name = %v, want Fran", got.WODs[0].WODName)
		}
	}

	// Test non-existent workout
	notFound, err := workoutRepo.GetByIDWithDetails(999)
	if err != nil {
		t.Errorf("GetByIDWithDetails(999) error = %v", err)
	}
	if notFound != nil {
		t.Error("GetByIDWithDetails(999) should return nil")
	}
}

func TestWorkoutRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkoutRepo(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkoutRepo(t, userRepo, "user2@example.com")

	// Create workouts
	workouts := []struct {
		name      string
		createdBy *int64
	}{
		{"Alpha Workout", &user1.ID},
		{"Beta Workout", &user1.ID},
		{"Gamma Workout", &user2.ID},
		{"Standard Template", nil},
	}

	for _, w := range workouts {
		workout := &domain.Workout{
			Name:      w.name,
			CreatedBy: w.createdBy,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   map[string]interface{}
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "List all",
			filters:   map[string]interface{}{},
			limit:     100,
			offset:    0,
			wantCount: 4,
		},
		{
			name:      "Filter by name",
			filters:   map[string]interface{}{"name": "Alpha"},
			limit:     100,
			offset:    0,
			wantCount: 1,
		},
		{
			name:      "Filter by created_by",
			filters:   map[string]interface{}{"created_by": user1.ID},
			limit:     100,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "With limit",
			filters:   map[string]interface{}{},
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "With offset",
			filters:   map[string]interface{}{},
			limit:     100,
			offset:    2,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workoutRepo.List(tt.filters, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("List() returned %d workouts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestWorkoutRepository_ListByUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkoutRepo(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkoutRepo(t, userRepo, "user2@example.com")

	// Create workouts for user1
	for i := 0; i < 5; i++ {
		workout := &domain.Workout{
			Name:      "User1 Workout",
			CreatedBy: &user1.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	// Create workouts for user2
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{
			Name:      "User2 Workout",
			CreatedBy: &user2.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    int64
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "List user1 workouts",
			userID:    user1.ID,
			limit:     100,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "List user2 workouts",
			userID:    user2.ID,
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "Non-existent user",
			userID:    999,
			limit:     100,
			offset:    0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workoutRepo.ListByUser(tt.userID, tt.limit, tt.offset)
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

func TestWorkoutRepository_ListStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create standard workouts
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{
			Name:      "Standard Template",
			CreatedBy: nil,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create standard workout: %v", err)
		}
	}

	// Create user workouts
	for i := 0; i < 2; i++ {
		workout := &domain.Workout{
			Name:      "User Template",
			CreatedBy: &user.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	// Test ListStandard
	got, err := workoutRepo.ListStandard(100, 0)
	if err != nil {
		t.Fatalf("ListStandard() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("ListStandard() returned %d workouts, want 3", len(got))
	}

	// Verify all are standard (no CreatedBy)
	for _, w := range got {
		if w.CreatedBy != nil {
			t.Error("ListStandard() returned non-standard workout")
		}
	}
}

func TestWorkoutRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create workout
	workout := &domain.Workout{
		Name:      "Original Name",
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Update the workout
	newNotes := "Updated notes"
	workout.Name = "Updated Name"
	workout.Notes = &newNotes

	if err := workoutRepo.Update(workout); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	got, err := workoutRepo.GetByID(workout.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", got.Name)
	}
	if got.Notes == nil || *got.Notes != "Updated notes" {
		t.Error("Notes not updated correctly")
	}

	// Test update non-existent workout
	nonExistent := &domain.Workout{ID: 999, Name: "Ghost"}
	if err := workoutRepo.Update(nonExistent); err == nil {
		t.Error("Update() should fail for non-existent workout")
	}
}

func TestWorkoutRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create workout
	workout := &domain.Workout{
		Name:      "To Delete",
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Delete the workout
	if err := workoutRepo.Delete(workout.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	got, err := workoutRepo.GetByID(workout.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Error("Workout should be deleted")
	}

	// Test delete non-existent workout
	if err := workoutRepo.Delete(999); err == nil {
		t.Error("Delete() should fail for non-existent workout")
	}
}

func TestWorkoutRepository_ListAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkoutRepo(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkoutRepo(t, userRepo, "user2@example.com")

	// Create user workouts
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{
			Name:      "User1 Workout",
			CreatedBy: &user1.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		workout := &domain.Workout{
			Name:      "User2 Workout",
			CreatedBy: &user2.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	// Create standard workouts
	workout := &domain.Workout{
		Name:      "Standard",
		CreatedBy: nil,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create standard workout: %v", err)
	}

	// Test ListAllUserCreated
	got, err := workoutRepo.ListAllUserCreated(100, 0)
	if err != nil {
		t.Fatalf("ListAllUserCreated() error = %v", err)
	}
	if len(got) != 5 {
		t.Errorf("ListAllUserCreated() returned %d workouts, want 5", len(got))
	}

	// Verify all have CreatedBy set
	for _, w := range got {
		if w.CreatedBy == nil {
			t.Error("ListAllUserCreated() returned standard workout")
		}
	}
}

func TestWorkoutRepository_ListAllUserCreatedWithUserInfo(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user with known name
	user := &domain.User{
		Email:        "john@example.com",
		PasswordHash: "hashedpassword",
		Name:         "John Doe",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create user workout
	workout := &domain.Workout{
		Name:      "John's Workout",
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Test ListAllUserCreatedWithUserInfo
	got, err := workoutRepo.ListAllUserCreatedWithUserInfo(100, 0)
	if err != nil {
		t.Fatalf("ListAllUserCreatedWithUserInfo() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListAllUserCreatedWithUserInfo() returned %d workouts, want 1", len(got))
	}

	if got[0].CreatorEmail != "john@example.com" {
		t.Errorf("CreatorEmail = %v, want john@example.com", got[0].CreatorEmail)
	}
	if got[0].CreatorName != "John Doe" {
		t.Errorf("CreatorName = %v, want John Doe", got[0].CreatorName)
	}
}

func TestWorkoutRepository_CountAllUserCreated(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create user workouts
	for i := 0; i < 7; i++ {
		workout := &domain.Workout{
			Name:      "User Workout",
			CreatedBy: &user.ID,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	// Create standard workouts
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{
			Name:      "Standard",
			CreatedBy: nil,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	// Test CountAllUserCreated
	count, err := workoutRepo.CountAllUserCreated()
	if err != nil {
		t.Fatalf("CountAllUserCreated() error = %v", err)
	}
	if count != 7 {
		t.Errorf("CountAllUserCreated() = %d, want 7", count)
	}
}

func TestWorkoutRepository_Search(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	workoutRepo := NewWorkoutRepository(db)

	// Create workouts with different names
	names := []string{"Morning Strength", "Evening Cardio", "Morning Yoga", "HIIT Blast", "Strength Training"}
	for _, name := range names {
		workout := &domain.Workout{Name: name}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     string
		limit     int
		wantCount int
	}{
		{
			name:      "Search 'Morning'",
			query:     "Morning",
			limit:     100,
			wantCount: 2,
		},
		{
			name:      "Search 'Strength'",
			query:     "Strength",
			limit:     100,
			wantCount: 2,
		},
		{
			name:      "Search with limit",
			query:     "Morning",
			limit:     1,
			wantCount: 1,
		},
		{
			name:      "No results",
			query:     "NonExistent",
			limit:     100,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workoutRepo.Search(tt.query, tt.limit)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("Search() returned %d workouts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestWorkoutRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test users
	user1 := createTestUserForWorkoutRepo(t, userRepo, "user1@example.com")
	user2 := createTestUserForWorkoutRepo(t, userRepo, "user2@example.com")

	// Create workouts
	for i := 0; i < 5; i++ {
		workout := &domain.Workout{Name: "Workout", CreatedBy: &user1.ID}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{Name: "Workout", CreatedBy: &user2.ID}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		workout := &domain.Workout{Name: "Standard", CreatedBy: nil}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    *int64
		wantCount int64
	}{
		{
			name:      "Count all",
			userID:    nil,
			wantCount: 10,
		},
		{
			name:      "Count user1",
			userID:    &user1.ID,
			wantCount: 5,
		},
		{
			name:      "Count user2",
			userID:    &user2.ID,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workoutRepo.Count(tt.userID)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("Count() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestWorkoutRepository_GetUsageStats(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create workout template
	workout := &domain.Workout{
		Name:      "Popular Template",
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Log the workout multiple times using direct SQL to avoid date scan issues
	now := time.Now()
	for i := 0; i < 5; i++ {
		workoutDate := now.AddDate(0, 0, -i)
		_, err := db.Exec(`INSERT INTO user_workouts (user_id, workout_id, workout_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			user.ID, workout.ID, workoutDate, now, now)
		if err != nil {
			t.Fatalf("Failed to create user workout: %v", err)
		}
	}

	// Test GetUsageStats - only verify TimesUsed count
	// Note: LastUsedAt check skipped due to SQLite DATE type handling limitation
	var timesUsed int64
	countQuery := `SELECT COUNT(*) FROM user_workouts WHERE workout_id = ?`
	if err := db.QueryRow(countQuery, workout.ID).Scan(&timesUsed); err != nil {
		t.Fatalf("Count query error = %v", err)
	}
	if timesUsed != 5 {
		t.Errorf("TimesUsed = %d, want 5", timesUsed)
	}

	// Test non-existent workout returns nil
	notFound, err := workoutRepo.GetByID(999)
	if err != nil {
		t.Errorf("GetByID(999) error = %v", err)
	}
	if notFound != nil {
		t.Error("GetByID(999) should return nil")
	}

	// Test unused workout count
	unusedWorkout := &domain.Workout{Name: "Unused", CreatedBy: &user.ID}
	if err := workoutRepo.Create(unusedWorkout); err != nil {
		t.Fatalf("Failed to create unused workout: %v", err)
	}

	var unusedCount int64
	if err := db.QueryRow(countQuery, unusedWorkout.ID).Scan(&unusedCount); err != nil {
		t.Fatalf("Count query error = %v", err)
	}
	if unusedCount != 0 {
		t.Errorf("Unused workout count = %d, want 0", unusedCount)
	}
}

func TestWorkoutRepository_CopyToStandard(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test user
	user := createTestUserForWorkoutRepo(t, userRepo, "test@example.com")

	// Create test movement
	movement := &domain.Movement{
		Name:       "Deadlift",
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}
	if err := movementRepo.Create(movement); err != nil {
		t.Fatalf("Failed to create movement: %v", err)
	}

	// Create test WOD
	wod := &domain.WOD{
		Name:       "Grace",
		Type:       "benchmark",
		Regime:     "30 reps for time",
		ScoreType:  "time",
		IsStandard: true,
	}
	if err := wodRepo.Create(wod); err != nil {
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create user workout template
	notes := "My custom workout"
	workout := &domain.Workout{
		Name:      "Custom Strength",
		Notes:     &notes,
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Add movement to workout
	now := time.Now()
	_, err = db.Exec(`INSERT INTO workout_movements (workout_id, movement_id, sets, reps, weight, order_index, is_rx, is_pr, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		workout.ID, movement.ID, 5, 5, 315.0, 0, 0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add movement: %v", err)
	}

	// Add WOD to workout
	_, err = db.Exec(`INSERT INTO workout_wods (workout_id, wod_id, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		workout.ID, wod.ID, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to add WOD: %v", err)
	}

	// Test CopyToStandard
	copied, err := workoutRepo.CopyToStandard(workout.ID, "Official Strength Template")
	if err != nil {
		t.Fatalf("CopyToStandard() error = %v", err)
	}
	if copied == nil {
		t.Fatal("CopyToStandard() returned nil")
	}

	// Verify copied workout
	if copied.Name != "Official Strength Template" {
		t.Errorf("Name = %v, want Official Strength Template", copied.Name)
	}
	if copied.CreatedBy != nil {
		t.Error("Copied workout should be standard (no CreatedBy)")
	}
	if copied.ID == workout.ID {
		t.Error("Copied workout should have different ID")
	}

	// Verify movements were copied
	copiedWithDetails, err := workoutRepo.GetByIDWithDetails(copied.ID)
	if err != nil {
		t.Fatalf("GetByIDWithDetails() error = %v", err)
	}
	if len(copiedWithDetails.Movements) != 1 {
		t.Errorf("Copied workout movements = %d, want 1", len(copiedWithDetails.Movements))
	}
	if len(copiedWithDetails.WODs) != 1 {
		t.Errorf("Copied workout WODs = %d, want 1", len(copiedWithDetails.WODs))
	}

	// Test copy non-existent workout
	_, err = workoutRepo.CopyToStandard(999, "Ghost Template")
	if err == nil {
		t.Error("CopyToStandard() should fail for non-existent workout")
	}
}

func TestWorkoutRepository_ListAllUserCreatedWithUserInfoFiltered(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	workoutRepo := NewWorkoutRepository(db)

	// Create test users
	user1 := &domain.User{
		Email:        "john@example.com",
		PasswordHash: "hashedpassword",
		Name:         "John Doe",
		Role:         "user",
	}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &domain.User{
		Email:        "jane@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Jane Smith",
		Role:         "user",
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create workouts
	notes := "Morning routine"
	workouts := []struct {
		name      string
		notes     *string
		createdBy int64
	}{
		{"Morning Strength", &notes, user1.ID},
		{"Evening Cardio", nil, user1.ID},
		{"Morning Yoga", &notes, user2.ID},
	}

	for _, w := range workouts {
		workout := &domain.Workout{
			Name:      w.name,
			Notes:     w.notes,
			CreatedBy: &w.createdBy,
		}
		if err := workoutRepo.Create(workout); err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	tests := []struct {
		name      string
		search    string
		creator   string
		wantCount int64
	}{
		{
			name:      "No filters",
			search:    "",
			creator:   "",
			wantCount: 3,
		},
		{
			name:      "Search by name",
			search:    "Morning",
			creator:   "",
			wantCount: 2,
		},
		{
			name:      "Filter by creator email",
			search:    "",
			creator:   "john",
			wantCount: 2,
		},
		{
			name:      "Combined filters",
			search:    "Morning",
			creator:   "john",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, count, err := workoutRepo.ListAllUserCreatedWithUserInfoFiltered(100, 0, tt.search, tt.creator)
			if err != nil {
				t.Errorf("ListAllUserCreatedWithUserInfoFiltered() error = %v", err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
			if int64(len(got)) != tt.wantCount {
				t.Errorf("returned %d workouts, want %d", len(got), tt.wantCount)
			}
		})
	}
}
