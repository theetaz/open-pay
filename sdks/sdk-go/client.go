// Package openpay provides a Go SDK for the Open Pay payment processing API.
//
// Usage:
//
//	client, err := openpay.NewClient("ak_live_xxx.sk_live_yyy",
//	    openpay.WithBaseURL("https://olp-api.nipuntheekshana.com"),
//	)
//
//	payment, err := client.Payments.Create(ctx, openpay.CreatePaymentInput{
//	    Amount:          "1000.00",
//	    Currency:        "LKR",
//	    MerchantTradeNo: "ORDER-123",
//	})
package openpay

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openlankapay/openlankapay/pkg/auth"
)

const (
	defaultBaseURL = "https://api.openpay.lk"
	sandboxBaseURL = "https://sandbox-api.openpay.lk"
)

// Client is the Open Pay API client.
type Client struct {
	keyID      string
	secret     string
	baseURL    string
	httpClient *http.Client

	// Resource services
	Payments *PaymentsService
	Checkout *CheckoutService
	Webhooks *WebhooksService
}

// Option configures the client.
type Option func(*Client)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(url, "/") }
}

// WithSandbox configures the client to use the sandbox environment.
func WithSandbox() Option {
	return func(c *Client) { c.baseURL = sandboxBaseURL }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a new Open Pay API client.
// apiKey format: "ak_{env}_{id}.sk_{env}_{secret}"
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	keyID, secret, err := auth.ParseAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	c := &Client{
		keyID:      keyID,
		secret:     secret,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Payments = &PaymentsService{client: c}
	c.Checkout = &CheckoutService{client: c}
	c.Webhooks = &WebhooksService{client: c}

	return c, nil
}

// SignHeaders generates the authentication headers for an API request.
func (c *Client) SignHeaders(method, path, body string) map[string]string {
	timestamp := auth.CurrentTimestamp()
	signature := auth.SignRequest(c.secret, timestamp, strings.ToUpper(method), path, body)

	return map[string]string{
		"x-api-key":   c.keyID,
		"x-timestamp": timestamp,
		"x-signature": signature,
	}
}

// doRequest executes an authenticated API request.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyStr string
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyStr = string(data)
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.SignHeaders(method, path, bodyStr) {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(respBody, &errResp)
		return &APIError{
			Code:    errResp.Error.Code,
			Message: errResp.Error.Message,
			Status:  resp.StatusCode,
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// VerifyWebhookSignature verifies an ED25519-signed webhook payload.
func VerifyWebhookSignature(publicKeyBase64, timestamp string, payload []byte, signatureBase64 string) error {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return fmt.Errorf("decoding public key: %w", err)
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}

	message := append([]byte(timestamp), payload...)
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), message, sigBytes) {
		return errors.New("signature verification failed")
	}

	return nil
}
