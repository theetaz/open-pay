package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
)

// MerchantRepository is the PostgreSQL implementation for merchant persistence.
type MerchantRepository struct {
	pool *pgxpool.Pool
}

// NewMerchantRepository creates a new PostgreSQL-backed MerchantRepository.
func NewMerchantRepository(pool *pgxpool.Pool) *MerchantRepository {
	return &MerchantRepository{pool: pool}
}

func (r *MerchantRepository) Create(ctx context.Context, m *domain.Merchant) error {
	query := `
		INSERT INTO merchants (
			id, business_name, business_type, registration_no, website,
			contact_email, contact_phone, contact_name,
			address_line1, address_line2, city, district, postal_code, country,
			kyc_status, instant_access_remaining,
			default_currency, default_provider, status,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16,
			$17, $18, $19,
			$20, $21
		)`

	_, err := r.pool.Exec(ctx, query,
		m.ID, m.BusinessName, m.BusinessType, m.RegistrationNo, m.Website,
		m.ContactEmail, m.ContactPhone, m.ContactName,
		m.AddressLine1, m.AddressLine2, m.City, m.District, m.PostalCode, m.Country,
		string(m.KYCStatus), m.InstantAccessRemaining,
		m.DefaultCurrency, m.DefaultProvider, string(m.Status),
		m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateEmail
		}
		return fmt.Errorf("inserting merchant: %w", err)
	}
	return nil
}

const merchantSelectCols = `id, business_name, COALESCE(business_type,''), COALESCE(registration_no,''), COALESCE(website,''),
	contact_email, COALESCE(contact_phone,''), COALESCE(contact_name,''),
	COALESCE(address_line1,''), COALESCE(address_line2,''), COALESCE(city,''), COALESCE(district,''), COALESCE(postal_code,''), country,
	kyc_status, kyc_submitted_at, kyc_reviewed_at, COALESCE(kyc_rejection_reason,''),
	transaction_limit_usdt, daily_limit_usdt, monthly_limit_usdt,
	COALESCE(instant_access_remaining, 0),
	COALESCE(bank_name,''), COALESCE(bank_branch,''), COALESCE(bank_account_no,''), COALESCE(bank_account_name,''),
	COALESCE(default_currency,'LKR'), COALESCE(default_provider,''), status,
	COALESCE(status_reason,''), status_changed_at, COALESCE(kyc_review_notes,''),
	created_at, updated_at, deleted_at`

func (r *MerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	return r.getOne(ctx, "SELECT "+merchantSelectCols+" FROM merchants WHERE id = $1 AND deleted_at IS NULL", id)
}

func (r *MerchantRepository) GetByEmail(ctx context.Context, email string) (*domain.Merchant, error) {
	return r.getOne(ctx, "SELECT "+merchantSelectCols+" FROM merchants WHERE contact_email = $1 AND deleted_at IS NULL", email)
}

func (r *MerchantRepository) Update(ctx context.Context, m *domain.Merchant) error {
	m.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE merchants SET
			business_name = $2, business_type = $3, registration_no = $4, website = $5,
			contact_email = $6, contact_phone = $7, contact_name = $8,
			address_line1 = $9, address_line2 = $10, city = $11, district = $12,
			postal_code = $13, country = $14,
			kyc_status = $15, kyc_submitted_at = $16, kyc_reviewed_at = $17,
			kyc_rejection_reason = $18,
			bank_name = $19, bank_branch = $20, bank_account_no = $21, bank_account_name = $22,
			default_currency = $23, default_provider = $24, status = $25,
			status_reason = $26, status_changed_at = $27, kyc_review_notes = $28,
			updated_at = $29
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		m.ID, m.BusinessName, m.BusinessType, m.RegistrationNo, m.Website,
		m.ContactEmail, m.ContactPhone, m.ContactName,
		m.AddressLine1, m.AddressLine2, m.City, m.District,
		m.PostalCode, m.Country,
		string(m.KYCStatus), m.KYCSubmittedAt, m.KYCReviewedAt,
		m.KYCRejectionReason,
		m.BankName, m.BankBranch, m.BankAccountNo, m.BankAccountName,
		m.DefaultCurrency, m.DefaultProvider, string(m.Status),
		m.StatusReason, m.StatusChangedAt, m.KYCReviewNotes,
		m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating merchant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMerchantNotFound
	}
	return nil
}

func (r *MerchantRepository) List(ctx context.Context, params service.ListParams) ([]*domain.Merchant, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	countQuery := "SELECT COUNT(*) FROM merchants WHERE deleted_at IS NULL"
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting merchants: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	query := "SELECT " + merchantSelectCols + " FROM merchants WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	rows, err := r.pool.Query(ctx, query, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing merchants: %w", err)
	}
	defer rows.Close()

	var merchants []*domain.Merchant
	for rows.Next() {
		m, err := scanMerchant(rows)
		if err != nil {
			return nil, 0, err
		}
		merchants = append(merchants, m)
	}

	return merchants, total, nil
}

func (r *MerchantRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE merchants SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-deleting merchant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMerchantNotFound
	}
	return nil
}

func (r *MerchantRepository) getOne(ctx context.Context, query string, args ...any) (*domain.Merchant, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying merchant: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrMerchantNotFound
	}

	return scanMerchant(rows)
}

func scanMerchant(rows pgx.Rows) (*domain.Merchant, error) {
	var m domain.Merchant
	var kycStatus, status string

	err := rows.Scan(
		&m.ID, &m.BusinessName, &m.BusinessType, &m.RegistrationNo, &m.Website,
		&m.ContactEmail, &m.ContactPhone, &m.ContactName,
		&m.AddressLine1, &m.AddressLine2, &m.City, &m.District, &m.PostalCode, &m.Country,
		&kycStatus, &m.KYCSubmittedAt, &m.KYCReviewedAt, &m.KYCRejectionReason,
		&m.TransactionLimitUSDT, &m.DailyLimitUSDT, &m.MonthlyLimitUSDT,
		&m.InstantAccessRemaining,
		&m.BankName, &m.BankBranch, &m.BankAccountNo, &m.BankAccountName,
		&m.DefaultCurrency, &m.DefaultProvider, &status,
		&m.StatusReason, &m.StatusChangedAt, &m.KYCReviewNotes,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning merchant: %w", err)
	}

	m.KYCStatus = domain.KYCStatus(kycStatus)
	m.Status = domain.MerchantStatus(status)

	return &m, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return errors.As(err, new(*duplicateKeyError)) ||
		// pgx wraps the error message
		contains(err.Error(), "duplicate key") ||
		contains(err.Error(), "23505")
}

type duplicateKeyError struct{}

func (e *duplicateKeyError) Error() string { return "duplicate key" }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
