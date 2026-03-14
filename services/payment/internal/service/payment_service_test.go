package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/provider"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPaymentRepo struct {
	payments map[uuid.UUID]*domain.Payment
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{payments: make(map[uuid.UUID]*domain.Payment)}
}

func (m *mockPaymentRepo) Create(_ context.Context, p *domain.Payment) error {
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

func (m *mockPaymentRepo) Update(_ context.Context, p *domain.Payment) error {
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepo) List(_ context.Context, _ uuid.UUID, _ service.ListParams) ([]*domain.Payment, int, error) {
	var result []*domain.Payment
	for _, p := range m.payments {
		result = append(result, p)
	}
	return result, len(result), nil
}

type mockEvents struct {
	published []string
}

func (m *mockEvents) Publish(_ context.Context, subject string, _ any) error {
	m.published = append(m.published, subject)
	return nil
}

func TestCreatePayment(t *testing.T) {
	ctx := context.Background()

	t.Run("successful USDT payment", func(t *testing.T) {
		mockProvider := provider.NewMockProvider()
		events := &mockEvents{}
		svc := service.NewPaymentService(
			newMockPaymentRepo(),
			map[string]domain.PaymentProvider{"TEST": mockProvider},
			events,
		)

		payment, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
			MerchantID: uuid.New(),
			Amount:     decimal.NewFromFloat(10),
			Currency:   "USDT",
			Provider:   "TEST",
		})
		require.NoError(t, err)
		assert.Equal(t, domain.StatusInitiated, payment.Status)
		assert.NotEmpty(t, payment.QRContent)
		assert.NotEmpty(t, payment.CheckoutLink)
		assert.NotEmpty(t, payment.ProviderPayID)
		assert.Contains(t, events.published, "payment.initiated")
	})

	t.Run("invalid amount", func(t *testing.T) {
		svc := service.NewPaymentService(
			newMockPaymentRepo(),
			map[string]domain.PaymentProvider{"TEST": provider.NewMockProvider()},
			&mockEvents{},
		)

		_, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
			MerchantID: uuid.New(),
			Amount:     decimal.NewFromFloat(-5),
			Currency:   "USDT",
			Provider:   "TEST",
		})
		require.Error(t, err)
	})

	t.Run("unknown provider", func(t *testing.T) {
		svc := service.NewPaymentService(
			newMockPaymentRepo(),
			map[string]domain.PaymentProvider{"TEST": provider.NewMockProvider()},
			&mockEvents{},
		)

		_, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
			MerchantID: uuid.New(),
			Amount:     decimal.NewFromFloat(10),
			Currency:   "USDT",
			Provider:   "UNKNOWN",
		})
		require.Error(t, err)
	})
}

func TestGetPayment(t *testing.T) {
	ctx := context.Background()
	mockProvider := provider.NewMockProvider()
	svc := service.NewPaymentService(
		newMockPaymentRepo(),
		map[string]domain.PaymentProvider{"TEST": mockProvider},
		&mockEvents{},
	)

	created, _ := svc.CreatePayment(ctx, service.CreatePaymentInput{
		MerchantID: uuid.New(),
		Amount:     decimal.NewFromFloat(10),
		Currency:   "USDT",
		Provider:   "TEST",
	})

	t.Run("get existing payment", func(t *testing.T) {
		payment, err := svc.GetPayment(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, payment.ID)
	})

	t.Run("get nonexistent payment", func(t *testing.T) {
		_, err := svc.GetPayment(ctx, uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
	})
}

func TestHandleProviderCallback(t *testing.T) {
	ctx := context.Background()
	mockProv := provider.NewMockProvider()
	events := &mockEvents{}
	svc := service.NewPaymentService(
		newMockPaymentRepo(),
		map[string]domain.PaymentProvider{"TEST": mockProv},
		events,
	)

	created, _ := svc.CreatePayment(ctx, service.CreatePaymentInput{
		MerchantID: uuid.New(),
		Amount:     decimal.NewFromFloat(10),
		Currency:   "USDT",
		Provider:   "TEST",
	})

	t.Run("payment confirmed by provider", func(t *testing.T) {
		// Simulate the provider confirming payment
		mockProv.SimulatePayment(created.ProviderPayID)

		err := svc.HandleProviderCallback(ctx, created.ID)
		require.NoError(t, err)

		updated, _ := svc.GetPayment(ctx, created.ID)
		assert.Equal(t, domain.StatusPaid, updated.Status)
		assert.NotNil(t, updated.PaidAt)
		assert.Contains(t, events.published, "payment.paid")
	})
}
