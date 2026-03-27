package money

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAmount   = errors.New("amount must be positive")
	ErrInvalidCurrency = errors.New("unsupported currency")
)

// Supported currencies.
var validCurrencies = map[string]bool{
	"USDT": true,
	"USDC": true,
	"BTC":  true,
	"ETH":  true,
	"BNB":  true,
	"LKR":  true,
}

// IsValidCurrency checks if a currency code is supported.
func IsValidCurrency(currency string) bool {
	return validCurrencies[currency]
}

// Amount represents a monetary value with its currency.
type Amount struct {
	Value    decimal.Decimal
	Currency string
}

// NewAmount creates a validated Amount from a string value and currency.
func NewAmount(value string, currency string) (Amount, error) {
	if !IsValidCurrency(currency) {
		return Amount{}, fmt.Errorf("%w: %s", ErrInvalidCurrency, currency)
	}

	d, err := decimal.NewFromString(value)
	if err != nil {
		return Amount{}, fmt.Errorf("invalid number format: %w", err)
	}

	if d.LessThanOrEqual(decimal.Zero) {
		return Amount{}, ErrInvalidAmount
	}

	return Amount{Value: d, Currency: currency}, nil
}

// FeeBreakdown contains the detailed fee calculation for a payment.
type FeeBreakdown struct {
	GrossAmount        decimal.Decimal
	ExchangeFeePercent decimal.Decimal
	ExchangeFeeAmount  decimal.Decimal
	PlatformFeePercent decimal.Decimal
	PlatformFeeAmount  decimal.Decimal
	TotalFees          decimal.Decimal
	NetAmount          decimal.Decimal
}

// CalculateFees computes the fee breakdown for a given amount.
// Fee rates are percentages (e.g., 0.5 means 0.5%, 1.5 means 1.5%).
func CalculateFees(amount, exchangeFeeRate, platformFeeRate decimal.Decimal) FeeBreakdown {
	hundred := decimal.NewFromInt(100)

	exchangeFee := amount.Mul(exchangeFeeRate).Div(hundred)
	platformFee := amount.Mul(platformFeeRate).Div(hundred)
	totalFees := exchangeFee.Add(platformFee)
	netAmount := amount.Sub(totalFees)

	return FeeBreakdown{
		GrossAmount:        amount,
		ExchangeFeePercent: exchangeFeeRate,
		ExchangeFeeAmount:  exchangeFee,
		PlatformFeePercent: platformFeeRate,
		PlatformFeeAmount:  platformFee,
		TotalFees:          totalFees,
		NetAmount:          netAmount,
	}
}

// ConvertToUSDT converts an LKR amount to USDT using the given rate.
// rate = how many LKR per 1 USDT.
func ConvertToUSDT(lkr, rate decimal.Decimal) decimal.Decimal {
	return lkr.Div(rate)
}

// ConvertToLKR converts a USDT amount to LKR using the given rate.
// rate = how many LKR per 1 USDT.
func ConvertToLKR(usdt, rate decimal.Decimal) decimal.Decimal {
	return usdt.Mul(rate)
}
