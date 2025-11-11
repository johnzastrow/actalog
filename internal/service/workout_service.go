package service

import (
	"errors"
	"fmt"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrUnauthorized = errors.New("unauthorized access")
)

// WorkoutService handles workout-related business logic
type WorkoutService struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
	workoutWODRepo      domain.WorkoutWODRepository
	movementRepo        domain.MovementRepository
}

// DeleteWorkout deletes a workout with authorization check
func (s *WorkoutService) DeleteWorkout(workoutID, userID int64) error {
	// Get existing workout
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization: only creator can delete custom templates; standard templates cannot be deleted
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
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
// GetWorkoutsByDateRange removed in v0.4.0 (use UserWorkoutService for instances)

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
// Note: This now needs to be called with a user_workout_id context, not workout template
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
	// Authorization based on workout template ownership removed; PR toggling allowed if movement belongs to user's logged workout context upstream.

	// Toggle the PR flag
	wm.IsPR = !wm.IsPR

	// Update the movement
	err = s.workoutMovementRepo.Update(wm)
	if err != nil {
		return fmt.Errorf("failed to update workout movement: %w", err)
	}

	return nil
}
