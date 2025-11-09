package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteMovementRepository implements MovementRepository using SQLite
type SQLiteMovementRepository struct {
	db *sql.DB
}

// NewSQLiteMovementRepository creates a new SQLite movement repository
func NewSQLiteMovementRepository(db *sql.DB) *SQLiteMovementRepository {
	return &SQLiteMovementRepository{db: db}
}

// Create creates a new movement
func (r *SQLiteMovementRepository) Create(movement *domain.Movement) error {
	query := `
		INSERT INTO movements (name, description, type, is_standard, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	movement.CreatedAt = now
	movement.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		movement.Name,
		movement.Description,
		movement.Type,
		movement.IsStandard,
		movement.CreatedBy,
		movement.CreatedAt,
		movement.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	movement.ID = id
	return nil
}

// GetByID retrieves a movement by ID
func (r *SQLiteMovementRepository) GetByID(id int64) (*domain.Movement, error) {
	query := `
		SELECT id, name, description, type, is_standard, created_by,
		       created_at, updated_at
		FROM movements
		WHERE id = ?
	`

	movement := &domain.Movement{}
	var description sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&movement.ID,
		&movement.Name,
		&movement.Description,
		&description,
		&movement.Type,
		&movement.IsStandard,
		&createdBy,
		&movement.CreatedAt,
		&movement.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if description.Valid {
		movement.Description = description.String
	}
	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// GetByName retrieves a movement by name
func (r *SQLiteMovementRepository) GetByName(name string) (*domain.Movement, error) {
	query := `
		SELECT id, name, description, type, is_standard, created_by,
		       created_at, updated_at
		FROM movements
		WHERE name = ?
	`

	movement := &domain.Movement{}
	var description sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(
		&movement.ID,
		&movement.Name,
		&movement.Description,
		&description,
		&movement.Type,
		&movement.IsStandard,
		&createdBy,
		&movement.CreatedAt,
		&movement.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if description.Valid {
		movement.Description = description.String
	}
	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// ListStandard retrieves all standard movements
func (r *SQLiteMovementRepository) ListStandard() ([]*domain.Movement, error) {
	query := `
		SELECT id, name, description, type, is_standard, created_by,
		       created_at, updated_at
		FROM movements
		WHERE is_standard = 1
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*domain.Movement
	for rows.Next() {
		movement := &domain.Movement{}
		var createdBy sql.NullInt64

		err := rows.Scan(
			&movement.ID,
			&movement.Name,
			&movement.Description,
			&movement.Type,
			&movement.IsStandard,
			&createdBy,
			&movement.CreatedAt,
			&movement.UpdatedAt,
		)
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

// ListByUser retrieves movements created by a user
func (r *SQLiteMovementRepository) ListByUser(userID int64) ([]*domain.Movement, error) {
	query := `
		SELECT id, name, description, type, is_standard, created_by,
		       created_at, updated_at
		FROM movements
		WHERE created_by = ?
		ORDER BY name
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*domain.Movement
	for rows.Next() {
		movement := &domain.Movement{}
		var createdBy sql.NullInt64

		err := rows.Scan(
			&movement.ID,
			&movement.Name,
			&movement.Description,
			&movement.Type,
			&movement.IsStandard,
			&createdBy,
			&movement.CreatedAt,
			&movement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if createdBy.Valid {
			movement.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, movement)
	}

	return movements, rows.Err()
	return r.queryMovementsWithParam(query, userID)
}

// Update updates a movement
func (r *SQLiteMovementRepository) Update(movement *domain.Movement) error {
	query := `
		UPDATE movements
		SET name = ?, description = ?, type = ?, updated_at = ?
		WHERE id = ?
	`

	movement.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		movement.Name,
		movement.Description,
		movement.Type,
		movement.UpdatedAt,
		movement.ID,
	)

	return err
}

// Delete deletes a movement
func (r *SQLiteMovementRepository) Delete(id int64) error {
	query := `DELETE FROM movements WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// Search searches for movements by name
func (r *SQLiteMovementRepository) Search(query string, limit int) ([]*domain.Movement, error) {
	sqlQuery := `
		SELECT id, name, description, type, is_standard, created_by,
		       created_at, updated_at
		FROM movements
		WHERE name LIKE ?
		ORDER BY is_standard DESC, name
		LIMIT ?
	`

	return r.queryMovementsWithParams(sqlQuery, "%"+query+"%", limit)
}

// Helper methods

func (r *SQLiteMovementRepository) queryMovements(query string) ([]*domain.Movement, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

func (r *SQLiteMovementRepository) queryMovementsWithParam(query string, arg interface{}) ([]*domain.Movement, error) {
	rows, err := r.db.Query(query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

func (r *SQLiteMovementRepository) queryMovementsWithParams(query string, args ...interface{}) ([]*domain.Movement, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

func (r *SQLiteMovementRepository) scanMovements(rows *sql.Rows) ([]*domain.Movement, error) {
	var movements []*domain.Movement
	for rows.Next() {
		movement := &domain.Movement{}
		var description sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&movement.ID,
			&movement.Name,
			&movement.Description,
			&description,
			&movement.Type,
			&movement.IsStandard,
			&createdBy,
			&movement.CreatedAt,
			&movement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			movement.Description = description.String
		}
		if createdBy.Valid {
			movement.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, movement)
	}

	return movements, rows.Err()
}
