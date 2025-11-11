package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrUnauthorized = errors.New("unauthorized access")
)

// WorkoutService handles workout template-related business logic
type WorkoutService struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
	movementRepo        domain.MovementRepository
	workoutWODRepo      domain.WorkoutWODRepository
}

// NewWorkoutService creates a new workout service
func NewWorkoutService(
	workoutRepo domain.WorkoutRepository,
	workoutMovementRepo domain.WorkoutMovementRepository,
	movementRepo domain.MovementRepository,
	workoutWODRepo domain.WorkoutWODRepository,
) *WorkoutService {
	return &WorkoutService{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
		movementRepo:        movementRepo,
		workoutWODRepo:      workoutWODRepo,
	}
}

// CreateTemplate creates a new workout template
func (s *WorkoutService) CreateTemplate(userID int64, workout *domain.Workout) error {
	// Set creator and timestamps
	workout.CreatedBy = &userID
	now := time.Now()
	workout.CreatedAt = now
	workout.UpdatedAt = now

	// Create workout template
	err := s.workoutRepo.Create(workout)
	if err != nil {
		return fmt.Errorf("failed to create workout template: %w", err)
	}

	// Create workout movements if provided
	if len(workout.Movements) > 0 {
		now := time.Now()
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

// GetTemplate retrieves a workout template by ID with movements and WODs
func (s *WorkoutService) GetTemplate(templateID int64) (*domain.Workout, error) {
	workout, err := s.workoutRepo.GetByIDWithDetails(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return nil, ErrWorkoutNotFound
	}

	return workout, nil
}

// ListTemplates lists workout templates (standard or user-specific)
func (s *WorkoutService) ListTemplates(userID *int64, limit, offset int) ([]*domain.Workout, error) {
	var workouts []*domain.Workout
	var err error

	if userID == nil {
		// List all standard templates (created_by IS NULL)
		workouts, err = s.workoutRepo.ListStandard(limit, offset)
	} else {
		// List user's templates
		workouts, err = s.workoutRepo.ListByUser(*userID, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list workout templates: %w", err)
	}

	return workouts, nil
}

// UpdateTemplate updates a workout template (only if user is the creator)
func (s *WorkoutService) UpdateTemplate(templateID, userID int64, updates *domain.Workout) error {
	// Get existing template
	workout, err := s.workoutRepo.GetByID(templateID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can update
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Update fields
	workout.Name = updates.Name
	workout.Notes = updates.Notes
	workout.UpdatedAt = time.Now()

	err = s.workoutRepo.Update(workout)
	if err != nil {
		return fmt.Errorf("failed to update workout template: %w", err)
	}

	// Update movements if provided
	if updates.Movements != nil {
		// Delete existing movements
		err = s.workoutMovementRepo.DeleteByWorkoutID(templateID)
		if err != nil {
			return fmt.Errorf("failed to delete workout movements: %w", err)
		}

		// Create new movements
		now := time.Now()
		for i, movement := range updates.Movements {
			movement.WorkoutID = templateID
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

// DeleteTemplate deletes a workout template (only if user is the creator)
func (s *WorkoutService) DeleteTemplate(templateID, userID int64) error {
	// Get existing template
	workout, err := s.workoutRepo.GetByID(templateID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can delete
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Delete template (movements and WODs will be cascade deleted)
	err = s.workoutRepo.Delete(templateID)
	if err != nil {
		return fmt.Errorf("failed to delete workout template: %w", err)
	}

	return nil
}

// GetTemplateUsageStats gets usage statistics for a template
func (s *WorkoutService) GetTemplateUsageStats(templateID int64) (*domain.WorkoutWithUsageStats, error) {
	stats, err := s.workoutRepo.GetUsageStats(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}
	return stats, nil
}

// AddMovementToTemplate adds a movement to a workout template
func (s *WorkoutService) AddMovementToTemplate(templateID, movementID int64, userID int64, wm *domain.WorkoutMovement) error {
	// Get existing template
	workout, err := s.workoutRepo.GetByID(templateID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Verify movement exists
	movement, err := s.movementRepo.GetByID(movementID)
	if err != nil {
		return fmt.Errorf("failed to get movement: %w", err)
	}
	if movement == nil {
		return errors.New("movement not found")
	}

	// Set IDs and timestamps
	wm.WorkoutID = templateID
	wm.MovementID = movementID
	now := time.Now()
	wm.CreatedAt = now
	wm.UpdatedAt = now

	err = s.workoutMovementRepo.Create(wm)
	if err != nil {
		return fmt.Errorf("failed to add movement to template: %w", err)
	}

	return nil
}

// AddWODToTemplate adds a WOD to a workout template
func (s *WorkoutService) AddWODToTemplate(templateID, wodID int64, userID int64, wod *domain.WorkoutWOD) error {
	// Get existing template
	workout, err := s.workoutRepo.GetByID(templateID)
	if err != nil {
		return fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return ErrWorkoutNotFound
	}

	// Authorization check - only creator can modify
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Set IDs and timestamps
	wod.WorkoutID = templateID
	wod.WODID = wodID
	now := time.Now()
	wod.CreatedAt = now
	wod.UpdatedAt = now

	err = s.workoutWODRepo.Create(wod)
	if err != nil {
		return fmt.Errorf("failed to add WOD to template: %w", err)
	}

	return nil
}

// SearchTemplates searches workout templates by name
func (s *WorkoutService) SearchTemplates(userID *int64, query string, limit int) ([]*domain.Workout, error) {
	workouts, err := s.workoutRepo.Search(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search workout templates: %w", err)
	}

	// Filter by user if specified
	if userID != nil {
		var filtered []*domain.Workout
		for _, w := range workouts {
			// Include standard templates (created_by IS NULL) or user's own templates
			if w.CreatedBy == nil || *w.CreatedBy == *userID {
				filtered = append(filtered, w)
			}
		}
		return filtered, nil
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

	// For v0.4.0: workout movements are part of templates, need different authorization
	// Since templates can be standard (created_by = NULL) or user-created,
	// we need to verify the user has access
	workout, err := s.workoutRepo.GetByID(wm.WorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return errors.New("workout not found")
	}

	// Only allow toggling PR on user's own templates
	if workout.CreatedBy == nil || *workout.CreatedBy != userID {
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
