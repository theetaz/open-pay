package auth_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAPIKey(t *testing.T) {
	t.Run("valid live key", func(t *testing.T) {
		keyID, secret, err := auth.ParseAPIKey("ak_live_abc123.sk_live_xyz789")
		require.NoError(t, err)
		assert.Equal(t, "ak_live_abc123", keyID)
		assert.Equal(t, "sk_live_xyz789", secret)
	})

	t.Run("valid test key", func(t *testing.T) {
		keyID, secret, err := auth.ParseAPIKey("ak_test_abc123.sk_test_xyz789")
		require.NoError(t, err)
		assert.Equal(t, "ak_test_abc123", keyID)
		assert.Equal(t, "sk_test_xyz789", secret)
	})

	t.Run("missing separator", func(t *testing.T) {
		_, _, err := auth.ParseAPIKey("ak_live_abc123sk_live_xyz789")
		require.Error(t, err)
	})

	t.Run("empty key", func(t *testing.T) {
		_, _, err := auth.ParseAPIKey("")
		require.Error(t, err)
	})

	t.Run("invalid prefix", func(t *testing.T) {
		_, _, err := auth.ParseAPIKey("xx_live_abc.yy_live_xyz")
		require.Error(t, err)
	})
}

func TestGenerateAPIKey(t *testing.T) {
	t.Run("live key", func(t *testing.T) {
		key, err := auth.GenerateAPIKey("live")
		require.NoError(t, err)

		keyID, secret, err := auth.ParseAPIKey(key)
		require.NoError(t, err)
		assert.Contains(t, keyID, "ak_live_")
		assert.Contains(t, secret, "sk_live_")
	})

	t.Run("test key", func(t *testing.T) {
		key, err := auth.GenerateAPIKey("test")
		require.NoError(t, err)

		keyID, secret, err := auth.ParseAPIKey(key)
		require.NoError(t, err)
		assert.Contains(t, keyID, "ak_test_")
		assert.Contains(t, secret, "sk_test_")
	})

	t.Run("keys are unique", func(t *testing.T) {
		key1, _ := auth.GenerateAPIKey("live")
		key2, _ := auth.GenerateAPIKey("live")
		assert.NotEqual(t, key1, key2)
	})

	t.Run("invalid environment", func(t *testing.T) {
		_, err := auth.GenerateAPIKey("staging")
		require.Error(t, err)
	})
}

func TestHMACSignature(t *testing.T) {
	secret := "sk_live_test_secret_key_12345"

	t.Run("sign and verify", func(t *testing.T) {
		timestamp := "1710410400000"
		method := "POST"
		path := "/v1/payments"
		body := `{"amount":"10","currency":"USDT"}`

		signature := auth.SignRequest(secret, timestamp, method, path, body)
		assert.NotEmpty(t, signature)

		valid := auth.VerifySignature(secret, timestamp, method, path, body, signature)
		assert.True(t, valid)
	})

	t.Run("wrong secret fails verification", func(t *testing.T) {
		timestamp := "1710410400000"
		method := "POST"
		path := "/v1/payments"
		body := `{"amount":"10"}`

		signature := auth.SignRequest(secret, timestamp, method, path, body)
		valid := auth.VerifySignature("wrong_secret", timestamp, method, path, body, signature)
		assert.False(t, valid)
	})

	t.Run("tampered body fails verification", func(t *testing.T) {
		timestamp := "1710410400000"
		method := "POST"
		path := "/v1/payments"
		body := `{"amount":"10"}`

		signature := auth.SignRequest(secret, timestamp, method, path, body)
		valid := auth.VerifySignature(secret, timestamp, method, path, `{"amount":"999"}`, signature)
		assert.False(t, valid)
	})

	t.Run("GET request with empty body", func(t *testing.T) {
		timestamp := "1710410400000"
		method := "GET"
		path := "/v1/payments/123"
		body := ""

		signature := auth.SignRequest(secret, timestamp, method, path, body)
		valid := auth.VerifySignature(secret, timestamp, method, path, body, signature)
		assert.True(t, valid)
	})

	t.Run("deterministic signatures", func(t *testing.T) {
		sig1 := auth.SignRequest(secret, "123", "GET", "/test", "")
		sig2 := auth.SignRequest(secret, "123", "GET", "/test", "")
		assert.Equal(t, sig1, sig2)
	})
}

func TestTimestampValidation(t *testing.T) {
	t.Run("current timestamp is valid", func(t *testing.T) {
		ts := auth.CurrentTimestamp()
		assert.True(t, auth.IsTimestampValid(ts, 5*60*1000)) // 5 min window
	})

	t.Run("old timestamp is invalid", func(t *testing.T) {
		ts := "1000000000000" // very old
		assert.False(t, auth.IsTimestampValid(ts, 5*60*1000))
	})

	t.Run("invalid format", func(t *testing.T) {
		assert.False(t, auth.IsTimestampValid("not-a-number", 5*60*1000))
	})
}
