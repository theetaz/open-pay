package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// UserRepository is the PostgreSQL implementation for user persistence.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, merchant_id, email, password_hash, name, role, branch_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.MerchantID, u.Email, u.PasswordHash,
		u.Name, u.Role, u.BranchID, u.IsActive,
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateEmail
		}
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, merchant_id, email, password_hash, name, role, branch_id,
	          is_active, last_login_at, created_at, updated_at
	          FROM users WHERE id = $1`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.MerchantID, &u.Email, &u.PasswordHash,
		&u.Name, &u.Role, &u.BranchID,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, merchantID uuid.UUID, email string) (*domain.User, error) {
	query := `SELECT id, merchant_id, email, password_hash, name, role, branch_id,
	          is_active, last_login_at, created_at, updated_at
	          FROM users WHERE merchant_id = $1 AND email = $2`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, merchantID, email).Scan(
		&u.ID, &u.MerchantID, &u.Email, &u.PasswordHash,
		&u.Name, &u.Role, &u.BranchID,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return &u, nil
}

func (r *UserRepository) GetByEmailGlobal(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, merchant_id, email, password_hash, name, role, branch_id,
	          is_active, last_login_at, created_at, updated_at
	          FROM users WHERE email = $1 AND is_active = TRUE
	          LIMIT 1`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.MerchantID, &u.Email, &u.PasswordHash,
		&u.Name, &u.Role, &u.BranchID,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return &u, nil
}

func (r *UserRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.User, error) {
	query := `SELECT id, merchant_id, email, password_hash, name, role, branch_id,
	          is_active, last_login_at, created_at, updated_at
	          FROM users WHERE merchant_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.MerchantID, &u.Email, &u.PasswordHash,
			&u.Name, &u.Role, &u.BranchID,
			&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now().UTC()
	query := `UPDATE users SET email = $2, name = $3, role = $4, branch_id = $5,
	          is_active = $6, last_login_at = $7, updated_at = $8
	          WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		u.ID, u.Email, u.Name, u.Role, u.BranchID,
		u.IsActive, u.LastLoginAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
