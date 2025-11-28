package domain

import "time"

// UserWorkout represents a user's logged instance of a workout template or ad-hoc workout
// This is the junction table between users and workouts, with the date and user-specific data
type UserWorkout struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	WorkoutID   *int64    `json:"workout_id,omitempty" db:"workout_id"`     // References the workout template (NULL for ad-hoc workouts)
	WorkoutName *string   `json:"workout_name,omitempty" db:"workout_name"` // Name for ad-hoc workouts (used when WorkoutID is NULL)
	WorkoutDate time.Time `json:"workout_date" db:"workout_date"`           // Date the workout was performed
	WorkoutType *string   `json:"workout_type,omitempty" db:"workout_type"` // strength, metcon, cardio, mixed
	TotalTime   *int      `json:"total_time,omitempty" db:"total_time"`     // Total workout duration in seconds
	Notes       *string   `json:"notes,omitempty" db:"notes"`               // User's notes for this specific workout instance
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserWorkoutWithDetails includes the workout template details and performance data
type UserWorkoutWithDetails struct {
	UserWorkout
	WorkoutName          string                   `json:"workout_name"`
	WorkoutDescription   *string                  `json:"workout_description,omitempty"`
	Movements            []*WorkoutMovement       `json:"movements,omitempty"`             // Template movements
	WODs                 []*WorkoutWODWithDetails `json:"wods,omitempty"`                  // Template WODs
	PerformanceMovements []*UserWorkoutMovement   `json:"performance_movements,omitempty"` // Actual performance data for movements
	PerformanceWODs      []*UserWorkoutWOD        `json:"performance_wods,omitempty"`      // Actual performance data for WODs
}

// UserWorkoutRepository defines the interface for user workout data access
type UserWorkoutRepository interface {
	// Create creates a new user workout (logs a workout instance)
	Create(userWorkout *UserWorkout) error

	// GetByID retrieves a user workout by ID
	GetByID(id int64) (*UserWorkout, error)

	// GetByIDWithDetails retrieves a user workout with full details (movements, WODs)
	GetByIDWithDetails(id int64, userID int64) (*UserWorkoutWithDetails, error)

	// ListByUser retrieves all workouts logged by a specific user
	ListByUser(userID int64, limit, offset int) ([]*UserWorkout, error)

	// ListByUserWithDetails retrieves all workouts logged by a user with details
	ListByUserWithDetails(userID int64, limit, offset int) ([]*UserWorkoutWithDetails, error)

	// ListByUserAndDateRange retrieves workouts within a date range
	ListByUserAndDateRange(userID int64, startDate, endDate time.Time) ([]*UserWorkout, error)

	// Update updates an existing user workout
	Update(userWorkout *UserWorkout) error

	// Delete deletes a user workout
	Delete(id int64, userID int64) error

	// GetByUserWorkoutDate checks if a user has already logged a specific workout on a date
	GetByUserWorkoutDate(userID, workoutID int64, date time.Time) (*UserWorkout, error)
}
