package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// MovementRepository implements domain.MovementRepository
// Note: After v0.4.0 migration, this accesses the 'strength_movements' table
type MovementRepository struct {
	db *sql.DB
}

// NewMovementRepository creates a new movement repository
func NewMovementRepository(db *sql.DB) *MovementRepository {
	return &MovementRepository{db: db}
}

// Create creates a new movement
func (r *MovementRepository) Create(movement *domain.Movement) error {
	movement.CreatedAt = time.Now()
	movement.UpdatedAt = time.Now()

	query := `INSERT INTO strength_movements (name, description, type, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, movement.Name, movement.Description, movement.Type, movement.IsStandard, movement.CreatedBy, movement.CreatedAt, movement.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get movement ID: %w", err)
	}

	movement.ID = id
	return nil
}

// GetByID retrieves a movement by ID
func (r *MovementRepository) GetByID(id int64) (*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM strength_movements WHERE id = ?`

	movement := &domain.Movement{}
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get movement: %w", err)
	}

	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// GetByName retrieves a movement by name
func (r *MovementRepository) GetByName(name string) (*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM strength_movements WHERE name = ?`

	movement := &domain.Movement{}
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get movement by name: %w", err)
	}

	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// ListStandard retrieves all standard movements
func (r *MovementRepository) ListStandard() ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM strength_movements WHERE is_standard = 1 ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListByUser retrieves movements created by a user
func (r *MovementRepository) ListByUser(userID int64) ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM strength_movements WHERE created_by = ? ORDER BY name`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// Update updates a movement (only for user-created movements)
func (r *MovementRepository) Update(movement *domain.Movement) error {
	movement.UpdatedAt = time.Now()

	query := `UPDATE strength_movements
	          SET name = ?, description = ?, type = ?, updated_at = ?
	          WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, movement.Name, movement.Description, movement.Type, movement.UpdatedAt, movement.ID)
	if err != nil {
		return fmt.Errorf("failed to update movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("movement not found or is a standard movement (cannot update)")
	}

	return nil
}

// Delete deletes a movement (only for user-created movements)
func (r *MovementRepository) Delete(id int64) error {
	query := `DELETE FROM strength_movements WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("movement not found or is a standard movement (cannot delete)")
	}

	return nil
}

// Search searches for movements by name
func (r *MovementRepository) Search(query string, limit int) ([]*domain.Movement, error) {
	searchQuery := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM strength_movements
	                WHERE name LIKE ?
	                ORDER BY is_standard DESC, name
	                LIMIT ?`

	rows, err := r.db.Query(searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// scanMovements scans multiple movement rows
func (r *MovementRepository) scanMovements(rows *sql.Rows) ([]*domain.Movement, error) {
	var movements []*domain.Movement
	for rows.Next() {
		movement := &domain.Movement{}
		var createdBy sql.NullInt64

		err := rows.Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if createdBy.Valid {
			movement.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, movement)
	}

	return movements, rows.Err()
}
