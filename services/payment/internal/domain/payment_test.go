package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid USDT payment", func(t *testing.T) {
		p, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(10),
			Currency:   "USDT",
			Provider:   "BYBIT",
		})
		require.NoError(t, err)
		assert.Equal(t, merchantID, p.MerchantID)
		assert.True(t, decimal.NewFromFloat(10).Equal(p.Amount))
		assert.Equal(t, "USDT", p.Currency)
		assert.Equal(t, domain.StatusInitiated, p.Status)
		assert.NotEmpty(t, p.ID)
		assert.NotEmpty(t, p.PaymentNo)
		assert.False(t, p.ExpireTime.IsZero())
	})

	t.Run("valid LKR payment", func(t *testing.T) {
		p, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(5000),
			Currency:   "LKR",
			Provider:   "BINANCE",
		})
		require.NoError(t, err)
		assert.Equal(t, "LKR", p.Currency)
	})

	t.Run("zero amount is invalid", func(t *testing.T) {
		_, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.Zero,
			Currency:   "USDT",
			Provider:   "BYBIT",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPayment)
	})

	t.Run("negative amount is invalid", func(t *testing.T) {
		_, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(-5),
			Currency:   "USDT",
			Provider:   "BYBIT",
		})
		require.Error(t, err)
	})

	t.Run("unsupported currency", func(t *testing.T) {
		_, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(10),
			Currency:   "EUR",
			Provider:   "BYBIT",
		})
		require.Error(t, err)
	})

	t.Run("unsupported provider", func(t *testing.T) {
		_, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(10),
			Currency:   "USDT",
			Provider:   "STRIPE",
		})
		require.Error(t, err)
	})

	t.Run("custom expiration", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		p, err := domain.NewPayment(domain.CreatePaymentInput{
			MerchantID: merchantID,
			Amount:     decimal.NewFromFloat(10),
			Currency:   "USDT",
			Provider:   "BYBIT",
			ExpireTime: &expiry,
		})
		require.NoError(t, err)
		assert.WithinDuration(t, expiry, p.ExpireTime, time.Second)
	})
}

func TestPaymentStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.PaymentStatus
		to      domain.PaymentStatus
		wantErr bool
	}{
		{"initiated to paid", domain.StatusInitiated, domain.StatusPaid, false},
		{"initiated to expired", domain.StatusInitiated, domain.StatusExpired, false},
		{"initiated to failed", domain.StatusInitiated, domain.StatusFailed, false},
		{"initiated to user_review", domain.StatusInitiated, domain.StatusUserReview, false},
		{"user_review to paid", domain.StatusUserReview, domain.StatusPaid, false},
		{"user_review to failed", domain.StatusUserReview, domain.StatusFailed, false},
		{"user_review to expired", domain.StatusUserReview, domain.StatusExpired, false},
		{"paid to expired (invalid)", domain.StatusPaid, domain.StatusExpired, true},
		{"paid to failed (invalid)", domain.StatusPaid, domain.StatusFailed, true},
		{"expired to paid (invalid)", domain.StatusExpired, domain.StatusPaid, true},
		{"failed to paid (invalid)", domain.StatusFailed, domain.StatusPaid, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &domain.Payment{Status: tt.from}
			err := p.TransitionTo(tt.to)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.to, p.Status)
			}
		})
	}
}

func TestPaymentMarkPaid(t *testing.T) {
	p := &domain.Payment{Status: domain.StatusInitiated}
	txHash := "0xabc123"

	err := p.MarkPaid(txHash)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPaid, p.Status)
	assert.Equal(t, txHash, p.TxHash)
	assert.NotNil(t, p.PaidAt)
}

func TestPaymentIsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		p := &domain.Payment{
			Status:     domain.StatusInitiated,
			ExpireTime: time.Now().Add(10 * time.Minute),
		}
		assert.False(t, p.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		p := &domain.Payment{
			Status:     domain.StatusInitiated,
			ExpireTime: time.Now().Add(-1 * time.Minute),
		}
		assert.True(t, p.IsExpired())
	})

	t.Run("already paid is not expired", func(t *testing.T) {
		p := &domain.Payment{
			Status:     domain.StatusPaid,
			ExpireTime: time.Now().Add(-1 * time.Minute),
		}
		assert.False(t, p.IsExpired())
	})
}

func TestPaymentSetFees(t *testing.T) {
	p := &domain.Payment{
		Amount:     decimal.NewFromFloat(100),
		AmountUSDT: decimal.NewFromFloat(100),
		Currency:   "USDT",
	}

	p.SetFees(decimal.NewFromFloat(0.5), decimal.NewFromFloat(1.5))

	assert.True(t, decimal.NewFromFloat(0.5).Equal(p.ExchangeFeePct))
	assert.True(t, decimal.NewFromFloat(0.5).Equal(p.ExchangeFeeUSDT))
	assert.True(t, decimal.NewFromFloat(1.5).Equal(p.PlatformFeePct))
	assert.True(t, decimal.NewFromFloat(1.5).Equal(p.PlatformFeeUSDT))
	assert.True(t, decimal.NewFromFloat(2).Equal(p.TotalFeesUSDT))
	assert.True(t, decimal.NewFromFloat(98).Equal(p.NetAmountUSDT))
}
