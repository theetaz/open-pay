package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookConfig(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid config", func(t *testing.T) {
		cfg, err := domain.NewWebhookConfig(merchantID, "https://merchant.com/webhook")
		require.NoError(t, err)
		assert.Equal(t, merchantID, cfg.MerchantID)
		assert.Equal(t, "https://merchant.com/webhook", cfg.URL)
		assert.NotEmpty(t, cfg.SigningPublicKey)
		assert.NotEmpty(t, cfg.SigningPrivateKey)
		assert.True(t, cfg.IsActive)
	})

	t.Run("empty URL", func(t *testing.T) {
		_, err := domain.NewWebhookConfig(merchantID, "")
		require.Error(t, err)
	})

	t.Run("non-HTTPS URL", func(t *testing.T) {
		_, err := domain.NewWebhookConfig(merchantID, "http://insecure.com/webhook")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWebhookConfig)
	})
}

func TestNewDelivery(t *testing.T) {
	configID := uuid.New()
	merchantID := uuid.New()
	payload := []byte(`{"event":"payment.paid","paymentId":"123"}`)

	t.Run("valid delivery", func(t *testing.T) {
		d := domain.NewDelivery(configID, merchantID, "payment.paid", payload)
		assert.Equal(t, configID, d.WebhookConfigID)
		assert.Equal(t, "payment.paid", d.EventType)
		assert.Equal(t, domain.DeliveryPending, d.Status)
		assert.Equal(t, 0, d.AttemptCount)
		assert.Equal(t, 5, d.MaxAttempts)
	})
}

func TestDeliveryRetry(t *testing.T) {
	d := domain.NewDelivery(uuid.New(), uuid.New(), "payment.paid", []byte(`{}`))

	t.Run("first retry schedules 1 minute delay", func(t *testing.T) {
		d.RecordFailure(500, "server error", "")
		assert.Equal(t, 1, d.AttemptCount)
		assert.Equal(t, domain.DeliveryPending, d.Status)
		assert.NotNil(t, d.NextAttemptAt)
		assert.WithinDuration(t, time.Now().Add(1*time.Minute), *d.NextAttemptAt, 5*time.Second)
	})

	t.Run("delays double each retry", func(t *testing.T) {
		d.RecordFailure(500, "server error", "")
		assert.Equal(t, 2, d.AttemptCount)
		// 2nd retry = 2 min delay
		assert.WithinDuration(t, time.Now().Add(2*time.Minute), *d.NextAttemptAt, 5*time.Second)

		d.RecordFailure(500, "server error", "")
		assert.Equal(t, 3, d.AttemptCount)
		// 3rd retry = 4 min delay

		d.RecordFailure(500, "server error", "")
		assert.Equal(t, 4, d.AttemptCount)
		// 4th retry = 8 min delay
	})

	t.Run("exhausted after max attempts", func(t *testing.T) {
		d.RecordFailure(500, "server error", "")
		assert.Equal(t, 5, d.AttemptCount)
		assert.Equal(t, domain.DeliveryExhausted, d.Status)
	})
}

func TestDeliverySuccess(t *testing.T) {
	d := domain.NewDelivery(uuid.New(), uuid.New(), "payment.paid", []byte(`{}`))

	d.RecordSuccess(200)
	assert.Equal(t, domain.DeliveryDelivered, d.Status)
	assert.Equal(t, 1, d.AttemptCount)
	assert.Equal(t, 200, d.LastResponseCode)
	assert.NotNil(t, d.DeliveredAt)
}

func TestDeliveryCanRetry(t *testing.T) {
	t.Run("pending delivery can retry", func(t *testing.T) {
		d := domain.NewDelivery(uuid.New(), uuid.New(), "payment.paid", []byte(`{}`))
		assert.True(t, d.CanRetry())
	})

	t.Run("delivered cannot retry", func(t *testing.T) {
		d := domain.NewDelivery(uuid.New(), uuid.New(), "payment.paid", []byte(`{}`))
		d.RecordSuccess(200)
		assert.False(t, d.CanRetry())
	})

	t.Run("exhausted can be manually retried", func(t *testing.T) {
		d := domain.NewDelivery(uuid.New(), uuid.New(), "payment.paid", []byte(`{}`))
		d.Status = domain.DeliveryExhausted
		d.AttemptCount = 5

		d.ResetForRetry()
		assert.Equal(t, domain.DeliveryPending, d.Status)
		assert.Equal(t, 0, d.AttemptCount)
		assert.True(t, d.CanRetry())
	})
}
