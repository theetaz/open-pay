package fraud_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/pkg/fraud"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestEngine_LowRisk(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:          "merchant-1",
		Amount:              decimal.NewFromFloat(25),
		Currency:            "USDT",
		PaymentsLast1Hour:   2,
		PaymentsLast24Hours: 5,
		TotalAmountLast24H:  decimal.NewFromFloat(100),
		AverageAmount:       decimal.NewFromFloat(20),
	}

	result := engine.Assess(ctx)
	assert.Equal(t, fraud.RiskLow, result.Level)
	assert.False(t, result.ShouldBlock)
	assert.Empty(t, result.Flags)
}

func TestEngine_HighVelocity(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:          "merchant-1",
		Amount:              decimal.NewFromFloat(25),
		Currency:            "USDT",
		PaymentsLast1Hour:   12,
		PaymentsLast24Hours: 30,
		AverageAmount:       decimal.NewFromFloat(20),
	}

	result := engine.Assess(ctx)
	assert.True(t, result.Score >= 30)
	assert.Contains(t, result.Flags[0], "velocity")
}

func TestEngine_AmountAnomaly(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:    "merchant-1",
		Amount:        decimal.NewFromFloat(600), // 6x average
		Currency:      "USDT",
		AverageAmount: decimal.NewFromFloat(100),
	}

	result := engine.Assess(ctx)
	assert.True(t, result.Score >= 25)
	found := false
	for _, f := range result.Flags {
		if f != "" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestEngine_CriticalRisk(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:          "merchant-1",
		Amount:              decimal.NewFromFloat(15000), // Large amount
		Currency:            "USDT",
		PaymentsLast1Hour:   15,                                         // High velocity
		PaymentsLast24Hours: 55,                                         // Excessive daily
		TotalAmountLast24H:  decimal.NewFromFloat(60000),                // High volume
		AverageAmount:       decimal.NewFromFloat(100),                  // Huge anomaly
		IsNewCustomer:       true,
	}

	result := engine.Assess(ctx)
	assert.Equal(t, fraud.RiskCritical, result.Level)
	assert.True(t, result.ShouldBlock)
	assert.True(t, len(result.Flags) >= 3)
}

func TestEngine_NewCustomerLargeAmount(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:    "merchant-1",
		Amount:        decimal.NewFromFloat(2000),
		Currency:      "USDT",
		IsNewCustomer: true,
		AverageAmount: decimal.NewFromFloat(500),
	}

	result := engine.Assess(ctx)
	assert.True(t, result.Score >= 15)
}

func TestEngine_BlockThreshold(t *testing.T) {
	engine := fraud.NewEngine()
	engine.SetBlockThreshold(50)

	ctx := fraud.PaymentContext{
		MerchantID:          "merchant-1",
		Amount:              decimal.NewFromFloat(15000),
		Currency:            "USDT",
		PaymentsLast1Hour:   12,
		PaymentsLast24Hours: 55,
		TotalAmountLast24H:  decimal.NewFromFloat(60000),
		AverageAmount:       decimal.NewFromFloat(100),
	}

	result := engine.Assess(ctx)
	assert.True(t, result.ShouldBlock)
}

func TestEngine_ScoreCappedAt100(t *testing.T) {
	engine := fraud.NewEngine()
	ctx := fraud.PaymentContext{
		MerchantID:          "merchant-1",
		Amount:              decimal.NewFromFloat(20000),
		Currency:            "USDT",
		PaymentsLast1Hour:   20,
		PaymentsLast24Hours: 100,
		TotalAmountLast24H:  decimal.NewFromFloat(100000),
		AverageAmount:       decimal.NewFromFloat(10),
		IsNewCustomer:       true,
	}

	result := engine.Assess(ctx)
	assert.LessOrEqual(t, result.Score, 100)
}
