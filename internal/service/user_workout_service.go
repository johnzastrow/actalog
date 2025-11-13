package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrUserWorkoutNotFound       = errors.New("user workout not found")
	ErrUnauthorizedWorkoutAccess = errors.New("unauthorized workout access")
)

// UserWorkoutService handles logging workout instances (when users perform workouts)
type UserWorkoutService struct {
	userWorkoutRepo         domain.UserWorkoutRepository
	workoutRepo             domain.WorkoutRepository
	workoutMovementRepo     domain.WorkoutMovementRepository
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository
	userWorkoutWODRepo      domain.UserWorkoutWODRepository
}

// NewUseroutService creates a new user workout service
func NewUserWorkoutService(
	userWorkoutRepo domain.UserWorkoutRepository,
	workoutRepo domain.WorkoutRepository,
	workoutMovementRepo domain.WorkoutMovementRepository,
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository,
	userWorkoutWODRepo domain.UserWorkoutWODRepository,
) *UserWorkoutService {
	return &UserWorkoutService{
		userWorkoutRepo:         userWorkoutRepo,
		workoutRepo:             workoutRepo,
		workoutMovementRepo:     workoutMovementRepo,
		userWorkoutMovementRepo: userWorkoutMovementRepo,
		userWorkoutWODRepo:      userWorkoutWODRepo,
	}
}

// LogWorkout logs that a user performed a workout (template-based or ad-hoc) on a specific date
func (s *UserWorkoutService) LogWorkout(userID int64, templateID *int64, workoutName *string, date time.Time, notes *string, totalTime *int, workoutType *string) (*domain.UserWorkout, error) {
	// If template ID is provided, verify it exists and check authorization
	if templateID != nil && *templateID != 0 {
		workout, err := s.workoutRepo.GetByID(*templateID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workout template: %w", err)
		}
		if workout == nil {
			return nil, ErrWorkoutNotFound
		}

		// Check authorization: user can only log workouts they created or standard workouts (created_by = null)
		if workout.CreatedBy != nil && *workout.CreatedBy != userID {
			return nil, ErrUnauthorizedWorkoutAccess
		}
	}

	// Create user workout (users can log the same workout multiple times per day)
	userWorkout := &domain.UserWorkout{
		UserID:      userID,
		WorkoutID:   templateID,
		WorkoutName: workoutName,
		WorkoutDate: date,
		WorkoutType: workoutType,
		TotalTime:   totalTime,
		Notes:       notes,
	}

	err := s.userWorkoutRepo.Create(userWorkout)
	if err != nil {
		return nil, fmt.Errorf("failed to log workout: %w", err)
	}

	return userWorkout, nil
}

// LogWorkoutWithPerformance logs a workout with full performance data for movements and WODs
func (s *UserWorkoutService) LogWorkoutWithPerformance(
	userID int64,
	templateID *int64,
	workoutName *string,
	date time.Time,
	notes *string,
	totalTime *int,
	workoutType *string,
	movements []*domain.UserWorkoutMovement,
	wods []*domain.UserWorkoutWOD,
) (*domain.UserWorkout, error) {
	// First create the base user workout
	userWorkout, err := s.LogWorkout(userID, templateID, workoutName, date, notes, totalTime, workoutType)
	if err != nil {
		return nil, err
	}

	// Set the user_workout_id for all movements
	for _, m := range movements {
		m.UserWorkoutID = userWorkout.ID
	}

	// Set the user_workout_id for all WODs
	for _, w := range wods {
		w.UserWorkoutID = userWorkout.ID
	}

	// Save movement performance data
	if len(movements) > 0 {
		if err := s.userWorkoutMovementRepo.CreateBatch(movements); err != nil {
			// Rollback: delete the user workout if performance data fails
			_ = s.userWorkoutRepo.Delete(userWorkout.ID, userID)
			return nil, fmt.Errorf("failed to save movement performance data: %w", err)
		}
	}

	// Save WOD performance data
	if len(wods) > 0 {
		if err := s.userWorkoutWODRepo.CreateBatch(wods); err != nil {
			// Rollback: delete movement data and user workout
			_ = s.userWorkoutMovementRepo.DeleteByUserWorkoutID(userWorkout.ID)
			_ = s.userWorkoutRepo.Delete(userWorkout.ID, userID)
			return nil, fmt.Errorf("failed to save WOD performance data: %w", err)
		}
	}

	return userWorkout, nil
}

// GetLoggedWorkout retrieves a logged workout by ID with full details including performance data
func (s *UserWorkoutService) GetLoggedWorkout(userWorkoutID, userID int64) (*domain.UserWorkoutWithDetails, error) {
	// First check if workout exists (without user filtering)
	basic, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logged workout: %w", err)
	}
	if basic == nil {
		return nil, ErrUserWorkoutNotFound
	}

	// Check authorization
	if basic.UserID != userID {
		return nil, ErrUnauthorizedWorkoutAccess
	}

	// Get full details
	userWorkout, err := s.userWorkoutRepo.GetByIDWithDetails(userWorkoutID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logged workout details: %w", err)
	}

	// Load performance data for movements
	performanceMovements, err := s.userWorkoutMovementRepo.GetByUserWorkoutID(userWorkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movement performance data: %w", err)
	}
	userWorkout.PerformanceMovements = performanceMovements

	// Load performance data for WODs
	performanceWODs, err := s.userWorkoutWODRepo.GetByUserWorkoutID(userWorkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD performance data: %w", err)
	}
	userWorkout.PerformanceWODs = performanceWODs

	return userWorkout, nil
}

// ListLoggedWorkouts retrieves all workouts logged by a user
func (s *UserWorkoutService) ListLoggedWorkouts(userID int64, limit, offset int) ([]*domain.UserWorkoutWithDetails, error) {
	workouts, err := s.userWorkoutRepo.ListByUserWithDetails(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list logged workouts: %w", err)
	}

	return workouts, nil
}

// ListLoggedWorkoutsByDateRange retrieves workouts within a date range
func (s *UserWorkoutService) ListLoggedWorkoutsByDateRange(userID int64, startDate, endDate time.Time) ([]*domain.UserWorkout, error) {
	workouts, err := s.userWorkoutRepo.ListByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts by date range: %w", err)
	}

	return workouts, nil
}

// UpdateLoggedWorkout updates a logged workout with authorization check
func (s *UserWorkoutService) UpdateLoggedWorkout(userWorkoutID, userID int64, notes *string, totalTime *int, workoutType *string) error {
	// Get existing logged workout
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get logged workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}

	// Authorization check
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	// Update fields
	if notes != nil {
		existing.Notes = notes
	}
	if totalTime != nil {
		existing.TotalTime = totalTime
	}
	if workoutType != nil {
		existing.WorkoutType = workoutType
	}

	err = s.userWorkoutRepo.Update(existing)
	if err != nil {
		return fmt.Errorf("failed to update logged workout: %w", err)
	}
	return nil
}

// DeleteLoggedWorkout deletes a logged workout with authorization check
func (s *UserWorkoutService) DeleteLoggedWorkout(userWorkoutID, userID int64) error {
	// Get existing logged workout
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get logged workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}

	// Authorization check
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	// Delete logged workout
	err = s.userWorkoutRepo.Delete(userWorkoutID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete logged workout: %w", err)
	}
	return nil
}

// UpdateWorkoutMovements updates the movements for a logged workout
func (s *UserWorkoutService) UpdateWorkoutMovements(userWorkoutID, userID int64, movements []domain.UserWorkoutMovement) error {
	// Authorization check
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get logged workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	// Delete existing movements
	if err := s.userWorkoutMovementRepo.DeleteByUserWorkoutID(userWorkoutID); err != nil {
		return fmt.Errorf("failed to delete existing movements: %w", err)
	}

	// Insert new movements
	for _, movement := range movements {
		movement.UserWorkoutID = userWorkoutID
		if err := s.userWorkoutMovementRepo.Create(&movement); err != nil {
			return fmt.Errorf("failed to create movement: %w", err)
		}
	}

	return nil
}

// UpdateWorkoutWODs updates the WODs for a logged workout
func (s *UserWorkoutService) UpdateWorkoutWODs(userWorkoutID, userID int64, wods []domain.UserWorkoutWOD) error {
	// Authorization check
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get logged workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	// Delete existing WODs
	if err := s.userWorkoutWODRepo.DeleteByUserWorkoutID(userWorkoutID); err != nil {
		return fmt.Errorf("failed to delete existing WODs: %w", err)
	}

	// Insert new WODs
	for _, wod := range wods {
		wod.UserWorkoutID = userWorkoutID
		if err := s.userWorkoutWODRepo.Create(&wod); err != nil {
			return fmt.Errorf("failed to create WOD: %w", err)
		}
	}

	return nil
}

// GetWorkoutStatsForMonth counts workouts logged in a specific month
func (s *UserWorkoutService) GetWorkoutStatsForMonth(userID int64, year, month int) (int, error) {
	// Calculate start and end dates for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get workouts in range
	workouts, err := s.userWorkoutRepo.ListByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to list workouts by month: %w", err)
	}
	return len(workouts), nil
}
