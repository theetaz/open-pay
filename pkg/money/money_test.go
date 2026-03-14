package money_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/pkg/money"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyValidation(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		valid    bool
	}{
		{"USDT is valid", "USDT", true},
		{"LKR is valid", "LKR", true},
		{"BTC is valid", "BTC", true},
		{"ETH is valid", "ETH", true},
		{"empty is invalid", "", false},
		{"random string is invalid", "FAKE", false},
		{"lowercase is invalid", "usdt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, money.IsValidCurrency(tt.currency))
		})
	}
}

func TestNewAmount(t *testing.T) {
	t.Run("valid amount", func(t *testing.T) {
		amt, err := money.NewAmount("10.50", "USDT")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(10.50).Equal(amt.Value))
		assert.Equal(t, "USDT", amt.Currency)
	})

	t.Run("zero amount is invalid", func(t *testing.T) {
		_, err := money.NewAmount("0", "USDT")
		require.Error(t, err)
		assert.ErrorIs(t, err, money.ErrInvalidAmount)
	})

	t.Run("negative amount is invalid", func(t *testing.T) {
		_, err := money.NewAmount("-5", "USDT")
		require.Error(t, err)
		assert.ErrorIs(t, err, money.ErrInvalidAmount)
	})

	t.Run("invalid currency", func(t *testing.T) {
		_, err := money.NewAmount("10", "FAKE")
		require.Error(t, err)
		assert.ErrorIs(t, err, money.ErrInvalidCurrency)
	})

	t.Run("invalid number format", func(t *testing.T) {
		_, err := money.NewAmount("abc", "USDT")
		require.Error(t, err)
	})
}

func TestFeeCalculation(t *testing.T) {
	tests := []struct {
		name            string
		amount          string
		exchangeFeeRate string
		platformFeeRate string
		wantExchangeFee string
		wantPlatformFee string
		wantTotalFees   string
		wantNet         string
	}{
		{
			name:            "standard 2% total fee on 100 USDT",
			amount:          "100",
			exchangeFeeRate: "0.5",
			platformFeeRate: "1.5",
			wantExchangeFee: "0.5",
			wantPlatformFee: "1.5",
			wantTotalFees:   "2",
			wantNet:         "98",
		},
		{
			name:            "fees on 10 USDT",
			amount:          "10",
			exchangeFeeRate: "0.5",
			platformFeeRate: "1.5",
			wantExchangeFee: "0.05",
			wantPlatformFee: "0.15",
			wantTotalFees:   "0.2",
			wantNet:         "9.8",
		},
		{
			name:            "fees on small amount 0.50 USDT",
			amount:          "0.50",
			exchangeFeeRate: "0.5",
			platformFeeRate: "1.5",
			wantExchangeFee: "0.0025",
			wantPlatformFee: "0.0075",
			wantTotalFees:   "0.01",
			wantNet:         "0.49",
		},
		{
			name:            "zero exchange fee",
			amount:          "100",
			exchangeFeeRate: "0",
			platformFeeRate: "2",
			wantExchangeFee: "0",
			wantPlatformFee: "2",
			wantTotalFees:   "2",
			wantNet:         "98",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := decimal.RequireFromString(tt.amount)
			exchangeRate := decimal.RequireFromString(tt.exchangeFeeRate)
			platformRate := decimal.RequireFromString(tt.platformFeeRate)

			breakdown := money.CalculateFees(amount, exchangeRate, platformRate)

			assert.True(t, decimal.RequireFromString(tt.wantExchangeFee).Equal(breakdown.ExchangeFeeAmount),
				"exchange fee: got %s, want %s", breakdown.ExchangeFeeAmount, tt.wantExchangeFee)
			assert.True(t, decimal.RequireFromString(tt.wantPlatformFee).Equal(breakdown.PlatformFeeAmount),
				"platform fee: got %s, want %s", breakdown.PlatformFeeAmount, tt.wantPlatformFee)
			assert.True(t, decimal.RequireFromString(tt.wantTotalFees).Equal(breakdown.TotalFees),
				"total fees: got %s, want %s", breakdown.TotalFees, tt.wantTotalFees)
			assert.True(t, decimal.RequireFromString(tt.wantNet).Equal(breakdown.NetAmount),
				"net amount: got %s, want %s", breakdown.NetAmount, tt.wantNet)
		})
	}
}

func TestFeeBreakdownInvariant(t *testing.T) {
	// Property: gross = net + totalFees ALWAYS
	amounts := []string{"0.01", "1", "10", "100", "999.99", "50000"}
	for _, amtStr := range amounts {
		amount := decimal.RequireFromString(amtStr)
		breakdown := money.CalculateFees(
			amount,
			decimal.NewFromFloat(0.5),
			decimal.NewFromFloat(1.5),
		)

		reconstructed := breakdown.NetAmount.Add(breakdown.TotalFees)
		assert.True(t, amount.Equal(reconstructed),
			"invariant violated for %s: net(%s) + fees(%s) = %s",
			amtStr, breakdown.NetAmount, breakdown.TotalFees, reconstructed)
	}
}

func TestConvertCurrency(t *testing.T) {
	t.Run("LKR to USDT", func(t *testing.T) {
		lkr := decimal.NewFromFloat(32500)
		rate := decimal.NewFromFloat(325) // 1 USDT = 325 LKR
		usdt := money.ConvertToUSDT(lkr, rate)
		assert.True(t, decimal.NewFromFloat(100).Equal(usdt))
	})

	t.Run("USDT to LKR", func(t *testing.T) {
		usdt := decimal.NewFromFloat(100)
		rate := decimal.NewFromFloat(325)
		lkr := money.ConvertToLKR(usdt, rate)
		assert.True(t, decimal.NewFromFloat(32500).Equal(lkr))
	})
}
