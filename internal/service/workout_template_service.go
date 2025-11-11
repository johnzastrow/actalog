package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type WorkoutTemplateService struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
}

func NewWorkoutTemplateService(workoutRepo domain.WorkoutRepository, workoutMovementRepo domain.WorkoutMovementRepository) *WorkoutTemplateService {
	return &WorkoutTemplateService{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
	}
}

// Create creates a new workout template
func (s *WorkoutTemplateService) Create(userID int64, name string, notes *string, movements []domain.WorkoutMovement) (*domain.Workout, error) {
	// Create the workout template
	workout := &domain.Workout{
		Name:      name,
		Notes:     notes,
		CreatedBy: &userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.workoutRepo.Create(workout); err != nil {
		return nil, fmt.Errorf("failed to create workout template: %w", err)
	}

	// Add movements if provided
	if len(movements) > 0 {
		for i, movement := range movements {
			wm := &domain.WorkoutMovement{
				WorkoutID:  workout.ID,
				MovementID: movement.MovementID,
				Sets:       movement.Sets,
				Reps:       movement.Reps,
				Weight:     movement.Weight,
				Time:       movement.Time,
				Distance:   movement.Distance,
				Notes:      movement.Notes,
				OrderIndex: i + 1,
			}

			if err := s.workoutMovementRepo.Create(wm); err != nil {
				return nil, fmt.Errorf("failed to add movement: %w", err)
			}
		}
	}

	// Reload with details
	return s.GetByIDWithDetails(workout.ID)
}

// GetByID retrieves a workout template by ID
func (s *WorkoutTemplateService) GetByID(id int64) (*domain.Workout, error) {
	workout, err := s.workoutRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workout template not found")
		}
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	return workout, nil
}

// GetByIDWithDetails retrieves a workout with movements and WODs
func (s *WorkoutTemplateService) GetByIDWithDetails(id int64) (*domain.Workout, error) {
	workout, err := s.workoutRepo.GetByIDWithDetails(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workout template not found")
		}
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	return workout, nil
}

// ListByUser retrieves all workout templates created by a specific user
func (s *WorkoutTemplateService) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	templates, err := s.workoutRepo.ListByUser(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user templates: %w", err)
	}
	return templates, nil
}

// ListStandard retrieves all standard (system) workout templates
func (s *WorkoutTemplateService) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	templates, err := s.workoutRepo.ListStandard(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard templates: %w", err)
	}
	return templates, nil
}

// Update updates an existing workout template
func (s *WorkoutTemplateService) Update(id, userID int64, name string, notes *string, movements []domain.WorkoutMovement) (*domain.Workout, error) {
	// Get existing workout to verify ownership
	existing, err := s.workoutRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workout template not found")
		}
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}

	// Verify user owns this template
	if existing.CreatedBy == nil || *existing.CreatedBy != userID {
		return nil, fmt.Errorf("you don't have permission to edit this template")
	}

	// Update the workout
	existing.Name = name
	existing.Notes = notes
	existing.UpdatedAt = time.Now()

	if err := s.workoutRepo.Update(existing); err != nil {
		return nil, fmt.Errorf("failed to update workout template: %w", err)
	}

	// Delete existing movements
	if err := s.workoutMovementRepo.DeleteByWorkoutID(id); err != nil {
		return nil, fmt.Errorf("failed to delete existing movements: %w", err)
	}

	// Add new movements
	if len(movements) > 0 {
		for i, movement := range movements {
			wm := &domain.WorkoutMovement{
				WorkoutID:  id,
				MovementID: movement.MovementID,
				Sets:       movement.Sets,
				Reps:       movement.Reps,
				Weight:     movement.Weight,
				Time:       movement.Time,
				Distance:   movement.Distance,
				Notes:      movement.Notes,
				OrderIndex: i + 1,
			}

			if err := s.workoutMovementRepo.Create(wm); err != nil {
				return nil, fmt.Errorf("failed to add movement: %w", err)
			}
		}
	}

	// Reload with details
	return s.GetByIDWithDetails(id)
}

// Delete deletes a workout template
func (s *WorkoutTemplateService) Delete(id, userID int64) error {
	// Get existing workout to verify ownership
	existing, err := s.workoutRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("workout template not found")
		}
		return fmt.Errorf("failed to get workout template: %w", err)
	}

	// Verify user owns this template
	if existing.CreatedBy == nil || *existing.CreatedBy != userID {
		return fmt.Errorf("you don't have permission to delete this template")
	}

	// Delete movements first
	if err := s.workoutMovementRepo.DeleteByWorkoutID(id); err != nil {
		return fmt.Errorf("failed to delete movements: %w", err)
	}

	// Delete the workout
	if err := s.workoutRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete workout template: %w", err)
	}

	return nil
}
