// Package main provides the CLI for the ActaLog scheduler
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnzastrow/actalog/configs"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/scheduler"
	"github.com/johnzastrow/actalog/pkg/version"
	"github.com/joho/godotenv"

	// Database drivers
	_ "github.com/go-sql-driver/mysql" // MySQL/MariaDB
	_ "github.com/lib/pq"              // PostgreSQL
	_ "github.com/mattn/go-sqlite3"    // SQLite
)

func main() {
	// Define subcommands
	materializeCmd := flag.NewFlagSet("materialize", flag.ExitOnError)
	daysAhead := materializeCmd.Int("days-ahead", 14, "Number of days ahead to materialize sessions")
	orgID := materializeCmd.Int64("org-id", 0, "Organization ID (0 for all organizations)")

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "materialize":
		materializeCmd.Parse(os.Args[2:])
		runMaterialize(*daysAhead, *orgID)
	case "version":
		fmt.Printf("ActaLog Scheduler v%s\n", version.Version())
		os.Exit(0)
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ActaLog Scheduler - Session Materialization Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scheduler <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  materialize    Generate class sessions from templates and schedule slots")
	fmt.Println("  version        Print version information")
	fmt.Println("  help           Print this help message")
	fmt.Println()
	fmt.Println("Options for 'materialize':")
	fmt.Println("  -days-ahead int    Number of days ahead to materialize sessions (default: 14)")
	fmt.Println("  -org-id int        Organization ID to materialize (0 for all organizations)")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DB_DRIVER    Database driver (sqlite3, postgres, mysql)")
	fmt.Println("  DB_HOST      Database host")
	fmt.Println("  DB_PORT      Database port")
	fmt.Println("  DB_USER      Database user")
	fmt.Println("  DB_PASSWORD  Database password")
	fmt.Println("  DB_NAME      Database name")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scheduler materialize                    # Materialize for all orgs, 14 days ahead")
	fmt.Println("  scheduler materialize -days-ahead 7      # Materialize 7 days ahead")
	fmt.Println("  scheduler materialize -org-id 1          # Materialize for org 1 only")
}

func runMaterialize(daysAhead int, orgID int64) {
	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load config
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("Starting session materialization...")
	appLogger.Info("Days ahead: %d", daysAhead)
	if orgID > 0 {
		appLogger.Info("Organization ID: %d", orgID)
	} else {
		appLogger.Info("Materializing for all organizations")
	}

	// Build DSN
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

	// Connect to database
	db, err := repository.InitDatabase(cfg.Database.Driver, dsn, cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()
	appLogger.Info("Database connected")

	// Initialize repositories
	templateRepo := repository.NewClassTemplateRepository(db)
	slotRepo := repository.NewScheduleSlotRepository(db)
	sessionRepo := repository.NewClassSessionRepository(db)

	// Create materializer
	config := scheduler.MaterializerConfig{
		DaysAhead: daysAhead,
	}
	materializer := scheduler.NewMaterializer(templateRepo, slotRepo, sessionRepo, appLogger, config)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run materialization
	var result *scheduler.MaterializeResult
	if orgID > 0 {
		result, err = materializer.MaterializeForOrganization(ctx, orgID)
	} else {
		result, err = materializer.MaterializeAll(ctx)
	}

	if err != nil {
		appLogger.Fatal("Materialization failed: %v", err)
	}

	// Print results
	appLogger.Info("=== Materialization Results ===")
	appLogger.Info("Templates processed: %d", result.TemplatesProcessed)
	appLogger.Info("Slots processed: %d", result.SlotsProcessed)
	appLogger.Info("Sessions created: %d", result.SessionsCreated)
	appLogger.Info("Sessions skipped (already exist): %d", result.SessionsSkipped)

	if len(result.Errors) > 0 {
		appLogger.Warn("Errors encountered: %d", len(result.Errors))
		for _, e := range result.Errors {
			appLogger.Error("  - %v", e)
		}
	}

	appLogger.Info("Materialization complete!")
}
