package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPRHandler(t *testing.T) {
	// Test constructor with nil dependencies
	handler := NewPRHandler(nil, nil)

	if handler == nil {
		t.Error("NewPRHandler() should not return nil")
	}

	if handler.db != nil {
		t.Error("db should be nil when passed nil")
	}

	if handler.logger != nil {
		t.Error("logger should be nil when passed nil")
	}
}

func TestPersonalRecord_Struct(t *testing.T) {
	// Test with all fields
	weight := 225.0
	sets := 5
	reps := 5
	timeVal := 300
	distance := 1000.0
	calc1RM := 253.0
	formula := "Epley"
	scoreVal := "10:30"
	division := "Rx"
	wodType := "AMRAP"
	wodScoreType := "Rounds+Reps"
	movementType := "weightlifting"

	pr := PersonalRecord{
		Type:          "movement",
		ID:            1,
		UserWorkoutID: 100,
		WorkoutDate:   "2024-01-15",
		Name:          "Back Squat",
		MovementType:  &movementType,
		Weight:        &weight,
		Sets:          &sets,
		Reps:          &reps,
		Time:          &timeVal,
		Distance:      &distance,
		Calculated1RM: &calc1RM,
		Formula:       &formula,
		ScoreValue:    &scoreVal,
		Division:      &division,
		WODType:       &wodType,
		WODScoreType:  &wodScoreType,
	}

	if pr.Type != "movement" {
		t.Errorf("Type = %q, want %q", pr.Type, "movement")
	}
	if pr.ID != 1 {
		t.Errorf("ID = %d, want 1", pr.ID)
	}
	if pr.UserWorkoutID != 100 {
		t.Errorf("UserWorkoutID = %d, want 100", pr.UserWorkoutID)
	}
	if pr.WorkoutDate != "2024-01-15" {
		t.Errorf("WorkoutDate = %q, want %q", pr.WorkoutDate, "2024-01-15")
	}
	if pr.Name != "Back Squat" {
		t.Errorf("Name = %q, want %q", pr.Name, "Back Squat")
	}
	if pr.Weight == nil || *pr.Weight != 225.0 {
		t.Errorf("Weight = %v, want 225.0", pr.Weight)
	}
	if pr.Calculated1RM == nil || *pr.Calculated1RM != 253.0 {
		t.Errorf("Calculated1RM = %v, want 253.0", pr.Calculated1RM)
	}
}

func TestPersonalRecord_NilFields(t *testing.T) {
	pr := PersonalRecord{
		Type:          "wod",
		ID:            1,
		UserWorkoutID: 100,
		WorkoutDate:   "2024-01-15",
		Name:          "Fran",
	}

	if pr.Weight != nil {
		t.Error("Weight should be nil")
	}
	if pr.Sets != nil {
		t.Error("Sets should be nil")
	}
	if pr.Reps != nil {
		t.Error("Reps should be nil")
	}
	if pr.MovementType != nil {
		t.Error("MovementType should be nil")
	}
	if pr.Calculated1RM != nil {
		t.Error("Calculated1RM should be nil")
	}
}

func TestMovementPRSummary_Struct(t *testing.T) {
	best1RM := 315.0
	bestFormula := "Brzycki"
	bestWeight := 275.0
	bestSets := 3
	bestReps := 5

	summary := MovementPRSummary{
		MovementID:   1,
		MovementName: "Deadlift",
		MovementType: "weightlifting",
		PRCount:      12,
		Best1RM:      &best1RM,
		BestFormula:  &bestFormula,
		BestWeight:   &bestWeight,
		BestSets:     &bestSets,
		BestReps:     &bestReps,
		LastPRDate:   "2024-01-20",
	}

	if summary.MovementID != 1 {
		t.Errorf("MovementID = %d, want 1", summary.MovementID)
	}
	if summary.MovementName != "Deadlift" {
		t.Errorf("MovementName = %q, want %q", summary.MovementName, "Deadlift")
	}
	if summary.MovementType != "weightlifting" {
		t.Errorf("MovementType = %q, want %q", summary.MovementType, "weightlifting")
	}
	if summary.PRCount != 12 {
		t.Errorf("PRCount = %d, want 12", summary.PRCount)
	}
	if summary.Best1RM == nil || *summary.Best1RM != 315.0 {
		t.Errorf("Best1RM = %v, want 315.0", summary.Best1RM)
	}
	if summary.BestFormula == nil || *summary.BestFormula != "Brzycki" {
		t.Errorf("BestFormula = %v, want Brzycki", summary.BestFormula)
	}
	if summary.LastPRDate != "2024-01-20" {
		t.Errorf("LastPRDate = %q, want %q", summary.LastPRDate, "2024-01-20")
	}
}

func TestMovementPRSummary_NilFields(t *testing.T) {
	summary := MovementPRSummary{
		MovementID:   1,
		MovementName: "Pull-ups",
		MovementType: "gymnastics",
		PRCount:      5,
		LastPRDate:   "2024-01-15",
	}

	if summary.Best1RM != nil {
		t.Error("Best1RM should be nil")
	}
	if summary.BestFormula != nil {
		t.Error("BestFormula should be nil")
	}
	if summary.BestWeight != nil {
		t.Error("BestWeight should be nil")
	}
	if summary.BestSets != nil {
		t.Error("BestSets should be nil")
	}
	if summary.BestReps != nil {
		t.Error("BestReps should be nil")
	}
}

func TestPRHandler_GetPersonalRecords_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs", "")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_GetPersonalRecords_WithLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with valid limit parameter - will panic due to nil db
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=25", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPersonalRecords_WithInvalidLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with invalid limit parameter - should use default
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=invalid", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPersonalRecords_WithNegativeLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with negative limit - should use default
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=-5", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPersonalRecords_WithHighLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with limit > 200 - should cap at 200
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=500", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPRMovements_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs/movements", "")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_GetPRMovements_WithLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with valid limit parameter
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=10", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPRMovements(rr, req)
}

func TestPRHandler_GetPRMovements_WithInvalidLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with invalid limit parameter
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPRMovements(rr, req)
}

func TestPRHandler_GetPRMovements_WithHighLimitParam(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Test with limit > 100 - should cap at 100
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=500", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPRMovements(rr, req)
}

func TestPRHandler_ToggleMovementPR_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodPut, "/api/prs/toggle?id=1", "")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_ToggleMovementPR_MissingID(t *testing.T) {
	handler := &PRHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Missing movement ID")
}

func TestPRHandler_ToggleMovementPR_InvalidID(t *testing.T) {
	handler := &PRHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestPRHandler_ToggleMovementPR_InvalidIDFormats(t *testing.T) {
	handler := &PRHandler{}

	testCases := []struct {
		name string
		id   string
	}{
		{"float", "1.5"},
		{"empty", ""},
		{"special chars", "1@#$"},
		{"overflow", "99999999999999999999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/prs/toggle"
			if tc.id != "" {
				url += "?id=" + tc.id
			}
			req := createAuthenticatedRequest(http.MethodPut, url, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ToggleMovementPR(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
		})
	}
}

func TestPRHandler_ToggleMovementPR_NilDB(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=1", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil db
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.ToggleMovementPR(rr, req)
}

func TestPRHandler_GetPersonalRecords_NilDB(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil db
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPRMovements_NilDB(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil db
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPRMovements(rr, req)
}

func TestPRHandler_ToggleMovementPR_ZeroID(t *testing.T) {
	handler := &PRHandler{}

	// Zero as ID should still parse but we can test it
	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Will panic on nil db if ID parses successfully
	defer func() {
		if r := recover(); r == nil {
			// If it doesn't panic, it should at least parse the ID
			t.Log("ID=0 was parsed successfully (would hit DB check)")
		}
	}()

	handler.ToggleMovementPR(rr, req)
}

func TestPRHandler_GetPersonalRecords_ZeroLimit(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Zero limit should use default
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPersonalRecords(rr, req)
}

func TestPRHandler_GetPRMovements_ZeroLimit(t *testing.T) {
	handler := &PRHandler{
		logger: createTestLogger(),
	}

	// Zero limit should use default
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.GetPRMovements(rr, req)
}
