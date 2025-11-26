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
	workoutWODRepo      domain.WorkoutWODRepository
}

func NewWorkoutTemplateService(workoutRepo domain.WorkoutRepository, workoutMovementRepo domain.WorkoutMovementRepository, workoutWODRepo domain.WorkoutWODRepository) *WorkoutTemplateService {
	return &WorkoutTemplateService{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
		workoutWODRepo:      workoutWODRepo,
	}
}

// Create creates a new workout template
func (s *WorkoutTemplateService) Create(userID int64, name string, notes *string, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error) {
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

	// Add WODs if provided
	if len(wods) > 0 {
		for i, wod := range wods {
			ww := &domain.WorkoutWOD{
				WorkoutID:  workout.ID,
				WODID:      wod.WODID,
				OrderIndex: i + 1,
			}

			if err := s.workoutWODRepo.Create(ww); err != nil {
				return nil, fmt.Errorf("failed to add WOD: %w", err)
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
func (s *WorkoutTemplateService) Update(id, userID int64, name string, notes *string, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error) {
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

	// Delete existing WODs
	if err := s.workoutWODRepo.DeleteByWorkout(id); err != nil {
		return nil, fmt.Errorf("failed to delete existing WODs: %w", err)
	}

	// Add new WODs
	if len(wods) > 0 {
		for i, wod := range wods {
			ww := &domain.WorkoutWOD{
				WorkoutID:  id,
				WODID:      wod.WODID,
				OrderIndex: i + 1,
			}

			if err := s.workoutWODRepo.Create(ww); err != nil {
				return nil, fmt.Errorf("failed to add WOD: %w", err)
			}
		}
	}

	// Reload with details
	return s.GetByIDWithDetails(id)
}

// ListAllUserCreated retrieves all user-created workout templates across all users (admin only)
func (s *WorkoutTemplateService) ListAllUserCreated(limit, offset int) ([]*domain.Workout, int64, error) {
	// Get the list
	workouts, err := s.workoutRepo.ListAllUserCreated(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created workouts: %w", err)
	}

	// Get the count
	count, err := s.workoutRepo.CountAllUserCreated()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user-created workouts: %w", err)
	}

	return workouts, count, nil
}

// ListAllUserCreatedWithUserInfo retrieves all user-created workout templates with creator info (admin only)
func (s *WorkoutTemplateService) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WorkoutWithCreator, int64, error) {
	// Get the list with user info
	workouts, err := s.workoutRepo.ListAllUserCreatedWithUserInfo(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created workouts with user info: %w", err)
	}

	// Get the count
	count, err := s.workoutRepo.CountAllUserCreated()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user-created workouts: %w", err)
	}

	return workouts, count, nil
}

// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created workout templates with creator info and filters (admin only)
func (s *WorkoutTemplateService) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*domain.WorkoutWithCreator, int64, error) {
	// Get the list with user info and filters
	workouts, count, err := s.workoutRepo.ListAllUserCreatedWithUserInfoFiltered(limit, offset, search, creator)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created workouts with filters: %w", err)
	}

	return workouts, count, nil
}

// CopyToStandard creates a standard workout template from a user-created one (admin only)
func (s *WorkoutTemplateService) CopyToStandard(id int64, newName string) (*domain.Workout, error) {
	// Validate new name
	if newName == "" {
		return nil, fmt.Errorf("new name is required")
	}

	// Copy to standard (including movements and WODs)
	workout, err := s.workoutRepo.CopyToStandard(id, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to copy workout to standard: %w", err)
	}

	return workout, nil
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
