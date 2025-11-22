package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// ImportService handles data import operations
type ImportService struct {
	wodRepo                 domain.WODRepository
	movementRepo            domain.MovementRepository
	userRepo                domain.UserRepository
	userWorkoutRepo         domain.UserWorkoutRepository
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository
	userWorkoutWODRepo      domain.UserWorkoutWODRepository
}

// NewImportService creates a new import service
func NewImportService(
	wodRepo domain.WODRepository,
	movementRepo domain.MovementRepository,
	userRepo domain.UserRepository,
	userWorkoutRepo domain.UserWorkoutRepository,
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository,
	userWorkoutWODRepo domain.UserWorkoutWODRepository,
) *ImportService {
	return &ImportService{
		wodRepo:                 wodRepo,
		movementRepo:            movementRepo,
		userRepo:                userRepo,
		userWorkoutRepo:         userWorkoutRepo,
		userWorkoutMovementRepo: userWorkoutMovementRepo,
		userWorkoutWODRepo:      userWorkoutWODRepo,
	}
}

// WODImportResult represents the result of a WOD import operation
type WODImportResult struct {
	TotalRows      int                    `json:"total_rows"`
	ValidRows      int                    `json:"valid_rows"`
	InvalidRows    int                    `json:"invalid_rows"`
	DuplicateRows  int                    `json:"duplicate_rows"`
	CreatedCount   int                    `json:"created_count"`
	UpdatedCount   int                    `json:"updated_count"`
	SkippedCount   int                    `json:"skipped_count"`
	Rows           []WODImportRow         `json:"rows"`
	ValidationErrors []string             `json:"validation_errors,omitempty"`
}

// WODImportRow represents a single WOD row during import
type WODImportRow struct {
	RowNumber       int      `json:"row_number"`
	Name            string   `json:"name"`
	Source          string   `json:"source"`
	Type            string   `json:"type"`
	Regime          string   `json:"regime"`
	ScoreType       string   `json:"score_type"`
	Description     string   `json:"description"`
	URL             string   `json:"url"`
	Notes           string   `json:"notes"`
	IsStandard      bool     `json:"is_standard"`
	CreatedByEmail  string   `json:"created_by_email"`
	IsValid         bool     `json:"is_valid"`
	IsDuplicate     bool     `json:"is_duplicate"`
	Errors          []string `json:"errors,omitempty"`
}

// MovementImportResult represents the result of a movement import operation
type MovementImportResult struct {
	TotalRows        int                    `json:"total_rows"`
	ValidRows        int                    `json:"valid_rows"`
	InvalidRows      int                    `json:"invalid_rows"`
	DuplicateRows    int                    `json:"duplicate_rows"`
	CreatedCount     int                    `json:"created_count"`
	UpdatedCount     int                    `json:"updated_count"`
	SkippedCount     int                    `json:"skipped_count"`
	Rows             []MovementImportRow    `json:"rows"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
}

// MovementImportRow represents a single movement row during import
type MovementImportRow struct {
	RowNumber      int      `json:"row_number"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	IsStandard     bool     `json:"is_standard"`
	CreatedByEmail string   `json:"created_by_email"`
	IsValid        bool     `json:"is_valid"`
	IsDuplicate    bool     `json:"is_duplicate"`
	Errors         []string `json:"errors,omitempty"`
}

// Valid enum values for WODs
var (
	validSources    = []string{"CrossFit", "Other Coach", "Self-recorded"}
	validTypes      = []string{"Benchmark", "Hero", "Girl", "Notables", "Games", "Endurance", "Self-created"}
	validRegimes    = []string{"EMOM", "AMRAP", "Fastest Time", "Slowest Round", "Get Stronger", "Skills"}
	validScoreTypes = []string{"Time (HH:MM:SS)", "Rounds+Reps", "Max Weight"}
	validMovementTypes = []string{"weightlifting", "bodyweight", "cardio", "gymnastics"}
)

// PreviewWODImport validates and previews WOD CSV data without saving
func (s *ImportService) PreviewWODImport(csvData io.Reader, userID int64, isAdmin bool) (*WODImportResult, error) {
	reader := csv.NewReader(csvData)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeader := []string{"name", "source", "type", "regime", "score_type", "description", "url", "notes", "is_standard", "created_by_email"}
	if !equalStringSlices(header, expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header. Expected: %v, Got: %v", expectedHeader, header)
	}

	result := &WODImportResult{
		Rows: []WODImportRow{},
	}

	rowNumber := 1 // Start at 1 (header is 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNumber+1, err)
		}

		rowNumber++
		row := s.parseWODRow(record, rowNumber)

		// Validate the row
		s.validateWODRow(&row, userID, isAdmin)

		// Check for duplicate by name
		existingWOD, err := s.wodRepo.GetByName(row.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check for duplicate WOD: %w", err)
		}
		if existingWOD != nil {
			row.IsDuplicate = true
			result.DuplicateRows++
		}

		if row.IsValid {
			result.ValidRows++
		} else {
			result.InvalidRows++
		}

		result.Rows = append(result.Rows, row)
		result.TotalRows++
	}

	return result, nil
}

// ConfirmWODImport actually imports WOD data after preview
func (s *ImportService) ConfirmWODImport(csvData io.Reader, userID int64, isAdmin bool, skipDuplicates, updateDuplicates bool) (*WODImportResult, error) {
	// First, run preview to validate
	preview, err := s.PreviewWODImport(csvData, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Process each valid row
	for _, row := range preview.Rows {
		if !row.IsValid {
			preview.SkippedCount++
			continue
		}

		if row.IsDuplicate {
			if skipDuplicates {
				preview.SkippedCount++
				continue
			}
			if updateDuplicates {
				// Update existing WOD
				existingWOD, err := s.wodRepo.GetByName(row.Name)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch existing WOD: %w", err)
				}

				// Update fields
				existingWOD.Source = row.Source
				existingWOD.Type = row.Type
				existingWOD.Regime = row.Regime
				existingWOD.ScoreType = row.ScoreType
				existingWOD.Description = row.Description
				if row.URL != "" {
					existingWOD.URL = &row.URL
				}
				if row.Notes != "" {
					existingWOD.Notes = &row.Notes
				}

				if err := s.wodRepo.Update(existingWOD); err != nil {
					return nil, fmt.Errorf("failed to update WOD: %w", err)
				}
				preview.UpdatedCount++
			} else {
				preview.SkippedCount++
			}
			continue
		}

		// Create new WOD
		wod := &domain.WOD{
			Name:        row.Name,
			Source:      row.Source,
			Type:        row.Type,
			Regime:      row.Regime,
			ScoreType:   row.ScoreType,
			Description: row.Description,
			IsStandard:  row.IsStandard,
		}

		if row.URL != "" {
			wod.URL = &row.URL
		}
		if row.Notes != "" {
			wod.Notes = &row.Notes
		}

		// Handle created_by
		if row.CreatedByEmail != "" {
			user, err := s.userRepo.GetByEmail(row.CreatedByEmail)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user by email: %w", err)
			}
			if user != nil {
				wod.CreatedBy = &user.ID
			}
		} else if !row.IsStandard {
			// If not standard and no email, use current user
			wod.CreatedBy = &userID
		}

		if err := s.wodRepo.Create(wod); err != nil {
			return nil, fmt.Errorf("failed to create WOD: %w", err)
		}
		preview.CreatedCount++
	}

	return preview, nil
}

// PreviewMovementImport validates and previews movement CSV data without saving
func (s *ImportService) PreviewMovementImport(csvData io.Reader, userID int64, isAdmin bool) (*MovementImportResult, error) {
	reader := csv.NewReader(csvData)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeader := []string{"name", "type", "description", "is_standard", "created_by_email"}
	if !equalStringSlices(header, expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header. Expected: %v, Got: %v", expectedHeader, header)
	}

	result := &MovementImportResult{
		Rows: []MovementImportRow{},
	}

	rowNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNumber+1, err)
		}

		rowNumber++
		row := s.parseMovementRow(record, rowNumber)

		// Validate the row
		s.validateMovementRow(&row, userID, isAdmin)

		// Check for duplicate by name
		existingMovement, err := s.movementRepo.GetByName(row.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check for duplicate movement: %w", err)
		}
		if existingMovement != nil {
			row.IsDuplicate = true
			result.DuplicateRows++
		}

		if row.IsValid {
			result.ValidRows++
		} else {
			result.InvalidRows++
		}

		result.Rows = append(result.Rows, row)
		result.TotalRows++
	}

	return result, nil
}

// ConfirmMovementImport actually imports movement data after preview
func (s *ImportService) ConfirmMovementImport(csvData io.Reader, userID int64, isAdmin bool, skipDuplicates, updateDuplicates bool) (*MovementImportResult, error) {
	// First, run preview to validate
	preview, err := s.PreviewMovementImport(csvData, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Process each valid row
	for _, row := range preview.Rows {
		if !row.IsValid {
			preview.SkippedCount++
			continue
		}

		if row.IsDuplicate {
			if skipDuplicates {
				preview.SkippedCount++
				continue
			}
			if updateDuplicates {
				// Update existing movement
				existingMovement, err := s.movementRepo.GetByName(row.Name)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch existing movement: %w", err)
				}

				// Update fields
				existingMovement.Type = domain.MovementType(row.Type)
				existingMovement.Description = row.Description

				if err := s.movementRepo.Update(existingMovement); err != nil {
					return nil, fmt.Errorf("failed to update movement: %w", err)
				}
				preview.UpdatedCount++
			} else {
				preview.SkippedCount++
			}
			continue
		}

		// Create new movement
		movement := &domain.Movement{
			Name:        row.Name,
			Type:        domain.MovementType(row.Type),
			Description: row.Description,
			IsStandard:  row.IsStandard,
		}

		// Handle created_by
		if row.CreatedByEmail != "" {
			user, err := s.userRepo.GetByEmail(row.CreatedByEmail)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch user by email: %w", err)
			}
			if user != nil {
				movement.CreatedBy = &user.ID
			}
		} else if !row.IsStandard {
			// If not standard and no email, use current user
			movement.CreatedBy = &userID
		}

		if err := s.movementRepo.Create(movement); err != nil {
			return nil, fmt.Errorf("failed to create movement: %w", err)
		}
		preview.CreatedCount++
	}

	return preview, nil
}

// Helper functions

func (s *ImportService) parseWODRow(record []string, rowNumber int) WODImportRow {
	row := WODImportRow{
		RowNumber:      rowNumber,
		Name:           strings.TrimSpace(record[0]),
		Source:         strings.TrimSpace(record[1]),
		Type:           strings.TrimSpace(record[2]),
		Regime:         strings.TrimSpace(record[3]),
		ScoreType:      strings.TrimSpace(record[4]),
		Description:    strings.TrimSpace(record[5]),
		URL:            strings.TrimSpace(record[6]),
		Notes:          strings.TrimSpace(record[7]),
		CreatedByEmail: strings.TrimSpace(record[9]),
		IsValid:        true,
		Errors:         []string{},
	}

	// Parse boolean
	isStandard := strings.ToLower(strings.TrimSpace(record[8]))
	row.IsStandard = isStandard == "true" || isStandard == "1"

	return row
}

func (s *ImportService) parseMovementRow(record []string, rowNumber int) MovementImportRow {
	row := MovementImportRow{
		RowNumber:      rowNumber,
		Name:           strings.TrimSpace(record[0]),
		Type:           strings.TrimSpace(record[1]),
		Description:    strings.TrimSpace(record[2]),
		CreatedByEmail: strings.TrimSpace(record[4]),
		IsValid:        true,
		Errors:         []string{},
	}

	// Parse boolean
	isStandard := strings.ToLower(strings.TrimSpace(record[3]))
	row.IsStandard = isStandard == "true" || isStandard == "1"

	return row
}

func (s *ImportService) validateWODRow(row *WODImportRow, userID int64, isAdmin bool) {
	// Validate required fields
	if row.Name == "" {
		row.Errors = append(row.Errors, "name is required")
		row.IsValid = false
	}
	if row.Source == "" {
		row.Errors = append(row.Errors, "source is required")
		row.IsValid = false
	}
	if row.Type == "" {
		row.Errors = append(row.Errors, "type is required")
		row.IsValid = false
	}
	if row.Regime == "" {
		row.Errors = append(row.Errors, "regime is required")
		row.IsValid = false
	}
	if row.ScoreType == "" {
		row.Errors = append(row.Errors, "score_type is required")
		row.IsValid = false
	}

	// Validate enum values
	if !contains(validSources, row.Source) {
		row.Errors = append(row.Errors, fmt.Sprintf("invalid source: %s (must be one of: %v)", row.Source, validSources))
		row.IsValid = false
	}
	if !contains(validTypes, row.Type) {
		row.Errors = append(row.Errors, fmt.Sprintf("invalid type: %s (must be one of: %v)", row.Type, validTypes))
		row.IsValid = false
	}
	if !contains(validRegimes, row.Regime) {
		row.Errors = append(row.Errors, fmt.Sprintf("invalid regime: %s (must be one of: %v)", row.Regime, validRegimes))
		row.IsValid = false
	}
	if !contains(validScoreTypes, row.ScoreType) {
		row.Errors = append(row.Errors, fmt.Sprintf("invalid score_type: %s (must be one of: %v)", row.ScoreType, validScoreTypes))
		row.IsValid = false
	}

	// Check permissions for standard WODs
	if row.IsStandard && !isAdmin {
		row.Errors = append(row.Errors, "only admins can import standard WODs")
		row.IsValid = false
	}
}

func (s *ImportService) validateMovementRow(row *MovementImportRow, userID int64, isAdmin bool) {
	// Validate required fields
	if row.Name == "" {
		row.Errors = append(row.Errors, "name is required")
		row.IsValid = false
	}
	if row.Type == "" {
		row.Errors = append(row.Errors, "type is required")
		row.IsValid = false
	}

	// Validate enum values
	if !contains(validMovementTypes, row.Type) {
		row.Errors = append(row.Errors, fmt.Sprintf("invalid type: %s (must be one of: %v)", row.Type, validMovementTypes))
		row.IsValid = false
	}

	// Check permissions for standard movements
	if row.IsStandard && !isAdmin {
		row.Errors = append(row.Errors, "only admins can import standard movements")
		row.IsValid = false
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// UserWorkoutImportResult represents the result of a user workout import operation
type UserWorkoutImportResult struct {
	TotalWorkouts    int      `json:"total_workouts"`
	ValidWorkouts    int      `json:"valid_workouts"`
	InvalidWorkouts  int      `json:"invalid_workouts"`
	DuplicateWorkouts int     `json:"duplicate_workouts"`
	CreatedCount     int      `json:"created_count"`
	UpdatedCount     int      `json:"updated_count"`
	SkippedCount     int      `json:"skipped_count"`
	MovementsCreated int      `json:"movements_created"`
	WODsCreated      int      `json:"wods_created"`
	Errors           []string `json:"errors,omitempty"`
}

// PreviewUserWorkoutImport validates and previews user workout JSON import
func (s *ImportService) PreviewUserWorkoutImport(jsonData []byte, userID int64) (*UserWorkoutImportResult, error) {
	// Parse JSON into export structure
	var exportData struct {
		ExportMetadata struct {
			UserEmail  string `json:"user_email"`
			Version    string `json:"version"`
			TotalCount int    `json:"total_count"`
		} `json:"export_metadata"`
		UserWorkouts []struct {
			WorkoutDate string                       `json:"workout_date"`
			WorkoutType *string                      `json:"workout_type,omitempty"`
			WorkoutName *string                      `json:"workout_name,omitempty"`
			TotalTime   *int                         `json:"total_time,omitempty"`
			Notes       *string                      `json:"notes,omitempty"`
			Movements   []map[string]interface{}     `json:"movements,omitempty"`
			WODs        []map[string]interface{}     `json:"wods,omitempty"`
		} `json:"user_workouts"`
	}

	if err := json.Unmarshal(jsonData, &exportData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &UserWorkoutImportResult{
		TotalWorkouts: len(exportData.UserWorkouts),
		Errors:        []string{},
	}

	// Validate each workout
	for _, workout := range exportData.UserWorkouts {
		// Parse workout date
		_, err := time.Parse("2006-01-02", workout.WorkoutDate)
		if err != nil {
			result.InvalidWorkouts++
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid date format: %s", workout.WorkoutDate))
			continue
		}

		// Check for duplicate (same user, same date)
		// TODO: Add duplicate detection using userWorkoutRepo.ListByUserAndDateRange
		// For now, mark as valid if date is parseable

		// Validate movements exist
		for _, movement := range workout.Movements {
			movementName, ok := movement["movement_name"].(string)
			if !ok || movementName == "" {
				result.InvalidWorkouts++
				result.Errors = append(result.Errors, fmt.Sprintf("Movement missing name in workout on %s", workout.WorkoutDate))
				continue
			}

			// Check if movement exists
			existingMovement, err := s.movementRepo.GetByName(movementName)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking movement %s: %v", movementName, err))
				continue
			}
			if existingMovement == nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Movement not found: %s (will be created)", movementName))
			}
		}

		// Validate WODs exist
		for _, wod := range workout.WODs {
			wodName, ok := wod["wod_name"].(string)
			if !ok || wodName == "" {
				result.InvalidWorkouts++
				result.Errors = append(result.Errors, fmt.Sprintf("WOD missing name in workout on %s", workout.WorkoutDate))
				continue
			}

			// Check if WOD exists
			existingWOD, err := s.wodRepo.GetByName(wodName)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking WOD %s: %v", wodName, err))
				continue
			}
			if existingWOD == nil {
				result.Errors = append(result.Errors, fmt.Sprintf("WOD not found: %s (will be created)", wodName))
			}
		}

		result.ValidWorkouts++
	}

	return result, nil
}

// ConfirmUserWorkoutImport actually imports user workout data after preview
func (s *ImportService) ConfirmUserWorkoutImport(jsonData []byte, userID int64, skipDuplicates bool) (*UserWorkoutImportResult, error) {
	// Parse JSON into export structure
	var exportData struct {
		ExportMetadata struct {
			UserEmail  string `json:"user_email"`
			Version    string `json:"version"`
			TotalCount int    `json:"total_count"`
		} `json:"export_metadata"`
		UserWorkouts []struct {
			WorkoutDate string  `json:"workout_date"`
			WorkoutType *string `json:"workout_type,omitempty"`
			WorkoutName *string `json:"workout_name,omitempty"`
			TotalTime   *int    `json:"total_time,omitempty"`
			Notes       *string `json:"notes,omitempty"`
			Movements   []struct {
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
			} `json:"movements,omitempty"`
			WODs []struct {
				WODName     string   `json:"wod_name"`
				WODType     string   `json:"wod_type"`
				ScoreType   *string  `json:"score_type,omitempty"`
				ScoreValue  *string  `json:"score_value,omitempty"`
				TimeSeconds *int     `json:"time_seconds,omitempty"`
				Rounds      *int     `json:"rounds,omitempty"`
				Reps        *int     `json:"reps,omitempty"`
				Weight      *float64 `json:"weight,omitempty"`
				Notes       string   `json:"notes,omitempty"`
				IsPR        bool     `json:"is_pr"`
				OrderIndex  int      `json:"order_index"`
			} `json:"wods,omitempty"`
		} `json:"user_workouts"`
	}

	if err := json.Unmarshal(jsonData, &exportData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &UserWorkoutImportResult{
		TotalWorkouts: len(exportData.UserWorkouts),
		Errors:        []string{},
	}

	// Import each workout
	for _, workoutData := range exportData.UserWorkouts {
		// Parse workout date
		workoutDate, err := time.Parse("2006-01-02", workoutData.WorkoutDate)
		if err != nil {
			result.InvalidWorkouts++
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid date: %s", workoutData.WorkoutDate))
			continue
		}

		// Check for duplicate workout on same date
		// Note: This is simplified - in production you'd check against actual user_workouts table
		// For now, we'll just create the workout
		// TODO: Add duplicate detection using userWorkoutRepo.ListByUserAndDateRange

		// Create or get movements
		movementIDs := make(map[string]int64)
		for _, movement := range workoutData.Movements {
			existingMovement, err := s.movementRepo.GetByName(movement.MovementName)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking movement: %v", err))
				continue
			}

			if existingMovement == nil {
				// Create movement
				newMovement := &domain.Movement{
					Name:        movement.MovementName,
					Type:        domain.MovementType(movement.MovementType),
					Description: "",
					IsStandard:  false,
					CreatedBy:   &userID,
				}
				if err := s.movementRepo.Create(newMovement); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to create movement: %v", err))
					continue
				}
				movementIDs[movement.MovementName] = newMovement.ID
				result.MovementsCreated++
			} else {
				movementIDs[movement.MovementName] = existingMovement.ID
			}
		}

		// Create or get WODs
		wodIDs := make(map[string]int64)
		for _, wod := range workoutData.WODs {
			existingWOD, err := s.wodRepo.GetByName(wod.WODName)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking WOD: %v", err))
				continue
			}

			if existingWOD == nil {
				// Create WOD with minimal info
				newWOD := &domain.WOD{
					Name:        wod.WODName,
					Type:        wod.WODType,
					Source:      "Self-recorded",
					Regime:      "AMRAP",
					ScoreType:   "Rounds+Reps",
					Description: fmt.Sprintf("Imported from backup on %s", time.Now().Format("2006-01-02")),
					IsStandard:  false,
					CreatedBy:   &userID,
				}
				if wod.ScoreType != nil {
					newWOD.ScoreType = *wod.ScoreType
				}
				if err := s.wodRepo.Create(newWOD); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to create WOD: %v", err))
					continue
				}
				wodIDs[wod.WODName] = newWOD.ID
				result.WODsCreated++
			} else {
				wodIDs[wod.WODName] = existingWOD.ID
			}
		}

		// Create UserWorkout record
		userWorkout := &domain.UserWorkout{
			UserID:      userID,
			WorkoutDate: workoutDate,
			WorkoutName: workoutData.WorkoutName,
			WorkoutType: workoutData.WorkoutType,
			Notes:       workoutData.Notes,
			TotalTime:   workoutData.TotalTime,
		}

		if err := s.userWorkoutRepo.Create(userWorkout); err != nil {
			result.InvalidWorkouts++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create workout: %v", err))
			continue
		}

		// Create UserWorkoutMovement records
		for _, movement := range workoutData.Movements {
			movementID, exists := movementIDs[movement.MovementName]
			if !exists {
				continue // Skip if movement wasn't found/created
			}

			userWorkoutMovement := &domain.UserWorkoutMovement{
				UserWorkoutID: userWorkout.ID,
				MovementID:    movementID,
				Sets:          movement.Sets,
				Reps:          movement.Reps,
				Weight:        movement.Weight,
				Time:          movement.Time,
				Distance:      movement.Distance,
				Notes:         movement.Notes,
				IsPR:          movement.IsPR,
				OrderIndex:    movement.OrderIndex,
			}

			if err := s.userWorkoutMovementRepo.Create(userWorkoutMovement); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create workout movement: %v", err))
			}
		}

		// Create UserWorkoutWOD records
		for _, wod := range workoutData.WODs {
			wodID, exists := wodIDs[wod.WODName]
			if !exists {
				continue // Skip if WOD wasn't found/created
			}

			userWorkoutWOD := &domain.UserWorkoutWOD{
				UserWorkoutID: userWorkout.ID,
				WODID:         wodID,
				TimeSeconds:   wod.TimeSeconds,
				Rounds:        wod.Rounds,
				Reps:          wod.Reps,
				Weight:        wod.Weight,
				Notes:         wod.Notes,
				IsPR:          wod.IsPR,
				OrderIndex:    wod.OrderIndex,
			}

			if err := s.userWorkoutWODRepo.Create(userWorkoutWOD); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create workout WOD: %v", err))
			}
		}

		result.CreatedCount++
		result.ValidWorkouts++
	}

	return result, nil
}
