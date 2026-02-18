package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// ConsistencyRepository implements domain.ConsistencyRepository
type ConsistencyRepository struct {
	db *sql.DB
}

// NewConsistencyRepository creates a new consistency repository
func NewConsistencyRepository(db *sql.DB) *ConsistencyRepository {
	return &ConsistencyRepository{db: db}
}

// HasAchievement checks if a user already has a specific achievement recorded
func (r *ConsistencyRepository) HasAchievement(userID int64, achievementType string, achievementValue int) (bool, error) {
	query := rebindQuery(`
		SELECT COUNT(*) FROM consistency_achievements
		WHERE user_id = ? AND achievement_type = ? AND achievement_value = ?
	`)
	var count int
	err := r.db.QueryRow(query, userID, achievementType, achievementValue).Scan(&count)
	return count > 0, err
}

// RecordAchievement records a new achievement
func (r *ConsistencyRepository) RecordAchievement(achievement *domain.ConsistencyAchievement) error {
	query := rebindQuery(`
		INSERT INTO consistency_achievements (user_id, achievement_type, achievement_value, notified_at)
		VALUES (?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			achievement.UserID,
			achievement.AchievementType,
			achievement.AchievementValue,
			achievement.NotifiedAt,
		).Scan(&achievement.ID)
	}

	result, err := r.db.Exec(query,
		achievement.UserID,
		achievement.AchievementType,
		achievement.AchievementValue,
		achievement.NotifiedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	achievement.ID = id
	return nil
}

// GetTotalWorkoutCount returns the total number of workouts logged by a user
func (r *ConsistencyRepository) GetTotalWorkoutCount(userID int64) (int, error) {
	query := rebindQuery(`SELECT COUNT(*) FROM user_workouts WHERE user_id = ?`)
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetWorkoutDatesInRange returns distinct workout dates for a user in a date range
func (r *ConsistencyRepository) GetWorkoutDatesInRange(userID int64, start, end time.Time) ([]time.Time, error) {
	var query string
	switch currentDriver {
	case "postgres":
		query = rebindQuery(`SELECT DISTINCT workout_date::date FROM user_workouts WHERE user_id = ? AND workout_date >= ? AND workout_date <= ? ORDER BY workout_date::date`)
	default:
		query = rebindQuery(`SELECT DISTINCT DATE(workout_date) FROM user_workouts WHERE user_id = ? AND workout_date >= ? AND workout_date <= ? ORDER BY DATE(workout_date)`)
	}

	rows, err := r.db.Query(query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout dates: %w", err)
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("failed to scan workout date: %w", err)
		}
		dates = append(dates, d)
	}
	return dates, rows.Err()
}

// GetLastWorkoutDate returns the most recent workout date for a user
func (r *ConsistencyRepository) GetLastWorkoutDate(userID int64) (*time.Time, error) {
	query := rebindQuery(`SELECT MAX(workout_date) FROM user_workouts WHERE user_id = ?`)
	var d sql.NullTime
	err := r.db.QueryRow(query, userID).Scan(&d)
	if err != nil {
		return nil, err
	}
	if !d.Valid {
		return nil, nil
	}
	return &d.Time, nil
}

// GetAllActiveUserIDs returns all non-disabled user IDs
func (r *ConsistencyRepository) GetAllActiveUserIDs() ([]int64, error) {
	query := fmt.Sprintf(`SELECT id FROM users WHERE account_disabled = %s`, getBoolValue(currentDriver, false))
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user IDs: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ResetWeeklyAchievements deletes weekly_frequency achievements older than the given time
func (r *ConsistencyRepository) ResetWeeklyAchievements(olderThan time.Time) error {
	query := rebindQuery(`DELETE FROM consistency_achievements WHERE achievement_type = ? AND notified_at < ?`)
	_, err := r.db.Exec(query, domain.AchievementTypeWeeklyFrequency, olderThan)
	return err
}
