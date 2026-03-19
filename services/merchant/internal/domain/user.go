package domain

import (
	"fmt"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Role values for merchant dashboard users.
type Role = string

const (
	RoleAdmin   Role = "ADMIN"
	RoleManager Role = "MANAGER"
	RoleUser    Role = "USER"
)

var validRoles = map[string]bool{
	RoleAdmin:   true,
	RoleManager: true,
	RoleUser:    true,
}

// User represents a dashboard user within a merchant organization.
type User struct {
	ID              uuid.UUID
	MerchantID      uuid.UUID
	Email           string
	PasswordHash    string
	Name            string
	Role            Role
	BranchID        *uuid.UUID
	IsActive        bool
	TOTPSecret      string
	TOTPEnabled     bool
	TOTPBackupCodes []string
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUser creates a validated User entity with a hashed password.
func NewUser(merchantID uuid.UUID, email, password, name string, role Role, branchID *uuid.UUID) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidUser, err.Error())
	}

	if !validRoles[role] {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRole, role)
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		MerchantID:   merchantID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		BranchID:     branchID,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// VerifyPassword checks if a plain password matches the stored hash.
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ValidatePassword checks password strength requirements.
func ValidatePassword(password string) error {
	return validatePassword(password)
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	var hasUpper, hasNumber bool
	for _, c := range password {
		if unicode.IsUpper(c) {
			hasUpper = true
		}
		if unicode.IsDigit(c) {
			hasNumber = true
		}
	}

	if !hasUpper || !hasNumber {
		return ErrWeakPassword
	}

	return nil
}
