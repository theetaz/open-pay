package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// BybitConfig holds configuration for the Bybit provider.
type BybitConfig struct {
	BaseURL   string // e.g., "https://api.bybit.com"
	APIKey    string
	APISecret string
}

// BybitProvider implements PaymentProvider for Bybit Pay.
type BybitProvider struct {
	config     BybitConfig
	httpClient *http.Client
}

// NewBybitProvider creates a new Bybit payment provider.
func NewBybitProvider(cfg BybitConfig) *BybitProvider {
	return &BybitProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *BybitProvider) Name() string {
	return "BYBIT"
}

func (p *BybitProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	// TODO: Implement actual Bybit Pay API integration
	// Reference: https://bybit-exchange.github.io/docs/v5/otc/create-order
	//
	// The implementation should:
	// 1. Build the Bybit API request body with amount, currency, orderID
	// 2. Sign the request with HMAC-SHA256 using the API secret
	// 3. POST to {baseURL}/fiat/otc/item/create
	// 4. Parse the response to extract payID, QR content, checkout link, deep link
	//
	// For now, return an error indicating the provider is not yet configured.

	return nil, fmt.Errorf("bybit provider not configured: set BYBIT_API_KEY and BYBIT_API_SECRET environment variables")
}

func (p *BybitProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	// TODO: Implement actual Bybit Pay status check
	// Reference: https://bybit-exchange.github.io/docs/v5/otc/query-order
	//
	// The implementation should:
	// 1. Build the status query request with providerPayID
	// 2. Sign and send to {baseURL}/fiat/otc/item/query
	// 3. Map Bybit status to domain.PaymentStatus
	//   - Bybit "COMPLETE" → StatusPaid
	//   - Bybit "EXPIRED" → StatusExpired
	//   - Bybit "FAILED" → StatusFailed
	//   - Bybit "PENDING" → StatusInitiated

	return nil, fmt.Errorf("bybit provider not configured")
}

// bybitSign creates the HMAC-SHA256 signature for Bybit API requests.
// This is a helper for when the actual implementation is built.
func (p *BybitProvider) signRequest(timestamp, payload string) string {
	// TODO: Implement Bybit HMAC signing
	// message = timestamp + apiKey + payload
	// signature = HMAC-SHA256(message, apiSecret)
	_ = timestamp
	_ = payload
	return ""
}

// Ensure interface compliance at compile time.
var _ domain.PaymentProvider = (*BybitProvider)(nil)

// suppress unused import warnings for future use
var (
	_ = bytes.NewReader
	_ = json.Marshal
)
