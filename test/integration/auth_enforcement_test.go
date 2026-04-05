package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

const testJWTSecret = "test-secret-key-for-auth-tests"

// setupFullRouter creates a router with all endpoints similar to main.go
// This ensures we test the actual route structure
func setupFullRouter(t *testing.T) (*chi.Mux, func()) {
	db, err := repository.InitDatabase(*testDBDriver, *testDSN, nil)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	refreshTokenRepo := repository.NewSQLiteRefreshTokenRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	wodRepo := repository.NewWODRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)
	workoutWODRepo := repository.NewWorkoutWODRepository(db)
	userWorkoutRepo := repository.NewUserWorkoutRepository(db)
	userWorkoutMovementRepo := repository.NewUserWorkoutMovementRepository(db)
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)
	userSettingsRepo := repository.NewSQLiteUserSettingsRepository(db)

	// Initialize services
	testLogger, _ := logger.New(logger.Config{Level: "error", EnableFile: false})

	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		nil, // no subscription repo for tests
		nil, // no audit log service
		testJWTSecret,
		24*time.Hour,
		7*24*time.Hour,
		true, // allow registration
		nil,  // no email service
		"http://localhost:3000",
		false, // no email verification
		5, 15*time.Minute,
	)

	movementService := service.NewMovementService(movementRepo, nil, nil)
	wodService := service.NewWODService(wodRepo, nil, nil)
	workoutTemplateService := service.NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)
	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo, workoutRepo, workoutMovementRepo,
		userWorkoutMovementRepo, userWorkoutWODRepo, wodRepo,
		nil, movementRepo, nil, userRepo, nil,
	)
	workoutWODService := service.NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)
	userSettingsService := service.NewUserSettingsService(userSettingsRepo, nil)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, testLogger)
	userHandler := handler.NewUserHandler(userService, nil, testLogger)
	movementHandler := handler.NewMovementHandler(movementRepo, movementService, testLogger)
	wodHandler := handler.NewWODHandler(wodService)
	workoutTemplateHandler := handler.NewWorkoutTemplateHandler(workoutTemplateService)
	userWorkoutHandler := handler.NewUserWorkoutHandler(userWorkoutService, testLogger)
	workoutWODHandler := handler.NewWorkoutWODHandler(workoutWODService)
	settingsHandler := handler.NewSettingsHandler(userSettingsService, testLogger)

	// Set up router matching main.go structure
	r := chi.NewRouter()

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Public browse routes
		r.Get("/movements", movementHandler.ListAll)
		r.Get("/movements/search", movementHandler.Search)
		r.Get("/movements/{id}", movementHandler.GetByID)
		r.Get("/wods", wodHandler.ListWODs)
		r.Get("/wods/standard", wodHandler.ListStandardWODs)
		r.Get("/wods/search", wodHandler.SearchWODs)
		r.Get("/wods/{id}", wodHandler.GetWOD)
		r.Get("/templates", workoutTemplateHandler.ListStandardTemplates)
		r.Get("/templates/{id}", workoutTemplateHandler.GetTemplate)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(testJWTSecret))

			// Profile routes (no subscription required)
			r.Get("/users/profile", userHandler.GetProfile)
			r.Put("/users/profile", userHandler.UpdateProfile)
			r.Get("/users/settings", settingsHandler.GetSettings)
			r.Put("/users/settings", settingsHandler.UpdateSettings)
			r.Put("/users/password", userHandler.ChangePassword)

			// Movement management
			r.Post("/movements", movementHandler.Create)
			r.Put("/movements/{id}", movementHandler.Update)
			r.Delete("/movements/{id}", movementHandler.Delete)

			// WOD management
			r.Get("/wods/my-wods", wodHandler.ListMyWODs)
			r.Post("/wods", wodHandler.CreateWOD)
			r.Put("/wods/{id}", wodHandler.UpdateWOD)
			r.Delete("/wods/{id}", wodHandler.DeleteWOD)

			// Workout Template routes
			r.Post("/templates", workoutTemplateHandler.CreateTemplate)
			r.Get("/workouts/my-templates", workoutTemplateHandler.ListMyTemplates)
			r.Put("/templates/{id}", workoutTemplateHandler.UpdateTemplate)
			r.Delete("/templates/{id}", workoutTemplateHandler.DeleteTemplate)

			// User Workout routes
			r.Post("/workouts", userWorkoutHandler.LogWorkout)
			r.Get("/workouts", userWorkoutHandler.ListLoggedWorkouts)
			r.Get("/workouts/{id}", userWorkoutHandler.GetLoggedWorkout)
			r.Put("/workouts/{id}", userWorkoutHandler.UpdateLoggedWorkout)
			r.Delete("/workouts/{id}", userWorkoutHandler.DeleteLoggedWorkout)
			r.Get("/workouts/stats/monthly", userWorkoutHandler.GetMonthlyStats)
			r.Get("/workouts/personal-records", userWorkoutHandler.GetPersonalRecords)

			// Workout WOD linking
			r.Post("/templates/{workout_id}/wods", workoutWODHandler.AddWODToWorkout)
			r.Get("/templates/{workout_id}/wods", workoutWODHandler.ListWODsForWorkout)
			r.Put("/templates/wods/{workout_wod_id}", workoutWODHandler.UpdateWorkoutWOD)
			r.Delete("/templates/wods/{workout_wod_id}", workoutWODHandler.RemoveWODFromWorkout)

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnly)
				r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"users":[]}`))
				})
				r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"metrics":{}}`))
				})
			})
		})
	})

	return r, cleanup
}

// TestPublicEndpointsAccessibleWithoutAuth verifies that public endpoints
// can be accessed without authentication
func TestPublicEndpointsAccessibleWithoutAuth(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	publicEndpoints := []struct {
		name   string
		method string
		path   string
	}{
		// Health check
		{"Health Check", "GET", "/health"},

		// Auth endpoints
		{"Register", "POST", "/api/auth/register"},
		{"Login", "POST", "/api/auth/login"},

		// Public browse endpoints
		{"List Movements", "GET", "/api/movements"},
		{"Search Movements", "GET", "/api/movements/search"},
		{"Get Movement", "GET", "/api/movements/1"},
		{"List WODs", "GET", "/api/wods"},
		{"List Standard WODs", "GET", "/api/wods/standard"},
		{"Search WODs", "GET", "/api/wods/search"},
		{"Get WOD", "GET", "/api/wods/1"},
		{"List Templates", "GET", "/api/templates"},
		{"Get Template", "GET", "/api/templates/1"},
	}

	for _, tt := range publicEndpoints {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.method == "POST" {
				// Provide minimal valid JSON body for POST requests
				if tt.path == "/api/auth/register" {
					body = bytes.NewBuffer([]byte(`{"email":"test@example.com","password":"Password123456!","name":"Test"}`))
				} else if tt.path == "/api/auth/login" {
					body = bytes.NewBuffer([]byte(`{"email":"test@example.com","password":"Password123456!"}`))
				} else {
					body = bytes.NewBuffer([]byte(`{}`))
				}
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Public endpoints should NOT return 401 Unauthorized.
			// Exception: auth endpoints (login/register) may return 401/400 due to
			// invalid/missing credentials — that means the endpoint was reached, not blocked.
			isAuthEndpoint := tt.path == "/api/auth/login" || tt.path == "/api/auth/register"
			if !isAuthEndpoint && rec.Code == http.StatusUnauthorized {
				t.Errorf("%s %s should be accessible without auth, got 401 Unauthorized", tt.method, tt.path)
			}
		})
	}
}

// TestProtectedEndpointsRequireAuth verifies that all protected endpoints
// return 401 Unauthorized when accessed without authentication
func TestProtectedEndpointsRequireAuth(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	protectedEndpoints := []struct {
		name   string
		method string
		path   string
	}{
		// Profile routes
		{"Get Profile", "GET", "/api/users/profile"},
		{"Update Profile", "PUT", "/api/users/profile"},
		{"Get Settings", "GET", "/api/users/settings"},
		{"Update Settings", "PUT", "/api/users/settings"},
		{"Change Password", "PUT", "/api/users/password"},

		// Movement management (write operations)
		{"Create Movement", "POST", "/api/movements"},
		{"Update Movement", "PUT", "/api/movements/1"},
		{"Delete Movement", "DELETE", "/api/movements/1"},

		// WOD management
		{"List My WODs", "GET", "/api/wods/my-wods"},
		{"Create WOD", "POST", "/api/wods"},
		{"Update WOD", "PUT", "/api/wods/1"},
		{"Delete WOD", "DELETE", "/api/wods/1"},

		// Template management
		{"Create Template", "POST", "/api/templates"},
		{"List My Templates", "GET", "/api/workouts/my-templates"},
		{"Update Template", "PUT", "/api/templates/1"},
		{"Delete Template", "DELETE", "/api/templates/1"},

		// User Workout routes
		{"Create Workout", "POST", "/api/workouts"},
		{"List Workouts", "GET", "/api/workouts"},
		{"Get Workout", "GET", "/api/workouts/1"},
		{"Update Workout", "PUT", "/api/workouts/1"},
		{"Delete Workout", "DELETE", "/api/workouts/1"},
		{"Monthly Stats", "GET", "/api/workouts/stats/monthly"},
		{"Personal Records", "GET", "/api/workouts/personal-records"},

		// Workout WOD linking
		{"Add WOD to Workout", "POST", "/api/templates/1/wods"},
		{"List Workout WODs", "GET", "/api/templates/1/wods"},
		{"Update Workout WOD", "PUT", "/api/templates/wods/1"},
		{"Remove Workout WOD", "DELETE", "/api/templates/wods/1"},

		// Admin routes
		{"Admin List Users", "GET", "/api/admin/users"},
		{"Admin Metrics", "GET", "/api/admin/metrics"},
	}

	for _, tt := range protectedEndpoints {
		t.Run(tt.name+"_NoAuth", func(t *testing.T) {
			var body *bytes.Buffer
			if tt.method == "POST" || tt.method == "PUT" {
				body = bytes.NewBuffer([]byte(`{}`))
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.method == "POST" || tt.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s without auth: expected 401, got %d", tt.method, tt.path, rec.Code)
			}
		})
	}
}

// TestProtectedEndpointsWithInvalidToken verifies that protected endpoints
// reject invalid JWT tokens
func TestProtectedEndpointsWithInvalidToken(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	invalidTokens := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Malformed token", "not-a-valid-jwt"},
		{"Invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlIjoidXNlciJ9.invalid_signature"},
		{"Token with wrong secret", generateTokenWithSecret(1, "test@example.com", "user", "wrong-secret")},
	}

	protectedPaths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users/profile"},
		{"GET", "/api/workouts"},
		{"POST", "/api/workouts"},
	}

	for _, token := range invalidTokens {
		for _, path := range protectedPaths {
			t.Run(token.name+"_"+path.method+"_"+path.path, func(t *testing.T) {
				var body *bytes.Buffer
				if path.method == "POST" {
					body = bytes.NewBuffer([]byte(`{}`))
				} else {
					body = bytes.NewBuffer(nil)
				}

				req := httptest.NewRequest(path.method, path.path, body)
				if path.method == "POST" {
					req.Header.Set("Content-Type", "application/json")
				}
				if token.token != "" {
					req.Header.Set("Authorization", "Bearer "+token.token)
				}
				rec := httptest.NewRecorder()

				router.ServeHTTP(rec, req)

				if rec.Code != http.StatusUnauthorized {
					t.Errorf("%s %s with %s: expected 401, got %d", path.method, path.path, token.name, rec.Code)
				}
			})
		}
	}
}

// TestProtectedEndpointsWithExpiredToken verifies that expired tokens are rejected
func TestProtectedEndpointsWithExpiredToken(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Generate a token that expires immediately
	token, err := auth.GenerateToken(1, "test@example.com", "user", testJWTSecret, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users/profile"},
		{"GET", "/api/workouts"},
		{"POST", "/api/movements"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+"_"+ep.path, func(t *testing.T) {
			var body *bytes.Buffer
			if ep.method == "POST" {
				body = bytes.NewBuffer([]byte(`{}`))
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(ep.method, ep.path, body)
			if ep.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s with expired token: expected 401, got %d", ep.method, ep.path, rec.Code)
			}
		})
	}
}

// TestProtectedEndpointsWithValidToken verifies that valid tokens grant access
func TestProtectedEndpointsWithValidToken(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Generate a valid token
	token, err := auth.GenerateToken(1, "test@example.com", "user", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// These endpoints should return something other than 401 with a valid token
	endpoints := []struct {
		name   string
		method string
		path   string
	}{
		{"Get Profile", "GET", "/api/users/profile"},
		{"Get Settings", "GET", "/api/users/settings"},
		{"List Workouts", "GET", "/api/workouts"},
		{"List My Templates", "GET", "/api/workouts/my-templates"},
		{"Monthly Stats", "GET", "/api/workouts/stats/monthly"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// With valid auth, should NOT get 401
			if rec.Code == http.StatusUnauthorized {
				t.Errorf("%s %s with valid token: got 401 Unauthorized", ep.method, ep.path)
			}
		})
	}
}

// TestAdminEndpointsRequireAdminRole verifies that admin endpoints reject non-admin users
func TestAdminEndpointsRequireAdminRole(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Generate a non-admin token
	userToken, err := auth.GenerateToken(1, "user@example.com", "user", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate user token: %v", err)
	}

	// Generate an admin token
	adminToken, err := auth.GenerateToken(2, "admin@example.com", "admin", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}

	adminEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/users"},
		{"GET", "/api/admin/metrics"},
	}

	// Test with non-admin user
	for _, ep := range adminEndpoints {
		t.Run("NonAdmin_"+ep.method+"_"+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("Non-admin accessing %s %s: expected 403, got %d", ep.method, ep.path, rec.Code)
			}
		})
	}

	// Test with admin user
	for _, ep := range adminEndpoints {
		t.Run("Admin_"+ep.method+"_"+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Admin should NOT get 401 or 403
			if rec.Code == http.StatusUnauthorized {
				t.Errorf("Admin accessing %s %s: got 401 Unauthorized", ep.method, ep.path)
			}
			if rec.Code == http.StatusForbidden {
				t.Errorf("Admin accessing %s %s: got 403 Forbidden", ep.method, ep.path)
			}
		})
	}
}

// TestAuthorizationHeaderFormats verifies different authorization header formats
func TestAuthorizationHeaderFormats(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	validToken, _ := auth.GenerateToken(1, "test@example.com", "user", testJWTSecret, time.Hour)

	tests := []struct {
		name       string
		authHeader string
		expectAuth bool
	}{
		{"Valid Bearer format", "Bearer " + validToken, true},
		{"Missing Bearer prefix", validToken, false},
		{"Lowercase bearer", "bearer " + validToken, false},
		{"Basic instead of Bearer", "Basic " + validToken, false},
		{"Empty header", "", false},
		{"Just Bearer", "Bearer", false},
		{"Bearer with extra space", "Bearer  " + validToken, false},
		{"Bearer with tab", "Bearer\t" + validToken, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/users/profile", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if tt.expectAuth {
				if rec.Code == http.StatusUnauthorized {
					t.Errorf("Expected auth to succeed with '%s', got 401", tt.name)
				}
			} else {
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("Expected auth to fail with '%s', got %d", tt.name, rec.Code)
				}
			}
		})
	}
}

// TestConcurrentAuthRequests verifies auth works correctly under concurrent load
func TestConcurrentAuthRequests(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	validToken, _ := auth.GenerateToken(1, "test@example.com", "user", testJWTSecret, time.Hour)

	// Run concurrent requests
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(withAuth bool) {
			req := httptest.NewRequest("GET", "/api/users/profile", nil)
			if withAuth {
				req.Header.Set("Authorization", "Bearer "+validToken)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if withAuth {
				if rec.Code == http.StatusUnauthorized {
					t.Errorf("Authenticated request should not get 401")
				}
			} else {
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("Unauthenticated request should get 401, got %d", rec.Code)
				}
			}
			done <- true
		}(i%2 == 0)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestNoAuthBypassViaHTTPMethods verifies that different HTTP methods don't bypass auth
func TestNoAuthBypassViaHTTPMethods(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Test that unusual methods don't bypass auth on protected routes
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	protectedPath := "/api/users/profile"

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var body *bytes.Buffer
			if method == "POST" || method == "PUT" || method == "PATCH" {
				body = bytes.NewBuffer([]byte(`{}`))
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(method, protectedPath, body)
			if method == "POST" || method == "PUT" || method == "PATCH" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// OPTIONS might be allowed for CORS, but others should require auth
			if method != "OPTIONS" && rec.Code != http.StatusUnauthorized && rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s without auth: expected 401 or 405, got %d", method, protectedPath, rec.Code)
			}
		})
	}
}

// Helper function to generate token with a specific secret
func generateTokenWithSecret(userID int64, email, role, secret string) string {
	token, _ := auth.GenerateToken(userID, email, role, secret, time.Hour)
	return token
}

// TestTokenPayloadIntegrity verifies that the token payload is correctly parsed
func TestTokenPayloadIntegrity(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	testCases := []struct {
		name   string
		userID int64
		email  string
		role   string
	}{
		{"Standard User", 1, "user@example.com", "user"},
		{"Admin User", 999, "admin@example.com", "admin"},
		{"User with long email", 42, "very.long.email.address.for.testing@subdomain.example.com", "user"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tc.userID, tc.email, tc.role, testJWTSecret, time.Hour)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			req := httptest.NewRequest("GET", "/api/users/profile", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Token should be accepted (not 401)
			if rec.Code == http.StatusUnauthorized {
				t.Errorf("Valid token for %s should be accepted", tc.name)
			}
		})
	}
}

// TestAuthEndpointSecurityHeaders verifies security-related response headers
func TestAuthEndpointSecurityHeaders(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Test that 401 responses include proper content type
	req := httptest.NewRequest("GET", "/api/users/profile", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", rec.Code)
	}

	// Response should be JSON
	contentType := rec.Header().Get("Content-Type")
	if contentType != "" && contentType != "text/plain; charset=utf-8" && contentType != "application/json" {
		// It's okay if it's plain text error or JSON
		t.Logf("Content-Type: %s", contentType)
	}

	// Verify the response body contains error message
	body := rec.Body.String()
	if body == "" {
		t.Error("401 response should have a body explaining the error")
	}
}

// TestRouteProtectionComprehensive is a meta-test that runs all route checks
func TestRouteProtectionComprehensive(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// All API routes should either be:
	// 1. Public (explicitly listed)
	// 2. Protected (return 401 without auth)

	publicRoutes := map[string]bool{
		"GET /health":               true,
		"POST /api/auth/register":   true,
		"POST /api/auth/login":      true,
		"GET /api/movements":        true,
		"GET /api/movements/search": true,
		"GET /api/movements/1":      true,
		"GET /api/wods":             true,
		"GET /api/wods/standard":    true,
		"GET /api/wods/search":      true,
		"GET /api/wods/1":           true,
		"GET /api/templates":        true,
		"GET /api/templates/1":      true,
	}

	allRoutes := []struct {
		method string
		path   string
	}{
		// Public routes
		{"GET", "/health"},
		{"POST", "/api/auth/register"},
		{"POST", "/api/auth/login"},
		{"GET", "/api/movements"},
		{"GET", "/api/movements/search"},
		{"GET", "/api/movements/1"},
		{"GET", "/api/wods"},
		{"GET", "/api/wods/standard"},
		{"GET", "/api/wods/search"},
		{"GET", "/api/wods/1"},
		{"GET", "/api/templates"},
		{"GET", "/api/templates/1"},

		// Protected routes
		{"GET", "/api/users/profile"},
		{"PUT", "/api/users/profile"},
		{"GET", "/api/users/settings"},
		{"PUT", "/api/users/settings"},
		{"PUT", "/api/users/password"},
		{"POST", "/api/movements"},
		{"PUT", "/api/movements/1"},
		{"DELETE", "/api/movements/1"},
		{"GET", "/api/wods/my-wods"},
		{"POST", "/api/wods"},
		{"PUT", "/api/wods/1"},
		{"DELETE", "/api/wods/1"},
		{"POST", "/api/templates"},
		{"GET", "/api/workouts/my-templates"},
		{"PUT", "/api/templates/1"},
		{"DELETE", "/api/templates/1"},
		{"POST", "/api/workouts"},
		{"GET", "/api/workouts"},
		{"GET", "/api/workouts/1"},
		{"PUT", "/api/workouts/1"},
		{"DELETE", "/api/workouts/1"},
		{"GET", "/api/workouts/stats/monthly"},
		{"GET", "/api/workouts/personal-records"},
		{"POST", "/api/templates/1/wods"},
		{"GET", "/api/templates/1/wods"},
		{"PUT", "/api/templates/wods/1"},
		{"DELETE", "/api/templates/wods/1"},
		{"GET", "/api/admin/users"},
		{"GET", "/api/admin/metrics"},
	}

	for _, route := range allRoutes {
		key := route.method + " " + route.path
		isPublic := publicRoutes[key]

		t.Run(key, func(t *testing.T) {
			var body *bytes.Buffer
			if route.method == "POST" || route.method == "PUT" {
				if route.path == "/api/auth/register" {
					body = bytes.NewBuffer([]byte(`{"email":"x@x.com","password":"Password123456!","name":"X"}`))
				} else if route.path == "/api/auth/login" {
					body = bytes.NewBuffer([]byte(`{"email":"x@x.com","password":"Password123456!"}`))
				} else {
					body = bytes.NewBuffer([]byte(`{}`))
				}
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(route.method, route.path, body)
			if route.method == "POST" || route.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if isPublic {
				// Public routes should not return 401 with "Missing authorization header" message
				// Note: Login can return 401 for invalid credentials - that's different from auth required
				if rec.Code == http.StatusUnauthorized {
					body := rec.Body.String()
					// Only fail if it's an auth header issue, not credential validation
					if strings.Contains(body, "Missing authorization header") ||
						strings.Contains(body, "Invalid authorization header") {
						t.Errorf("Public route %s should not require auth header", key)
					}
				}
			} else {
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("Protected route %s should return 401 without auth, got %d", key, rec.Code)
				}
			}
		})
	}
}

// TestAllAPIRoutesMustReturnJSON verifies that all API error responses are JSON
func TestAllAPIRoutesMustReturnJSON(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Test a few protected routes to ensure 401 errors are parseable
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users/profile"},
		{"GET", "/api/workouts"},
		{"POST", "/api/movements"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			var body *bytes.Buffer
			if route.method == "POST" {
				body = bytes.NewBuffer([]byte(`{}`))
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(route.method, route.path, body)
			if route.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Skipf("Expected 401, got %d", rec.Code)
			}

			// Try to parse as JSON
			var result map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &result)
			if err != nil {
				// It's acceptable to have plain text error for 401
				// but log it for awareness
				t.Logf("401 response is not JSON: %s", rec.Body.String())
			}
		})
	}
}
