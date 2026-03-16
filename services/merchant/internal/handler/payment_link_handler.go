package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
}

// PaymentLinkHandler handles payment link HTTP requests.
type PaymentLinkHandler struct {
	repo paymentLinkRepo
}

// RegisterPaymentLinkRoutes registers payment link routes on the router.
func RegisterPaymentLinkRoutes(r chi.Router, repo paymentLinkRepo, jwtSecret string) {
	h := &PaymentLinkHandler{repo: repo}

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

	if err := h.repo.Create(r.Context(), pl); err != nil {
		if errors.Is(err, domain.ErrDuplicateSlug) {
			writeError(w, http.StatusConflict, "DUPLICATE_SLUG", "a payment link with this slug already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create payment link")
		return
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

	links, total, err := h.repo.List(r.Context(), postgres.PaymentLinkListParams{
		MerchantID: claims.MerchantID,
		Page:       page,
		PerPage:    perPage,
	})
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

	if err := h.repo.Update(r.Context(), pl); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update payment link")
		return
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

	writeJSON(w, http.StatusOK, envelope{"data": paymentLinkResponse(pl)})
}

// --- Request types ---

type createPaymentLinkRequest struct {
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	Description       string `json:"description"`
	Currency          string `json:"currency"`
	Amount            string `json:"amount"`
	AllowCustomAmount bool   `json:"allowCustomAmount"`
	IsReusable        bool   `json:"isReusable"`
	ShowOnQRPage      bool   `json:"showOnQrPage"`
	ExpireAt          string `json:"expireAt"`
}

type updatePaymentLinkRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
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
	return resp
}
