package service

import (
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WorkoutWODService handles business logic for linking WODs to workout templates
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

// AddWODToWorkout adds a WOD to a workout template
func (s *WorkoutWODService) AddWODToWorkout(
	workoutID int64,
	wodID int64,
	userID int64,
	orderIndex int,
	division *string,
) (*domain.WorkoutWOD, error) {
	// Verify workout template exists and user has permission
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return nil, ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return nil, ErrUnauthorized
	}

	// Verify WOD exists
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return nil, ErrWODNotFound
	}

	// Create workout WOD link
	workoutWOD := &domain.WorkoutWOD{
		WorkoutID:  workoutID,
		WODID:      wodID,
		OrderIndex: orderIndex,
		Division:   division,
		IsPR:       false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.workoutWODRepo.Create(workoutWOD)
	if err != nil {
		return nil, fmt.Errorf("failed to add WOD to workout: %w", err)
	}

	return workoutWOD, nil
}

// RemoveWODFromWorkout removes a WOD from a workout template
func (s *WorkoutWODService) RemoveWODFromWorkout(workoutWODID int64, userID int64) error {
	// Get workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Verify workout template exists and user has permission
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify template
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	err = s.workoutWODRepo.Delete(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to remove WOD from workout: %w", err)
	}

	return nil
}

// UpdateWorkoutWOD updates a WOD in a workout template (score, division, etc.)
func (s *WorkoutWODService) UpdateWorkoutWOD(
	workoutWODID int64,
	userID int64,
	scoreValue *string,
	division *string,
) error {
	// Get workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Verify workout template exists and user has permission
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify template
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

// ToggleWODPR toggles the PR flag on a WOD in a workout template
func (s *WorkoutWODService) ToggleWODPR(workoutWODID int64, userID int64) error {
	// Get workout WOD
	workoutWOD, err := s.workoutWODRepo.GetByID(workoutWODID)
	if err != nil {
		return fmt.Errorf("failed to get workout WOD: %w", err)
	}
	if workoutWOD == nil {
		return fmt.Errorf("workout WOD not found")
	}

	// Verify workout template exists and user has permission
	workout, err := s.workoutRepo.GetByID(workoutWOD.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify template
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

// ListWODsForWorkout retrieves all WODs in a workout template with details
func (s *WorkoutWODService) ListWODsForWorkout(workoutID int64) ([]*domain.WorkoutWODWithDetails, error) {
	wods, err := s.workoutWODRepo.ListByWorkoutWithDetails(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to list WODs for workout: %w", err)
	}

	return wods, nil
}
