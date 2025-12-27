package service

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewExportService(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	if svc == nil {
		t.Fatal("NewExportService returned nil")
	}
}

func TestExportService_ExportWODsToCSV_StandardOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create standard WODs
	wod1 := &domain.WOD{Name: "Fran", Source: "CrossFit", Type: "For Time", Regime: "21-15-9", ScoreType: "time", Description: "Thrusters and Pull-ups", IsStandard: true}
	wod2 := &domain.WOD{Name: "Grace", Source: "CrossFit", Type: "For Time", Regime: "30 reps", ScoreType: "time", Description: "Clean and Jerks", IsStandard: true}
	_ = wodRepo.Create(wod1)
	_ = wodRepo.Create(wod2)

	// Export standard WODs only
	data, err := svc.ExportWODsToCSV(1, false, true, false)
	if err != nil {
		t.Errorf("ExportWODsToCSV() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportWODsToCSV() returned empty data")
	}

	// Parse CSV and verify
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Errorf("Failed to parse CSV: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV has %d rows, want 3 (header + 2 WODs)", len(records))
	}

	// Verify header
	expectedHeader := []string{"name", "source", "type", "regime", "score_type", "description", "url", "notes", "is_standard", "created_by_email"}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("Header[%d] = %s, want %s", i, records[0][i], h)
		}
	}
}

func TestExportService_ExportWODsToCSV_CustomOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create custom WOD owned by user
	customWOD := &domain.WOD{Name: "My Custom WOD", Source: "Custom", Type: "AMRAP", IsStandard: false, CreatedBy: &user.ID}
	_ = wodRepo.Create(customWOD)

	// Export custom WODs only
	data, err := svc.ExportWODsToCSV(user.ID, false, false, true)
	if err != nil {
		t.Errorf("ExportWODsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("CSV has %d rows, want 2 (header + 1 custom WOD)", len(records))
	}

	// Verify the WOD name and creator email
	if len(records) > 1 {
		if records[1][0] != "My Custom WOD" {
			t.Errorf("WOD name = %s, want My Custom WOD", records[1][0])
		}
		if records[1][9] != "test@example.com" {
			t.Errorf("Created by email = %s, want test@example.com", records[1][9])
		}
	}
}

func TestExportService_ExportWODsToCSV_NothingSelected(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export with nothing selected
	_, err := svc.ExportWODsToCSV(1, false, false, false)
	if err == nil {
		t.Error("ExportWODsToCSV() should return error when nothing selected")
	}
}

func TestExportService_ExportWODsToCSV_AdminAll(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create standard WOD
	standardWOD := &domain.WOD{Name: "Fran", IsStandard: true}
	_ = wodRepo.Create(standardWOD)

	// Create custom WOD
	userID := int64(1)
	customWOD := &domain.WOD{Name: "Custom", IsStandard: false, CreatedBy: &userID}
	_ = wodRepo.Create(customWOD)

	// Admin exports all
	data, err := svc.ExportWODsToCSV(1, true, true, true)
	if err != nil {
		t.Errorf("ExportWODsToCSV() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 2 data rows (standard + custom)
	if len(records) != 3 {
		t.Errorf("CSV has %d rows, want 3", len(records))
	}
}

func TestExportService_ExportWODsToJSON_StandardOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create standard WODs
	wod1 := &domain.WOD{Name: "Fran", IsStandard: true}
	wod2 := &domain.WOD{Name: "Grace", IsStandard: true}
	_ = wodRepo.Create(wod1)
	_ = wodRepo.Create(wod2)

	// Export to JSON
	data, err := svc.ExportWODsToJSON(1, false, true, false)
	if err != nil {
		t.Errorf("ExportWODsToJSON() error = %v", err)
	}

	// Verify valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("JSON has %d items, want 2", len(result))
	}
}

func TestExportService_ExportWODsToJSON_NothingSelected(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export with nothing selected
	_, err := svc.ExportWODsToJSON(1, false, false, false)
	if err == nil {
		t.Error("ExportWODsToJSON() should return error when nothing selected")
	}
}

func TestExportService_ExportMovementsToCSV_StandardOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create standard movements
	movement1 := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, Description: "Barbell back squat", IsStandard: true}
	movement2 := &domain.Movement{Name: "Deadlift", Type: domain.MovementTypeWeightlifting, Description: "Conventional deadlift", IsStandard: true}
	_ = movementRepo.Create(movement1)
	_ = movementRepo.Create(movement2)

	// Export standard movements only
	data, err := svc.ExportMovementsToCSV(1, false, true, false)
	if err != nil {
		t.Errorf("ExportMovementsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV has %d rows, want 3 (header + 2 movements)", len(records))
	}

	// Verify header
	expectedHeader := []string{"name", "type", "description", "is_standard", "created_by_email"}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("Header[%d] = %s, want %s", i, records[0][i], h)
		}
	}
}

func TestExportService_ExportMovementsToCSV_CustomOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create custom movement
	customMovement := &domain.Movement{Name: "My Custom Movement", Type: domain.MovementTypeWeightlifting, IsStandard: false, CreatedBy: &user.ID}
	_ = movementRepo.Create(customMovement)

	// Export custom movements only
	data, err := svc.ExportMovementsToCSV(user.ID, false, false, true)
	if err != nil {
		t.Errorf("ExportMovementsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("CSV has %d rows, want 2 (header + 1 custom movement)", len(records))
	}

	// Verify the movement name and creator email
	if len(records) > 1 {
		if records[1][0] != "My Custom Movement" {
			t.Errorf("Movement name = %s, want My Custom Movement", records[1][0])
		}
		if records[1][4] != "test@example.com" {
			t.Errorf("Created by email = %s, want test@example.com", records[1][4])
		}
	}
}

func TestExportService_ExportMovementsToCSV_NothingSelected(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export with nothing selected
	_, err := svc.ExportMovementsToCSV(1, false, false, false)
	if err == nil {
		t.Error("ExportMovementsToCSV() should return error when nothing selected")
	}
}

func TestExportService_ExportMovementsToJSON_StandardOnly(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create standard movements
	movement1 := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	movement2 := &domain.Movement{Name: "Deadlift", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	_ = movementRepo.Create(movement1)
	_ = movementRepo.Create(movement2)

	// Export to JSON
	data, err := svc.ExportMovementsToJSON(1, false, true, false)
	if err != nil {
		t.Errorf("ExportMovementsToJSON() error = %v", err)
	}

	// Verify valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("JSON has %d items, want 2", len(result))
	}
}

func TestExportService_ExportMovementsToJSON_NothingSelected(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export with nothing selected
	_, err := svc.ExportMovementsToJSON(1, false, false, false)
	if err == nil {
		t.Error("ExportMovementsToJSON() should return error when nothing selected")
	}
}

func TestExportService_ExportUserWorkoutsToJSON(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create user workouts
	workoutName := "Morning Workout"
	workout1 := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now().AddDate(0, 0, -1),
		WorkoutName: &workoutName,
	}
	_ = userWorkoutRepo.Create(workout1)

	// Add details
	userWorkoutRepo.userWorkoutsDetails[workout1.ID] = &domain.UserWorkoutWithDetails{
		UserWorkout: *workout1,
		WorkoutName: workoutName,
	}

	// Export to JSON
	data, err := svc.ExportUserWorkoutsToJSON(user.ID, nil, nil)
	if err != nil {
		t.Errorf("ExportUserWorkoutsToJSON() error = %v", err)
	}

	// Verify valid JSON structure
	var result UserWorkoutExport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	// Verify metadata
	if result.ExportMetadata.UserEmail != "test@example.com" {
		t.Errorf("UserEmail = %s, want test@example.com", result.ExportMetadata.UserEmail)
	}
	if result.ExportMetadata.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.ExportMetadata.TotalCount)
	}
	if result.ExportMetadata.DateRange != nil {
		t.Error("DateRange should be nil when no date filter")
	}
}

func TestExportService_ExportUserWorkoutsToJSON_WithDateRange(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create user workouts
	workoutName := "Workout 1"
	workout1 := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now().AddDate(0, 0, -5),
		WorkoutName: &workoutName,
	}
	_ = userWorkoutRepo.Create(workout1)

	userWorkoutRepo.userWorkoutsDetails[workout1.ID] = &domain.UserWorkoutWithDetails{
		UserWorkout: *workout1,
		WorkoutName: workoutName,
	}

	// Export with date range
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()
	data, err := svc.ExportUserWorkoutsToJSON(user.ID, &startDate, &endDate)
	if err != nil {
		t.Errorf("ExportUserWorkoutsToJSON() error = %v", err)
	}

	// Verify JSON structure
	var result UserWorkoutExport
	_ = json.Unmarshal(data, &result)

	// Verify date range is included
	if result.ExportMetadata.DateRange == nil {
		t.Error("DateRange should not be nil when date filter is applied")
	}
}

func TestExportService_ExportUserWorkoutsToJSON_UserNotFound(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export for non-existent user
	_, err := svc.ExportUserWorkoutsToJSON(999, nil, nil)
	if err == nil {
		t.Error("ExportUserWorkoutsToJSON() should return error when user not found")
	}
}

func TestExportService_ExportUserWorkoutsToCSV(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create user workouts
	workoutName := "Morning Workout"
	workout1 := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now(),
		WorkoutName: &workoutName,
	}
	_ = userWorkoutRepo.Create(workout1)

	// Add details with movements
	userWorkoutRepo.userWorkoutsDetails[workout1.ID] = &domain.UserWorkoutWithDetails{
		UserWorkout: *workout1,
		WorkoutName: workoutName,
		PerformanceMovements: []*domain.UserWorkoutMovement{
			{
				MovementName: "Back Squat",
				MovementType: "strength",
				Sets:         intPtr(3),
				Reps:         intPtr(5),
				Weight:       floatPtr(225.0),
				IsPR:         true,
				OrderIndex:   0,
			},
		},
	}

	// Export to CSV
	data, err := svc.ExportUserWorkoutsToCSV(user.ID, nil, nil)
	if err != nil {
		t.Errorf("ExportUserWorkoutsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 1 data row (for the movement)
	if len(records) < 2 {
		t.Errorf("CSV has %d rows, want at least 2", len(records))
	}

	// Verify header has expected columns
	if len(records[0]) != 19 {
		t.Errorf("Header has %d columns, want 19", len(records[0]))
	}
}

func TestExportService_ExportUserWorkoutsToCSV_EmptyWorkout(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create user workout with no movements or WODs
	workoutName := "Empty Workout"
	workout1 := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now(),
		WorkoutName: &workoutName,
	}
	_ = userWorkoutRepo.Create(workout1)

	// Add details with no performances
	userWorkoutRepo.userWorkoutsDetails[workout1.ID] = &domain.UserWorkoutWithDetails{
		UserWorkout:          *workout1,
		WorkoutName:          workoutName,
		PerformanceMovements: []*domain.UserWorkoutMovement{},
		PerformanceWODs:      []*domain.UserWorkoutWOD{},
	}

	// Export to CSV
	data, err := svc.ExportUserWorkoutsToCSV(user.ID, nil, nil)
	if err != nil {
		t.Errorf("ExportUserWorkoutsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 1 empty row (for workout with no performances)
	if len(records) != 2 {
		t.Errorf("CSV has %d rows, want 2 (header + 1 empty workout row)", len(records))
	}
}

func TestExportService_ExportUserWorkoutsToCSV_UserNotFound(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Try to export for non-existent user
	_, err := svc.ExportUserWorkoutsToCSV(999, nil, nil)
	if err == nil {
		t.Error("ExportUserWorkoutsToCSV() should return error when user not found")
	}
}

func TestExportService_ExportWODsToCSV_BothStandardAndCustom(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create standard WOD
	standardWOD := &domain.WOD{Name: "Fran", IsStandard: true}
	_ = wodRepo.Create(standardWOD)

	// Create custom WOD
	customWOD := &domain.WOD{Name: "My WOD", IsStandard: false, CreatedBy: &user.ID}
	_ = wodRepo.Create(customWOD)

	// Export both standard and custom (as regular user)
	data, err := svc.ExportWODsToCSV(user.ID, false, true, true)
	if err != nil {
		t.Errorf("ExportWODsToCSV() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV has %d rows, want 3 (header + 2 WODs)", len(records))
	}
}

func TestExportService_ExportMovementsToCSV_BothStandardAndCustom(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create standard movement
	standardMovement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	_ = movementRepo.Create(standardMovement)

	// Create custom movement
	customMovement := &domain.Movement{Name: "My Movement", Type: domain.MovementTypeWeightlifting, IsStandard: false, CreatedBy: &user.ID}
	_ = movementRepo.Create(customMovement)

	// Export both standard and custom
	data, err := svc.ExportMovementsToCSV(user.ID, false, true, true)
	if err != nil {
		t.Errorf("ExportMovementsToCSV() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV has %d rows, want 3 (header + 2 movements)", len(records))
	}
}

func TestExportService_ExportUserWorkoutsToCSV_WithWODs(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()

	svc := NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	// Create user workout
	workoutName := "WOD Day"
	workout := &domain.UserWorkout{
		UserID:      user.ID,
		WorkoutDate: time.Now(),
		WorkoutName: &workoutName,
	}
	_ = userWorkoutRepo.Create(workout)

	// Add details with WOD
	scoreType := "time"
	scoreValue := "5:30"
	userWorkoutRepo.userWorkoutsDetails[workout.ID] = &domain.UserWorkoutWithDetails{
		UserWorkout: *workout,
		WorkoutName: workoutName,
		PerformanceWODs: []*domain.UserWorkoutWOD{
			{
				WODName:    "Fran",
				WODType:    "For Time",
				ScoreType:  &scoreType,
				ScoreValue: &scoreValue,
				IsPR:       true,
				OrderIndex: 0,
			},
		},
	}

	// Export to CSV
	data, err := svc.ExportUserWorkoutsToCSV(user.ID, nil, nil)
	if err != nil {
		t.Errorf("ExportUserWorkoutsToCSV() error = %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, _ := reader.ReadAll()

	// Should have header + 1 data row (for the WOD)
	if len(records) < 2 {
		t.Errorf("CSV has %d rows, want at least 2", len(records))
	}

	// Verify the WOD row has performance_type = "wod"
	if len(records) > 1 && records[1][5] != "wod" {
		t.Errorf("performance_type = %s, want wod", records[1][5])
	}
}
