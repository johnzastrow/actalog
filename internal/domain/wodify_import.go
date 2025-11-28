package domain

import "time"

// WodifyPerformanceRow represents a single row from the Wodify CSV export
type WodifyPerformanceRow struct {
	CustomerName              string
	LocationName              string
	Date                      string // MM/DD/YYYY format
	ProgramName               string
	ClassName                 string
	ComponentType             string // "Weightlifting", "Metcon", "Gymnastics"
	ComponentID               string
	ComponentName             string
	ComponentDescription      string
	PerformanceResultType     string // "Weight", "Time", "AMRAP - Rounds and Reps", etc.
	RepScheme                 string
	FullyFormattedResult      string
	FromWeightliftingTotal    bool
	FromVariableSet           bool
	IsRx                      bool
	IsRxPlus                  bool
	IsPersonalRecord          bool
	PersonalRecordDescription string
	Comment                   string
}

// WodifyImportPreview represents the preview of what will be imported
type WodifyImportPreview struct {
	TotalRows            int                    `json:"total_rows"`
	ValidRows            int                    `json:"valid_rows"`
	InvalidRows          int                    `json:"invalid_rows"`
	UniqueWorkoutDates   int                    `json:"unique_workout_dates"`
	MovementsToCreate    int                    `json:"movements_to_create"`
	WODsToCreate         int                    `json:"wods_to_create"`
	UserWorkoutsToCreate int                    `json:"user_workouts_to_create"`
	UserWorkoutsToUpdate int                    `json:"user_workouts_to_update"`
	PerformancesToCreate int                    `json:"performances_to_create"`
	Errors               []WodifyImportError    `json:"errors,omitempty"`
	WorkoutSummary       []WodifyWorkoutSummary `json:"workout_summary"`
	NewMovements         []string               `json:"new_movements"`
	NewWODs              []string               `json:"new_wods"`
}

// WodifyImportError represents an error in the import
type WodifyImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// WodifyWorkoutSummary represents a summary of a workout to be created
type WodifyWorkoutSummary struct {
	Date              string `json:"date"`
	MovementCount     int    `json:"movement_count"`
	WODCount          int    `json:"wod_count"`
	HasPRs            bool   `json:"has_prs"`
	ComponentTypes    string `json:"component_types"`
	ExistingWorkoutID *int64 `json:"existing_workout_id,omitempty"`
	IsUpdate          bool   `json:"is_update"`
}

// WodifyImportResult represents the result of the import
type WodifyImportResult struct {
	WorkoutsCreated     int `json:"workouts_created"`
	WorkoutsUpdated     int `json:"workouts_updated"`
	MovementsCreated    int `json:"movements_created"`
	WODsCreated         int `json:"wods_created"`
	PerformancesCreated int `json:"performances_created"`
	PerformancesUpdated int `json:"performances_updated"`
	PRsFlagged          int `json:"prs_flagged"`
}

// ParsedPerformanceResult represents a parsed performance result
type ParsedPerformanceResult struct {
	// For Weightlifting
	Sets   *int     `json:"sets,omitempty"`
	Reps   *int     `json:"reps,omitempty"`
	Weight *float64 `json:"weight,omitempty"`

	// For Metcons
	TimeSeconds *int     `json:"time_seconds,omitempty"`
	Rounds      *int     `json:"rounds,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	Distance    *float64 `json:"distance,omitempty"`

	// Notes
	Notes string `json:"notes,omitempty"`

	// PR flag
	IsPR bool `json:"is_pr"`
}

// WodifyGroupedWorkout represents performances grouped by date
type WodifyGroupedWorkout struct {
	Date         time.Time
	Performances []WodifyPerformanceRow
}
