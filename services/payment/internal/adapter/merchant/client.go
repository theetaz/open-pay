package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client calls the merchant service's internal API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new merchant service client.
func NewClient(merchantServiceURL string) *Client {
	return &Client{
		baseURL:    merchantServiceURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// IncrementPaymentLinkUsage increments the usage count for a payment link by slug.
func (c *Client) IncrementPaymentLinkUsage(ctx context.Context, slug string) error {
	body, err := json.Marshal(map[string]string{"slug": slug})
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/payment-links/increment-usage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling merchant service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("merchant service returned %d", resp.StatusCode)
	}

	return nil
}
