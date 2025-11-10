package domain

import (
	"time"
)

// MovementType represents the type of movement
type MovementType string

const (
	MovementTypeWeightlifting MovementType = "weightlifting"
	MovementTypeBodyweight    MovementType = "bodyweight"
	MovementTypeCardio        MovementType = "cardio"
	MovementTypeGymnastics    MovementType = "gymnastics"
)

// Movement represents a specific exercise or movement (strength_movements table)
// Note: After v0.4.0 migration, this maps to the 'strength_movements' table in the database
type Movement struct {
	ID          int64        `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description,omitempty" db:"description"`
	Type        MovementType `json:"type" db:"type"`
	IsStandard  bool         `json:"is_standard" db:"is_standard"`         // True for predefined movements
	CreatedBy   *int64       `json:"created_by,omitempty" db:"created_by"` // User ID if custom
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// WorkoutMovement represents a movement in a workout template (workout_strength table)
// Note: After v0.4.0 migration, this maps to the 'workout_strength' table in the database
// WorkoutID references a workout template, not a user-specific workout instance
type WorkoutMovement struct {
	ID         int64     `json:"id" db:"id"`
	WorkoutID  int64     `json:"workout_id" db:"workout_id"`     // References workout template
	MovementID int64     `json:"movement_id" db:"movement_id"`   // References strength_movements table (column may still be named movement_id in some DBs)
	Weight     *float64  `json:"weight,omitempty" db:"weight"` // in lbs or kg
	Sets       *int      `json:"sets,omitempty" db:"sets"`
	Reps       *int      `json:"reps,omitempty" db:"reps"`
	Time       *int      `json:"time,omitempty" db:"time"`         // in seconds
	Distance   *float64  `json:"distance,omitempty" db:"distance"` // in meters or miles
	IsRx       bool      `json:"is_rx" db:"is_rx"`                 // Prescribed weight/standard
	IsPR       bool      `json:"is_pr" db:"is_pr"`                 // Personal record flag
	Notes      string    `json:"notes,omitempty" db:"notes"`
	OrderIndex int       `json:"order_index" db:"order_index"` // Order in the workout
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	Movement *Movement `json:"movement,omitempty" db:"-"`
}

// MovementRepository defines the interface for movement data access
type MovementRepository interface {
	Create(movement *Movement) error
	GetByID(id int64) (*Movement, error)
	GetByName(name string) (*Movement, error)
	ListStandard() ([]*Movement, error)
	ListByUser(userID int64) ([]*Movement, error)
	Update(movement *Movement) error
	Delete(id int64) error
	Search(query string, limit int) ([]*Movement, error)
}

// PersonalRecord represents a user's personal record for a movement
// After v0.4.0, this aggregates data from user_workouts and workout_strength tables
type PersonalRecord struct {
	MovementID      int64     `json:"movement_id"`
	MovementName    string    `json:"movement_name"`
	MaxWeight       *float64  `json:"max_weight,omitempty"`
	MaxReps         *int      `json:"max_reps,omitempty"`
	BestTime        *int      `json:"best_time,omitempty"` // Fastest time in seconds
	UserWorkoutID   int64     `json:"user_workout_id"`     // References user_workouts table (logged workout instance)
	WorkoutID       int64     `json:"workout_id"`          // References workout template
	WorkoutDate     time.Time `json:"workout_date"`        // From user_workouts.workout_date
}

// WorkoutMovementRepository defines the interface for workout movement data access
type WorkoutMovementRepository interface {
	Create(wm *WorkoutMovement) error
	GetByID(id int64) (*WorkoutMovement, error)
	GetByWorkoutID(workoutID int64) ([]*WorkoutMovement, error)
	GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*WorkoutMovement, error)
	Update(wm *WorkoutMovement) error
	Delete(id int64) error
	DeleteByWorkoutID(workoutID int64) error
	// PR tracking methods
	GetPersonalRecords(userID int64) ([]*PersonalRecord, error)
	GetMaxWeightForMovement(userID, movementID int64) (*float64, error)
	GetPRMovements(userID int64, limit int) ([]*WorkoutMovement, error)
}
