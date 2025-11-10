package domain

import "time"

// WorkoutWOD represents the junction between a workout template and a WOD
// A workout template can contain multiple WODs
type WorkoutWOD struct {
	ID         int64     `json:"id" db:"id"`
	WorkoutID  int64     `json:"workout_id" db:"workout_id"` // References workout template
	WODID      int64     `json:"wod_id" db:"wod_id"`         // References WOD
	ScoreValue *string   `json:"score_value,omitempty" db:"score_value"` // Actual score when logged (time, rounds+reps, weight)
	Division   *string   `json:"division,omitempty" db:"division"`       // rx, scaled, beginner
	IsPR       bool      `json:"is_pr" db:"is_pr"`                       // Personal record flag
	OrderIndex int       `json:"order_index" db:"order_index"`           // Order in workout sequence
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// WorkoutWODWithDetails includes WOD details
type WorkoutWODWithDetails struct {
	WorkoutWOD
	WODName        string  `json:"wod_name"`
	WODType        string  `json:"wod_type"`
	WODRegime      string  `json:"wod_regime"`
	WODScoreType   string  `json:"wod_score_type"`
	WODDescription string  `json:"wod_description"`
}

// WorkoutWODRepository defines the interface for workout-WOD junction data access
type WorkoutWODRepository interface {
	// Create creates a new workout-WOD association
	Create(workoutWOD *WorkoutWOD) error

	// GetByID retrieves a workout-WOD by ID
	GetByID(id int64) (*WorkoutWOD, error)

	// ListByWorkout retrieves all WODs associated with a workout template
	ListByWorkout(workoutID int64) ([]*WorkoutWOD, error)

	// ListByWorkoutWithDetails retrieves WODs with full WOD details
	ListByWorkoutWithDetails(workoutID int64) ([]*WorkoutWODWithDetails, error)

	// Update updates an existing workout-WOD association
	Update(workoutWOD *WorkoutWOD) error

	// Delete deletes a workout-WOD association
	Delete(id int64) error

	// DeleteByWorkout deletes all WOD associations for a workout
	DeleteByWorkout(workoutID int64) error

	// TogglePR toggles the PR flag for a workout-WOD
	TogglePR(id int64) error
}
