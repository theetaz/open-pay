package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
)

// ContractRepository implements contract persistence.
type ContractRepository struct {
	pool *pgxpool.Pool
}

// NewContractRepository creates a new contract repository.
func NewContractRepository(pool *pgxpool.Pool) *ContractRepository {
	return &ContractRepository{pool: pool}
}

func (r *ContractRepository) Create(ctx context.Context, c *domain.Contract) error {
	query := `INSERT INTO direct_debit_contracts (
		id, merchant_id, branch_id, merchant_contract_code, service_name, scenario_id,
		payment_provider, currency, single_upper_limit, status,
		pre_contract_id, contract_id, biz_id, open_user_id, merchant_account_no,
		qr_content, deep_link, webhook_url, return_url, cancel_url,
		periodic, contract_end_time, request_expire_time,
		payment_count, total_amount_charged, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)`

	_, err := r.pool.Exec(ctx, query,
		c.ID, c.MerchantID, c.BranchID, c.MerchantContractCode, c.ServiceName, c.ScenarioID,
		c.PaymentProvider, c.Currency, c.SingleUpperLimit, c.Status,
		c.PreContractID, c.ContractID, c.BizID, c.OpenUserID, c.MerchantAccountNo,
		c.QRContent, c.DeepLink, c.WebhookURL, c.ReturnURL, c.CancelURL,
		c.Periodic, c.ContractEndTime, c.RequestExpireTime,
		c.PaymentCount, c.TotalAmountCharged, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting contract: %w", err)
	}
	return nil
}

func (r *ContractRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	query := `SELECT id, merchant_id, branch_id, merchant_contract_code, service_name, scenario_id,
		payment_provider, currency, single_upper_limit, status,
		pre_contract_id, contract_id, biz_id, open_user_id, merchant_account_no,
		qr_content, deep_link, webhook_url, return_url, cancel_url,
		periodic, contract_end_time, request_expire_time,
		termination_way, termination_time, termination_notes,
		payment_count, total_amount_charged, last_payment_at, created_at, updated_at
		FROM direct_debit_contracts WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying contract: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrContractNotFound
	}
	return scanContract(rows)
}

func (r *ContractRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, status string, page, limit int) ([]*domain.Contract, int, error) {
	baseWhere := " WHERE merchant_id = $1"
	args := []any{merchantID}
	argIdx := 2

	if status != "" {
		baseWhere += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM direct_debit_contracts" + baseWhere
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting contracts: %w", err)
	}

	// Fetch page
	query := `SELECT id, merchant_id, branch_id, merchant_contract_code, service_name, scenario_id,
		payment_provider, currency, single_upper_limit, status,
		pre_contract_id, contract_id, biz_id, open_user_id, merchant_account_no,
		qr_content, deep_link, webhook_url, return_url, cancel_url,
		periodic, contract_end_time, request_expire_time,
		termination_way, termination_time, termination_notes,
		payment_count, total_amount_charged, last_payment_at, created_at, updated_at
		FROM direct_debit_contracts` + baseWhere + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing contracts: %w", err)
	}
	defer rows.Close()

	var contracts []*domain.Contract
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, 0, err
		}
		contracts = append(contracts, c)
	}
	return contracts, total, nil
}

func (r *ContractRepository) Update(ctx context.Context, c *domain.Contract) error {
	query := `UPDATE direct_debit_contracts SET
		status = $2, contract_id = $3, open_user_id = $4, qr_content = $5, deep_link = $6,
		termination_way = $7, termination_time = $8, termination_notes = $9,
		payment_count = $10, total_amount_charged = $11, last_payment_at = $12, updated_at = $13
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		c.ID, c.Status, c.ContractID, c.OpenUserID, c.QRContent, c.DeepLink,
		c.TerminationWay, c.TerminationTime, c.TerminationNotes,
		c.PaymentCount, c.TotalAmountCharged, c.LastPaymentAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrContractNotFound
	}
	return nil
}

func scanContract(rows pgx.Rows) (*domain.Contract, error) {
	var c domain.Contract
	err := rows.Scan(
		&c.ID, &c.MerchantID, &c.BranchID, &c.MerchantContractCode, &c.ServiceName, &c.ScenarioID,
		&c.PaymentProvider, &c.Currency, &c.SingleUpperLimit, &c.Status,
		&c.PreContractID, &c.ContractID, &c.BizID, &c.OpenUserID, &c.MerchantAccountNo,
		&c.QRContent, &c.DeepLink, &c.WebhookURL, &c.ReturnURL, &c.CancelURL,
		&c.Periodic, &c.ContractEndTime, &c.RequestExpireTime,
		&c.TerminationWay, &c.TerminationTime, &c.TerminationNotes,
		&c.PaymentCount, &c.TotalAmountCharged, &c.LastPaymentAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning contract: %w", err)
	}
	return &c, nil
}
