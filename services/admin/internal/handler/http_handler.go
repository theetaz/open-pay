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
	"github.com/openlankapay/openlankapay/services/admin/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
	"github.com/openlankapay/openlankapay/services/admin/internal/service"
)

type AuditServiceInterface interface {
	CreateLog(ctx context.Context, input domain.AuditInput) (*domain.AuditLog, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)
	List(ctx context.Context, params postgres.ListParams) ([]*domain.AuditLog, int, error)
}

type AdminAuthServiceInterface interface {
	Login(ctx context.Context, email, password string) (*service.AdminLoginResult, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.AdminUser, error)
	RefreshToken(ctx context.Context, refreshToken string) (*service.AdminLoginResult, error)
}

type AdminHandler struct {
	auditSvc AuditServiceInterface
	authSvc  AdminAuthServiceInterface
}

func NewAdminHandler(auditSvc AuditServiceInterface, authSvc AdminAuthServiceInterface) *AdminHandler {
	return &AdminHandler{auditSvc: auditSvc, authSvc: authSvc}
}

func NewRouter(h *AdminHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public admin auth routes
	r.Post("/v1/admin/auth/login", h.AdminLogin)
	r.Post("/v1/admin/auth/refresh", h.AdminRefreshToken)

	// Protected admin routes (JWT + platform admin role)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Use(auth.RequirePlatformAdmin())

		r.Get("/v1/admin/auth/me", h.AdminMe)
		r.Get("/v1/audit-logs", h.ListAuditLogs)
		r.Get("/v1/audit-logs/{id}", h.GetAuditLog)
	})

	// Internal endpoint for creating audit logs from other services
	r.Post("/internal/audit-logs", h.CreateAuditLog)

	return r
}

// --- Admin Auth Handlers ---

func (h *AdminHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	result, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
			return
		}
		if errors.Is(err, domain.ErrAdminAccountInactive) {
			writeError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", "account is inactive")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to login")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": adminAuthResponse(result)})
}

func (h *AdminHandler) AdminRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	result, err := h.authSvc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired refresh token")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": adminAuthResponse(result)})
}

func (h *AdminHandler) AdminMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	user, err := h.authSvc.GetCurrentUser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "admin user not found")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": adminUserResponse(user)})
}

// --- Audit Handlers ---

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	params := postgres.ListParams{
		Page:         intQuery(r, "page", 1),
		PerPage:      intQuery(r, "perPage", 20),
		Action:       r.URL.Query().Get("action"),
		ActorType:    r.URL.Query().Get("actorType"),
		ResourceType: r.URL.Query().Get("resourceType"),
	}

	logs, total, err := h.auditSvc.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list audit logs")
		return
	}

	items := make([]map[string]any, 0, len(logs))
	for _, l := range logs {
		items = append(items, auditResponse(l))
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{"total": total, "page": params.Page, "perPage": params.PerPage},
	})
}

func (h *AdminHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid audit log ID")
		return
	}

	log, err := h.auditSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "audit log not found")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": auditResponse(log)})
}

func (h *AdminHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	var req createAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	actorID, _ := uuid.Parse(req.ActorID)
	var merchantID *uuid.UUID
	if req.MerchantID != "" {
		id, _ := uuid.Parse(req.MerchantID)
		merchantID = &id
	}
	var resourceID *uuid.UUID
	if req.ResourceID != "" {
		id, _ := uuid.Parse(req.ResourceID)
		resourceID = &id
	}

	log, err := h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
		ActorID:      actorID,
		ActorType:    req.ActorType,
		MerchantID:   merchantID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   resourceID,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create audit log")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": auditResponse(log)})
}

// --- Request/Response types ---

type createAuditRequest struct {
	ActorID      string `json:"actorId"`
	ActorType    string `json:"actorType"`
	MerchantID   string `json:"merchantId"`
	Action       string `json:"action"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	IPAddress    string `json:"ipAddress"`
	UserAgent    string `json:"userAgent"`
}

type envelope map[string]any

func adminAuthResponse(r *service.AdminLoginResult) map[string]any {
	return map[string]any{
		"accessToken":  r.AccessToken,
		"refreshToken": r.RefreshToken,
		"user":         adminUserResponse(r.User),
	}
}

func adminUserResponse(u *domain.AdminUser) map[string]any {
	resp := map[string]any{
		"id":       u.ID.String(),
		"email":    u.Email,
		"name":     u.Name,
		"isActive": u.IsActive,
	}
	if u.Role != nil {
		resp["role"] = map[string]any{
			"name":        u.Role.Name,
			"permissions": u.Role.Permissions,
		}
	}
	return resp
}

func auditResponse(l *domain.AuditLog) map[string]any {
	resp := map[string]any{
		"id":           l.ID.String(),
		"actorId":      l.ActorID.String(),
		"actorType":    l.ActorType,
		"action":       l.Action,
		"resourceType": l.ResourceType,
		"ipAddress":    l.IPAddress,
		"userAgent":    l.UserAgent,
		"createdAt":    l.CreatedAt.Format(time.RFC3339),
	}
	if l.MerchantID != nil {
		resp["merchantId"] = l.MerchantID.String()
	}
	if l.ResourceID != nil {
		resp["resourceId"] = l.ResourceID.String()
	}
	if l.Changes != nil {
		resp["changes"] = l.Changes
	}
	if l.Metadata != nil {
		resp["metadata"] = l.Metadata
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
