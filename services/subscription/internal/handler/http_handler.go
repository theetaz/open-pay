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
	"github.com/openlankapay/openlankapay/services/subscription/internal/domain"
)

type SubscriptionServiceInterface interface {
	CreatePlan(ctx context.Context, merchantID uuid.UUID, name, description string, amount decimal.Decimal, currency string, interval domain.IntervalType, intervalCount, trialDays int) (*domain.SubscriptionPlan, error)
	GetPlan(ctx context.Context, id uuid.UUID) (*domain.SubscriptionPlan, error)
	ListPlans(ctx context.Context, merchantID uuid.UUID) ([]*domain.SubscriptionPlan, error)
	ArchivePlan(ctx context.Context, id uuid.UUID) error
	Subscribe(ctx context.Context, planID uuid.UUID, email, wallet string) (*domain.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	ListSubscriptions(ctx context.Context, merchantID uuid.UUID) ([]*domain.Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID, reason string) error
}

type SubscriptionHandler struct {
	svc SubscriptionServiceInterface
}

func NewSubscriptionHandler(svc SubscriptionServiceInterface) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func NewRouter(h *SubscriptionHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		r.Post("/v1/subscription-plans", h.CreatePlan)
		r.Get("/v1/subscription-plans", h.ListPlans)
		r.Get("/v1/subscription-plans/{id}", h.GetPlan)
		r.Post("/v1/subscription-plans/{id}/archive", h.ArchivePlan)
		r.Get("/v1/subscriptions", h.ListSubscriptions)
		r.Get("/v1/subscriptions/{id}", h.GetSubscription)
		r.Post("/v1/subscriptions/{id}/cancel", h.CancelSubscription)
	})

	// Public route for subscribing
	r.Post("/v1/subscription-plans/{id}/subscribe", h.Subscribe)

	return r
}

func (h *SubscriptionHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req createPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid amount")
		return
	}

	plan, err := h.svc.CreatePlan(r.Context(), claims.MerchantID, req.Name, req.Description, amount, req.Currency, domain.IntervalType(req.IntervalType), req.IntervalCount, req.TrialDays)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPlan) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create plan")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": planResponse(plan)})
}

func (h *SubscriptionHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid plan ID")
		return
	}

	plan, err := h.svc.GetPlan(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "plan not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get plan")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": planResponse(plan)})
}

func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	plans, err := h.svc.ListPlans(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list plans")
		return
	}

	items := make([]map[string]any, 0, len(plans))
	for _, p := range plans {
		items = append(items, planResponse(p))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *SubscriptionHandler) ArchivePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid plan ID")
		return
	}

	if err := h.svc.ArchivePlan(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to archive plan")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "archived"}})
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid plan ID")
		return
	}

	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	sub, err := h.svc.Subscribe(r.Context(), id, req.Email, req.Wallet)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidSubscription) || errors.Is(err, domain.ErrPlanNotFound) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to subscribe")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": subscriptionResponse(sub)})
}

func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid subscription ID")
		return
	}

	sub, err := h.svc.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get subscription")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": subscriptionResponse(sub)})
}

func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	subs, err := h.svc.ListSubscriptions(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list subscriptions")
		return
	}

	items := make([]map[string]any, 0, len(subs))
	for _, s := range subs {
		items = append(items, subscriptionResponse(s))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid subscription ID")
		return
	}

	var req cancelRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.CancelSubscription(r.Context(), id, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel subscription")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "cancelled"}})
}

// --- Request/Response types ---

type createPlanRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	IntervalType  string `json:"intervalType"`
	IntervalCount int    `json:"intervalCount"`
	TrialDays     int    `json:"trialDays"`
}

type subscribeRequest struct {
	Email  string `json:"email"`
	Wallet string `json:"wallet"`
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

type envelope map[string]any

func planResponse(p *domain.SubscriptionPlan) map[string]any {
	return map[string]any{
		"id":            p.ID.String(),
		"merchantId":    p.MerchantID.String(),
		"name":          p.Name,
		"description":   p.Description,
		"amount":        p.Amount.String(),
		"currency":      p.Currency,
		"intervalType":  string(p.IntervalType),
		"intervalCount": p.IntervalCount,
		"trialDays":     p.TrialDays,
		"status":        string(p.Status),
		"createdAt":     p.CreatedAt.Format(time.RFC3339),
	}
}

func subscriptionResponse(s *domain.Subscription) map[string]any {
	resp := map[string]any{
		"id":              s.ID.String(),
		"planId":          s.PlanID.String(),
		"merchantId":      s.MerchantID.String(),
		"subscriberEmail": s.SubscriberEmail,
		"status":          string(s.Status),
		"nextBillingDate": s.NextBillingDate.Format(time.RFC3339),
		"totalPaidUsdt":   s.TotalPaidUSDT.String(),
		"billingCount":    s.BillingCount,
		"createdAt":       s.CreatedAt.Format(time.RFC3339),
	}
	if s.TrialEnd != nil {
		resp["trialEnd"] = s.TrialEnd.Format(time.RFC3339)
	}
	if s.CancelledAt != nil {
		resp["cancelledAt"] = s.CancelledAt.Format(time.RFC3339)
		resp["cancellationReason"] = s.CancellationReason
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{"error": map[string]string{"code": code, "message": message}})
}
