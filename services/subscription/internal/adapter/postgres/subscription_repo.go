package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/subscription/internal/domain"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *domain.Subscription) error {
	query := `INSERT INTO subscriptions (id, plan_id, merchant_id, subscriber_email, subscriber_wallet,
		status, current_period_start, current_period_end, next_billing_date, trial_end,
		cancel_at_end, total_paid_usdt, billing_count, failed_payment_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.PlanID, s.MerchantID, s.SubscriberEmail, s.SubscriberWallet,
		string(s.Status), s.CurrentPeriodStart, s.CurrentPeriodEnd, s.NextBillingDate, s.TrialEnd,
		s.CancelAtEnd, s.TotalPaidUSDT, s.BillingCount, s.FailedPaymentCount, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `SELECT id, plan_id, merchant_id, subscriber_email, subscriber_wallet,
		status, current_period_start, current_period_end, next_billing_date, trial_end,
		cancel_at_end, cancelled_at, cancellation_reason, total_paid_usdt, billing_count,
		failed_payment_count, created_at, updated_at
		FROM subscriptions WHERE id = $1`

	return r.scanOne(ctx, query, id)
}

func (r *SubscriptionRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Subscription, error) {
	query := `SELECT id, plan_id, merchant_id, subscriber_email, subscriber_wallet,
		status, current_period_start, current_period_end, next_billing_date, trial_end,
		cancel_at_end, cancelled_at, cancellation_reason, total_paid_usdt, billing_count,
		failed_payment_count, created_at, updated_at
		FROM subscriptions WHERE merchant_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*domain.Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *domain.Subscription) error {
	query := `UPDATE subscriptions SET status = $2, current_period_start = $3, current_period_end = $4,
		next_billing_date = $5, cancel_at_end = $6, cancelled_at = $7, cancellation_reason = $8,
		total_paid_usdt = $9, billing_count = $10, failed_payment_count = $11, updated_at = $12
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		s.ID, string(s.Status), s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.NextBillingDate, s.CancelAtEnd, s.CancelledAt, s.CancellationReason,
		s.TotalPaidUSDT, s.BillingCount, s.FailedPaymentCount, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSubscriptionNotFound
	}
	return nil
}

func (r *SubscriptionRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Subscription, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying subscription: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrSubscriptionNotFound
	}
	return scanSubscription(rows)
}

func scanSubscription(rows pgx.Rows) (*domain.Subscription, error) {
	var s domain.Subscription
	var status string

	err := rows.Scan(
		&s.ID, &s.PlanID, &s.MerchantID, &s.SubscriberEmail, &s.SubscriberWallet,
		&status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.NextBillingDate, &s.TrialEnd,
		&s.CancelAtEnd, &s.CancelledAt, &s.CancellationReason, &s.TotalPaidUSDT, &s.BillingCount,
		&s.FailedPaymentCount, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning subscription: %w", err)
	}

	s.Status = domain.SubscriptionStatus(status)
	return &s, nil
}
