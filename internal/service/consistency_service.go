package service

import (
	"context"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// ConsistencyService handles consistency achievement checks and notifications
type ConsistencyService struct {
	consistencyRepo domain.ConsistencyRepository
	notificationRepo domain.NotificationRepository
	userRepo        domain.UserRepository
	orgRepo         domain.OrganizationRepository
}

// NewConsistencyService creates a new consistency service
func NewConsistencyService(
	consistencyRepo domain.ConsistencyRepository,
	notificationRepo domain.NotificationRepository,
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
) *ConsistencyService {
	return &ConsistencyService{
		consistencyRepo:  consistencyRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		orgRepo:          orgRepo,
	}
}

// CheckAllUsers runs the daily consistency check for all active users
func (s *ConsistencyService) CheckAllUsers(ctx context.Context) error {
	// Reset stale weekly achievements (older than 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := s.consistencyRepo.ResetWeeklyAchievements(weekAgo); err != nil {
		return fmt.Errorf("failed to reset weekly achievements: %w", err)
	}

	userIDs, err := s.consistencyRepo.GetAllActiveUserIDs()
	if err != nil {
		return fmt.Errorf("failed to get active user IDs: %w", err)
	}

	for _, userID := range userIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.checkUserMilestones(userID); err != nil {
			fmt.Printf("[consistency] error checking user %d: %v\n", userID, err)
		}
	}

	return nil
}

// checkUserMilestones checks all milestone types for a single user
func (s *ConsistencyService) checkUserMilestones(userID int64) error {
	// 1. Total workout milestones
	totalCount, err := s.consistencyRepo.GetTotalWorkoutCount(userID)
	if err != nil {
		return fmt.Errorf("failed to get total workout count: %w", err)
	}

	for _, milestone := range domain.TotalWorkoutMilestones {
		if totalCount >= milestone {
			if err := s.notifyIfNew(userID, domain.AchievementTypeTotalWorkouts, milestone,
				"Workout Milestone!",
				fmt.Sprintf("You've completed %d workouts! Keep up the great work.", milestone),
				domain.NotificationTypeConsistencyMilestone,
			); err != nil {
				return err
			}
		}
	}

	// 2. Weekly frequency
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	weekDates, err := s.consistencyRepo.GetWorkoutDatesInRange(userID, weekStart, now)
	if err != nil {
		return fmt.Errorf("failed to get weekly dates: %w", err)
	}

	if len(weekDates) >= domain.WeeklyFrequencyThreshold {
		if err := s.notifyIfNew(userID, domain.AchievementTypeWeeklyFrequency, len(weekDates),
			"Consistency Champion!",
			fmt.Sprintf("You've worked out %d days this week! Amazing consistency.", len(weekDates)),
			domain.NotificationTypeConsistencyMilestone,
		); err != nil {
			return err
		}
	}

	// 3. Day streak
	streak := s.calculateCurrentStreak(userID)
	for _, milestone := range domain.DayStreakMilestones {
		if streak >= milestone {
			if err := s.notifyIfNew(userID, domain.AchievementTypeDayStreak, milestone,
				"Streak Milestone!",
				fmt.Sprintf("You're on a %d-day workout streak! Don't break the chain.", milestone),
				domain.NotificationTypeConsistencyMilestone,
			); err != nil {
				return err
			}
		}
	}

	// 4. Inactivity warning
	lastDate, err := s.consistencyRepo.GetLastWorkoutDate(userID)
	if err != nil {
		return fmt.Errorf("failed to get last workout date: %w", err)
	}

	if lastDate != nil {
		daysSince := int(now.Sub(*lastDate).Hours() / 24)
		if daysSince >= 7 {
			weeksSince := daysSince / 7
			if err := s.notifyIfNew(userID, domain.AchievementTypeInactivity, weeksSince,
				"We Miss You!",
				fmt.Sprintf("It's been %d days since your last workout. Ready to get back at it?", daysSince),
				domain.NotificationTypeConsistencyInactivity,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// calculateCurrentStreak calculates how many consecutive days the user has worked out
func (s *ConsistencyService) calculateCurrentStreak(userID int64) int {
	now := time.Now()
	start := now.AddDate(0, 0, -60) // Look back 60 days
	dates, err := s.consistencyRepo.GetWorkoutDatesInRange(userID, start, now)
	if err != nil || len(dates) == 0 {
		return 0
	}

	// Build a set of dates (normalized to date only)
	dateSet := make(map[string]bool, len(dates))
	for _, d := range dates {
		dateSet[d.Format("2006-01-02")] = true
	}

	// Walk backward from today/yesterday
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	streak := 0

	// Check if today counts
	checkDate := today
	if !dateSet[checkDate.Format("2006-01-02")] {
		// Try yesterday as start
		checkDate = today.AddDate(0, 0, -1)
		if !dateSet[checkDate.Format("2006-01-02")] {
			return 0
		}
	}

	for dateSet[checkDate.Format("2006-01-02")] {
		streak++
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	return streak
}

// GetUserStats returns consistency stats for a user
func (s *ConsistencyService) GetUserStats(userID int64) (*domain.ConsistencyStats, error) {
	stats := &domain.ConsistencyStats{}

	// Total workouts
	total, err := s.consistencyRepo.GetTotalWorkoutCount(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalWorkouts = total

	// Workouts this week
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	weekDates, err := s.consistencyRepo.GetWorkoutDatesInRange(userID, weekStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly dates: %w", err)
	}
	stats.WorkoutsThisWeek = len(weekDates)

	// Current streak
	stats.CurrentStreak = s.calculateCurrentStreak(userID)

	// Longest streak (look back over all workout dates in last year)
	yearStart := now.AddDate(-1, 0, 0)
	allDates, err := s.consistencyRepo.GetWorkoutDatesInRange(userID, yearStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get all dates: %w", err)
	}
	stats.LongestStreak = calculateLongestStreak(allDates)

	// Days since last workout
	lastDate, err := s.consistencyRepo.GetLastWorkoutDate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last workout date: %w", err)
	}
	if lastDate != nil {
		stats.DaysSinceLastWorkout = int(now.Sub(*lastDate).Hours() / 24)
	} else {
		stats.DaysSinceLastWorkout = -1 // No workouts
	}

	return stats, nil
}

// notifyIfNew creates a notification if the achievement hasn't been recorded yet
func (s *ConsistencyService) notifyIfNew(userID int64, achievementType string, value int, title, message, notifType string) error {
	has, err := s.consistencyRepo.HasAchievement(userID, achievementType, value)
	if err != nil {
		return fmt.Errorf("failed to check achievement: %w", err)
	}
	if has {
		return nil
	}

	now := time.Now()

	// Create the notification
	notification := &domain.Notification{
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Data:      "{}",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Record the achievement
	achievement := &domain.ConsistencyAchievement{
		UserID:          userID,
		AchievementType: achievementType,
		AchievementValue: value,
		NotifiedAt:      now,
	}

	if err := s.consistencyRepo.RecordAchievement(achievement); err != nil {
		return fmt.Errorf("failed to record achievement: %w", err)
	}

	return nil
}

// calculateLongestStreak finds the longest consecutive day streak from a sorted list of dates
func calculateLongestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	// Normalize and deduplicate
	dateSet := make(map[string]bool, len(dates))
	var unique []string
	for _, d := range dates {
		key := d.Format("2006-01-02")
		if !dateSet[key] {
			dateSet[key] = true
			unique = append(unique, key)
		}
	}

	longest := 1
	current := 1

	for i := 1; i < len(unique); i++ {
		prev, _ := time.Parse("2006-01-02", unique[i-1])
		curr, _ := time.Parse("2006-01-02", unique[i])

		if curr.Sub(prev).Hours() == 24 {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}

	return longest
}
