package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/provider"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
)

// PaymentServiceInterface defines the operations the handler depends on.
type PaymentServiceInterface interface {
	CreatePayment(ctx context.Context, input service.CreatePaymentInput) (*domain.Payment, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	ListPayments(ctx context.Context, merchantID uuid.UUID, params service.ListParams) ([]*domain.Payment, int, error)
	HandleProviderCallback(ctx context.Context, paymentID uuid.UUID) error
}

// PaymentHandler handles HTTP requests for payment operations.
type PaymentHandler struct {
	svc          PaymentServiceInterface
	mockProvider *provider.MockProvider
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc PaymentServiceInterface, mockProvider *provider.MockProvider) *PaymentHandler {
	return &PaymentHandler{svc: svc, mockProvider: mockProvider}
}

// NewRouter creates a chi router with payment routes.
func NewRouter(h *PaymentHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Protected routes (require JWT)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		r.Post("/v1/payments", h.CreatePayment)
		r.Get("/v1/payments", h.ListPayments)
		r.Get("/v1/payments/{id}", h.GetPayment)
	})

	// Public routes
	r.Get("/v1/payments/{id}/checkout", h.GetCheckout)
	r.Post("/v1/payments/{id}/callback", h.HandleCallback)

	// Test/sandbox routes
	if h.mockProvider != nil {
		r.Post("/test/simulate/{providerPayID}", h.SimulatePayment)
	}

	return r
}

// CreatePayment handles POST /v1/payments.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be a valid number")
		return
	}

	prov := req.Provider
	if prov == "" {
		prov = "TEST"
	}
	currency := req.Currency
	if currency == "" {
		currency = "USDT"
	}

	payment, err := h.svc.CreatePayment(r.Context(), service.CreatePaymentInput{
		MerchantID:      claims.MerchantID,
		BranchID:        req.BranchID,
		Amount:          amount,
		Currency:        currency,
		Provider:        prov,
		MerchantTradeNo: req.MerchantTradeNo,
		WebhookURL:      req.WebhookURL,
		CustomerEmail:   req.CustomerEmail,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPayment) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create payment")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": paymentResponse(payment)})
}

// GetPayment handles GET /v1/payments/{id}.
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment ID format")
		return
	}

	payment, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment")
		return
	}

	if payment.MerchantID != claims.MerchantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant's payment")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": paymentResponse(payment)})
}

// ListPayments handles GET /v1/payments.
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	params := service.ListParams{
		Page:    intQuery(r, "page", 1),
		PerPage: intQuery(r, "perPage", 20),
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.PaymentStatus(statusStr)
		params.Status = &status
	}

	payments, total, err := h.svc.ListPayments(r.Context(), claims.MerchantID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list payments")
		return
	}

	items := make([]map[string]any, 0, len(payments))
	for _, p := range payments {
		items = append(items, paymentResponse(p))
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{
			"total":   total,
			"page":    params.Page,
			"perPage": params.PerPage,
		},
	})
}

// GetCheckout handles GET /v1/payments/{id}/checkout (public, for customer).
func (h *PaymentHandler) GetCheckout(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment ID format")
		return
	}

	payment, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": checkoutResponse(payment)})
}

// HandleCallback handles POST /v1/payments/{id}/callback.
func (h *PaymentHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment ID format")
		return
	}

	if err := h.svc.HandleProviderCallback(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to process callback")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "processed"}})
}

// SimulatePayment handles POST /test/simulate/{providerPayID} (sandbox only).
func (h *PaymentHandler) SimulatePayment(w http.ResponseWriter, r *http.Request) {
	providerPayID := chi.URLParam(r, "providerPayID")
	if providerPayID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "provider pay ID is required")
		return
	}

	h.mockProvider.SimulatePayment(providerPayID)
	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "simulated", "providerPayId": providerPayID}})
}

// --- Request/Response types ---

type createPaymentRequest struct {
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	Provider        string     `json:"provider"`
	MerchantTradeNo string     `json:"merchantTradeNo"`
	WebhookURL      string     `json:"webhookUrl"`
	CustomerEmail   string     `json:"customerEmail"`
	BranchID        *uuid.UUID `json:"branchId"`
}

type envelope map[string]any

func paymentResponse(p *domain.Payment) map[string]any {
	resp := map[string]any{
		"id":              p.ID.String(),
		"merchantId":      p.MerchantID.String(),
		"paymentNo":       p.PaymentNo,
		"merchantTradeNo": p.MerchantTradeNo,
		"amount":          p.Amount.String(),
		"currency":        p.Currency,
		"amountUsdt":      p.AmountUSDT.String(),
		"exchangeFeePct":  p.ExchangeFeePct.String(),
		"exchangeFeeUsdt": p.ExchangeFeeUSDT.String(),
		"platformFeePct":  p.PlatformFeePct.String(),
		"platformFeeUsdt": p.PlatformFeeUSDT.String(),
		"totalFeesUsdt":   p.TotalFeesUSDT.String(),
		"netAmountUsdt":   p.NetAmountUSDT.String(),
		"provider":        p.Provider,
		"providerPayId":   p.ProviderPayID,
		"qrContent":       p.QRContent,
		"checkoutLink":    p.CheckoutLink,
		"deepLink":        p.DeepLink,
		"status":          string(p.Status),
		"customerEmail":   p.CustomerEmail,
		"txHash":          p.TxHash,
		"expireTime":      p.ExpireTime.Format(time.RFC3339),
		"createdAt":       p.CreatedAt.Format(time.RFC3339),
	}
	if p.BranchID != nil {
		resp["branchId"] = p.BranchID.String()
	}
	if p.ExchangeRateSnapshot != nil {
		resp["exchangeRate"] = p.ExchangeRateSnapshot.String()
	}
	if p.PaidAt != nil {
		resp["paidAt"] = p.PaidAt.Format(time.RFC3339)
	}
	return resp
}

func checkoutResponse(p *domain.Payment) map[string]any {
	resp := map[string]any{
		"id":         p.ID.String(),
		"paymentNo":  p.PaymentNo,
		"amount":     p.Amount.String(),
		"currency":   p.Currency,
		"amountUsdt": p.AmountUSDT.String(),
		"qrContent":  p.QRContent,
		"deepLink":   p.DeepLink,
		"status":     string(p.Status),
		"expireTime": p.ExpireTime.Format(time.RFC3339),
		"createdAt":  p.CreatedAt.Format(time.RFC3339),
	}
	if p.ExchangeRateSnapshot != nil {
		resp["exchangeRate"] = p.ExchangeRateSnapshot.String()
	}
	if p.PaidAt != nil {
		resp["paidAt"] = p.PaidAt.Format(time.RFC3339)
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func intQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}
