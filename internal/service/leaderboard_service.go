package service

import (
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
)

// LeaderboardService handles leaderboard business logic
type LeaderboardService struct {
	leaderboardRepo *repository.LeaderboardRepository
	orgRepo         domain.OrganizationRepository
	movementRepo    *repository.MovementRepository
	wodRepo         *repository.WODRepository
	settingsRepo    domain.UserSettingsRepository
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(
	leaderboardRepo *repository.LeaderboardRepository,
	orgRepo domain.OrganizationRepository,
	movementRepo *repository.MovementRepository,
	wodRepo *repository.WODRepository,
	settingsRepo domain.UserSettingsRepository,
) *LeaderboardService {
	return &LeaderboardService{
		leaderboardRepo: leaderboardRepo,
		orgRepo:         orgRepo,
		movementRepo:    movementRepo,
		wodRepo:         wodRepo,
		settingsRepo:    settingsRepo,
	}
}

// GetMovementLeaderboard returns the leaderboard for a specific movement
func (s *LeaderboardService) GetMovementLeaderboard(userID, movementID int64, period string, limit, offset int) (*domain.LeaderboardResult, error) {
	orgIDs, err := s.orgRepo.GetUserOrganizationIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgIDs) == 0 {
		return &domain.LeaderboardResult{
			Type:    "movement",
			ItemID:  movementID,
			Period:  period,
			Entries: []*domain.LeaderboardEntry{},
		}, nil
	}

	startDate, endDate := periodToDateRange(period)

	movement, err := s.movementRepo.GetByID(movementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movement: %w", err)
	}
	if movement == nil {
		return nil, fmt.Errorf("movement not found")
	}

	entries, total, err := s.leaderboardRepo.GetMovementLeaderboard(movementID, orgIDs, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get movement leaderboard: %w", err)
	}

	// Mark current user's entries
	for _, entry := range entries {
		if entry.UserID == userID {
			entry.IsCurrentUser = true
		}
	}

	return &domain.LeaderboardResult{
		Type:      "movement",
		ItemID:    movementID,
		ItemName:  movement.Name,
		ScoreType: "weight",
		Period:    period,
		Entries:   entries,
		Total:     total,
	}, nil
}

// GetWODLeaderboard returns the leaderboard for a specific WOD
func (s *LeaderboardService) GetWODLeaderboard(userID, wodID int64, period string, limit, offset int) (*domain.LeaderboardResult, error) {
	orgIDs, err := s.orgRepo.GetUserOrganizationIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgIDs) == 0 {
		return &domain.LeaderboardResult{
			Type:    "wod",
			ItemID:  wodID,
			Period:  period,
			Entries: []*domain.LeaderboardEntry{},
		}, nil
	}

	startDate, endDate := periodToDateRange(period)

	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return nil, fmt.Errorf("WOD not found")
	}

	entries, total, err := s.leaderboardRepo.GetWODLeaderboard(wodID, wod.ScoreType, orgIDs, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD leaderboard: %w", err)
	}

	for _, entry := range entries {
		if entry.UserID == userID {
			entry.IsCurrentUser = true
		}
	}

	return &domain.LeaderboardResult{
		Type:      "wod",
		ItemID:    wodID,
		ItemName:  wod.Name,
		ScoreType: wod.ScoreType,
		Period:    period,
		Entries:   entries,
		Total:     total,
	}, nil
}

// periodToDateRange converts a period string to start/end dates
func periodToDateRange(period string) (*time.Time, *time.Time) {
	now := time.Now()

	switch period {
	case "this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return &start, &now
	case "this_year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return &start, &now
	default: // "all_time"
		return nil, nil
	}
}
