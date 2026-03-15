package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/exchange/internal/domain"
	"github.com/openlankapay/openlankapay/services/exchange/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRateRepo struct {
	mu    sync.RWMutex
	rates []*domain.ExchangeRate
}

func newMockRateRepo() *mockRateRepo {
	return &mockRateRepo{}
}

func (m *mockRateRepo) Create(_ context.Context, rate *domain.ExchangeRate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rates = append(m.rates, rate)
	return nil
}

func (m *mockRateRepo) GetActive(_ context.Context, base, quote string) (*domain.ExchangeRate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := len(m.rates) - 1; i >= 0; i-- {
		r := m.rates[i]
		if r.BaseCurrency == base && r.QuoteCurrency == quote && r.IsActive {
			return r, nil
		}
	}
	return nil, domain.ErrRateNotFound
}

func (m *mockRateRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ExchangeRate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.rates {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, domain.ErrRateNotFound
}

func (m *mockRateRepo) GetHistorical(_ context.Context, base, quote string, from, to time.Time) ([]*domain.ExchangeRate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ExchangeRate
	for _, r := range m.rates {
		if r.BaseCurrency == base && r.QuoteCurrency == quote &&
			!r.FetchedAt.Before(from) && !r.FetchedAt.After(to) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRateRepo) GetAtTime(_ context.Context, base, quote string, at time.Time) (*domain.ExchangeRate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var closest *domain.ExchangeRate
	for _, r := range m.rates {
		if r.BaseCurrency == base && r.QuoteCurrency == quote {
			if closest == nil || r.FetchedAt.Sub(at).Abs() < closest.FetchedAt.Sub(at).Abs() {
				closest = r
			}
		}
	}
	if closest == nil {
		return nil, domain.ErrRateNotFound
	}
	return closest, nil
}

type mockEvents struct {
	published []string
}

func (m *mockEvents) Publish(_ context.Context, subject string, _ any) error {
	m.published = append(m.published, subject)
	return nil
}

func TestGetActiveRate(t *testing.T) {
	ctx := context.Background()
	repo := newMockRateRepo()
	events := &mockEvents{}
	svc := service.NewExchangeService(repo, events)

	// Seed a rate
	rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "MOCK")
	_ = repo.Create(ctx, rate)

	t.Run("returns active rate", func(t *testing.T) {
		result, err := svc.GetActiveRate(ctx, "USDT", "LKR")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(325).Equal(result.Rate))
	})

	t.Run("returns error for unknown pair", func(t *testing.T) {
		_, err := svc.GetActiveRate(ctx, "BTC", "EUR")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrRateNotFound)
	})
}

func TestUpdateRate(t *testing.T) {
	ctx := context.Background()
	repo := newMockRateRepo()
	events := &mockEvents{}
	svc := service.NewExchangeService(repo, events)

	t.Run("stores new rate and publishes event", func(t *testing.T) {
		err := svc.UpdateRate(ctx, "USDT", "LKR", decimal.NewFromFloat(326.50), "BINANCE")
		require.NoError(t, err)

		active, err := svc.GetActiveRate(ctx, "USDT", "LKR")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(326.50).Equal(active.Rate))
		assert.Contains(t, events.published, "exchange.rate.updated")
	})
}

func TestGetSnapshot(t *testing.T) {
	ctx := context.Background()
	repo := newMockRateRepo()
	svc := service.NewExchangeService(repo, &mockEvents{})

	rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "MOCK")
	_ = repo.Create(ctx, rate)

	t.Run("returns snapshot of active rate", func(t *testing.T) {
		snapshot, err := svc.GetSnapshot(ctx, "USDT", "LKR")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(325).Equal(snapshot.Rate))
		assert.Equal(t, rate.ID, snapshot.RateID)
	})
}

func TestConvertLKRToUSDT(t *testing.T) {
	ctx := context.Background()
	repo := newMockRateRepo()
	svc := service.NewExchangeService(repo, &mockEvents{})

	rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "MOCK")
	_ = repo.Create(ctx, rate)

	t.Run("converts LKR to USDT", func(t *testing.T) {
		usdt, snapshot, err := svc.ConvertLKRToUSDT(ctx, decimal.NewFromFloat(32500))
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(100).Equal(usdt))
		assert.NotNil(t, snapshot)
	})
}
