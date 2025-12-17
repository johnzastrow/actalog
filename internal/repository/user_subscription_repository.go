package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteUserSubscriptionRepository implements UserSubscriptionRepository
type SQLiteUserSubscriptionRepository struct {
	db *sql.DB
}

// NewSQLiteUserSubscriptionRepository creates a new user subscription repository
func NewSQLiteUserSubscriptionRepository(db *sql.DB) *SQLiteUserSubscriptionRepository {
	return &SQLiteUserSubscriptionRepository{db: db}
}

// Create creates a new user subscription
func (r *SQLiteUserSubscriptionRepository) Create(sub *domain.UserSubscription) error {
	query := rebindQuery(`
		INSERT INTO user_subscriptions (
			user_id, subscription_type, status, is_permanent_free,
			start_date, end_date, last_payment_date, next_billing_date,
			cancelled_at, cancelled_reason, notes, created_at, updated_at, created_by_user_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		err := r.db.QueryRow(
			query,
			sub.UserID,
			sub.SubscriptionType,
			sub.Status,
			sub.IsPermanentFree,
			sub.StartDate,
			sub.EndDate,
			sub.LastPaymentDate,
			sub.NextBillingDate,
			sub.CancelledAt,
			sub.CancelledReason,
			sub.Notes,
			sub.CreatedAt,
			sub.UpdatedAt,
			sub.CreatedByUserID,
		).Scan(&sub.ID)
		return err
	}

	result, err := r.db.Exec(
		query,
		sub.UserID,
		sub.SubscriptionType,
		sub.Status,
		sub.IsPermanentFree,
		sub.StartDate,
		sub.EndDate,
		sub.LastPaymentDate,
		sub.NextBillingDate,
		sub.CancelledAt,
		sub.CancelledReason,
		sub.Notes,
		sub.CreatedAt,
		sub.UpdatedAt,
		sub.CreatedByUserID,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	sub.ID = id
	return nil
}

// GetByID retrieves a user subscription by ID
func (r *SQLiteUserSubscriptionRepository) GetByID(id int64) (*domain.UserSubscription, error) {
	query := rebindQuery(`
		SELECT id, user_id, subscription_type, status, is_permanent_free,
		       start_date, end_date, last_payment_date, next_billing_date,
		       cancelled_at, cancelled_reason, notes, created_at, updated_at, created_by_user_id
		FROM user_subscriptions
		WHERE id = ?
	`)

	sub := &domain.UserSubscription{}
	var endDate, lastPaymentDate, nextBillingDate, cancelledAt sql.NullTime
	var cancelledReason, notes sql.NullString
	var createdByUserID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.SubscriptionType,
		&sub.Status,
		&sub.IsPermanentFree,
		&sub.StartDate,
		&endDate,
		&lastPaymentDate,
		&nextBillingDate,
		&cancelledAt,
		&cancelledReason,
		&notes,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&createdByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	if lastPaymentDate.Valid {
		sub.LastPaymentDate = &lastPaymentDate.Time
	}
	if nextBillingDate.Valid {
		sub.NextBillingDate = &nextBillingDate.Time
	}
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if cancelledReason.Valid {
		sub.CancelledReason = &cancelledReason.String
	}
	if notes.Valid {
		sub.Notes = &notes.String
	}
	if createdByUserID.Valid {
		sub.CreatedByUserID = &createdByUserID.Int64
	}

	return sub, nil
}

// GetActiveByUserID retrieves the active subscription for a user
func (r *SQLiteUserSubscriptionRepository) GetActiveByUserID(userID int64) (*domain.UserSubscription, error) {
	query := rebindQuery(`
		SELECT id, user_id, subscription_type, status, is_permanent_free,
		       start_date, end_date, last_payment_date, next_billing_date,
		       cancelled_at, cancelled_reason, notes, created_at, updated_at, created_by_user_id
		FROM user_subscriptions
		WHERE user_id = ? AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`)

	sub := &domain.UserSubscription{}
	var endDate, lastPaymentDate, nextBillingDate, cancelledAt sql.NullTime
	var cancelledReason, notes sql.NullString
	var createdByUserID sql.NullInt64

	err := r.db.QueryRow(query, userID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.SubscriptionType,
		&sub.Status,
		&sub.IsPermanentFree,
		&sub.StartDate,
		&endDate,
		&lastPaymentDate,
		&nextBillingDate,
		&cancelledAt,
		&cancelledReason,
		&notes,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&createdByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	if lastPaymentDate.Valid {
		sub.LastPaymentDate = &lastPaymentDate.Time
	}
	if nextBillingDate.Valid {
		sub.NextBillingDate = &nextBillingDate.Time
	}
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if cancelledReason.Valid {
		sub.CancelledReason = &cancelledReason.String
	}
	if notes.Valid {
		sub.Notes = &notes.String
	}
	if createdByUserID.Valid {
		sub.CreatedByUserID = &createdByUserID.Int64
	}

	return sub, nil
}

// GetByUserID retrieves all subscriptions for a user (including history)
func (r *SQLiteUserSubscriptionRepository) GetByUserID(userID int64) ([]*domain.UserSubscription, error) {
	query := rebindQuery(`
		SELECT id, user_id, subscription_type, status, is_permanent_free,
		       start_date, end_date, last_payment_date, next_billing_date,
		       cancelled_at, cancelled_reason, notes, created_at, updated_at, created_by_user_id
		FROM user_subscriptions
		WHERE user_id = ?
		ORDER BY created_at DESC
	`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*domain.UserSubscription
	for rows.Next() {
		sub := &domain.UserSubscription{}
		var endDate, lastPaymentDate, nextBillingDate, cancelledAt sql.NullTime
		var cancelledReason, notes sql.NullString
		var createdByUserID sql.NullInt64

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.SubscriptionType,
			&sub.Status,
			&sub.IsPermanentFree,
			&sub.StartDate,
			&endDate,
			&lastPaymentDate,
			&nextBillingDate,
			&cancelledAt,
			&cancelledReason,
			&notes,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&createdByUserID,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		if lastPaymentDate.Valid {
			sub.LastPaymentDate = &lastPaymentDate.Time
		}
		if nextBillingDate.Valid {
			sub.NextBillingDate = &nextBillingDate.Time
		}
		if cancelledAt.Valid {
			sub.CancelledAt = &cancelledAt.Time
		}
		if cancelledReason.Valid {
			sub.CancelledReason = &cancelledReason.String
		}
		if notes.Valid {
			sub.Notes = &notes.String
		}
		if createdByUserID.Valid {
			sub.CreatedByUserID = &createdByUserID.Int64
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// Update updates a user subscription
func (r *SQLiteUserSubscriptionRepository) Update(sub *domain.UserSubscription) error {
	query := rebindQuery(`
		UPDATE user_subscriptions
		SET subscription_type = ?, status = ?, is_permanent_free = ?,
		    start_date = ?, end_date = ?, last_payment_date = ?, next_billing_date = ?,
		    cancelled_at = ?, cancelled_reason = ?, notes = ?, updated_at = ?, created_by_user_id = ?
		WHERE id = ?
	`)

	sub.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		sub.SubscriptionType,
		sub.Status,
		sub.IsPermanentFree,
		sub.StartDate,
		sub.EndDate,
		sub.LastPaymentDate,
		sub.NextBillingDate,
		sub.CancelledAt,
		sub.CancelledReason,
		sub.Notes,
		sub.UpdatedAt,
		sub.CreatedByUserID,
		sub.ID,
	)

	return err
}

// Delete deletes a user subscription
func (r *SQLiteUserSubscriptionRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM user_subscriptions WHERE id = ?`)
	_, err := r.db.Exec(query, id)
	return err
}

// MarkAsPaid marks a subscription as paid and extends the end_date
func (r *SQLiteUserSubscriptionRepository) MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64) error {
	// First, get the subscription to determine the type
	sub, err := r.GetByID(id)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("subscription not found")
	}

	// Calculate new end_date based on subscription type
	var newEndDate time.Time
	var newNextBillingDate time.Time

	switch sub.SubscriptionType {
	case domain.SubscriptionTypeMonthly:
		// Extend by 30 days from payment date
		newEndDate = paymentDate.AddDate(0, 0, 30)
		newNextBillingDate = newEndDate
	case domain.SubscriptionTypeAnnual:
		// Extend by 365 days from payment date
		newEndDate = paymentDate.AddDate(0, 0, 365)
		newNextBillingDate = newEndDate
	case domain.SubscriptionTypeFree:
		// Free subscriptions don't need payment
		return fmt.Errorf("cannot mark free subscription as paid")
	default:
		return fmt.Errorf("unknown subscription type: %s", sub.SubscriptionType)
	}

	// Update the subscription
	query := rebindQuery(`
		UPDATE user_subscriptions
		SET status = 'active',
		    last_payment_date = ?,
		    end_date = ?,
		    next_billing_date = ?,
		    updated_at = ?,
		    created_by_user_id = ?
		WHERE id = ?
	`)

	now := time.Now()
	_, err = r.db.Exec(query, paymentDate, newEndDate, newNextBillingDate, now, adminUserID, id)
	return err
}

// MarkAsExpired marks a subscription as expired
func (r *SQLiteUserSubscriptionRepository) MarkAsExpired(id int64) error {
	query := rebindQuery(`
		UPDATE user_subscriptions
		SET status = 'expired',
		    updated_at = ?
		WHERE id = ?
	`)

	now := time.Now()
	_, err := r.db.Exec(query, now, id)
	return err
}

// Cancel cancels a subscription
func (r *SQLiteUserSubscriptionRepository) Cancel(id int64, reason string, adminUserID int64) error {
	query := rebindQuery(`
		UPDATE user_subscriptions
		SET status = 'cancelled',
		    cancelled_at = ?,
		    cancelled_reason = ?,
		    updated_at = ?,
		    created_by_user_id = ?
		WHERE id = ?
	`)

	now := time.Now()
	_, err := r.db.Exec(query, now, reason, now, adminUserID, id)
	return err
}

// ListAll returns all user subscriptions
func (r *SQLiteUserSubscriptionRepository) ListAll() ([]*domain.UserSubscription, error) {
	query := rebindQuery(`
		SELECT
			id, user_id, subscription_type, status, is_permanent_free,
			start_date, end_date, last_payment_date, next_billing_date,
			cancelled_at, cancelled_reason, notes, created_at, updated_at, created_by_user_id
		FROM user_subscriptions
		ORDER BY created_at DESC
	`)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query user subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*domain.UserSubscription
	for rows.Next() {
		var sub domain.UserSubscription
		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.SubscriptionType,
			&sub.Status,
			&sub.IsPermanentFree,
			&sub.StartDate,
			&sub.EndDate,
			&sub.LastPaymentDate,
			&sub.NextBillingDate,
			&sub.CancelledAt,
			&sub.CancelledReason,
			&sub.Notes,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&sub.CreatedByUserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user subscriptions: %w", err)
	}

	return subscriptions, nil
}
