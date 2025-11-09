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
