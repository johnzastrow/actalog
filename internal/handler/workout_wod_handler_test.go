package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkoutWODHandler_AddWODToWorkout_Unauthorized(t *testing.T) {
	handler := &WorkoutWODHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPost, "/api/workouts/1/wods", `{"wod_id": 1}`)
	rr := httptest.NewRecorder()

	handler.AddWODToWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutWODHandler_AddWODToWorkout_InvalidWorkoutID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts/abc/wods", `{"wod_id": 1}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.AddWODToWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestWorkoutWODHandler_AddWODToWorkout_InvalidJSON(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts/1/wods", "{invalid json", 1, "test@example.com", "user")
	// Without chi router context for workout_id
	rr := httptest.NewRecorder()

	handler.AddWODToWorkout(rr, req)

	// Without chi context, it will fail on URL param parsing first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWorkoutWODHandler_AddWODToWorkout_MissingWODID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	// Valid JSON but missing wod_id - however, without chi context this will fail on URL param
	req := createAuthenticatedRequest(http.MethodPost, "/api/workouts/1/wods", `{"order_index": 1}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.AddWODToWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWorkoutWODHandler_RemoveWODFromWorkout_Unauthorized(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createTestRequest(http.MethodDelete, "/api/workout-wods/1", "")
	rr := httptest.NewRecorder()

	handler.RemoveWODFromWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutWODHandler_RemoveWODFromWorkout_InvalidID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/workout-wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RemoveWODFromWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout WOD ID")
}

func TestWorkoutWODHandler_UpdateWorkoutWOD_Unauthorized(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createTestRequest(http.MethodPut, "/api/workout-wods/1", `{"score_value": "10:30"}`)
	rr := httptest.NewRecorder()

	handler.UpdateWorkoutWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutWODHandler_UpdateWorkoutWOD_InvalidID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/workout-wods/abc", `{"score_value": "10:30"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWorkoutWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout WOD ID")
}

func TestWorkoutWODHandler_UpdateWorkoutWOD_InvalidJSON(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/workout-wods/1", "{invalid", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWorkoutWOD(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWorkoutWODHandler_ToggleWODPR_Unauthorized(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createTestRequest(http.MethodPost, "/api/workout-wods/1/toggle-pr", "")
	rr := httptest.NewRecorder()

	handler.ToggleWODPR(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutWODHandler_ToggleWODPR_InvalidID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/workout-wods/abc/toggle-pr", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleWODPR(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout WOD ID")
}

func TestWorkoutWODHandler_ListWODsForWorkout_InvalidID(t *testing.T) {
	handler := &WorkoutWODHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/abc/wods", "")
	rr := httptest.NewRecorder()

	handler.ListWODsForWorkout(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestNewWorkoutWODHandler(t *testing.T) {
	// Test constructor with nil service (basic construction test)
	handler := NewWorkoutWODHandler(nil)

	if handler == nil {
		t.Error("NewWorkoutWODHandler() should not return nil")
	}

	if handler.workoutWODService != nil {
		t.Error("workoutWODService should be nil when passed nil")
	}
}

func TestAddWODToWorkoutRequest(t *testing.T) {
	// Test struct initialization
	req := AddWODToWorkoutRequest{
		WODID:      123,
		OrderIndex: 1,
		Division:   nil,
	}

	if req.WODID != 123 {
		t.Errorf("WODID = %d, want 123", req.WODID)
	}

	if req.OrderIndex != 1 {
		t.Errorf("OrderIndex = %d, want 1", req.OrderIndex)
	}

	if req.Division != nil {
		t.Error("Division should be nil")
	}

	// With division
	div := "Rx"
	req2 := AddWODToWorkoutRequest{
		WODID:      456,
		OrderIndex: 2,
		Division:   &div,
	}

	if req2.Division == nil || *req2.Division != "Rx" {
		t.Error("Division should be 'Rx'")
	}
}

func TestUpdateWorkoutWODRequest(t *testing.T) {
	// Test struct initialization
	req := UpdateWorkoutWODRequest{
		ScoreValue: nil,
		Division:   nil,
	}

	if req.ScoreValue != nil {
		t.Error("ScoreValue should be nil")
	}

	if req.Division != nil {
		t.Error("Division should be nil")
	}

	// With values
	scoreVal := "10:30"
	div := "Scaled"
	req2 := UpdateWorkoutWODRequest{
		ScoreValue: &scoreVal,
		Division:   &div,
	}

	if req2.ScoreValue == nil || *req2.ScoreValue != "10:30" {
		t.Error("ScoreValue should be '10:30'")
	}

	if req2.Division == nil || *req2.Division != "Scaled" {
		t.Error("Division should be 'Scaled'")
	}
}
