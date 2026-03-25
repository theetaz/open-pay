package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// RateUpdater is the interface for updating exchange rates.
type RateUpdater interface {
	UpdateRate(ctx context.Context, base, quote string, rate decimal.Decimal, source string) error
}

// CoinGeckoFetcher periodically fetches USDT/LKR rate from CoinGecko.
type CoinGeckoFetcher struct {
	svc      RateUpdater
	logger   zerolog.Logger
	interval time.Duration
	client   *http.Client
}

// NewCoinGeckoFetcher creates a new rate fetcher.
func NewCoinGeckoFetcher(svc RateUpdater, logger zerolog.Logger, interval time.Duration) *CoinGeckoFetcher {
	return &CoinGeckoFetcher{
		svc:      svc,
		logger:   logger,
		interval: interval,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Start begins periodic rate fetching. Blocks until ctx is cancelled.
func (f *CoinGeckoFetcher) Start(ctx context.Context) {
	// Fetch immediately on start
	f.fetchAndUpdate(ctx)

	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Info().Msg("rate fetcher stopped")
			return
		case <-ticker.C:
			f.fetchAndUpdate(ctx)
		}
	}
}

func (f *CoinGeckoFetcher) fetchAndUpdate(ctx context.Context) {
	rate, err := f.fetchUSDTLKR(ctx)
	if err != nil {
		f.logger.Warn().Err(err).Msg("failed to fetch USDT/LKR rate from CoinGecko")
		return
	}

	if err := f.svc.UpdateRate(ctx, "USDT", "LKR", rate, "COINGECKO"); err != nil {
		f.logger.Warn().Err(err).Msg("failed to update USDT/LKR rate")
		return
	}

	f.logger.Info().Str("rate", rate.String()).Msg("updated USDT/LKR rate from CoinGecko")
}

// fetchUSDTLKR gets the USDT price in LKR from CoinGecko's free API.
// CoinGecko uses "tether" for USDT and "lkr" for Sri Lankan Rupee.
func (f *CoinGeckoFetcher) fetchUSDTLKR(ctx context.Context) (decimal.Decimal, error) {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=lkr"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("creating request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("fetching rate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Response: {"tether":{"lkr":325.50}}
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return decimal.Zero, fmt.Errorf("decoding response: %w", err)
	}

	tether, ok := result["tether"]
	if !ok {
		return decimal.Zero, fmt.Errorf("missing tether data in response")
	}
	lkr, ok := tether["lkr"]
	if !ok {
		return decimal.Zero, fmt.Errorf("missing lkr rate in response")
	}

	rate := decimal.NewFromFloat(lkr)
	if rate.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("invalid rate: %s", rate.String())
	}

	return rate, nil
}
