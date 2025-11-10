package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Mock repositories for testing
type mockWorkoutRepo struct {
	workouts map[int64]*domain.Workout
}

func (m *mockWorkoutRepo) Create(workout *domain.Workout) error {
	workout.ID = int64(len(m.workouts) + 1)
	workout.CreatedAt = time.Now()
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *mockWorkoutRepo) GetByID(id int64) (*domain.Workout, error) {
	workout, ok := m.workouts[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return workout, nil
}

func (m *mockWorkoutRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.Workout, error) {
	var workouts []*domain.Workout
	for _, w := range m.workouts {
		if w.UserID == userID {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}

func (m *mockWorkoutRepo) Update(workout *domain.Workout) error {
	if _, ok := m.workouts[workout.ID]; !ok {
		return sql.ErrNoRows
	}
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *mockWorkoutRepo) Delete(id int64) error {
	if _, ok := m.workouts[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.workouts, id)
	return nil
}

func (m *mockWorkoutRepo) GetByDate(userID int64, date time.Time) ([]*domain.Workout, error) {
	return nil, nil
}

type mockWorkoutMovementRepo struct {
	movements map[int64]*domain.WorkoutMovement
	nextID    int64
}

func (m *mockWorkoutMovementRepo) Create(movement *domain.WorkoutMovement) error {
	m.nextID++
	movement.ID = m.nextID
	movement.CreatedAt = time.Now()
	movement.UpdatedAt = time.Now()
	m.movements[movement.ID] = movement
	return nil
}

func (m *mockWorkoutMovementRepo) GetByID(id int64) (*domain.WorkoutMovement, error) {
	movement, ok := m.movements[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return movement, nil
}

func (m *mockWorkoutMovementRepo) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	var movements []*domain.WorkoutMovement
	for _, mv := range m.movements {
		if mv.WorkoutID == workoutID {
			movements = append(movements, mv)
		}
	}
	return movements, nil
}

func (m *mockWorkoutMovementRepo) Update(movement *domain.WorkoutMovement) error {
	if _, ok := m.movements[movement.ID]; !ok {
		return sql.ErrNoRows
	}
	movement.UpdatedAt = time.Now()
	m.movements[movement.ID] = movement
	return nil
}

func (m *mockWorkoutMovementRepo) Delete(id int64) error {
	if _, ok := m.movements[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.movements, id)
	return nil
}

func (m *mockWorkoutMovementRepo) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	// Group by movement_id and find max values
	records := make(map[int64]*domain.PersonalRecord)

	for _, mv := range m.movements {
		// Need to verify this movement belongs to user's workout
		// For simplicity in mock, we'll assume it does

		pr, exists := records[mv.MovementID]
		if !exists {
			pr = &domain.PersonalRecord{
				MovementID:   mv.MovementID,
				MovementName: "Test Movement",
				WorkoutID:    mv.WorkoutID,
				WorkoutDate:  time.Now(),
			}
			records[mv.MovementID] = pr
		}

		if mv.Weight != nil && (pr.MaxWeight == nil || *mv.Weight > *pr.MaxWeight) {
			pr.MaxWeight = mv.Weight
		}
		if mv.Reps != nil && (pr.MaxReps == nil || *mv.Reps > *pr.MaxReps) {
			pr.MaxReps = mv.Reps
		}
		if mv.Time != nil && (pr.BestTime == nil || *mv.Time < *pr.BestTime) {
			pr.BestTime = mv.Time
		}
	}

	var result []*domain.PersonalRecord
	for _, pr := range records {
		result = append(result, pr)
	}
	return result, nil
}

func (m *mockWorkoutMovementRepo) GetPRMovements(userID int64, limit int) ([]*domain.PRMovement, error) {
	var prMovements []*domain.PRMovement
	count := 0

	for _, mv := range m.movements {
		if mv.IsPR && count < limit {
			prMovements = append(prMovements, &domain.PRMovement{
				ID:           mv.ID,
				WorkoutID:    mv.WorkoutID,
				MovementID:   mv.MovementID,
				MovementName: "Test Movement",
				Weight:       mv.Weight,
				Reps:         mv.Reps,
				Time:         mv.Time,
				IsPR:         mv.IsPR,
				WorkoutDate:  time.Now(),
			})
			count++
		}
	}
	return prMovements, nil
}

func (m *mockWorkoutMovementRepo) TogglePRFlag(movementID int64) error {
	movement, ok := m.movements[movementID]
	if !ok {
		return sql.ErrNoRows
	}
	movement.IsPR = !movement.IsPR
	return nil
}

func (m *mockWorkoutMovementRepo) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	var maxWeight *float64

	for _, mv := range m.movements {
		// Skip if not for this movement
		if mv.MovementID != movementID {
			continue
		}

		// Check if this movement has weight and it's a new max
		if mv.Weight != nil {
			if maxWeight == nil || *mv.Weight > *maxWeight {
				maxWeight = mv.Weight
			}
		}
	}

	return maxWeight, nil
}

type mockMovementRepo struct {
	movements map[int64]*domain.Movement
}

func (m *mockMovementRepo) Create(movement *domain.Movement) error {
	return nil
}

func (m *mockMovementRepo) GetByID(id int64) (*domain.Movement, error) {
	return nil, nil
}

func (m *mockMovementRepo) GetByName(name string) (*domain.Movement, error) {
	return nil, nil
}

func (m *mockMovementRepo) ListStandard() ([]*domain.Movement, error) {
	return nil, nil
}

func (m *mockMovementRepo) Search(query string) ([]*domain.Movement, error) {
	return nil, nil
}

func (m *mockMovementRepo) ListByUser(userID int64) ([]*domain.Movement, error) {
	return nil, nil
}

func (m *mockMovementRepo) Update(movement *domain.Movement) error {
	return nil
}

func (m *mockMovementRepo) Delete(id int64) error {
	return nil
}

// Helper function to create test service
func newTestWorkoutService() *WorkoutService {
	return NewWorkoutService(
		&mockWorkoutRepo{workouts: make(map[int64]*domain.Workout)},
		&mockWorkoutMovementRepo{movements: make(map[int64]*domain.WorkoutMovement), nextID: 0},
		&mockMovementRepo{movements: make(map[int64]*domain.Movement)},
	)
}

// Test PR Detection
func TestDetectAndFlagPRs(t *testing.T) {
	service := newTestWorkoutService()
	userID := int64(1)
	workoutID := int64(1)
	movementID := int64(1)

	tests := []struct {
		name           string
		existingWeight *float64
		newWeight      *float64
		shouldBePR     bool
	}{
		{
			name:           "First workout with movement should be PR",
			existingWeight: nil,
			newWeight:      floatPtr(135.0),
			shouldBePR:     true,
		},
		{
			name:           "Heavier weight should be PR",
			existingWeight: floatPtr(135.0),
			newWeight:      floatPtr(185.0),
			shouldBePR:     true,
		},
		{
			name:           "Lighter weight should not be PR",
			existingWeight: floatPtr(185.0),
			newWeight:      floatPtr(135.0),
			shouldBePR:     false,
		},
		{
			name:           "Equal weight should not be PR",
			existingWeight: floatPtr(185.0),
			newWeight:      floatPtr(185.0),
			shouldBePR:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset service state
			service = newTestWorkoutService()

			// Add existing movement if specified
			if tt.existingWeight != nil {
				existingMv := &domain.WorkoutMovement{
					WorkoutID:  workoutID,
					MovementID: movementID,
					Weight:     tt.existingWeight,
				}
				service.workoutMovementRepo.Create(existingMv)
			}

			// Create new movement
			newMv := &domain.WorkoutMovement{
				WorkoutID:  workoutID,
				MovementID: movementID,
				Weight:     tt.newWeight,
			}
			service.workoutMovementRepo.Create(newMv)

			// Detect PRs
			err := service.DetectAndFlagPRs(userID, workoutID)
			if err != nil {
				t.Fatalf("DetectAndFlagPRs failed: %v", err)
			}

			// Verify PR flag
			movements, _ := service.workoutMovementRepo.GetByWorkoutID(workoutID)
			if len(movements) == 0 {
				t.Fatal("No movements found")
			}

			// Find the new movement (last one created)
			lastMovement := movements[len(movements)-1]
			if lastMovement.IsPR != tt.shouldBePR {
				t.Errorf("Expected IsPR=%v, got %v", tt.shouldBePR, lastMovement.IsPR)
			}
		})
	}
}

// Test Get Personal Records
func TestGetPersonalRecords(t *testing.T) {
	service := newTestWorkoutService()
	userID := int64(1)

	// Add test data
	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  1,
		MovementID: 1,
		Weight:     floatPtr(135.0),
		Reps:       intPtr(5),
	})

	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  2,
		MovementID: 1,
		Weight:     floatPtr(185.0), // PR weight
		Reps:       intPtr(3),
	})

	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  3,
		MovementID: 1,
		Weight:     floatPtr(155.0),
		Reps:       intPtr(8), // PR reps
	})

	// Get PRs
	records, err := service.GetPersonalRecords(userID)
	if err != nil {
		t.Fatalf("GetPersonalRecords failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("Expected records, got none")
	}

	// Verify max weight
	if records[0].MaxWeight == nil || *records[0].MaxWeight != 185.0 {
		t.Errorf("Expected max weight 185.0, got %v", records[0].MaxWeight)
	}

	// Verify max reps
	if records[0].MaxReps == nil || *records[0].MaxReps != 8 {
		t.Errorf("Expected max reps 8, got %v", records[0].MaxReps)
	}
}

// Test Toggle PR Flag
func TestTogglePRFlag(t *testing.T) {
	service := newTestWorkoutService()
	userID := int64(1)

	// Create movement
	mv := &domain.WorkoutMovement{
		WorkoutID:  1,
		MovementID: 1,
		Weight:     floatPtr(135.0),
		IsPR:       false,
	}
	service.workoutMovementRepo.Create(mv)

	// Toggle to true
	err := service.TogglePRFlag(mv.ID, userID)
	if err != nil {
		t.Fatalf("TogglePRFlag failed: %v", err)
	}

	// Verify it's now true
	updated, _ := service.workoutMovementRepo.GetByID(mv.ID)
	if !updated.IsPR {
		t.Error("Expected IsPR=true after toggle")
	}

	// Toggle back to false
	err = service.TogglePRFlag(mv.ID, userID)
	if err != nil {
		t.Fatalf("TogglePRFlag failed: %v", err)
	}

	// Verify it's now false
	updated, _ = service.workoutMovementRepo.GetByID(mv.ID)
	if updated.IsPR {
		t.Error("Expected IsPR=false after second toggle")
	}
}

// Test Get PR Movements
func TestGetPRMovements(t *testing.T) {
	service := newTestWorkoutService()
	userID := int64(1)

	// Create multiple movements, some PRs
	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  1,
		MovementID: 1,
		Weight:     floatPtr(135.0),
		IsPR:       false,
	})

	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  2,
		MovementID: 2,
		Weight:     floatPtr(185.0),
		IsPR:       true,
	})

	service.workoutMovementRepo.Create(&domain.WorkoutMovement{
		WorkoutID:  3,
		MovementID: 3,
		Weight:     floatPtr(225.0),
		IsPR:       true,
	})

	// Get PR movements
	prMovements, err := service.GetPRMovements(userID, 10)
	if err != nil {
		t.Fatalf("GetPRMovements failed: %v", err)
	}

	// Should get 2 PR movements
	if len(prMovements) != 2 {
		t.Errorf("Expected 2 PR movements, got %d", len(prMovements))
	}

	// Verify they're all marked as PR
	for _, mv := range prMovements {
		if !mv.IsPR {
			t.Error("Expected all returned movements to be PRs")
		}
	}
}

// Test Limit on PR Movements
func TestGetPRMovementsLimit(t *testing.T) {
	service := newTestWorkoutService()
	userID := int64(1)

	// Create 10 PR movements
	for i := 0; i < 10; i++ {
		service.workoutMovementRepo.Create(&domain.WorkoutMovement{
			WorkoutID:  int64(i + 1),
			MovementID: 1,
			Weight:     floatPtr(float64(100 + i*10)),
			IsPR:       true,
		})
	}

	// Get only 5
	prMovements, err := service.GetPRMovements(userID, 5)
	if err != nil {
		t.Fatalf("GetPRMovements failed: %v", err)
	}

	// Should respect limit
	if len(prMovements) != 5 {
		t.Errorf("Expected 5 PR movements (limited), got %d", len(prMovements))
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
