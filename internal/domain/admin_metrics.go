package domain

// AdminMetrics holds all dashboard statistics for the admin metrics dashboard
type AdminMetrics struct {
	UserStats         UserStats         `json:"user_stats"`
	WorkoutStats      WorkoutStats      `json:"workout_stats"`
	ContentStats      ContentStats      `json:"content_stats"`
	SubscriptionStats SubscriptionStats `json:"subscription_stats"`
	SystemHealth      SystemHealth      `json:"system_health"`
	WorkoutTrends     []WorkoutTrend    `json:"workout_trends"`
}

// UserStats contains user-related metrics
type UserStats struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveThisMonth int64 `json:"active_this_month"`
	NewThisMonth    int64 `json:"new_this_month"`
	DisabledUsers   int64 `json:"disabled_users"`
}

// WorkoutStats contains workout-related metrics
type WorkoutStats struct {
	TotalWorkouts      int64   `json:"total_workouts"`
	WorkoutsThisMonth  int64   `json:"workouts_this_month"`
	AvgWorkoutsPerUser float64 `json:"avg_workouts_per_user"`
}

// ContentStats contains content-related metrics (movements, WODs, templates)
type ContentStats struct {
	TotalMovements        int64 `json:"total_movements"`
	TotalWODs             int64 `json:"total_wods"`
	TotalWorkoutTemplates int64 `json:"total_workout_templates"`
	UserCreatedMovements  int64 `json:"user_created_movements"`
	UserCreatedWODs       int64 `json:"user_created_wods"`
	UserCreatedTemplates  int64 `json:"user_created_templates"`
}

// SubscriptionStats contains subscription-related metrics
type SubscriptionStats struct {
	ActiveSubscriptions  int64 `json:"active_subscriptions"`
	ExpiredSubscriptions int64 `json:"expired_subscriptions"`
	ExpiringSoon         int64 `json:"expiring_soon"`
}

// SystemHealth contains system health metrics
type SystemHealth struct {
	RecentAuditEvents24h int64   `json:"recent_audit_events_24h"`
	EmailSuccessRate     float64 `json:"email_success_rate"`
	TotalEmailsSent      int64   `json:"total_emails_sent"`
	FailedEmails         int64   `json:"failed_emails"`
}

// WorkoutTrend represents daily workout counts for trend visualization
type WorkoutTrend struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}
