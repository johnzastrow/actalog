package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
)

func TestUserWorkoutHandler_LogWorkout_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodPost, "/api/workouts", `{"workout_id": 1, "workout_date": "2024-01-15"}`)
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_LogWorkout_InvalidJSON(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestUserWorkoutHandler_LogWorkout_MissingWorkoutIDAndName(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_date": "2024-01-15"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Either workout_id or workout_name is required")
}

func TestUserWorkoutHandler_LogWorkout_MissingDate(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_id": 1}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Workout date is required")
}

func TestUserWorkoutHandler_LogWorkout_InvalidDateFormat(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_id": 1, "workout_date": "invalid-date"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout date format")
}

func TestUserWorkoutHandler_GetLoggedWorkout_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/1", "")
	rr := httptest.NewRecorder()

	handler.GetLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_GetLoggedWorkout_InvalidID(t *testing.T) {
	handler := &UserWorkoutHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestUserWorkoutHandler_ListLoggedWorkouts_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts", "")
	rr := httptest.NewRecorder()

	handler.ListLoggedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_ListLoggedWorkouts_InvalidDateRange(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts?start_date=invalid&end_date=also-invalid", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListLoggedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid date format")
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodPut, "/api/workouts/1", `{"notes": "Updated notes"}`)
	rr := httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_InvalidID(t *testing.T) {
	handler := &UserWorkoutHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPut, "/api/workouts/abc", `{"notes": "Updated"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_InvalidJSON(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/workouts/1", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	// Without chi router context, URL param is empty -> Invalid workout ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestUserWorkoutHandler_DeleteLoggedWorkout_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodDelete, "/api/workouts/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_DeleteLoggedWorkout_InvalidID(t *testing.T) {
	handler := &UserWorkoutHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/workouts/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestUserWorkoutHandler_GetMonthlyStats_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/stats/monthly?year=2024&month=1", "")
	rr := httptest.NewRecorder()

	handler.GetMonthlyStats(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_GetMonthlyStats_MissingParams(t *testing.T) {
	handler := &UserWorkoutHandler{}

	tests := []struct {
		name      string
		query     string
		wantError string
	}{
		{
			name:      "missing both",
			query:     "",
			wantError: "Year and month are required",
		},
		{
			name:      "missing year",
			query:     "?month=1",
			wantError: "Year and month are required",
		},
		{
			name:      "missing month",
			query:     "?year=2024",
			wantError: "Year and month are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/stats/monthly"+tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.GetMonthlyStats(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestUserWorkoutHandler_GetMonthlyStats_InvalidParams(t *testing.T) {
	handler := &UserWorkoutHandler{}

	tests := []struct {
		name      string
		query     string
		wantError string
	}{
		{
			name:      "invalid year",
			query:     "?year=abc&month=1",
			wantError: "Invalid year or month",
		},
		{
			name:      "invalid month",
			query:     "?year=2024&month=abc",
			wantError: "Invalid year or month",
		},
		{
			name:      "month too low",
			query:     "?year=2024&month=0",
			wantError: "Invalid year or month",
		},
		{
			name:      "month too high",
			query:     "?year=2024&month=13",
			wantError: "Invalid year or month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/stats/monthly"+tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.GetMonthlyStats(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestUserWorkoutHandler_GetPersonalRecords_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/personal-records", "")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_RetroactiveFlagPRs_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodPost, "/api/workouts/retroactive-prs", "")
	rr := httptest.NewRecorder()

	handler.RetroactiveFlagPRs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_GetActiveUsersStats_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/stats/active-users-this-month", "")
	rr := httptest.NewRecorder()

	handler.GetActiveUsersStats(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNewUserWorkoutHandler(t *testing.T) {
	handler := NewUserWorkoutHandler(nil, nil)
	if handler == nil {
		t.Error("NewUserWorkoutHandler should return a non-nil handler")
	}
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_ValidIDInvalidJSON(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/workouts/1", "{bad json", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Removed 35 panic-expectation tests:
// - TestUserWorkoutHandler_GetLoggedWorkout_ValidIDNilService
// - TestUserWorkoutHandler_DeleteLoggedWorkout_ValidIDNilService
// - TestUserWorkoutHandler_GetPersonalRecords_WithLimitParams
// - TestUserWorkoutHandler_ListLoggedWorkouts_WithPaginationParams
// - TestUserWorkoutHandler_ListLoggedWorkouts_ValidDateRange
// - TestUserWorkoutHandler_GetMonthlyStats_ValidParams
// - TestUserWorkoutHandler_RetroactiveFlagPRs_ValidAuth
// - TestUserWorkoutHandler_GetActiveUsersStats_ValidAuth
// - TestUserWorkoutHandler_LogWorkout_ValidInputWithWorkoutID
// - TestUserWorkoutHandler_LogWorkout_ValidInputWithWorkoutName
// - TestUserWorkoutHandler_LogWorkout_WithMovements
// - TestUserWorkoutHandler_LogWorkout_WithWODs
// - TestUserWorkoutHandler_UpdateLoggedWorkout_ValidInputNilService
// - TestUserWorkoutHandler_UpdateLoggedWorkout_WithMovements
// - TestUserWorkoutHandler_UpdateLoggedWorkout_WithWODs
// - TestUserWorkoutHandler_UpdateLoggedWorkout_WithMovementsAndWODs
// - TestUserWorkoutHandler_UpdateLoggedWorkout_DifferentIDs
// - TestUserWorkoutHandler_UpdateLoggedWorkout_WithWorkoutName
// - TestUserWorkoutHandler_UpdateLoggedWorkout_WithWorkoutType
// - TestUserWorkoutHandler_UpdateLoggedWorkout_FullMovementData
// - TestUserWorkoutHandler_UpdateLoggedWorkout_FullWODData
// - TestUserWorkoutHandler_UpdateLoggedWorkout_MultipleMovements
// - TestUserWorkoutHandler_UpdateLoggedWorkout_MultipleWODs
// - TestUserWorkoutHandler_DeleteLoggedWorkout_DifferentIDs
// - TestUserWorkoutHandler_GetLoggedWorkout_DifferentIDs
// - TestUserWorkoutHandler_GetPersonalRecords_DifferentUserIDs
// - TestUserWorkoutHandler_RetroactiveFlagPRs_DifferentUserIDs
// - TestUserWorkoutHandler_GetActiveUsersStats_DifferentUserIDs
// - TestUserWorkoutHandler_GetMonthlyStats_DifferentMonths
// - TestUserWorkoutHandler_LogWorkout_FullMovementData
// - TestUserWorkoutHandler_LogWorkout_FullWODData
// - TestUserWorkoutHandler_LogWorkout_WithAllOptionalFields
// - TestUserWorkoutHandler_LogWorkout_AdHocWithMovementsAndWODs
// - TestUserWorkoutHandler_ListLoggedWorkouts_OnlyStartDate
// - TestUserWorkoutHandler_ListLoggedWorkouts_OnlyEndDate
// These tests verified nil pointer panics, not business logic.

// ===== Tests with real database =====

func createTestUserWorkoutHandler(t *testing.T) (*UserWorkoutHandler, int64, func()) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	userWorkoutRepo := repository.NewUserWorkoutRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)
	userWorkoutMovementRepo := repository.NewUserWorkoutMovementRepository(db)
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)
	wodRepo := repository.NewWODRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db, "sqlite3")
	movementRepo := repository.NewMovementRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userSettingsRepo := repository.NewSQLiteUserSettingsRepository(db)

	// Create test user
	now := time.Now()
	testUser := &domain.User{
		Email:        "workouttest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		cleanup()
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test movement
	movement := &domain.Movement{
		Name:        "Back Squat",
		Type:        "weightlifting",
		Description: "Barbell back squat",
		IsStandard:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := movementRepo.Create(movement); err != nil {
		cleanup()
		t.Fatalf("Failed to create movement: %v", err)
	}

	// Create test WOD
	wod := &domain.WOD{
		Name:        "Fran",
		Type:        "For Time",
		ScoreType:   "Time (HH:MM:SS)",
		Description: "21-15-9 Thrusters and Pull-ups",
		IsStandard:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := wodRepo.Create(wod); err != nil {
		cleanup()
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create test workout template
	testNotes := "Test workout notes"
	workout := &domain.Workout{
		Name:      "Test Workout",
		Notes:     &testNotes,
		CreatedBy: &testUser.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := workoutRepo.Create(workout); err != nil {
		cleanup()
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Create notification service (required by UserWorkoutService)
	// Pass nil for emailService since we won't send emails in tests
	notificationService := service.NewNotificationService(notificationRepo, orgRepo, userRepo, userSettingsRepo, nil)

	// Create UserWorkoutService
	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		workoutMovementRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
		wodRepo,
		auditLogRepo,
		movementRepo,
		notificationService,
		userRepo,
		orgRepo,
	)

	handler := NewUserWorkoutHandler(userWorkoutService, createTestLogger())
	return handler, testUser.ID, cleanup
}

func TestUserWorkoutHandler_LogWorkout_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_LogWorkout_WithAdHocName(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	body := `{"workout_name": "My Custom Workout", "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_ListLoggedWorkouts_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Now list workouts
	req = createAuthenticatedRequest(http.MethodGet, "/api/workouts", "", userID, "workouttest@example.com", "user")
	rr = httptest.NewRecorder()

	handler.ListLoggedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	if !strings.Contains(rr.Body.String(), "workouts") {
		t.Error("Response should contain 'workouts' field")
	}
}

func TestUserWorkoutHandler_ListLoggedWorkouts_WithDateRange(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// Log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)

	// List with date range
	req = createAuthenticatedRequest(http.MethodGet, "/api/workouts?start_date=2024-01-01&end_date=2024-01-31", "", userID, "workouttest@example.com", "user")
	rr = httptest.NewRecorder()

	handler.ListLoggedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserWorkoutHandler_GetLoggedWorkout_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Now get the workout
	req = createAuthenticatedRequest(http.MethodGet, "/api/workouts/1", "", userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr = httptest.NewRecorder()

	handler.GetLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_GetLoggedWorkout_NotFound(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/999", "", userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetLoggedWorkout(rr, req)

	// Handler returns 500 with error message for not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "not found")
}

func TestUserWorkoutHandler_DeleteLoggedWorkout_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Now delete it
	req = createAuthenticatedRequest(http.MethodDelete, "/api/workouts/1", "", userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr = httptest.NewRecorder()

	handler.DeleteLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "deleted successfully")
}

func TestUserWorkoutHandler_DeleteLoggedWorkout_NotFound(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/workouts/999", "", userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.DeleteLoggedWorkout(rr, req)

	// Handler returns 500 with error message for not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "not found")
}

func TestUserWorkoutHandler_GetMonthlyStats_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// Log some workouts
	for i := 1; i <= 3; i++ {
		body := `{"workout_id": 1, "workout_date": "2024-01-` + string(rune('0'+i)) + `5"}`
		req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
		rr := httptest.NewRecorder()
		handler.LogWorkout(rr, req)
	}

	// Get monthly stats
	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/stats/monthly?year=2024&month=1", "", userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMonthlyStats(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_GetPersonalRecords_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/personal-records", "", userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_GetPersonalRecords_WithLimit(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/personal-records?limit=5", "", userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserWorkoutHandler_RetroactiveFlagPRs_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts/retroactive-prs", "", userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RetroactiveFlagPRs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "flagged")
}

func TestUserWorkoutHandler_GetActiveUsersStats_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/stats/active-users-this-month", "", userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetActiveUsersStats(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_Success(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Now update it (total_time is in seconds as integer)
	updateBody := `{"workout_name": "Updated Workout", "notes": "Updated notes", "total_time": 2700}`
	req = createAuthenticatedRequest(http.MethodPut, "/api/workouts/1", updateBody, userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr = httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "updated")
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_NotFound(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	updateBody := `{"workout_name": "Updated Workout"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/workouts/999", updateBody, userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_WithMovements(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Update with movements
	updateBody := `{"workout_name": "Updated", "movements": [{"movement_id": 1, "sets": 5, "reps": 5, "weight": 225.0}]}`
	req = createAuthenticatedRequest(http.MethodPut, "/api/workouts/1", updateBody, userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr = httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserWorkoutHandler_UpdateLoggedWorkout_WithWODs(t *testing.T) {
	handler, userID, cleanup := createTestUserWorkoutHandler(t)
	defer cleanup()

	// First, log a workout
	body := `{"workout_id": 1, "workout_date": "2024-01-15"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, userID, "workouttest@example.com", "user")
	rr := httptest.NewRecorder()
	handler.LogWorkout(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Update with WODs (Fran is time-based, need time_seconds)
	updateBody := `{"workout_name": "Updated", "wods": [{"wod_id": 1, "time_seconds": 330}]}`
	req = createAuthenticatedRequest(http.MethodPut, "/api/workouts/1", updateBody, userID, "workouttest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr = httptest.NewRecorder()

	handler.UpdateLoggedWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}
