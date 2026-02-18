package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// LeaderboardRepository implements domain.LeaderboardRepository
type LeaderboardRepository struct {
	db *sql.DB
}

// NewLeaderboardRepository creates a new leaderboard repository
func NewLeaderboardRepository(db *sql.DB) *LeaderboardRepository {
	return &LeaderboardRepository{db: db}
}

// GetMovementLeaderboard returns ranked entries for a movement, sorted by max weight DESC
func (r *LeaderboardRepository) GetMovementLeaderboard(movementID int64, orgIDs []int64, startDate, endDate *time.Time, limit, offset int) ([]*domain.LeaderboardEntry, int, error) {
	if len(orgIDs) == 0 {
		return []*domain.LeaderboardEntry{}, 0, nil
	}

	// Build org ID placeholders
	orgPlaceholders, orgArgs := buildINClause(orgIDs)

	// Build date filter
	dateFilter, dateArgs := buildDateFilter(startDate, endDate)

	args := []interface{}{movementID}
	args = append(args, orgArgs...)
	args = append(args, dateArgs...)

	baseQuery := fmt.Sprintf(`
		FROM user_workout_movements uwm
		JOIN user_workouts uw ON uwm.user_workout_id = uw.id
		JOIN users u ON uw.user_id = u.id
		JOIN user_settings us ON us.user_id = u.id AND us.leaderboard_opt_in = %s
		JOIN user_organizations uo ON uo.user_id = u.id AND uo.organization_id IN (%s)
		WHERE uwm.movement_id = ?
		AND u.account_disabled = %s
		%s
	`, getBoolValue(currentDriver, true), orgPlaceholders, getBoolValue(currentDriver, false), dateFilter)

	// Count query
	countQuery := rebindQuery(`SELECT COUNT(DISTINCT uw.user_id) ` + baseQuery)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count leaderboard entries: %w", err)
	}

	// Main query — best weight per user
	selectQuery := rebindQuery(fmt.Sprintf(`
		SELECT u.id, u.name, COALESCE(u.profile_image, ''),
		       MAX(uwm.weight) as best_weight,
		       uw.workout_date
		%s
		GROUP BY u.id, u.name, u.profile_image, uw.workout_date
		ORDER BY best_weight DESC
		LIMIT ? OFFSET ?
	`, baseQuery))

	args = append(args, limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query movement leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*domain.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		entry := &domain.LeaderboardEntry{}
		if err := rows.Scan(&entry.UserID, &entry.UserName, &entry.ProfileImage, &entry.Score, &entry.WorkoutDate); err != nil {
			return nil, 0, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entry.Rank = rank
		entry.ScoreDisplay = fmt.Sprintf("%.1f", entry.Score)
		entries = append(entries, entry)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("leaderboard rows error: %w", err)
	}

	return entries, total, nil
}

// GetWODLeaderboard returns ranked entries for a WOD, sorted by score type
func (r *LeaderboardRepository) GetWODLeaderboard(wodID int64, scoreType string, orgIDs []int64, startDate, endDate *time.Time, limit, offset int) ([]*domain.LeaderboardEntry, int, error) {
	if len(orgIDs) == 0 {
		return []*domain.LeaderboardEntry{}, 0, nil
	}

	orgPlaceholders, orgArgs := buildINClause(orgIDs)
	dateFilter, dateArgs := buildDateFilter(startDate, endDate)

	args := []interface{}{wodID}
	args = append(args, orgArgs...)
	args = append(args, dateArgs...)

	baseQuery := fmt.Sprintf(`
		FROM user_workout_wods uww
		JOIN user_workouts uw ON uww.user_workout_id = uw.id
		JOIN users u ON uw.user_id = u.id
		JOIN user_settings us ON us.user_id = u.id AND us.leaderboard_opt_in = %s
		JOIN user_organizations uo ON uo.user_id = u.id AND uo.organization_id IN (%s)
		WHERE uww.wod_id = ?
		AND u.account_disabled = %s
		%s
	`, getBoolValue(currentDriver, true), orgPlaceholders, getBoolValue(currentDriver, false), dateFilter)

	// Count
	countQuery := rebindQuery(`SELECT COUNT(DISTINCT uw.user_id) ` + baseQuery)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count WOD leaderboard entries: %w", err)
	}

	// Determine aggregate and sort based on score type
	var scoreExpr, orderDir string
	switch scoreType {
	case "Time":
		scoreExpr = "MIN(uww.time_seconds)"
		orderDir = "ASC"
	case "Rounds+Reps":
		scoreExpr = "MAX(uww.rounds * 1000 + COALESCE(uww.reps, 0))"
		orderDir = "DESC"
	default: // "Max Weight" or other
		scoreExpr = "MAX(uww.weight)"
		orderDir = "DESC"
	}

	selectQuery := rebindQuery(fmt.Sprintf(`
		SELECT u.id, u.name, COALESCE(u.profile_image, ''),
		       %s as best_score,
		       uw.workout_date
		%s
		GROUP BY u.id, u.name, u.profile_image, uw.workout_date
		ORDER BY best_score %s
		LIMIT ? OFFSET ?
	`, scoreExpr, baseQuery, orderDir))

	args = append(args, limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query WOD leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*domain.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		entry := &domain.LeaderboardEntry{}
		if err := rows.Scan(&entry.UserID, &entry.UserName, &entry.ProfileImage, &entry.Score, &entry.WorkoutDate); err != nil {
			return nil, 0, fmt.Errorf("failed to scan WOD leaderboard entry: %w", err)
		}
		entry.Rank = rank
		entry.ScoreDisplay = formatWODScore(scoreType, entry.Score)
		entries = append(entries, entry)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("WOD leaderboard rows error: %w", err)
	}

	return entries, total, nil
}

// buildINClause builds a SQL IN clause for a slice of int64 IDs
func buildINClause(ids []int64) (string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return strings.Join(placeholders, ", "), args
}

// buildDateFilter builds optional date range WHERE clauses
func buildDateFilter(startDate, endDate *time.Time) (string, []interface{}) {
	var clauses []string
	var args []interface{}
	if startDate != nil {
		clauses = append(clauses, "AND uw.workout_date >= ?")
		args = append(args, *startDate)
	}
	if endDate != nil {
		clauses = append(clauses, "AND uw.workout_date <= ?")
		args = append(args, *endDate)
	}
	return strings.Join(clauses, " "), args
}

// formatWODScore formats a WOD score for display based on score type
func formatWODScore(scoreType string, score float64) string {
	switch scoreType {
	case "Time":
		totalSec := int(score)
		minutes := totalSec / 60
		seconds := totalSec % 60
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	case "Rounds+Reps":
		rounds := int(score) / 1000
		reps := int(score) % 1000
		if reps > 0 {
			return fmt.Sprintf("%d+%d", rounds, reps)
		}
		return fmt.Sprintf("%d rounds", rounds)
	default:
		return fmt.Sprintf("%.1f", score)
	}
}
