package integration

import (
	// standard library
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	// internal / external
	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/internal/testhelpers"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

var (
	testDBDriver = flag.String("db", "sqlite3", "database driver for integration tests (sqlite3|postgres|mysql)")
	testDSN      = flag.String("dsn", ":memory:", "database DSN for integration tests")
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Allow environment variables to override flags for CI convenience
	if envDriver := os.Getenv("DB_DRIVER"); envDriver != "" {
		*testDBDriver = envDriver
	}
	if envDSN := os.Getenv("DB_DSN"); envDSN != "" {
		*testDSN = envDSN
	}

	// Setup a temporary DB for postgres/mysql to isolate CI runs
	var teardown func()
	if *testDBDriver == "postgres" || *testDBDriver == "mysql" {
		td, tdFn, err := testhelpers.SetupTempDB(*testDBDriver, *testDSN)
		if err == nil {
			*testDSN = td
			teardown = tdFn
			// export env for other helpers
			_ = os.Setenv("DB_DRIVER", *testDBDriver)
			_ = os.Setenv("DB_DSN", *testDSN)
		}
	}

	code := m.Run()

	if teardown != nil {
		teardown()
	}

	os.Exit(code)
}

// Test helper to set up test router with dependencies
func setupTestRouter(t *testing.T) (*chi.Mux, *repository.SQLiteUserRepository, *sql.DB, int64, error) {
	// Initialize using the configured test DB driver and DSN (defaults to sqlite in-memory)
	db, err := repository.InitDatabase(*testDBDriver, *testDSN)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)

	// Create a minimal workout template so tests can log a workout referencing it
	// Schema may differ depending on migrations. Inspect table columns and insert accordingly.
	cols := map[string]bool{}
	rows, err := db.Query("PRAGMA table_info(workouts)")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			_ = rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
			cols[name] = true
		}
	}

	var templateID int64
	if cols["user_id"] {
		// Old schema: workouts are user-specific and require a user_id. Create a user and insert accordingly.
		u := &domain.User{
			Email:        "template-owner@example.com",
			PasswordHash: "",
			Name:         "Template Owner",
			Role:         "user",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := userRepo.Create(u); err != nil {
			return nil, nil, nil, 0, err
		}

		// If the migrated schema added a `name` column (templates), insert into `name` to keep WorkoutRepository happy.
		var res sql.Result
		if cols["name"] {
			// include workout_date because older schema columns may still be present and NOT NULL
			insertQuery := `INSERT INTO workouts (name, workout_date, workout_type, created_at, updated_at, user_id) VALUES (?, ?, ?, ?, ?, ?)`
			res, err = db.Exec(insertQuery, "Test Template", time.Now().Format("2006-01-02"), "strength", time.Now(), time.Now(), u.ID)
		} else {
			insertQuery := `INSERT INTO workouts (user_id, workout_date, workout_type, workout_name, notes, total_time, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
			res, err = db.Exec(insertQuery, u.ID, time.Now().Format("2006-01-02"), "strength", "Test Template", nil, nil, time.Now(), time.Now())
		}
		if err != nil {
			return nil, nil, nil, 0, err
		}
		id, _ := res.LastInsertId()
		templateID = id
	} else {
		// New schema: workouts are templates
		defaultWorkout := &domain.Workout{
			Name: "Test Template",
		}
		if err := workoutRepo.Create(defaultWorkout); err != nil {
			return nil, nil, nil, 0, err
		}
		templateID = defaultWorkout.ID
	}

	// Add a movement to the template if strength movements exist (movement_id 1 seeded)
	wm := &domain.WorkoutMovement{
		WorkoutID:  templateID,
		MovementID: 1,
		OrderIndex: 1,
	}
	if err := workoutMovementRepo.Create(wm); err != nil {
		return nil, nil, nil, 0, err
	}

	refreshTokenRepo := repository.NewSQLiteRefreshTokenRepository(db)

	// Initialize services
	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		"test-secret-key",
		24*time.Hour,
		7*24*time.Hour,
		true, // allow registration
		nil,  // no email service for tests
		"http://localhost:3000",
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService)
	// Create user workout handler
	userWorkoutService := service.NewUserWorkoutService(
		repository.NewUserWorkoutRepository(db),
		workoutRepo,
		workoutMovementRepo,
	)
	userWorkoutHandler := handler.NewUserWorkoutHandler(userWorkoutService)

	// Create workout service for PR endpoints
	movementRepo := repository.NewMovementRepository(db)
	workoutWODRepo := repository.NewWorkoutWODRepository(db)
	workoutService := service.NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	// Set up router
	r := chi.NewRouter()

	// Auth routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth("test-secret-key"))
		r.Post("/api/workouts", userWorkoutHandler.LogWorkout)
		r.Get("/api/workouts", userWorkoutHandler.ListLoggedWorkouts)
		r.Get("/api/workouts/{id}", userWorkoutHandler.GetLoggedWorkout)
		r.Put("/api/workouts/{id}", userWorkoutHandler.UpdateLoggedWorkout)
		r.Delete("/api/workouts/{id}", userWorkoutHandler.DeleteLoggedWorkout)

		// PR endpoints (used by integration tests)
		r.Get("/api/workouts/prs", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := middleware.GetUserID(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			prs, err := workoutService.GetPersonalRecords(userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"prs": prs})
		})

		r.Get("/api/workouts/pr-movements", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := middleware.GetUserID(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			movs, err := workoutService.GetPRMovements(userID, 10)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"movements": movs})
		})
	})

	return r, userRepo, db, templateID, nil
}

// Test Registration and Login Flow
func TestRegistrationAndLogin(t *testing.T) {
	router, _, _, _, err := setupTestRouter(t)
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
	router, _, _, templateID, err := setupTestRouter(t)
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
			"workout_id":   templateID,
			"workout_date": time.Now().Format("2006-01-02"),
			"workout_type": "strength",
			"notes":        "Test Workout notes",
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
	router, _, _, _, err := setupTestRouter(t)
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
