package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/service"
)

// DirectDebitServiceInterface defines the service methods used by handlers.
type DirectDebitServiceInterface interface {
	ListScenarioCodes(ctx context.Context, provider string, activeOnly bool) ([]*domain.ScenarioCode, error)
	CreateContract(ctx context.Context, merchantID uuid.UUID, input service.CreateContractInput) (*domain.Contract, error)
	GetContract(ctx context.Context, id uuid.UUID) (*domain.Contract, error)
	ListContracts(ctx context.Context, merchantID uuid.UUID, status string, page, limit int) ([]*domain.Contract, int, error)
	SyncContractStatus(ctx context.Context, id uuid.UUID) (*domain.Contract, error)
	TerminateContract(ctx context.Context, id uuid.UUID, notes string) (*domain.Contract, error)
	ExecutePayment(ctx context.Context, contractID uuid.UUID, input service.ExecutePaymentInput) (*domain.Payment, error)
}

// DirectDebitHandler handles HTTP requests for direct debit operations.
type DirectDebitHandler struct {
	svc DirectDebitServiceInterface
}

// NewDirectDebitHandler creates a new handler.
func NewDirectDebitHandler(svc DirectDebitServiceInterface) *DirectDebitHandler {
	return &DirectDebitHandler{svc: svc}
}

// NewRouter creates the direct debit HTTP router.
func NewRouter(h *DirectDebitHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, envelope{"status": "ok"})
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		// Scenario codes
		r.Get("/v1/direct-debit/scenario-codes", h.ListScenarioCodes)

		// Contracts
		r.Post("/v1/direct-debit", h.CreateContract)
		r.Get("/v1/direct-debit/list", h.ListContracts)
		r.Get("/v1/direct-debit/{id}", h.GetContract)
		r.Post("/v1/direct-debit/{id}/sync", h.SyncContractStatus)
		r.Post("/v1/direct-debit/{id}/terminate", h.TerminateContract)
		r.Post("/v1/direct-debit/{id}/payment", h.ExecutePayment)
	})

	return r
}

func (h *DirectDebitHandler) ListScenarioCodes(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	activeOnly := r.URL.Query().Get("active") != "false"

	scenarios, err := h.svc.ListScenarioCodes(r.Context(), provider, activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list scenario codes")
		return
	}

	items := make([]map[string]any, 0, len(scenarios))
	for _, s := range scenarios {
		items = append(items, scenarioResponse(s))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *DirectDebitHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req createContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	scenarioID, err := uuid.Parse(req.ScenarioID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCENARIO_ID", "invalid scenario ID")
		return
	}

	limit, err := decimal.NewFromString(req.SingleUpperLimit)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid single upper limit")
		return
	}

	input := service.CreateContractInput{
		MerchantContractCode: req.MerchantContractCode,
		ServiceName:          req.ServiceName,
		ScenarioID:           scenarioID,
		SingleUpperLimit:     limit,
		WebhookURL:           req.WebhookURL,
		ReturnURL:            req.ReturnURL,
		CancelURL:            req.CancelURL,
	}
	if req.BranchID != "" {
		bid, err := uuid.Parse(req.BranchID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_BRANCH_ID", "invalid branch ID")
			return
		}
		input.BranchID = &bid
	}

	contract, err := h.svc.CreateContract(r.Context(), claims.MerchantID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidContract) || errors.Is(err, domain.ErrScenarioNotFound) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create contract")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": contractResponse(contract)})
}

func (h *DirectDebitHandler) GetContract(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid contract ID")
		return
	}

	contract, err := h.svc.GetContract(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "contract not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get contract")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": contractResponse(contract)})
}

func (h *DirectDebitHandler) ListContracts(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	contracts, total, err := h.svc.ListContracts(r.Context(), claims.MerchantID, status, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list contracts")
		return
	}

	items := make([]map[string]any, 0, len(contracts))
	for _, c := range contracts {
		items = append(items, contractResponse(c))
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	totalPages := (total + limit - 1) / limit

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{
			"page":       page,
			"limit":      limit,
			"totalItems": total,
			"totalPages": totalPages,
		},
	})
}

func (h *DirectDebitHandler) SyncContractStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid contract ID")
		return
	}

	contract, err := h.svc.SyncContractStatus(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "contract not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to sync contract status")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": contractResponse(contract)})
}

func (h *DirectDebitHandler) TerminateContract(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid contract ID")
		return
	}

	var req terminateRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	contract, err := h.svc.TerminateContract(r.Context(), id, req.TerminationNotes)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "contract not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to terminate contract")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": contractResponse(contract)})
}

func (h *DirectDebitHandler) ExecutePayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid contract ID")
		return
	}

	var req executePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid amount")
		return
	}

	input := service.ExecutePaymentInput{
		Amount:            amount,
		ProductName:       req.ProductName,
		ProductDetail:     req.ProductDetail,
		WebhookURL:        req.WebhookURL,
		CustomerFirstName: req.CustomerFirstName,
		CustomerLastName:  req.CustomerLastName,
		CustomerEmail:     req.CustomerEmail,
		CustomerPhone:     req.CustomerPhone,
		CustomerAddress:   req.CustomerAddress,
	}

	payment, err := h.svc.ExecutePayment(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "contract not found")
			return
		}
		if errors.Is(err, domain.ErrContractNotSigned) || errors.Is(err, domain.ErrAmountExceedsLimit) || errors.Is(err, domain.ErrInvalidPayment) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to execute payment")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": paymentResponse(payment)})
}

// --- Request types ---

type createContractRequest struct {
	MerchantContractCode string `json:"merchantContractCode"`
	BranchID             string `json:"branchId"`
	ServiceName          string `json:"serviceName"`
	ScenarioID           string `json:"scenarioId"`
	SingleUpperLimit     string `json:"singleUpperLimit"`
	WebhookURL           string `json:"webhookUrl"`
	ReturnURL            string `json:"returnUrl"`
	CancelURL            string `json:"cancelUrl"`
}

type terminateRequest struct {
	TerminationNotes string `json:"terminationNotes"`
}

type executePaymentRequest struct {
	Amount            string `json:"amount"`
	ProductName       string `json:"productName"`
	ProductDetail     string `json:"productDetail"`
	WebhookURL        string `json:"webhookUrl"`
	CustomerFirstName string `json:"customerFirstName"`
	CustomerLastName  string `json:"customerLastName"`
	CustomerEmail     string `json:"customerEmail"`
	CustomerPhone     string `json:"customerPhone"`
	CustomerAddress   string `json:"customerAddress"`
}

// --- Response helpers ---

type envelope map[string]any

func scenarioResponse(s *domain.ScenarioCode) map[string]any {
	return map[string]any{
		"id":              s.ID.String(),
		"scenarioId":      s.ScenarioID,
		"scenarioName":    s.ScenarioName,
		"paymentProvider": s.PaymentProvider,
		"maxLimit":        s.MaxLimit.String(),
		"isActive":        s.IsActive,
	}
}

func contractResponse(c *domain.Contract) map[string]any {
	resp := map[string]any{
		"id":                   c.ID.String(),
		"merchantId":           c.MerchantID.String(),
		"merchantContractCode": c.MerchantContractCode,
		"serviceName":          c.ServiceName,
		"scenarioId":           c.ScenarioID.String(),
		"paymentProvider":      c.PaymentProvider,
		"currency":             c.Currency,
		"singleUpperLimit":     c.SingleUpperLimit.String(),
		"status":               c.Status,
		"qrContent":            c.QRContent,
		"deepLink":             c.DeepLink,
		"periodic":             c.Periodic,
		"paymentCount":         c.PaymentCount,
		"totalAmountCharged":   c.TotalAmountCharged.String(),
		"createdAt":            c.CreatedAt.Format(time.RFC3339),
		"updatedAt":            c.UpdatedAt.Format(time.RFC3339),
	}
	if c.BranchID != nil {
		resp["branchId"] = c.BranchID.String()
	}
	if c.ContractID != "" {
		resp["contractId"] = c.ContractID
	}
	if c.OpenUserID != "" {
		resp["openUserId"] = c.OpenUserID
	}
	if c.WebhookURL != "" {
		resp["webhookUrl"] = c.WebhookURL
	}
	if c.TerminationTime != nil {
		resp["terminationTime"] = c.TerminationTime.Format(time.RFC3339)
		resp["terminationNotes"] = c.TerminationNotes
	}
	if c.LastPaymentAt != nil {
		resp["lastPaymentAt"] = c.LastPaymentAt.Format(time.RFC3339)
	}
	return resp
}

func paymentResponse(p *domain.Payment) map[string]any {
	return map[string]any{
		"id":              p.ID.String(),
		"contractId":      p.ContractID.String(),
		"merchantId":      p.MerchantID.String(),
		"payId":           p.PayID,
		"paymentNo":       p.PaymentNo,
		"amount":          p.Amount.String(),
		"currency":        p.Currency,
		"status":          p.Status,
		"productName":     p.ProductName,
		"paymentProvider": p.PaymentProvider,
		"createdAt":       p.CreatedAt.Format(time.RFC3339),
		"feeBreakdown": map[string]any{
			"grossAmountUSDT":       p.GrossAmountUSDT.String(),
			"exchangeFeePercentage": p.ExchangeFeePct.String(),
			"exchangeFeeAmountUSDT": p.ExchangeFeeUSDT.String(),
			"platformFeePercentage": p.PlatformFeePct.String(),
			"platformFeeAmountUSDT": p.PlatformFeeUSDT.String(),
			"totalFeesUSDT":         p.TotalFeesUSDT.String(),
			"netAmountUSDT":         p.NetAmountUSDT.String(),
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{"error": map[string]string{"code": code, "message": message}})
}
