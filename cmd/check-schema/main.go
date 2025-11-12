package main

import (
	"database/sql"
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
	)

	// Open database connection
	db, err := sql.Open(cfg.Database.Driver, dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Query tables
	var query string
	if cfg.Database.Driver == "sqlite3" {
		query = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
	} else if cfg.Database.Driver == "postgres" {
		query = "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename"
	} else {
		query = "SHOW TABLES"
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== Database Tables ===")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("  - %s\n", tableName)
	}

	// Check migrations
	fmt.Println("\n=== Applied Migrations ===")
	rows2, err := db.Query("SELECT version, description FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		log.Fatalf("Failed to query migrations: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var version, description string
		if err := rows2.Scan(&version, &description); err != nil {
			log.Fatalf("Failed to scan migration: %v", err)
		}
		fmt.Printf("  ✓ %s - %s\n", version, description)
	}
}
