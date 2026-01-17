package service

import (
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// AdminMetricsService provides admin dashboard metrics
type AdminMetricsService struct {
	userRepo        domain.UserRepository
	userWorkoutRepo *repository.UserWorkoutRepository
	movementRepo    domain.MovementRepository
	wodRepo         domain.WODRepository
	workoutRepo     domain.WorkoutRepository
	userSubRepo     domain.UserSubscriptionRepository
	auditLogRepo    domain.AuditLogRepository
	emailLogService *EmailLogService
	logger          *logger.Logger
}

// NewAdminMetricsService creates a new admin metrics service
func NewAdminMetricsService(
	userRepo domain.UserRepository,
	userWorkoutRepo *repository.UserWorkoutRepository,
	movementRepo domain.MovementRepository,
	wodRepo domain.WODRepository,
	workoutRepo domain.WorkoutRepository,
	userSubRepo domain.UserSubscriptionRepository,
	auditLogRepo domain.AuditLogRepository,
	emailLogService *EmailLogService,
	logger *logger.Logger,
) *AdminMetricsService {
	return &AdminMetricsService{
		userRepo:        userRepo,
		userWorkoutRepo: userWorkoutRepo,
		movementRepo:    movementRepo,
		wodRepo:         wodRepo,
		workoutRepo:     workoutRepo,
		userSubRepo:     userSubRepo,
		auditLogRepo:    auditLogRepo,
		emailLogService: emailLogService,
		logger:          logger,
	}
}

// GetMetrics retrieves all admin dashboard metrics
func (s *AdminMetricsService) GetMetrics() (*domain.AdminMetrics, error) {
	metrics := &domain.AdminMetrics{}

	// User stats
	userStats, err := s.getUserStats()
	if err != nil {
		s.logger.Error("Failed to get user stats error=%v", err)
		// Continue with zeros instead of failing entirely
	}
	metrics.UserStats = userStats

	// Workout stats
	workoutStats, err := s.getWorkoutStats()
	if err != nil {
		s.logger.Error("Failed to get workout stats error=%v", err)
	}
	metrics.WorkoutStats = workoutStats

	// Content stats
	contentStats, err := s.getContentStats()
	if err != nil {
		s.logger.Error("Failed to get content stats error=%v", err)
	}
	metrics.ContentStats = contentStats

	// Subscription stats
	subStats, err := s.getSubscriptionStats()
	if err != nil {
		s.logger.Error("Failed to get subscription stats error=%v", err)
	}
	metrics.SubscriptionStats = subStats

	// System health
	systemHealth, err := s.getSystemHealth()
	if err != nil {
		s.logger.Error("Failed to get system health error=%v", err)
	}
	metrics.SystemHealth = systemHealth

	// Workout trends (30 days)
	trends, err := s.userWorkoutRepo.GetWorkoutTrends(30)
	if err != nil {
		s.logger.Error("Failed to get workout trends error=%v", err)
	}
	metrics.WorkoutTrends = trends

	return metrics, nil
}

func (s *AdminMetricsService) getUserStats() (domain.UserStats, error) {
	stats := domain.UserStats{}

	totalUsers, err := s.userRepo.Count()
	if err != nil {
		return stats, err
	}
	stats.TotalUsers = totalUsers

	activeUsers, err := s.userWorkoutRepo.CountActiveUsersThisMonth()
	if err != nil {
		return stats, err
	}
	stats.ActiveThisMonth = activeUsers

	newUsers, err := s.userRepo.CountNewThisMonth()
	if err != nil {
		return stats, err
	}
	stats.NewThisMonth = newUsers

	disabledUsers, err := s.userRepo.CountDisabled()
	if err != nil {
		return stats, err
	}
	stats.DisabledUsers = disabledUsers

	return stats, nil
}

func (s *AdminMetricsService) getWorkoutStats() (domain.WorkoutStats, error) {
	stats := domain.WorkoutStats{}

	totalWorkouts, err := s.userWorkoutRepo.CountAll()
	if err != nil {
		return stats, err
	}
	stats.TotalWorkouts = totalWorkouts

	workoutsThisMonth, err := s.userWorkoutRepo.CountThisMonth()
	if err != nil {
		return stats, err
	}
	stats.WorkoutsThisMonth = workoutsThisMonth

	// Calculate average workouts per user
	totalUsers, err := s.userRepo.Count()
	if err != nil {
		return stats, err
	}
	if totalUsers > 0 {
		stats.AvgWorkoutsPerUser = float64(totalWorkouts) / float64(totalUsers)
	}

	return stats, nil
}

func (s *AdminMetricsService) getContentStats() (domain.ContentStats, error) {
	stats := domain.ContentStats{}

	// Total movements - get all and count
	movements, err := s.movementRepo.ListAll()
	if err != nil {
		return stats, err
	}
	stats.TotalMovements = int64(len(movements))

	// User-created movements
	userCreatedMovements, err := s.movementRepo.CountAllUserCreated()
	if err != nil {
		return stats, err
	}
	stats.UserCreatedMovements = userCreatedMovements

	// Total WODs - list with high limit to get count
	wods, err := s.wodRepo.List(nil, 100000, 0)
	if err != nil {
		return stats, err
	}
	stats.TotalWODs = int64(len(wods))

	// User-created WODs
	userCreatedWODs, err := s.wodRepo.CountAllUserCreated()
	if err != nil {
		return stats, err
	}
	stats.UserCreatedWODs = userCreatedWODs

	// Total workout templates - list with high limit
	templates, err := s.workoutRepo.List(nil, 100000, 0)
	if err != nil {
		return stats, err
	}
	stats.TotalWorkoutTemplates = int64(len(templates))

	// User-created templates
	userCreatedTemplates, err := s.workoutRepo.CountAllUserCreated()
	if err != nil {
		return stats, err
	}
	stats.UserCreatedTemplates = userCreatedTemplates

	return stats, nil
}

func (s *AdminMetricsService) getSubscriptionStats() (domain.SubscriptionStats, error) {
	stats := domain.SubscriptionStats{}

	activeCount, err := s.userSubRepo.CountActive()
	if err != nil {
		return stats, err
	}
	stats.ActiveSubscriptions = activeCount

	expiredCount, err := s.userSubRepo.CountExpired()
	if err != nil {
		return stats, err
	}
	stats.ExpiredSubscriptions = expiredCount

	expiringSoonCount, err := s.userSubRepo.CountExpiringSoon(7)
	if err != nil {
		return stats, err
	}
	stats.ExpiringSoon = expiringSoonCount

	return stats, nil
}

func (s *AdminMetricsService) getSystemHealth() (domain.SystemHealth, error) {
	health := domain.SystemHealth{}

	// Audit events count (no time filter - gets all)
	filters := domain.AuditLogFilters{}
	auditCount, err := s.auditLogRepo.Count(filters)
	if err != nil {
		return health, err
	}
	health.RecentAuditEvents24h = int64(auditCount)

	// Email stats from email log service
	if s.emailLogService != nil {
		emailStats, err := s.emailLogService.GetEmailStats()
		if err == nil && emailStats != nil {
			health.TotalEmailsSent = int64(emailStats.TotalCount)
			health.EmailSuccessRate = emailStats.SuccessRate
			health.FailedEmails = int64(emailStats.FailedCount)
		}
	}

	return health, nil
}
