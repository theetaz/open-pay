package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

type BranchRepository struct {
	pool *pgxpool.Pool
}

func NewBranchRepository(pool *pgxpool.Pool) *BranchRepository {
	return &BranchRepository{pool: pool}
}

func (r *BranchRepository) Create(ctx context.Context, b *domain.Branch) error {
	query := `INSERT INTO branches (id, merchant_id, name, description, address, city,
		bank_name, bank_branch, bank_account_no, bank_account_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		b.ID, b.MerchantID, b.Name, b.Description, b.Address, b.City,
		b.BankName, b.BankBranch, b.BankAccountNo, b.BankAccountName,
		b.IsActive, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting branch: %w", err)
	}
	return nil
}

func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	query := `SELECT id, merchant_id, name, description, address, city,
		bank_name, bank_branch, bank_account_no, bank_account_name,
		is_active, created_at, updated_at, deleted_at
		FROM branches WHERE id = $1 AND deleted_at IS NULL`

	var b domain.Branch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.MerchantID, &b.Name, &b.Description, &b.Address, &b.City,
		&b.BankName, &b.BankBranch, &b.BankAccountNo, &b.BankAccountName,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	if err != nil {
		return nil, domain.ErrBranchNotFound
	}
	return &b, nil
}

func (r *BranchRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Branch, error) {
	query := `SELECT id, merchant_id, name, description, address, city,
		bank_name, bank_branch, bank_account_no, bank_account_name,
		is_active, created_at, updated_at, deleted_at
		FROM branches WHERE merchant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}
	defer rows.Close()

	var branches []*domain.Branch
	for rows.Next() {
		var b domain.Branch
		if err := rows.Scan(
			&b.ID, &b.MerchantID, &b.Name, &b.Description, &b.Address, &b.City,
			&b.BankName, &b.BankBranch, &b.BankAccountNo, &b.BankAccountName,
			&b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning branch: %w", err)
		}
		branches = append(branches, &b)
	}
	return branches, nil
}

func (r *BranchRepository) Update(ctx context.Context, b *domain.Branch) error {
	b.UpdatedAt = time.Now().UTC()
	query := `UPDATE branches SET name = $2, description = $3, address = $4, city = $5,
		bank_name = $6, bank_branch = $7, bank_account_no = $8, bank_account_name = $9,
		is_active = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		b.ID, b.Name, b.Description, b.Address, b.City,
		b.BankName, b.BankBranch, b.BankAccountNo, b.BankAccountName,
		b.IsActive, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating branch: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBranchNotFound
	}
	return nil
}

func (r *BranchRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE branches SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-deleting branch: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBranchNotFound
	}
	return nil
}
