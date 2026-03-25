package openpay

import "time"

// Payment represents a payment object from the API.
type Payment struct {
	ID              string     `json:"id"`
	MerchantID      string     `json:"merchantId"`
	BranchID        string     `json:"branchId,omitempty"`
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	Provider        string     `json:"provider"`
	ProviderPayID   string     `json:"providerPayId,omitempty"`
	MerchantTradeNo string     `json:"merchantTradeNo"`
	QRContent       string     `json:"qrContent,omitempty"`
	CheckoutLink    string     `json:"checkoutLink,omitempty"`
	DeepLink        string     `json:"deepLink,omitempty"`
	WebhookURL      string     `json:"webhookURL,omitempty"`
	SuccessURL      string     `json:"successURL,omitempty"`
	CancelURL       string     `json:"cancelURL,omitempty"`
	CustomerEmail   string     `json:"customerEmail,omitempty"`
	AmountLKR       string     `json:"amountLkr,omitempty"`
	ExchangeRate    string     `json:"exchangeRate,omitempty"`
	PlatformFeeLKR  string     `json:"platformFeeLkr,omitempty"`
	ExchangeFeeLKR  string     `json:"exchangeFeeLkr,omitempty"`
	NetAmountLKR    string     `json:"netAmountLkr,omitempty"`
	PaidAt          *time.Time `json:"paidAt,omitempty"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
}

// CreatePaymentInput holds parameters for creating a payment.
type CreatePaymentInput struct {
	Amount          string `json:"amount"`
	Currency        string `json:"currency,omitempty"`
	Provider        string `json:"provider,omitempty"`
	MerchantTradeNo string `json:"merchantTradeNo,omitempty"`
	Description     string `json:"description,omitempty"`
	WebhookURL      string `json:"webhookURL,omitempty"`
	SuccessURL      string `json:"successURL,omitempty"`
	CancelURL       string `json:"cancelURL,omitempty"`
	CustomerEmail   string `json:"customerEmail,omitempty"`
}

// ListPaymentsParams holds query parameters for listing payments.
type ListPaymentsParams struct {
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"perPage,omitempty"`
	Status   string `json:"status,omitempty"`
	Search   string `json:"search,omitempty"`
	BranchID string `json:"branchId,omitempty"`
	DateFrom string `json:"dateFrom,omitempty"`
	DateTo   string `json:"dateTo,omitempty"`
}

// PaginatedPayments holds a paginated list of payments.
type PaginatedPayments struct {
	Data []Payment `json:"data"`
	Meta struct {
		Page    int `json:"page"`
		PerPage int `json:"perPage"`
		Total   int `json:"total"`
	} `json:"meta"`
}

// WebhookEvent represents a parsed webhook event.
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
