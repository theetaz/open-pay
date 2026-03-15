package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/exchange/internal/domain"
)

// RateRepository is the PostgreSQL implementation for exchange rate persistence.
type RateRepository struct {
	pool *pgxpool.Pool
}

// NewRateRepository creates a new PostgreSQL-backed RateRepository.
func NewRateRepository(pool *pgxpool.Pool) *RateRepository {
	return &RateRepository{pool: pool}
}

func (r *RateRepository) Create(ctx context.Context, rate *domain.ExchangeRate) error {
	// Deactivate previous active rates for this pair
	deactivateQuery := `UPDATE exchange_rates SET is_active = FALSE
		WHERE base_currency = $1 AND quote_currency = $2 AND is_active = TRUE`
	_, _ = r.pool.Exec(ctx, deactivateQuery, rate.BaseCurrency, rate.QuoteCurrency)

	query := `INSERT INTO exchange_rates (id, base_currency, quote_currency, rate, source, is_active, fetched_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		rate.ID, rate.BaseCurrency, rate.QuoteCurrency,
		rate.Rate, rate.Source, rate.IsActive,
		rate.FetchedAt, rate.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting exchange rate: %w", err)
	}
	return nil
}

func (r *RateRepository) GetActive(ctx context.Context, base, quote string) (*domain.ExchangeRate, error) {
	query := `SELECT id, base_currency, quote_currency, rate, source, is_active, fetched_at, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND quote_currency = $2 AND is_active = TRUE
		ORDER BY fetched_at DESC LIMIT 1`

	return r.scanOne(ctx, query, base, quote)
}

func (r *RateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ExchangeRate, error) {
	query := `SELECT id, base_currency, quote_currency, rate, source, is_active, fetched_at, created_at
		FROM exchange_rates WHERE id = $1`

	return r.scanOne(ctx, query, id)
}

func (r *RateRepository) GetHistorical(ctx context.Context, base, quote string, from, to time.Time) ([]*domain.ExchangeRate, error) {
	query := `SELECT id, base_currency, quote_currency, rate, source, is_active, fetched_at, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND quote_currency = $2
		AND fetched_at BETWEEN $3 AND $4
		ORDER BY fetched_at DESC`

	rows, err := r.pool.Query(ctx, query, base, quote, from, to)
	if err != nil {
		return nil, fmt.Errorf("listing historical rates: %w", err)
	}
	defer rows.Close()

	var rates []*domain.ExchangeRate
	for rows.Next() {
		rate, err := scanRate(rows)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

func (r *RateRepository) GetAtTime(ctx context.Context, base, quote string, at time.Time) (*domain.ExchangeRate, error) {
	query := `SELECT id, base_currency, quote_currency, rate, source, is_active, fetched_at, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND quote_currency = $2 AND fetched_at <= $3
		ORDER BY fetched_at DESC LIMIT 1`

	return r.scanOne(ctx, query, base, quote, at)
}

func (r *RateRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.ExchangeRate, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rate: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrRateNotFound
	}

	return scanRate(rows)
}

func scanRate(rows pgx.Rows) (*domain.ExchangeRate, error) {
	var rate domain.ExchangeRate
	err := rows.Scan(
		&rate.ID, &rate.BaseCurrency, &rate.QuoteCurrency,
		&rate.Rate, &rate.Source, &rate.IsActive,
		&rate.FetchedAt, &rate.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning rate: %w", err)
	}
	return &rate, nil
}
