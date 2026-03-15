package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
)

// SettlementServiceInterface defines the operations the handler depends on.
type SettlementServiceInterface interface {
	GetBalance(ctx context.Context, merchantID uuid.UUID) (*domain.MerchantBalance, error)
	CreditPayment(ctx context.Context, merchantID uuid.UUID, netUSDT, feesUSDT decimal.Decimal) error
	RequestWithdrawal(ctx context.Context, merchantID uuid.UUID, amountUSDT, exchangeRate decimal.Decimal, bankName, bankAccountNo, bankAccountName string) (*domain.Withdrawal, error)
	ListWithdrawals(ctx context.Context, merchantID uuid.UUID, status *domain.WithdrawalStatus) ([]*domain.Withdrawal, error)
	GetWithdrawal(ctx context.Context, id uuid.UUID) (*domain.Withdrawal, error)
	ApproveWithdrawal(ctx context.Context, withdrawalID, adminID uuid.UUID) error
	RejectWithdrawal(ctx context.Context, withdrawalID uuid.UUID, reason string) error
	CompleteWithdrawal(ctx context.Context, withdrawalID uuid.UUID, bankReference string) error
}

// SettlementHandler handles HTTP requests for settlement operations.
type SettlementHandler struct {
	svc SettlementServiceInterface
}

// NewSettlementHandler creates a new SettlementHandler.
func NewSettlementHandler(svc SettlementServiceInterface) *SettlementHandler {
	return &SettlementHandler{svc: svc}
}

// NewRouter creates a chi router with settlement routes.
func NewRouter(h *SettlementHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		r.Get("/v1/settlements/balance", h.GetBalance)
		r.Post("/v1/settlements/credit", h.CreditPayment)
		r.Post("/v1/withdrawals", h.RequestWithdrawal)
		r.Get("/v1/withdrawals", h.ListWithdrawals)
		r.Get("/v1/withdrawals/{id}", h.GetWithdrawal)
		r.Post("/v1/withdrawals/{id}/approve", h.ApproveWithdrawal)
		r.Post("/v1/withdrawals/{id}/reject", h.RejectWithdrawal)
		r.Post("/v1/withdrawals/{id}/complete", h.CompleteWithdrawal)
	})

	return r
}

// GetBalance handles GET /v1/settlements/balance.
func (h *SettlementHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	balance, err := h.svc.GetBalance(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get balance")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": balanceResponse(balance)})
}

// CreditPayment handles POST /v1/settlements/credit (internal, called by payment service).
func (h *SettlementHandler) CreditPayment(w http.ResponseWriter, r *http.Request) {
	var req creditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	netUSDT, err := decimal.NewFromString(req.NetUSDT)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid net amount")
		return
	}

	feesUSDT, err := decimal.NewFromString(req.FeesUSDT)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid fees amount")
		return
	}

	if err := h.svc.CreditPayment(r.Context(), merchantID, netUSDT, feesUSDT); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to credit payment")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "credited"}})
}

// RequestWithdrawal handles POST /v1/withdrawals.
func (h *SettlementHandler) RequestWithdrawal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req withdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	amountUSDT, err := decimal.NewFromString(req.AmountUSDT)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid amount")
		return
	}

	exchangeRate, err := decimal.NewFromString(req.ExchangeRate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RATE", "invalid exchange rate")
		return
	}

	withdrawal, err := h.svc.RequestWithdrawal(r.Context(), claims.MerchantID, amountUSDT, exchangeRate, req.BankName, req.BankAccountNo, req.BankAccountName)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) {
			writeError(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "insufficient balance for withdrawal")
			return
		}
		if errors.Is(err, domain.ErrInvalidWithdrawal) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create withdrawal")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": withdrawalResponse(withdrawal)})
}

// ListWithdrawals handles GET /v1/withdrawals.
func (h *SettlementHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var status *domain.WithdrawalStatus
	if s := r.URL.Query().Get("status"); s != "" {
		ws := domain.WithdrawalStatus(s)
		status = &ws
	}

	withdrawals, err := h.svc.ListWithdrawals(r.Context(), claims.MerchantID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list withdrawals")
		return
	}

	items := make([]map[string]any, 0, len(withdrawals))
	for _, w := range withdrawals {
		items = append(items, withdrawalResponse(w))
	}

	writeJSON(w, http.StatusOK, envelope{"data": items})
}

// GetWithdrawal handles GET /v1/withdrawals/{id}.
func (h *SettlementHandler) GetWithdrawal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal ID")
		return
	}

	withdrawal, err := h.svc.GetWithdrawal(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWithdrawalNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "withdrawal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get withdrawal")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": withdrawalResponse(withdrawal)})
}

// ApproveWithdrawal handles POST /v1/withdrawals/{id}/approve.
func (h *SettlementHandler) ApproveWithdrawal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal ID")
		return
	}

	if err := h.svc.ApproveWithdrawal(r.Context(), id, claims.UserID); err != nil {
		if errors.Is(err, domain.ErrInvalidWithdrawalTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to approve withdrawal")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "approved"}})
}

// RejectWithdrawal handles POST /v1/withdrawals/{id}/reject.
func (h *SettlementHandler) RejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal ID")
		return
	}

	var req rejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if err := h.svc.RejectWithdrawal(r.Context(), id, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reject withdrawal")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "rejected"}})
}

// CompleteWithdrawal handles POST /v1/withdrawals/{id}/complete.
func (h *SettlementHandler) CompleteWithdrawal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid withdrawal ID")
		return
	}

	var req completeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if err := h.svc.CompleteWithdrawal(r.Context(), id, req.BankReference); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete withdrawal")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "completed"}})
}

// --- Request/Response types ---

type creditRequest struct {
	MerchantID string `json:"merchantId"`
	NetUSDT    string `json:"netUsdt"`
	FeesUSDT   string `json:"feesUsdt"`
}

type withdrawalRequest struct {
	AmountUSDT      string `json:"amountUsdt"`
	ExchangeRate    string `json:"exchangeRate"`
	BankName        string `json:"bankName"`
	BankAccountNo   string `json:"bankAccountNo"`
	BankAccountName string `json:"bankAccountName"`
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

type completeRequest struct {
	BankReference string `json:"bankReference"`
}

type envelope map[string]any

func balanceResponse(b *domain.MerchantBalance) map[string]any {
	return map[string]any{
		"merchantId":        b.MerchantID.String(),
		"availableUsdt":     b.AvailableUSDT.String(),
		"pendingUsdt":       b.PendingUSDT.String(),
		"totalEarnedUsdt":   b.TotalEarnedUSDT.String(),
		"totalWithdrawnUsdt": b.TotalWithdrawnUSDT.String(),
		"totalFeesUsdt":     b.TotalFeesUSDT.String(),
		"updatedAt":         b.UpdatedAt.Format(time.RFC3339),
	}
}

func withdrawalResponse(w *domain.Withdrawal) map[string]any {
	resp := map[string]any{
		"id":              w.ID.String(),
		"merchantId":      w.MerchantID.String(),
		"amountUsdt":      w.AmountUSDT.String(),
		"exchangeRate":    w.ExchangeRate.String(),
		"amountLkr":       w.AmountLKR.String(),
		"bankName":        w.BankName,
		"bankAccountNo":   w.BankAccountNo,
		"bankAccountName": w.BankAccountName,
		"status":          string(w.Status),
		"createdAt":       w.CreatedAt.Format(time.RFC3339),
	}
	if w.BankReference != "" {
		resp["bankReference"] = w.BankReference
	}
	if w.RejectedReason != "" {
		resp["rejectedReason"] = w.RejectedReason
	}
	if w.CompletedAt != nil {
		resp["completedAt"] = w.CompletedAt.Format(time.RFC3339)
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
