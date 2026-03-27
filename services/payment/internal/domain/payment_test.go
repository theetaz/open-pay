package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSplits_EmptySlice(t *testing.T) {
	err := ValidateSplits(nil)
	assert.NoError(t, err)
}

func TestValidateSplits_Valid5050(t *testing.T) {
	splits := []SplitRule{
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(50)},
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(50)},
	}
	err := ValidateSplits(splits)
	assert.NoError(t, err)
}

func TestValidateSplits_ValidUnderHundred(t *testing.T) {
	splits := []SplitRule{
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(30)},
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(20)},
	}
	err := ValidateSplits(splits)
	assert.NoError(t, err)
}

func TestValidateSplits_ExceedsHundred(t *testing.T) {
	splits := []SplitRule{
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(60)},
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(50)},
	}
	err := ValidateSplits(splits)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSplitPercentage)
}

func TestValidateSplits_ZeroPercentage(t *testing.T) {
	splits := []SplitRule{
		{MerchantID: uuid.New(), Percentage: decimal.Zero},
	}
	err := ValidateSplits(splits)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSplitPercentage)
}

func TestValidateSplits_NegativePercentage(t *testing.T) {
	splits := []SplitRule{
		{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(-10)},
	}
	err := ValidateSplits(splits)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSplitPercentage)
}

func TestNewPayment_WithValidSplits(t *testing.T) {
	input := CreatePaymentInput{
		MerchantID:      uuid.New(),
		Amount:          decimal.NewFromInt(100),
		Currency:        "USDT",
		Provider:        "BYBIT",
		MerchantTradeNo: "TXN-001",
		Splits: []SplitRule{
			{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(50)},
			{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(50)},
		},
	}

	payment, err := NewPayment(input)
	require.NoError(t, err)
	assert.Len(t, payment.Splits, 2)
}

func TestNewPayment_WithInvalidSplits(t *testing.T) {
	input := CreatePaymentInput{
		MerchantID:      uuid.New(),
		Amount:          decimal.NewFromInt(100),
		Currency:        "USDT",
		Provider:        "BYBIT",
		MerchantTradeNo: "TXN-001",
		Splits: []SplitRule{
			{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(70)},
			{MerchantID: uuid.New(), Percentage: decimal.NewFromInt(40)},
		},
	}

	_, err := NewPayment(input)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSplitPercentage)
}
