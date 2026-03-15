package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminUserNotFound    = errors.New("admin user not found")
	ErrAdminRoleNotFound    = errors.New("admin role not found")
	ErrInvalidAdminUser     = errors.New("invalid admin user data")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAdminAccountInactive = errors.New("admin account is inactive")
	ErrDuplicateAdminEmail  = errors.New("admin email already in use")
	ErrPermissionDenied     = errors.New("permission denied")
)

// AdminRole represents a role with associated permissions.
type AdminRole struct {
	ID          uuid.UUID
	Name        string
	Description string
	Permissions []string
	IsSystem    bool
	CreatedAt   time.Time
}

// HasPermission checks if the role has a specific permission.
func (r *AdminRole) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AdminUser represents a platform administrator.
type AdminUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	RoleID       uuid.UUID
	Role         *AdminRole
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewAdminUser creates a validated admin user.
func NewAdminUser(email, password, name string, roleID uuid.UUID) (*AdminUser, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidAdminUser)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidAdminUser)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidAdminUser)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	return &AdminUser{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// VerifyPassword checks if a plain password matches the stored hash.
func (u *AdminUser) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// HasPermission checks if the admin user has a specific permission via their role.
func (u *AdminUser) HasPermission(permission string) bool {
	if u.Role == nil {
		return false
	}
	return u.Role.HasPermission(permission)
}
