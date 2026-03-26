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

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
)

// WebhookServiceInterface defines the operations the handler depends on.
type WebhookServiceInterface interface {
	Configure(ctx context.Context, merchantID uuid.UUID, url string) (*domain.WebhookConfig, error)
	GetPublicKey(ctx context.Context, merchantID uuid.UUID) (string, error)
	Deliver(ctx context.Context, merchantID uuid.UUID, eventType string, payload []byte, httpClient *http.Client) (*domain.Delivery, error)
	ListDeliveries(ctx context.Context, merchantID uuid.UUID, page, perPage int) ([]*domain.Delivery, int, error)
}

// WebhookHandler handles HTTP requests for webhook operations.
type WebhookHandler struct {
	svc WebhookServiceInterface
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(svc WebhookServiceInterface) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

// NewRouter creates a chi router with webhook routes.
func NewRouter(h *WebhookHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))

		r.Post("/v1/webhooks/configure", h.Configure)
		r.Get("/v1/webhooks/public-key", h.GetPublicKey)
		r.Post("/v1/webhooks/test", h.TestWebhook)
		r.Get("/v1/webhooks/deliveries", h.ListDeliveries)
	})

	// Internal endpoint for delivering webhooks (called by other services)
	r.Post("/internal/webhooks/deliver", h.Deliver)

	return r
}

// Configure handles POST /v1/webhooks/configure.
func (h *WebhookHandler) Configure(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req configureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	cfg, err := h.svc.Configure(r.Context(), claims.MerchantID, req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidWebhookConfig) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to configure webhook")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": configResponse(cfg)})
}

// GetPublicKey handles GET /v1/webhooks/public-key.
func (h *WebhookHandler) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	pubKey, err := h.svc.GetPublicKey(r.Context(), claims.MerchantID)
	if err != nil {
		if errors.Is(err, domain.ErrWebhookConfigNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "webhook not configured")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get public key")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"publicKey": pubKey}})
}

// Deliver handles POST /internal/webhooks/deliver (internal, called by other services).
func (h *WebhookHandler) Deliver(w http.ResponseWriter, r *http.Request) {
	var req deliverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	payload, _ := json.Marshal(req.Payload)

	delivery, err := h.svc.Deliver(r.Context(), merchantID, req.EventType, payload, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to deliver webhook")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]any{
		"deliveryId": delivery.ID.String(),
		"status":     string(delivery.Status),
	}})
}

// TestWebhook handles POST /v1/webhooks/test — sends a test webhook to verify merchant integration.
func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	testPayload, _ := json.Marshal(map[string]any{
		"event": "test.webhook",
		"data": map[string]string{
			"message": "This is a test webhook from OpenLankaPay",
		},
	})

	start := time.Now()
	delivery, err := h.svc.Deliver(r.Context(), claims.MerchantID, "test.webhook", testPayload, nil)
	deliveryTime := time.Since(start).Milliseconds()
	if err != nil {
		if errors.Is(err, domain.ErrWebhookConfigNotFound) {
			writeError(w, http.StatusNotFound, "NOT_CONFIGURED", "webhook not configured — call POST /v1/webhooks/configure first")
			return
		}
		writeError(w, http.StatusInternalServerError, "DELIVERY_FAILED", "failed to deliver test webhook")
		return
	}

	success := delivery.Status == domain.DeliveryDelivered
	writeJSON(w, http.StatusOK, envelope{"data": map[string]any{
		"success":      success,
		"deliveryId":   delivery.ID.String(),
		"status":       string(delivery.Status),
		"deliveryTime": deliveryTime,
		"message":      "Test webhook delivered",
	}})
}

// ListDeliveries handles GET /v1/webhooks/deliveries.
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	page, perPage := 1, 20
	if v := r.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			page = p
		}
	}
	if v := r.URL.Query().Get("perPage"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			perPage = p
		}
	}

	deliveries, total, err := h.svc.ListDeliveries(r.Context(), claims.MerchantID, page, perPage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list deliveries")
		return
	}

	items := make([]map[string]any, 0, len(deliveries))
	for _, d := range deliveries {
		item := map[string]any{
			"id":               d.ID.String(),
			"eventType":        d.EventType,
			"status":           string(d.Status),
			"attemptCount":     d.AttemptCount,
			"maxAttempts":      d.MaxAttempts,
			"lastResponseCode": d.LastResponseCode,
			"lastError":        d.LastError,
			"nextAttemptAt":    d.NextAttemptAt,
			"deliveredAt":      d.DeliveredAt,
			"createdAt":        d.CreatedAt,
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{"page": page, "perPage": perPage, "total": total},
	})
}

// --- Request/Response types ---

type configureRequest struct {
	URL string `json:"url"`
}

type deliverRequest struct {
	MerchantID string `json:"merchantId"`
	EventType  string `json:"eventType"`
	Payload    any    `json:"payload"`
}

type envelope map[string]any

func configResponse(cfg *domain.WebhookConfig) map[string]any {
	return map[string]any{
		"id":        cfg.ID.String(),
		"url":       cfg.URL,
		"publicKey": cfg.SigningPublicKey,
		"events":    cfg.Events,
		"isActive":  cfg.IsActive,
		"createdAt": cfg.CreatedAt.Format(time.RFC3339),
	}
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
