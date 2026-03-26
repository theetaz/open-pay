package openpay

import "context"

// CheckoutService handles checkout session operations.
type CheckoutService struct {
	client *Client
}

// CheckoutSessionInput holds parameters for creating a checkout session.
type CheckoutSessionInput struct {
	Amount           string     `json:"amount"`
	Currency         string     `json:"currency,omitempty"`
	Provider         string     `json:"provider,omitempty"`
	MerchantTradeNo  string     `json:"merchantTradeNo,omitempty"`
	Description      string     `json:"description,omitempty"`
	SuccessURL       string     `json:"successUrl,omitempty"`
	CancelURL        string     `json:"cancelUrl,omitempty"`
	CustomerEmail    string     `json:"customerEmail,omitempty"`
	ExpiresInMinutes int        `json:"expiresInMinutes,omitempty"`
	LineItems        []LineItem `json:"lineItems,omitempty"`
}

// LineItem represents an item in the checkout session.
type LineItem struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Amount      string `json:"amount,omitempty"`
}

// CheckoutSession represents a checkout session response.
type CheckoutSession struct {
	ID              string `json:"id"`
	PaymentID       string `json:"paymentId"`
	URL             string `json:"url"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	AmountUSDT      string `json:"amountUsdt"`
	Status          string `json:"status"`
	QRContent       string `json:"qrContent"`
	DeepLink        string `json:"deepLink"`
	MerchantTradeNo string `json:"merchantTradeNo"`
	SuccessURL      string `json:"successUrl"`
	CancelURL       string `json:"cancelUrl"`
	ExchangeRate    string `json:"exchangeRate,omitempty"`
	ExpiresAt       string `json:"expiresAt"`
	CreatedAt       string `json:"createdAt"`
}

// CreateSession creates a new checkout session.
func (s *CheckoutService) CreateSession(ctx context.Context, input CheckoutSessionInput) (*CheckoutSession, error) {
	var resp struct {
		Data CheckoutSession `json:"data"`
	}
	if err := s.client.doRequest(ctx, "POST", "/v1/sdk/checkout/sessions", input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
