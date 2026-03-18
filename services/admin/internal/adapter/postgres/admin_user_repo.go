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
		u.must_change_password, u.last_login_at, u.created_at, u.updated_at,
		r.id, r.name, r.description, r.permissions, r.is_system, r.created_at
		FROM admin_users u
		JOIN admin_roles r ON u.role_id = r.id
		WHERE u.email = $1`

	var user domain.AdminUser
	var role domain.AdminRole
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID,
		&user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
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
		u.must_change_password, u.last_login_at, u.created_at, u.updated_at,
		r.id, r.name, r.description, r.permissions, r.is_system, r.created_at
		FROM admin_users u
		JOIN admin_roles r ON u.role_id = r.id
		WHERE u.id = $1`

	var user domain.AdminUser
	var role domain.AdminRole
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID,
		&user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
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
	query := `INSERT INTO admin_users (id, email, password_hash, name, role_id, is_active, must_change_password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name,
		user.RoleID, user.IsActive, user.MustChangePassword, user.CreatedAt, user.UpdatedAt,
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

// ListUsers returns paginated admin users.
func (r *AdminUserRepository) ListUsers(ctx context.Context, page, perPage int) ([]*domain.AdminUser, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT u.id, u.email, u.password_hash, u.name, u.role_id, u.is_active,
		u.must_change_password, u.last_login_at, u.created_at, u.updated_at,
		r.id, r.name, r.description, r.permissions, r.is_system, r.created_at
		FROM admin_users u
		JOIN admin_roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.AdminUser
	for rows.Next() {
		var user domain.AdminUser
		var role domain.AdminRole
		var permissionsJSON []byte
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID,
			&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
			&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.IsSystem, &role.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(permissionsJSON, &role.Permissions)
		user.Role = &role
		users = append(users, &user)
	}
	return users, total, rows.Err()
}

// UpdateUser updates admin user fields.
func (r *AdminUserRepository) UpdateUser(ctx context.Context, user *domain.AdminUser) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE admin_users SET name = $1, role_id = $2, is_active = $3, updated_at = $4 WHERE id = $5`,
		user.Name, user.RoleID, user.IsActive, user.UpdatedAt, user.ID)
	return err
}

// ListRoles returns all admin roles.
func (r *AdminUserRepository) ListRoles(ctx context.Context) ([]*domain.AdminRole, error) {
	query := `SELECT id, name, description, permissions, is_system, created_at FROM admin_roles ORDER BY is_system DESC, name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.AdminRole
	for rows.Next() {
		var role domain.AdminRole
		var permissionsJSON []byte
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(permissionsJSON, &role.Permissions)
		roles = append(roles, &role)
	}
	return roles, rows.Err()
}

// CreateRole creates a new admin role.
func (r *AdminUserRepository) CreateRole(ctx context.Context, role *domain.AdminRole) error {
	permJSON, _ := json.Marshal(role.Permissions)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admin_roles (id, name, description, permissions, is_system, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		role.ID, role.Name, role.Description, permJSON, role.IsSystem, role.CreatedAt)
	return err
}

// UpdateRole updates a role's permissions and description.
func (r *AdminUserRepository) UpdateRole(ctx context.Context, role *domain.AdminRole) error {
	permJSON, _ := json.Marshal(role.Permissions)
	_, err := r.pool.Exec(ctx,
		`UPDATE admin_roles SET description = $1, permissions = $2 WHERE id = $3 AND is_system = FALSE`,
		role.Description, permJSON, role.ID)
	return err
}

// ChangePassword updates the password hash and clears must_change_password flag.
func (r *AdminUserRepository) ChangePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE admin_users SET password_hash = $1, must_change_password = FALSE, updated_at = $2 WHERE id = $3`,
		passwordHash, now, id)
	return err
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
