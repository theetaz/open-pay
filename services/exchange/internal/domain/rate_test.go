package domain_test

import (
	"testing"
	"time"

	"github.com/openlankapay/openlankapay/services/exchange/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExchangeRate(t *testing.T) {
	t.Run("valid rate", func(t *testing.T) {
		rate, err := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325.50), "BINANCE")
		require.NoError(t, err)
		assert.Equal(t, "USDT", rate.BaseCurrency)
		assert.Equal(t, "LKR", rate.QuoteCurrency)
		assert.True(t, decimal.NewFromFloat(325.50).Equal(rate.Rate))
		assert.Equal(t, "BINANCE", rate.Source)
		assert.True(t, rate.IsActive)
		assert.NotEmpty(t, rate.ID)
	})

	t.Run("zero rate is invalid", func(t *testing.T) {
		_, err := domain.NewExchangeRate("USDT", "LKR", decimal.Zero, "BINANCE")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidRate)
	})

	t.Run("negative rate is invalid", func(t *testing.T) {
		_, err := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(-1), "BINANCE")
		require.Error(t, err)
	})

	t.Run("empty base currency", func(t *testing.T) {
		_, err := domain.NewExchangeRate("", "LKR", decimal.NewFromFloat(325), "BINANCE")
		require.Error(t, err)
	})

	t.Run("empty source", func(t *testing.T) {
		_, err := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "")
		require.Error(t, err)
	})
}

func TestConvertAmount(t *testing.T) {
	rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "BINANCE")

	t.Run("convert USDT to LKR", func(t *testing.T) {
		lkr := rate.ConvertBaseToQuote(decimal.NewFromFloat(100))
		assert.True(t, decimal.NewFromFloat(32500).Equal(lkr))
	})

	t.Run("convert LKR to USDT", func(t *testing.T) {
		usdt := rate.ConvertQuoteToBase(decimal.NewFromFloat(32500))
		assert.True(t, decimal.NewFromFloat(100).Equal(usdt))
	})

	t.Run("small amount precision", func(t *testing.T) {
		usdt := rate.ConvertQuoteToBase(decimal.NewFromFloat(325))
		assert.True(t, decimal.NewFromFloat(1).Equal(usdt))
	})
}

func TestRateSnapshot(t *testing.T) {
	t.Run("create snapshot from rate", func(t *testing.T) {
		rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "BINANCE")
		snapshot := rate.Snapshot()

		assert.Equal(t, rate.ID, snapshot.RateID)
		assert.True(t, rate.Rate.Equal(snapshot.Rate))
		assert.Equal(t, rate.Source, snapshot.Source)
		assert.WithinDuration(t, time.Now(), snapshot.SnapshotAt, time.Second)
	})
}

func TestRateIsStale(t *testing.T) {
	t.Run("fresh rate is not stale", func(t *testing.T) {
		rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "BINANCE")
		assert.False(t, rate.IsStale(5*time.Minute))
	})

	t.Run("old rate is stale", func(t *testing.T) {
		rate, _ := domain.NewExchangeRate("USDT", "LKR", decimal.NewFromFloat(325), "BINANCE")
		rate.FetchedAt = time.Now().Add(-10 * time.Minute)
		assert.True(t, rate.IsStale(5*time.Minute))
	})
}
