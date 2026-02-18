package domain

import "time"

// Achievement type constants
const (
	AchievementTypeTotalWorkouts   = "total_workouts"
	AchievementTypeWeeklyFrequency = "weekly_frequency"
	AchievementTypeDayStreak       = "day_streak"
	AchievementTypeInactivity      = "inactivity_warning"
)

// Milestone thresholds
var (
	TotalWorkoutMilestones  = []int{50, 100, 200, 300}
	DayStreakMilestones     = []int{4, 7, 14, 21, 30}
	WeeklyFrequencyThreshold = 4
)

// ConsistencyAchievement records that a user was notified about an achievement
type ConsistencyAchievement struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	AchievementType  string    `json:"achievement_type"`
	AchievementValue int       `json:"achievement_value"`
	NotifiedAt       time.Time `json:"notified_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// ConsistencyStats holds a user's current consistency metrics
type ConsistencyStats struct {
	TotalWorkouts      int `json:"total_workouts"`
	WorkoutsThisWeek   int `json:"workouts_this_week"`
	CurrentStreak      int `json:"current_streak"`
	LongestStreak      int `json:"longest_streak"`
	DaysSinceLastWorkout int `json:"days_since_last_workout"`
}

// ConsistencyRepository defines the interface for consistency achievement data access
type ConsistencyRepository interface {
	HasAchievement(userID int64, achievementType string, achievementValue int) (bool, error)
	RecordAchievement(achievement *ConsistencyAchievement) error
	GetTotalWorkoutCount(userID int64) (int, error)
	GetWorkoutDatesInRange(userID int64, start, end time.Time) ([]time.Time, error)
	GetLastWorkoutDate(userID int64) (*time.Time, error)
	GetAllActiveUserIDs() ([]int64, error)
	ResetWeeklyAchievements(olderThan time.Time) error
}
