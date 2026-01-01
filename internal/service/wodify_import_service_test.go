package service

import (
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewWodifyImportService(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	if svc == nil {
		t.Fatal("NewWodifyImportService returned nil")
	}
	if svc.parser == nil {
		t.Error("parser is nil")
	}
}

// createValidWodifyCSV creates a valid Wodify CSV for testing
func createValidWodifyCSV() string {
	header := "Customer Name,Location Name,Date,Program Name,Class Name,Component Type,Component ID,Component Name,Component Description,Performance Result Type,Rep Scheme,Fully Formatted Result,From Weightlifting Total,From Variable Set,Is Rx,Is Rx Plus,Is Personal Record,Personal Record Description,Comment"
	row1 := "John Doe,Main Gym,01/15/2024,CrossFit,Morning Class,Weightlifting,1,Back Squat,5x5 Back Squat,Weight,5-5-5-5-5,3 x 5 @ 225 lbs,FALSE,FALSE,TRUE,FALSE,FALSE,,Heavy day"
	row2 := "John Doe,Main Gym,01/15/2024,CrossFit,Morning Class,Metcon,2,Fran,21-15-9 Thrusters and Pull-ups,Time,,5:30,FALSE,FALSE,TRUE,FALSE,TRUE,New PR!,Felt great"
	return header + "\n" + row1 + "\n" + row2 + "\n"
}

func TestWodifyImportService_PreviewImport_ValidCSV(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	if preview.TotalRows != 2 {
		t.Errorf("TotalRows = %d, want 2", preview.TotalRows)
	}
	if preview.ValidRows != 2 {
		t.Errorf("ValidRows = %d, want 2", preview.ValidRows)
	}
	if preview.InvalidRows != 0 {
		t.Errorf("InvalidRows = %d, want 0", preview.InvalidRows)
	}
	if preview.UniqueWorkoutDates != 1 {
		t.Errorf("UniqueWorkoutDates = %d, want 1", preview.UniqueWorkoutDates)
	}
	if preview.MovementsToCreate != 1 {
		t.Errorf("MovementsToCreate = %d, want 1 (Back Squat)", preview.MovementsToCreate)
	}
	if preview.WODsToCreate != 1 {
		t.Errorf("WODsToCreate = %d, want 1 (Fran)", preview.WODsToCreate)
	}
	if len(preview.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", preview.Errors)
	}
}

func TestWodifyImportService_PreviewImport_InvalidHeader(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := "Col1,Col2,Col3\nval1,val2,val3\n"
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Should have an error about insufficient columns
	if len(preview.Errors) == 0 {
		t.Error("Expected errors for invalid header")
	}
}

func TestWodifyImportService_PreviewImport_EmptyDate(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	header := "Customer Name,Location Name,Date,Program Name,Class Name,Component Type,Component ID,Component Name,Component Description,Performance Result Type,Rep Scheme,Fully Formatted Result,From Weightlifting Total,From Variable Set,Is Rx,Is Rx Plus,Is Personal Record,Personal Record Description,Comment"
	row := "John Doe,Main Gym,,CrossFit,Morning Class,Weightlifting,1,Back Squat,5x5,Weight,5-5-5-5-5,225 lbs,FALSE,FALSE,TRUE,FALSE,FALSE,,"
	csv := header + "\n" + row + "\n"
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Row with empty date should be skipped
	if preview.ValidRows != 0 {
		t.Errorf("ValidRows = %d, want 0 (row with empty date should be skipped)", preview.ValidRows)
	}
}

func TestWodifyImportService_PreviewImport_MissingComponentFields(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	header := "Customer Name,Location Name,Date,Program Name,Class Name,Component Type,Component ID,Component Name,Component Description,Performance Result Type,Rep Scheme,Fully Formatted Result,From Weightlifting Total,From Variable Set,Is Rx,Is Rx Plus,Is Personal Record,Personal Record Description,Comment"
	// Missing Component Type and Name
	row := "John Doe,Main Gym,01/15/2024,CrossFit,Morning Class,,,,,Weight,5-5-5-5-5,225 lbs,FALSE,FALSE,TRUE,FALSE,FALSE,,"
	csv := header + "\n" + row + "\n"
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Should have an error for missing component fields
	if len(preview.Errors) == 0 {
		t.Error("Expected error for missing component type or name")
	}
}

func TestWodifyImportService_PreviewImport_ExistingMovement(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	// Add existing movement
	movementRepo.movements[1] = &domain.Movement{
		ID:         1,
		Name:       "Back Squat",
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Back Squat already exists, should not need to create
	if preview.MovementsToCreate != 0 {
		t.Errorf("MovementsToCreate = %d, want 0 (Back Squat already exists)", preview.MovementsToCreate)
	}
}

func TestWodifyImportService_PreviewImport_ExistingWorkout(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	// Add existing workout for the date
	existingWorkout := &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	userWorkoutRepo.userWorkouts[1] = existingWorkout

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Should detect as an update
	if preview.UserWorkoutsToUpdate != 1 {
		t.Errorf("UserWorkoutsToUpdate = %d, want 1", preview.UserWorkoutsToUpdate)
	}
	if preview.UserWorkoutsToCreate != 0 {
		t.Errorf("UserWorkoutsToCreate = %d, want 0", preview.UserWorkoutsToCreate)
	}
}

func TestWodifyImportService_ConfirmImport_CreateNew(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	result, err := svc.ConfirmImport(reader, 1, false, false)
	if err != nil {
		t.Fatalf("ConfirmImport returned error: %v", err)
	}

	if result.WorkoutsCreated != 1 {
		t.Errorf("WorkoutsCreated = %d, want 1", result.WorkoutsCreated)
	}
	if result.MovementsCreated != 1 {
		t.Errorf("MovementsCreated = %d, want 1", result.MovementsCreated)
	}
	if result.WODsCreated != 1 {
		t.Errorf("WODsCreated = %d, want 1", result.WODsCreated)
	}
	if result.PerformancesCreated != 2 {
		t.Errorf("PerformancesCreated = %d, want 2", result.PerformancesCreated)
	}
	if result.PRsFlagged != 1 {
		t.Errorf("PRsFlagged = %d, want 1", result.PRsFlagged)
	}
}

func TestWodifyImportService_ConfirmImport_SkipDuplicates(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	// Add existing workout
	existingWorkout := &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	userWorkoutRepo.userWorkouts[1] = existingWorkout

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	result, err := svc.ConfirmImport(reader, 1, true, false) // skipDuplicates=true
	if err != nil {
		t.Fatalf("ConfirmImport returned error: %v", err)
	}

	if result.WorkoutsSkipped != 1 {
		t.Errorf("WorkoutsSkipped = %d, want 1", result.WorkoutsSkipped)
	}
	if result.WorkoutsCreated != 0 {
		t.Errorf("WorkoutsCreated = %d, want 0", result.WorkoutsCreated)
	}
}

func TestWodifyImportService_ConfirmImport_UpdateDuplicates(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	// Add existing workout
	existingWorkout := &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	userWorkoutRepo.userWorkouts[1] = existingWorkout

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	result, err := svc.ConfirmImport(reader, 1, false, true) // updateDuplicates=true
	if err != nil {
		t.Fatalf("ConfirmImport returned error: %v", err)
	}

	if result.WorkoutsUpdated != 1 {
		t.Errorf("WorkoutsUpdated = %d, want 1", result.WorkoutsUpdated)
	}
	if result.WorkoutsCreated != 0 {
		t.Errorf("WorkoutsCreated = %d, want 0", result.WorkoutsCreated)
	}
	if result.PerformancesUpdated != 2 {
		t.Errorf("PerformancesUpdated = %d, want 2", result.PerformancesUpdated)
	}
}

func TestWodifyImportService_ConfirmImport_MultipleDates(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	header := "Customer Name,Location Name,Date,Program Name,Class Name,Component Type,Component ID,Component Name,Component Description,Performance Result Type,Rep Scheme,Fully Formatted Result,From Weightlifting Total,From Variable Set,Is Rx,Is Rx Plus,Is Personal Record,Personal Record Description,Comment"
	row1 := "John Doe,Main Gym,01/15/2024,CrossFit,Morning,Weightlifting,1,Back Squat,5x5,Weight,5-5-5-5-5,225 lbs,FALSE,FALSE,TRUE,FALSE,FALSE,,"
	row2 := "John Doe,Main Gym,01/16/2024,CrossFit,Morning,Weightlifting,1,Deadlift,5x5,Weight,5-5-5-5-5,315 lbs,FALSE,FALSE,TRUE,FALSE,FALSE,,"
	csv := header + "\n" + row1 + "\n" + row2 + "\n"
	reader := strings.NewReader(csv)

	result, err := svc.ConfirmImport(reader, 1, false, false)
	if err != nil {
		t.Fatalf("ConfirmImport returned error: %v", err)
	}

	if result.WorkoutsCreated != 2 {
		t.Errorf("WorkoutsCreated = %d, want 2", result.WorkoutsCreated)
	}
}

func TestWodifyImportService_DetermineWorkoutType(t *testing.T) {
	svc := &WodifyImportService{}

	tests := []struct {
		name         string
		performances []domain.WodifyPerformanceRow
		want         string
	}{
		{
			name: "mostly metcon",
			performances: []domain.WodifyPerformanceRow{
				{ComponentType: "Metcon"},
				{ComponentType: "Metcon"},
				{ComponentType: "Weightlifting"},
			},
			want: "metcon",
		},
		{
			name: "mostly weightlifting",
			performances: []domain.WodifyPerformanceRow{
				{ComponentType: "Weightlifting"},
				{ComponentType: "Weightlifting"},
				{ComponentType: "Metcon"},
			},
			want: "strength",
		},
		{
			name: "mostly gymnastics",
			performances: []domain.WodifyPerformanceRow{
				{ComponentType: "Gymnastics"},
				{ComponentType: "Gymnastics"},
				{ComponentType: "Metcon"},
			},
			want: "gymnastics",
		},
		{
			name: "empty",
			performances: []domain.WodifyPerformanceRow{},
			want: "metcon", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.determineWorkoutType(tt.performances)
			if got != tt.want {
				t.Errorf("determineWorkoutType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWodifyImportService_FormatScoreValue(t *testing.T) {
	svc := &WodifyImportService{}

	tests := []struct {
		name      string
		parsed    *domain.ParsedPerformanceResult
		scoreType string
		want      *string
	}{
		{
			name:      "time format with hours",
			parsed:    &domain.ParsedPerformanceResult{TimeSeconds: intPtr(3930)}, // 1:05:30
			scoreType: "Time (HH:MM:SS)",
			want:      stringPtr("01:05:30"),
		},
		{
			name:      "time format without hours",
			parsed:    &domain.ParsedPerformanceResult{TimeSeconds: intPtr(330)}, // 5:30
			scoreType: "Time (HH:MM:SS)",
			want:      stringPtr("05:30"),
		},
		{
			name:      "rounds and reps",
			parsed:    &domain.ParsedPerformanceResult{Rounds: intPtr(7), Reps: intPtr(3)},
			scoreType: "Rounds+Reps",
			want:      stringPtr("7 rounds + 3 reps"),
		},
		{
			name:      "rounds only",
			parsed:    &domain.ParsedPerformanceResult{Rounds: intPtr(10)},
			scoreType: "Rounds+Reps",
			want:      stringPtr("10 rounds"),
		},
		{
			name:      "reps only",
			parsed:    &domain.ParsedPerformanceResult{Reps: intPtr(50)},
			scoreType: "Rounds+Reps",
			want:      stringPtr("50 reps"),
		},
		{
			name:      "max weight",
			parsed:    &domain.ParsedPerformanceResult{Weight: floatPtr(225.0)},
			scoreType: "Max Weight",
			want:      stringPtr("225 lbs"),
		},
		{
			name:      "empty result",
			parsed:    &domain.ParsedPerformanceResult{},
			scoreType: "Time (HH:MM:SS)",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.formatScoreValue(tt.parsed, tt.scoreType)
			if tt.want == nil {
				if got != nil {
					t.Errorf("formatScoreValue() = %q, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("formatScoreValue() = nil, want %q", *tt.want)
				} else if *got != *tt.want {
					t.Errorf("formatScoreValue() = %q, want %q", *got, *tt.want)
				}
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"TRUE", true},
		{"true", true},
		{"True", true},
		{" TRUE ", true},
		{"FALSE", false},
		{"false", false},
		{"", false},
		{"yes", false},
		{"1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseBool(tt.input)
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestWodifyImportService_ConfirmImport_WithExistingMovementAndWOD(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	// Add existing movement and WOD
	movementRepo.movements[1] = &domain.Movement{
		ID:         1,
		Name:       "Back Squat",
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}
	wodRepo.wods[1] = &domain.WOD{
		ID:         1,
		Name:       "Fran",
		IsStandard: true,
	}

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	result, err := svc.ConfirmImport(reader, 1, false, false)
	if err != nil {
		t.Fatalf("ConfirmImport returned error: %v", err)
	}

	// Should use existing movement and WOD
	if result.MovementsCreated != 0 {
		t.Errorf("MovementsCreated = %d, want 0 (should use existing)", result.MovementsCreated)
	}
	if result.WODsCreated != 0 {
		t.Errorf("WODsCreated = %d, want 0 (should use existing)", result.WODsCreated)
	}
}

func TestWodifyImportService_PreviewImport_PRDetection(t *testing.T) {
	userRepo := newMockUserRepo()
	movementRepo := newMockMovementRepo()
	wodRepo := newMockWODRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := newMockUserWorkoutMovementRepo()
	userWorkoutWODRepo := newMockUserWorkoutWODRepo()

	svc := NewWodifyImportService(
		userRepo,
		movementRepo,
		wodRepo,
		userWorkoutRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
	)

	// CSV with a PR row
	csv := createValidWodifyCSV()
	reader := strings.NewReader(csv)

	preview, err := svc.PreviewImport(reader, 1)
	if err != nil {
		t.Fatalf("PreviewImport returned error: %v", err)
	}

	// Workout summary should indicate PRs
	if len(preview.WorkoutSummary) == 0 {
		t.Fatal("Expected workout summary")
	}
	if !preview.WorkoutSummary[0].HasPRs {
		t.Error("WorkoutSummary should have HasPRs=true")
	}
}

// Helper function stringPtr is defined in test_helpers.go
