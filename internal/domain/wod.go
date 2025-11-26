package domain

import "time"

// WOD represents a Workout of the Day (CrossFit benchmark workout)
// WODs are predefined workouts like "Fran", "Murph", "Helen", etc.
// Standard WODs are pre-seeded, users can also create custom WODs
type WOD struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Source      string     `json:"source,omitempty" db:"source"`           // CrossFit, Other Coach, Self-recorded
	Type        string     `json:"type,omitempty" db:"type"`               // Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
	Regime      string     `json:"regime,omitempty" db:"regime"`           // EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
	ScoreType   string     `json:"score_type,omitempty" db:"score_type"`   // Time (HH:MM:SS), Rounds+Reps, Max Weight
	Description string     `json:"description,omitempty" db:"description"` // Full WOD description/instructions
	URL         *string    `json:"url,omitempty" db:"url"`                 // Optional video or reference URL
	Notes       *string    `json:"notes,omitempty" db:"notes"`             // Additional notes
	IsStandard  bool       `json:"is_standard" db:"is_standard"`           // TRUE for pre-seeded WODs, FALSE for user-created
	CreatedBy   *int64     `json:"created_by,omitempty" db:"created_by"`   // User ID if custom WOD (NULL for standard)
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// UserWorkoutWOD represents a WOD's performance in a logged workout (user_workout_wods table)
// This stores the actual performance data when a user logs a WOD
type UserWorkoutWOD struct {
	ID            int64     `json:"id" db:"id"`
	UserWorkoutID int64     `json:"user_workout_id" db:"user_workout_id"` // References user_workouts (logged workout instance)
	WODID         int64     `json:"wod_id" db:"wod_id"`                   // References wods table
	ScoreType     *string   `json:"score_type,omitempty" db:"score_type"` // Time (HH:MM:SS), Rounds+Reps, Max Weight
	ScoreValue    *string   `json:"score_value,omitempty" db:"score_value"` // Formatted score (e.g., "12:34", "10+15", "225.5")
	TimeSeconds   *int      `json:"time_seconds,omitempty" db:"time_seconds"` // For Time-based WODs
	Rounds        *int      `json:"rounds,omitempty" db:"rounds"` // For AMRAP WODs
	Reps          *int      `json:"reps,omitempty" db:"reps"` // Remaining reps in AMRAP
	Weight        *float64  `json:"weight,omitempty" db:"weight"` // For Max Weight WODs
	Notes         string    `json:"notes,omitempty" db:"notes"`
	IsPR          bool      `json:"is_pr" db:"is_pr"` // Personal record flag
	OrderIndex    int       `json:"order_index" db:"order_index"` // Order in the workout
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// Related data (loaded via joins)
	WOD            *WOD      `json:"wod,omitempty" db:"-"`
	WODName        string    `json:"wod_name,omitempty" db:"-"`        // Flattened for convenience
	WODType        string    `json:"wod_type,omitempty" db:"-"`        // Flattened for convenience (Benchmark, Hero, Girl, etc.)
	WODScoreType   string    `json:"wod_score_type,omitempty" db:"-"`  // WOD's defined score_type from wods table
	WorkoutDate    time.Time `json:"workout_date" db:"-"`              // From user_workouts.workout_date
}

// WODRepository defines the interface for WOD data access
type WODRepository interface {
	// Create creates a new custom WOD
	Create(wod *WOD) error

	// GetByID retrieves a WOD by ID
	GetByID(id int64) (*WOD, error)

	// GetByName retrieves a WOD by name
	GetByName(name string) (*WOD, error)

	// List retrieves all WODs with optional filtering and pagination
	List(filters map[string]interface{}, limit, offset int) ([]*WOD, error)

	// ListStandard retrieves all standard (pre-seeded) WODs with pagination
	ListStandard(limit, offset int) ([]*WOD, error)

	// ListByUser retrieves all custom WODs created by a specific user with pagination
	ListByUser(userID int64, limit, offset int) ([]*WOD, error)

	// Update updates an existing WOD (only for user-created WODs)
	Update(wod *WOD) error

	// UpdateStandard updates an existing standard WOD (for admin import)
	UpdateStandard(wod *WOD) error

	// Delete deletes a WOD (only for user-created WODs)
	Delete(id int64) error

	// Search searches WODs by name (partial match) with limit
	Search(query string, limit int) ([]*WOD, error)
}

// UserWorkoutWODRepository defines the interface for user workout WOD performance data
type UserWorkoutWODRepository interface {
	// Create creates a new user workout WOD performance record
	Create(uww *UserWorkoutWOD) error

	// CreateBatch creates multiple user workout WOD records at once
	CreateBatch(wods []*UserWorkoutWOD) error

	// GetByID retrieves a user workout WOD by ID
	GetByID(id int64) (*UserWorkoutWOD, error)

	// GetByUserWorkoutID retrieves all WODs for a specific logged workout
	GetByUserWorkoutID(userWorkoutID int64) ([]*UserWorkoutWOD, error)

	// Update updates an existing user workout WOD
	Update(uww *UserWorkoutWOD) error

	// Delete deletes a user workout WOD
	Delete(id int64) error

	// DeleteByUserWorkoutID deletes all WODs for a logged workout
	DeleteByUserWorkoutID(userWorkoutID int64) error

	// GetBestTimeForWOD retrieves the fastest time for a specific WOD for a user
	GetBestTimeForWOD(userID, wodID int64) (*int, error)

	// GetBestRoundsRepsForWOD retrieves the best rounds+reps for a specific WOD for a user
	GetBestRoundsRepsForWOD(userID, wodID int64) (rounds *int, reps *int, err error)

	// GetPRWODs retrieves recent PR-flagged WODs for a user
	GetPRWODs(userID int64, limit int) ([]*UserWorkoutWOD, error)

	// UpdatePRFlag updates the is_pr flag for a user workout WOD
	UpdatePRFlag(id int64, isPR bool) error
}
