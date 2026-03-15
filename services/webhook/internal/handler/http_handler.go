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

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
)

// WebhookServiceInterface defines the operations the handler depends on.
type WebhookServiceInterface interface {
	Configure(ctx context.Context, merchantID uuid.UUID, url string) (*domain.WebhookConfig, error)
	GetPublicKey(ctx context.Context, merchantID uuid.UUID) (string, error)
	Deliver(ctx context.Context, merchantID uuid.UUID, eventType string, payload []byte, httpClient *http.Client) (*domain.Delivery, error)
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
