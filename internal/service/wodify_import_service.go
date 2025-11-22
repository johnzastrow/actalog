package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WodifyImportService handles Wodify performance data imports
type WodifyImportService struct {
	userRepo                domain.UserRepository
	movementRepo            domain.MovementRepository
	wodRepo                 domain.WODRepository
	userWorkoutRepo         domain.UserWorkoutRepository
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository
	userWorkoutWODRepo      domain.UserWorkoutWODRepository
	parser                  *WodifyResultParser
}

// NewWodifyImportService creates a new Wodify import service
func NewWodifyImportService(
	userRepo domain.UserRepository,
	movementRepo domain.MovementRepository,
	wodRepo domain.WODRepository,
	userWorkoutRepo domain.UserWorkoutRepository,
	userWorkoutMovementRepo domain.UserWorkoutMovementRepository,
	userWorkoutWODRepo domain.UserWorkoutWODRepository,
) *WodifyImportService {
	return &WodifyImportService{
		userRepo:                userRepo,
		movementRepo:            movementRepo,
		wodRepo:                 wodRepo,
		userWorkoutRepo:         userWorkoutRepo,
		userWorkoutMovementRepo: userWorkoutMovementRepo,
		userWorkoutWODRepo:      userWorkoutWODRepo,
		parser:                  NewWodifyResultParser(),
	}
}

// PreviewImport parses the CSV and returns a preview of what will be imported
func (s *WodifyImportService) PreviewImport(csvData io.Reader, userID int64) (*domain.WodifyImportPreview, error) {
	// Parse CSV
	rows, errors := s.parseCSV(csvData)

	// Group by date
	grouped := s.groupByDate(rows)

	// Analyze what needs to be created
	newMovements, newWODs := s.analyzeNewEntities(rows, userID)

	// Create preview
	preview := &domain.WodifyImportPreview{
		TotalRows:            len(rows) + len(errors),
		ValidRows:            len(rows),
		InvalidRows:          len(errors),
		UniqueWorkoutDates:   len(grouped),
		MovementsToCreate:    len(newMovements),
		WODsToCreate:         len(newWODs),
		UserWorkoutsToCreate: len(grouped),
		PerformancesToCreate: len(rows),
		Errors:               errors,
		WorkoutSummary:       s.createWorkoutSummary(grouped),
		NewMovements:         newMovements,
		NewWODs:              newWODs,
	}

	return preview, nil
}

// ConfirmImport executes the import after preview
func (s *WodifyImportService) ConfirmImport(csvData io.Reader, userID int64) (*domain.WodifyImportResult, error) {
	// Parse CSV
	rows, _ := s.parseCSV(csvData)

	// Group by date
	grouped := s.groupByDate(rows)

	result := &domain.WodifyImportResult{}

	// Process each workout date
	for _, workout := range grouped {
		if err := s.importWorkout(workout, userID, result); err != nil {
			return nil, fmt.Errorf("failed to import workout for %s: %w", workout.Date.Format("2006-01-02"), err)
		}
	}

	return result, nil
}

// parseCSV reads and parses the Wodify CSV file
func (s *WodifyImportService) parseCSV(csvData io.Reader) ([]domain.WodifyPerformanceRow, []domain.WodifyImportError) {
	reader := csv.NewReader(csvData)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var rows []domain.WodifyPerformanceRow
	var errors []domain.WodifyImportError

	// Read header
	header, err := reader.Read()
	if err != nil {
		errors = append(errors, domain.WodifyImportError{
			Row:     0,
			Message: fmt.Sprintf("Failed to read header: %v", err),
		})
		return rows, errors
	}

	// Validate header
	expectedCols := 19
	if len(header) < expectedCols {
		errors = append(errors, domain.WodifyImportError{
			Row:     0,
			Message: fmt.Sprintf("Invalid header: expected %d columns, got %d", expectedCols, len(header)),
		})
		return rows, errors
	}

	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, domain.WodifyImportError{
				Row:     rowNum,
				Message: fmt.Sprintf("Failed to parse row: %v", err),
			})
			rowNum++
			continue
		}

		// Skip rows with insufficient columns
		if len(record) < expectedCols {
			rowNum++
			continue
		}

		// Skip empty rows (no date)
		if strings.TrimSpace(record[2]) == "" {
			rowNum++
			continue
		}

		// Parse row
		row := domain.WodifyPerformanceRow{
			CustomerName:             record[0],
			LocationName:             record[1],
			Date:                     record[2],
			ProgramName:              record[3],
			ClassName:                record[4],
			ComponentType:            record[5],
			ComponentID:              record[6],
			ComponentName:            record[7],
			ComponentDescription:     record[8],
			PerformanceResultType:    record[9],
			RepScheme:                record[10],
			FullyFormattedResult:     record[11],
			FromWeightliftingTotal:   parseBool(record[12]),
			FromVariableSet:          parseBool(record[13]),
			IsRx:                     parseBool(record[14]),
			IsRxPlus:                 parseBool(record[15]),
			IsPersonalRecord:         parseBool(record[16]),
			PersonalRecordDescription: record[17],
			Comment:                  record[18],
		}

		// Validate required fields
		if row.ComponentType == "" || row.ComponentName == "" {
			errors = append(errors, domain.WodifyImportError{
				Row:     rowNum,
				Message: "Missing component type or name",
			})
			rowNum++
			continue
		}

		rows = append(rows, row)
		rowNum++
	}

	return rows, errors
}

// groupByDate groups performance rows by workout date
func (s *WodifyImportService) groupByDate(rows []domain.WodifyPerformanceRow) []domain.WodifyGroupedWorkout {
	groupMap := make(map[string][]domain.WodifyPerformanceRow)

	for _, row := range rows {
		groupMap[row.Date] = append(groupMap[row.Date], row)
	}

	var grouped []domain.WodifyGroupedWorkout
	for dateStr, performances := range groupMap {
		date, err := s.parser.ParseDate(dateStr)
		if err != nil {
			continue // Skip invalid dates
		}

		grouped = append(grouped, domain.WodifyGroupedWorkout{
			Date:         date,
			Performances: performances,
		})
	}

	// Sort by date
	sort.Slice(grouped, func(i, j int) bool {
		return grouped[i].Date.Before(grouped[j].Date)
	})

	return grouped
}

// analyzeNewEntities determines which movements and WODs need to be created
func (s *WodifyImportService) analyzeNewEntities(rows []domain.WodifyPerformanceRow, userID int64) ([]string, []string) {
	movementNames := make(map[string]bool)
	wodNames := make(map[string]bool)

	for _, row := range rows {
		if row.ComponentType == "Metcon" {
			wodNames[row.ComponentName] = true
		} else {
			movementNames[row.ComponentName] = true
		}
	}

	// Check which movements already exist
	var newMovements []string
	for name := range movementNames {
		movements, _ := s.movementRepo.Search(name, 10)
		exists := false
		for _, m := range movements {
			if strings.EqualFold(m.Name, name) {
				exists = true
				break
			}
		}
		if !exists {
			newMovements = append(newMovements, name)
		}
	}

	// Check which WODs already exist
	var newWODs []string
	for name := range wodNames {
		wods, _ := s.wodRepo.Search(name, 10)
		exists := false
		for _, w := range wods {
			if strings.EqualFold(w.Name, name) {
				exists = true
				break
			}
		}
		if !exists {
			newWODs = append(newWODs, name)
		}
	}

	sort.Strings(newMovements)
	sort.Strings(newWODs)

	return newMovements, newWODs
}

// createWorkoutSummary creates a summary of workouts to be imported
func (s *WodifyImportService) createWorkoutSummary(grouped []domain.WodifyGroupedWorkout) []domain.WodifyWorkoutSummary {
	var summary []domain.WodifyWorkoutSummary

	for _, workout := range grouped {
		movementCount := 0
		wodCount := 0
		hasPRs := false
		types := make(map[string]bool)

		for _, perf := range workout.Performances {
			if perf.ComponentType == "Metcon" {
				wodCount++
			} else {
				movementCount++
			}
			if perf.IsPersonalRecord {
				hasPRs = true
			}
			types[perf.ComponentType] = true
		}

		typesList := []string{}
		for t := range types {
			typesList = append(typesList, t)
		}

		summary = append(summary, domain.WodifyWorkoutSummary{
			Date:           workout.Date.Format("2006-01-02"),
			MovementCount:  movementCount,
			WODCount:       wodCount,
			HasPRs:         hasPRs,
			ComponentTypes: strings.Join(typesList, ", "),
		})
	}

	return summary
}

// importWorkout imports a single grouped workout
func (s *WodifyImportService) importWorkout(workout domain.WodifyGroupedWorkout, userID int64, result *domain.WodifyImportResult) error {
	// Determine workout type based on predominant component type
	workoutType := s.determineWorkoutType(workout.Performances)

	// Create UserWorkout
	workoutName := fmt.Sprintf("Workout %s", workout.Date.Format("2006-01-02"))
	userWorkout := &domain.UserWorkout{
		UserID:      userID,
		WorkoutDate: workout.Date,
		WorkoutName: &workoutName,
		WorkoutType: &workoutType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userWorkoutRepo.Create(userWorkout); err != nil {
		return fmt.Errorf("failed to create user workout: %w", err)
	}

	result.WorkoutsCreated++

	// Process each performance
	for orderIndex, perf := range workout.Performances {
		if perf.ComponentType == "Metcon" {
			if err := s.importWODPerformance(userWorkout.ID, userID, perf, orderIndex, result); err != nil {
				return fmt.Errorf("failed to import WOD performance: %w", err)
			}
		} else {
			if err := s.importMovementPerformance(userWorkout.ID, userID, perf, orderIndex, result); err != nil {
				return fmt.Errorf("failed to import movement performance: %w", err)
			}
		}

		if perf.IsPersonalRecord {
			result.PRsFlagged++
		}
	}

	return nil
}

// importMovementPerformance imports a movement performance
func (s *WodifyImportService) importMovementPerformance(userWorkoutID, userID int64, perf domain.WodifyPerformanceRow, orderIndex int, result *domain.WodifyImportResult) error {
	// Get or create movement
	movement, created, err := s.getOrCreateMovement(perf, userID)
	if err != nil {
		return err
	}
	if created {
		result.MovementsCreated++
	}

	// Parse result
	parsed, err := s.parser.ParseResult(perf.PerformanceResultType, perf.FullyFormattedResult, perf.Comment)
	if err != nil {
		parsed = &domain.ParsedPerformanceResult{
			Notes: fmt.Sprintf("%s: %s. %s", perf.PerformanceResultType, perf.FullyFormattedResult, perf.Comment),
		}
	}
	parsed.IsPR = perf.IsPersonalRecord

	// Create UserWorkoutMovement
	uwm := &domain.UserWorkoutMovement{
		UserWorkoutID: userWorkoutID,
		MovementID:    movement.ID,
		Sets:          parsed.Sets,
		Reps:          parsed.Reps,
		Weight:        parsed.Weight,
		Time:          parsed.TimeSeconds,
		Distance:      parsed.Distance,
		Notes:         parsed.Notes,
		IsPR:          parsed.IsPR,
		OrderIndex:    orderIndex,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userWorkoutMovementRepo.Create(uwm); err != nil {
		return fmt.Errorf("failed to create user workout movement: %w", err)
	}

	result.PerformancesCreated++
	return nil
}

// importWODPerformance imports a WOD performance
func (s *WodifyImportService) importWODPerformance(userWorkoutID, userID int64, perf domain.WodifyPerformanceRow, orderIndex int, result *domain.WodifyImportResult) error {
	// Get or create WOD
	wod, created, err := s.getOrCreateWOD(perf, userID)
	if err != nil {
		return err
	}
	if created {
		result.WODsCreated++
	}

	// Parse result
	parsed, err := s.parser.ParseResult(perf.PerformanceResultType, perf.FullyFormattedResult, perf.Comment)
	if err != nil {
		parsed = &domain.ParsedPerformanceResult{
			Notes: fmt.Sprintf("%s: %s. %s", perf.PerformanceResultType, perf.FullyFormattedResult, perf.Comment),
		}
	}
	parsed.IsPR = perf.IsPersonalRecord

	// Determine score type and value
	scoreType := s.parser.DetermineWODScoreType(perf.PerformanceResultType)
	scoreValue := s.formatScoreValue(parsed, scoreType)

	// Create UserWorkoutWOD
	uww := &domain.UserWorkoutWOD{
		UserWorkoutID: userWorkoutID,
		WODID:         wod.ID,
		ScoreType:     &scoreType,
		ScoreValue:    scoreValue,
		TimeSeconds:   parsed.TimeSeconds,
		Rounds:        parsed.Rounds,
		Reps:          parsed.Reps,
		Weight:        parsed.Weight,
		Notes:         parsed.Notes,
		IsPR:          parsed.IsPR,
		OrderIndex:    orderIndex,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userWorkoutWODRepo.Create(uww); err != nil {
		return fmt.Errorf("failed to create user workout WOD: %w", err)
	}

	result.PerformancesCreated++
	return nil
}

// getOrCreateMovement gets an existing movement or creates a new one
func (s *WodifyImportService) getOrCreateMovement(perf domain.WodifyPerformanceRow, userID int64) (*domain.Movement, bool, error) {
	// Search for existing movement
	movements, err := s.movementRepo.Search(perf.ComponentName, 10)
	if err == nil {
		for _, m := range movements {
			if strings.EqualFold(m.Name, perf.ComponentName) {
				return m, false, nil
			}
		}
	}

	// Create new movement
	movementType := s.parser.DetermineMovementType(perf.ComponentType)
	isStandard := false // User-created from import

	movement := &domain.Movement{
		Name:        perf.ComponentName,
		Type:        domain.MovementType(movementType),
		Description: perf.ComponentDescription,
		IsStandard:  isStandard,
		CreatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.movementRepo.Create(movement); err != nil {
		return nil, false, fmt.Errorf("failed to create movement: %w", err)
	}

	return movement, true, nil
}

// getOrCreateWOD gets an existing WOD or creates a new one
func (s *WodifyImportService) getOrCreateWOD(perf domain.WodifyPerformanceRow, userID int64) (*domain.WOD, bool, error) {
	// Search for existing WOD
	wods, err := s.wodRepo.Search(perf.ComponentName, 10)
	if err == nil {
		for _, w := range wods {
			if strings.EqualFold(w.Name, perf.ComponentName) {
				return w, false, nil
			}
		}
	}

	// Create new WOD
	scoreType := s.parser.DetermineWODScoreType(perf.PerformanceResultType)
	isStandard := false // User-created from import

	wod := &domain.WOD{
		Name:        perf.ComponentName,
		Source:      "Wodify Import",
		Type:        "Self-created",
		Regime:      "AMRAP",
		ScoreType:   scoreType,
		Description: perf.ComponentDescription,
		IsStandard:  isStandard,
		CreatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.wodRepo.Create(wod); err != nil {
		return nil, false, fmt.Errorf("failed to create WOD: %w", err)
	}

	return wod, true, nil
}

// determineWorkoutType determines the workout type based on predominant component type
func (s *WodifyImportService) determineWorkoutType(performances []domain.WodifyPerformanceRow) string {
	counts := make(map[string]int)

	for _, perf := range performances {
		if perf.ComponentType == "Metcon" {
			counts["metcon"]++
		} else if perf.ComponentType == "Weightlifting" {
			counts["strength"]++
		} else {
			counts["gymnastics"]++
		}
	}

	// Return the predominant type
	maxType := "metcon"
	maxCount := 0
	for typ, count := range counts {
		if count > maxCount {
			maxCount = count
			maxType = typ
		}
	}

	return maxType
}

// formatScoreValue formats the score value string based on parsed result
func (s *WodifyImportService) formatScoreValue(parsed *domain.ParsedPerformanceResult, scoreType string) *string {
	var value string

	switch scoreType {
	case "Time (HH:MM:SS)":
		if parsed.TimeSeconds != nil {
			seconds := *parsed.TimeSeconds
			hours := seconds / 3600
			minutes := (seconds % 3600) / 60
			secs := seconds % 60
			if hours > 0 {
				value = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
			} else {
				value = fmt.Sprintf("%02d:%02d", minutes, secs)
			}
		}
	case "Rounds+Reps":
		if parsed.Rounds != nil && parsed.Reps != nil {
			value = fmt.Sprintf("%d rounds + %d reps", *parsed.Rounds, *parsed.Reps)
		} else if parsed.Rounds != nil {
			value = fmt.Sprintf("%d rounds", *parsed.Rounds)
		} else if parsed.Reps != nil {
			value = fmt.Sprintf("%d reps", *parsed.Reps)
		}
	case "Max Weight":
		if parsed.Weight != nil {
			value = fmt.Sprintf("%.0f lbs", *parsed.Weight)
		}
	}

	if value == "" {
		return nil
	}
	return &value
}

// Helper function to parse boolean strings
func parseBool(s string) bool {
	return strings.ToUpper(strings.TrimSpace(s)) == "TRUE"
}
