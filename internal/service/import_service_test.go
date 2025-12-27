package service

import (
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewImportService(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	if svc == nil {
		t.Fatal("NewImportService returned nil")
	}
}

func TestImportService_PreviewWODImport_ValidCSV(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),21-15-9 Thrusters and Pull-ups,https://example.com,Classic benchmark,true,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result == nil {
		t.Fatal("PreviewWODImport() returned nil")
	}
	if result.TotalRows != 1 {
		t.Errorf("TotalRows = %d, want 1", result.TotalRows)
	}
	if result.ValidRows != 1 {
		t.Errorf("ValidRows = %d, want 1", result.ValidRows)
	}
	if result.InvalidRows != 0 {
		t.Errorf("InvalidRows = %d, want 0", result.InvalidRows)
	}
}

func TestImportService_PreviewWODImport_InvalidHeader(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `wrong,header,columns
Fran,CrossFit,Benchmark`

	_, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err == nil {
		t.Error("PreviewWODImport() should return error for invalid header")
	}
}

func TestImportService_PreviewWODImport_MissingRequiredFields(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Missing name, source, type
	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
,,,,,Test description,,,false,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if result.ValidRows != 0 {
		t.Errorf("ValidRows = %d, want 0", result.ValidRows)
	}
}

func TestImportService_PreviewWODImport_InvalidEnumValues(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Invalid source, type, regime, score_type values
	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Test WOD,InvalidSource,InvalidType,InvalidRegime,InvalidScore,Test,,,false,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}

	// Check that errors were recorded
	if len(result.Rows) == 0 || len(result.Rows[0].Errors) == 0 {
		t.Error("Expected validation errors to be recorded")
	}
}

func TestImportService_PreviewWODImport_DuplicateDetection(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing WOD
	existingWOD := &domain.WOD{Name: "Fran", IsStandard: true}
	_ = wodRepo.Create(existingWOD)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),21-15-9 Thrusters and Pull-ups,,,true,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.DuplicateRows != 1 {
		t.Errorf("DuplicateRows = %d, want 1", result.DuplicateRows)
	}
	if len(result.Rows) > 0 && !result.Rows[0].IsDuplicate {
		t.Error("Row should be marked as duplicate")
	}
}

func TestImportService_PreviewWODImport_NonAdminCannotImportStandard(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Non-admin trying to import standard WOD
	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),Test,,,true,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, false) // isAdmin = false
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
}

func TestImportService_ConfirmWODImport_CreateNew(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Grace,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),30 Clean and Jerks,https://example.com,Classic,true,`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), 1, true, false, false)
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify WOD was created
	createdWOD, _ := wodRepo.GetByName("Grace")
	if createdWOD == nil {
		t.Error("WOD should be created")
	}
}

func TestImportService_ConfirmWODImport_SkipDuplicates(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing WOD
	existingWOD := &domain.WOD{Name: "Fran", IsStandard: true, Source: "CrossFit", Type: "Benchmark", Regime: "Fastest Time", ScoreType: "Time (HH:MM:SS)"}
	_ = wodRepo.Create(existingWOD)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),Updated description,,,true,`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), 1, true, true, false) // skipDuplicates = true
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", result.SkippedCount)
	}
	if result.CreatedCount != 0 {
		t.Errorf("CreatedCount = %d, want 0", result.CreatedCount)
	}
}

func TestImportService_PreviewMovementImport_ValidCSV(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Barbell back squat,true,`

	result, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewMovementImport() error = %v", err)
	}
	if result == nil {
		t.Fatal("PreviewMovementImport() returned nil")
	}
	if result.TotalRows != 1 {
		t.Errorf("TotalRows = %d, want 1", result.TotalRows)
	}
	if result.ValidRows != 1 {
		t.Errorf("ValidRows = %d, want 1", result.ValidRows)
	}
}

func TestImportService_PreviewMovementImport_InvalidHeader(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `wrong,columns
Back Squat,weightlifting`

	_, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, true)
	if err == nil {
		t.Error("PreviewMovementImport() should return error for invalid header")
	}
}

func TestImportService_PreviewMovementImport_MissingRequiredFields(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Missing name and type
	csv := `name,type,description,is_standard,created_by_email
,,Test description,false,`

	result, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewMovementImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
}

func TestImportService_PreviewMovementImport_InvalidType(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,invalid_type,Test,false,`

	result, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewMovementImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
}

func TestImportService_PreviewMovementImport_DuplicateDetection(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing movement
	existingMovement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	_ = movementRepo.Create(existingMovement)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Updated description,true,`

	result, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewMovementImport() error = %v", err)
	}
	if result.DuplicateRows != 1 {
		t.Errorf("DuplicateRows = %d, want 1", result.DuplicateRows)
	}
}

func TestImportService_PreviewMovementImport_NonAdminCannotImportStandard(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Test,true,`

	result, err := svc.PreviewMovementImport(strings.NewReader(csv), 1, false) // isAdmin = false
	if err != nil {
		t.Errorf("PreviewMovementImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
}

func TestImportService_ConfirmMovementImport_CreateNew(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Deadlift,weightlifting,Conventional deadlift,true,`

	result, err := svc.ConfirmMovementImport(strings.NewReader(csv), 1, true, false, false)
	if err != nil {
		t.Errorf("ConfirmMovementImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify movement was created
	createdMovement, _ := movementRepo.GetByName("Deadlift")
	if createdMovement == nil {
		t.Error("Movement should be created")
	}
}

func TestImportService_ConfirmMovementImport_SkipDuplicates(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing movement
	existingMovement := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	_ = movementRepo.Create(existingMovement)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Updated description,true,`

	result, err := svc.ConfirmMovementImport(strings.NewReader(csv), 1, true, true, false) // skipDuplicates = true
	if err != nil {
		t.Errorf("ConfirmMovementImport() error = %v", err)
	}
	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", result.SkippedCount)
	}
}

func TestImportService_PreviewUserWorkoutImport_ValidJSON(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create movement for reference
	backSquat := &domain.Movement{Name: "Back Squat", Type: domain.MovementTypeWeightlifting, IsStandard: true}
	_ = movementRepo.Create(backSquat)

	json := `{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "0.5.1",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_type": "strength",
				"workout_name": "Morning Workout",
				"movements": [
					{
						"movement_name": "Back Squat",
						"movement_type": "weightlifting",
						"sets": 3,
						"reps": 5,
						"weight": 225.0
					}
				]
			}
		]
	}`

	result, err := svc.PreviewUserWorkoutImport([]byte(json), 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	if result == nil {
		t.Fatal("PreviewUserWorkoutImport() returned nil")
	}
	if result.TotalWorkouts != 1 {
		t.Errorf("TotalWorkouts = %d, want 1", result.TotalWorkouts)
	}
	if result.ValidWorkouts != 1 {
		t.Errorf("ValidWorkouts = %d, want 1", result.ValidWorkouts)
	}
}

func TestImportService_PreviewUserWorkoutImport_InvalidDate(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	json := `{
		"export_metadata": {"user_email": "test@example.com"},
		"user_workouts": [
			{
				"workout_date": "invalid-date",
				"workout_name": "Test"
			}
		]
	}`

	result, err := svc.PreviewUserWorkoutImport([]byte(json), 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	if result.InvalidWorkouts != 1 {
		t.Errorf("InvalidWorkouts = %d, want 1", result.InvalidWorkouts)
	}
}

func TestImportService_PreviewUserWorkoutImport_InvalidJSON(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	_, err := svc.PreviewUserWorkoutImport([]byte("invalid json"), 1)
	if err == nil {
		t.Error("PreviewUserWorkoutImport() should return error for invalid JSON")
	}
}

func TestImportService_PreviewUserWorkoutImport_MissingMovement(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Movement doesn't exist in repo
	json := `{
		"export_metadata": {"user_email": "test@example.com"},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"movements": [
					{
						"movement_name": "NonExistentMovement",
						"movement_type": "weightlifting"
					}
				]
			}
		]
	}`

	result, err := svc.PreviewUserWorkoutImport([]byte(json), 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	// Should have error about missing movement
	if len(result.Errors) == 0 {
		t.Error("Expected error about missing movement")
	}
}

func TestImportService_MultipleRows(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),21-15-9,,,true,
Grace,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),30 C&J,,,true,
Helen,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),3 rounds,,,true,`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.TotalRows != 3 {
		t.Errorf("TotalRows = %d, want 3", result.TotalRows)
	}
	if result.ValidRows != 3 {
		t.Errorf("ValidRows = %d, want 3", result.ValidRows)
	}
}

func TestImportService_PreviewWODImport_WithCreatedByEmail(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create a user
	user := &domain.User{Email: "test@example.com"}
	_ = userRepo.Create(user)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Custom WOD,Self-recorded,Self-created,AMRAP,Rounds+Reps,Custom workout,,,false,test@example.com`

	result, err := svc.PreviewWODImport(strings.NewReader(csv), 1, true)
	if err != nil {
		t.Errorf("PreviewWODImport() error = %v", err)
	}
	if result.ValidRows != 1 {
		t.Errorf("ValidRows = %d, want 1", result.ValidRows)
	}
	if len(result.Rows) > 0 && result.Rows[0].CreatedByEmail != "test@example.com" {
		t.Errorf("CreatedByEmail = %s, want test@example.com", result.Rows[0].CreatedByEmail)
	}
}

func TestEqualStringSlices(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{"equal slices", []string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{"different length", []string{"a", "b"}, []string{"a", "b", "c"}, false},
		{"different content", []string{"a", "b", "c"}, []string{"a", "x", "c"}, false},
		{"empty slices", []string{}, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := equalStringSlices(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("equalStringSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("contains() should return true for existing item")
	}

	if contains(slice, "orange") {
		t.Error("contains() should return false for non-existing item")
	}

	if contains([]string{}, "anything") {
		t.Error("contains() should return false for empty slice")
	}
}
