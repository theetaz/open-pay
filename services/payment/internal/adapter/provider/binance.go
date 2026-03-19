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

// BinanceConfig holds configuration for the Binance Pay provider.
type BinanceConfig struct {
	BaseURL   string // e.g., "https://bpay.binanceapi.com"
	APIKey    string
	APISecret string
}

// BinanceProvider implements PaymentProvider for Binance Pay.
type BinanceProvider struct {
	config     BinanceConfig
	httpClient *http.Client
}

// NewBinanceProvider creates a new Binance Pay payment provider.
func NewBinanceProvider(cfg BinanceConfig) *BinanceProvider {
	return &BinanceProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *BinanceProvider) Name() string {
	return "BINANCE"
}

func (p *BinanceProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	// TODO: Implement actual Binance Pay API integration
	// Reference: https://developers.binance.com/docs/binance-pay/api-order-v2
	//
	// The implementation should:
	// 1. Build the request body with amount, currency, merchantTradeNo
	// 2. Sign with HMAC-SHA512 using the API secret
	// 3. POST to {baseURL}/binancepay/openapi/v2/order
	// 4. Parse the response for qrcodeLink, checkoutUrl, deeplink, prepayId

	return nil, fmt.Errorf("binance provider not configured: set BINANCE_API_KEY and BINANCE_API_SECRET environment variables")
}

func (p *BinanceProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	// TODO: Implement actual Binance Pay status check
	// Reference: https://developers.binance.com/docs/binance-pay/api-order-query-v2
	//
	// Status mapping:
	//   - Binance "SUCCESS" → StatusPaid
	//   - Binance "EXPIRED" → StatusExpired
	//   - Binance "FAILED" / "REFUNDED" → StatusFailed
	//   - Binance "INITIAL" → StatusInitiated

	return nil, fmt.Errorf("binance provider not configured")
}

var _ domain.PaymentProvider = (*BinanceProvider)(nil)

var (
	_ = bytes.NewReader
	_ = json.Marshal
)
