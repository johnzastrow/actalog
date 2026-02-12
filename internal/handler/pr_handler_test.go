package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
)

func TestNewPRHandler(t *testing.T) {
	handler := NewPRHandler(nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewPRHandler() should not return nil")
	}
}

// createTestPRHandler creates a PR handler with test database and sample data
func createTestPRHandler(t *testing.T) (*PRHandler, func()) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "prtest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		cleanup()
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a movement
	_, err = db.Exec(`
		INSERT INTO movements (name, type, description, created_at, updated_at)
		VALUES ('Back Squat', 'weightlifting', 'Barbell back squat', ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create movement: %v", err)
	}

	// Create a workout
	_, err = db.Exec(`
		INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
		VALUES ('Test Workout', 'Test notes', 1, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Create a user workout
	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_id, workout_date, created_at, updated_at)
		VALUES (1, 1, ?, ?, ?)
	`, now.Format("2006-01-02"), now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Create a workout movement with PR
	_, err = db.Exec(`
		INSERT INTO workout_movements (workout_id, movement_id, sets, reps, weight, is_pr, order_index, created_at, updated_at)
		VALUES (1, 1, 5, 5, 315.0, 1, 0, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout movement: %v", err)
	}

	// Create a WOD
	_, err = db.Exec(`
		INSERT INTO wods (name, type, score_type, description, created_by, created_at, updated_at)
		VALUES ('Fran', 'For Time', 'time', '21-15-9 Thrusters and Pull-ups', 1, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create a workout WOD with PR
	_, err = db.Exec(`
		INSERT INTO workout_wods (workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at)
		VALUES (1, 1, '5:30', 'rx', 1, 0, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout WOD: %v", err)
	}

	handler := NewPRHandler(db, createTestLogger())
	return handler, cleanup
}

// Removed struct field assignment tests:
// - TestPersonalRecord_Struct, TestPersonalRecord_NilFields
// - TestMovementPRSummary_Struct, TestMovementPRSummary_NilFields
// These tests verified Go struct assignment works, not business logic.

func TestPRHandler_GetPersonalRecords_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs", "")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPRHandler_GetPRMovements_Unauthorized(t *testing.T) {
	handler := &PRHandler{}

	req := createTestRequest(http.MethodGet, "/api/prs/movements", "")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
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

// Success path tests with real database

func TestPRHandler_GetPersonalRecords_Success(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "prs") {
		t.Error("Response should contain 'prs' field")
	}
	// Should have both movement and WOD PRs
	if !strings.Contains(body, "Back Squat") {
		t.Error("Response should contain the movement PR 'Back Squat'")
	}
	if !strings.Contains(body, "Fran") {
		t.Error("Response should contain the WOD PR 'Fran'")
	}
	// Check for calculated 1RM
	if !strings.Contains(body, "calculated_1rm") {
		t.Error("Response should contain 'calculated_1rm' for weight-based PRs")
	}
}

func TestPRHandler_GetPersonalRecords_WithLimit(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=1", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPRHandler_GetPersonalRecords_WithHighLimit(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Request with limit over max (200) - should be capped
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=500", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPRHandler_GetPersonalRecords_WithInvalidLimit(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Request with invalid limit (should use default)
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs?limit=abc", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPRHandler_GetPRMovements_Success(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "movements") {
		t.Error("Response should contain 'movements' field")
	}
	if !strings.Contains(body, "Back Squat") {
		t.Error("Response should contain 'Back Squat' movement")
	}
	if !strings.Contains(body, "pr_count") {
		t.Error("Response should contain 'pr_count' field")
	}
	if !strings.Contains(body, "best_1rm") {
		t.Error("Response should contain 'best_1rm' field for weightlifting movements")
	}
}

func TestPRHandler_GetPRMovements_WithLimit(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=5", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPRHandler_GetPRMovements_WithHighLimit(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Request with limit over max (100) - should be capped
	req := createAuthenticatedRequest(http.MethodGet, "/api/prs/movements?limit=500", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPRMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPRHandler_ToggleMovementPR_Success(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Toggle the PR for workout_movement id=1
	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=1", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "is_pr") {
		t.Error("Response should contain 'is_pr' field")
	}
	if !strings.Contains(body, "message") {
		t.Error("Response should contain 'message' field")
	}
}

func TestPRHandler_ToggleMovementPR_NotFound(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Try to toggle a non-existent movement
	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=9999", "", 1, "prtest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "not found")
}

func TestPRHandler_ToggleMovementPR_WrongUser(t *testing.T) {
	handler, cleanup := createTestPRHandler(t)
	defer cleanup()

	// Try to toggle with a different user (user_id=2 instead of 1)
	req := createAuthenticatedRequest(http.MethodPut, "/api/prs/toggle?id=1", "", 2, "other@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ToggleMovementPR(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "not found")
}

func TestPRHandler_GetPersonalRecords_EmptyResult(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create user but no PRs
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "empty@example.com",
		PasswordHash: "hashedpassword",
		Role:         "athlete",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	handler := NewPRHandler(db, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/prs", "", 1, "empty@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetPersonalRecords(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Should return empty list
	body := rr.Body.String()
	if !strings.Contains(body, "prs") {
		t.Error("Response should contain 'prs' field")
	}
}
