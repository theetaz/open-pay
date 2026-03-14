package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKey(t *testing.T) {
	merchantID := uuid.New()

	t.Run("live key", func(t *testing.T) {
		key, plainSecret, err := domain.NewAPIKey(merchantID, "live", "Production Key")
		require.NoError(t, err)
		assert.Equal(t, merchantID, key.MerchantID)
		assert.Equal(t, "live", key.Environment)
		assert.Equal(t, "Production Key", key.Name)
		assert.Contains(t, key.KeyID, "ak_live_")
		assert.NotEmpty(t, key.SecretHash)
		assert.True(t, key.IsActive)
		assert.NotEmpty(t, plainSecret)
		assert.Contains(t, plainSecret, "sk_live_")
	})

	t.Run("test key", func(t *testing.T) {
		key, plainSecret, err := domain.NewAPIKey(merchantID, "test", "Test Key")
		require.NoError(t, err)
		assert.Contains(t, key.KeyID, "ak_test_")
		assert.Contains(t, plainSecret, "sk_test_")
		assert.Equal(t, "test", key.Environment)
	})

	t.Run("invalid environment", func(t *testing.T) {
		_, _, err := domain.NewAPIKey(merchantID, "staging", "Key")
		require.Error(t, err)
	})

	t.Run("verify secret", func(t *testing.T) {
		key, plainSecret, err := domain.NewAPIKey(merchantID, "live", "Key")
		require.NoError(t, err)

		assert.True(t, key.VerifySecret(plainSecret))
		assert.False(t, key.VerifySecret("wrong_secret"))
	})
}

func TestAPIKeyRevoke(t *testing.T) {
	merchantID := uuid.New()
	key, _, _ := domain.NewAPIKey(merchantID, "live", "Key")

	t.Run("revoke active key", func(t *testing.T) {
		err := key.Revoke("No longer needed")
		require.NoError(t, err)
		assert.False(t, key.IsActive)
		assert.Equal(t, "No longer needed", key.RevokedReason)
		assert.NotNil(t, key.RevokedAt)
	})

	t.Run("revoke already revoked key", func(t *testing.T) {
		err := key.Revoke("Again")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrKeyAlreadyRevoked)
	})
}
