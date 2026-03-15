package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/notification/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailNotification(t *testing.T) {
	t.Run("valid email notification", func(t *testing.T) {
		n, err := domain.NewNotification(domain.NotificationInput{
			MerchantID: uuid.New(),
			Channel:    domain.ChannelEmail,
			Recipient:  "merchant@example.com",
			Subject:    "Payment Received",
			Body:       "You received 10 USDT",
			EventType:  "payment.paid",
		})
		require.NoError(t, err)
		assert.Equal(t, domain.ChannelEmail, n.Channel)
		assert.Equal(t, domain.StatusPending, n.Status)
		assert.NotEmpty(t, n.ID)
	})

	t.Run("empty recipient", func(t *testing.T) {
		_, err := domain.NewNotification(domain.NotificationInput{
			MerchantID: uuid.New(),
			Channel:    domain.ChannelEmail,
			Recipient:  "",
			Subject:    "Test",
			Body:       "Test",
		})
		require.Error(t, err)
	})

	t.Run("invalid channel", func(t *testing.T) {
		_, err := domain.NewNotification(domain.NotificationInput{
			MerchantID: uuid.New(),
			Channel:    "TELEGRAM",
			Recipient:  "user@test.com",
			Subject:    "Test",
			Body:       "Test",
		})
		require.Error(t, err)
	})
}

func TestNotificationStatus(t *testing.T) {
	n, _ := domain.NewNotification(domain.NotificationInput{
		MerchantID: uuid.New(),
		Channel:    domain.ChannelEmail,
		Recipient:  "test@test.com",
		Subject:    "Test",
		Body:       "Test body",
	})

	t.Run("mark sent", func(t *testing.T) {
		n.MarkSent()
		assert.Equal(t, domain.StatusSent, n.Status)
		assert.NotNil(t, n.SentAt)
	})

	t.Run("mark failed", func(t *testing.T) {
		n2, _ := domain.NewNotification(domain.NotificationInput{
			MerchantID: uuid.New(),
			Channel:    domain.ChannelEmail,
			Recipient:  "fail@test.com",
			Subject:    "Test",
			Body:       "Test",
		})
		n2.MarkFailed("SMTP connection refused")
		assert.Equal(t, domain.StatusFailed, n2.Status)
		assert.Equal(t, "SMTP connection refused", n2.FailureReason)
	})
}

func TestRenderTemplate(t *testing.T) {
	t.Run("payment confirmation", func(t *testing.T) {
		body := domain.RenderTemplate("payment.paid", map[string]string{
			"amount":   "10.00",
			"currency": "USDT",
			"paymentNo": "PAY-20260315-abc123",
		})
		assert.Contains(t, body, "10.00")
		assert.Contains(t, body, "USDT")
		assert.Contains(t, body, "PAY-20260315-abc123")
	})

	t.Run("unknown template returns default", func(t *testing.T) {
		body := domain.RenderTemplate("unknown.event", nil)
		assert.Contains(t, body, "notification")
	})
}
