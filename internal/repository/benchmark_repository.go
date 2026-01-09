package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// BenchmarkRepository implements domain.BenchmarkRepository
type BenchmarkRepository struct {
	db     *sql.DB
	driver string
}

// NewBenchmarkRepository creates a new benchmark repository
func NewBenchmarkRepository(db *sql.DB, driver string) *BenchmarkRepository {
	return &BenchmarkRepository{
		db:     db,
		driver: driver,
	}
}

// Create creates a new benchmark data entry
func (r *BenchmarkRepository) Create(data *domain.BenchmarkData) error {
	now := time.Now()
	data.CreatedAt = now
	data.UpdatedAt = now

	switch r.driver {
	case "sqlite3", "mysql":
		query := `INSERT INTO benchmark_data (test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := r.db.Exec(query, data.TestKey, data.TestValue, data.NumValue, data.IntValue, data.BoolValue, data.CreatedBy, data.CreatedAt, data.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create benchmark data: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get benchmark data ID: %w", err)
		}
		data.ID = id

	case "postgres":
		query := `INSERT INTO benchmark_data (test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
		err := r.db.QueryRow(query, data.TestKey, data.TestValue, data.NumValue, data.IntValue, data.BoolValue, data.CreatedBy, data.CreatedAt, data.UpdatedAt).Scan(&data.ID)
		if err != nil {
			return fmt.Errorf("failed to create benchmark data: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	return nil
}

// CreateBatch creates multiple benchmark data entries
func (r *BenchmarkRepository) CreateBatch(data []*domain.BenchmarkData) error {
	if len(data) == 0 {
		return nil
	}

	now := time.Now()

	// Use transaction for batch insert
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `INSERT INTO benchmark_data (test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	case "postgres":
		query = `INSERT INTO benchmark_data (test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range data {
		d.CreatedAt = now
		d.UpdatedAt = now
		_, err := stmt.Exec(d.TestKey, d.TestValue, d.NumValue, d.IntValue, d.BoolValue, d.CreatedBy, d.CreatedAt, d.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert benchmark data: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a benchmark data entry by ID
func (r *BenchmarkRepository) GetByID(id int64) (*domain.BenchmarkData, error) {
	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data WHERE id = ?`
	case "postgres":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data WHERE id = $1`
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	data := &domain.BenchmarkData{}
	var testValue sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&data.ID, &data.TestKey, &testValue, &data.NumValue, &data.IntValue, &data.BoolValue,
		&data.CreatedBy, &data.CreatedAt, &data.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("benchmark data not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark data: %w", err)
	}

	if testValue.Valid {
		data.TestValue = testValue.String
	}

	return data, nil
}

// GetByKey retrieves a benchmark data entry by test key
func (r *BenchmarkRepository) GetByKey(testKey string) (*domain.BenchmarkData, error) {
	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data WHERE test_key = ? LIMIT 1`
	case "postgres":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data WHERE test_key = $1 LIMIT 1`
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	data := &domain.BenchmarkData{}
	var testValue sql.NullString
	err := r.db.QueryRow(query, testKey).Scan(
		&data.ID, &data.TestKey, &testValue, &data.NumValue, &data.IntValue, &data.BoolValue,
		&data.CreatedBy, &data.CreatedAt, &data.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("benchmark data not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark data: %w", err)
	}

	if testValue.Valid {
		data.TestValue = testValue.String
	}

	return data, nil
}

// List retrieves benchmark data entries with pagination
func (r *BenchmarkRepository) List(limit, offset int) ([]*domain.BenchmarkData, error) {
	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data ORDER BY id LIMIT ? OFFSET ?`
	case "postgres":
		query = `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
			FROM benchmark_data ORDER BY id LIMIT $1 OFFSET $2`
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark data: %w", err)
	}
	defer rows.Close()

	var results []*domain.BenchmarkData
	for rows.Next() {
		data := &domain.BenchmarkData{}
		var testValue sql.NullString
		err := rows.Scan(
			&data.ID, &data.TestKey, &testValue, &data.NumValue, &data.IntValue, &data.BoolValue,
			&data.CreatedBy, &data.CreatedAt, &data.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan benchmark data: %w", err)
		}
		if testValue.Valid {
			data.TestValue = testValue.String
		}
		results = append(results, data)
	}

	return results, rows.Err()
}

// ListFiltered retrieves benchmark data with filters
func (r *BenchmarkRepository) ListFiltered(filters domain.BenchmarkFilters, limit, offset int) ([]*domain.BenchmarkData, error) {
	query := `SELECT id, test_key, test_value, num_value, int_value, bool_value, created_by, created_at, updated_at
		FROM benchmark_data WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.CreatedBy != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		} else {
			query += " AND created_by = ?"
		}
		args = append(args, *filters.CreatedBy)
		argIndex++
	}

	if filters.TestKey != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND test_key LIKE $%d", argIndex)
		} else {
			query += " AND test_key LIKE ?"
		}
		args = append(args, *filters.TestKey+"%")
		argIndex++
	}

	if filters.BoolValue != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND bool_value = $%d", argIndex)
		} else {
			query += " AND bool_value = ?"
		}
		args = append(args, *filters.BoolValue)
		argIndex++
	}

	query += " ORDER BY id"

	// Add pagination
	if r.driver == "postgres" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	} else {
		query += " LIMIT ? OFFSET ?"
	}
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list filtered benchmark data: %w", err)
	}
	defer rows.Close()

	var results []*domain.BenchmarkData
	for rows.Next() {
		data := &domain.BenchmarkData{}
		var testValue sql.NullString
		err := rows.Scan(
			&data.ID, &data.TestKey, &testValue, &data.NumValue, &data.IntValue, &data.BoolValue,
			&data.CreatedBy, &data.CreatedAt, &data.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan benchmark data: %w", err)
		}
		if testValue.Valid {
			data.TestValue = testValue.String
		}
		results = append(results, data)
	}

	return results, rows.Err()
}

// Update updates a benchmark data entry
func (r *BenchmarkRepository) Update(data *domain.BenchmarkData) error {
	data.UpdatedAt = time.Now()

	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `UPDATE benchmark_data SET test_key = ?, test_value = ?, num_value = ?, int_value = ?, bool_value = ?, updated_at = ?
			WHERE id = ?`
		_, err := r.db.Exec(query, data.TestKey, data.TestValue, data.NumValue, data.IntValue, data.BoolValue, data.UpdatedAt, data.ID)
		if err != nil {
			return fmt.Errorf("failed to update benchmark data: %w", err)
		}

	case "postgres":
		query = `UPDATE benchmark_data SET test_key = $1, test_value = $2, num_value = $3, int_value = $4, bool_value = $5, updated_at = $6
			WHERE id = $7`
		_, err := r.db.Exec(query, data.TestKey, data.TestValue, data.NumValue, data.IntValue, data.BoolValue, data.UpdatedAt, data.ID)
		if err != nil {
			return fmt.Errorf("failed to update benchmark data: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	return nil
}

// Delete deletes a benchmark data entry by ID
func (r *BenchmarkRepository) Delete(id int64) error {
	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `DELETE FROM benchmark_data WHERE id = ?`
	case "postgres":
		query = `DELETE FROM benchmark_data WHERE id = $1`
	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete benchmark data: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all benchmark data for a specific user
func (r *BenchmarkRepository) DeleteByUserID(userID int64) error {
	var query string
	switch r.driver {
	case "sqlite3", "mysql":
		query = `DELETE FROM benchmark_data WHERE created_by = ?`
	case "postgres":
		query = `DELETE FROM benchmark_data WHERE created_by = $1`
	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete benchmark data by user ID: %w", err)
	}

	return nil
}

// DeleteAll deletes all benchmark data (admin only)
func (r *BenchmarkRepository) DeleteAll() error {
	_, err := r.db.Exec("DELETE FROM benchmark_data")
	if err != nil {
		return fmt.Errorf("failed to delete all benchmark data: %w", err)
	}
	return nil
}

// Count returns the total number of benchmark data entries
func (r *BenchmarkRepository) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM benchmark_data").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count benchmark data: %w", err)
	}
	return count, nil
}

// Driver returns the database driver name
func (r *BenchmarkRepository) Driver() string {
	return r.driver
}

// DatabaseVersion returns the database server version
func (r *BenchmarkRepository) DatabaseVersion() string {
	var version string
	var query string

	switch r.driver {
	case "sqlite3":
		query = "SELECT sqlite_version()"
	case "postgres":
		query = "SELECT version()"
	case "mysql":
		query = "SELECT version()"
	default:
		return "unknown"
	}

	err := r.db.QueryRow(query).Scan(&version)
	if err != nil {
		return "unknown"
	}

	// For PostgreSQL, the version string is verbose, extract just the version
	if r.driver == "postgres" && len(version) > 0 {
		// PostgreSQL returns something like "PostgreSQL 16.2 on x86_64-pc-linux-gnu..."
		// Extract just the version number
		parts := strings.Fields(version)
		if len(parts) >= 2 {
			return parts[0] + " " + parts[1]
		}
	}

	return version
}

// rebindQueryBenchmark converts ? placeholders to $1, $2, etc. for PostgreSQL
// This is a helper function for complex queries
func rebindQueryBenchmark(query string, driver string) string {
	if driver != "postgres" {
		return query
	}

	argIndex := 1
	var result strings.Builder
	for _, c := range query {
		if c == '?' {
			result.WriteString(fmt.Sprintf("$%d", argIndex))
			argIndex++
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}
