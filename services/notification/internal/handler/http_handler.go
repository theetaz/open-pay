package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/openlankapay/openlankapay/pkg/auth"
	pgadapter "github.com/openlankapay/openlankapay/services/notification/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/notification/internal/domain"
)

type NotificationServiceInterface interface {
	Send(ctx context.Context, input domain.NotificationInput) (*domain.Notification, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Notification, error)
}

type EmailTemplateRepo interface {
	List(ctx context.Context) ([]*pgadapter.EmailTemplate, error)
	Create(ctx context.Context, t *pgadapter.EmailTemplate) error
	Update(ctx context.Context, t *pgadapter.EmailTemplate) error
}

type NotificationHandler struct {
	svc       NotificationServiceInterface
	templates EmailTemplateRepo
}

func NewNotificationHandler(svc NotificationServiceInterface, templates EmailTemplateRepo) *NotificationHandler {
	return &NotificationHandler{svc: svc, templates: templates}
}

func NewRouter(h *NotificationHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Get("/v1/notifications", h.ListNotifications)
	})

	// Internal endpoint for sending notifications (called by other services)
	r.Post("/internal/notifications/send", h.SendNotification)

	// Admin email template management (protected by platform admin JWT)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Use(auth.RequirePlatformAdmin())

		r.Get("/v1/admin/email-templates", h.ListEmailTemplates)
		r.Post("/v1/admin/email-templates", h.CreateEmailTemplate)
		r.Put("/v1/admin/email-templates/{id}", h.UpdateEmailTemplate)
	})

	return r
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	notifications, err := h.svc.ListByMerchant(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list notifications")
		return
	}

	items := make([]map[string]any, 0, len(notifications))
	for _, n := range notifications {
		items = append(items, notificationResponse(n))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	n, err := h.svc.Send(r.Context(), domain.NotificationInput{
		MerchantID: merchantID,
		Channel:    domain.Channel(req.Channel),
		Recipient:  req.Recipient,
		Subject:    req.Subject,
		Body:       req.Body,
		EventType:  req.EventType,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to send notification")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": notificationResponse(n)})
}

type sendRequest struct {
	MerchantID string `json:"merchantId"`
	Channel    string `json:"channel"`
	Recipient  string `json:"recipient"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	EventType  string `json:"eventType"`
}

type envelope map[string]any

func notificationResponse(n *domain.Notification) map[string]any {
	resp := map[string]any{
		"id":        n.ID.String(),
		"channel":   string(n.Channel),
		"recipient": n.Recipient,
		"subject":   n.Subject,
		"eventType": n.EventType,
		"status":    string(n.Status),
		"createdAt": n.CreatedAt.Format(time.RFC3339),
	}
	if n.SentAt != nil {
		resp["sentAt"] = n.SentAt.Format(time.RFC3339)
	}
	if n.FailureReason != "" {
		resp["failureReason"] = n.FailureReason
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

// --- Email Template Handlers ---

func (h *NotificationHandler) ListEmailTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.templates.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list templates")
		return
	}

	items := make([]map[string]any, 0, len(templates))
	for _, t := range templates {
		items = append(items, templateResponse(t))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *NotificationHandler) CreateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventType string   `json:"eventType"`
		Name      string   `json:"name"`
		Subject   string   `json:"subject"`
		BodyHTML  string   `json:"bodyHtml"`
		Variables []string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	t := &pgadapter.EmailTemplate{
		EventType: req.EventType,
		Name:      req.Name,
		Subject:   req.Subject,
		BodyHTML:  req.BodyHTML,
		Variables: req.Variables,
	}

	if err := h.templates.Create(r.Context(), t); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create template")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": templateResponse(t)})
}

func (h *NotificationHandler) UpdateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid template ID")
		return
	}

	var req struct {
		Name      string   `json:"name"`
		Subject   string   `json:"subject"`
		BodyHTML  string   `json:"bodyHtml"`
		Variables []string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	t := &pgadapter.EmailTemplate{
		ID:        id,
		Name:      req.Name,
		Subject:   req.Subject,
		BodyHTML:  req.BodyHTML,
		Variables: req.Variables,
	}

	if err := h.templates.Update(r.Context(), t); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update template")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "updated"}})
}

func templateResponse(t *pgadapter.EmailTemplate) map[string]any {
	return map[string]any{
		"id":        t.ID.String(),
		"eventType": t.EventType,
		"name":      t.Name,
		"subject":   t.Subject,
		"bodyHtml":  t.BodyHTML,
		"variables": t.Variables,
		"isActive":  t.IsActive,
		"createdAt": t.CreatedAt.Format(time.RFC3339),
		"updatedAt": t.UpdatedAt.Format(time.RFC3339),
	}
}
