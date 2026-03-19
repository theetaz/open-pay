package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// BybitConfig holds configuration for the Bybit provider.
type BybitConfig struct {
	BaseURL   string // e.g., "https://api-testnet.bybit.com"
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

// --- request/response structs ---

type bybitCreateReq struct {
	TokenID    string         `json:"tokenId"`
	Amount     string         `json:"amount"`
	OutTradeNo string         `json:"outTradeNo"`
	Env        bybitEnvDetail `json:"env"`
}

type bybitEnvDetail struct {
	TerminalType string `json:"terminalType"`
}

type bybitCreateResp struct {
	RetCode int              `json:"retCode"`
	RetMsg  string           `json:"retMsg"`
	Result  bybitCreateData  `json:"result"`
}

type bybitCreateData struct {
	PayID       string `json:"payId"`
	QRCodeURL   string `json:"qrCodeUrl"`
	CheckoutURL string `json:"checkoutUrl"`
	DeepLink    string `json:"deepLink"`
}

type bybitQueryReq struct {
	PayID string `json:"payId"`
}

type bybitQueryResp struct {
	RetCode int             `json:"retCode"`
	RetMsg  string          `json:"retMsg"`
	Result  bybitQueryData  `json:"result"`
}

type bybitQueryData struct {
	Status string `json:"status"`
	TxHash string `json:"txHash"`
}

// --- signing ---

func (p *BybitProvider) sign(timestamp, body string) string {
	recvWindow := "5000"
	message := timestamp + p.config.APIKey + recvWindow + body
	mac := hmac.New(sha256.New, []byte(p.config.APISecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- HTTP helper ---

func (p *BybitProvider) doRequest(ctx context.Context, url string, body []byte) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sig := p.sign(timestamp, string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bybit: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", p.config.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-SIGN", sig)
	req.Header.Set("X-BAPI-RECV-WINDOW", "5000")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bybit: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bybit: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bybit: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// --- PaymentProvider interface ---

func (p *BybitProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	payload := bybitCreateReq{
		TokenID:    req.Currency,
		Amount:     req.Amount,
		OutTradeNo: req.OrderID,
		Env:        bybitEnvDetail{TerminalType: "WEB"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bybit: marshal request: %w", err)
	}

	respBody, err := p.doRequest(ctx, p.config.BaseURL+"/fiat/otc/item/create", body)
	if err != nil {
		return nil, err
	}

	var apiResp bybitCreateResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("bybit: unmarshal response: %w", err)
	}
	if apiResp.RetCode != 0 {
		return nil, fmt.Errorf("bybit: API error %d: %s", apiResp.RetCode, apiResp.RetMsg)
	}

	return &domain.ProviderPaymentResponse{
		ProviderPayID: apiResp.Result.PayID,
		QRContent:     apiResp.Result.QRCodeURL,
		CheckoutLink:  apiResp.Result.CheckoutURL,
		DeepLink:      apiResp.Result.DeepLink,
	}, nil
}

func (p *BybitProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	payload := bybitQueryReq{PayID: providerPayID}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bybit: marshal request: %w", err)
	}

	respBody, err := p.doRequest(ctx, p.config.BaseURL+"/fiat/otc/item/query", body)
	if err != nil {
		return nil, err
	}

	var apiResp bybitQueryResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("bybit: unmarshal response: %w", err)
	}
	if apiResp.RetCode != 0 {
		return nil, fmt.Errorf("bybit: API error %d: %s", apiResp.RetCode, apiResp.RetMsg)
	}

	return &domain.ProviderPaymentStatus{
		Status: mapBybitStatus(apiResp.Result.Status),
		TxHash: apiResp.Result.TxHash,
	}, nil
}

func mapBybitStatus(s string) domain.PaymentStatus {
	switch s {
	case "COMPLETE", "SUCCESS":
		return domain.StatusPaid
	case "EXPIRED":
		return domain.StatusExpired
	case "FAILED", "CANCELLED":
		return domain.StatusFailed
	default:
		return domain.StatusInitiated
	}
}

// Ensure interface compliance at compile time.
var _ domain.PaymentProvider = (*BybitProvider)(nil)
