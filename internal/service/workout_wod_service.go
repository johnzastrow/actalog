package service

import (
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WorkoutWODService handles linking WODs to workout templates
type WorkoutWODService struct {
	workoutWODRepo domain.WorkoutWODRepository
	workoutRepo    domain.WorkoutRepository
	wodRepo        domain.WODRepository
}

// NewWorkoutWODService creates a new workout WOD service
func NewWorkoutWODService(
	workoutWODRepo domain.WorkoutWODRepository,
	workoutRepo domain.WorkoutRepository,
	wodRepo domain.WODRepository,
) *WorkoutWODService {
	return &WorkoutWODService{
		workoutWODRepo: workoutWODRepo,
		workoutRepo:    workoutRepo,
		wodRepo:        wodRepo,
	}
}

// AddWODToWorkout adds a WOD to a workout template with authorization check
func (s *WorkoutWODService) AddWODToWorkout(workoutID, wodID int64, userID int64, orderIndex int, division *string) (*domain.WorkoutWOD, error) {
	// Verify workout template exists and user has permission
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return nil, fmt.Errorf("workout template not found")
	}

	// Authorization check: only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return nil, ErrUnauthorized
	}

	// Verify WOD exists
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return nil, fmt.Errorf("WOD not found")
	}

	// Create workout WOD association
	now := time.Now()
	workoutWOD := &domain.WorkoutWOD{
		WorkoutID:  workoutID,
		WODID:      wodID,
		Division:   division,
		OrderIndex: orderIndex,
		IsPR:       false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err = s.workoutWODRepo.Create(workoutWOD)
	if err != nil {
		return nil, fmt.Errorf("failed to add WOD to workout: %w", err)
	}

	return workoutWOD, nil
}

// RemoveWODFromWorkout removes a WOD from a workout template with authorization check
func (s *WorkoutWODService) RemoveWODFromWorkout(workoutWODID, userID int64) error {
	// Get the workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Get the workout to check authorization
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return fmt.Errorf("workout template not found")
	}

	// Authorization check: only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Delete the association
	err = s.workoutWODRepo.Delete(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to remove WOD from workout: %w", err)
	}

	return nil
}

// UpdateWorkoutWOD updates a WOD in a workout with authorization check
func (s *WorkoutWODService) UpdateWorkoutWOD(workoutWODID, userID int64, scoreValue *string, division *string) error {
	// Get the workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Get the workout to check authorization
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return fmt.Errorf("workout template not found")
	}

	// Authorization check: only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Update fields
	if scoreValue != nil {
		workoutWOD.ScoreValue = scoreValue
	}
	if division != nil {
		workoutWOD.Division = division
	}
	workoutWOD.UpdatedAt = time.Now()

	err = s.workoutWODRepo.Update(workoutWOD)
	if err != nil {
		return fmt.Errorf("failed to update workout WOD: %w", err)
	}

	return nil
}

// ToggleWODPR toggles the PR flag on a WOD in a workout with authorization check
func (s *WorkoutWODService) ToggleWODPR(workoutWODID, userID int64) error {
	// Get the workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Get the workout to check authorization
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return fmt.Errorf("workout template not found")
	}

	// Authorization check: only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Toggle PR flag
	err = s.workoutWODRepo.TogglePR(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to toggle WOD PR: %w", err)
	}

	return nil
}

// ListWODsForWorkout retrieves all WODs associated with a workout template
func (s *WorkoutWODService) ListWODsForWorkout(workoutID int64) ([]*domain.WorkoutWODWithDetails, error) {
	wods, err := s.workoutWODRepo.ListByWorkoutWithDetails(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to list WODs for workout: %w", err)
	}

	return wods, nil
}
