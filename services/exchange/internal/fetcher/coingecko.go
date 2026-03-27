package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// RateUpdater is the interface for updating exchange rates.
type RateUpdater interface {
	UpdateRate(ctx context.Context, base, quote string, rate decimal.Decimal, source string) error
}

// cryptoPair maps our currency code to CoinGecko's coin ID.
var cryptoPairs = []struct {
	Base       string // Our currency code
	CoinGeckoID string // CoinGecko API ID
}{
	{"USDT", "tether"},
	{"USDC", "usd-coin"},
	{"BTC", "bitcoin"},
	{"ETH", "ethereum"},
	{"BNB", "binancecoin"},
}

// CoinGeckoFetcher periodically fetches crypto/LKR rates from CoinGecko.
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
	rates, err := f.fetchAllRates(ctx)
	if err != nil {
		f.logger.Warn().Err(err).Msg("failed to fetch rates from CoinGecko")
		return
	}

	for base, lkrRate := range rates {
		if err := f.svc.UpdateRate(ctx, base, "LKR", lkrRate, "COINGECKO"); err != nil {
			f.logger.Warn().Err(err).Str("pair", base+"/LKR").Msg("failed to update rate")
			continue
		}
		f.logger.Info().Str("pair", base+"/LKR").Str("rate", lkrRate.String()).Msg("updated rate from CoinGecko")
	}
}

// fetchAllRates gets prices for all supported cryptocurrencies in LKR from CoinGecko.
func (f *CoinGeckoFetcher) fetchAllRates(ctx context.Context) (map[string]decimal.Decimal, error) {
	// Build comma-separated list of CoinGecko IDs
	ids := make([]string, len(cryptoPairs))
	for i, p := range cryptoPairs {
		ids[i] = p.CoinGeckoID
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=lkr", strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching rates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Response: {"tether":{"lkr":325.50},"bitcoin":{"lkr":30000000},...}
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	rates := make(map[string]decimal.Decimal)
	for _, pair := range cryptoPairs {
		coinData, ok := result[pair.CoinGeckoID]
		if !ok {
			f.logger.Warn().Str("coin", pair.CoinGeckoID).Msg("missing data in CoinGecko response")
			continue
		}
		lkr, ok := coinData["lkr"]
		if !ok {
			f.logger.Warn().Str("coin", pair.CoinGeckoID).Msg("missing LKR rate in response")
			continue
		}
		rate := decimal.NewFromFloat(lkr)
		if rate.LessThanOrEqual(decimal.Zero) {
			f.logger.Warn().Str("coin", pair.CoinGeckoID).Msg("invalid rate (<=0)")
			continue
		}
		rates[pair.Base] = rate
	}

	return rates, nil
}
