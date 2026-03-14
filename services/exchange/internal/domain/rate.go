package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidRate  = errors.New("invalid exchange rate")
	ErrRateNotFound = errors.New("exchange rate not found")
)

// ExchangeRate represents a currency pair exchange rate.
type ExchangeRate struct {
	ID            uuid.UUID
	BaseCurrency  string          // e.g., USDT
	QuoteCurrency string          // e.g., LKR
	Rate          decimal.Decimal // How many quote units per 1 base unit
	Source        string          // e.g., BINANCE, COINGECKO
	IsActive      bool
	FetchedAt     time.Time
	CreatedAt     time.Time
}

// RateSnapshot is a frozen copy of a rate at a specific point in time.
type RateSnapshot struct {
	RateID     uuid.UUID
	Rate       decimal.Decimal
	Source     string
	SnapshotAt time.Time
}

// NewExchangeRate creates a validated exchange rate.
func NewExchangeRate(base, quote string, rate decimal.Decimal, source string) (*ExchangeRate, error) {
	if base == "" {
		return nil, fmt.Errorf("%w: base currency is required", ErrInvalidRate)
	}
	if quote == "" {
		return nil, fmt.Errorf("%w: quote currency is required", ErrInvalidRate)
	}
	if rate.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: rate must be positive", ErrInvalidRate)
	}
	if source == "" {
		return nil, fmt.Errorf("%w: source is required", ErrInvalidRate)
	}

	now := time.Now().UTC()
	return &ExchangeRate{
		ID:            uuid.New(),
		BaseCurrency:  base,
		QuoteCurrency: quote,
		Rate:          rate,
		Source:        source,
		IsActive:      true,
		FetchedAt:     now,
		CreatedAt:     now,
	}, nil
}

// ConvertBaseToQuote converts an amount from base currency to quote currency.
// E.g., USDT → LKR: amount * rate.
func (r *ExchangeRate) ConvertBaseToQuote(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(r.Rate)
}

// ConvertQuoteToBase converts an amount from quote currency to base currency.
// E.g., LKR → USDT: amount / rate.
func (r *ExchangeRate) ConvertQuoteToBase(amount decimal.Decimal) decimal.Decimal {
	return amount.Div(r.Rate)
}

// Snapshot creates a frozen copy of this rate for recording with a payment.
func (r *ExchangeRate) Snapshot() RateSnapshot {
	return RateSnapshot{
		RateID:     r.ID,
		Rate:       r.Rate,
		Source:     r.Source,
		SnapshotAt: time.Now().UTC(),
	}
}

// IsStale returns true if the rate was fetched longer ago than maxAge.
func (r *ExchangeRate) IsStale(maxAge time.Duration) bool {
	return time.Since(r.FetchedAt) > maxAge
}
