// Package domain contains the core business entities
package domain

import (
	"time"
)

// Notification types
const (
	NotificationTypePRAchievement  = "pr_achievement"
	NotificationTypeWeeklyStreak   = "weekly_streak"
	NotificationTypeWODMilestone   = "wod_milestone"
	NotificationTypeAdminUserEvent = "admin_user_event"
	NotificationTypeWelcome        = "welcome"
)

// Notification represents a notification to a user
type Notification struct {
	ID             int64      `json:"id" db:"id"`
	UserID         int64      `json:"user_id" db:"user_id"`
	OrganizationID *int64     `json:"organization_id,omitempty" db:"organization_id"`
	Type           string     `json:"type" db:"type"`
	Title          string     `json:"title" db:"title"`
	Message        string     `json:"message" db:"message"`
	Data           string     `json:"data,omitempty" db:"data"` // JSON string
	ReadAt         *time.Time `json:"read_at,omitempty" db:"read_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationData represents the structured data stored in the notification
type NotificationData struct {
	AchievingUserID   int64   `json:"achieving_user_id,omitempty"`
	AchievingUserName string  `json:"achieving_user_name,omitempty"`
	MovementName      *string `json:"movement_name,omitempty"`
	WODName           *string `json:"wod_name,omitempty"`
	Value             *string `json:"value,omitempty"`    // Weight, time, rounds, etc.
	Count             *int    `json:"count,omitempty"`    // For streaks and milestones
	WODType           *string `json:"wod_type,omitempty"` // hero, girl, games
}

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(notification *Notification) error
	GetByID(id int64) (*Notification, error)
	GetByUserID(userID int64, limit, offset int) ([]*Notification, error)
	GetUnreadByUserID(userID int64, limit, offset int) ([]*Notification, error)
	CountUnreadByUserID(userID int64) (int64, error)
	MarkAsRead(id int64) error
	MarkAsUnread(id int64) error
	MarkAllAsReadForUser(userID int64) error
	Delete(id int64) error
	DeleteOlderThan(cutoff time.Time) (int64, error) // Returns count of deleted notifications
}
