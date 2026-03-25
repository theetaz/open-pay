package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
)

type checkoutSessionRequest struct {
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Provider        string `json:"provider"`
	MerchantTradeNo string `json:"merchantTradeNo"`
	Description     string `json:"description"`
	SuccessURL      string `json:"successUrl"`
	CancelURL       string `json:"cancelUrl"`
	CustomerEmail   string `json:"customerEmail"`
	LineItems       []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Amount      string `json:"amount"`
	} `json:"lineItems"`
	ExpiresInMinutes int `json:"expiresInMinutes"`
}

// CreateCheckoutSession handles POST /v1/checkout/sessions.
// This is the Stripe-like flow: SDK creates a session, gets back a hosted checkout URL.
func (h *PaymentHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req checkoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if req.Amount == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "amount is required")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be a valid number")
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "LKR"
	}
	provider := req.Provider
	if provider == "" {
		provider = "TEST"
	}

	if req.MerchantTradeNo == "" {
		req.MerchantTradeNo = "cs_" + uuid.New().String()[:8]
	}

	input := service.CreatePaymentInput{
		MerchantID:      merchantID,
		Amount:          amount,
		Currency:        currency,
		Provider:        provider,
		MerchantTradeNo: req.MerchantTradeNo,
		SuccessURL:      req.SuccessURL,
		CancelURL:       req.CancelURL,
		CustomerEmail:   req.CustomerEmail,
	}

	// Convert line items to goods
	for _, item := range req.LineItems {
		input.Goods = append(input.Goods, domain.GoodItem{Name: item.Name, Description: item.Description})
	}

	// Set expiration
	if req.ExpiresInMinutes > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInMinutes) * time.Minute)
		input.ExpireTime = &t
	}

	payment, err := h.svc.CreatePayment(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPayment) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create checkout session")
		return
	}

	// Build the hosted checkout URL
	checkoutURL := payment.CheckoutLink
	if checkoutURL == "" {
		// Fallback: generate our own hosted checkout URL
		gatewayBase := r.Header.Get("X-Forwarded-Host")
		if gatewayBase == "" {
			gatewayBase = r.Host
		}
		scheme := "https"
		if r.Header.Get("X-Forwarded-Proto") != "" {
			scheme = r.Header.Get("X-Forwarded-Proto")
		}
		checkoutURL = scheme + "://" + gatewayBase + "/v1/payments/" + payment.ID.String() + "/checkout"
	}

	// Session response
	sessionResp := map[string]any{
		"id":              "cs_" + payment.ID.String(),
		"paymentId":       payment.ID.String(),
		"url":             checkoutURL,
		"amount":          payment.Amount.String(),
		"currency":        payment.Currency,
		"amountUsdt":      payment.AmountUSDT.String(),
		"status":          "open",
		"qrContent":       payment.QRContent,
		"deepLink":        payment.DeepLink,
		"merchantTradeNo": payment.MerchantTradeNo,
		"successUrl":      req.SuccessURL,
		"cancelUrl":       req.CancelURL,
		"expiresAt":       payment.ExpireTime.Format(time.RFC3339),
		"createdAt":       payment.CreatedAt.Format(time.RFC3339),
	}

	if payment.ExchangeRateSnapshot != nil {
		sessionResp["exchangeRate"] = payment.ExchangeRateSnapshot.String()
	}

	writeJSON(w, http.StatusCreated, envelope{"data": sessionResp})
}
