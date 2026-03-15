package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/exchange/internal/domain"
)

// RateRepository defines the data access contract for exchange rates.
type RateRepository interface {
	Create(ctx context.Context, rate *domain.ExchangeRate) error
	GetActive(ctx context.Context, base, quote string) (*domain.ExchangeRate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ExchangeRate, error)
	GetHistorical(ctx context.Context, base, quote string, from, to time.Time) ([]*domain.ExchangeRate, error)
	GetAtTime(ctx context.Context, base, quote string, at time.Time) (*domain.ExchangeRate, error)
}

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data any) error
}

// ExchangeService orchestrates exchange rate operations.
type ExchangeService struct {
	repo   RateRepository
	events EventPublisher
}

// NewExchangeService creates a new ExchangeService.
func NewExchangeService(repo RateRepository, events EventPublisher) *ExchangeService {
	return &ExchangeService{repo: repo, events: events}
}

// GetActiveRate returns the current active exchange rate for a currency pair.
func (s *ExchangeService) GetActiveRate(ctx context.Context, base, quote string) (*domain.ExchangeRate, error) {
	return s.repo.GetActive(ctx, base, quote)
}

// UpdateRate stores a new exchange rate and publishes an event.
func (s *ExchangeService) UpdateRate(ctx context.Context, base, quote string, rate decimal.Decimal, source string) error {
	exchangeRate, err := domain.NewExchangeRate(base, quote, rate, source)
	if err != nil {
		return fmt.Errorf("creating rate: %w", err)
	}

	if err := s.repo.Create(ctx, exchangeRate); err != nil {
		return fmt.Errorf("storing rate: %w", err)
	}

	_ = s.events.Publish(ctx, "exchange.rate.updated", exchangeRate)
	return nil
}

// GetSnapshot returns a frozen rate snapshot for recording with a payment.
func (s *ExchangeService) GetSnapshot(ctx context.Context, base, quote string) (*domain.RateSnapshot, error) {
	rate, err := s.repo.GetActive(ctx, base, quote)
	if err != nil {
		return nil, err
	}
	snapshot := rate.Snapshot()
	return &snapshot, nil
}

// GetHistoricalRates returns rates within a time range.
func (s *ExchangeService) GetHistoricalRates(ctx context.Context, base, quote string, from, to time.Time) ([]*domain.ExchangeRate, error) {
	return s.repo.GetHistorical(ctx, base, quote, from, to)
}

// GetRateAtTime returns the closest rate to a specific point in time.
func (s *ExchangeService) GetRateAtTime(ctx context.Context, base, quote string, at time.Time) (*domain.ExchangeRate, error) {
	return s.repo.GetAtTime(ctx, base, quote, at)
}

// ConvertLKRToUSDT converts an LKR amount to USDT using the active rate.
// Returns the USDT amount and the rate snapshot used.
func (s *ExchangeService) ConvertLKRToUSDT(ctx context.Context, lkrAmount decimal.Decimal) (decimal.Decimal, *domain.RateSnapshot, error) {
	rate, err := s.repo.GetActive(ctx, "USDT", "LKR")
	if err != nil {
		return decimal.Zero, nil, fmt.Errorf("getting USDT/LKR rate: %w", err)
	}

	usdt := rate.ConvertQuoteToBase(lkrAmount)
	snapshot := rate.Snapshot()
	return usdt, &snapshot, nil
}
