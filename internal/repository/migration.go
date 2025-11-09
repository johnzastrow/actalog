package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// MigrationInfo holds information about a migration
type MigrationInfo struct {
	Version   string
	UpSQL     string
	DownSQL   string
	AppliedAt *string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migrations
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			fmt.Printf("Migration %s already applied, skipping\n", migration.Version)
			continue
		}

		fmt.Printf("Applying migration %s...\n", migration.Version)
		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
		fmt.Printf("Migration %s applied successfully\n", migration.Version)
	}

	return nil
}

// RollbackMigration rolls back the most recent migration
func RollbackMigration(db *sql.DB) error {
	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Get the most recent migration
	var versions []string
	for v := range applied {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	latestVersion := versions[len(versions)-1]

	// Load all migrations
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Find the migration to rollback
	var migrationToRollback *MigrationInfo
	for _, m := range migrations {
		if m.Version == latestVersion {
			migrationToRollback = &m
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("migration %s not found", latestVersion)
	}

	fmt.Printf("Rolling back migration %s...\n", latestVersion)
	if err := rollbackMigration(db, *migrationToRollback); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", latestVersion, err)
	}
	fmt.Printf("Migration %s rolled back successfully\n", latestVersion)

	return nil
}

// createMigrationsTable creates the schema_migrations table to track applied migrations
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem
func loadMigrations() ([]MigrationInfo, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	migrationsMap := make(map[string]*MigrationInfo)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse filename: 000001_v0.3.0_schema.up.sql or 000001_v0.3.0_schema.down.sql
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		direction := ""
		if strings.HasSuffix(name, ".up.sql") {
			direction = "up"
		} else if strings.HasSuffix(name, ".down.sql") {
			direction = "down"
		} else {
			continue
		}

		// Read file content
		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		// Get or create migration info
		if _, ok := migrationsMap[version]; !ok {
			migrationsMap[version] = &MigrationInfo{Version: version}
		}

		// Set SQL based on direction
		if direction == "up" {
			migrationsMap[version].UpSQL = string(content)
		} else {
			migrationsMap[version].DownSQL = string(content)
		}
	}

	// Convert map to sorted slice
	var migrations []MigrationInfo
	var versions []string
	for v := range migrationsMap {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	for _, v := range versions {
		migrations = append(migrations, *migrationsMap[v])
	}

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[string]string, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]string)
	for rows.Next() {
		var version, appliedAt string
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}

	return applied, rows.Err()
}

// applyMigration executes an "up" migration
func applyMigration(db *sql.DB, migration MigrationInfo) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	query := `INSERT INTO schema_migrations (version) VALUES (?)`
	if _, err := tx.Exec(query, migration.Version); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// rollbackMigration executes a "down" migration
func rollbackMigration(db *sql.DB, migration MigrationInfo) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.Exec(migration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	query := `DELETE FROM schema_migrations WHERE version = ?`
	if _, err := tx.Exec(query, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return tx.Commit()
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *sql.DB) ([]MigrationInfo, error) {
	// Load all migrations
	migrations, err := loadMigrations()
	if err != nil {
		return nil, err
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return nil, err
	}

	// Set applied_at for each migration
	for i := range migrations {
		if appliedAt, ok := applied[migrations[i].Version]; ok {
			migrations[i].AppliedAt = &appliedAt
		}
	}

	return migrations, nil
}
