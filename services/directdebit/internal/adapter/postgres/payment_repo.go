package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
)

// PaymentRepository implements direct debit payment persistence.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a new payment repository.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `INSERT INTO direct_debit_payments (
		id, contract_id, merchant_id, pay_id, payment_no, amount, currency, status,
		product_name, product_detail, payment_provider, webhook_url,
		gross_amount_usdt, exchange_fee_pct, exchange_fee_usdt,
		platform_fee_pct, platform_fee_usdt, total_fees_usdt, net_amount_usdt,
		customer_first_name, customer_last_name, customer_email, customer_phone, customer_address,
		created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.ContractID, p.MerchantID, p.PayID, p.PaymentNo, p.Amount, p.Currency, p.Status,
		p.ProductName, p.ProductDetail, p.PaymentProvider, p.WebhookURL,
		p.GrossAmountUSDT, p.ExchangeFeePct, p.ExchangeFeeUSDT,
		p.PlatformFeePct, p.PlatformFeeUSDT, p.TotalFeesUSDT, p.NetAmountUSDT,
		p.CustomerFirstName, p.CustomerLastName, p.CustomerEmail, p.CustomerPhone, p.CustomerAddress,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `SELECT id, contract_id, merchant_id, pay_id, payment_no, amount, currency, status,
		product_name, product_detail, payment_provider, webhook_url,
		gross_amount_usdt, exchange_fee_pct, exchange_fee_usdt,
		platform_fee_pct, platform_fee_usdt, total_fees_usdt, net_amount_usdt,
		customer_first_name, customer_last_name, customer_email, customer_phone, customer_address,
		created_at, updated_at
		FROM direct_debit_payments WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying payment: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrPaymentNotFound
	}
	return scanPayment(rows)
}

func (r *PaymentRepository) ListByContract(ctx context.Context, contractID uuid.UUID) ([]*domain.Payment, error) {
	query := `SELECT id, contract_id, merchant_id, pay_id, payment_no, amount, currency, status,
		product_name, product_detail, payment_provider, webhook_url,
		gross_amount_usdt, exchange_fee_pct, exchange_fee_usdt,
		platform_fee_pct, platform_fee_usdt, total_fees_usdt, net_amount_usdt,
		customer_first_name, customer_last_name, customer_email, customer_phone, customer_address,
		created_at, updated_at
		FROM direct_debit_payments WHERE contract_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, contractID)
	if err != nil {
		return nil, fmt.Errorf("listing payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	query := `UPDATE direct_debit_payments SET status = $2, pay_id = $3, updated_at = $4 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, p.ID, p.Status, p.PayID, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating payment: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func scanPayment(rows pgx.Rows) (*domain.Payment, error) {
	var p domain.Payment
	err := rows.Scan(
		&p.ID, &p.ContractID, &p.MerchantID, &p.PayID, &p.PaymentNo, &p.Amount, &p.Currency, &p.Status,
		&p.ProductName, &p.ProductDetail, &p.PaymentProvider, &p.WebhookURL,
		&p.GrossAmountUSDT, &p.ExchangeFeePct, &p.ExchangeFeeUSDT,
		&p.PlatformFeePct, &p.PlatformFeeUSDT, &p.TotalFeesUSDT, &p.NetAmountUSDT,
		&p.CustomerFirstName, &p.CustomerLastName, &p.CustomerEmail, &p.CustomerPhone, &p.CustomerAddress,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning payment: %w", err)
	}
	return &p, nil
}
