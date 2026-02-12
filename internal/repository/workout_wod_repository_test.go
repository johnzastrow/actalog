package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestWorkoutWODRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{
		Name:      "Fran",
		Type:      "Benchmark",
		ScoreType: "Time",
	}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Create workout-WOD association
	scoreValue := "2:45"
	division := "Rx"
	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		ScoreValue: &scoreValue,
		Division:   &division,
		IsPR:       true,
		OrderIndex: 0,
	}

	err = wwRepo.Create(ww)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if ww.ID == 0 {
		t.Error("Create() should set ID")
	}
	if ww.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestWorkoutWODRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time"}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	scoreValue := "2:30"
	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		ScoreValue: &scoreValue,
		IsPR:       true,
		OrderIndex: 0,
	}
	wwRepo.Create(ww)

	// Get by ID
	got, err := wwRepo.GetByID(ww.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.WorkoutID != workout.ID {
		t.Errorf("WorkoutID = %d, want %d", got.WorkoutID, workout.ID)
	}
	if got.ScoreValue == nil || *got.ScoreValue != "2:30" {
		t.Error("ScoreValue should be '2:30'")
	}
	if !got.IsPR {
		t.Error("IsPR should be true")
	}

	// Non-existent
	got, err = wwRepo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestWorkoutWODRepository_ListByWorkout(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wods := []*domain.WOD{
		{Name: "Fran", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Grace", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Helen", Type: "Benchmark", ScoreType: "Time"},
	}
	for _, w := range wods {
		wodRepo.Create(w)
	}

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Add WODs to workout
	for i, w := range wods {
		ww := &domain.WorkoutWOD{
			WorkoutID:  workout.ID,
			WODID:      w.ID,
			OrderIndex: i,
		}
		wwRepo.Create(ww)
	}

	// List by workout
	got, err := wwRepo.ListByWorkout(workout.ID)
	if err != nil {
		t.Fatalf("ListByWorkout() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("ListByWorkout() returned %d WODs, want 3", len(got))
	}

	// Verify order
	for i, ww := range got {
		if ww.OrderIndex != i {
			t.Errorf("WOD %d has OrderIndex %d, want %d", i, ww.OrderIndex, i)
		}
	}
}

func TestWorkoutWODRepository_ListByWorkoutWithDetails(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{
		Name:        "Fran",
		Type:        "Benchmark",
		Regime:      "21-15-9",
		ScoreType:   "Time",
		Description: "Thrusters and Pull-ups",
	}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		OrderIndex: 0,
	}
	wwRepo.Create(ww)

	// List with details
	got, err := wwRepo.ListByWorkoutWithDetails(workout.ID)
	if err != nil {
		t.Fatalf("ListByWorkoutWithDetails() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Expected 1 WOD, got %d", len(got))
	}

	// Verify WOD details are populated
	if got[0].WODName != "Fran" {
		t.Errorf("WODName = %v, want 'Fran'", got[0].WODName)
	}
	if got[0].WODType != "Benchmark" {
		t.Errorf("WODType = %v, want 'Benchmark'", got[0].WODType)
	}
	if got[0].WODRegime != "21-15-9" {
		t.Errorf("WODRegime = %v, want '21-15-9'", got[0].WODRegime)
	}
}

func TestWorkoutWODRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time"}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	scoreValue := "3:00"
	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		ScoreValue: &scoreValue,
		IsPR:       false,
		OrderIndex: 0,
	}
	wwRepo.Create(ww)

	// Update
	newScore := "2:45"
	newDivision := "Rx"
	ww.ScoreValue = &newScore
	ww.Division = &newDivision
	ww.IsPR = true
	err = wwRepo.Update(ww)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, _ := wwRepo.GetByID(ww.ID)
	if got.ScoreValue == nil || *got.ScoreValue != "2:45" {
		t.Error("ScoreValue should be updated to '2:45'")
	}
	if got.Division == nil || *got.Division != "Rx" {
		t.Error("Division should be 'Rx'")
	}
	if !got.IsPR {
		t.Error("IsPR should be true")
	}
}

func TestWorkoutWODRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time"}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		OrderIndex: 0,
	}
	wwRepo.Create(ww)

	// Delete
	err = wwRepo.Delete(ww.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := wwRepo.GetByID(ww.ID)
	if got != nil {
		t.Error("WorkoutWOD should be deleted")
	}

	// Delete non-existent
	err = wwRepo.Delete(999)
	if err == nil {
		t.Error("Delete(999) should return error")
	}
}

func TestWorkoutWODRepository_GetByWODID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time"}
	wodRepo.Create(wod)

	// Create multiple workouts using the same WOD
	for i := 0; i < 3; i++ {
		workout := &domain.Workout{Name: "Workout " + string(rune('A'+i)), CreatedBy: &user.ID}
		workoutRepo.Create(workout)

		ww := &domain.WorkoutWOD{
			WorkoutID:  workout.ID,
			WODID:      wod.ID,
			OrderIndex: 0,
		}
		wwRepo.Create(ww)
	}

	// Get by WOD ID
	got, err := wwRepo.GetByWODID(wod.ID)
	if err != nil {
		t.Fatalf("GetByWODID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByWODID() returned %d entries, want 3", len(got))
	}
}

func TestWorkoutWODRepository_DeleteByWorkout(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wods := []*domain.WOD{
		{Name: "Fran", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Grace", Type: "Benchmark", ScoreType: "Time"},
	}
	for _, w := range wods {
		wodRepo.Create(w)
	}

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Add WODs
	for i, w := range wods {
		ww := &domain.WorkoutWOD{
			WorkoutID:  workout.ID,
			WODID:      w.ID,
			OrderIndex: i,
		}
		wwRepo.Create(ww)
	}

	// Verify we have 2 WODs
	list, _ := wwRepo.ListByWorkout(workout.ID)
	if len(list) != 2 {
		t.Fatalf("Expected 2 WODs, got %d", len(list))
	}

	// Delete all
	err = wwRepo.DeleteByWorkout(workout.ID)
	if err != nil {
		t.Fatalf("DeleteByWorkout() error = %v", err)
	}

	// Verify all deleted
	list, _ = wwRepo.ListByWorkout(workout.ID)
	if len(list) != 0 {
		t.Errorf("Expected 0 WODs, got %d", len(list))
	}
}

func TestWorkoutWODRepository_BatchCreate(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wods := []*domain.WOD{
		{Name: "Fran", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Grace", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Helen", Type: "Benchmark", ScoreType: "Time"},
	}
	var wodIDs []int64
	for _, w := range wods {
		wodRepo.Create(w)
		wodIDs = append(wodIDs, w.ID)
	}

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Batch create
	err = wwRepo.BatchCreate(workout.ID, wodIDs)
	if err != nil {
		t.Fatalf("BatchCreate() error = %v", err)
	}

	// Verify all created
	list, _ := wwRepo.ListByWorkout(workout.ID)
	if len(list) != 3 {
		t.Errorf("Expected 3 WODs, got %d", len(list))
	}

	// Verify ordering
	for i, ww := range list {
		if ww.OrderIndex != i {
			t.Errorf("WOD %d has OrderIndex %d, want %d", i, ww.OrderIndex, i)
		}
	}
}

func TestWorkoutWODRepository_Reorder(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wods := []*domain.WOD{
		{Name: "Fran", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Grace", Type: "Benchmark", ScoreType: "Time"},
		{Name: "Helen", Type: "Benchmark", ScoreType: "Time"},
	}
	for _, w := range wods {
		wodRepo.Create(w)
	}

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	// Add WODs in original order
	for i, w := range wods {
		ww := &domain.WorkoutWOD{
			WorkoutID:  workout.ID,
			WODID:      w.ID,
			OrderIndex: i,
		}
		wwRepo.Create(ww)
	}

	// Reorder: Helen, Fran, Grace
	newOrder := []int64{wods[2].ID, wods[0].ID, wods[1].ID}
	err = wwRepo.Reorder(workout.ID, newOrder)
	if err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}

	// Verify new order
	list, _ := wwRepo.ListByWorkout(workout.ID)
	if list[0].WODID != wods[2].ID {
		t.Error("First WOD should be Helen")
	}
	if list[1].WODID != wods[0].ID {
		t.Error("Second WOD should be Fran")
	}
	if list[2].WODID != wods[1].ID {
		t.Error("Third WOD should be Grace")
	}
}

func TestWorkoutWODRepository_TogglePR(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	wwRepo := NewWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "athlete"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time"}
	wodRepo.Create(wod)

	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &user.ID}
	workoutRepo.Create(workout)

	ww := &domain.WorkoutWOD{
		WorkoutID:  workout.ID,
		WODID:      wod.ID,
		IsPR:       false,
		OrderIndex: 0,
	}
	wwRepo.Create(ww)

	// Toggle PR on
	err = wwRepo.TogglePR(ww.ID)
	if err != nil {
		t.Fatalf("TogglePR() error = %v", err)
	}

	got, _ := wwRepo.GetByID(ww.ID)
	if !got.IsPR {
		t.Error("IsPR should be true after toggle")
	}

	// Toggle PR off
	err = wwRepo.TogglePR(ww.ID)
	if err != nil {
		t.Fatalf("TogglePR() error = %v", err)
	}

	got, _ = wwRepo.GetByID(ww.ID)
	if got.IsPR {
		t.Error("IsPR should be false after second toggle")
	}
}
