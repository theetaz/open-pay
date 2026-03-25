package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// KuCoinConfig holds configuration for the KuCoin provider.
type KuCoinConfig struct {
	BaseURL    string // e.g., "https://openapi-sandbox.kucoin.com"
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

// --- request/response structs ---

type kucoinCreateReq struct {
	MerchantOrderID string `json:"merchantOrderId"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
}

type kucoinAPIResp struct {
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

type kucoinCreateData struct {
	OrderID      string `json:"orderId"`
	QRCode       string `json:"qrCode"`
	CheckoutLink string `json:"checkoutLink"`
	DeepLink     string `json:"deepLink"`
}

type kucoinQueryData struct {
	Status string `json:"status"`
	TxHash string `json:"txHash"`
}

// --- signing ---

func (p *KuCoinProvider) signHMAC(message string) string {
	mac := hmac.New(sha256.New, []byte(p.config.APISecret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (p *KuCoinProvider) signPassphrase() string {
	mac := hmac.New(sha256.New, []byte(p.config.APISecret))
	mac.Write([]byte(p.config.Passphrase))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// --- HTTP helper ---

func (p *KuCoinProvider) doRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	message := timestamp + method + endpoint + string(body)
	sig := p.signHMAC(message)
	passphrase := p.signPassphrase()

	fullURL := p.config.BaseURL + endpoint

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("kucoin: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("KC-API-KEY", p.config.APIKey)
	req.Header.Set("KC-API-SIGN", sig)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", passphrase)
	req.Header.Set("KC-API-KEY-VERSION", "2")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kucoin: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kucoin: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kucoin: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// --- PaymentProvider interface ---

func (p *KuCoinProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	payload := kucoinCreateReq{
		MerchantOrderID: req.OrderID,
		Amount:          req.Amount,
		Currency:        req.Currency,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("kucoin: marshal request: %w", err)
	}

	endpoint := "/api/v1/kucoin-pay/order"
	respBody, err := p.doRequest(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}

	var apiResp kucoinAPIResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("kucoin: unmarshal response: %w", err)
	}
	if apiResp.Code != "200000" {
		return nil, fmt.Errorf("kucoin: API error code=%s msg=%s", apiResp.Code, apiResp.Msg)
	}

	var data kucoinCreateData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("kucoin: unmarshal data: %w", err)
	}

	return &domain.ProviderPaymentResponse{
		ProviderPayID: data.OrderID,
		QRContent:     data.QRCode,
		CheckoutLink:  data.CheckoutLink,
		DeepLink:      data.DeepLink,
	}, nil
}

func (p *KuCoinProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	endpoint := "/api/v1/kucoin-pay/order/" + providerPayID
	respBody, err := p.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var apiResp kucoinAPIResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("kucoin: unmarshal response: %w", err)
	}
	if apiResp.Code != "200000" {
		return nil, fmt.Errorf("kucoin: API error code=%s msg=%s", apiResp.Code, apiResp.Msg)
	}

	var data kucoinQueryData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("kucoin: unmarshal data: %w", err)
	}

	return &domain.ProviderPaymentStatus{
		Status: mapKuCoinStatus(data.Status),
		TxHash: data.TxHash,
	}, nil
}

func mapKuCoinStatus(s string) domain.PaymentStatus {
	switch s {
	case "SUCCESS":
		return domain.StatusPaid
	case "EXPIRED":
		return domain.StatusExpired
	case "FAILED":
		return domain.StatusFailed
	default:
		return domain.StatusInitiated
	}
}

var _ domain.PaymentProvider = (*KuCoinProvider)(nil)
