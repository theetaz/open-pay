package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
	"github.com/openlankapay/openlankapay/services/webhook/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigRepo struct {
	configs map[uuid.UUID]*domain.WebhookConfig
}

func newMockConfigRepo() *mockConfigRepo {
	return &mockConfigRepo{configs: make(map[uuid.UUID]*domain.WebhookConfig)}
}

func (m *mockConfigRepo) Create(_ context.Context, cfg *domain.WebhookConfig) error {
	m.configs[cfg.MerchantID] = cfg
	return nil
}

func (m *mockConfigRepo) GetByMerchantID(_ context.Context, merchantID uuid.UUID) (*domain.WebhookConfig, error) {
	cfg, ok := m.configs[merchantID]
	if !ok {
		return nil, domain.ErrWebhookConfigNotFound
	}
	return cfg, nil
}

func (m *mockConfigRepo) Update(_ context.Context, cfg *domain.WebhookConfig) error {
	m.configs[cfg.MerchantID] = cfg
	return nil
}

type mockDeliveryRepo struct {
	deliveries map[uuid.UUID]*domain.Delivery
}

func newMockDeliveryRepo() *mockDeliveryRepo {
	return &mockDeliveryRepo{deliveries: make(map[uuid.UUID]*domain.Delivery)}
}

func (m *mockDeliveryRepo) Create(_ context.Context, d *domain.Delivery) error {
	m.deliveries[d.ID] = d
	return nil
}

func (m *mockDeliveryRepo) Update(_ context.Context, d *domain.Delivery) error {
	m.deliveries[d.ID] = d
	return nil
}

func (m *mockDeliveryRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Delivery, error) {
	d, ok := m.deliveries[id]
	if !ok {
		return nil, domain.ErrDeliveryNotFound
	}
	return d, nil
}

func (m *mockDeliveryRepo) ListRetryable(_ context.Context, _ int) ([]*domain.Delivery, error) {
	return nil, nil
}

func (m *mockDeliveryRepo) ListByMerchant(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.Delivery, int, error) {
	return nil, 0, nil
}

func TestConfigureWebhook(t *testing.T) {
	ctx := context.Background()
	svc := service.NewWebhookService(newMockConfigRepo(), newMockDeliveryRepo())

	t.Run("create new config", func(t *testing.T) {
		merchantID := uuid.New()
		cfg, err := svc.Configure(ctx, merchantID, "https://merchant.com/webhook")
		require.NoError(t, err)
		assert.Equal(t, merchantID, cfg.MerchantID)
		assert.NotEmpty(t, cfg.SigningPublicKey)
	})

	t.Run("reject non-HTTPS", func(t *testing.T) {
		_, err := svc.Configure(ctx, uuid.New(), "http://insecure.com/hook")
		require.Error(t, err)
	})
}

func TestDeliverWebhook(t *testing.T) {
	ctx := context.Background()
	configRepo := newMockConfigRepo()
	deliveryRepo := newMockDeliveryRepo()
	svc := service.NewWebhookService(configRepo, deliveryRepo)

	merchantID := uuid.New()
	cfg, _ := svc.Configure(ctx, merchantID, "https://merchant.com/webhook")

	t.Run("delivers to merchant endpoint", func(t *testing.T) {
		var received atomic.Bool
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received.Store(true)
			assert.Equal(t, "POST", r.Method)
			assert.NotEmpty(t, r.Header.Get("X-Webhook-Signature"))
			assert.NotEmpty(t, r.Header.Get("X-Webhook-Timestamp"))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// Update config to point to test server
		cfg.URL = ts.URL
		_ = configRepo.Update(ctx, cfg)

		payload := []byte(`{"event":"payment.paid","paymentId":"test-123"}`)
		delivery, err := svc.Deliver(ctx, merchantID, "payment.paid", payload, ts.Client())
		require.NoError(t, err)
		assert.Equal(t, domain.DeliveryDelivered, delivery.Status)
		assert.True(t, received.Load())
	})
}

func TestGetPublicKey(t *testing.T) {
	ctx := context.Background()
	svc := service.NewWebhookService(newMockConfigRepo(), newMockDeliveryRepo())

	merchantID := uuid.New()
	_, _ = svc.Configure(ctx, merchantID, "https://merchant.com/webhook")

	t.Run("returns public key", func(t *testing.T) {
		pubKey, err := svc.GetPublicKey(ctx, merchantID)
		require.NoError(t, err)
		assert.NotEmpty(t, pubKey)
	})

	t.Run("not found for unknown merchant", func(t *testing.T) {
		_, err := svc.GetPublicKey(ctx, uuid.New())
		require.Error(t, err)
	})
}
