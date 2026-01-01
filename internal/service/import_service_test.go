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

// Additional tests for better coverage

func TestImportService_ConfirmWODImport_UpdateDuplicates_Standard(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing standard WOD
	existingWOD := &domain.WOD{
		Name:        "Fran",
		IsStandard:  true,
		Source:      "CrossFit",
		Type:        "Benchmark",
		Regime:      "Fastest Time",
		ScoreType:   "Time (HH:MM:SS)",
		Description: "Original description",
	}
	_ = wodRepo.Create(existingWOD)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),Updated description,https://example.com,Updated notes,true,`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), 1, true, false, true) // updateDuplicates = true
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.UpdatedCount != 1 {
		t.Errorf("UpdatedCount = %d, want 1", result.UpdatedCount)
	}
	if result.CreatedCount != 0 {
		t.Errorf("CreatedCount = %d, want 0", result.CreatedCount)
	}

	// Verify WOD was updated
	updatedWOD, _ := wodRepo.GetByName("Fran")
	if updatedWOD.Description != "Updated description" {
		t.Errorf("Description = %s, want Updated description", updatedWOD.Description)
	}
}

func TestImportService_ConfirmWODImport_UpdateDuplicates_Custom(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing custom WOD
	userID := int64(1)
	existingWOD := &domain.WOD{
		Name:        "My Custom WOD",
		IsStandard:  false,
		Source:      "Self-recorded",
		Type:        "Self-created",
		Regime:      "AMRAP",
		ScoreType:   "Rounds+Reps",
		Description: "Original description",
		CreatedBy:   &userID,
	}
	_ = wodRepo.Create(existingWOD)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
My Custom WOD,Self-recorded,Self-created,AMRAP,Rounds+Reps,Updated description,,,false,`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), userID, false, false, true) // updateDuplicates = true
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.UpdatedCount != 1 {
		t.Errorf("UpdatedCount = %d, want 1", result.UpdatedCount)
	}

	// Verify WOD was updated using regular Update (not UpdateStandard)
	updatedWOD, _ := wodRepo.GetByName("My Custom WOD")
	if updatedWOD.Description != "Updated description" {
		t.Errorf("Description = %s, want Updated description", updatedWOD.Description)
	}
}

func TestImportService_ConfirmWODImport_WithURLAndNotes(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Grace,CrossFit,Benchmark,Fastest Time,Time (HH:MM:SS),30 Clean and Jerks,https://crossfit.com/grace,Classic Olympic lift benchmark,true,`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), 1, true, false, false)
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify WOD was created with URL and Notes
	createdWOD, _ := wodRepo.GetByName("Grace")
	if createdWOD == nil {
		t.Fatal("WOD not created")
	}
	if createdWOD.URL == nil || *createdWOD.URL != "https://crossfit.com/grace" {
		t.Error("URL not set correctly")
	}
	if createdWOD.Notes == nil || *createdWOD.Notes != "Classic Olympic lift benchmark" {
		t.Error("Notes not set correctly")
	}
}

func TestImportService_ConfirmWODImport_WithCreatedByEmail(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	// Create a user to be referenced
	user := &domain.User{Email: "creator@example.com"}
	_ = userRepo.Create(user)

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Custom WOD,Self-recorded,Self-created,AMRAP,Rounds+Reps,Custom workout,,,false,creator@example.com`

	result, err := svc.ConfirmWODImport(strings.NewReader(csv), 2, false, false, false) // Different user ID
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify WOD was created with correct created_by
	createdWOD, _ := wodRepo.GetByName("Custom WOD")
	if createdWOD == nil {
		t.Fatal("WOD not created")
	}
	if createdWOD.CreatedBy == nil || *createdWOD.CreatedBy != user.ID {
		t.Error("CreatedBy not set correctly from email")
	}
}

func TestImportService_ConfirmWODImport_CustomWithoutEmail(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Custom WOD No Email,Self-recorded,Self-created,AMRAP,Rounds+Reps,Custom workout,,,false,`

	userID := int64(5)
	result, err := svc.ConfirmWODImport(strings.NewReader(csv), userID, false, false, false)
	if err != nil {
		t.Errorf("ConfirmWODImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify WOD was created with current user as created_by
	createdWOD, _ := wodRepo.GetByName("Custom WOD No Email")
	if createdWOD == nil {
		t.Fatal("WOD not created")
	}
	if createdWOD.CreatedBy == nil || *createdWOD.CreatedBy != userID {
		t.Error("CreatedBy should be set to current user when no email provided")
	}
}

func TestImportService_ConfirmMovementImport_UpdateDuplicates(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Create existing movement
	existingMovement := &domain.Movement{
		Name:        "Back Squat",
		Type:        domain.MovementTypeWeightlifting,
		Description: "Original description",
		IsStandard:  true,
	}
	_ = movementRepo.Create(existingMovement)

	csv := `name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Updated description,true,`

	result, err := svc.ConfirmMovementImport(strings.NewReader(csv), 1, true, false, true) // updateDuplicates = true
	if err != nil {
		t.Errorf("ConfirmMovementImport() error = %v", err)
	}
	if result.UpdatedCount != 1 {
		t.Errorf("UpdatedCount = %d, want 1", result.UpdatedCount)
	}
	if result.CreatedCount != 0 {
		t.Errorf("CreatedCount = %d, want 0", result.CreatedCount)
	}

	// Verify movement was updated
	updatedMovement, _ := movementRepo.GetByName("Back Squat")
	if updatedMovement.Description != "Updated description" {
		t.Errorf("Description = %s, want Updated description", updatedMovement.Description)
	}
}

func TestImportService_ConfirmMovementImport_WithCreatedByEmail(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	// Create a user to be referenced
	user := &domain.User{Email: "coach@example.com"}
	_ = userRepo.Create(user)

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Custom Move,gymnastics,Custom movement,false,coach@example.com`

	result, err := svc.ConfirmMovementImport(strings.NewReader(csv), 2, false, false, false)
	if err != nil {
		t.Errorf("ConfirmMovementImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify movement was created with correct created_by
	createdMovement, _ := movementRepo.GetByName("Custom Move")
	if createdMovement == nil {
		t.Fatal("Movement not created")
	}
	if createdMovement.CreatedBy == nil || *createdMovement.CreatedBy != user.ID {
		t.Error("CreatedBy not set correctly from email")
	}
}

func TestImportService_ConfirmMovementImport_CustomWithoutEmail(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	csv := `name,type,description,is_standard,created_by_email
Custom Move No Email,gymnastics,Custom movement,false,`

	userID := int64(7)
	result, err := svc.ConfirmMovementImport(strings.NewReader(csv), userID, false, false, false)
	if err != nil {
		t.Errorf("ConfirmMovementImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// Verify movement was created with current user as created_by
	createdMovement, _ := movementRepo.GetByName("Custom Move No Email")
	if createdMovement == nil {
		t.Fatal("Movement not created")
	}
	if createdMovement.CreatedBy == nil || *createdMovement.CreatedBy != userID {
		t.Error("CreatedBy should be set to current user when no email provided")
	}
}

func TestImportService_PreviewUserWorkoutImport_WithMissingMovement(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"movements": [
					{
						"movement_name": "Back Squat",
						"movement_type": "weightlifting"
					}
				]
			}
		]
	}`)

	result, err := svc.PreviewUserWorkoutImport(jsonData, 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	// Should warn about missing movement but still be valid
	if result.ValidWorkouts != 1 {
		t.Errorf("ValidWorkouts = %d, want 1", result.ValidWorkouts)
	}
	if len(result.Errors) == 0 {
		t.Error("Should have warning about missing movement")
	}
}

func TestImportService_PreviewUserWorkoutImport_WithMissingWOD(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"wods": [
					{
						"wod_name": "Fran",
						"wod_type": "Benchmark"
					}
				]
			}
		]
	}`)

	result, err := svc.PreviewUserWorkoutImport(jsonData, 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	// Should warn about missing WOD but still be valid
	if result.ValidWorkouts != 1 {
		t.Errorf("ValidWorkouts = %d, want 1", result.ValidWorkouts)
	}
	if len(result.Errors) == 0 {
		t.Error("Should have warning about missing WOD")
	}
}

func TestImportService_PreviewUserWorkoutImport_MissingMovementName(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"movements": [
					{
						"movement_name": "",
						"movement_type": "weightlifting"
					}
				]
			}
		]
	}`)

	result, err := svc.PreviewUserWorkoutImport(jsonData, 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	// Should mark as invalid when movement name is missing
	if result.InvalidWorkouts != 1 {
		t.Errorf("InvalidWorkouts = %d, want 1", result.InvalidWorkouts)
	}
}

func TestImportService_PreviewUserWorkoutImport_MissingWODName(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"wods": [
					{
						"wod_name": "",
						"wod_type": "Benchmark"
					}
				]
			}
		]
	}`)

	result, err := svc.PreviewUserWorkoutImport(jsonData, 1)
	if err != nil {
		t.Errorf("PreviewUserWorkoutImport() error = %v", err)
	}
	// Should mark as invalid when WOD name is missing
	if result.InvalidWorkouts != 1 {
		t.Errorf("InvalidWorkouts = %d, want 1", result.InvalidWorkouts)
	}
}

func TestImportService_ConfirmUserWorkoutImport_BasicSuccess(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_type": "strength",
				"workout_name": "Morning Workout",
				"total_time": 3600,
				"notes": "Felt good"
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.TotalWorkouts != 1 {
		t.Errorf("TotalWorkouts = %d, want 1", result.TotalWorkouts)
	}
	// ValidWorkouts includes all successful imports
	if result.ValidWorkouts < 1 {
		t.Errorf("ValidWorkouts = %d, want at least 1", result.ValidWorkouts)
	}
}

func TestImportService_ConfirmUserWorkoutImport_InvalidJSON(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{invalid json}`)

	_, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err == nil {
		t.Error("ConfirmUserWorkoutImport() should return error for invalid JSON")
	}
}

func TestImportService_ConfirmUserWorkoutImport_InvalidDate(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "not-a-date",
				"workout_name": "Morning Workout"
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.InvalidWorkouts != 1 {
		t.Errorf("InvalidWorkouts = %d, want 1", result.InvalidWorkouts)
	}
}

func TestImportService_ConfirmUserWorkoutImport_WithMovements(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	// Create a movement that exists
	existingMovement := &domain.Movement{
		Name:       "Back Squat",
		Type:       domain.MovementTypeWeightlifting,
		IsStandard: true,
	}
	_ = movementRepo.Create(existingMovement)

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_name": "Squat Day",
				"movements": [
					{
						"movement_name": "Back Squat",
						"movement_type": "weightlifting",
						"sets": 5,
						"reps": 5,
						"weight": 225.0,
						"is_pr": true,
						"order_index": 0
					}
				]
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.ValidWorkouts < 1 {
		t.Errorf("ValidWorkouts = %d, want at least 1", result.ValidWorkouts)
	}
}

func TestImportService_ConfirmUserWorkoutImport_WithWODs(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	// Create a WOD that exists
	existingWOD := &domain.WOD{
		Name:       "Fran",
		Type:       "Benchmark",
		IsStandard: true,
	}
	_ = wodRepo.Create(existingWOD)

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_name": "WOD Day",
				"wods": [
					{
						"wod_name": "Fran",
						"wod_type": "Benchmark",
						"score_type": "Time",
						"score_value": "3:45",
						"time_seconds": 225,
						"is_pr": true,
						"order_index": 0
					}
				]
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.ValidWorkouts < 1 {
		t.Errorf("ValidWorkouts = %d, want at least 1", result.ValidWorkouts)
	}
}

func TestImportService_ConfirmUserWorkoutImport_CreatesMissingMovement(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_name": "Squat Day",
				"movements": [
					{
						"movement_name": "New Movement",
						"movement_type": "weightlifting",
						"sets": 3,
						"reps": 10,
						"order_index": 0
					}
				]
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.MovementsCreated != 1 {
		t.Errorf("MovementsCreated = %d, want 1", result.MovementsCreated)
	}

	// Verify movement was created
	createdMovement, _ := movementRepo.GetByName("New Movement")
	if createdMovement == nil {
		t.Error("Movement should have been created")
	}
}

func TestImportService_ConfirmUserWorkoutImport_CreatesMissingWOD(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_name": "WOD Day",
				"wods": [
					{
						"wod_name": "New WOD",
						"wod_type": "Custom",
						"score_type": "Rounds+Reps",
						"rounds": 10,
						"reps": 5,
						"order_index": 0
					}
				]
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.WODsCreated != 1 {
		t.Errorf("WODsCreated = %d, want 1", result.WODsCreated)
	}

	// Verify WOD was created
	createdWOD, _ := wodRepo.GetByName("New WOD")
	if createdWOD == nil {
		t.Error("WOD should have been created")
	}
}

func TestImportService_ConfirmUserWorkoutImport_DefaultWorkoutName(t *testing.T) {
	wodRepo := newMockWODRepo()
	movementRepo := newMockMovementRepo()
	userRepo := newMockUserRepo()
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutMovementRepo := &mockUserWorkoutMovementRepo{}
	userWorkoutWODRepo := &mockUserWorkoutWODRepo{}

	svc := NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// No workout_name provided
	jsonData := []byte(`{
		"export_metadata": {
			"user_email": "test@example.com",
			"version": "1.0",
			"total_count": 1
		},
		"user_workouts": [
			{
				"workout_date": "2024-01-15",
				"workout_type": "strength"
			}
		]
	}`)

	result, err := svc.ConfirmUserWorkoutImport(jsonData, 1, false, false)
	if err != nil {
		t.Errorf("ConfirmUserWorkoutImport() error = %v", err)
	}
	if result.ValidWorkouts < 1 {
		t.Errorf("ValidWorkouts = %d, want at least 1", result.ValidWorkouts)
	}
}
