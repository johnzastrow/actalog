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
	"github.com/johnzastrow/actalog/configs"
	"github.com/johnzastrow/actalog/internal/handler"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/version"

	// Database drivers
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Print version information
	fmt.Println(version.String())
	fmt.Println("Starting ActaLog server...")

	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration (without sensitive data)
	log.Printf("Environment: %s", cfg.App.Environment)
	log.Printf("Log Level: %s", cfg.App.LogLevel)
	log.Printf("Database Driver: %s", cfg.Database.Driver)
	log.Printf("Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Allow Registration: %t", cfg.App.AllowRegistration)

	// Initialize database
	db, err := repository.InitDatabase(cfg.Database.Driver, cfg.Database.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	movementRepo := repository.NewSQLiteMovementRepository(db)

	// Initialize services
	userService := service.NewUserService(
		userRepo,
		cfg.JWT.SecretKey,
		cfg.JWT.ExpirationTime,
		cfg.App.AllowRegistration,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService)
	movementHandler := handler.NewMovementHandler(movementRepo)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
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
		// Auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Movement routes
		r.Get("/movements", movementHandler.ListStandard)
		r.Get("/movements/search", movementHandler.Search)
		r.Get("/movements/{id}", movementHandler.GetByID)
		r.Post("/movements", movementHandler.Create)
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
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
