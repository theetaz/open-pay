package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/audit"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/merchant/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/shopspring/decimal"
)

type paymentLinkRepo interface {
	Create(ctx context.Context, pl *domain.PaymentLink) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentLink, error)
	GetBySlug(ctx context.Context, merchantID uuid.UUID, slug string) (*domain.PaymentLink, error)
	GetBySlugGlobal(ctx context.Context, slug string) (*domain.PaymentLink, error)
	SlugExists(ctx context.Context, merchantID uuid.UUID, slug string) (bool, error)
	List(ctx context.Context, params postgres.PaymentLinkListParams) ([]*domain.PaymentLink, int, error)
	Update(ctx context.Context, pl *domain.PaymentLink) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID) error
	ExpireStale(ctx context.Context) (int, error)
}

// PaymentLinkHandler handles payment link HTTP requests.
type PaymentLinkHandler struct {
	repo     paymentLinkRepo
	auditLog *audit.Client
}

// RegisterPaymentLinkRoutes registers payment link routes on the router.
func RegisterPaymentLinkRoutes(r chi.Router, repo paymentLinkRepo, jwtSecret string, auditClient ...*audit.Client) {
	h := &PaymentLinkHandler{repo: repo}
	if len(auditClient) > 0 {
		h.auditLog = auditClient[0]
	}

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		r.Post("/v1/payment-links", h.Create)
		r.Get("/v1/payment-links", h.List)
		r.Get("/v1/payment-links/check-slug/{slug}", h.CheckSlug)
		r.Get("/v1/payment-links/{id}", h.GetByID)
		r.Put("/v1/payment-links/{id}", h.Update)
		r.Delete("/v1/payment-links/{id}", h.Delete)
	})

	// Public route: resolve by slug (for checkout pages)
	r.Get("/v1/public/payment-links/by-slug/{slug}", h.GetBySlugPublic)

	// Internal routes (no auth — called by other services)
	r.Post("/internal/payment-links/increment-usage", h.IncrementUsage)
	r.Post("/internal/payment-links/expire-stale", h.ExpireStaleLinks)
}

// Create handles POST /v1/payment-links.
func (h *PaymentLinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req createPaymentLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be a valid number")
		return
	}

	pl, err := domain.NewPaymentLink(claims.MerchantID, req.Name, req.Slug, req.Currency, amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPaymentLink) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create payment link")
		return
	}

	pl.Description = req.Description
	pl.AllowCustomAmount = req.AllowCustomAmount
	pl.IsReusable = req.IsReusable
	pl.ShowOnQRPage = req.ShowOnQRPage
	pl.AllowQuantityBuy = req.AllowQuantityBuy
	pl.SuccessURL = req.SuccessURL
	pl.CancelURL = req.CancelURL
	pl.WebhookURL = req.WebhookURL
	pl.MerchantTradeNo = req.MerchantTradeNo
	pl.OrderExpireMinutes = req.OrderExpireMinutes

	if req.MaxQuantity != nil {
		pl.MaxQuantity = *req.MaxQuantity
	}

	if req.MinAmount != nil {
		v, err := decimal.NewFromString(*req.MinAmount)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "minAmount must be a valid number")
			return
		}
		pl.MinAmount = &v
	}
	if req.MaxAmount != nil {
		v, err := decimal.NewFromString(*req.MaxAmount)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "maxAmount must be a valid number")
			return
		}
		pl.MaxAmount = &v
	}

	if claims.BranchID != nil {
		pl.BranchID = claims.BranchID
	}

	if req.ExpireAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "expireAt must be in RFC3339 format")
			return
		}
		pl.ExpireAt = &t
	}

	if err := pl.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.repo.Create(r.Context(), pl); err != nil {
		if errors.Is(err, domain.ErrDuplicateSlug) {
			writeError(w, http.StatusConflict, "DUPLICATE_SLUG", "a payment link with this slug already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create payment link")
		return
	}

	if h.auditLog != nil {
		merchantID := claims.MerchantID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: claims.UserID, ActorType: "MERCHANT_USER", MerchantID: &merchantID,
			Action: "payment_link.created", ResourceType: "payment_link", ResourceID: &pl.ID,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": paymentLinkResponse(pl)})
}

// List handles GET /v1/payment-links.
func (h *PaymentLinkHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	page := intQuery(r, "page", 1)
	perPage := intQuery(r, "perPage", 20)

	listParams := postgres.PaymentLinkListParams{
		MerchantID: claims.MerchantID,
		Page:       page,
		PerPage:    perPage,
		Search:     r.URL.Query().Get("search"),
		Status:     r.URL.Query().Get("status"),
	}
	if bid := r.URL.Query().Get("branchId"); bid != "" {
		if id, err := uuid.Parse(bid); err == nil {
			listParams.BranchID = &id
		}
	}

	links, total, err := h.repo.List(r.Context(), listParams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list payment links")
		return
	}

	items := make([]map[string]any, 0, len(links))
	for _, pl := range links {
		items = append(items, paymentLinkResponse(pl))
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{
			"total":   total,
			"page":    page,
			"perPage": perPage,
		},
	})
}

// GetByID handles GET /v1/payment-links/{id}.
func (h *PaymentLinkHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment link ID")
		return
	}

	pl, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentLinkNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment link")
		return
	}

	if pl.MerchantID != claims.MerchantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": paymentLinkResponse(pl)})
}

// Update handles PUT /v1/payment-links/{id}.
func (h *PaymentLinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment link ID")
		return
	}

	pl, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentLinkNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment link")
		return
	}

	if pl.MerchantID != claims.MerchantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req updatePaymentLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if req.Name != nil {
		pl.Name = *req.Name
	}
	if req.Description != nil {
		pl.Description = *req.Description
	}
	if req.Status != nil {
		pl.Status = domain.PaymentLinkStatus(*req.Status)
	}
	if req.AllowQuantityBuy != nil {
		pl.AllowQuantityBuy = *req.AllowQuantityBuy
	}
	if req.MaxQuantity != nil {
		pl.MaxQuantity = *req.MaxQuantity
	}
	if req.SuccessURL != nil {
		pl.SuccessURL = *req.SuccessURL
	}
	if req.CancelURL != nil {
		pl.CancelURL = *req.CancelURL
	}
	if req.WebhookURL != nil {
		pl.WebhookURL = *req.WebhookURL
	}
	if req.OrderExpireMinutes != nil {
		pl.OrderExpireMinutes = req.OrderExpireMinutes
	}

	if err := h.repo.Update(r.Context(), pl); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update payment link")
		return
	}

	if h.auditLog != nil {
		merchantID := claims.MerchantID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: claims.UserID, ActorType: "MERCHANT_USER", MerchantID: &merchantID,
			Action: "payment_link.updated", ResourceType: "payment_link", ResourceID: &id,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": paymentLinkResponse(pl)})
}

// Delete handles DELETE /v1/payment-links/{id}.
func (h *PaymentLinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment link ID")
		return
	}

	pl, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentLinkNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment link")
		return
	}

	if pl.MerchantID != claims.MerchantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	if err := h.repo.SoftDelete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete payment link")
		return
	}

	if h.auditLog != nil {
		merchantID := claims.MerchantID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: claims.UserID, ActorType: "MERCHANT_USER", MerchantID: &merchantID,
			Action: "payment_link.deleted", ResourceType: "payment_link", ResourceID: &id,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "deleted"}})
}

// CheckSlug handles GET /v1/payment-links/check-slug/{slug}.
func (h *PaymentLinkHandler) CheckSlug(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "INVALID_SLUG", "slug is required")
		return
	}

	exists, err := h.repo.SlugExists(r.Context(), claims.MerchantID, slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check slug")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]bool{"available": !exists}})
}

// GetBySlugPublic handles GET /v1/public/payment-links/by-slug/{slug} (no auth).
func (h *PaymentLinkHandler) GetBySlugPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "slug is required")
		return
	}

	pl, err := h.repo.GetBySlugGlobal(r.Context(), slug)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentLinkNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get payment link")
		return
	}

	if pl.Status != domain.PaymentLinkActive {
		writeError(w, http.StatusNotFound, "INACTIVE", "this payment link is no longer active")
		return
	}

	if pl.IsExpired() {
		writeJSON(w, http.StatusGone, envelope{"error": map[string]string{"code": "EXPIRED", "message": "this payment link has expired"}})
		return
	}

	if pl.IsConsumed() {
		writeJSON(w, http.StatusGone, envelope{"error": map[string]string{"code": "CONSUMED", "message": "this single-use payment link has already been used"}})
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": paymentLinkResponse(pl)})
}

// IncrementUsage handles POST /internal/payment-links/increment-usage (no auth — internal).
func (h *PaymentLinkHandler) IncrementUsage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "slug is required")
		return
	}

	pl, err := h.repo.GetBySlugGlobal(r.Context(), req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentLinkNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "payment link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to find payment link")
		return
	}

	if err := h.repo.IncrementUsage(r.Context(), pl.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to increment usage")
		return
	}

	if h.auditLog != nil {
		merchantID := pl.MerchantID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorType:    "SYSTEM",
			MerchantID:   &merchantID,
			Action:       "payment_link.used",
			ResourceType: "payment_link",
			ResourceID:   &pl.ID,
			Metadata:     map[string]string{"slug": req.Slug},
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "incremented"}})
}

// ExpireStaleLinks handles POST /internal/payment-links/expire-stale.
func (h *PaymentLinkHandler) ExpireStaleLinks(w http.ResponseWriter, r *http.Request) {
	count, err := h.repo.ExpireStale(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to expire stale links")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]any{"expired": count}})
}

// --- Request types ---

type createPaymentLinkRequest struct {
	Name               string  `json:"name"`
	Slug               string  `json:"slug"`
	Description        string  `json:"description"`
	Currency           string  `json:"currency"`
	Amount             string  `json:"amount"`
	AllowCustomAmount  bool    `json:"allowCustomAmount"`
	MinAmount          *string `json:"minAmount"`
	MaxAmount          *string `json:"maxAmount"`
	AllowQuantityBuy   bool    `json:"allowQuantityBuy"`
	MaxQuantity        *int    `json:"maxQuantity"`
	IsReusable         bool    `json:"isReusable"`
	ShowOnQRPage       bool    `json:"showOnQrPage"`
	SuccessURL         string  `json:"successUrl"`
	CancelURL          string  `json:"cancelUrl"`
	WebhookURL         string  `json:"webhookUrl"`
	MerchantTradeNo    string  `json:"merchantTradeNo"`
	OrderExpireMinutes *int    `json:"orderExpireMinutes"`
	ExpireAt           string  `json:"expireAt"`
}

type updatePaymentLinkRequest struct {
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	Status             *string `json:"status"`
	AllowQuantityBuy   *bool   `json:"allowQuantityBuy"`
	MaxQuantity        *int    `json:"maxQuantity"`
	SuccessURL         *string `json:"successUrl"`
	CancelURL          *string `json:"cancelUrl"`
	WebhookURL         *string `json:"webhookUrl"`
	OrderExpireMinutes *int    `json:"orderExpireMinutes"`
}

// --- Response helper ---

func paymentLinkResponse(pl *domain.PaymentLink) map[string]any {
	resp := map[string]any{
		"id":                pl.ID.String(),
		"merchantId":        pl.MerchantID.String(),
		"name":              pl.Name,
		"slug":              pl.Slug,
		"description":       pl.Description,
		"currency":          pl.Currency,
		"amount":            pl.Amount.String(),
		"allowCustomAmount": pl.AllowCustomAmount,
		"allowQuantityBuy":  pl.AllowQuantityBuy,
		"maxQuantity":       pl.MaxQuantity,
		"isReusable":        pl.IsReusable,
		"showOnQrPage":      pl.ShowOnQRPage,
		"usageCount":        pl.UsageCount,
		"status":            string(pl.Status),
		"createdAt":         pl.CreatedAt.Format(time.RFC3339),
		"updatedAt":         pl.UpdatedAt.Format(time.RFC3339),
	}
	if pl.BranchID != nil {
		resp["branchId"] = pl.BranchID.String()
	}
	if pl.ExpireAt != nil {
		resp["expireAt"] = pl.ExpireAt.Format(time.RFC3339)
	}
	if pl.MinAmount != nil {
		resp["minAmount"] = pl.MinAmount.String()
	}
	if pl.MaxAmount != nil {
		resp["maxAmount"] = pl.MaxAmount.String()
	}
	if pl.SuccessURL != "" {
		resp["successUrl"] = pl.SuccessURL
	}
	if pl.CancelURL != "" {
		resp["cancelUrl"] = pl.CancelURL
	}
	if pl.WebhookURL != "" {
		resp["webhookUrl"] = pl.WebhookURL
	}
	if pl.MerchantTradeNo != "" {
		resp["merchantTradeNo"] = pl.MerchantTradeNo
	}
	if pl.OrderExpireMinutes != nil {
		resp["orderExpireMinutes"] = *pl.OrderExpireMinutes
	}
	return resp
}
