package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// ExportService handles data export operations
type ExportService struct {
	wodRepo         domain.WODRepository
	movementRepo    domain.MovementRepository
	userRepo        domain.UserRepository
	userWorkoutRepo domain.UserWorkoutRepository
}

// NewExportService creates a new export service
func NewExportService(
	wodRepo domain.WODRepository,
	movementRepo domain.MovementRepository,
	userRepo domain.UserRepository,
	userWorkoutRepo domain.UserWorkoutRepository,
) *ExportService {
	return &ExportService{
		wodRepo:         wodRepo,
		movementRepo:    movementRepo,
		userRepo:        userRepo,
		userWorkoutRepo: userWorkoutRepo,
	}
}

// UserWorkoutExport represents the complete export structure for user workouts
type UserWorkoutExport struct {
	ExportMetadata ExportMetadata          `json:"export_metadata"`
	UserWorkouts   []UserWorkoutExportItem `json:"user_workouts"`
}

// ExportMetadata contains information about the export
type ExportMetadata struct {
	UserEmail   string     `json:"user_email"`
	ExportDate  time.Time  `json:"export_date"`
	DateRange   *DateRange `json:"date_range,omitempty"`
	Version     string     `json:"version"`
	TotalCount  int        `json:"total_count"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start string `json:"start"` // YYYY-MM-DD format
	End   string `json:"end"`   // YYYY-MM-DD format
}

// UserWorkoutExportItem represents a single workout in the export
type UserWorkoutExportItem struct {
	WorkoutDate string                       `json:"workout_date"`
	WorkoutType *string                      `json:"workout_type,omitempty"`
	WorkoutName *string                      `json:"workout_name,omitempty"`
	TotalTime   *int                         `json:"total_time,omitempty"`
	Notes       *string                      `json:"notes,omitempty"`
	Movements   []MovementPerformanceExport  `json:"movements,omitempty"`
	WODs        []WODPerformanceExport       `json:"wods,omitempty"`
}

// MovementPerformanceExport represents movement performance data
type MovementPerformanceExport struct {
	MovementName string   `json:"movement_name"`
	MovementType string   `json:"movement_type"`
	Sets         *int     `json:"sets,omitempty"`
	Reps         *int     `json:"reps,omitempty"`
	Weight       *float64 `json:"weight,omitempty"`
	Time         *int     `json:"time_seconds,omitempty"`
	Distance     *float64 `json:"distance,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	IsPR         bool     `json:"is_pr"`
	OrderIndex   int      `json:"order_index"`
}

// WODPerformanceExport represents WOD performance data
type WODPerformanceExport struct {
	WODName      string   `json:"wod_name"`
	WODType      string   `json:"wod_type"`
	ScoreType    *string  `json:"score_type,omitempty"`
	ScoreValue   *string  `json:"score_value,omitempty"`
	TimeSeconds  *int     `json:"time_seconds,omitempty"`
	Rounds       *int     `json:"rounds,omitempty"`
	Reps         *int     `json:"reps,omitempty"`
	Weight       *float64 `json:"weight,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	IsPR         bool     `json:"is_pr"`
	OrderIndex   int      `json:"order_index"`
}

// ExportWODsToCSV exports WODs to CSV format
// If isAdmin is true, exports all WODs. Otherwise, only exports standard WODs and user's custom WODs
func (s *ExportService) ExportWODsToCSV(userID int64, isAdmin bool, includeStandard, includeCustom bool) ([]byte, error) {
	var wods []*domain.WOD
	var err error

	// Fetch WODs based on permissions and filters
	if isAdmin && includeStandard && includeCustom {
		// Admin wants everything
		wods, err = s.wodRepo.List(nil, 10000, 0)
	} else if includeStandard && includeCustom {
		// User wants standard + their custom
		standardWods, err1 := s.wodRepo.ListStandard(10000, 0)
		if err1 != nil {
			return nil, fmt.Errorf("failed to fetch standard WODs: %w", err1)
		}
		customWods, err2 := s.wodRepo.ListByUser(userID, 10000, 0)
		if err2 != nil {
			return nil, fmt.Errorf("failed to fetch custom WODs: %w", err2)
		}
		wods = append(standardWods, customWods...)
	} else if includeStandard {
		// Only standard WODs
		wods, err = s.wodRepo.ListStandard(10000, 0)
	} else if includeCustom {
		// Only user's custom WODs
		wods, err = s.wodRepo.ListByUser(userID, 10000, 0)
	} else {
		// Nothing selected
		return nil, fmt.Errorf("must select at least one option: standard or custom")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch WODs: %w", err)
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"name", "source", "type", "regime", "score_type", "description", "url", "notes", "is_standard", "created_by_email"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write WOD rows
	for _, wod := range wods {
		var createdByEmail string
		if wod.CreatedBy != nil {
			// Fetch user email for custom WODs
			user, err := s.userRepo.GetByID(*wod.CreatedBy)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user for WOD %d: %w", wod.ID, err)
			}
			if user != nil {
				createdByEmail = user.Email
			}
		}

		url := ""
		if wod.URL != nil {
			url = *wod.URL
		}

		notes := ""
		if wod.Notes != nil {
			notes = *wod.Notes
		}

		row := []string{
			wod.Name,
			wod.Source,
			wod.Type,
			wod.Regime,
			wod.ScoreType,
			wod.Description,
			url,
			notes,
			strconv.FormatBool(wod.IsStandard),
			createdByEmail,
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write WOD row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportWODsToJSON exports WODs to JSON format
// If isAdmin is true, exports all WODs. Otherwise, only exports standard WODs and user's custom WODs
func (s *ExportService) ExportWODsToJSON(userID int64, isAdmin bool, includeStandard, includeCustom bool) ([]byte, error) {
	var wods []*domain.WOD
	var err error

	// Fetch WODs based on permissions and filters (same logic as CSV)
	if isAdmin && includeStandard && includeCustom {
		wods, err = s.wodRepo.List(nil, 10000, 0)
	} else if includeStandard && includeCustom {
		standardWods, err1 := s.wodRepo.ListStandard(10000, 0)
		if err1 != nil {
			return nil, fmt.Errorf("failed to fetch standard WODs: %w", err1)
		}
		customWods, err2 := s.wodRepo.ListByUser(userID, 10000, 0)
		if err2 != nil {
			return nil, fmt.Errorf("failed to fetch custom WODs: %w", err2)
		}
		wods = append(standardWods, customWods...)
	} else if includeStandard {
		wods, err = s.wodRepo.ListStandard(10000, 0)
	} else if includeCustom {
		wods, err = s.wodRepo.ListByUser(userID, 10000, 0)
	} else {
		return nil, fmt.Errorf("must select at least one option: standard or custom")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch WODs: %w", err)
	}

	// Enrich WODs with creator email
	type WODExport struct {
		*domain.WOD
		CreatedByEmail string `json:"created_by_email,omitempty"`
	}

	exports := make([]WODExport, 0, len(wods))
	for _, wod := range wods {
		export := WODExport{WOD: wod}
		if wod.CreatedBy != nil {
			user, err := s.userRepo.GetByID(*wod.CreatedBy)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user for WOD %d: %w", wod.ID, err)
			}
			if user != nil {
				export.CreatedByEmail = user.Email
			}
		}
		exports = append(exports, export)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(exports, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}

// ExportMovementsToCSV exports movements to CSV format
// If isAdmin is true, exports all movements. Otherwise, only exports standard movements and user's custom movements
func (s *ExportService) ExportMovementsToCSV(userID int64, isAdmin bool, includeStandard, includeCustom bool) ([]byte, error) {
	var movements []*domain.Movement
	var err error

	// Fetch movements based on permissions and filters
	if isAdmin && includeStandard && includeCustom {
		// Admin wants everything
		movements, err = s.movementRepo.ListAll()
	} else if includeStandard && includeCustom {
		// User wants standard + their custom
		standardMovements, err1 := s.movementRepo.ListStandard()
		if err1 != nil {
			return nil, fmt.Errorf("failed to fetch standard movements: %w", err1)
		}
		customMovements, err2 := s.movementRepo.ListByUser(userID)
		if err2 != nil {
			return nil, fmt.Errorf("failed to fetch custom movements: %w", err2)
		}
		movements = append(standardMovements, customMovements...)
	} else if includeStandard {
		// Only standard movements
		movements, err = s.movementRepo.ListStandard()
	} else if includeCustom {
		// Only user's custom movements
		movements, err = s.movementRepo.ListByUser(userID)
	} else {
		// Nothing selected
		return nil, fmt.Errorf("must select at least one option: standard or custom")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch movements: %w", err)
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"name", "type", "description", "is_standard", "created_by_email"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write movement rows
	for _, movement := range movements {
		var createdByEmail string
		if movement.CreatedBy != nil {
			// Fetch user email for custom movements
			user, err := s.userRepo.GetByID(*movement.CreatedBy)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user for movement %d: %w", movement.ID, err)
			}
			if user != nil {
				createdByEmail = user.Email
			}
		}

		row := []string{
			movement.Name,
			string(movement.Type),
			movement.Description,
			strconv.FormatBool(movement.IsStandard),
			createdByEmail,
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write movement row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportMovementsToJSON exports movements to JSON format
// If isAdmin is true, exports all movements. Otherwise, only exports standard movements and user's custom movements
func (s *ExportService) ExportMovementsToJSON(userID int64, isAdmin bool, includeStandard, includeCustom bool) ([]byte, error) {
	var movements []*domain.Movement
	var err error

	// Fetch movements based on permissions and filters (same logic as CSV)
	if isAdmin && includeStandard && includeCustom {
		movements, err = s.movementRepo.ListAll()
	} else if includeStandard && includeCustom {
		standardMovements, err1 := s.movementRepo.ListStandard()
		if err1 != nil {
			return nil, fmt.Errorf("failed to fetch standard movements: %w", err1)
		}
		customMovements, err2 := s.movementRepo.ListByUser(userID)
		if err2 != nil {
			return nil, fmt.Errorf("failed to fetch custom movements: %w", err2)
		}
		movements = append(standardMovements, customMovements...)
	} else if includeStandard {
		movements, err = s.movementRepo.ListStandard()
	} else if includeCustom {
		movements, err = s.movementRepo.ListByUser(userID)
	} else {
		return nil, fmt.Errorf("must select at least one option: standard or custom")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch movements: %w", err)
	}

	// Enrich movements with creator email
	type MovementExport struct {
		*domain.Movement
		CreatedByEmail string `json:"created_by_email,omitempty"`
	}

	exports := make([]MovementExport, 0, len(movements))
	for _, movement := range movements {
		export := MovementExport{Movement: movement}
		if movement.CreatedBy != nil {
			user, err := s.userRepo.GetByID(*movement.CreatedBy)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user for movement %d: %w", movement.ID, err)
			}
			if user != nil {
				export.CreatedByEmail = user.Email
			}
		}
		exports = append(exports, export)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(exports, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}

// ExportUserWorkoutsToJSON exports user workouts with full nested data to JSON format
// Supports optional date range filtering
func (s *ExportService) ExportUserWorkoutsToJSON(userID int64, startDate, endDate *time.Time) ([]byte, error) {
	// Fetch user for metadata
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Fetch workouts (with or without date range)
	var workouts []*domain.UserWorkout
	var dateRange *DateRange

	if startDate != nil && endDate != nil {
		workouts, err = s.userWorkoutRepo.ListByUserAndDateRange(userID, *startDate, *endDate)
		dateRange = &DateRange{
			Start: startDate.Format("2006-01-02"),
			End:   endDate.Format("2006-01-02"),
		}
	} else {
		workouts, err = s.userWorkoutRepo.ListByUser(userID, 10000, 0)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user workouts: %w", err)
	}

	// Build export structure
	exportData := UserWorkoutExport{
		ExportMetadata: ExportMetadata{
			UserEmail:  user.Email,
			ExportDate: time.Now(),
			DateRange:  dateRange,
			Version:    "0.5.1",
			TotalCount: len(workouts),
		},
		UserWorkouts: make([]UserWorkoutExportItem, 0, len(workouts)),
	}

	// Process each workout
	for _, workout := range workouts {
		// Get full workout details with performance data
		details, err := s.userWorkoutRepo.GetByIDWithDetails(workout.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch workout details for workout %d: %w", workout.ID, err)
		}

		if details == nil {
			continue
		}

		// Build workout export item
		item := UserWorkoutExportItem{
			WorkoutDate: details.WorkoutDate.Format("2006-01-02"),
			WorkoutType: details.WorkoutType,
			WorkoutName: &details.WorkoutName,
			TotalTime:   details.TotalTime,
			Notes:       details.Notes,
			Movements:   make([]MovementPerformanceExport, 0),
			WODs:        make([]WODPerformanceExport, 0),
		}

		// Add movement performance data
		for _, perfMovement := range details.PerformanceMovements {
			movementExport := MovementPerformanceExport{
				MovementName: perfMovement.MovementName,
				MovementType: perfMovement.MovementType,
				Sets:         perfMovement.Sets,
				Reps:         perfMovement.Reps,
				Weight:       perfMovement.Weight,
				Time:         perfMovement.Time,
				Distance:     perfMovement.Distance,
				Notes:        perfMovement.Notes,
				IsPR:         perfMovement.IsPR,
				OrderIndex:   perfMovement.OrderIndex,
			}
			item.Movements = append(item.Movements, movementExport)
		}

		// Add WOD performance data
		for _, perfWOD := range details.PerformanceWODs {
			wodExport := WODPerformanceExport{
				WODName:     perfWOD.WODName,
				WODType:     perfWOD.WODType,
				ScoreType:   perfWOD.ScoreType,
				ScoreValue:  perfWOD.ScoreValue,
				TimeSeconds: perfWOD.TimeSeconds,
				Rounds:      perfWOD.Rounds,
				Reps:        perfWOD.Reps,
				Weight:      perfWOD.Weight,
				Notes:       perfWOD.Notes,
				IsPR:        perfWOD.IsPR,
				OrderIndex:  perfWOD.OrderIndex,
			}
			item.WODs = append(item.WODs, wodExport)
		}

		exportData.UserWorkouts = append(exportData.UserWorkouts, item)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}

// ExportUserWorkoutsToCSV exports user workouts to CSV format (flattened structure)
// Supports optional date range filtering
func (s *ExportService) ExportUserWorkoutsToCSV(userID int64, startDate, endDate *time.Time) ([]byte, error) {
	// Fetch user for metadata
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Fetch workouts (with or without date range)
	var workouts []*domain.UserWorkout
	if startDate != nil && endDate != nil {
		workouts, err = s.userWorkoutRepo.ListByUserAndDateRange(userID, *startDate, *endDate)
	} else {
		workouts, err = s.userWorkoutRepo.ListByUser(userID, 10000, 0)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user workouts: %w", err)
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"workout_date",
		"workout_type",
		"workout_name",
		"total_time_seconds",
		"notes",
		"performance_type",
		"entity_name",
		"entity_type",
		"sets",
		"reps",
		"weight",
		"time_seconds",
		"distance",
		"rounds",
		"score_type",
		"score_value",
		"is_pr",
		"performance_notes",
		"order_index",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Process each workout
	for _, workout := range workouts {
		// Get full workout details with performance data
		details, err := s.userWorkoutRepo.GetByIDWithDetails(workout.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch workout details for workout %d: %w", workout.ID, err)
		}

		if details == nil {
			continue
		}

		// Helper to format optional values
		formatInt := func(v *int) string {
			if v == nil {
				return ""
			}
			return strconv.Itoa(*v)
		}
		formatFloat := func(v *float64) string {
			if v == nil {
				return ""
			}
			return fmt.Sprintf("%.2f", *v)
		}
		formatString := func(v *string) string {
			if v == nil {
				return ""
			}
			return *v
		}

		workoutDate := details.WorkoutDate.Format("2006-01-02")
		workoutType := formatString(details.WorkoutType)
		workoutName := details.WorkoutName
		totalTime := formatInt(details.TotalTime)
		workoutNotes := formatString(details.Notes)

		// Export movement performances
		for _, perfMovement := range details.PerformanceMovements {
			row := []string{
				workoutDate,
				workoutType,
				workoutName,
				totalTime,
				workoutNotes,
				"movement",                     // performance_type
				perfMovement.MovementName,      // entity_name
				perfMovement.MovementType,      // entity_type
				formatInt(perfMovement.Sets),   // sets
				formatInt(perfMovement.Reps),   // reps
				formatFloat(perfMovement.Weight), // weight
				formatInt(perfMovement.Time),   // time_seconds
				formatFloat(perfMovement.Distance), // distance
				"",                             // rounds (n/a for movements)
				"",                             // score_type (n/a for movements)
				"",                             // score_value (n/a for movements)
				strconv.FormatBool(perfMovement.IsPR), // is_pr
				perfMovement.Notes,             // performance_notes
				strconv.Itoa(perfMovement.OrderIndex), // order_index
			}
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write movement row: %w", err)
			}
		}

		// Export WOD performances
		for _, perfWOD := range details.PerformanceWODs {
			row := []string{
				workoutDate,
				workoutType,
				workoutName,
				totalTime,
				workoutNotes,
				"wod",                          // performance_type
				perfWOD.WODName,                // entity_name
				perfWOD.WODType,                // entity_type
				"",                             // sets (n/a for WODs)
				formatInt(perfWOD.Reps),        // reps
				formatFloat(perfWOD.Weight),    // weight
				formatInt(perfWOD.TimeSeconds), // time_seconds
				"",                             // distance (n/a for WODs)
				formatInt(perfWOD.Rounds),      // rounds
				formatString(perfWOD.ScoreType), // score_type
				formatString(perfWOD.ScoreValue), // score_value
				strconv.FormatBool(perfWOD.IsPR), // is_pr
				perfWOD.Notes,                  // performance_notes
				strconv.Itoa(perfWOD.OrderIndex), // order_index
			}
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write WOD row: %w", err)
			}
		}

		// If workout has no performances, write a row with just workout info
		if len(details.PerformanceMovements) == 0 && len(details.PerformanceWODs) == 0 {
			row := []string{
				workoutDate,
				workoutType,
				workoutName,
				totalTime,
				workoutNotes,
				"",  // performance_type
				"",  // entity_name
				"",  // entity_type
				"",  // sets
				"",  // reps
				"",  // weight
				"",  // time_seconds
				"",  // distance
				"",  // rounds
				"",  // score_type
				"",  // score_value
				"",  // is_pr
				"",  // performance_notes
				"",  // order_index
			}
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write workout row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}
