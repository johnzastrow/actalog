package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestWorkoutMovementRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Create user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
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

	// Create workout
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &user.ID,
	}
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Create workout movement
	weight := 225.0
	sets := 5
	reps := 5
	wm := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: movement.ID,
		Weight:     &weight,
		Sets:       &sets,
		Reps:       &reps,
		IsRx:       true,
		IsPR:       false,
		OrderIndex: 0,
	}

	err = wmRepo.Create(wm)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if wm.ID == 0 {
		t.Error("Create() should set ID")
	}
	if wm.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestWorkoutMovementRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	weight := 315.0
	reps := 3
	wm := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: movement.ID,
		Weight:     &weight,
		Reps:       &reps,
		IsPR:       true,
		OrderIndex: 0,
	}
	wmRepo.Create(wm)

	// Get by ID
	got, err := wmRepo.GetByID(wm.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.WorkoutID != workout.ID {
		t.Errorf("WorkoutID = %d, want %d", got.WorkoutID, workout.ID)
	}
	if got.Weight == nil || *got.Weight != 315.0 {
		t.Error("Weight should be 315.0")
	}
	if !got.IsPR {
		t.Error("IsPR should be true")
	}

	// Non-existent
	got, err = wmRepo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestWorkoutMovementRepository_GetByWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	movements := []*domain.Movement{
		{Name: "Back Squat", Type: domain.MovementTypeWeightlifting},
		{Name: "Bench Press", Type: domain.MovementTypeWeightlifting},
		{Name: "Deadlift", Type: domain.MovementTypeWeightlifting},
	}
	for _, m := range movements {
		movementRepo.Create(m)
	}

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Add movements to workout
	for i, m := range movements {
		weight := float64(135 + i*50)
		wm := &domain.WorkoutMovement{
			WorkoutID:  workout.ID,
			MovementID: m.ID,
			Weight:     &weight,
			OrderIndex: i,
		}
		wmRepo.Create(wm)
	}

	// Get by workout ID
	got, err := wmRepo.GetByWorkoutID(workout.ID)
	if err != nil {
		t.Fatalf("GetByWorkoutID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByWorkoutID() returned %d movements, want 3", len(got))
	}

	// Verify order
	for i, wm := range got {
		if wm.OrderIndex != i {
			t.Errorf("Movement %d has OrderIndex %d, want %d", i, wm.OrderIndex, i)
		}
		if wm.Movement == nil {
			t.Error("Movement should be populated from JOIN")
		}
	}
}

func TestWorkoutMovementRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	weight := 225.0
	wm := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: movement.ID,
		Weight:     &weight,
		IsPR:       false,
		OrderIndex: 0,
	}
	wmRepo.Create(wm)

	// Update
	newWeight := 275.0
	wm.Weight = &newWeight
	wm.IsPR = true
	err = wmRepo.Update(wm)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, _ := wmRepo.GetByID(wm.ID)
	if got.Weight == nil || *got.Weight != 275.0 {
		t.Error("Weight should be updated to 275.0")
	}
	if !got.IsPR {
		t.Error("IsPR should be true")
	}
}

func TestWorkoutMovementRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	wm := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: movement.ID,
		OrderIndex: 0,
	}
	wmRepo.Create(wm)

	// Delete
	err = wmRepo.Delete(wm.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := wmRepo.GetByID(wm.ID)
	if got != nil {
		t.Error("Movement should be deleted")
	}

	// Delete non-existent
	err = wmRepo.Delete(999)
	if err == nil {
		t.Error("Delete(999) should return error")
	}
}

func TestWorkoutMovementRepository_DeleteByWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	movement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	movementRepo.Create(movement)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Add multiple movements
	for i := 0; i < 3; i++ {
		wm := &domain.WorkoutMovement{
			WorkoutID:  workout.ID,
			MovementID: movement.ID,
			OrderIndex: i,
		}
		wmRepo.Create(wm)
	}

	// Verify 3 movements
	movements, _ := wmRepo.GetByWorkoutID(workout.ID)
	if len(movements) != 3 {
		t.Fatalf("Expected 3 movements, got %d", len(movements))
	}

	// Delete all
	err = wmRepo.DeleteByWorkoutID(workout.ID)
	if err != nil {
		t.Fatalf("DeleteByWorkoutID() error = %v", err)
	}

	// Verify all deleted
	movements, _ = wmRepo.GetByWorkoutID(workout.ID)
	if len(movements) != 0 {
		t.Errorf("Expected 0 movements, got %d", len(movements))
	}
}

func TestWorkoutMovementRepository_DifferentMeasurements(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	movementRepo := NewMovementRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wmRepo := NewWorkoutMovementRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	// Different movement types
	squat := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting}
	run := &domain.Movement{Name: "Run", Type: domain.MovementTypeCardio}
	pullup := &domain.Movement{Name: "Pull-ups", Type: domain.MovementTypeGymnastics}
	movementRepo.Create(squat)
	movementRepo.Create(run)
	movementRepo.Create(pullup)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Weight-based movement
	weight := 315.0
	sets := 5
	reps := 3
	wmWeight := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: squat.ID,
		Weight:     &weight,
		Sets:       &sets,
		Reps:       &reps,
		OrderIndex: 0,
	}
	wmRepo.Create(wmWeight)

	// Time-based movement (cardio)
	timeSeconds := 300 // 5 minutes
	distance := 1.0    // 1 mile
	wmCardio := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: run.ID,
		Time:       &timeSeconds,
		Distance:   &distance,
		OrderIndex: 1,
	}
	wmRepo.Create(wmCardio)

	// Rep-based movement (gymnastics)
	gymReps := 10
	wmGym := &domain.WorkoutMovement{
		WorkoutID:  workout.ID,
		MovementID: pullup.ID,
		Reps:       &gymReps,
		OrderIndex: 2,
	}
	wmRepo.Create(wmGym)

	// Verify all retrieved correctly
	movements, err := wmRepo.GetByWorkoutID(workout.ID)
	if err != nil {
		t.Fatalf("GetByWorkoutID() error = %v", err)
	}
	if len(movements) != 3 {
		t.Errorf("Expected 3 movements, got %d", len(movements))
	}

	// Check each has its proper values
	for _, m := range movements {
		switch m.MovementID {
		case squat.ID:
			if m.Weight == nil || *m.Weight != 315.0 {
				t.Error("Squat should have weight 315")
			}
		case run.ID:
			if m.Time == nil || *m.Time != 300 {
				t.Error("Run should have time 300")
			}
			if m.Distance == nil || *m.Distance != 1.0 {
				t.Error("Run should have distance 1.0")
			}
		case pullup.ID:
			if m.Reps == nil || *m.Reps != 10 {
				t.Error("Pull-ups should have 10 reps")
			}
		}
	}
}
