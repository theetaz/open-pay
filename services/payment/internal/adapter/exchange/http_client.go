package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// Client calls the exchange service to get current rates.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an exchange service HTTP client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type rateResponse struct {
	Data struct {
		Rate string `json:"rate"`
	} `json:"data"`
}

// GetRate fetches the current active exchange rate for a currency pair.
func (c *Client) GetRate(ctx context.Context, base, quote string) (decimal.Decimal, error) {
	url := fmt.Sprintf("%s/v1/exchange-rates/active?base=%s&quote=%s", c.baseURL, base, quote)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("fetching exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("exchange service returned status %d", resp.StatusCode)
	}

	var rateResp rateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		return decimal.Zero, fmt.Errorf("decoding response: %w", err)
	}

	rate, err := decimal.NewFromString(rateResp.Data.Rate)
	if err != nil {
		return decimal.Zero, fmt.Errorf("parsing rate: %w", err)
	}

	return rate, nil
}
