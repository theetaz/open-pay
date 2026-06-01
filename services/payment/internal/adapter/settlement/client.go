// Package settlement is a thin client over the settlement service's internal
// API, used by the payment service to credit a merchant's balance when an
// on-chain payment is confirmed.
package settlement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Client calls the settlement service's internal API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new settlement service client.
func NewClient(settlementServiceURL string) *Client {
	return &Client{
		baseURL:    settlementServiceURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// CreditPayment credits a merchant's settlement balance with the net amount and
// records the platform fees. Calls POST /internal/settlements/credit.
func (c *Client) CreditPayment(ctx context.Context, merchantID uuid.UUID, netUSDT, feesUSDT decimal.Decimal) error {
	body, err := json.Marshal(map[string]string{
		"merchantId": merchantID.String(),
		"netUsdt":    netUSDT.String(),
		"feesUsdt":   feesUSDT.String(),
	})
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/settlements/credit", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling settlement service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("settlement service returned %d", resp.StatusCode)
	}
	return nil
}
