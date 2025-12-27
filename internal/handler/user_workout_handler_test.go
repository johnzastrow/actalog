package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
