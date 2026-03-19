package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
	"github.com/shopspring/decimal"
)

// PaymentRepository is the PostgreSQL implementation for payment persistence.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a new PostgreSQL-backed PaymentRepository.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	goodsJSON, _ := json.Marshal(p.Goods)
	if p.Goods == nil {
		goodsJSON = nil
	}

	query := `
		INSERT INTO payments (
			id, merchant_id, branch_id, payment_no, merchant_trade_no,
			amount, currency, amount_usdt, exchange_rate_snapshot,
			exchange_fee_pct, exchange_fee_usdt, platform_fee_pct, platform_fee_usdt,
			total_fees_usdt, net_amount_usdt,
			provider, provider_pay_id, qr_content, checkout_link, deep_link,
			status, customer_email, customer_first_name, customer_last_name,
			customer_phone, customer_address, goods,
			webhook_url, success_url, cancel_url,
			tx_hash, block_number, wallet_address,
			lkr_amount, lkr_exchange_fee, lkr_platform_fee, lkr_total_fees, lkr_net_amount,
			expire_time, paid_at, failed_at, idempotency_key,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24,
			$25, $26, $27,
			$28, $29, $30,
			$31, $32, $33,
			$34, $35, $36, $37, $38,
			$39, $40, $41, $42,
			$43, $44
		)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.MerchantID, p.BranchID, p.PaymentNo, p.MerchantTradeNo,
		p.Amount, p.Currency, p.AmountUSDT, p.ExchangeRateSnapshot,
		p.ExchangeFeePct, p.ExchangeFeeUSDT, p.PlatformFeePct, p.PlatformFeeUSDT,
		p.TotalFeesUSDT, p.NetAmountUSDT,
		p.Provider, p.ProviderPayID, p.QRContent, p.CheckoutLink, p.DeepLink,
		string(p.Status), p.CustomerEmail, p.CustomerFirstName, p.CustomerLastName,
		p.CustomerPhone, p.CustomerAddress, goodsJSON,
		p.WebhookURL, p.SuccessURL, p.CancelURL,
		p.TxHash, p.BlockNumber, p.WalletAddress,
		p.LKRAmount, p.LKRExchangeFee, p.LKRPlatformFee, p.LKRTotalFees, p.LKRNetAmount,
		p.ExpireTime, p.PaidAt, p.FailedAt, nilIfEmpty(p.IdempotencyKey),
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}
	return nil
}

const paymentSelectCols = `id, merchant_id, branch_id, payment_no, COALESCE(merchant_trade_no,''),
	amount, currency, amount_usdt, exchange_rate_snapshot,
	COALESCE(exchange_fee_pct,0), COALESCE(exchange_fee_usdt,0), COALESCE(platform_fee_pct,0), COALESCE(platform_fee_usdt,0),
	COALESCE(total_fees_usdt,0), COALESCE(net_amount_usdt,0),
	provider, COALESCE(provider_pay_id,''), COALESCE(qr_content,''), COALESCE(checkout_link,''), COALESCE(deep_link,''),
	status, COALESCE(customer_email,''), COALESCE(customer_first_name,''), COALESCE(customer_last_name,''),
	COALESCE(customer_phone,''), COALESCE(customer_address,''), goods,
	COALESCE(webhook_url,''), COALESCE(success_url,''), COALESCE(cancel_url,''),
	COALESCE(tx_hash,''), COALESCE(block_number,0), COALESCE(wallet_address,''),
	lkr_amount, lkr_exchange_fee, lkr_platform_fee, lkr_total_fees, lkr_net_amount,
	expire_time, paid_at, failed_at, COALESCE(idempotency_key,''),
	created_at, updated_at, deleted_at`

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `SELECT ` + paymentSelectCols + ` FROM payments WHERE id = $1 AND deleted_at IS NULL`
	return r.scanOne(ctx, query, id)
}

func (r *PaymentRepository) GetByProviderPayID(ctx context.Context, providerPayID string) (*domain.Payment, error) {
	query := `SELECT ` + paymentSelectCols + ` FROM payments WHERE provider_pay_id = $1 AND deleted_at IS NULL`
	return r.scanOne(ctx, query, providerPayID)
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	query := `
		UPDATE payments SET
			status = $2, provider_pay_id = $3, qr_content = $4, checkout_link = $5, deep_link = $6,
			tx_hash = $7, block_number = $8, wallet_address = $9,
			paid_at = $10, failed_at = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		p.ID, string(p.Status), p.ProviderPayID, p.QRContent, p.CheckoutLink, p.DeepLink,
		p.TxHash, p.BlockNumber, p.WalletAddress,
		p.PaidAt, p.FailedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating payment: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func (r *PaymentRepository) ListExpired(ctx context.Context) ([]*domain.Payment, error) {
	query := `SELECT ` + paymentSelectCols + `
		FROM payments
		WHERE status IN ('INITIATED', 'USER_REVIEW')
		AND expire_time < NOW()
		AND deleted_at IS NULL
		LIMIT 100`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing expired payments: %w", err)
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

func (r *PaymentRepository) List(ctx context.Context, merchantID uuid.UUID, params service.ListParams) ([]*domain.Payment, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	// Build WHERE clause
	conditions := []string{"merchant_id = $1", "deleted_at IS NULL"}
	args := []any{merchantID}
	argIdx := 2

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*params.Status))
		argIdx++
	}

	if params.BranchID != nil {
		conditions = append(conditions, fmt.Sprintf("branch_id = $%d", argIdx))
		args = append(args, *params.BranchID)
		argIdx++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(payment_no ILIKE $%d OR merchant_trade_no ILIKE $%d OR customer_email ILIKE $%d)",
			argIdx, argIdx, argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	if params.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *params.DateFrom)
		argIdx++
	}

	if params.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *params.DateTo)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count
	countQuery := "SELECT COUNT(*) FROM payments WHERE " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting payments: %w", err)
	}

	// Fetch
	offset := (params.Page - 1) * params.PerPage
	selectQuery := fmt.Sprintf(`SELECT %s FROM payments WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		paymentSelectCols, where, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, 0, err
		}
		payments = append(payments, p)
	}

	return payments, total, nil
}

func (r *PaymentRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Payment, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying payment: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrPaymentNotFound
	}

	return scanPayment(rows)
}

func scanPayment(rows pgx.Rows) (*domain.Payment, error) {
	var p domain.Payment
	var status string
	var goodsJSON []byte

	err := rows.Scan(
		&p.ID, &p.MerchantID, &p.BranchID, &p.PaymentNo, &p.MerchantTradeNo,
		&p.Amount, &p.Currency, &p.AmountUSDT, &p.ExchangeRateSnapshot,
		&p.ExchangeFeePct, &p.ExchangeFeeUSDT, &p.PlatformFeePct, &p.PlatformFeeUSDT,
		&p.TotalFeesUSDT, &p.NetAmountUSDT,
		&p.Provider, &p.ProviderPayID, &p.QRContent, &p.CheckoutLink, &p.DeepLink,
		&status, &p.CustomerEmail, &p.CustomerFirstName, &p.CustomerLastName,
		&p.CustomerPhone, &p.CustomerAddress, &goodsJSON,
		&p.WebhookURL, &p.SuccessURL, &p.CancelURL,
		&p.TxHash, &p.BlockNumber, &p.WalletAddress,
		&p.LKRAmount, &p.LKRExchangeFee, &p.LKRPlatformFee, &p.LKRTotalFees, &p.LKRNetAmount,
		&p.ExpireTime, &p.PaidAt, &p.FailedAt, &p.IdempotencyKey,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning payment: %w", err)
	}

	p.Status = domain.PaymentStatus(status)

	if len(goodsJSON) > 0 {
		_ = json.Unmarshal(goodsJSON, &p.Goods)
	}

	return &p, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// suppress unused import
var _ = decimal.Zero
