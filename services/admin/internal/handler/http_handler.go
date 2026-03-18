package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/pkg/notification"
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

type LegalDocRepo interface {
	GetActiveByType(ctx context.Context, docType string) (*postgres.LegalDocument, error)
	List(ctx context.Context) ([]*postgres.LegalDocument, error)
	Create(ctx context.Context, doc *postgres.LegalDocument) error
	Update(ctx context.Context, doc *postgres.LegalDocument) error
	Activate(ctx context.Context, id uuid.UUID) error
}

type SettingsRepo interface {
	GetAll(ctx context.Context) ([]*postgres.PlatformSetting, error)
	GetByCategory(ctx context.Context, category string) ([]*postgres.PlatformSetting, error)
	BulkUpdate(ctx context.Context, updates map[string]string, updatedBy uuid.UUID) error
}

type AdminUserRepo interface {
	ListUsers(ctx context.Context, page, perPage int) ([]*domain.AdminUser, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminUser, error)
	Create(ctx context.Context, user *domain.AdminUser) error
	UpdateUser(ctx context.Context, user *domain.AdminUser) error
	ChangePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	ListRoles(ctx context.Context) ([]*domain.AdminRole, error)
	CreateRole(ctx context.Context, role *domain.AdminRole) error
	UpdateRole(ctx context.Context, role *domain.AdminRole) error
	GetRoleByName(ctx context.Context, name string) (*domain.AdminRole, error)
}

type AdminHandler struct {
	auditSvc  AuditServiceInterface
	authSvc   AdminAuthServiceInterface
	legalDocs LegalDocRepo
	settings  SettingsRepo
	userRepo  AdminUserRepo
	notifier  *notification.Client
}

func NewAdminHandler(auditSvc AuditServiceInterface, authSvc AdminAuthServiceInterface, legalDocs LegalDocRepo, settings SettingsRepo, userRepo AdminUserRepo, notifier *notification.Client) *AdminHandler {
	return &AdminHandler{auditSvc: auditSvc, authSvc: authSvc, legalDocs: legalDocs, settings: settings, userRepo: userRepo, notifier: notifier}
}

func NewRouter(h *AdminHandler, jwtSecret string, uploadHandler ...*AdminUploadHandler) http.Handler {
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
		r.Post("/v1/admin/auth/change-password", h.ChangePassword)
		r.Get("/v1/audit-logs", h.ListAuditLogs)
		r.Get("/v1/audit-logs/{id}", h.GetAuditLog)
	})

	// Merchant-scoped audit log access (JWT required, auto-filtered by merchantId)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Get("/v1/merchant/audit-logs", h.ListMerchantAuditLogs)
	})

	// Protected admin management routes
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Use(auth.RequirePlatformAdmin())

		// Legal documents
		r.Get("/v1/admin/legal-documents", h.ListLegalDocuments)
		r.Post("/v1/admin/legal-documents", h.CreateLegalDocument)
		r.Put("/v1/admin/legal-documents/{id}", h.UpdateLegalDocument)
		r.Post("/v1/admin/legal-documents/{id}/activate", h.ActivateLegalDocument)

		// Platform settings
		r.Get("/v1/admin/settings", h.GetSettings)
		r.Put("/v1/admin/settings", h.UpdateSettings)
		r.Get("/v1/admin/settings/{category}", h.GetSettingsByCategory)

		// Admin user management
		r.Get("/v1/admin/users", h.ListAdminUsers)
		r.Post("/v1/admin/users", h.CreateAdminUser)
		r.Put("/v1/admin/users/{id}", h.UpdateAdminUser)
		r.Post("/v1/admin/users/{id}/deactivate", h.DeactivateAdminUser)

		// Admin role management
		r.Get("/v1/admin/roles", h.ListRoles)
		r.Post("/v1/admin/roles", h.CreateRole)
		r.Put("/v1/admin/roles/{id}", h.UpdateRoleHandler)

		// Admin file uploads (logos, branding)
		if len(uploadHandler) > 0 && uploadHandler[0] != nil {
			r.Post("/v1/admin/uploads", uploadHandler[0].Upload)
		}
	})

	// Public asset serving (logos, branding images from MinIO)
	if len(uploadHandler) > 0 && uploadHandler[0] != nil {
		r.Get("/v1/assets/*", uploadHandler[0].ServeAsset)
	}

	// Public endpoint for fetching active legal documents (used by merchant portal)
	r.Get("/v1/legal-documents/active", h.GetActiveLegalDocument)

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
			// Log failed login attempt
			h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
				ActorType: "ADMIN", Action: "admin.login.failed",
				ResourceType: "admin_user", IPAddress: stripPort(r.RemoteAddr),
				UserAgent: r.UserAgent(),
			})
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

	// Log successful login
	h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
		ActorID: result.User.ID, ActorType: "ADMIN", Action: "admin.login",
		ResourceType: "admin_user", ResourceID: &result.User.ID,
		IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
	})

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

func (h *AdminHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "new password must be at least 8 characters")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	if !user.VerifyPassword(req.CurrentPassword) {
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "current password is incorrect")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password")
		return
	}

	if err := h.userRepo.ChangePassword(r.Context(), user.ID, string(hash)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to change password")
		return
	}

	h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
		ActorID: claims.UserID, ActorType: "ADMIN", Action: "admin_user.password_changed",
		ResourceType: "admin_user", ResourceID: &user.ID,
		IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "password_changed"}})
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

	if mid := r.URL.Query().Get("merchantId"); mid != "" {
		if id, err := uuid.Parse(mid); err == nil {
			params.MerchantID = &id
		}
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

func (h *AdminHandler) ListMerchantAuditLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	merchantID := claims.MerchantID
	params := postgres.ListParams{
		Page:         intQuery(r, "page", 1),
		PerPage:      intQuery(r, "perPage", 20),
		Action:       r.URL.Query().Get("action"),
		MerchantID:   &merchantID,
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
		"id":                 u.ID.String(),
		"email":              u.Email,
		"name":               u.Name,
		"isActive":           u.IsActive,
		"mustChangePassword": u.MustChangePassword,
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

// --- Legal Document Handlers ---

func (h *AdminHandler) GetActiveLegalDocument(w http.ResponseWriter, r *http.Request) {
	docType := r.URL.Query().Get("type")
	if docType == "" {
		docType = "terms_and_conditions"
	}

	doc, err := h.legalDocs.GetActiveByType(r.Context(), docType)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no active document found for this type")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": legalDocResponse(doc)})
}

func (h *AdminHandler) ListLegalDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.legalDocs.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list documents")
		return
	}

	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		items = append(items, legalDocResponse(d))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *AdminHandler) CreateLegalDocument(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string `json:"type"`
		Version int    `json:"version"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}
	if req.Type == "" || req.Title == "" || req.Content == "" || req.Version < 1 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "type, version, title, and content are required")
		return
	}

	doc := &postgres.LegalDocument{
		Type:    req.Type,
		Version: req.Version,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.legalDocs.Create(r.Context(), doc); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create document")
		return
	}

	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
			ActorID: claims.UserID, ActorType: "ADMIN", Action: "legal_document.created",
			ResourceType: "legal_document", ResourceID: &doc.ID,
			IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": legalDocResponse(doc)})
}

func (h *AdminHandler) UpdateLegalDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	doc := &postgres.LegalDocument{
		ID:      id,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.legalDocs.Update(r.Context(), doc); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update document")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "updated"}})
}

func (h *AdminHandler) ActivateLegalDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	if err := h.legalDocs.Activate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to activate document")
		return
	}

	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
			ActorID: claims.UserID, ActorType: "ADMIN", Action: "legal_document.activated",
			ResourceType: "legal_document", ResourceID: &id,
			IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "activated"}})
}

func stripPort(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// --- Platform Settings Handlers ---

func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settings.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get settings")
		return
	}

	grouped := make(map[string][]map[string]any)
	for _, s := range settings {
		grouped[s.Category] = append(grouped[s.Category], map[string]any{
			"key": s.Key, "value": s.Value, "description": s.Description, "updatedAt": s.UpdatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, envelope{"data": grouped})
}

func (h *AdminHandler) GetSettingsByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	settings, err := h.settings.GetByCategory(r.Context(), category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get settings")
		return
	}

	items := make([]map[string]any, 0, len(settings))
	for _, s := range settings {
		items = append(items, map[string]any{
			"key": s.Key, "value": s.Value, "description": s.Description, "updatedAt": s.UpdatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *AdminHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Settings map[string]string `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Settings) == 0 {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "settings map is required")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	if err := h.settings.BulkUpdate(r.Context(), req.Settings, claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update settings")
		return
	}

	h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
		ActorID: claims.UserID, ActorType: "ADMIN", Action: "settings.updated",
		ResourceType: "platform_settings", IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "updated"}})
}

// --- Admin User Management Handlers ---

func (h *AdminHandler) ListAdminUsers(w http.ResponseWriter, r *http.Request) {
	page := intQuery(r, "page", 1)
	perPage := intQuery(r, "perPage", 20)

	users, total, err := h.userRepo.ListUsers(r.Context(), page, perPage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list users")
		return
	}

	items := make([]map[string]any, 0, len(users))
	for _, u := range users {
		items = append(items, adminUserListResponse(u))
	}
	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]int{"total": total, "page": page, "perPage": perPage},
	})
}

func (h *AdminHandler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		RoleID   string `json:"roleId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid role ID")
		return
	}

	user, err := domain.NewAdminUser(req.Email, req.Password, req.Name, roleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		if errors.Is(err, domain.ErrDuplicateAdminEmail) {
			writeError(w, http.StatusConflict, "DUPLICATE_EMAIL", "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}

	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
			ActorID: claims.UserID, ActorType: "ADMIN", Action: "admin_user.created",
			ResourceType: "admin_user", ResourceID: &user.ID,
			IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
		})
	}

	// Send invite email to the new admin user with temporary password
	if h.notifier != nil {
		h.notifier.SendEmail(r.Context(), notification.SendEmailInput{
			MerchantID: uuid.Nil,
			Recipient:  req.Email,
			Subject:    "You've been invited to Open Pay Admin",
			Body: fmt.Sprintf(
				`<p>Hi <strong>%s</strong>,</p>
				<p>You've been invited as an administrator on the Open Pay platform.</p>
				<p>Here are your login credentials:</p>
				<div style="background: #f3f4f6; padding: 16px; border-radius: 8px; margin: 16px 0;">
					<p style="margin: 4px 0;"><strong>Email:</strong> %s</p>
					<p style="margin: 4px 0;"><strong>Temporary Password:</strong> <code style="background: #e5e7eb; padding: 2px 6px; border-radius: 4px;">%s</code></p>
				</div>
				<p><strong>You will be required to change your password on first login.</strong></p>
				<p>If you did not expect this invitation, please ignore this email.</p>`,
				req.Name, req.Email, req.Password,
			),
			EventType: "admin.invite",
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": adminUserListResponse(user)})
}

func (h *AdminHandler) UpdateAdminUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid user ID")
		return
	}

	var req struct {
		Name   *string `json:"name"`
		RoleID *string `json:"roleId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid role ID")
			return
		}
		user.RoleID = roleID
	}

	if err := h.userRepo.UpdateUser(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "updated"}})
}

func (h *AdminHandler) DeactivateAdminUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	user.IsActive = !user.IsActive
	if err := h.userRepo.UpdateUser(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}

	action := "admin_user.deactivated"
	if user.IsActive {
		action = "admin_user.activated"
	}

	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		h.auditSvc.CreateLog(r.Context(), domain.AuditInput{
			ActorID: claims.UserID, ActorType: "ADMIN", Action: action,
			ResourceType: "admin_user", ResourceID: &id,
			IPAddress: stripPort(r.RemoteAddr), UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": action}})
}

// --- Admin Role Handlers ---

func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.userRepo.ListRoles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list roles")
		return
	}

	items := make([]map[string]any, 0, len(roles))
	for _, role := range roles {
		items = append(items, map[string]any{
			"id": role.ID.String(), "name": role.Name, "description": role.Description,
			"permissions": role.Permissions, "isSystem": role.IsSystem,
			"createdAt": role.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required")
		return
	}

	role := &domain.AdminRole{
		ID: uuid.New(), Name: req.Name, Description: req.Description,
		Permissions: req.Permissions, IsSystem: false, CreatedAt: time.Now().UTC(),
	}

	if err := h.userRepo.CreateRole(r.Context(), role); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create role")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": map[string]any{
		"id": role.ID.String(), "name": role.Name, "permissions": role.Permissions,
	}})
}

func (h *AdminHandler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid role ID")
		return
	}

	var req struct {
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	role := &domain.AdminRole{ID: id, Description: req.Description, Permissions: req.Permissions}
	if err := h.userRepo.UpdateRole(r.Context(), role); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update role")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "updated"}})
}

func adminUserListResponse(u *domain.AdminUser) map[string]any {
	resp := map[string]any{
		"id": u.ID.String(), "email": u.Email, "name": u.Name,
		"isActive": u.IsActive, "createdAt": u.CreatedAt.Format(time.RFC3339),
	}
	if u.Role != nil {
		resp["role"] = map[string]any{"id": u.Role.ID.String(), "name": u.Role.Name, "permissions": u.Role.Permissions}
	}
	if u.LastLoginAt != nil {
		resp["lastLoginAt"] = u.LastLoginAt.Format(time.RFC3339)
	}
	return resp
}

func legalDocResponse(d *postgres.LegalDocument) map[string]any {
	resp := map[string]any{
		"id":        d.ID.String(),
		"type":      d.Type,
		"version":   d.Version,
		"title":     d.Title,
		"content":   d.Content,
		"isActive":  d.IsActive,
		"createdAt": d.CreatedAt.Format(time.RFC3339),
		"updatedAt": d.UpdatedAt.Format(time.RFC3339),
	}
	if d.CreatedBy != nil {
		resp["createdBy"] = d.CreatedBy.String()
	}
	return resp
}
