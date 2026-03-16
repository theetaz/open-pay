package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/shopspring/decimal"
)

// PaymentLinkRepository is the PostgreSQL implementation for payment link persistence.
type PaymentLinkRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentLinkRepository creates a new PostgreSQL-backed PaymentLinkRepository.
func NewPaymentLinkRepository(pool *pgxpool.Pool) *PaymentLinkRepository {
	return &PaymentLinkRepository{pool: pool}
}

func (r *PaymentLinkRepository) Create(ctx context.Context, pl *domain.PaymentLink) error {
	query := `
		INSERT INTO payment_links (
			id, merchant_id, branch_id, name, slug, description, currency,
			amount, allow_custom_amount, is_reusable, show_on_qr_page,
			usage_count, status, expire_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16
		)`

	_, err := r.pool.Exec(ctx, query,
		pl.ID, pl.MerchantID, pl.BranchID, pl.Name, pl.Slug, pl.Description, pl.Currency,
		pl.Amount, pl.AllowCustomAmount, pl.IsReusable, pl.ShowOnQRPage,
		pl.UsageCount, string(pl.Status), pl.ExpireAt, pl.CreatedAt, pl.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateSlug
		}
		return fmt.Errorf("inserting payment link: %w", err)
	}
	return nil
}

const paymentLinkSelectCols = `id, merchant_id, branch_id, name, slug, COALESCE(description,''), currency,
	amount, allow_custom_amount, is_reusable, show_on_qr_page,
	usage_count, status, expire_at, created_at, updated_at, deleted_at`

func (r *PaymentLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentLink, error) {
	return r.getOne(ctx, "SELECT "+paymentLinkSelectCols+" FROM payment_links WHERE id = $1 AND deleted_at IS NULL", id)
}

func (r *PaymentLinkRepository) GetBySlug(ctx context.Context, merchantID uuid.UUID, slug string) (*domain.PaymentLink, error) {
	return r.getOne(ctx, "SELECT "+paymentLinkSelectCols+" FROM payment_links WHERE merchant_id = $1 AND slug = $2 AND deleted_at IS NULL", merchantID, slug)
}

func (r *PaymentLinkRepository) GetBySlugGlobal(ctx context.Context, slug string) (*domain.PaymentLink, error) {
	return r.getOne(ctx, "SELECT "+paymentLinkSelectCols+" FROM payment_links WHERE slug = $1 AND deleted_at IS NULL", slug)
}

func (r *PaymentLinkRepository) SlugExists(ctx context.Context, merchantID uuid.UUID, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM payment_links WHERE merchant_id = $1 AND slug = $2 AND deleted_at IS NULL)",
		merchantID, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking slug existence: %w", err)
	}
	return exists, nil
}

// ListParams holds pagination parameters for payment links.
type PaymentLinkListParams struct {
	MerchantID uuid.UUID
	Page       int
	PerPage    int
}

func (r *PaymentLinkRepository) List(ctx context.Context, params PaymentLinkListParams) ([]*domain.PaymentLink, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	var total int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM payment_links WHERE merchant_id = $1 AND deleted_at IS NULL",
		params.MerchantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting payment links: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	rows, err := r.pool.Query(ctx,
		"SELECT "+paymentLinkSelectCols+" FROM payment_links WHERE merchant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		params.MerchantID, params.PerPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing payment links: %w", err)
	}
	defer rows.Close()

	var links []*domain.PaymentLink
	for rows.Next() {
		pl, err := scanPaymentLink(rows)
		if err != nil {
			return nil, 0, err
		}
		links = append(links, pl)
	}

	return links, total, nil
}

func (r *PaymentLinkRepository) Update(ctx context.Context, pl *domain.PaymentLink) error {
	pl.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE payment_links SET
			name = $2, slug = $3, description = $4, currency = $5,
			amount = $6, allow_custom_amount = $7, is_reusable = $8, show_on_qr_page = $9,
			status = $10, expire_at = $11, usage_count = $12, updated_at = $13
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		pl.ID, pl.Name, pl.Slug, pl.Description, pl.Currency,
		pl.Amount, pl.AllowCustomAmount, pl.IsReusable, pl.ShowOnQRPage,
		string(pl.Status), pl.ExpireAt, pl.UsageCount, pl.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateSlug
		}
		return fmt.Errorf("updating payment link: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPaymentLinkNotFound
	}
	return nil
}

func (r *PaymentLinkRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE payment_links SET deleted_at = NOW(), status = 'INACTIVE' WHERE id = $1 AND deleted_at IS NULL"
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-deleting payment link: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPaymentLinkNotFound
	}
	return nil
}

func (r *PaymentLinkRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE payment_links SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("incrementing usage count: %w", err)
	}
	return nil
}

func (r *PaymentLinkRepository) getOne(ctx context.Context, query string, args ...any) (*domain.PaymentLink, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying payment link: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrPaymentLinkNotFound
	}
	return scanPaymentLink(rows)
}

func scanPaymentLink(rows pgx.Rows) (*domain.PaymentLink, error) {
	var pl domain.PaymentLink
	var status string
	var amount decimal.Decimal

	err := rows.Scan(
		&pl.ID, &pl.MerchantID, &pl.BranchID, &pl.Name, &pl.Slug, &pl.Description, &pl.Currency,
		&amount, &pl.AllowCustomAmount, &pl.IsReusable, &pl.ShowOnQRPage,
		&pl.UsageCount, &status, &pl.ExpireAt, &pl.CreatedAt, &pl.UpdatedAt, &pl.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning payment link: %w", err)
	}

	pl.Amount = amount
	pl.Status = domain.PaymentLinkStatus(status)
	return &pl, nil
}
