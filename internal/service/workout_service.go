package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrWorkoutNotFound = errors.New("workout not found")
	ErrUnauthorized    = errors.New("unauthorized access")
)

// WorkoutService handles workout-related business logic
type WorkoutService struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
	movementRepo        domain.MovementRepository
}

// NewWorkoutService creates a new workout service
func NewWorkoutService(
	workoutRepo domain.WorkoutRepository,
	workoutMovementRepo domain.WorkoutMovementRepository,
	movementRepo domain.MovementRepository,
) *WorkoutService {
	return &WorkoutService{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
		movementRepo:        movementRepo,
	}
}

// CreateWorkout creates a new workout with movements
func (s *WorkoutService) CreateWorkout(userID int64, workout *domain.Workout) error {
	// Set user ID and timestamps
	workout.UserID = userID
	now := time.Now()
	workout.CreatedAt = now
	workout.UpdatedAt = now

	// Create workout
	err := s.workoutRepo.Create(workout)
	if err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}

	// Create workout movements if provided
	if len(workout.Movements) > 0 {
		// Detect PRs before creating movements
		err = s.DetectAndFlagPRs(userID, workout.Movements)
		if err != nil {
			return fmt.Errorf("failed to detect PRs: %w", err)
		}

		for i, movement := range workout.Movements {
			movement.WorkoutID = workout.ID
			movement.OrderIndex = i
			movement.CreatedAt = now
			movement.UpdatedAt = now

			err = s.workoutMovementRepo.Create(movement)
			if err != nil {
				return fmt.Errorf("failed to create workout movement: %w", err)
			}
		}
	}

	return nil
}

// GetWorkout retrieves a workout by ID with authorization check
func (s *WorkoutService) GetWorkout(workoutID, userID int64) (*domain.Workout, error) {
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return nil, ErrWorkoutNotFound
	}

	// Authorization check
	if workout.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Load movements
	movements, err := s.workoutMovementRepo.GetByWorkoutID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout movements: %w", err)
	}

	// Load movement details for each workout movement
	for _, wm := range movements {
		movement, err := s.movementRepo.GetByID(wm.MovementID)
		if err != nil {
			return nil, fmt.Errorf("failed to get movement details: %w", err)
		}
		wm.Movement = movement
	}

	workout.Movements = movements

	return workout, nil
}

// ListUserWorkouts retrieves workouts for a user with pagination
func (s *WorkoutService) ListUserWorkouts(userID int64, limit, offset int) ([]*domain.Workout, error) {
	workouts, err := s.workoutRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}

	// Load movements for each workout
	for _, workout := range workouts {
		movements, err := s.workoutMovementRepo.GetByWorkoutID(workout.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workout movements: %w", err)
		}

		// Load movement details for each workout movement
		for _, wm := range movements {
			movement, err := s.movementRepo.GetByID(wm.MovementID)
			if err != nil {
				return nil, fmt.Errorf("failed to get movement details: %w", err)
			}
			wm.Movement = movement
		}

		workout.Movements = movements
	}

	return workouts, nil
}

// UpdateWorkout updates a workout with authorization check
func (s *WorkoutService) UpdateWorkout(workoutID, userID int64, updates *domain.Workout) error {
	// Get existing workout
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check
	if workout.UserID != userID {
		return ErrUnauthorized
	}

	// Update fields
	workout.WorkoutDate = updates.WorkoutDate
	workout.WorkoutType = updates.WorkoutType
	workout.WorkoutName = updates.WorkoutName
	workout.Notes = updates.Notes
	workout.TotalTime = updates.TotalTime
	workout.UpdatedAt = time.Now()

	err = s.workoutRepo.Update(workout)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	// Update movements if provided
	if updates.Movements != nil {
		// Delete existing movements
		err = s.workoutMovementRepo.DeleteByWorkoutID(workoutID)
		if err != nil {
			return fmt.Errorf("failed to delete workout movements: %w", err)
		}

		// Create new movements
		now := time.Now()
		for i, movement := range updates.Movements {
			movement.WorkoutID = workoutID
			movement.OrderIndex = i
			movement.CreatedAt = now
			movement.UpdatedAt = now

			err = s.workoutMovementRepo.Create(movement)
			if err != nil {
				return fmt.Errorf("failed to create workout movement: %w", err)
			}
		}
	}

	return nil
}

// DeleteWorkout deletes a workout with authorization check
func (s *WorkoutService) DeleteWorkout(workoutID, userID int64) error {
	// Get existing workout
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check
	if workout.UserID != userID {
		return ErrUnauthorized
	}

	// Delete workout (movements will be cascade deleted)
	err = s.workoutRepo.Delete(workoutID)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	return nil
}

// GetWorkoutsByDateRange retrieves workouts for a user within a date range
func (s *WorkoutService) GetWorkoutsByDateRange(userID int64, startDate, endDate time.Time) ([]*domain.Workout, error) {
	workouts, err := s.workoutRepo.GetByUserIDAndDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get workouts by date range: %w", err)
	}

	// Load movements for each workout
	for _, workout := range workouts {
		movements, err := s.workoutMovementRepo.GetByWorkoutID(workout.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workout movements: %w", err)
		}

		// Load movement details for each workout movement
		for _, wm := range movements {
			movement, err := s.movementRepo.GetByID(wm.MovementID)
			if err != nil {
				return nil, fmt.Errorf("failed to get movement details: %w", err)
			}
			wm.Movement = movement
		}

		workout.Movements = movements
	}

	return workouts, nil
}

// ListMovements retrieves all available movements (standard + user custom)
func (s *WorkoutService) ListMovements(userID int64) ([]*domain.Movement, error) {
	// Get standard movements
	standard, err := s.movementRepo.ListStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to list standard movements: %w", err)
	}

	// Get user's custom movements
	custom, err := s.movementRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user movements: %w", err)
	}

	// Combine both lists
	movements := append(standard, custom...)
	return movements, nil
}

// DetectAndFlagPRs automatically detects personal records for movements with weight
func (s *WorkoutService) DetectAndFlagPRs(userID int64, movements []*domain.WorkoutMovement) error {
	for _, wm := range movements {
		// Only check for PRs on movements with weight
		if wm.Weight == nil {
			continue
		}

		// Get max weight for this movement for this user
		maxWeight, err := s.workoutMovementRepo.GetMaxWeightForMovement(userID, wm.MovementID)
		if err != nil {
			return fmt.Errorf("failed to get max weight: %w", err)
		}

		// If this is the first time doing this movement, or if weight exceeds previous max, it's a PR
		if maxWeight == nil || *wm.Weight > *maxWeight {
			wm.IsPR = true
		}
	}
	return nil
}

// GetPersonalRecords retrieves all personal records for a user
func (s *WorkoutService) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	records, err := s.workoutMovementRepo.GetPersonalRecords(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal records: %w", err)
	}
	return records, nil
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (s *WorkoutService) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	movements, err := s.workoutMovementRepo.GetPRMovements(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR movements: %w", err)
	}
	return movements, nil
}

// TogglePRFlag manually toggles the PR flag on a workout movement
func (s *WorkoutService) TogglePRFlag(movementID, userID int64) error {
	// Get the workout movement
	wm, err := s.workoutMovementRepo.GetByID(movementID)
	if err != nil {
		return fmt.Errorf("failed to get workout movement: %w", err)
	}
	if wm == nil {
		return errors.New("workout movement not found")
	}

	// Get the workout to check authorization
	workout, err := s.workoutRepo.GetByID(wm.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return errors.New("workout not found")
	}
	if workout.UserID != userID {
		return ErrUnauthorized
	}

	// Toggle the PR flag
	wm.IsPR = !wm.IsPR

	// Update the movement
	err = s.workoutMovementRepo.Update(wm)
	if err != nil {
		return fmt.Errorf("failed to update workout movement: %w", err)
	}

	return nil
}
