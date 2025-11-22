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

// Movement represents a specific exercise or movement (movements table)
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

// WorkoutMovement represents a movement in a workout template (workout_movements table)
// WorkoutID references a workout template, not a user-specific workout instance
type WorkoutMovement struct {
	ID         int64     `json:"id" db:"id"`
	WorkoutID  int64     `json:"workout_id" db:"workout_id"`     // References workout template
	MovementID int64     `json:"movement_id" db:"movement_id"`   // References movements table
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
	ListAll() ([]*Movement, error)
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

// UserWorkoutMovement represents a movement's performance in a logged workout (user_workout_movements table)
// This stores the actual performance data when a user logs a workout
type UserWorkoutMovement struct {
	ID            int64     `json:"id" db:"id"`
	UserWorkoutID int64     `json:"user_workout_id" db:"user_workout_id"` // References user_workouts (logged workout instance)
	MovementID    int64     `json:"movement_id" db:"movement_id"`         // References movements table
	Sets          *int      `json:"sets,omitempty" db:"sets"`
	Reps          *int      `json:"reps,omitempty" db:"reps"`
	Weight        *float64  `json:"weight,omitempty" db:"weight"`     // in lbs or kg
	Time          *int      `json:"time_seconds,omitempty" db:"time"`         // in seconds
	Distance      *float64  `json:"distance,omitempty" db:"distance"` // in meters or miles
	Notes         string    `json:"notes,omitempty" db:"notes"`
	IsPR          bool      `json:"is_pr" db:"is_pr"` // Personal record flag
	OrderIndex    int       `json:"order_index" db:"order_index"` // Order in the workout
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// Related data (loaded via joins)
	Movement     *Movement `json:"movement,omitempty" db:"-"`
	MovementName string    `json:"movement_name,omitempty" db:"-"` // Flattened for convenience
	MovementType string    `json:"movement_type,omitempty" db:"-"` // Flattened for convenience
	WorkoutDate  time.Time `json:"workout_date" db:"-"`            // From user_workouts.workout_date
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

// UserWorkoutMovementRepository defines the interface for user workout movement performance data
type UserWorkoutMovementRepository interface {
	// Create creates a new user workout movement performance record
	Create(uwm *UserWorkoutMovement) error

	// CreateBatch creates multiple user workout movement records at once
	CreateBatch(movements []*UserWorkoutMovement) error

	// GetByID retrieves a user workout movement by ID
	GetByID(id int64) (*UserWorkoutMovement, error)

	// GetByUserWorkoutID retrieves all movements for a specific logged workout
	GetByUserWorkoutID(userWorkoutID int64) ([]*UserWorkoutMovement, error)

	// Update updates an existing user workout movement
	Update(uwm *UserWorkoutMovement) error

	// Delete deletes a user workout movement
	Delete(id int64) error

	// DeleteByUserWorkoutID deletes all movements for a logged workout
	DeleteByUserWorkoutID(userWorkoutID int64) error

	// GetMaxWeightForMovement retrieves the maximum weight for a specific movement for a user
	GetMaxWeightForMovement(userID, movementID int64) (*float64, error)

	// GetPRMovements retrieves recent PR-flagged movements for a user
	GetPRMovements(userID int64, limit int) ([]*UserWorkoutMovement, error)

	// UpdatePRFlag updates the is_pr flag for a user workout movement
	UpdatePRFlag(id int64, isPR bool) error
}
