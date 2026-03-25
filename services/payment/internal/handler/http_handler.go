package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/pkg/audit"
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
	ExpireStalePayments(ctx context.Context) (int, error)
}

// PaymentHandler handles HTTP requests for payment operations.
type PaymentHandler struct {
	svc          PaymentServiceInterface
	mockProvider *provider.MockProvider
	auditLog     *audit.Client
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc PaymentServiceInterface, mockProvider *provider.MockProvider, auditClient ...*audit.Client) *PaymentHandler {
	h := &PaymentHandler{svc: svc, mockProvider: mockProvider}
	if len(auditClient) > 0 {
		h.auditLog = auditClient[0]
	}
	return h
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
	r.Post("/v1/public/payments", h.CreatePublicPayment)
	r.Get("/v1/payments/{id}/checkout", h.GetCheckout)
	r.Post("/v1/payments/{id}/callback", h.HandleCallback)

	// Internal routes
	r.Post("/internal/payments/expire-stale", h.ExpireStalePayments)

	// Test/sandbox routes
	if h.mockProvider != nil {
		r.Post("/test/simulate/{providerPayID}", h.SimulatePayment)
	}

	return r
}

// merchantIDFromRequest extracts the merchant ID from JWT claims or X-Merchant-ID header (SDK auth).
func merchantIDFromRequest(r *http.Request) (uuid.UUID, bool) {
	// Try JWT claims first (portal users)
	if claims, ok := auth.ClaimsFromContext(r.Context()); ok && claims.MerchantID != uuid.Nil {
		return claims.MerchantID, true
	}
	// Fallback to X-Merchant-ID header (set by gateway for HMAC-authenticated SDK requests)
	if hdr := r.Header.Get("X-Merchant-ID"); hdr != "" {
		if id, err := uuid.Parse(hdr); err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

// CreatePayment handles POST /v1/payments.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}
	// Build a claims-like reference for audit logging
	claims, _ := auth.ClaimsFromContext(r.Context())

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

	input := service.CreatePaymentInput{
		MerchantID:      merchantID,
		BranchID:        req.BranchID,
		Amount:          amount,
		Currency:        currency,
		Provider:        prov,
		MerchantTradeNo: req.MerchantTradeNo,
		WebhookURL:      req.WebhookURL,
		SuccessURL:      req.SuccessURL,
		CancelURL:       req.CancelURL,
		CustomerEmail:   req.CustomerEmail,
	}
	if req.CustomerBilling != nil {
		input.CustomerFirstName = req.CustomerBilling.FirstName
		input.CustomerLastName = req.CustomerBilling.LastName
		input.CustomerPhone = req.CustomerBilling.Phone
		input.CustomerAddress = req.CustomerBilling.Address
		if input.CustomerEmail == "" {
			input.CustomerEmail = req.CustomerBilling.Email
		}
	}
	for _, g := range req.Goods {
		input.Goods = append(input.Goods, domain.GoodItem{Name: g.Name, Description: g.Description, MccCode: g.MccCode})
	}
	if req.OrderExpireTime != "" {
		if t, err := time.Parse(time.RFC3339, req.OrderExpireTime); err == nil {
			input.ExpireTime = &t
		}
	}

	payment, err := h.svc.CreatePayment(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPayment) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create payment")
		return
	}

	if h.auditLog != nil {
		actorID := merchantID
		actorType := "SDK"
		if claims != nil {
			actorID = claims.UserID
			actorType = "MERCHANT_USER"
		}
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: actorID, ActorType: actorType, MerchantID: &merchantID,
			Action: "payment.initiated", ResourceType: "payment", ResourceID: &payment.ID,
			IPAddress:  stripPort(r.RemoteAddr),
			Metadata:   map[string]string{"amount": payment.Amount.String(), "currency": payment.Currency, "provider": payment.Provider},
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": paymentResponse(payment)})
}

// CreatePublicPayment handles POST /v1/public/payments (no auth - for payment links).
// Requires merchantId in the request body instead of JWT claims.
func (h *PaymentHandler) CreatePublicPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MerchantID      string  `json:"merchantId"`
		Amount          string  `json:"amount"`
		Currency        string  `json:"currency"`
		Provider        string  `json:"provider"`
		MerchantTradeNo string  `json:"merchantTradeNo"`
		CustomerEmail   string  `json:"customerEmail"`
		BranchID        *string `json:"branchId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_MERCHANT_ID", "invalid merchant ID")
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

	var branchID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		bid, err := uuid.Parse(*req.BranchID)
		if err == nil {
			branchID = &bid
		}
	}

	payment, err := h.svc.CreatePayment(r.Context(), service.CreatePaymentInput{
		MerchantID:      merchantID,
		BranchID:        branchID,
		Amount:          amount,
		Currency:        currency,
		Provider:        prov,
		MerchantTradeNo: req.MerchantTradeNo,
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

	if h.auditLog != nil {
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorType: "SYSTEM", MerchantID: &merchantID,
			Action: "payment.initiated", ResourceType: "payment", ResourceID: &payment.ID,
			Metadata: map[string]string{"amount": payment.Amount.String(), "currency": payment.Currency, "provider": payment.Provider, "merchantTradeNo": req.MerchantTradeNo},
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": paymentResponse(payment)})
}

// GetPayment handles GET /v1/payments/{id}.
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromRequest(r)
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

	if payment.MerchantID != merchantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot access another merchant's payment")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": paymentResponse(payment)})
}

// ListPayments handles GET /v1/payments.
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := merchantIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	params := service.ListParams{
		Page:    intQuery(r, "page", 1),
		PerPage: intQuery(r, "perPage", 20),
		Search:  r.URL.Query().Get("search"),
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.PaymentStatus(statusStr)
		params.Status = &status
	}
	if bid := r.URL.Query().Get("branchId"); bid != "" {
		if id, err := uuid.Parse(bid); err == nil {
			params.BranchID = &id
		}
	}
	if df := r.URL.Query().Get("dateFrom"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			params.DateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("dateTo"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			params.DateTo = &t
		}
	}

	payments, total, err := h.svc.ListPayments(r.Context(), merchantID, params)
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

	if h.auditLog != nil {
		merchantID := payment.MerchantID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorType: "SYSTEM", MerchantID: &merchantID,
			Action: "payment.checkout_viewed", ResourceType: "payment", ResourceID: &payment.ID,
			IPAddress: stripPort(r.RemoteAddr),
		})
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

	if h.auditLog != nil {
		if payment, err := h.svc.GetPayment(r.Context(), id); err == nil {
			action := "payment.paid"
			switch payment.Status {
			case domain.StatusExpired:
				action = "payment.expired"
			case domain.StatusFailed:
				action = "payment.failed"
			}
			merchantID := payment.MerchantID
			meta := map[string]string{"status": string(payment.Status)}
			if payment.TxHash != "" {
				meta["txHash"] = payment.TxHash
			}
			h.auditLog.Log(r.Context(), audit.LogEntry{
				ActorType: "SYSTEM", MerchantID: &merchantID,
				Action: action, ResourceType: "payment", ResourceID: &payment.ID,
				Metadata: meta,
			})
		}
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "processed"}})
}

// ExpireStalePayments handles POST /internal/payments/expire-stale.
func (h *PaymentHandler) ExpireStalePayments(w http.ResponseWriter, r *http.Request) {
	count, err := h.svc.ExpireStalePayments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to expire payments")
		return
	}

	if h.auditLog != nil && count > 0 {
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorType: "SYSTEM",
			Action:    "payment.expired", ResourceType: "payment",
			Metadata: map[string]string{"expiredCount": strconv.Itoa(count)},
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]any{"expired": count}})
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
	Amount            string              `json:"amount"`
	Currency          string              `json:"currency"`
	Provider          string              `json:"provider"`
	MerchantTradeNo   string              `json:"merchantTradeNo"`
	WebhookURL        string              `json:"webhookUrl"`
	SuccessURL        string              `json:"successUrl"`
	CancelURL         string              `json:"cancelUrl"`
	CustomerEmail     string              `json:"customerEmail"`
	BranchID          *uuid.UUID          `json:"branchId"`
	OrderExpireTime   string              `json:"orderExpireTime"`
	CustomerBilling   *customerBillingReq `json:"customerBilling"`
	Goods             []goodItemReq       `json:"goods"`
}

type customerBillingReq struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
}

type goodItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MccCode     string `json:"mccCode"`
}

type envelope map[string]any

func paymentResponse(p *domain.Payment) map[string]any {
	feeBreakdown := map[string]any{
		"grossAmountUSDT":       p.AmountUSDT.String(),
		"exchangeFeePercentage": p.ExchangeFeePct.String(),
		"exchangeFeeAmountUSDT": p.ExchangeFeeUSDT.String(),
		"ceypayFeePercentage":   p.PlatformFeePct.String(),
		"ceypayFeeAmountUSDT":   p.PlatformFeeUSDT.String(),
		"totalFeesUSDT":         p.TotalFeesUSDT.String(),
		"netAmountUSDT":         p.NetAmountUSDT.String(),
	}
	if p.LKRAmount != nil {
		feeBreakdown["grossAmountLKR"] = p.LKRAmount.String()
		feeBreakdown["exchangeFeeAmountLKR"] = p.LKRExchangeFee.String()
		feeBreakdown["ceypayFeeAmountLKR"] = p.LKRPlatformFee.String()
		feeBreakdown["totalFeesLKR"] = p.LKRTotalFees.String()
		feeBreakdown["netAmountLKR"] = p.LKRNetAmount.String()
	}

	resp := map[string]any{
		"id":              p.ID.String(),
		"merchantId":      p.MerchantID.String(),
		"paymentNo":       p.PaymentNo,
		"merchantTradeNo": p.MerchantTradeNo,
		"amount":          p.Amount.String(),
		"currency":        p.Currency,
		"amountUsdt":      p.AmountUSDT.String(),
		"provider":        p.Provider,
		"providerPayId":   p.ProviderPayID,
		"qrContent":       p.QRContent,
		"checkoutLink":    p.CheckoutLink,
		"deepLink":        p.DeepLink,
		"status":          string(p.Status),
		"customerEmail":   p.CustomerEmail,
		"txHash":          p.TxHash,
		"feeBreakdown":    feeBreakdown,
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
	if p.CustomerFirstName != "" || p.CustomerLastName != "" {
		resp["customerBilling"] = map[string]string{
			"firstName": p.CustomerFirstName,
			"lastName":  p.CustomerLastName,
			"email":     p.CustomerEmail,
			"phone":     p.CustomerPhone,
			"address":   p.CustomerAddress,
		}
	}
	if len(p.Goods) > 0 {
		resp["goods"] = p.Goods
	}
	if p.SuccessURL != "" {
		resp["successUrl"] = p.SuccessURL
	}
	if p.CancelURL != "" {
		resp["cancelUrl"] = p.CancelURL
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

func stripPort(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
