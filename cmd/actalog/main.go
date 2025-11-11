// Package main is the entry point for ActaLog application
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/johnzastrow/actalog/configs"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/email"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/version"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"  // MySQL/MariaDB
	_ "github.com/lib/pq"                // PostgreSQL
	_ "github.com/mattn/go-sqlite3"     // SQLite
)

func main() {
	// Print version information
	fmt.Println(version.String())
	fmt.Println("Starting ActaLog server...")

	// Load .env file (ignore error if file doesn't exist)
	// In production, you should use actual environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables or defaults")
	}

	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:      cfg.Logging.Level,
		EnableFile: cfg.Logging.EnableFile,
		FilePath:   cfg.Logging.FilePath,
		MaxSizeMB:  cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	// Log configuration (without sensitive data)
	appLogger.Info("Environment: %s", cfg.App.Environment)
	appLogger.Info("Log Level: %s", cfg.Logging.Level)
	if cfg.Logging.EnableFile {
		appLogger.Info("File logging: enabled")
	} else {
		appLogger.Info("File logging: disabled (stdout only)")
	}
	appLogger.Info("Database Driver: %s", cfg.Database.Driver)
	appLogger.Info("Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.Info("Allow Registration: %t", cfg.App.AllowRegistration)

	// Build database connection string
	dsn := repository.BuildDSN(
		cfg.Database.Driver,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	// Initialize database
	db, err := repository.InitDatabase(cfg.Database.Driver, dsn)
	if err != nil {
		appLogger.Fatal("Failed to initialize database: %v", err)
	}
	defer db.Close()
	appLogger.Info("Database initialized successfully")

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	refreshTokenRepo := repository.NewSQLiteRefreshTokenRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)
	userWorkoutRepo := repository.NewUserWorkoutRepository(db)
	wodRepo := repository.NewWODRepository(db)
	workoutWODRepo := repository.NewWorkoutWODRepository(db)

	// Initialize email service
	var emailService *email.Service
	if cfg.Email.Enabled && cfg.Email.SMTPHost != "" {
		// Create a standard logger that writes to our custom logger
		stdLogger := log.New(appLogger.Writer(), "", 0)

		emailService = email.NewService(email.Config{
			SMTPHost:     cfg.Email.SMTPHost,
			SMTPPort:     cfg.Email.SMTPPort,
			SMTPUser:     cfg.Email.SMTPUser,
			SMTPPassword: cfg.Email.SMTPPassword,
			FromAddress:  cfg.Email.FromAddress,
			FromName:     cfg.Email.FromName,
		}, stdLogger)
		appLogger.Info("Email service: enabled (SMTP: %s:%d)", cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	} else {
		appLogger.Info("Email service: disabled (password reset emails will not be sent)")
	}

	// Determine app URL for password reset links
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		if cfg.App.Environment == "production" {
			appURL = "https://actalog.example.com" // Replace with your production URL
		} else {
			appURL = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		}
	}

	// Initialize services
	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		cfg.JWT.SecretKey,
		cfg.JWT.ExpirationTime,
		cfg.JWT.RefreshTokenDuration,
		cfg.App.AllowRegistration,
		emailService,
		appURL,
	)

	// workoutService := service.NewWorkoutService(
	// 	workoutRepo,
	// 	workoutMovementRepo,
	// 	movementRepo,
	// 	workoutWODRepo,
	// ) // Temporarily disabled until template handler is created

	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		workoutMovementRepo,
	)

	wodService := service.NewWODService(wodRepo)

	workoutWODService := service.NewWorkoutWODService(
		workoutWODRepo,
		workoutRepo,
		wodRepo,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService)
	userHandler := handler.NewUserHandler(userService)
	movementHandler := handler.NewMovementHandler(movementRepo)
	// workoutHandler := handler.NewWorkoutHandler(workoutRepo, workoutMovementRepo, workoutService) // DEPRECATED for v0.4.0
	userWorkoutHandler := handler.NewUserWorkoutHandler(userWorkoutService)
	wodHandler := handler.NewWODHandler(wodService)
	workoutWODHandler := handler.NewWorkoutWODHandler(workoutWODService)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestLogger(appLogger))
	r.Use(middleware.CORS(cfg.App.CORSOrigins))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"%s"}`, version.Version())
	})

	// Version endpoint
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"version":"%s","app":"%s"}`, version.Version(), cfg.App.Name)
	})

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Welcome to ActaLog API","version":"%s"}`, version.Version())
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (public)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.Post("/auth/reset-password", authHandler.ResetPassword)
		r.Get("/auth/verify-email", authHandler.VerifyEmail)
		r.Post("/auth/resend-verification", authHandler.ResendVerification)
		r.Post("/auth/refresh", authHandler.RefreshToken)
		r.Post("/auth/revoke", authHandler.RevokeToken)

		// Movement routes (public for browsing)
		r.Get("/movements", movementHandler.ListStandard)
		r.Get("/movements/search", movementHandler.Search)
		r.Get("/movements/{id}", movementHandler.GetByID)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWT.SecretKey))

			// Movement management (authenticated)
			r.Post("/movements", movementHandler.Create)

			// User profile routes (authenticated)
			r.Get("/users/profile", userHandler.GetProfile)
			r.Put("/users/profile", userHandler.UpdateProfile)

			// Workout routes (authenticated) - DEPRECATED: v0.3.x compatibility - temporarily disabled for v0.4.0 migration
			// r.Post("/workouts", workoutHandler.Create)
			// r.Get("/workouts", workoutHandler.ListByUser)
			// r.Get("/workouts/{id}", workoutHandler.GetByID)
			// r.Put("/workouts/{id}", workoutHandler.Update)
			// r.Delete("/workouts/{id}", workoutHandler.Delete)

			// PR tracking routes (authenticated) - temporarily disabled for v0.4.0 migration
			// r.Get("/workouts/prs", workoutHandler.GetPersonalRecords)
			// r.Get("/workouts/pr-movements", workoutHandler.GetPRMovements)
			// r.Post("/workouts/movements/{id}/toggle-pr", workoutHandler.TogglePRFlag)

			// Progress tracking (authenticated) - temporarily disabled for v0.4.0 migration
			// r.Get("/progress/movements/{movement_id}", workoutHandler.GetProgressByMovement)

			// === v0.4.0 NEW ROUTES ===

			// Workout Templates (authenticated) - temporarily disabled until template handler is created
			// r.Get("/templates", workoutHandler.ListByUser)       // List user's templates
			// r.Post("/templates", workoutHandler.Create)          // Create template
			// r.Get("/templates/{id}", workoutHandler.GetByID)     // Get template details
			// r.Put("/templates/{id}", workoutHandler.Update)      // Update template
			// r.Delete("/templates/{id}", workoutHandler.Delete)   // Delete template

			// WOD Management (authenticated)
			r.Get("/wods", wodHandler.ListWODs)              // List all WODs (standard + user's custom)
			r.Post("/wods", wodHandler.CreateWOD)            // Create custom WOD
			r.Get("/wods/search", wodHandler.SearchWODs)     // Search WODs
			r.Get("/wods/{id}", wodHandler.GetWOD)           // Get WOD details
			r.Put("/wods/{id}", wodHandler.UpdateWOD)        // Update custom WOD
			r.Delete("/wods/{id}", wodHandler.DeleteWOD)     // Delete custom WOD

			// Link WODs to Templates (authenticated)
			r.Post("/templates/{id}/wods", workoutWODHandler.AddWODToWorkout)              // Add WOD to template
			r.Get("/templates/{id}/wods", workoutWODHandler.ListWODsForWorkout)            // List WODs in template
			r.Put("/templates/{id}/wods/{wod_id}", workoutWODHandler.UpdateWorkoutWOD)     // Update WOD in template
			r.Delete("/templates/{id}/wods/{wod_id}", workoutWODHandler.RemoveWODFromWorkout) // Remove WOD from template
			r.Post("/templates/{id}/wods/{wod_id}/toggle-pr", workoutWODHandler.ToggleWODPR) // Toggle PR flag

			// User Workouts - Log workout instances (authenticated)
			r.Post("/user-workouts", userWorkoutHandler.LogWorkout)               // Log a workout instance
			r.Get("/user-workouts", userWorkoutHandler.ListLoggedWorkouts)        // List logged workouts
			r.Get("/user-workouts/{id}", userWorkoutHandler.GetLoggedWorkout)     // Get logged workout details
			r.Put("/user-workouts/{id}", userWorkoutHandler.UpdateLoggedWorkout)  // Update logged workout
			r.Delete("/user-workouts/{id}", userWorkoutHandler.DeleteLoggedWorkout) // Delete logged workout
			r.Get("/user-workouts/stats/month", userWorkoutHandler.GetWorkoutStatsForMonth) // Monthly stats
		})
	})

	// Configure HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}
