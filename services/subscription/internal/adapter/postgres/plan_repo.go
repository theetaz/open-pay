package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/subscription/internal/domain"
)

type PlanRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

func (r *PlanRepository) Create(ctx context.Context, p *domain.SubscriptionPlan) error {
	query := `INSERT INTO subscription_plans (id, merchant_id, name, description, amount, currency,
		interval_type, interval_count, trial_days, max_subscribers, contract_address, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.MerchantID, p.Name, p.Description, p.Amount, p.Currency,
		string(p.IntervalType), p.IntervalCount, p.TrialDays, p.MaxSubscribers,
		p.ContractAddress, string(p.Status), p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionPlan, error) {
	query := `SELECT id, merchant_id, name, description, amount, currency,
		interval_type, interval_count, trial_days, max_subscribers, contract_address,
		status, created_at, updated_at, deleted_at
		FROM subscription_plans WHERE id = $1 AND deleted_at IS NULL`

	return r.scanOne(ctx, query, id)
}

func (r *PlanRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.SubscriptionPlan, error) {
	query := `SELECT id, merchant_id, name, description, amount, currency,
		interval_type, interval_count, trial_days, max_subscribers, contract_address,
		status, created_at, updated_at, deleted_at
		FROM subscription_plans WHERE merchant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}
	defer rows.Close()

	var plans []*domain.SubscriptionPlan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *PlanRepository) Update(ctx context.Context, p *domain.SubscriptionPlan) error {
	query := `UPDATE subscription_plans SET name = $2, description = $3, status = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, p.ID, p.Name, p.Description, string(p.Status), p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating plan: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPlanNotFound
	}
	return nil
}

func (r *PlanRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.SubscriptionPlan, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying plan: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrPlanNotFound
	}
	return scanPlan(rows)
}

func scanPlan(rows pgx.Rows) (*domain.SubscriptionPlan, error) {
	var p domain.SubscriptionPlan
	var intervalType, status string

	err := rows.Scan(
		&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Amount, &p.Currency,
		&intervalType, &p.IntervalCount, &p.TrialDays, &p.MaxSubscribers,
		&p.ContractAddress, &status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning plan: %w", err)
	}

	p.IntervalType = domain.IntervalType(intervalType)
	p.Status = domain.PlanStatus(status)
	return &p, nil
}
