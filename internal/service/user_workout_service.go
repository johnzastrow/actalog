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
	wodRepo                 domain.WODRepository
}

// NewUseroutService creates a new user workout service
func NewUserWorkoutService(
	userWorkoutRepo domain.UserWorkoutRepository,
	workoutRepo domain.WorkoutRepository,
	workoutMovementRepo domain.WorkoutMovementRepository,
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository,
	userWorkoutWODRepo domain.UserWorkoutWODRepository,
	wodRepo domain.WODRepository,
) *UserWorkoutService {
	return &UserWorkoutService{
		userWorkoutRepo:         userWorkoutRepo,
		workoutRepo:             workoutRepo,
		workoutMovementRepo:     workoutMovementRepo,
		userWorkoutMovementRepo: userWorkoutMovementRepo,
		userWorkoutWODRepo:      userWorkoutWODRepo,
		wodRepo:                 wodRepo,
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

	// Detect and flag PRs for movements before saving
	if len(movements) > 0 {
		if err := s.DetectAndFlagMovementPRs(userID, movements); err != nil {
			_ = s.userWorkoutRepo.Delete(userWorkout.ID, userID)
			return nil, fmt.Errorf("failed to detect movement PRs: %w", err)
		}
	}

	// Validate WOD score types before saving
	if len(wods) > 0 {
		if err := s.ValidateWODScoreTypes(wods); err != nil {
			_ = s.userWorkoutRepo.Delete(userWorkout.ID, userID)
			return nil, fmt.Errorf("WOD validation failed: %w", err)
		}
	}

	// Detect and flag PRs for WODs before saving
	if len(wods) > 0 {
		if err := s.DetectAndFlagWODPRs(userID, wods); err != nil {
			_ = s.userWorkoutRepo.Delete(userWorkout.ID, userID)
			return nil, fmt.Errorf("failed to detect WOD PRs: %w", err)
		}
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
func (s *UserWorkoutService) UpdateLoggedWorkout(userWorkoutID, userID int64, workoutName *string, notes *string, totalTime *int, workoutType *string) error {
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
	if workoutName != nil {
		// Only allow updating workout_name for ad-hoc workouts (workout_id is null)
		if existing.WorkoutID == nil {
			existing.WorkoutName = workoutName
		}
	}
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

	// Validate WOD score types before updating
	wodPointers := make([]*domain.UserWorkoutWOD, len(wods))
	for i := range wods {
		wodPointers[i] = &wods[i]
	}
	if err := s.ValidateWODScoreTypes(wodPointers); err != nil {
		return fmt.Errorf("WOD validation failed: %w", err)
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

// DetectAndFlagMovementPRs automatically detects personal records for movements with weight
func (s *UserWorkoutService) DetectAndFlagMovementPRs(userID int64, movements []*domain.UserWorkoutMovement) error {
	for _, m := range movements {
		// Only check for PRs on movements with weight
		if m.Weight == nil {
			continue
		}

		// Get max weight for this movement for this user
		maxWeight, err := s.userWorkoutMovementRepo.GetMaxWeightForMovement(userID, m.MovementID)
		if err != nil {
			return fmt.Errorf("failed to get max weight for movement %d: %w", m.MovementID, err)
		}

		// If this is the first time doing this movement, or if weight exceeds previous max, it's a PR
		if maxWeight == nil || *m.Weight > *maxWeight {
			m.IsPR = true
		}
	}
	return nil
}

// DetectAndFlagWODPRs automatically detects personal records for WODs (time-based or rounds+reps)
func (s *UserWorkoutService) DetectAndFlagWODPRs(userID int64, wods []*domain.UserWorkoutWOD) error {
	for _, w := range wods {
		// Check for time-based PRs (fastest time)
		if w.TimeSeconds != nil {
			bestTime, err := s.userWorkoutWODRepo.GetBestTimeForWOD(userID, w.WODID)
			if err != nil {
				return fmt.Errorf("failed to get best time for WOD %d: %w", w.WODID, err)
			}

			// If this is the first time doing this WOD, or if time is faster than previous best, it's a PR
			if bestTime == nil || *w.TimeSeconds < *bestTime {
				w.IsPR = true
			}
			continue
		}

		// Check for rounds+reps PRs (most rounds, then most reps)
		if w.Rounds != nil {
			bestRounds, bestReps, err := s.userWorkoutWODRepo.GetBestRoundsRepsForWOD(userID, w.WODID)
			if err != nil {
				return fmt.Errorf("failed to get best rounds+reps for WOD %d: %w", w.WODID, err)
			}

			// If this is the first time doing this WOD, it's a PR
			if bestRounds == nil {
				w.IsPR = true
				continue
			}

			// Check if current rounds > best rounds
			if *w.Rounds > *bestRounds {
				w.IsPR = true
				continue
			}

			// If rounds are equal, check reps
			if *w.Rounds == *bestRounds && w.Reps != nil && bestReps != nil && *w.Reps > *bestReps {
				w.IsPR = true
				continue
			}
		}
	}
	return nil
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (s *UserWorkoutService) GetPRMovements(userID int64, limit int) ([]*domain.UserWorkoutMovement, error) {
	movements, err := s.userWorkoutMovementRepo.GetPRMovements(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR movements: %w", err)
	}
	return movements, nil
}

// GetPRWODs retrieves recent PR-flagged WODs for a user
func (s *UserWorkoutService) GetPRWODs(userID int64, limit int) ([]*domain.UserWorkoutWOD, error) {
	wods, err := s.userWorkoutWODRepo.GetPRWODs(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR WODs: %w", err)
	}
	return wods, nil
}

// RetroactivelyFlagPRs analyzes all existing workouts for a user and flags PRs based on historical max values
func (s *UserWorkoutService) RetroactivelyFlagPRs(userID int64) (int, int, error) {
	movementPRCount := 0
	wodPRCount := 0

	// Get all user workouts ordered by date (chronologically)
	workouts, err := s.userWorkoutRepo.ListByUserAndDateRange(userID, time.Time{}, time.Now().AddDate(0, 0, 1))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user workouts: %w", err)
	}

	// Track max weights per movement_id
	maxWeights := make(map[int64]float64)

	// Track best times per wod_id
	bestTimes := make(map[int64]int)

	// Track best rounds+reps per wod_id
	bestRoundsReps := make(map[int64]struct {
		rounds int
		reps   int
	})

	// Process each workout chronologically
	for _, workout := range workouts {
		// Get movements for this workout
		movements, err := s.userWorkoutMovementRepo.GetByUserWorkoutID(workout.ID)
		if err != nil {
			return movementPRCount, wodPRCount, fmt.Errorf("failed to get movements for workout %d: %w", workout.ID, err)
		}

		// Process each movement
		for _, movement := range movements {
			// Only process movements with weight
			if movement.Weight == nil {
				continue
			}

			currentWeight := *movement.Weight
			movementID := movement.MovementID

			// Check if this is a PR
			isPR := false
			if maxWeight, exists := maxWeights[movementID]; exists {
				// Compare with previous max
				if currentWeight > maxWeight {
					isPR = true
					maxWeights[movementID] = currentWeight
				}
			} else {
				// First time doing this movement - it's a PR
				isPR = true
				maxWeights[movementID] = currentWeight
			}

			// Update PR flag if needed
			if isPR != movement.IsPR {
				if err := s.userWorkoutMovementRepo.UpdatePRFlag(movement.ID, isPR); err != nil {
					return movementPRCount, wodPRCount, fmt.Errorf("failed to update PR flag for movement %d: %w", movement.ID, err)
				}
				if isPR {
					movementPRCount++
				}
			}
		}

		// Get WODs for this workout
		wods, err := s.userWorkoutWODRepo.GetByUserWorkoutID(workout.ID)
		if err != nil {
			return movementPRCount, wodPRCount, fmt.Errorf("failed to get WODs for workout %d: %w", workout.ID, err)
		}

		// Process each WOD
		for _, wod := range wods {
			wodID := wod.WODID
			isPR := false

			// Check time-based PRs
			if wod.TimeSeconds != nil {
				currentTime := *wod.TimeSeconds

				if bestTime, exists := bestTimes[wodID]; exists {
					// Compare with previous best (lower time is better)
					if currentTime < bestTime {
						isPR = true
						bestTimes[wodID] = currentTime
					}
				} else {
					// First time doing this WOD - it's a PR
					isPR = true
					bestTimes[wodID] = currentTime
				}
			}

			// Check rounds+reps PRs
			if wod.Rounds != nil {
				currentRounds := *wod.Rounds
				currentReps := 0
				if wod.Reps != nil {
					currentReps = *wod.Reps
				}

				if best, exists := bestRoundsReps[wodID]; exists {
					// Compare with previous best
					if currentRounds > best.rounds || (currentRounds == best.rounds && currentReps > best.reps) {
						isPR = true
						bestRoundsReps[wodID] = struct {
							rounds int
							reps   int
						}{currentRounds, currentReps}
					}
				} else {
					// First time doing this WOD - it's a PR
					isPR = true
					bestRoundsReps[wodID] = struct {
						rounds int
						reps   int
					}{currentRounds, currentReps}
				}
			}

			// Update PR flag if needed
			if isPR != wod.IsPR {
				if err := s.userWorkoutWODRepo.UpdatePRFlag(wod.ID, isPR); err != nil {
					return movementPRCount, wodPRCount, fmt.Errorf("failed to update PR flag for WOD %d: %w", wod.ID, err)
				}
				if isPR {
					wodPRCount++
				}
			}
		}
	}

	return movementPRCount, wodPRCount, nil
}

// ValidateWODScoreTypes validates that WOD performance data matches each WOD's defined score_type
func (s *UserWorkoutService) ValidateWODScoreTypes(wods []*domain.UserWorkoutWOD) error {
	for _, w := range wods {
		// Fetch the WOD definition
		wod, err := s.wodRepo.GetByID(w.WODID)
		if err != nil {
			return fmt.Errorf("failed to get WOD definition for WOD ID %d: %w", w.WODID, err)
		}
		if wod == nil {
			return fmt.Errorf("WOD with ID %d not found", w.WODID)
		}

		// Validate based on score_type
		scoreType := wod.ScoreType

		// Time-based WODs
		if scoreType == "Time (HH:MM:SS)" {
			// Must have time_seconds, must NOT have rounds/reps/weight
			if w.TimeSeconds == nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but time_seconds is missing", wod.Name, scoreType)
			}
			if w.Rounds != nil || w.Reps != nil || w.Weight != nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but contains invalid fields (rounds/reps/weight)", wod.Name, scoreType)
			}
		}

		// Rounds+Reps WODs
		if scoreType == "Rounds+Reps" {
			// Must have rounds, must NOT have time_seconds/weight
			if w.Rounds == nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but rounds is missing", wod.Name, scoreType)
			}
			if w.TimeSeconds != nil || w.Weight != nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but contains invalid fields (time_seconds/weight)", wod.Name, scoreType)
			}
		}

		// Max Weight WODs
		if scoreType == "Max Weight" {
			// Must have weight, must NOT have time_seconds/rounds/reps
			if w.Weight == nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but weight is missing", wod.Name, scoreType)
			}
			if w.TimeSeconds != nil || w.Rounds != nil || w.Reps != nil {
				return fmt.Errorf("WOD '%s' has score_type '%s' but contains invalid fields (time_seconds/rounds/reps)", wod.Name, scoreType)
			}
		}
	}

	return nil
}
