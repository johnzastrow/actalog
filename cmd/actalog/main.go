// Package main is the entry point for ActaLog application
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/configs"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/email"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/version"
	"github.com/joho/godotenv"

	// Database drivers
	_ "github.com/go-sql-driver/mysql" // MySQL/MariaDB
	_ "github.com/lib/pq"              // PostgreSQL
	_ "github.com/mattn/go-sqlite3"    // SQLite
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
	auditLogRepo := repository.NewAuditLogRepository(db, cfg.Database.Driver)
	movementRepo := repository.NewMovementRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)
	wodRepo := repository.NewWODRepository(db)
	userWorkoutRepo := repository.NewUserWorkoutRepository(db)
	workoutWODRepo := repository.NewWorkoutWODRepository(db)
	userSettingsRepo := repository.NewSQLiteUserSettingsRepository(db)
	userWorkoutMovementRepo := repository.NewUserWorkoutMovementRepository(db)
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)

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
	auditLogService := service.NewAuditLogService(auditLogRepo)

	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		auditLogService,
		cfg.JWT.SecretKey,
		cfg.JWT.ExpirationTime,
		cfg.JWT.RefreshTokenDuration,
		cfg.App.AllowRegistration,
		emailService,
		appURL,
		cfg.Email.RequireVerification,
		cfg.Security.MaxLoginAttempts,
		cfg.Security.AccountLockoutDuration,
	)

	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		workoutMovementRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
		wodRepo,
	)

	workoutTemplateService := service.NewWorkoutTemplateService(
		workoutRepo,
		workoutMovementRepo,
		workoutWODRepo,
	)

	wodService := service.NewWODService(wodRepo)

	workoutWODService := service.NewWorkoutWODService(
		workoutWODRepo,
		workoutRepo,
		wodRepo,
	)

	userSettingsService := service.NewUserSettingsService(userSettingsRepo)

	exportService := service.NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)
	importService := service.NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, appLogger)
	userHandler := handler.NewUserHandler(userService, appLogger)
	movementHandler := handler.NewMovementHandler(movementRepo, appLogger)
	workoutTemplateHandler := handler.NewWorkoutTemplateHandler(workoutTemplateService)
	userWorkoutHandler := handler.NewUserWorkoutHandler(userWorkoutService, appLogger)
	wodHandler := handler.NewWODHandler(wodService)
	workoutWODHandler := handler.NewWorkoutWODHandler(workoutWODService)
	settingsHandler := handler.NewSettingsHandler(userSettingsService, appLogger)
	prHandler := handler.NewPRHandler(db, appLogger)
	performanceHandler := handler.NewPerformanceHandler(movementRepo, wodRepo, userWorkoutMovementRepo, userWorkoutWODRepo, appLogger)
	adminHandler := handler.NewAdminHandler(db, userWorkoutWODRepo, wodRepo, userRepo, appLogger)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService, appLogger)
	adminUserHandler := handler.NewAdminUserHandler(userService, appLogger)
	sessionHandler := handler.NewSessionHandler(userService, appLogger)
	exportHandler := handler.NewExportHandler(exportService)
	importHandler := handler.NewImportHandler(importService)

	// Set up router
	r := chi.NewRouter()

	// Initialize rate limiters
	// Login/Register: 5 attempts per 15 minutes per IP
	authRateLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	// Password reset: 3 attempts per hour per IP
	passwordResetLimiter := middleware.NewRateLimiter(3, 1*time.Hour)

	// Middleware
	r.Use(middleware.LoggingMiddleware(appLogger))
	r.Use(middleware.CORS(cfg.App.CORSOrigins))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"%s"}`, version.Version())
	})

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Welcome to ActaLog API","version":"%s"}`, version.Version())
	})

	// Static file serving for uploads (avatars, etc.)
	workDir, _ := os.Getwd()
	uploadsDir := http.Dir(filepath.Join(workDir, "uploads"))
	FileServer(r, "/uploads", uploadsDir)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Version endpoint (public)
		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"version":"%s","build":%d,"fullVersion":"%s","app":"%s"}`,
				version.Version(), version.BuildNumber(), version.FullVersion(), cfg.App.Name)
		})

		// Auth routes (public with rate limiting)
		r.With(middleware.RateLimit(authRateLimiter)).Post("/auth/register", authHandler.Register)
		r.With(middleware.RateLimit(authRateLimiter)).Post("/auth/login", authHandler.Login)
		r.With(middleware.RateLimit(passwordResetLimiter)).Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.With(middleware.RateLimit(passwordResetLimiter)).Post("/auth/reset-password", authHandler.ResetPassword)
		r.Get("/auth/verify-email", authHandler.VerifyEmail)
		r.With(middleware.RateLimit(authRateLimiter)).Post("/auth/resend-verification", authHandler.ResendVerification)
		r.Post("/auth/refresh", authHandler.RefreshToken)
		r.Post("/auth/revoke", authHandler.RevokeToken)

		// Movement routes (public for browsing)
		r.Get("/movements", movementHandler.ListAll)
		r.Get("/movements/search", movementHandler.Search)
		r.Get("/movements/{id}", movementHandler.GetByID)

		// WOD routes (public for browsing standard WODs)
		r.Get("/wods", wodHandler.ListWODs)
		r.Get("/wods/search", wodHandler.SearchWODs)
		r.Get("/wods/{id}", wodHandler.GetWOD)

		// Template routes (public for browsing standard templates)
		r.Get("/templates", workoutTemplateHandler.ListStandardTemplates)
		r.Get("/templates/{id}", workoutTemplateHandler.GetTemplate)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWT.SecretKey))

			// Movement management (authenticated)
			r.Post("/movements", movementHandler.Create)
			r.Put("/movements/{id}", movementHandler.Update)
			r.Delete("/movements/{id}", movementHandler.Delete)

			// User profile routes (authenticated)
			r.Get("/users/profile", userHandler.GetProfile)
			r.Put("/users/profile", userHandler.UpdateProfile)
			r.Post("/users/avatar", userHandler.UploadAvatar)
			r.Delete("/users/avatar", userHandler.DeleteAvatar)

			// User settings routes (authenticated)
			r.Get("/users/settings", settingsHandler.GetSettings)
			r.Put("/users/settings", settingsHandler.UpdateSettings)
			r.Put("/users/password", userHandler.ChangePassword)

			// User audit log routes (authenticated - own logs only)
			r.Get("/users/me/audit-logs", auditLogHandler.GetMyAuditLogs)

			// Session management routes (authenticated)
			r.Get("/sessions", sessionHandler.ListSessions)
			r.Delete("/sessions/{id}", sessionHandler.RevokeSession)
			r.Post("/sessions/revoke-all", sessionHandler.RevokeAllSessions)

			// Workout Template routes (authenticated)
			r.Post("/templates", workoutTemplateHandler.CreateTemplate)
			r.Get("/workouts/my-templates", workoutTemplateHandler.ListMyTemplates)
			r.Put("/templates/{id}", workoutTemplateHandler.UpdateTemplate)
			r.Delete("/templates/{id}", workoutTemplateHandler.DeleteTemplate)

			// User Workout routes (logging workouts) (authenticated)
			r.Post("/workouts", userWorkoutHandler.LogWorkout)
			r.Get("/workouts", userWorkoutHandler.ListLoggedWorkouts)
			r.Get("/workouts/standard", workoutTemplateHandler.ListStandardTemplates)
			r.Get("/workouts/{id}", userWorkoutHandler.GetLoggedWorkout)
			r.Put("/workouts/{id}", userWorkoutHandler.UpdateLoggedWorkout)
			r.Delete("/workouts/{id}", userWorkoutHandler.DeleteLoggedWorkout)
			r.Get("/workouts/stats/monthly", userWorkoutHandler.GetMonthlyStats)
			r.Get("/workouts/personal-records", userWorkoutHandler.GetPersonalRecords)
			r.Post("/workouts/retroactive-flag-prs", userWorkoutHandler.RetroactiveFlagPRs)

			// WOD management (authenticated)
			r.Post("/wods", wodHandler.CreateWOD)
			r.Put("/wods/{id}", wodHandler.UpdateWOD)
			r.Delete("/wods/{id}", wodHandler.DeleteWOD)

			// Workout WOD linking (authenticated)
			r.Post("/templates/{workout_id}/wods", workoutWODHandler.AddWODToWorkout)
			r.Get("/templates/{workout_id}/wods", workoutWODHandler.ListWODsForWorkout)
			r.Put("/templates/wods/{workout_wod_id}", workoutWODHandler.UpdateWorkoutWOD)
			r.Delete("/templates/wods/{workout_wod_id}", workoutWODHandler.RemoveWODFromWorkout)
			r.Post("/templates/wods/{workout_wod_id}/toggle-pr", workoutWODHandler.ToggleWODPR)

			// PR tracking routes (authenticated)
			r.Get("/prs", prHandler.GetPersonalRecords)
			r.Get("/pr-movements", prHandler.GetPRMovements)
			r.Post("/movements/toggle-pr", prHandler.ToggleMovementPR)

			// Performance tracking routes (authenticated)
			r.Get("/performance/search", performanceHandler.UnifiedSearch)
			r.Get("/performance/movements/{id}", performanceHandler.GetMovementPerformance)
			r.Get("/performance/wods/{id}", performanceHandler.GetWODPerformance)

			// Export routes (authenticated)
			r.Get("/export/wods", exportHandler.ExportWODs)
			r.Get("/export/movements", exportHandler.ExportMovements)
			r.Get("/export/user-workouts", exportHandler.ExportUserWorkouts)

			// Import routes (authenticated)
			r.Post("/import/wods/preview", importHandler.PreviewWODImport)
			r.Post("/import/wods/confirm", importHandler.ConfirmWODImport)
			r.Post("/import/movements/preview", importHandler.PreviewMovementImport)
			r.Post("/import/movements/confirm", importHandler.ConfirmMovementImport)
			r.Post("/import/user-workouts/preview", importHandler.PreviewUserWorkoutImport)
			r.Post("/import/user-workouts/confirm", importHandler.ConfirmUserWorkoutImport)

			// Admin routes (authenticated + admin role check)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				// Data cleanup routes
				r.Get("/data-cleanup/wod-mismatches", adminHandler.DetectWODScoreTypeMismatches)
				r.Delete("/data-cleanup/wod-mismatches", adminHandler.FixWODScoreTypeMismatches)
				r.Put("/data-cleanup/wod-record/{id}", adminHandler.UpdateWODRecord)

				// Audit log routes (admin only)
				r.Get("/audit-logs", auditLogHandler.ListAuditLogs)
				r.Get("/audit-logs/{id}", auditLogHandler.GetAuditLog)
				r.Post("/audit-logs/cleanup", auditLogHandler.CleanupOldLogs)

				// User management routes (admin only)
				r.Get("/users", adminUserHandler.ListUsers)
				r.Post("/users/{id}/unlock", adminUserHandler.UnlockUser)
				r.Get("/users/{id}", adminUserHandler.GetUserDetails)
				r.Post("/users/{id}/disable", adminUserHandler.DisableUser)
				r.Post("/users/{id}/enable", adminUserHandler.EnableUser)
				r.Put("/users/{id}/role", adminUserHandler.ChangeUserRole)
				r.Post("/users/{id}/toggle-email-verification", adminUserHandler.ToggleEmailVerification)
				r.Delete("/users/{id}", adminUserHandler.DeleteUser)
			})
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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
