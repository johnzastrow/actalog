package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	// ErrWorkoutNotFound is returned when a workout template is not found
	ErrWorkoutNotFound = errors.New("workout template not found")
	// ErrUserWorkoutNotFound is returned when a user workout is not found
	ErrUserWorkoutNotFound = errors.New("user workout not found")
	// ErrUnauthorizedWorkoutAccess is returned when a user tries to access another user's workout
	ErrUnauthorizedWorkoutAccess = errors.New("unauthorized workout access")
)

// UserWorkoutService handles business logic for user workout instances (logged workouts)
type UserWorkoutService struct {
	userWorkoutRepo     domain.UserWorkoutRepository
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
}

// NewUserWorkoutService creates a new user workout service
func NewUserWorkoutService(
	userWorkoutRepo domain.UserWorkoutRepository,
	workoutRepo domain.WorkoutRepository,
	workoutMovementRepo domain.WorkoutMovementRepository,
) *UserWorkoutService {
	return &UserWorkoutService{
		userWorkoutRepo:     userWorkoutRepo,
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
	}
}

// LogWorkout creates a new user workout instance (logs that a user performed a workout template)
func (s *UserWorkoutService) LogWorkout(
	userID int64,
	workoutID int64,
	workoutDate time.Time,
	notes *string,
	totalTime *int,
	workoutType *string,
) (*domain.UserWorkout, error) {
	// Verify workout template exists
	workout, err := s.workoutRepo.GetByID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return nil, ErrWorkoutNotFound
	}

	// Verify user has access to this workout template
	// Users can access standard workouts (created_by = NULL) or their own workouts
	if workout.CreatedBy != nil && *workout.CreatedBy != userID {
		return nil, ErrUnauthorizedWorkoutAccess
	}

	// Create user workout instance
	userWorkout := &domain.UserWorkout{
		UserID:      userID,
		WorkoutID:   workoutID,
		WorkoutDate: workoutDate,
		Notes:       notes,
		TotalTime:   totalTime,
		WorkoutType: workoutType,
	}

	err = s.userWorkoutRepo.Create(userWorkout)
	if err != nil {
		return nil, fmt.Errorf("failed to create user workout: %w", err)
	}

	return userWorkout, nil
}

// GetLoggedWorkout retrieves a logged workout with full details (movements, WODs)
func (s *UserWorkoutService) GetLoggedWorkout(userWorkoutID int64, userID int64) (*domain.UserWorkoutWithDetails, error) {
	userWorkout, err := s.userWorkoutRepo.GetByIDWithDetails(userWorkoutID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout: %w", err)
	}
	if userWorkout == nil {
		return nil, ErrUserWorkoutNotFound
	}

	// Verify ownership
	if userWorkout.UserID != userID {
		return nil, ErrUnauthorizedWorkoutAccess
	}

	return userWorkout, nil
}

// ListLoggedWorkouts retrieves all logged workouts for a user with full details
func (s *UserWorkoutService) ListLoggedWorkouts(userID int64, limit, offset int) ([]*domain.UserWorkoutWithDetails, error) {
	workouts, err := s.userWorkoutRepo.ListByUserWithDetails(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user workouts: %w", err)
	}

	return workouts, nil
}

// ListLoggedWorkoutsByDateRange retrieves logged workouts within a date range
func (s *UserWorkoutService) ListLoggedWorkoutsByDateRange(
	userID int64,
	startDate, endDate time.Time,
) ([]*domain.UserWorkout, error) {
	workouts, err := s.userWorkoutRepo.ListByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts by date range: %w", err)
	}

	return workouts, nil
}

// UpdateLoggedWorkout updates a logged workout (notes, time, type)
func (s *UserWorkoutService) UpdateLoggedWorkout(
	userWorkoutID int64,
	userID int64,
	notes *string,
	totalTime *int,
	workoutType *string,
) error {
	// Verify ownership
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get user workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	// Update fields
	existing.Notes = notes
	existing.TotalTime = totalTime
	existing.WorkoutType = workoutType

	err = s.userWorkoutRepo.Update(existing)
	if err != nil {
		return fmt.Errorf("failed to update user workout: %w", err)
	}

	return nil
}

// DeleteLoggedWorkout deletes a logged workout
func (s *UserWorkoutService) DeleteLoggedWorkout(userWorkoutID int64, userID int64) error {
	// Verify ownership
	existing, err := s.userWorkoutRepo.GetByID(userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to get user workout: %w", err)
	}
	if existing == nil {
		return ErrUserWorkoutNotFound
	}
	if existing.UserID != userID {
		return ErrUnauthorizedWorkoutAccess
	}

	err = s.userWorkoutRepo.Delete(userWorkoutID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user workout: %w", err)
	}

	return nil
}

// GetWorkoutStatsForMonth returns the count of workouts logged in a specific month
func (s *UserWorkoutService) GetWorkoutStatsForMonth(userID int64, year, month int) (int, error) {
	// Create date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second) // Last second of the month

	workouts, err := s.userWorkoutRepo.ListByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get workout stats: %w", err)
	}

	return len(workouts), nil
}

// GetWorkoutForDate retrieves a logged workout for a specific user, workout and date
func (s *UserWorkoutService) GetWorkoutForDate(userID, workoutID int64, date time.Time) (*domain.UserWorkoutWithDetails, error) {
	userWorkout, err := s.userWorkoutRepo.GetByUserWorkoutDate(userID, workoutID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout for date: %w", err)
	}
	if userWorkout == nil {
		return nil, ErrUserWorkoutNotFound
	}

	// Get full details
	return s.userWorkoutRepo.GetByIDWithDetails(userWorkout.ID, userID)
}
