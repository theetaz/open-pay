package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
)

// RefundRepository implements refund persistence.
type RefundRepository struct {
	pool *pgxpool.Pool
}

// NewRefundRepository creates a new refund repository.
func NewRefundRepository(pool *pgxpool.Pool) *RefundRepository {
	return &RefundRepository{pool: pool}
}

func (r *RefundRepository) Create(ctx context.Context, ref *domain.Refund) error {
	query := `INSERT INTO refunds (id, merchant_id, payment_id, payment_no, amount_usdt, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query,
		ref.ID, ref.MerchantID, ref.PaymentID, ref.PaymentNo, ref.AmountUSDT, ref.Reason, string(ref.Status), ref.CreatedAt, ref.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting refund: %w", err)
	}
	return nil
}

func (r *RefundRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Refund, error) {
	query := `SELECT id, merchant_id, payment_id, payment_no, amount_usdt, reason, status,
		approved_by, approved_at, rejected_reason, completed_at, created_at, updated_at
		FROM refunds WHERE id = $1`
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying refund: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, domain.ErrRefundNotFound
	}
	return scanRefund(rows)
}

func (r *RefundRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Refund, error) {
	query := `SELECT id, merchant_id, payment_id, payment_no, amount_usdt, reason, status,
		approved_by, approved_at, rejected_reason, completed_at, created_at, updated_at
		FROM refunds WHERE merchant_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing refunds: %w", err)
	}
	defer rows.Close()

	var refunds []*domain.Refund
	for rows.Next() {
		ref, err := scanRefund(rows)
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, nil
}

func (r *RefundRepository) Update(ctx context.Context, ref *domain.Refund) error {
	query := `UPDATE refunds SET status = $2, approved_by = $3, approved_at = $4,
		rejected_reason = $5, completed_at = $6, updated_at = $7 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query,
		ref.ID, string(ref.Status), ref.ApprovedBy, ref.ApprovedAt,
		ref.RejectedReason, ref.CompletedAt, ref.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating refund: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRefundNotFound
	}
	return nil
}

func scanRefund(rows pgx.Rows) (*domain.Refund, error) {
	var ref domain.Refund
	var status string
	err := rows.Scan(
		&ref.ID, &ref.MerchantID, &ref.PaymentID, &ref.PaymentNo, &ref.AmountUSDT, &ref.Reason, &status,
		&ref.ApprovedBy, &ref.ApprovedAt, &ref.RejectedReason, &ref.CompletedAt, &ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning refund: %w", err)
	}
	ref.Status = domain.RefundStatus(status)
	return &ref, nil
}
