package provider_test

import (
	"context"
	"testing"

	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/provider"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockProvider(t *testing.T) {
	p := provider.NewMockProvider()

	t.Run("name returns TEST", func(t *testing.T) {
		assert.Equal(t, "TEST", p.Name())
	})

	t.Run("create payment returns QR and links", func(t *testing.T) {
		resp, err := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
			Amount:   "10",
			Currency: "USDT",
			OrderID:  "test-order-1",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ProviderPayID)
		assert.NotEmpty(t, resp.QRContent)
		assert.NotEmpty(t, resp.CheckoutLink)
	})

	t.Run("get status returns initiated for new payment", func(t *testing.T) {
		resp, _ := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
			Amount:  "10",
			OrderID: "test-order-2",
		})

		status, err := p.GetPaymentStatus(context.Background(), resp.ProviderPayID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusInitiated, status.Status)
	})

	t.Run("simulate payment marks as paid", func(t *testing.T) {
		resp, _ := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
			Amount:  "10",
			OrderID: "test-order-3",
		})

		p.SimulatePayment(resp.ProviderPayID)

		status, err := p.GetPaymentStatus(context.Background(), resp.ProviderPayID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusPaid, status.Status)
		assert.NotEmpty(t, status.TxHash)
	})

	t.Run("amount 0.01 auto-expires", func(t *testing.T) {
		resp, _ := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
			Amount:  "0.01",
			OrderID: "expire-test",
		})

		status, err := p.GetPaymentStatus(context.Background(), resp.ProviderPayID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusExpired, status.Status)
	})

	t.Run("amount 0.02 auto-fails", func(t *testing.T) {
		resp, _ := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
			Amount:  "0.02",
			OrderID: "fail-test",
		})

		status, err := p.GetPaymentStatus(context.Background(), resp.ProviderPayID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFailed, status.Status)
	})
}
