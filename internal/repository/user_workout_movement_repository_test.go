package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserWorkoutMovementRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Create user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create movement
	movement := &domain.Movement{
		Name: "Back Squat",
		Type: domain.MovementTypeWeightlifting,
	}
	if err := movementRepo.Create(movement); err != nil {
		t.Fatalf("Failed to create movement: %v", err)
	}

	// Create user workout
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now(),
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Create user workout movement
	weight := 225.0
	sets := 5
	reps := 5
	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkout.ID,
		MovementID:    movement.ID,
		Weight:        &weight,
		Sets:          &sets,
		Reps:          &reps,
		IsPR:          false,
		OrderIndex:    0,
	}

	err = uwmRepo.Create(uwm)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if uwm.ID == 0 {
		t.Error("Create() should set ID")
	}
	if uwm.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestUserWorkoutMovementRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	weight := 315.0
	reps := 3
	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkout.ID,
		MovementID:    movement.ID,
		Weight:        &weight,
		Reps:          &reps,
		IsPR:          true,
		OrderIndex:    0,
	}
	uwmRepo.Create(uwm)

	// Get by ID
	got, err := uwmRepo.GetByID(uwm.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.UserWorkoutID != userWorkout.ID {
		t.Errorf("UserWorkoutID = %d, want %d", got.UserWorkoutID, userWorkout.ID)
	}
	if got.Weight == nil || *got.Weight != 315.0 {
		t.Error("Weight should be 315.0")
	}

	// Non-existent
	got, err = uwmRepo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestUserWorkoutMovementRepository_GetByUserWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movements := []*domain.Movement{
		{Name: "Back Squat", Type: domain.MovementTypeWeightlifting},
		{Name: "Bench Press", Type: domain.MovementTypeWeightlifting},
		{Name: "Deadlift", Type: domain.MovementTypeWeightlifting},
	}
	for _, m := range movements {
		movementRepo.Create(m)
	}

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Add movements to user workout
	for i, m := range movements {
		weight := float64(135 + i*50)
		uwm := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    m.ID,
			Weight:        &weight,
			OrderIndex:    i,
		}
		uwmRepo.Create(uwm)
	}

	// Get by user workout ID
	got, err := uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if err != nil {
		t.Fatalf("GetByUserWorkoutID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByUserWorkoutID() returned %d movements, want 3", len(got))
	}

	// Verify order and that Movement is populated
	for i, uwm := range got {
		if uwm.OrderIndex != i {
			t.Errorf("Movement %d has OrderIndex %d, want %d", i, uwm.OrderIndex, i)
		}
		if uwm.Movement == nil {
			t.Error("Movement should be populated from JOIN")
		}
	}
}

func TestUserWorkoutMovementRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	weight := 225.0
	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkout.ID,
		MovementID:    movement.ID,
		Weight:        &weight,
		OrderIndex:    0,
	}
	uwmRepo.Create(uwm)

	// Update
	newWeight := 275.0
	newReps := 5
	uwm.Weight = &newWeight
	uwm.Reps = &newReps
	err = uwmRepo.Update(uwm)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, _ := uwmRepo.GetByID(uwm.ID)
	if got.Weight == nil || *got.Weight != 275.0 {
		t.Error("Weight should be updated to 275.0")
	}
	if got.Reps == nil || *got.Reps != 5 {
		t.Error("Reps should be updated to 5")
	}
}

func TestUserWorkoutMovementRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkout.ID,
		MovementID:    movement.ID,
		OrderIndex:    0,
	}
	uwmRepo.Create(uwm)

	// Delete
	err = uwmRepo.Delete(uwm.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := uwmRepo.GetByID(uwm.ID)
	if got != nil {
		t.Error("Movement should be deleted")
	}

	// Delete non-existent
	err = uwmRepo.Delete(999)
	if err == nil {
		t.Error("Delete(999) should return error")
	}
}

func TestUserWorkoutMovementRepository_DeleteByUserWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Add multiple movements
	for i := 0; i < 3; i++ {
		uwm := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement.ID,
			OrderIndex:    i,
		}
		uwmRepo.Create(uwm)
	}

	// Verify 3 movements
	movements, _ := uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(movements) != 3 {
		t.Fatalf("Expected 3 movements, got %d", len(movements))
	}

	// Delete all
	err = uwmRepo.DeleteByUserWorkoutID(userWorkout.ID)
	if err != nil {
		t.Fatalf("DeleteByUserWorkoutID() error = %v", err)
	}

	// Verify all deleted
	movements, _ = uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(movements) != 0 {
		t.Errorf("Expected 0 movements, got %d", len(movements))
	}
}

func TestUserWorkoutMovementRepository_CreateBatch(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Create batch
	var movements []*domain.UserWorkoutMovement
	for i := 0; i < 5; i++ {
		weight := float64(135 + i*20)
		reps := 5 + i
		movements = append(movements, &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement.ID,
			Weight:        &weight,
			Reps:          &reps,
			OrderIndex:    i,
		})
	}

	err = uwmRepo.CreateBatch(movements)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}

	// Verify all created with IDs
	for i, m := range movements {
		if m.ID == 0 {
			t.Errorf("Movement %d should have ID set", i)
		}
	}

	// Verify in database
	got, _ := uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(got) != 5 {
		t.Errorf("Expected 5 movements, got %d", len(got))
	}
}

func TestUserWorkoutMovementRepository_GetMaxWeightForMovement(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	// Create multiple workouts with different weights
	weights := []float64{225.0, 315.0, 275.0, 295.0}
	for i, w := range weights {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		weight := w
		uwm := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement.ID,
			Weight:        &weight,
			OrderIndex:    0,
		}
		uwmRepo.Create(uwm)
	}

	// Get max weight
	maxWeight, err := uwmRepo.GetMaxWeightForMovement(user.ID, movement.ID)
	if err != nil {
		t.Fatalf("GetMaxWeightForMovement() error = %v", err)
	}
	if maxWeight == nil {
		t.Fatal("GetMaxWeightForMovement() returned nil")
	}
	if *maxWeight != 315.0 {
		t.Errorf("Max weight = %v, want 315.0", *maxWeight)
	}

	// Test for user with no movements
	noMovements, err := uwmRepo.GetMaxWeightForMovement(999, movement.ID)
	if err != nil {
		t.Fatalf("GetMaxWeightForMovement() error = %v", err)
	}
	if noMovements != nil {
		t.Error("GetMaxWeightForMovement() should return nil for user with no movements")
	}
}

func TestUserWorkoutMovementRepository_UpdatePRFlag(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	weight := 315.0
	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkout.ID,
		MovementID:    movement.ID,
		Weight:        &weight,
		IsPR:          false,
		OrderIndex:    0,
	}
	uwmRepo.Create(uwm)

	// Update PR flag to true
	err = uwmRepo.UpdatePRFlag(uwm.ID, true)
	if err != nil {
		t.Fatalf("UpdatePRFlag() error = %v", err)
	}

	// Verify using GetByUserWorkoutID which includes is_pr field
	// Note: GetByID doesn't include is_pr in its SELECT (potential bug in production code)
	movements, _ := uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(movements) == 0 {
		t.Fatal("Expected at least one movement")
	}
	if !movements[0].IsPR {
		t.Error("IsPR should be true")
	}

	// Update PR flag to false
	err = uwmRepo.UpdatePRFlag(uwm.ID, false)
	if err != nil {
		t.Fatalf("UpdatePRFlag() error = %v", err)
	}

	// Verify
	movements, _ = uwmRepo.GetByUserWorkoutID(userWorkout.ID)
	if movements[0].IsPR {
		t.Error("IsPR should be false")
	}
}

func TestUserWorkoutMovementRepository_GetPRMovements(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	// Create workouts with some PRs
	for i := 0; i < 5; i++ {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		weight := float64(225 + i*25)
		uwm := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement.ID,
			Weight:        &weight,
			IsPR:          i%2 == 0, // Every other one is a PR
			OrderIndex:    0,
		}
		uwmRepo.Create(uwm)
	}

	// Get PR movements
	prs, err := uwmRepo.GetPRMovements(user.ID, 10)
	if err != nil {
		t.Fatalf("GetPRMovements() error = %v", err)
	}
	if len(prs) != 3 { // 0, 2, 4 are PRs
		t.Errorf("GetPRMovements() returned %d movements, want 3", len(prs))
	}

	// Verify all returned are PRs
	for _, pr := range prs {
		if !pr.IsPR {
			t.Error("All returned movements should be PRs")
		}
	}
}

func TestUserWorkoutMovementRepository_GetAllPerformancesForMovement(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	// Create multiple workouts with different weights and dates
	weights := []float64{225.0, 250.0, 275.0, 300.0, 315.0}
	for i, w := range weights {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -30+i*5), // Spread over time
		}
		userWorkoutRepo.Create(userWorkout)

		weight := w
		reps := 5
		uwm := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement.ID,
			Weight:        &weight,
			Reps:          &reps,
			IsPR:          i == 4, // Last one is PR
			OrderIndex:    0,
		}
		uwmRepo.Create(uwm)
	}

	// Get all performances (nil excludeWorkoutID means don't exclude any)
	performances, err := uwmRepo.GetAllPerformancesForMovement(user.ID, movement.ID, nil)
	if err != nil {
		t.Fatalf("GetAllPerformancesForMovement() error = %v", err)
	}
	if len(performances) != 5 {
		t.Errorf("GetAllPerformancesForMovement() returned %d performances, want 5", len(performances))
	}

	// Test for non-existent user
	performances, err = uwmRepo.GetAllPerformancesForMovement(999, movement.ID, nil)
	if err != nil {
		t.Fatalf("GetAllPerformancesForMovement(999) error = %v", err)
	}
	if len(performances) != 0 {
		t.Errorf("GetAllPerformancesForMovement(999) returned %d performances, want 0", len(performances))
	}

	// Test for non-existent movement
	performances, err = uwmRepo.GetAllPerformancesForMovement(user.ID, 999, nil)
	if err != nil {
		t.Fatalf("GetAllPerformancesForMovement(movement=999) error = %v", err)
	}
	if len(performances) != 0 {
		t.Errorf("GetAllPerformancesForMovement(movement=999) returned %d performances, want 0", len(performances))
	}
}

func TestUserWorkoutMovementRepository_GetByUserIDAndMovementID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwmRepo := NewUserWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	movement1 := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movement2 := &domain.Movement{Name: "Deadlift", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement1)
	movementRepo.Create(movement2)

	// Create workouts with both movements
	for i := 0; i < 3; i++ {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		// Add movement1
		weight1 := float64(225 + i*25)
		uwm1 := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement1.ID,
			Weight:        &weight1,
			OrderIndex:    0,
		}
		uwmRepo.Create(uwm1)

		// Add movement2
		weight2 := float64(315 + i*25)
		uwm2 := &domain.UserWorkoutMovement{
			UserWorkoutID: userWorkout.ID,
			MovementID:    movement2.ID,
			Weight:        &weight2,
			OrderIndex:    1,
		}
		uwmRepo.Create(uwm2)
	}

	// Get by user ID and movement1 ID (limit 100)
	performances, err := uwmRepo.GetByUserIDAndMovementID(user.ID, movement1.ID, 100)
	if err != nil {
		t.Fatalf("GetByUserIDAndMovementID() error = %v", err)
	}
	if len(performances) != 3 {
		t.Errorf("GetByUserIDAndMovementID(movement1) returned %d performances, want 3", len(performances))
	}

	// Verify all are for the right movement
	for _, p := range performances {
		if p.MovementID != movement1.ID {
			t.Errorf("Movement ID = %d, want %d", p.MovementID, movement1.ID)
		}
	}

	// Get by user ID and movement2 ID
	performances, err = uwmRepo.GetByUserIDAndMovementID(user.ID, movement2.ID, 100)
	if err != nil {
		t.Fatalf("GetByUserIDAndMovementID(movement2) error = %v", err)
	}
	if len(performances) != 3 {
		t.Errorf("GetByUserIDAndMovementID(movement2) returned %d performances, want 3", len(performances))
	}

	// Test for non-existent user
	performances, err = uwmRepo.GetByUserIDAndMovementID(999, movement1.ID, 100)
	if err != nil {
		t.Fatalf("GetByUserIDAndMovementID(999) error = %v", err)
	}
	if len(performances) != 0 {
		t.Errorf("GetByUserIDAndMovementID(999) returned %d performances, want 0", len(performances))
	}

	// Test with limit
	performances, err = uwmRepo.GetByUserIDAndMovementID(user.ID, movement1.ID, 2)
	if err != nil {
		t.Fatalf("GetByUserIDAndMovementID(limit=2) error = %v", err)
	}
	if len(performances) != 2 {
		t.Errorf("GetByUserIDAndMovementID(limit=2) returned %d performances, want 2", len(performances))
	}
}
