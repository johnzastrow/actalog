package domain

import "time"

// LeaderboardEntry represents a single user's ranking in a leaderboard
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name"`
	ProfileImage string    `json:"profile_image,omitempty"`
	Score        float64   `json:"score"`
	ScoreDisplay string    `json:"score_display"`
	WorkoutDate  time.Time `json:"workout_date"`
	IsCurrentUser bool    `json:"is_current_user"`
}

// LeaderboardResult represents a complete leaderboard response
type LeaderboardResult struct {
	Type       string             `json:"type"`        // "movement" or "wod"
	ItemID     int64              `json:"item_id"`
	ItemName   string             `json:"item_name"`
	ScoreType  string             `json:"score_type"`  // "weight", "time", "rounds_reps"
	Period     string             `json:"period"`       // "all_time", "this_year", "this_month"
	Entries    []*LeaderboardEntry `json:"entries"`
	Total      int                `json:"total"`
}

// LeaderboardRepository defines the interface for leaderboard data access
type LeaderboardRepository interface {
	// GetMovementLeaderboard returns ranked entries for a movement across org members
	GetMovementLeaderboard(movementID int64, orgIDs []int64, startDate, endDate *time.Time, limit, offset int) ([]*LeaderboardEntry, int, error)

	// GetWODLeaderboard returns ranked entries for a WOD across org members
	GetWODLeaderboard(wodID int64, scoreType string, orgIDs []int64, startDate, endDate *time.Time, limit, offset int) ([]*LeaderboardEntry, int, error)
}
