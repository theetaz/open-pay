// Package openpay provides a Go SDK for the Open Pay payment processing API.
package openpay

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/openlankapay/openlankapay/pkg/auth"
)

const (
	defaultBaseURL   = "https://api.openpay.lk"
	sandboxBaseURL   = "https://sandbox-api.openpay.lk"
)

// Client is the Open Pay API client.
type Client struct {
	keyID   string
	secret  string
	baseURL string
}

// Option configures the client.
type Option func(*Client)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithSandbox configures the client to use the sandbox environment.
func WithSandbox() Option {
	return func(c *Client) { c.baseURL = sandboxBaseURL }
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
		keyID:   keyID,
		secret:  secret,
		baseURL: defaultBaseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

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
