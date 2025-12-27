package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserWorkoutWODRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Create user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create WOD
	wod := &domain.WOD{
		Name:        "Fran",
		Type:        "Benchmark",
		ScoreType:   "Time",
		Description: "21-15-9 Thrusters and Pull-ups",
		IsStandard:  true,
	}
	if err := wodRepo.Create(wod); err != nil {
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create user workout
	userWorkout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now(),
	}
	if err := userWorkoutRepo.Create(userWorkout); err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Create user workout WOD
	scoreType := "Time"
	scoreValue := "3:25"
	timeSeconds := 205
	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkout.ID,
		WODID:         wod.ID,
		ScoreType:     &scoreType,
		ScoreValue:    &scoreValue,
		TimeSeconds:   &timeSeconds,
		IsPR:          true,
		OrderIndex:    0,
	}

	err = uwwRepo.Create(uww)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if uww.ID == 0 {
		t.Error("Create() should set ID")
	}
	if uww.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestUserWorkoutWODRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	scoreType := "Time"
	scoreValue := "3:25"
	timeSeconds := 205
	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkout.ID,
		WODID:         wod.ID,
		ScoreType:     &scoreType,
		ScoreValue:    &scoreValue,
		TimeSeconds:   &timeSeconds,
		IsPR:          true,
		OrderIndex:    0,
	}
	uwwRepo.Create(uww)

	// Get by ID
	got, err := uwwRepo.GetByID(uww.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.UserWorkoutID != userWorkout.ID {
		t.Errorf("UserWorkoutID = %d, want %d", got.UserWorkoutID, userWorkout.ID)
	}
	if got.TimeSeconds == nil || *got.TimeSeconds != 205 {
		t.Error("TimeSeconds should be 205")
	}

	// Non-existent
	got, err = uwwRepo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestUserWorkoutWODRepository_GetByUserWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wods := []*domain.WOD{
		{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true},
		{Name: "Grace", Type: "Benchmark", ScoreType: "Time", IsStandard: true},
		{Name: "Cindy", Type: "Benchmark", ScoreType: "Rounds", IsStandard: true},
	}
	for _, w := range wods {
		wodRepo.Create(w)
	}

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Add WODs to user workout
	for i, w := range wods {
		scoreType := "Time"
		uww := &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         w.ID,
			ScoreType:     &scoreType,
			OrderIndex:    i,
		}
		uwwRepo.Create(uww)
	}

	// Get by user workout ID
	got, err := uwwRepo.GetByUserWorkoutID(userWorkout.ID)
	if err != nil {
		t.Fatalf("GetByUserWorkoutID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByUserWorkoutID() returned %d WODs, want 3", len(got))
	}

	// Verify order and that WOD is populated
	for i, uww := range got {
		if uww.OrderIndex != i {
			t.Errorf("WOD %d has OrderIndex %d, want %d", i, uww.OrderIndex, i)
		}
		if uww.WOD == nil {
			t.Error("WOD should be populated from JOIN")
		}
	}
}

func TestUserWorkoutWODRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	scoreType := "Time"
	scoreValue := "4:00"
	timeSeconds := 240
	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkout.ID,
		WODID:         wod.ID,
		ScoreType:     &scoreType,
		ScoreValue:    &scoreValue,
		TimeSeconds:   &timeSeconds,
		OrderIndex:    0,
	}
	uwwRepo.Create(uww)

	// Update
	newScoreValue := "3:30"
	newTimeSeconds := 210
	uww.ScoreValue = &newScoreValue
	uww.TimeSeconds = &newTimeSeconds
	err = uwwRepo.Update(uww)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, _ := uwwRepo.GetByID(uww.ID)
	if got.ScoreValue == nil || *got.ScoreValue != "3:30" {
		t.Error("ScoreValue should be updated to 3:30")
	}
	if got.TimeSeconds == nil || *got.TimeSeconds != 210 {
		t.Error("TimeSeconds should be updated to 210")
	}
}

func TestUserWorkoutWODRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkout.ID,
		WODID:         wod.ID,
		OrderIndex:    0,
	}
	uwwRepo.Create(uww)

	// Delete
	err = uwwRepo.Delete(uww.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := uwwRepo.GetByID(uww.ID)
	if got != nil {
		t.Error("WOD should be deleted")
	}

	// Delete non-existent
	err = uwwRepo.Delete(999)
	if err == nil {
		t.Error("Delete(999) should return error")
	}
}

func TestUserWorkoutWODRepository_DeleteByUserWorkoutID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Add multiple WODs
	for i := 0; i < 3; i++ {
		uww := &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         wod.ID,
			OrderIndex:    i,
		}
		uwwRepo.Create(uww)
	}

	// Verify 3 WODs
	wods, _ := uwwRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(wods) != 3 {
		t.Fatalf("Expected 3 WODs, got %d", len(wods))
	}

	// Delete all
	err = uwwRepo.DeleteByUserWorkoutID(userWorkout.ID)
	if err != nil {
		t.Fatalf("DeleteByUserWorkoutID() error = %v", err)
	}

	// Verify all deleted
	wods, _ = uwwRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(wods) != 0 {
		t.Errorf("Expected 0 WODs, got %d", len(wods))
	}
}

func TestUserWorkoutWODRepository_CreateBatch(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	// Create batch
	var wods []*domain.UserWorkoutWOD
	for i := 0; i < 5; i++ {
		scoreType := "Time"
		timeSeconds := 200 + i*10
		wods = append(wods, &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         wod.ID,
			ScoreType:     &scoreType,
			TimeSeconds:   &timeSeconds,
			OrderIndex:    i,
		})
	}

	err = uwwRepo.CreateBatch(wods)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}

	// Verify all created with IDs
	for i, w := range wods {
		if w.ID == 0 {
			t.Errorf("WOD %d should have ID set", i)
		}
	}

	// Verify in database
	got, _ := uwwRepo.GetByUserWorkoutID(userWorkout.ID)
	if len(got) != 5 {
		t.Errorf("Expected 5 WODs, got %d", len(got))
	}
}

func TestUserWorkoutWODRepository_GetBestTimeForWOD(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	// Create multiple workouts with different times
	times := []int{240, 205, 220, 230}
	for i, t := range times {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		timeSeconds := t
		uww := &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         wod.ID,
			TimeSeconds:   &timeSeconds,
			OrderIndex:    0,
		}
		uwwRepo.Create(uww)
	}

	// Get best time (lowest)
	bestTime, err := uwwRepo.GetBestTimeForWOD(user.ID, wod.ID)
	if err != nil {
		t.Fatalf("GetBestTimeForWOD() error = %v", err)
	}
	if bestTime == nil {
		t.Fatal("GetBestTimeForWOD() returned nil")
	}
	if *bestTime != 205 {
		t.Errorf("Best time = %v, want 205", *bestTime)
	}

	// Test for user with no WODs
	noTime, err := uwwRepo.GetBestTimeForWOD(999, wod.ID)
	if err != nil {
		t.Fatalf("GetBestTimeForWOD() error = %v", err)
	}
	if noTime != nil {
		t.Error("GetBestTimeForWOD() should return nil for user with no WODs")
	}
}

func TestUserWorkoutWODRepository_GetBestRoundsRepsForWOD(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Cindy", Type: "Benchmark", ScoreType: "Rounds", IsStandard: true}
	wodRepo.Create(wod)

	// Create workouts with different rounds/reps
	performances := []struct {
		rounds int
		reps   int
	}{
		{15, 10},
		{18, 5},
		{18, 12}, // Best - highest rounds with highest reps for tied rounds
		{16, 8},
	}

	for i, p := range performances {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		rounds := p.rounds
		reps := p.reps
		uww := &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         wod.ID,
			Rounds:        &rounds,
			Reps:          &reps,
			OrderIndex:    0,
		}
		uwwRepo.Create(uww)
	}

	// Get best rounds+reps
	rounds, reps, err := uwwRepo.GetBestRoundsRepsForWOD(user.ID, wod.ID)
	if err != nil {
		t.Fatalf("GetBestRoundsRepsForWOD() error = %v", err)
	}
	if rounds == nil || *rounds != 18 {
		t.Errorf("Best rounds = %v, want 18", rounds)
	}
	if reps == nil || *reps != 12 {
		t.Errorf("Best reps = %v, want 12", reps)
	}
}

func TestUserWorkoutWODRepository_UpdatePRFlag(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	userWorkout := &domain.UserWorkout{UserID: user.ID, WorkoutDate: time.Now()}
	userWorkoutRepo.Create(userWorkout)

	timeSeconds := 205
	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkout.ID,
		WODID:         wod.ID,
		TimeSeconds:   &timeSeconds,
		IsPR:          false,
		OrderIndex:    0,
	}
	uwwRepo.Create(uww)

	// Update PR flag to true
	err = uwwRepo.UpdatePRFlag(uww.ID, true)
	if err != nil {
		t.Fatalf("UpdatePRFlag() error = %v", err)
	}

	// Verify using direct SQL since GetByID doesn't include is_pr in its SELECT
	// Note: This is a workaround for a potential bug in the production GetByID method
	var isPR int
	err = db.QueryRow("SELECT is_pr FROM user_workout_wods WHERE id = ?", uww.ID).Scan(&isPR)
	if err != nil {
		t.Fatalf("Failed to query is_pr: %v", err)
	}
	if isPR != 1 {
		t.Error("IsPR should be true (1)")
	}

	// Update PR flag to false
	err = uwwRepo.UpdatePRFlag(uww.ID, false)
	if err != nil {
		t.Fatalf("UpdatePRFlag() error = %v", err)
	}

	// Verify
	err = db.QueryRow("SELECT is_pr FROM user_workout_wods WHERE id = ?", uww.ID).Scan(&isPR)
	if err != nil {
		t.Fatalf("Failed to query is_pr: %v", err)
	}
	if isPR != 0 {
		t.Error("IsPR should be false (0)")
	}
}

func TestUserWorkoutWODRepository_GetPRWODs(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	wodRepo := NewWODRepository(db)
	userWorkoutRepo := NewUserWorkoutRepository(db)
	uwwRepo := NewUserWorkoutWODRepository(db)

	// Setup
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	userRepo.Create(user)

	wod := &domain.WOD{Name: "Fran", Type: "Benchmark", ScoreType: "Time", IsStandard: true}
	wodRepo.Create(wod)

	// Create workouts with some PRs
	for i := 0; i < 5; i++ {
		userWorkout := &domain.UserWorkout{
			UserID:      user.ID,
			WorkoutDate: time.Now().AddDate(0, 0, -i),
		}
		userWorkoutRepo.Create(userWorkout)

		timeSeconds := 200 + i*10
		uww := &domain.UserWorkoutWOD{
			UserWorkoutID: userWorkout.ID,
			WODID:         wod.ID,
			TimeSeconds:   &timeSeconds,
			IsPR:          i%2 == 0, // Every other one is a PR
			OrderIndex:    0,
		}
		uwwRepo.Create(uww)
	}

	// Get PR WODs
	prs, err := uwwRepo.GetPRWODs(user.ID, 10)
	if err != nil {
		t.Fatalf("GetPRWODs() error = %v", err)
	}
	if len(prs) != 3 { // 0, 2, 4 are PRs
		t.Errorf("GetPRWODs() returned %d WODs, want 3", len(prs))
	}

	// Verify all returned are PRs
	for _, pr := range prs {
		if !pr.IsPR {
			t.Error("All returned WODs should be PRs")
		}
	}
}
