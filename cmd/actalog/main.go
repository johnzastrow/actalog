// Package main is the entry point for ActaLog application
//
// @title           ActaLog API
// @version         0.17.0
// @description     ActaLog is a mobile-first CrossFit workout tracker API. It provides endpoints for user authentication, workout logging, movement tracking, personal records, and administrative functions.
// @termsOfService  https://actalog.com/terms/
//
// @contact.name   ActaLog Support
// @contact.url    https://github.com/johnzastrow/actalog/issues
// @contact.email  support@actalog.com
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host      localhost:8080
// @BasePath  /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
//
// @tag.name auth
// @tag.description Authentication and authorization endpoints
//
// @tag.name users
// @tag.description User profile and settings management
//
// @tag.name workouts
// @tag.description Workout logging and history
//
// @tag.name movements
// @tag.description Movement/exercise definitions and management
//
// @tag.name wods
// @tag.description Workout of the Day management
//
// @tag.name templates
// @tag.description Workout template management
//
// @tag.name performance
// @tag.description Performance analytics and statistics
//
// @tag.name prs
// @tag.description Personal records tracking
//
// @tag.name notifications
// @tag.description User notifications and announcements
//
// @tag.name sessions
// @tag.description Session management
//
// @tag.name subscriptions
// @tag.description Subscription and billing management
//
// @tag.name organizations
// @tag.description Organization management
//
// @tag.name admin
// @tag.description Administrative operations (admin only)
//
// @tag.name import-export
// @tag.description Data import and export operations
//
// @tag.name backups
// @tag.description System backup and restore (admin only)
//
// @tag.name audit
// @tag.description Audit log operations
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
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/johnzastrow/actalog/configs"
	_ "github.com/johnzastrow/actalog/docs" // Swagger docs
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
	startTime := time.Now()

	// Print startup banner with timestamp
	fmt.Println("============================================================")
	fmt.Printf("[%s] ActaLog Server Starting\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("============================================================")
	fmt.Println(version.String())

	// Log basic environment info for diagnostics
	hostname, _ := os.Hostname()
	fmt.Printf("[STARTUP] Hostname: %s\n", hostname)
	fmt.Printf("[STARTUP] Working Directory: %s\n", mustGetCwd())
	fmt.Printf("[STARTUP] Process ID: %d\n", os.Getpid())
	fmt.Println("[STARTUP] Loading configuration...")

	// Load .env file (ignore error if file doesn't exist)
	// In production, you should use actual environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("[STARTUP] No .env file found or error loading it, using environment variables or defaults")
	} else {
		fmt.Println("[STARTUP] ✓ .env file loaded")
	}

	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("[STARTUP] ✗ Failed to load configuration: %v", err)
	}
	fmt.Println("[STARTUP] ✓ Configuration loaded")

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		EnableFile: cfg.Logging.EnableFile,
		FilePath:   cfg.Logging.FilePath,
		MaxSizeMB:  cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
	})
	if err != nil {
		log.Fatalf("[STARTUP] ✗ Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()
	fmt.Println("[STARTUP] ✓ Logger initialized")

	// Log comprehensive configuration (without sensitive data)
	fmt.Println("------------------------------------------------------------")
	fmt.Println("[CONFIG] Application Settings:")
	appLogger.Info("Environment: %s", cfg.App.Environment)
	appLogger.Info("Log Level: %s", cfg.Logging.Level)
	if cfg.Logging.EnableFile {
		appLogger.Info("File logging: enabled (%s)", cfg.Logging.FilePath)
	} else {
		appLogger.Info("File logging: disabled (stdout only)")
	}

	// Database configuration (detailed for debugging)
	fmt.Println("[CONFIG] Database Settings:")
	appLogger.Info("Database Driver: %s", cfg.Database.Driver)
	if cfg.Database.Driver != "sqlite3" {
		appLogger.Info("Database Host: %s", cfg.Database.Host)
		appLogger.Info("Database Port: %d", cfg.Database.Port)
		appLogger.Info("Database Name: %s", cfg.Database.Database)
		appLogger.Info("Database User: %s", cfg.Database.User)
		if cfg.Database.Schema != "" {
			appLogger.Info("Database Schema: %s", cfg.Database.Schema)
		}
	} else {
		appLogger.Info("Database File: %s", cfg.Database.Database)
	}

	fmt.Println("[CONFIG] Server Settings:")
	appLogger.Info("Server Address: %s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.Info("Allow Registration: %t", cfg.App.AllowRegistration)
	fmt.Println("------------------------------------------------------------")

	// Build database connection string
	dsn := repository.BuildDSN(
		cfg.Database.Driver,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
		cfg.Database.Schema,
	)

	// Initialize database with connection pooling
	db, err := repository.InitDatabase(cfg.Database.Driver, dsn, cfg.Database)
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
	dataChangeLogRepo := repository.NewDataChangeLogRepository(db, cfg.Database.Driver)
	orgRepo := repository.NewOrganizationRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	notificationLikeRepo := repository.NewNotificationLikeRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db, cfg.Database.Driver)

	// Subscription repositories
	userSubscriptionRepo := repository.NewSQLiteUserSubscriptionRepository(db)
	orgSubscriptionRepo := repository.NewSQLiteOrganizationSubscriptionRepository(db)
	subscriptionAccessRepo := repository.NewSubscriptionAccessRepository(userSubscriptionRepo, orgSubscriptionRepo, orgRepo)

	// Benchmark repository
	benchmarkRepo := repository.NewBenchmarkRepository(db, cfg.Database.Driver)

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
	dataChangeLogService := service.NewDataChangeLogService(dataChangeLogRepo)
	emailLogService := service.NewEmailLogService(emailLogRepo)

	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		userSubscriptionRepo,
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

	// Create admin notification service for user event emails and in-app notifications
	stdLogger := log.New(appLogger.Writer(), "[AdminNotify] ", 0)
	// Pass nil explicitly for email service interface to avoid typed-nil issue
	// (a typed nil *email.Service assigned to interface is not nil)
	var adminEmailSvc email.EmailService
	if emailService != nil {
		adminEmailSvc = emailService
	}
	adminNotificationService := service.NewAdminNotificationService(
		userRepo,
		userSettingsRepo,
		notificationRepo,
		adminEmailSvc,
		emailLogService,
		appURL,
		stdLogger,
	)
	userService.SetAdminNotificationService(adminNotificationService)
	appLogger.Info("Admin notification service: enabled (email: %v, in-app: true)", emailService != nil)

	notificationService := service.NewNotificationService(
		notificationRepo,
		orgRepo,
		userRepo,
		userSettingsRepo,
		emailService,
	)

	notificationLikeService := service.NewNotificationLikeService(
		notificationLikeRepo,
		notificationRepo,
	)

	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		workoutMovementRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
		wodRepo,
		auditLogRepo,
		movementRepo,
		notificationService,
		userRepo,
		orgRepo,
	)

	workoutTemplateService := service.NewWorkoutTemplateService(
		workoutRepo,
		workoutMovementRepo,
		workoutWODRepo,
		auditLogRepo,
	)

	wodService := service.NewWODService(wodRepo, dataChangeLogService, auditLogRepo)

	movementService := service.NewMovementService(movementRepo, dataChangeLogService, auditLogRepo)

	workoutWODService := service.NewWorkoutWODService(
		workoutWODRepo,
		workoutRepo,
		wodRepo,
	)

	userSettingsService := service.NewUserSettingsService(userSettingsRepo, auditLogRepo)

	orgService := service.NewOrganizationService(orgRepo, userRepo, auditLogRepo)

	subscriptionService := service.NewSubscriptionService(
		userSubscriptionRepo,
		orgSubscriptionRepo,
		subscriptionAccessRepo,
		auditLogRepo,
		userRepo,
		orgRepo,
	)

	exportService := service.NewExportService(wodRepo, movementRepo, userRepo, userWorkoutRepo)
	importService := service.NewImportService(wodRepo, movementRepo, userRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)
	wodifyImportService := service.NewWodifyImportService(userRepo, movementRepo, wodRepo, userWorkoutRepo, userWorkoutMovementRepo, userWorkoutWODRepo)

	// Determine backups and uploads directories
	workDir, _ := os.Getwd()
	backupDir := filepath.Join(workDir, "backups")
	uploadsPath := filepath.Join(workDir, "uploads")

	backupService := service.NewBackupService(
		db,
		cfg.Database.Driver,
		cfg.Database.Database,
		backupDir,
		uploadsPath,
		userRepo,
		auditLogRepo,
	)

	// Benchmark service
	benchmarkService := service.NewBenchmarkService(benchmarkRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, appLogger)
	userHandler := handler.NewUserHandler(userService, appLogger)
	movementHandler := handler.NewMovementHandler(movementRepo, movementService, appLogger)
	workoutTemplateHandler := handler.NewWorkoutTemplateHandler(workoutTemplateService)
	userWorkoutHandler := handler.NewUserWorkoutHandler(userWorkoutService, appLogger)
	wodHandler := handler.NewWODHandler(wodService)
	workoutWODHandler := handler.NewWorkoutWODHandler(workoutWODService)
	settingsHandler := handler.NewSettingsHandler(userSettingsService, appLogger)
	prHandler := handler.NewPRHandler(db, appLogger)
	performanceHandler := handler.NewPerformanceHandler(movementRepo, wodRepo, userWorkoutMovementRepo, userWorkoutWODRepo, appLogger)
	adminHandler := handler.NewAdminHandler(db, userWorkoutWODRepo, wodRepo, movementRepo, workoutRepo, userRepo, wodService, movementService, workoutTemplateService, appLogger)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService, appLogger)
	dataChangeLogHandler := handler.NewDataChangeLogHandler(dataChangeLogService, appLogger)
	adminUserHandler := handler.NewAdminUserHandler(userService, appLogger)
	sessionHandler := handler.NewSessionHandler(userService, appLogger)
	exportHandler := handler.NewExportHandler(exportService)
	importHandler := handler.NewImportHandler(importService)
	wodifyImportHandler := handler.NewWodifyImportHandler(wodifyImportService)
	backupHandler := handler.NewBackupHandler(backupService, auditLogRepo)
	orgHandler := handler.NewOrganizationHandler(orgService, appLogger)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, appLogger)
	notificationHandler := handler.NewNotificationHandler(notificationService, appLogger)
	notificationLikeHandler := handler.NewNotificationLikeHandler(notificationLikeService)
	emailHandler := handler.NewEmailHandler(emailService, emailLogService, appLogger)
	emailLogHandler := handler.NewEmailLogHandler(emailLogService, appLogger)
	benchmarkHandler := handler.NewBenchmarkHandler(benchmarkService, appLogger)

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

	// Get frontend directory from environment or use default
	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = filepath.Join(workDir, "web", "dist")
	}

	// Check if frontend directory exists
	frontendExists := false
	if _, err := os.Stat(frontendDir); err == nil {
		frontendExists = true
		appLogger.Info("Serving frontend from: %s", frontendDir)
	} else {
		appLogger.Info("Frontend directory not found: %s (API-only mode)", frontendDir)
	}

	// Root endpoint - only serve API JSON if no frontend available
	if !frontendExists {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"message":"Welcome to ActaLog API","version":"%s"}`, version.Version())
		})
	}

	// Static file serving for uploads (avatars, etc.)
	uploadsDir := http.Dir(uploadsPath)
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

		// Swagger documentation UI (public)
		r.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/api/docs/doc.json"), // The url pointing to API definition
		))

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
		r.Get("/wods/standard", wodHandler.ListStandardWODs)
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

			// Notification routes (authenticated)
			r.Get("/notifications", notificationHandler.ListNotifications)
			r.Get("/notifications/unread", notificationHandler.ListUnreadNotifications)
			r.Get("/notifications/count", notificationHandler.GetUnreadCount)
			r.Put("/notifications/{id}/read", notificationHandler.MarkAsRead)
			r.Put("/notifications/read-all", notificationHandler.MarkAllAsRead)
			r.Delete("/notifications/{id}", notificationHandler.DeleteNotification)

			// Notification Like routes (authenticated)
			r.Post("/notifications/{id}/like", notificationLikeHandler.LikeNotification)
			r.Delete("/notifications/{id}/like", notificationLikeHandler.UnlikeNotification)
			r.Get("/notifications/{id}/likes", notificationLikeHandler.GetNotificationLikes)

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
			r.Get("/wods/my-wods", wodHandler.ListMyWODs)
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

			// Statistics routes (authenticated)
			r.Get("/stats/active-users-this-month", userWorkoutHandler.GetActiveUsersStats)

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
			r.Post("/import/wodify/preview", wodifyImportHandler.PreviewWodifyImport)
			r.Post("/import/wodify/confirm", wodifyImportHandler.ConfirmWodifyImport)

			// Subscription status (user-accessible)
			r.Get("/subscriptions/status", subscriptionHandler.GetMySubscriptionStatus)

			// Benchmark routes (authenticated)
			r.Post("/benchmark", benchmarkHandler.RunBenchmark)
			r.Get("/benchmark/status", benchmarkHandler.GetBenchmarkStatus)

			// Admin routes (authenticated + admin role check)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				// Backup routes (admin only)
				r.Post("/backups", backupHandler.CreateBackup)
				r.Get("/backups", backupHandler.ListBackups)
				r.Post("/backups/upload", backupHandler.UploadBackup)
				r.Get("/backups/{filename}", backupHandler.DownloadBackup)
				r.Get("/backups/{filename}/metadata", backupHandler.GetBackupMetadata)
				r.Delete("/backups/{filename}", backupHandler.DeleteBackup)
				r.Post("/backups/{filename}/restore", backupHandler.RestoreBackup)

				// Email admin routes (admin only)
				r.Route("/email", func(r chi.Router) {
					r.Get("/config", emailHandler.GetEmailConfig)
					r.Post("/test", emailHandler.SendTestEmail)
				})

				// Email logs routes (admin only)
				r.Route("/email-logs", func(r chi.Router) {
					r.Get("/", emailLogHandler.ListEmailLogs)
					r.Get("/stats", emailLogHandler.GetEmailStats)
					r.Get("/failures", emailLogHandler.GetRecentFailures)
					r.Post("/cleanup", emailLogHandler.CleanupEmailLogs)
					r.Get("/{id}", emailLogHandler.GetEmailLog)
				})

				// Announcement routes (admin only)
				r.Post("/notifications/announce", notificationHandler.CreateAnnouncement)

				// Data cleanup routes
				r.Get("/data-cleanup/wod-mismatches", adminHandler.DetectWODScoreTypeMismatches)
				r.Delete("/data-cleanup/wod-mismatches", adminHandler.FixWODScoreTypeMismatches)
				r.Put("/data-cleanup/wod-record/{id}", adminHandler.UpdateWODRecord)

				// Benchmark data cleanup (admin only)
				r.Delete("/benchmark/data", benchmarkHandler.CleanupBenchmarkData)

				// Audit log routes (admin only)
				r.Get("/audit-logs", auditLogHandler.ListAuditLogs)
				r.Get("/audit-logs/{id}", auditLogHandler.GetAuditLog)
				r.Post("/audit-logs/cleanup", auditLogHandler.CleanupOldLogs)

				// Data change log routes (admin only)
				r.Get("/data-change-logs", dataChangeLogHandler.ListDataChangeLogs)
				r.Get("/data-change-logs/{id}", dataChangeLogHandler.GetDataChangeLog)
				r.Get("/data-change-logs/entity/{entity_type}/{entity_id}", dataChangeLogHandler.GetEntityHistory)
				r.Post("/data-change-logs/cleanup", dataChangeLogHandler.CleanupOldLogs)

				// User management routes (admin only)
				r.Get("/users", adminUserHandler.ListUsers)
				r.Post("/users/{id}/unlock", adminUserHandler.UnlockUser)
				r.Get("/users/{id}", adminUserHandler.GetUserDetails)
				r.Post("/users/{id}/disable", adminUserHandler.DisableUser)
				r.Post("/users/{id}/enable", adminUserHandler.EnableUser)
				r.Put("/users/{id}/role", adminUserHandler.ChangeUserRole)
				r.Post("/users/{id}/toggle-email-verification", adminUserHandler.ToggleEmailVerification)
				r.Delete("/users/{id}", adminUserHandler.DeleteUser)

				// User-created content management routes (admin only)
				r.Get("/user-created/wods", adminHandler.ListUserCreatedWODs)
				r.Post("/user-created/wods/{id}/copy-to-standard", adminHandler.CopyWODToStandard)
				r.Get("/user-created/movements", adminHandler.ListUserCreatedMovements)
				r.Post("/user-created/movements/{id}/copy-to-standard", adminHandler.CopyMovementToStandard)
				r.Get("/user-created/workouts", adminHandler.ListUserCreatedWorkouts)
				r.Post("/user-created/workouts/{id}/copy-to-standard", adminHandler.CopyWorkoutToStandard)

				// Organization management routes (admin only)
				r.Post("/organizations", orgHandler.CreateOrganization)
				r.Get("/organizations", orgHandler.ListOrganizations)
				r.Get("/organizations/{id}", orgHandler.GetOrganization)
				r.Put("/organizations/{id}", orgHandler.UpdateOrganization)
				r.Delete("/organizations/{id}", orgHandler.DeleteOrganization)

				// User-organization assignment (admin only)
				r.Post("/users/{id}/organization", orgHandler.AssignUserToOrganization)
				r.Delete("/users/{id}/organization/{org_id}", orgHandler.RemoveUserFromOrganization)
				r.Get("/users/{id}/organizations", orgHandler.GetUserOrganizations)
				r.Get("/organizations/{id}/users", orgHandler.GetOrganizationUsers)

				// Subscription management routes (admin only)
				r.Route("/subscriptions", func(r chi.Router) {
					// List all subscriptions (must come before parameterized routes)
					r.Get("/users", subscriptionHandler.ListAllUserSubscriptions)
					r.Get("/organizations", subscriptionHandler.ListAllOrganizationSubscriptions)

					// Expiring and expired subscriptions (must come before parameterized routes)
					r.Get("/users/expiring", subscriptionHandler.ListExpiringUserSubscriptions)
					r.Get("/users/expired", subscriptionHandler.ListExpiredUserSubscriptions)
					r.Get("/organizations/expiring", subscriptionHandler.ListExpiringOrganizationSubscriptions)
					r.Get("/organizations/expired", subscriptionHandler.ListExpiredOrganizationSubscriptions)

					// User subscriptions
					r.Post("/user", subscriptionHandler.CreateUserSubscription)
					r.Get("/user/{user_id}", subscriptionHandler.GetUserSubscriptions)
					r.Post("/user/{id}/mark-paid", subscriptionHandler.MarkUserSubscriptionAsPaid)
					r.Post("/user/{id}/cancel", subscriptionHandler.CancelUserSubscription)
					r.Post("/user/{id}/set-permanent", subscriptionHandler.SetUserSubscriptionPermanent)

					// Organization subscriptions
					r.Post("/organization", subscriptionHandler.CreateOrganizationSubscription)
					r.Get("/organization/{org_id}", subscriptionHandler.GetOrganizationSubscriptions)
					r.Post("/organization/{id}/mark-paid", subscriptionHandler.MarkOrganizationSubscriptionAsPaid)
					r.Post("/organization/{id}/cancel", subscriptionHandler.CancelOrganizationSubscription)
					r.Post("/organization/{id}/set-permanent", subscriptionHandler.SetOrganizationSubscriptionPermanent)
				})
			})
		})
	})

	// Serve frontend static files (must be after API routes to allow API to take precedence)
	// Serve static assets (CSS, JS, images, etc.)
	if frontendExists {
		fs := http.FileServer(http.Dir(frontendDir))
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			// Build the full file path
			filePath := filepath.Join(frontendDir, req.URL.Path)

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// File doesn't exist, serve index.html for SPA routing
				http.ServeFile(w, req, filepath.Join(frontendDir, "index.html"))
				return
			}

			// File exists, serve it
			fs.ServeHTTP(w, req)
		})
	}

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
		fmt.Println("============================================================")
		fmt.Printf("[%s] ✓ SERVER READY\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("[STARTUP] Total startup time: %v\n", time.Since(startTime))
		fmt.Printf("[STARTUP] Server listening on %s\n", addr)
		fmt.Println("============================================================")
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

// mustGetCwd returns the current working directory or "unknown" if it fails
func mustGetCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}
