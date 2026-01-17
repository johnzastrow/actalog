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

// createTestPerformanceHandler creates a performance handler with test database and sample data
func createTestPerformanceHandler(t *testing.T) (*PerformanceHandler, func()) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "perftest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		cleanup()
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create repositories
	movementRepo := repository.NewMovementRepository(db)
	wodRepo := repository.NewWODRepository(db)
	userWorkoutMovementRepo := repository.NewUserWorkoutMovementRepository(db)
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)

	// Create test movement
	movement := &domain.Movement{
		Name:        "Back Squat",
		Type:        "weightlifting",
		Description: "Barbell back squat",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := movementRepo.Create(movement); err != nil {
		cleanup()
		t.Fatalf("Failed to create movement: %v", err)
	}

	// Create test WOD
	createdBy := int64(1)
	wod := &domain.WOD{
		Name:        "Fran",
		Type:        "For Time",
		ScoreType:   "time",
		Description: "21-15-9 Thrusters and Pull-ups",
		CreatedBy:   &createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := wodRepo.Create(wod); err != nil {
		cleanup()
		t.Fatalf("Failed to create WOD: %v", err)
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

	// Create workout movement performance
	_, err = db.Exec(`
		INSERT INTO workout_movements (workout_id, movement_id, sets, reps, weight, is_pr, order_index, created_at, updated_at)
		VALUES (1, 1, 5, 5, 315.0, 1, 0, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout movement: %v", err)
	}

	// Create workout WOD performance
	_, err = db.Exec(`
		INSERT INTO workout_wods (workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at)
		VALUES (1, 1, '5:30', 'rx', 1, 0, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout WOD: %v", err)
	}

	handler := NewPerformanceHandler(movementRepo, wodRepo, userWorkoutMovementRepo, userWorkoutWODRepo, createTestLogger())
	return handler, cleanup
}

func TestPerformanceHandler_UnifiedSearch_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_UnifiedSearch_MissingQuery(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestPerformanceHandler_GetMovementPerformance_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/movements/1", "")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_GetMovementPerformance_InvalidID(t *testing.T) {
	handler := &PerformanceHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/movements/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestPerformanceHandler_GetWODPerformance_Unauthorized(t *testing.T) {
	handler := &PerformanceHandler{}

	req := createTestRequest(http.MethodGet, "/api/performance/wods/1", "")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestPerformanceHandler_GetWODPerformance_InvalidID(t *testing.T) {
	handler := &PerformanceHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestNewPerformanceHandler(t *testing.T) {
	handler := NewPerformanceHandler(nil, nil, nil, nil, nil)
	if handler == nil {
		t.Error("NewPerformanceHandler should return a non-nil handler")
	}
}

// Success path tests with real database

func TestPerformanceHandler_UnifiedSearch_Success(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search?q=squat", "", 1, "perftest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "results") {
		t.Error("Response should contain 'results' field")
	}
	if !strings.Contains(body, "Back Squat") {
		t.Error("Response should contain the movement 'Back Squat'")
	}
}

func TestPerformanceHandler_UnifiedSearch_WithLimit(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search?q=squat&limit=5", "", 1, "perftest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPerformanceHandler_UnifiedSearch_InvalidLimit(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	// Invalid limit should use default
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search?q=squat&limit=abc", "", 1, "perftest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestPerformanceHandler_UnifiedSearch_NoResults(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/search?q=nonexistent", "", 1, "perftest@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnifiedSearch(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Should return empty results
	body := rr.Body.String()
	if !strings.Contains(body, "results") {
		t.Error("Response should contain 'results' field")
	}
}

func TestPerformanceHandler_GetMovementPerformance_Success(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/movements/1", "", 1, "perftest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "performances") {
		t.Error("Response should contain 'performances' field")
	}
	if !strings.Contains(body, "count") {
		t.Error("Response should contain 'count' field")
	}
}

func TestPerformanceHandler_GetMovementPerformance_EmptyResult(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	// Non-existent movement returns empty performances (not 404)
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/movements/9999", "", 1, "perftest@example.com", "user")
	req = addChiURLParam(req, "id", "9999")
	rr := httptest.NewRecorder()

	handler.GetMovementPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	body := rr.Body.String()
	if !strings.Contains(body, `"count":0`) {
		t.Error("Response should contain count of 0 for non-existent movement")
	}
}

func TestPerformanceHandler_GetWODPerformance_Success(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/wods/1", "", 1, "perftest@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "count") {
		t.Error("Response should contain 'count' field")
	}
}

func TestPerformanceHandler_GetWODPerformance_EmptyResult(t *testing.T) {
	handler, cleanup := createTestPerformanceHandler(t)
	defer cleanup()

	// Non-existent WOD returns empty performances (not 404)
	req := createAuthenticatedRequest(http.MethodGet, "/api/performance/wods/9999", "", 1, "perftest@example.com", "user")
	req = addChiURLParam(req, "id", "9999")
	rr := httptest.NewRecorder()

	handler.GetWODPerformance(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	body := rr.Body.String()
	if !strings.Contains(body, `"count":0`) {
		t.Error("Response should contain count of 0 for non-existent WOD")
	}
}
