package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

// --- request/response structs ---

type binanceCreateReq struct {
	Env             binanceEnv   `json:"env"`
	MerchantTradeNo string       `json:"merchantTradeNo"`
	OrderAmount     string       `json:"orderAmount"`
	Currency        string       `json:"currency"`
	Goods           binanceGoods `json:"goods"`
}

type binanceEnv struct {
	TerminalType string `json:"terminalType"`
}

type binanceGoods struct {
	GoodsType        string `json:"goodsType"`
	GoodsCategory    string `json:"goodsCategory"`
	ReferenceGoodsID string `json:"referenceGoodsId"`
	GoodsName        string `json:"goodsName"`
}

type binanceAPIResp struct {
	Status string          `json:"status"`
	Code   string          `json:"code"`
	Data   json.RawMessage `json:"data"`
}

type binanceCreateData struct {
	PrepayID    string `json:"prepayId"`
	QRCodeLink  string `json:"qrcodeLink"`
	CheckoutURL string `json:"checkoutUrl"`
	DeepLink    string `json:"deeplink"`
}

type binanceQueryReq struct {
	PrepayID string `json:"prepayId"`
}

type binanceQueryData struct {
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
}

// --- signing ---

func binanceNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (p *BinanceProvider) sign(timestamp, nonce, body string) string {
	message := timestamp + "\n" + nonce + "\n" + body + "\n"
	mac := hmac.New(sha512.New, []byte(p.config.APISecret))
	mac.Write([]byte(message))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

// --- HTTP helper ---

func (p *BinanceProvider) doRequest(ctx context.Context, url string, body []byte) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	nonce := binanceNonce()
	sig := p.sign(timestamp, nonce, string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("binance: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("BinancePay-Timestamp", timestamp)
	req.Header.Set("BinancePay-Nonce", nonce)
	req.Header.Set("BinancePay-Certificate-SN", p.config.APIKey)
	req.Header.Set("BinancePay-Signature", sig)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("binance: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("binance: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// --- PaymentProvider interface ---

func (p *BinanceProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	// Binance requires merchantTradeNo without hyphens
	tradeNo := strings.ReplaceAll(req.OrderID, "-", "")

	payload := binanceCreateReq{
		Env:             binanceEnv{TerminalType: "WEB"},
		MerchantTradeNo: tradeNo,
		OrderAmount:     req.Amount,
		Currency:        req.Currency,
		Goods: binanceGoods{
			GoodsType:        "02",
			GoodsCategory:    "Z000",
			ReferenceGoodsID: req.OrderID,
			GoodsName:        "Payment",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("binance: marshal request: %w", err)
	}

	respBody, err := p.doRequest(ctx, p.config.BaseURL+"/binancepay/openapi/v2/order", body)
	if err != nil {
		return nil, err
	}

	var apiResp binanceAPIResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("binance: unmarshal response: %w", err)
	}
	if apiResp.Status != "SUCCESS" || apiResp.Code != "000000" {
		return nil, fmt.Errorf("binance: API error status=%s code=%s", apiResp.Status, apiResp.Code)
	}

	var data binanceCreateData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("binance: unmarshal data: %w", err)
	}

	return &domain.ProviderPaymentResponse{
		ProviderPayID: data.PrepayID,
		QRContent:     data.QRCodeLink,
		CheckoutLink:  data.CheckoutURL,
		DeepLink:      data.DeepLink,
	}, nil
}

func (p *BinanceProvider) GetPaymentStatus(ctx context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	payload := binanceQueryReq{PrepayID: providerPayID}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("binance: marshal request: %w", err)
	}

	respBody, err := p.doRequest(ctx, p.config.BaseURL+"/binancepay/openapi/v2/order/query", body)
	if err != nil {
		return nil, err
	}

	var apiResp binanceAPIResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("binance: unmarshal response: %w", err)
	}
	if apiResp.Status != "SUCCESS" || apiResp.Code != "000000" {
		return nil, fmt.Errorf("binance: API error status=%s code=%s", apiResp.Status, apiResp.Code)
	}

	var data binanceQueryData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("binance: unmarshal data: %w", err)
	}

	return &domain.ProviderPaymentStatus{
		Status: mapBinanceStatus(data.Status),
		TxHash: data.TransactionID,
	}, nil
}

func mapBinanceStatus(s string) domain.PaymentStatus {
	switch s {
	case "PAID":
		return domain.StatusPaid
	case "EXPIRED", "CANCELED":
		return domain.StatusExpired
	case "ERROR", "REFUNDED":
		return domain.StatusFailed
	default:
		return domain.StatusInitiated
	}
}

var _ domain.PaymentProvider = (*BinanceProvider)(nil)
