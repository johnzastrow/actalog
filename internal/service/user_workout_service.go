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

// LogWorkout logs that a user performed a workout template on a specific date
func (s *UserWorkoutService) LogWorkout(userID, templateID int64, date time.Time, notes *string, totalTime *int, workoutType *string) (*domain.UserWorkout, error) {
	// Verify template exists
	workout, err := s.workoutRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout template: %w", err)
	}
	if workout == nil {
		return nil, fmt.Errorf("workout template not found")
	}

	// Check if user already logged this workout on this date
	existing, err := s.userWorkoutRepo.GetByUserWorkoutDate(userID, templateID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing workout: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("workout already logged for this date")
	}

	// Create user workout
	userWorkout := &domain.UserWorkout{
		UserID:      userID,
		WorkoutID:   templateID,
		WorkoutDate: date,
		WorkoutType: workoutType,
		TotalTime:   totalTime,
		Notes:       notes,
	}

	err = s.userWorkoutRepo.Create(userWorkout)
	if err != nil {
		return nil, fmt.Errorf("failed to log workout: %w", err)
	}

	return userWorkout, nil
}

// GetLoggedWorkout retrieves a logged workout by ID with full details
func (s *UserWorkoutService) GetLoggedWorkout(userWorkoutID, userID int64) (*domain.UserWorkoutWithDetails, error) {
	userWorkout, err := s.userWorkoutRepo.GetByIDWithDetails(userWorkoutID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logged workout: %w", err)
	}
	if userWorkout == nil {
		return nil, fmt.Errorf("logged workout not found")
	}

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
// UpdateLoggedWorkout updates a logged workout (notes, time, type - not date)
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

	// Update fields (not including date)
	existing.WorkoutType = workoutType
	existing.TotalTime = totalTime
	existing.Notes = notes
	existing.UpdatedAt = time.Now()

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
		return fmt.Errorf("logged workout not found")
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
