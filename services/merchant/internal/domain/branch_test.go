package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBranch(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid branch", func(t *testing.T) {
		b, err := domain.NewBranch(merchantID, "Main Branch")
		require.NoError(t, err)
		assert.Equal(t, merchantID, b.MerchantID)
		assert.Equal(t, "Main Branch", b.Name)
		assert.True(t, b.IsActive)
		assert.NotEmpty(t, b.ID)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := domain.NewBranch(merchantID, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidBranch)
	})

	t.Run("zero merchant ID", func(t *testing.T) {
		_, err := domain.NewBranch(uuid.Nil, "Branch")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidBranch)
	})
}

func TestNewUser(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid admin user", func(t *testing.T) {
		u, err := domain.NewUser(merchantID, "admin@test.com", "SecurePass1", "Admin", domain.RoleAdmin, nil)
		require.NoError(t, err)
		assert.Equal(t, "admin@test.com", u.Email)
		assert.Equal(t, domain.RoleAdmin, u.Role)
		assert.True(t, u.IsActive)
		assert.NotEmpty(t, u.PasswordHash)
		// Password hash should not be the plaintext
		assert.NotEqual(t, "SecurePass1", u.PasswordHash)
	})

	t.Run("valid branch-scoped user", func(t *testing.T) {
		branchID := uuid.New()
		u, err := domain.NewUser(merchantID, "user@test.com", "SecurePass1", "User", domain.RoleUser, &branchID)
		require.NoError(t, err)
		assert.Equal(t, domain.RoleUser, u.Role)
		assert.Equal(t, &branchID, u.BranchID)
	})

	t.Run("invalid email", func(t *testing.T) {
		_, err := domain.NewUser(merchantID, "bad", "SecurePass1", "User", domain.RoleUser, nil)
		require.Error(t, err)
	})

	t.Run("weak password", func(t *testing.T) {
		_, err := domain.NewUser(merchantID, "user@test.com", "short", "User", domain.RoleUser, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWeakPassword)
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := domain.NewUser(merchantID, "user@test.com", "SecurePass1", "User", "SUPERADMIN", nil)
		require.Error(t, err)
	})

	t.Run("verify password", func(t *testing.T) {
		u, _ := domain.NewUser(merchantID, "user@test.com", "SecurePass1", "User", domain.RoleAdmin, nil)
		assert.True(t, u.VerifyPassword("SecurePass1"))
		assert.False(t, u.VerifyPassword("WrongPass1"))
	})
}
