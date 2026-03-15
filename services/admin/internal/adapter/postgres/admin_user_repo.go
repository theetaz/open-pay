package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
)

type AdminUserRepository struct {
	pool *pgxpool.Pool
}

func NewAdminUserRepository(pool *pgxpool.Pool) *AdminUserRepository {
	return &AdminUserRepository{pool: pool}
}

func (r *AdminUserRepository) GetByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	query := `SELECT u.id, u.email, u.password_hash, u.name, u.role_id, u.is_active,
		u.last_login_at, u.created_at, u.updated_at,
		r.id, r.name, r.description, r.permissions, r.is_system, r.created_at
		FROM admin_users u
		JOIN admin_roles r ON u.role_id = r.id
		WHERE u.email = $1`

	var user domain.AdminUser
	var role domain.AdminRole
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.IsSystem, &role.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrAdminUserNotFound
	}

	_ = json.Unmarshal(permissionsJSON, &role.Permissions)
	user.Role = &role
	return &user, nil
}

func (r *AdminUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminUser, error) {
	query := `SELECT u.id, u.email, u.password_hash, u.name, u.role_id, u.is_active,
		u.last_login_at, u.created_at, u.updated_at,
		r.id, r.name, r.description, r.permissions, r.is_system, r.created_at
		FROM admin_users u
		JOIN admin_roles r ON u.role_id = r.id
		WHERE u.id = $1`

	var user domain.AdminUser
	var role domain.AdminRole
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.IsSystem, &role.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrAdminUserNotFound
	}

	_ = json.Unmarshal(permissionsJSON, &role.Permissions)
	user.Role = &role
	return &user, nil
}

func (r *AdminUserRepository) Create(ctx context.Context, user *domain.AdminUser) error {
	query := `INSERT INTO admin_users (id, email, password_hash, name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name,
		user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateAdminEmail
		}
		return fmt.Errorf("inserting admin user: %w", err)
	}
	return nil
}

func (r *AdminUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	query := `UPDATE admin_users SET last_login_at = $2, updated_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, now)
	return err
}

func (r *AdminUserRepository) GetRoleByName(ctx context.Context, name string) (*domain.AdminRole, error) {
	query := `SELECT id, name, description, permissions, is_system, created_at FROM admin_roles WHERE name = $1`

	var role domain.AdminRole
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.IsSystem, &role.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrAdminRoleNotFound
	}

	_ = json.Unmarshal(permissionsJSON, &role.Permissions)
	return &role, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	for i := 0; i <= len(s)-5; i++ {
		if s[i:i+5] == "23505" {
			return true
		}
	}
	return false
}
