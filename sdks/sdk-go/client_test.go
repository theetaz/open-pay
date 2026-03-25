package openpay_test

import (
	"testing"

	openpay "github.com/openlankapay/openlankapay/sdks/sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		client, err := openpay.NewClient("ak_live_testid.sk_live_testsecret", openpay.WithBaseURL("https://api.example.com"))
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Payments)
		assert.NotNil(t, client.Webhooks)
	})

	t.Run("empty API key", func(t *testing.T) {
		_, err := openpay.NewClient("")
		require.Error(t, err)
	})

	t.Run("invalid API key format", func(t *testing.T) {
		_, err := openpay.NewClient("bad-key")
		require.Error(t, err)
	})

	t.Run("sandbox mode", func(t *testing.T) {
		client, err := openpay.NewClient("ak_test_id.sk_test_secret", openpay.WithSandbox())
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestSignRequest(t *testing.T) {
	client, _ := openpay.NewClient("ak_live_testid.sk_live_testsecret")

	t.Run("generates valid headers", func(t *testing.T) {
		headers := client.SignHeaders("POST", "/v1/payments", `{"amount":"10"}`)
		assert.NotEmpty(t, headers["x-api-key"])
		assert.NotEmpty(t, headers["x-timestamp"])
		assert.NotEmpty(t, headers["x-signature"])
		assert.Equal(t, "ak_live_testid", headers["x-api-key"])
	})

	t.Run("different bodies produce different signatures", func(t *testing.T) {
		h1 := client.SignHeaders("POST", "/v1/payments", `{"amount":"10"}`)
		h2 := client.SignHeaders("POST", "/v1/payments", `{"amount":"20"}`)
		assert.NotEqual(t, h1["x-signature"], h2["x-signature"])
	})
}

func TestVerifyWebhook(t *testing.T) {
	t.Run("returns error for invalid signature", func(t *testing.T) {
		err := openpay.VerifyWebhookSignature("bad-pub-key", "12345", []byte("payload"), "bad-sig")
		assert.Error(t, err)
	})
}

func TestAPIError(t *testing.T) {
	err := &openpay.APIError{Code: "NOT_FOUND", Message: "payment not found", Status: 404}
	assert.Contains(t, err.Error(), "payment not found")
	assert.True(t, openpay.IsNotFound(err))
	assert.False(t, openpay.IsAuthError(err))

	authErr := &openpay.APIError{Code: "UNAUTHORIZED", Message: "bad key", Status: 401}
	assert.True(t, openpay.IsAuthError(authErr))
	assert.False(t, openpay.IsNotFound(authErr))
}
