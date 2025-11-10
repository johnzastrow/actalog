package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// Test helper to set up test router with dependencies
func setupTestRouter(t *testing.T) (*chi.Mux, *repository.SQLiteUserRepository, error) {
	// Use in-memory SQLite for testing
	db, err := repository.InitDatabase("sqlite3", ":memory:")
	if err != nil {
		return nil, nil, err
	}

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	movementRepo := repository.NewSQLiteMovementRepository(db)
	workoutRepo := repository.NewSQLiteWorkoutRepository(db)
	workoutMovementRepo := repository.NewSQLiteWorkoutMovementRepository(db)

	// Initialize services
	userService := service.NewUserService(
		userRepo,
		"test-secret-key",
		24*time.Hour,
		true, // allow registration
		nil,  // no email service for tests
		"http://localhost:3000",
	)

	workoutService := service.NewWorkoutService(
		workoutRepo,
		workoutMovementRepo,
		movementRepo,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService)
	workoutHandler := handler.NewWorkoutHandler(workoutRepo, workoutMovementRepo, workoutService)

	// Set up router
	r := chi.NewRouter()

	// Auth routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth("test-secret-key"))
		r.Post("/api/workouts", workoutHandler.Create)
		r.Get("/api/workouts", workoutHandler.ListByUser)
		r.Get("/api/workouts/prs", workoutHandler.GetPersonalRecords)
		r.Get("/api/workouts/pr-movements", workoutHandler.GetPRMovements)
		r.Post("/api/workouts/movements/{id}/toggle-pr", workoutHandler.TogglePRFlag)
	})

	return r, userRepo, nil
}

// Test Registration and Login Flow
func TestRegistrationAndLogin(t *testing.T) {
	router, _, err := setupTestRouter(t)
	if err != nil {
		t.Fatalf("Failed to setup router: %v", err)
	}

	// Test Registration
	t.Run("Register User", func(t *testing.T) {
		body := map[string]string{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
			t.Logf("Response: %s", w.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["token"] == nil {
			t.Error("Expected token in response")
		}

		if response["user"] == nil {
			t.Error("Expected user in response")
		}
	})

	// Test Login
	t.Run("Login User", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "Password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["token"] == nil {
			t.Error("Expected token in response")
		}
	})

	// Test Login with Invalid Credentials
	t.Run("Login with Invalid Password", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "WrongPassword",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 401 or 500, got %d", w.Code)
		}
	})
}

// Test Workout Creation and PR Detection
func TestWorkoutAndPRFlow(t *testing.T) {
	router, _, err := setupTestRouter(t)
	if err != nil {
		t.Fatalf("Failed to setup router: %v", err)
	}

	// Register and login to get token
	var token string
	t.Run("Setup: Register and Login", func(t *testing.T) {
		// Register
		regBody := map[string]string{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123",
		}
		jsonBody, _ := json.Marshal(regBody)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		token = response["token"].(string)
	})

	// Test Creating Workout
	t.Run("Create Workout", func(t *testing.T) {
		weight := 135.0
		reps := 5

		body := map[string]interface{}{
			"workout_date": time.Now().Format(time.RFC3339),
			"workout_type": "strength",
			"workout_name": "Test Workout",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"weight":      weight,
					"reps":        reps,
					"sets":        3,
					"is_rx":       true,
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/workouts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
			t.Logf("Response: %s", w.Body.String())
		}
	})

	// Test Listing Workouts
	t.Run("List Workouts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/workouts", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["workouts"] == nil {
			t.Error("Expected workouts in response")
		}
	})

	// Test Getting Personal Records
	t.Run("Get Personal Records", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/workouts/prs", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

// Test Authentication Required
func TestAuthenticationRequired(t *testing.T) {
	router, _, err := setupTestRouter(t)
	if err != nil {
		t.Fatalf("Failed to setup router: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"List Workouts", "GET", "/api/workouts"},
		{"Get PRs", "GET", "/api/workouts/prs"},
		{"Get PR Movements", "GET", "/api/workouts/pr-movements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}
