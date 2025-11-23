package main

import (
	"fmt"
	"log"

	"github.com/johnzastrow/actalog/configs"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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

	// Initialize database
	db, err := repository.InitDatabase(cfg.Database.Driver, dsn, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database connected successfully")
	fmt.Println("Running migrations...")

	// Run migrations
	if err := repository.RunMigrations(db, cfg.Database.Driver); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("✓ All migrations completed successfully")
}
