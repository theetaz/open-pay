package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
)

// BalanceRepository is the PostgreSQL implementation for balance persistence.
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// NewBalanceRepository creates a new PostgreSQL-backed BalanceRepository.
func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

func (r *BalanceRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*domain.MerchantBalance, error) {
	query := `SELECT id, merchant_id, available_usdt, pending_usdt,
		total_earned_usdt, total_withdrawn_usdt, total_fees_usdt,
		total_earned_lkr, total_withdrawn_lkr, updated_at
		FROM merchant_balances WHERE merchant_id = $1`

	var b domain.MerchantBalance
	err := r.pool.QueryRow(ctx, query, merchantID).Scan(
		&b.ID, &b.MerchantID, &b.AvailableUSDT, &b.PendingUSDT,
		&b.TotalEarnedUSDT, &b.TotalWithdrawnUSDT, &b.TotalFeesUSDT,
		&b.TotalEarnedLKR, &b.TotalWithdrawnLKR, &b.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrBalanceNotFound
	}
	return &b, nil
}

func (r *BalanceRepository) Create(ctx context.Context, b *domain.MerchantBalance) error {
	query := `INSERT INTO merchant_balances (id, merchant_id, available_usdt, pending_usdt,
		total_earned_usdt, total_withdrawn_usdt, total_fees_usdt,
		total_earned_lkr, total_withdrawn_lkr, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		b.ID, b.MerchantID, b.AvailableUSDT, b.PendingUSDT,
		b.TotalEarnedUSDT, b.TotalWithdrawnUSDT, b.TotalFeesUSDT,
		b.TotalEarnedLKR, b.TotalWithdrawnLKR, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) Update(ctx context.Context, b *domain.MerchantBalance) error {
	query := `UPDATE merchant_balances SET
		available_usdt = $2, pending_usdt = $3,
		total_earned_usdt = $4, total_withdrawn_usdt = $5, total_fees_usdt = $6,
		total_earned_lkr = $7, total_withdrawn_lkr = $8, updated_at = $9
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		b.ID, b.AvailableUSDT, b.PendingUSDT,
		b.TotalEarnedUSDT, b.TotalWithdrawnUSDT, b.TotalFeesUSDT,
		b.TotalEarnedLKR, b.TotalWithdrawnLKR, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating balance: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBalanceNotFound
	}
	return nil
}
