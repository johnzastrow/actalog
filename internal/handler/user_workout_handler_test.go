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

func TestNewUserWorkoutHandler(t *testing.T) {
	// Test constructor with nil dependencies
	handler := NewUserWorkoutHandler(nil, nil)

	if handler == nil {
		t.Error("NewUserWorkoutHandler() should not return nil")
	}

	if handler.userWorkoutService != nil {
		t.Error("userWorkoutService should be nil when passed nil")
	}

	if handler.logger != nil {
		t.Error("logger should be nil when passed nil")
	}
}

func TestLogWorkoutRequest_Struct(t *testing.T) {
	workoutID := int64(123)
	workoutName := "Morning WOD"
	workoutType := "AMRAP"
	totalTime := 3600
	notes := "Great workout"

	req := LogWorkoutRequest{
		WorkoutID:   &workoutID,
		WorkoutName: &workoutName,
		WorkoutDate: "2024-01-15",
		WorkoutType: &workoutType,
		TotalTime:   &totalTime,
		Notes:       &notes,
		Movements:   []MovementPerformance{},
		WODs:        []WODPerformance{},
	}

	if req.WorkoutID == nil || *req.WorkoutID != 123 {
		t.Errorf("WorkoutID = %v, want 123", req.WorkoutID)
	}
	if req.WorkoutName == nil || *req.WorkoutName != "Morning WOD" {
		t.Errorf("WorkoutName = %v, want 'Morning WOD'", req.WorkoutName)
	}
	if req.WorkoutDate != "2024-01-15" {
		t.Errorf("WorkoutDate = %q, want '2024-01-15'", req.WorkoutDate)
	}
}

func TestLogWorkoutRequest_NilOptionalFields(t *testing.T) {
	req := LogWorkoutRequest{
		WorkoutDate: "2024-01-15",
	}

	if req.WorkoutID != nil {
		t.Error("WorkoutID should be nil")
	}
	if req.WorkoutName != nil {
		t.Error("WorkoutName should be nil")
	}
	if req.WorkoutType != nil {
		t.Error("WorkoutType should be nil")
	}
	if req.TotalTime != nil {
		t.Error("TotalTime should be nil")
	}
	if req.Notes != nil {
		t.Error("Notes should be nil")
	}
}

func TestMovementPerformance_Struct(t *testing.T) {
	sets := 5
	reps := 10
	weight := 135.5
	time := 120
	distance := 500.0

	mp := MovementPerformance{
		MovementID: 1,
		Sets:       &sets,
		Reps:       &reps,
		Weight:     &weight,
		Time:       &time,
		Distance:   &distance,
		Notes:      "Felt strong",
		OrderIndex: 0,
	}

	if mp.MovementID != 1 {
		t.Errorf("MovementID = %d, want 1", mp.MovementID)
	}
	if mp.Sets == nil || *mp.Sets != 5 {
		t.Errorf("Sets = %v, want 5", mp.Sets)
	}
	if mp.Reps == nil || *mp.Reps != 10 {
		t.Errorf("Reps = %v, want 10", mp.Reps)
	}
	if mp.Weight == nil || *mp.Weight != 135.5 {
		t.Errorf("Weight = %v, want 135.5", mp.Weight)
	}
	if mp.Notes != "Felt strong" {
		t.Errorf("Notes = %q, want 'Felt strong'", mp.Notes)
	}
}

func TestMovementPerformance_NilFields(t *testing.T) {
	mp := MovementPerformance{
		MovementID: 1,
		OrderIndex: 0,
	}

	if mp.Sets != nil {
		t.Error("Sets should be nil")
	}
	if mp.Reps != nil {
		t.Error("Reps should be nil")
	}
	if mp.Weight != nil {
		t.Error("Weight should be nil")
	}
	if mp.Time != nil {
		t.Error("Time should be nil")
	}
	if mp.Distance != nil {
		t.Error("Distance should be nil")
	}
}

func TestWODPerformance_Struct(t *testing.T) {
	scoreType := "Time"
	scoreValue := "10:30"
	timeSeconds := 630
	rounds := 5
	reps := 10
	weight := 95.0

	wp := WODPerformance{
		WODID:       1,
		ScoreType:   &scoreType,
		ScoreValue:  &scoreValue,
		TimeSeconds: &timeSeconds,
		Rounds:      &rounds,
		Reps:        &reps,
		Weight:      &weight,
		Notes:       "PR attempt",
		OrderIndex:  0,
	}

	if wp.WODID != 1 {
		t.Errorf("WODID = %d, want 1", wp.WODID)
	}
	if wp.ScoreType == nil || *wp.ScoreType != "Time" {
		t.Errorf("ScoreType = %v, want 'Time'", wp.ScoreType)
	}
	if wp.ScoreValue == nil || *wp.ScoreValue != "10:30" {
		t.Errorf("ScoreValue = %v, want '10:30'", wp.ScoreValue)
	}
	if wp.TimeSeconds == nil || *wp.TimeSeconds != 630 {
		t.Errorf("TimeSeconds = %v, want 630", wp.TimeSeconds)
	}
}

func TestWODPerformance_NilFields(t *testing.T) {
	wp := WODPerformance{
		WODID:      1,
		OrderIndex: 0,
	}

	if wp.ScoreType != nil {
		t.Error("ScoreType should be nil")
	}
	if wp.ScoreValue != nil {
		t.Error("ScoreValue should be nil")
	}
	if wp.TimeSeconds != nil {
		t.Error("TimeSeconds should be nil")
	}
	if wp.Rounds != nil {
		t.Error("Rounds should be nil")
	}
	if wp.Reps != nil {
		t.Error("Reps should be nil")
	}
	if wp.Weight != nil {
		t.Error("Weight should be nil")
	}
}

func TestUpdateLoggedWorkoutRequest_Struct(t *testing.T) {
	workoutName := "Updated Workout"
	workoutType := "For Time"
	totalTime := 1800
	notes := "Updated notes"

	req := UpdateLoggedWorkoutRequest{
		WorkoutName: &workoutName,
		WorkoutType: &workoutType,
		TotalTime:   &totalTime,
		Notes:       &notes,
		Movements:   []MovementPerformance{},
		WODs:        []WODPerformance{},
	}

	if req.WorkoutName == nil || *req.WorkoutName != "Updated Workout" {
		t.Errorf("WorkoutName = %v, want 'Updated Workout'", req.WorkoutName)
	}
	if req.WorkoutType == nil || *req.WorkoutType != "For Time" {
		t.Errorf("WorkoutType = %v, want 'For Time'", req.WorkoutType)
	}
	if req.TotalTime == nil || *req.TotalTime != 1800 {
		t.Errorf("TotalTime = %v, want 1800", req.TotalTime)
	}
}

func TestUserWorkoutResponse_Struct(t *testing.T) {
	workoutID := int64(100)
	workoutType := "AMRAP"
	totalTime := 1200
	notes := "Test notes"
	workoutNotes := "Template notes"

	resp := UserWorkoutResponse{
		ID:           1,
		UserID:       10,
		WorkoutID:    &workoutID,
		WorkoutName:  "Test Workout",
		WorkoutDate:  "2024-01-15",
		WorkoutType:  &workoutType,
		TotalTime:    &totalTime,
		Notes:        &notes,
		CreatedAt:    "2024-01-15T10:00:00Z",
		UpdatedAt:    "2024-01-15T10:00:00Z",
		WorkoutNotes: &workoutNotes,
	}

	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}
	if resp.UserID != 10 {
		t.Errorf("UserID = %d, want 10", resp.UserID)
	}
	if resp.WorkoutID == nil || *resp.WorkoutID != 100 {
		t.Errorf("WorkoutID = %v, want 100", resp.WorkoutID)
	}
	if resp.WorkoutName != "Test Workout" {
		t.Errorf("WorkoutName = %q, want 'Test Workout'", resp.WorkoutName)
	}
}

func TestUserWorkoutResponse_NilFields(t *testing.T) {
	resp := UserWorkoutResponse{
		ID:          1,
		UserID:      10,
		WorkoutName: "Ad-hoc Workout",
		WorkoutDate: "2024-01-15",
		CreatedAt:   "2024-01-15T10:00:00Z",
		UpdatedAt:   "2024-01-15T10:00:00Z",
	}

	if resp.WorkoutID != nil {
		t.Error("WorkoutID should be nil for ad-hoc workouts")
	}
	if resp.WorkoutType != nil {
		t.Error("WorkoutType should be nil")
	}
	if resp.TotalTime != nil {
		t.Error("TotalTime should be nil")
	}
	if resp.Notes != nil {
		t.Error("Notes should be nil")
	}
}

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

func TestUserWorkoutHandler_LogWorkout_ZeroWorkoutID(t *testing.T) {
	handler := &UserWorkoutHandler{}

	// Zero workout_id with no name
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_id": 0, "workout_date": "2024-01-15"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LogWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Either workout_id or workout_name is required")
}

func TestUserWorkoutHandler_LogWorkout_EmptyWorkoutName(t *testing.T) {
	handler := &UserWorkoutHandler{}

	// Empty workout_name with no ID
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_name": "", "workout_date": "2024-01-15"}`, 1, "test@example.com", "user")
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

func TestUserWorkoutHandler_LogWorkout_InvalidDateFormats(t *testing.T) {
	handler := &UserWorkoutHandler{}

	testCases := []struct {
		name string
		date string
	}{
		{"wrong format", "15-01-2024"},
		{"partial date", "2024-01"},
		{"slash format", "2024/01/15"},
		{"text format", "January 15, 2024"},
		{"invalid month", "2024-13-01"},
		{"invalid day", "2024-01-32"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"workout_id": 1, "workout_date": "` + tc.date + `"}`
			req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.LogWorkout(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, "Invalid workout date format")
		})
	}
}

func TestUserWorkoutHandler_LogWorkout_NilService(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts", `{"workout_id": 1, "workout_date": "2024-01-15"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.LogWorkout(rr, req)
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

func TestUserWorkoutHandler_ListLoggedWorkouts_WithLimitOffset(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts?limit=50&offset=10", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.ListLoggedWorkouts(rr, req)
}

func TestUserWorkoutHandler_ListLoggedWorkouts_WithInvalidLimitOffset(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	// Invalid limit/offset should use defaults
	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts?limit=abc&offset=xyz", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.ListLoggedWorkouts(rr, req)
}

func TestUserWorkoutHandler_ListLoggedWorkouts_NilService(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.ListLoggedWorkouts(rr, req)
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

func TestUserWorkoutHandler_GetMonthlyStats_NilService(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/stats/monthly?year=2024&month=1", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.GetMonthlyStats(rr, req)
}

func TestUserWorkoutHandler_GetPersonalRecords_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/personal-records", "")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_GetPersonalRecords_NilService(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/personal-records", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestUserWorkoutHandler_RetroactiveFlagPRs_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodPost, "/api/workouts/retroactive-prs", "")
	rr := httptest.NewRecorder()

	handler.RetroactiveFlagPRs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserWorkoutHandler_RetroactiveFlagPRs_NilService(t *testing.T) {
	handler := &UserWorkoutHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts/retroactive-prs", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userWorkoutService")
		}
	}()

	handler.RetroactiveFlagPRs(rr, req)
}

func TestUserWorkoutHandler_GetActiveUsersStats_Unauthorized(t *testing.T) {
	handler := &UserWorkoutHandler{}

	req := createTestRequest(http.MethodGet, "/api/stats/active-users-this-month", "")
	rr := httptest.NewRecorder()

	handler.GetActiveUsersStats(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
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
		Role:         "athlete",
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
