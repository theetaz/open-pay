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

// KuCoinConfig holds configuration for the KuCoin provider.
type KuCoinConfig struct {
	BaseURL    string // e.g., "https://openapi-v2.kucoin.com"
	APIKey     string
	APISecret  string
	Passphrase string
}

// KuCoinProvider implements PaymentProvider for KuCoin.
type KuCoinProvider struct {
	config     KuCoinConfig
	httpClient *http.Client
}

// NewKuCoinProvider creates a new KuCoin payment provider.
func NewKuCoinProvider(cfg KuCoinConfig) *KuCoinProvider {
	return &KuCoinProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *KuCoinProvider) Name() string {
	return "KUCOIN"
}

func (p *KuCoinProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	// TODO: Implement actual KuCoin Pay API integration
	// Reference: https://www.kucoin.com/docs/rest/other/kucoin-pay/create-order
	//
	// The implementation should:
	// 1. Build the request body with amount, currency, orderId
	// 2. Sign with HMAC-SHA256 using the API secret and passphrase
	// 3. POST to {baseURL}/api/v1/kucoin-pay/order
	// 4. Parse the response for QR code, checkout link, deep link

	return nil, fmt.Errorf("kucoin provider not configured: set KUCOIN_API_KEY and KUCOIN_API_SECRET environment variables")
}

func (p *KuCoinProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	// TODO: Implement actual KuCoin status check
	//
	// Status mapping:
	//   - KuCoin "SUCCESS" → StatusPaid
	//   - KuCoin "EXPIRED" → StatusExpired
	//   - KuCoin "FAILED" → StatusFailed
	//   - KuCoin "PENDING" → StatusInitiated

	return nil, fmt.Errorf("kucoin provider not configured")
}

var _ domain.PaymentProvider = (*KuCoinProvider)(nil)

var (
	_ = bytes.NewReader
	_ = json.Marshal
)
