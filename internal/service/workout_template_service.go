package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type WorkoutTemplateService struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
	workoutWODRepo      domain.WorkoutWODRepository
	auditLogRepo        domain.AuditLogRepository
}

func NewWorkoutTemplateService(workoutRepo domain.WorkoutRepository, workoutMovementRepo domain.WorkoutMovementRepository, workoutWODRepo domain.WorkoutWODRepository, auditLogRepo domain.AuditLogRepository) *WorkoutTemplateService {
	return &WorkoutTemplateService{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
		workoutWODRepo:      workoutWODRepo,
		auditLogRepo:        auditLogRepo,
	}
}

// Create creates a new workout template
func (s *WorkoutTemplateService) Create(userID int64, userEmail, userRole, name string, introWarmup, notes *string, isStandard bool, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error) {
	// Determine CreatedBy based on isStandard and role
	var createdBy *int64
	if isStandard && userRole == "admin" {
		// Standard workout: CreatedBy is NULL, making it available to all users
		createdBy = nil
	} else {
		// Personal workout: CreatedBy is the user's ID
		createdBy = &userID
	}

	// Create the workout template
	workout := &domain.Workout{
		Name:        name,
		IntroWarmup: introWarmup,
		Notes:       notes,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.workoutRepo.Create(workout); err != nil {
		return nil, fmt.Errorf("failed to create workout template: %w", err)
	}

	// Add movements if provided
	if len(movements) > 0 {
		for i, movement := range movements {
			wm := &domain.WorkoutMovement{
				WorkoutID:    workout.ID,
				MovementID:   movement.MovementID,
				Sets:         movement.Sets,
				Reps:         movement.Reps,
				Weight:       movement.Weight,
				Time:         movement.Time,
				Distance:     movement.Distance,
				IsRx:         movement.IsRx,
				IsPR:         movement.IsPR,
				Instructions: movement.Instructions,
				Notes:        movement.Notes,
				OrderIndex:   i + 1,
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
				WorkoutID:    workout.ID,
				WODID:        wod.WODID,
				ScoreValue:   wod.ScoreValue,
				Division:     wod.Division,
				IsPR:         wod.IsPR,
				Instructions: wod.Instructions,
				Notes:        wod.Notes,
				OrderIndex:   i + 1,
			}

			if err := s.workoutWODRepo.Create(ww); err != nil {
				return nil, fmt.Errorf("failed to add WOD: %w", err)
			}
		}
	}

	// Audit log
	if s.auditLogRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"workout_id":     workout.ID,
			"workout_name":   name,
			"movement_count": len(movements),
			"wod_count":      len(wods),
			"created_by":     userEmail,
		})
		detailsStr := string(details)
		targetUserID := userID
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:       &targetUserID,
			TargetUserID: &targetUserID,
			EventType:    domain.EventWorkoutTemplateCreated,
			Details:      &detailsStr,
			CreatedAt:    time.Now(),
		})
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

// ListByUser retrieves all workout templates created by a specific user with full details
func (s *WorkoutTemplateService) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	templates, err := s.workoutRepo.ListByUser(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user templates: %w", err)
	}

	// Load full details for each workout
	var result []*domain.Workout
	for _, t := range templates {
		workout, err := s.workoutRepo.GetByIDWithDetails(t.ID)
		if err != nil {
			// If we can't get details, use the basic info
			result = append(result, t)
			continue
		}
		result = append(result, workout)
	}

	return result, nil
}

// ListStandard retrieves all standard (system) workout templates with full details
func (s *WorkoutTemplateService) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	templates, err := s.workoutRepo.ListStandard(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard templates: %w", err)
	}

	// Load full details for each workout
	var result []*domain.Workout
	for _, t := range templates {
		workout, err := s.workoutRepo.GetByIDWithDetails(t.ID)
		if err != nil {
			// If we can't get details, use the basic info
			result = append(result, t)
			continue
		}
		result = append(result, workout)
	}

	return result, nil
}

// Update updates an existing workout template
func (s *WorkoutTemplateService) Update(id, userID int64, userEmail, userRole, name string, introWarmup, notes *string, isStandard bool, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error) {
	// Get existing workout to verify ownership
	existing, err := s.workoutRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workout template not found")
		}
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}

	// Check permissions:
	// - Admins can edit any workout (including standard ones)
	// - Non-admins can only edit their own workouts
	isOwner := existing.CreatedBy != nil && *existing.CreatedBy == userID
	isAdmin := userRole == "admin"

	if !isOwner && !isAdmin {
		return nil, fmt.Errorf("you don't have permission to edit this workout")
	}

	// Store old values for audit logging
	oldName := existing.Name
	oldIntroWarmup := existing.IntroWarmup
	oldNotes := existing.Notes

	// Determine CreatedBy based on isStandard and role
	if isAdmin {
		if isStandard {
			existing.CreatedBy = nil // Make it standard
		} else if existing.CreatedBy == nil {
			// Converting from standard to personal - assign to this admin
			existing.CreatedBy = &userID
		}
	}

	// Update the workout
	existing.Name = name
	existing.IntroWarmup = introWarmup
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
				WorkoutID:    id,
				MovementID:   movement.MovementID,
				Sets:         movement.Sets,
				Reps:         movement.Reps,
				Weight:       movement.Weight,
				Time:         movement.Time,
				Distance:     movement.Distance,
				IsRx:         movement.IsRx,
				IsPR:         movement.IsPR,
				Instructions: movement.Instructions,
				Notes:        movement.Notes,
				OrderIndex:   i + 1,
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
				WorkoutID:    id,
				WODID:        wod.WODID,
				ScoreValue:   wod.ScoreValue,
				Division:     wod.Division,
				IsPR:         wod.IsPR,
				Instructions: wod.Instructions,
				Notes:        wod.Notes,
				OrderIndex:   i + 1,
			}

			if err := s.workoutWODRepo.Create(ww); err != nil {
				return nil, fmt.Errorf("failed to add WOD: %w", err)
			}
		}
	}

	// Audit log
	if s.auditLogRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"workout_id":     id,
			"workout_name":   name,
			"movement_count": len(movements),
			"wod_count":      len(wods),
			"updated_by":     userEmail,
			"changes": map[string]interface{}{
				"name_old":         oldName,
				"name_new":         name,
				"intro_warmup_old": oldIntroWarmup,
				"intro_warmup_new": introWarmup,
				"notes_old":        oldNotes,
				"notes_new":        notes,
			},
		})
		detailsStr := string(details)
		targetUserID := userID
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:       &targetUserID,
			TargetUserID: &targetUserID,
			EventType:    domain.EventWorkoutTemplateUpdated,
			Details:      &detailsStr,
			CreatedAt:    time.Now(),
		})
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
func (s *WorkoutTemplateService) Delete(id, userID int64, userEmail, userRole string) error {
	// Get existing workout to verify ownership
	existing, err := s.workoutRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("workout template not found")
		}
		return fmt.Errorf("failed to get workout template: %w", err)
	}

	// Check permissions:
	// - Admins can delete any workout (including standard ones)
	// - Non-admins can only delete their own workouts
	isOwner := existing.CreatedBy != nil && *existing.CreatedBy == userID
	isAdmin := userRole == "admin"

	if !isOwner && !isAdmin {
		return fmt.Errorf("you don't have permission to delete this workout")
	}

	// Store details for audit log
	workoutName := existing.Name

	// Delete movements first
	if err := s.workoutMovementRepo.DeleteByWorkoutID(id); err != nil {
		return fmt.Errorf("failed to delete movements: %w", err)
	}

	// Delete the workout
	if err := s.workoutRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete workout template: %w", err)
	}

	// Audit log
	if s.auditLogRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"workout_id":   id,
			"workout_name": workoutName,
			"deleted_by":   userEmail,
		})
		detailsStr := string(details)
		targetUserID := userID
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:       &targetUserID,
			TargetUserID: &targetUserID,
			EventType:    domain.EventWorkoutTemplateDeleted,
			Details:      &detailsStr,
			CreatedAt:    time.Now(),
		})
	}

	return nil
}
