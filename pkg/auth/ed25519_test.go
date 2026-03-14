package auth_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestED25519KeyPair(t *testing.T) {
	t.Run("generate key pair", func(t *testing.T) {
		pub, priv, err := auth.GenerateED25519KeyPair()
		require.NoError(t, err)
		assert.NotEmpty(t, pub)
		assert.NotEmpty(t, priv)
	})

	t.Run("key pairs are unique", func(t *testing.T) {
		pub1, _, _ := auth.GenerateED25519KeyPair()
		pub2, _, _ := auth.GenerateED25519KeyPair()
		assert.NotEqual(t, pub1, pub2)
	})
}

func TestWebhookSigning(t *testing.T) {
	pub, priv, err := auth.GenerateED25519KeyPair()
	require.NoError(t, err)

	t.Run("sign and verify webhook", func(t *testing.T) {
		timestamp := "1710410400000"
		payload := `{"event":"payment.paid","paymentId":"123"}`

		signature, err := auth.SignWebhook(priv, timestamp, []byte(payload))
		require.NoError(t, err)
		assert.NotEmpty(t, signature)

		valid, err := auth.VerifyWebhook(pub, timestamp, []byte(payload), signature)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("wrong public key fails", func(t *testing.T) {
		otherPub, _, _ := auth.GenerateED25519KeyPair()

		timestamp := "1710410400000"
		payload := `{"event":"payment.paid"}`

		signature, _ := auth.SignWebhook(priv, timestamp, []byte(payload))

		valid, err := auth.VerifyWebhook(otherPub, timestamp, []byte(payload), signature)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("tampered payload fails", func(t *testing.T) {
		timestamp := "1710410400000"
		payload := `{"event":"payment.paid"}`

		signature, _ := auth.SignWebhook(priv, timestamp, []byte(payload))

		valid, err := auth.VerifyWebhook(pub, timestamp, []byte(`{"event":"payment.failed"}`), signature)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("tampered timestamp fails", func(t *testing.T) {
		payload := `{"event":"payment.paid"}`

		signature, _ := auth.SignWebhook(priv, "1710410400000", []byte(payload))

		valid, err := auth.VerifyWebhook(pub, "9999999999999", []byte(payload), signature)
		require.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestED25519KeySerialization(t *testing.T) {
	pub, priv, err := auth.GenerateED25519KeyPair()
	require.NoError(t, err)

	t.Run("round-trip public key", func(t *testing.T) {
		encoded := auth.EncodePublicKey(pub)
		decoded, err := auth.DecodePublicKey(encoded)
		require.NoError(t, err)
		assert.Equal(t, pub, decoded)
	})

	t.Run("round-trip private key", func(t *testing.T) {
		encoded := auth.EncodePrivateKey(priv)
		decoded, err := auth.DecodePrivateKey(encoded)
		require.NoError(t, err)
		assert.Equal(t, priv, decoded)
	})
}
