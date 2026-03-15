package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
)

// WithdrawalRepository is the PostgreSQL implementation for withdrawal persistence.
type WithdrawalRepository struct {
	pool *pgxpool.Pool
}

// NewWithdrawalRepository creates a new PostgreSQL-backed WithdrawalRepository.
func NewWithdrawalRepository(pool *pgxpool.Pool) *WithdrawalRepository {
	return &WithdrawalRepository{pool: pool}
}

func (r *WithdrawalRepository) Create(ctx context.Context, w *domain.Withdrawal) error {
	query := `INSERT INTO withdrawals (id, merchant_id, amount_usdt, exchange_rate, amount_lkr,
		bank_name, bank_account_no, bank_account_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		w.ID, w.MerchantID, w.AmountUSDT, w.ExchangeRate, w.AmountLKR,
		w.BankName, w.BankAccountNo, w.BankAccountName, string(w.Status),
		w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting withdrawal: %w", err)
	}
	return nil
}

const withdrawalSelectCols = `id, merchant_id, amount_usdt, exchange_rate, amount_lkr,
	bank_name, bank_account_no, bank_account_name, status,
	approved_by, approved_at, COALESCE(rejected_reason,''), COALESCE(bank_reference,''), completed_at,
	created_at, updated_at`

func (r *WithdrawalRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Withdrawal, error) {
	return scanWithdrawal(r.pool.QueryRow(ctx, "SELECT "+withdrawalSelectCols+" FROM withdrawals WHERE id = $1", id))
}

func (r *WithdrawalRepository) Update(ctx context.Context, w *domain.Withdrawal) error {
	query := `UPDATE withdrawals SET status = $2, approved_by = $3, approved_at = $4,
		rejected_reason = $5, bank_reference = $6, completed_at = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		w.ID, string(w.Status), w.ApprovedBy, w.ApprovedAt,
		w.RejectedReason, w.BankReference, w.CompletedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating withdrawal: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrWithdrawalNotFound
	}
	return nil
}

func (r *WithdrawalRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, status *domain.WithdrawalStatus) ([]*domain.Withdrawal, error) {
	conditions := []string{"merchant_id = $1"}
	args := []any{merchantID}

	if status != nil {
		conditions = append(conditions, "status = $2")
		args = append(args, string(*status))
	}

	query := fmt.Sprintf(`SELECT %s FROM withdrawals WHERE %s ORDER BY created_at DESC`,
		withdrawalSelectCols, strings.Join(conditions, " AND "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*domain.Withdrawal
	for rows.Next() {
		w, err := scanWithdrawalRow(rows)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil
}

func scanWithdrawal(row pgx.Row) (*domain.Withdrawal, error) {
	var w domain.Withdrawal
	var status string
	err := row.Scan(
		&w.ID, &w.MerchantID, &w.AmountUSDT, &w.ExchangeRate, &w.AmountLKR,
		&w.BankName, &w.BankAccountNo, &w.BankAccountName, &status,
		&w.ApprovedBy, &w.ApprovedAt, &w.RejectedReason, &w.BankReference, &w.CompletedAt,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrWithdrawalNotFound
	}
	w.Status = domain.WithdrawalStatus(status)
	return &w, nil
}

func scanWithdrawalRow(rows pgx.Rows) (*domain.Withdrawal, error) {
	var w domain.Withdrawal
	var status string
	err := rows.Scan(
		&w.ID, &w.MerchantID, &w.AmountUSDT, &w.ExchangeRate, &w.AmountLKR,
		&w.BankName, &w.BankAccountNo, &w.BankAccountName, &status,
		&w.ApprovedBy, &w.ApprovedAt, &w.RejectedReason, &w.BankReference, &w.CompletedAt,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning withdrawal: %w", err)
	}
	w.Status = domain.WithdrawalStatus(status)
	return &w, nil
}
