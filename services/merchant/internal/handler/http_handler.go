package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	minio "github.com/minio/minio-go/v7"
	"github.com/openlankapay/openlankapay/pkg/audit"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
)

// MerchantServiceInterface defines the operations the handler depends on.
type MerchantServiceInterface interface {
	Register(ctx context.Context, input service.RegisterInput) (*domain.Merchant, error)
	RegisterWithUser(ctx context.Context, input service.RegisterWithUserInput) (*service.LoginResult, error)
	Login(ctx context.Context, email, password string) (*service.LoginResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*service.LoginResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateMerchantProfile(ctx context.Context, id uuid.UUID, input service.UpdateProfileInput) (*domain.Merchant, error)
	Approve(ctx context.Context, id uuid.UUID, force bool, forceReason string) error
	Reject(ctx context.Context, id uuid.UUID, reason string) error
	List(ctx context.Context, params service.ListParams) ([]*domain.Merchant, int, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	Freeze(ctx context.Context, id uuid.UUID, reason string) error
	Unfreeze(ctx context.Context, id uuid.UUID) error
	Terminate(ctx context.Context, id uuid.UUID, reason string) error
	CreateDirector(ctx context.Context, merchantID uuid.UUID, email string) (*domain.Director, error)
	ListDirectors(ctx context.Context, merchantID uuid.UUID) ([]*domain.Director, error)
	ResendDirectorVerification(ctx context.Context, merchantID, directorID uuid.UUID) error
	RemoveDirector(ctx context.Context, merchantID, directorID uuid.UUID) error
	GetDirectorByToken(ctx context.Context, token string) (*domain.Director, *domain.Merchant, error)
	SubmitDirectorVerification(ctx context.Context, token string, input service.SubmitDirectorInput) (*domain.Director, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	SetupTOTP(ctx context.Context, userID uuid.UUID) (secret, uri string, err error)
	VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) ([]string, error)
	DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error
	CreateAPIKey(ctx context.Context, merchantID uuid.UUID, env, name string) (*domain.APIKey, string, error)
	RevokeAPIKey(ctx context.Context, id uuid.UUID, reason string) error
	ListAPIKeys(ctx context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error)
	ValidateAPIKey(ctx context.Context, keyID, secret string) (*domain.Merchant, error)
}

// MerchantHandler handles HTTP requests for merchant operations.
type MerchantHandler struct {
	svc         MerchantServiceInterface
	jwtSecret   string
	auditLog    *audit.Client
	docRepo     DocumentRepository
	minioClient *minio.Client
	minioBucket string
}

// NewMerchantHandler creates a new MerchantHandler.
func NewMerchantHandler(svc MerchantServiceInterface, jwtSecret string, auditClient *audit.Client, docRepo DocumentRepository, minioClient *minio.Client, minioBucket string) *MerchantHandler {
	return &MerchantHandler{
		svc:         svc,
		jwtSecret:   jwtSecret,
		auditLog:    auditClient,
		docRepo:     docRepo,
		minioClient: minioClient,
		minioBucket: minioBucket,
	}
}

// NewRouter creates a chi router with merchant routes.
func NewRouter(h *MerchantHandler, branchRepo branchRepo, paymentLinkRepo paymentLinkRepo, uploadHandler ...*FileUploadHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public auth routes
	r.Post("/v1/auth/register", h.RegisterWithUser)
	r.Post("/v1/auth/login", h.Login)
	r.Post("/v1/auth/refresh", h.RefreshToken)

	// Public director verification routes (no auth required)
	r.Get("/v1/public/directors/verify/{token}", h.GetDirectorByToken)
	r.Post("/v1/public/directors/verify/{token}", h.SubmitDirectorVerification)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(h.jwtSecret))

		r.Get("/v1/auth/me", h.GetMe)
		r.Post("/v1/auth/change-password", h.ChangePassword)
		r.Post("/v1/auth/2fa/setup", h.SetupTOTP)
		r.Post("/v1/auth/2fa/verify", h.VerifyTOTP)
		r.Post("/v1/auth/2fa/disable", h.DisableTOTP)
		r.Get("/v1/merchants", h.ListMerchants)
		r.Put("/v1/merchants/{id}", h.UpdateProfile)
		r.Get("/v1/merchants/{id}", h.GetByID)
		r.Post("/v1/merchants/{id}/approve", h.ApproveMerchant)
		r.Post("/v1/merchants/{id}/reject", h.RejectMerchant)
		r.Post("/v1/merchants/{id}/deactivate", h.DeactivateMerchant)
		r.Post("/v1/merchants/{id}/freeze", h.FreezeMerchant)
		r.Post("/v1/merchants/{id}/unfreeze", h.UnfreezeMerchant)
		r.Post("/v1/merchants/{id}/terminate", h.TerminateMerchant)
		r.Get("/v1/merchants/{id}/documents", h.GetMerchantDocuments)
		r.Post("/v1/merchants/{id}/directors", h.CreateDirector)
		r.Get("/v1/merchants/{id}/directors", h.ListDirectors)
		r.Post("/v1/merchants/{id}/directors/{directorId}/resend", h.ResendDirectorVerification)
		r.Delete("/v1/merchants/{id}/directors/{directorId}", h.RemoveDirector)

		// API key management
		r.Post("/v1/api-keys", h.CreateAPIKey)
		r.Get("/v1/api-keys", h.ListAPIKeys)
		r.Delete("/v1/api-keys/{id}", h.RevokeAPIKey)
	})

	// Branch routes
	if branchRepo != nil {
		RegisterBranchRoutes(r, branchRepo, h.jwtSecret)
	}

	// Payment link routes
	if paymentLinkRepo != nil {
		RegisterPaymentLinkRoutes(r, paymentLinkRepo, h.jwtSecret, h.auditLog)
	}

	// Upload routes
	if len(uploadHandler) > 0 && uploadHandler[0] != nil {
		RegisterUploadRoutes(r, uploadHandler[0], h.jwtSecret)
	}

	return r
}

// RegisterInternalRoutes adds internal service-to-service endpoints (no auth).
func RegisterInternalRoutes(r chi.Router, ih *InternalHandler) {
	r.Post("/internal/api-keys/validate", ih.ValidateAPIKey)
}

// RegisterWithUser handles POST /v1/auth/register.
func (h *MerchantHandler) RegisterWithUser(w http.ResponseWriter, r *http.Request) {
	var req registerWithUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	result, err := h.svc.RegisterWithUser(r.Context(), service.RegisterWithUserInput{
		BusinessName: req.BusinessName,
		ContactEmail: req.Email,
		ContactPhone: req.Phone,
		ContactName:  req.Name,
		AdminEmail:   req.Email,
		AdminPassword: req.Password,
		AdminName:    req.Name,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidMerchant) || errors.Is(err, domain.ErrInvalidUser) || errors.Is(err, domain.ErrWeakPassword) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, domain.ErrDuplicateEmail) {
			writeError(w, http.StatusConflict, "DUPLICATE_EMAIL", "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register")
		return
	}

	if h.auditLog != nil {
		merchantID := result.Merchant.ID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: result.User.ID, ActorType: "MERCHANT_USER", MerchantID: &merchantID,
			Action: "merchant.registered", ResourceType: "merchant", ResourceID: &merchantID,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusCreated, envelope{"data": authResponse(result)})
}

// Login handles POST /v1/auth/login.
func (h *MerchantHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
			return
		}
		if errors.Is(err, service.ErrAccountInactive) {
			writeError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", "account is inactive")
			return
		}
		if errors.Is(err, service.ErrAccountTerminated) {
			writeError(w, http.StatusForbidden, "ACCOUNT_TERMINATED", "your merchant account has been terminated")
			return
		}
		if errors.Is(err, service.ErrAccountFrozen) {
			writeError(w, http.StatusForbidden, "ACCOUNT_FROZEN", "your merchant account has been frozen, please contact support")
			return
		}
		if errors.Is(err, service.ErrAccountNotActive) {
			writeError(w, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE", "your merchant account is not active")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to login")
		return
	}

	if h.auditLog != nil {
		merchantID := result.Merchant.ID
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: result.User.ID, ActorType: "MERCHANT_USER", MerchantID: &merchantID,
			Action: "merchant.login", ResourceType: "user", ResourceID: &result.User.ID,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": authResponse(result)})
}

// RefreshToken handles POST /v1/auth/refresh.
func (h *MerchantHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	result, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired refresh token")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": authResponse(result)})
}

// GetMe handles GET /v1/auth/me.
func (h *MerchantHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	merchant, err := h.svc.GetByID(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
		return
	}

	// Force logout if merchant account is no longer active
	switch merchant.Status {
	case domain.MerchantTerminated:
		writeError(w, http.StatusForbidden, "ACCOUNT_TERMINATED", "your merchant account has been terminated")
		return
	case domain.MerchantFrozen:
		writeError(w, http.StatusForbidden, "ACCOUNT_FROZEN", "your merchant account has been frozen")
		return
	case domain.MerchantInactive:
		writeError(w, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE", "your merchant account is not active")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": meResponse(user, merchant)})
}

// ChangePassword handles POST /v1/auth/change-password.
func (h *MerchantHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", "current password is incorrect")
			return
		}
		if errors.Is(err, domain.ErrWeakPassword) {
			writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to change password")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "password_changed"}})
}

// SetupTOTP handles POST /v1/auth/2fa/setup.
func (h *MerchantHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	secret, uri, err := h.svc.SetupTOTP(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SETUP_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{
		"secret": secret,
		"uri":    uri,
	}})
}

// VerifyTOTP handles POST /v1/auth/2fa/verify.
func (h *MerchantHandler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	backupCodes, err := h.svc.VerifyAndEnableTOTP(r.Context(), claims.UserID, req.Code)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTOTP) {
			writeError(w, http.StatusBadRequest, "INVALID_CODE", "invalid 2FA code")
			return
		}
		writeError(w, http.StatusBadRequest, "VERIFY_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]any{
		"enabled":     true,
		"backupCodes": backupCodes,
	}})
}

// DisableTOTP handles POST /v1/auth/2fa/disable.
func (h *MerchantHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if err := h.svc.DisableTOTP(r.Context(), claims.UserID, req.Code); err != nil {
		if errors.Is(err, domain.ErrInvalidTOTP) {
			writeError(w, http.StatusBadRequest, "INVALID_CODE", "invalid 2FA code")
			return
		}
		writeError(w, http.StatusBadRequest, "DISABLE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "2fa_disabled"}})
}

// UpdateProfile handles PUT /v1/merchants/{id}.
func (h *MerchantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID format")
		return
	}

	if claims.MerchantID != id {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot update another merchant's profile")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchant, err := h.svc.UpdateMerchantProfile(r.Context(), id, service.UpdateProfileInput{
		BusinessName:    req.BusinessName,
		BusinessType:    req.BusinessType,
		RegistrationNo:  req.RegistrationNo,
		Website:         req.Website,
		ContactPhone:    req.ContactPhone,
		ContactName:     req.ContactName,
		AddressLine1:    req.AddressLine1,
		AddressLine2:    req.AddressLine2,
		City:            req.City,
		District:        req.District,
		PostalCode:      req.PostalCode,
		BankName:        req.BankName,
		BankBranch:      req.BankBranch,
		BankAccountNo:   req.BankAccountNo,
		BankAccountName: req.BankAccountName,
		SubmitKYC:       req.SubmitKYC,
	})
	if err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidKYCTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_KYC_TRANSITION", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update profile")
		return
	}

	if h.auditLog != nil {
		action := "merchant.profile_updated"
		if req.SubmitKYC {
			action = "merchant.kyc_submitted"
		}
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: claims.UserID, ActorType: "MERCHANT_USER", MerchantID: &id,
			Action: action, ResourceType: "merchant", ResourceID: &id,
			IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": merchantResponse(merchant)})
}

// GetByID handles GET /v1/merchants/{id}.
func (h *MerchantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID format")
		return
	}

	merchant, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get merchant")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": merchantResponse(merchant)})
}

// ListMerchants handles GET /v1/merchants.
func (h *MerchantHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	params := service.ListParams{
		Page:    intQuery(r, "page", 1),
		PerPage: intQuery(r, "perPage", 20),
		Search:  r.URL.Query().Get("search"),
	}
	if v := r.URL.Query().Get("kycStatus"); v != "" {
		s := domain.KYCStatus(v)
		params.KYCStatus = &s
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := domain.MerchantStatus(v)
		params.Status = &s
	}

	merchants, total, err := h.svc.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list merchants")
		return
	}

	items := make([]map[string]any, 0, len(merchants))
	for _, m := range merchants {
		items = append(items, merchantResponse(m))
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": items,
		"meta": map[string]any{"total": total, "page": params.Page, "perPage": params.PerPage},
	})
}

// ApproveMerchant handles POST /v1/merchants/{id}/approve.
func (h *MerchantHandler) ApproveMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	var req struct {
		Force       bool   `json:"force"`
		ForceReason string `json:"forceReason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.Approve(r.Context(), id, req.Force, req.ForceReason); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidKYCTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
			return
		}
		if errors.Is(err, domain.ErrDirectorsNotVerified) {
			writeError(w, http.StatusUnprocessableEntity, "DIRECTORS_NOT_VERIFIED", "not all directors have been verified; use force=true to override")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to approve merchant")
		return
	}

	action := "merchant.approved"
	if req.Force {
		action = "merchant.force_approved"
	}

	if h.auditLog != nil {
		var actorID uuid.UUID
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			actorID = claims.UserID
		}
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: actorID, ActorType: "ADMIN", Action: action,
			ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "approved"}})
}

// RejectMerchant handles POST /v1/merchants/{id}/reject.
func (h *MerchantHandler) RejectMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Rejected by admin"
	}

	if err := h.svc.Reject(r.Context(), id, req.Reason); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reject merchant")
		return
	}

	if h.auditLog != nil {
		var actorID uuid.UUID
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			actorID = claims.UserID
		}
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: actorID, ActorType: "ADMIN", Action: "merchant.rejected",
			ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "rejected"}})
}

// DeactivateMerchant handles POST /v1/merchants/{id}/deactivate.
func (h *MerchantHandler) DeactivateMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if err := h.svc.Deactivate(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to deactivate merchant")
		return
	}

	if h.auditLog != nil {
		var actorID uuid.UUID
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			actorID = claims.UserID
		}
		h.auditLog.Log(r.Context(), audit.LogEntry{
			ActorID: actorID, ActorType: "ADMIN", Action: "merchant.deactivated",
			ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "deactivated"}})
}

// FreezeMerchant handles POST /v1/merchants/{id}/freeze.
func (h *MerchantHandler) FreezeMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Frozen by admin"
	}

	if err := h.svc.Freeze(r.Context(), id, req.Reason); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to freeze merchant")
		return
	}

	if h.auditLog != nil {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			h.auditLog.Log(r.Context(), audit.LogEntry{
				ActorID: claims.UserID, ActorType: "ADMIN", Action: "merchant.frozen",
				ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
			})
		}
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "frozen"}})
}

// UnfreezeMerchant handles POST /v1/merchants/{id}/unfreeze.
func (h *MerchantHandler) UnfreezeMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if err := h.svc.Unfreeze(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to unfreeze merchant")
		return
	}

	if h.auditLog != nil {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			h.auditLog.Log(r.Context(), audit.LogEntry{
				ActorID: claims.UserID, ActorType: "ADMIN", Action: "merchant.unfrozen",
				ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
			})
		}
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "active"}})
}

// TerminateMerchant handles POST /v1/merchants/{id}/terminate.
func (h *MerchantHandler) TerminateMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Terminated by admin"
	}

	if err := h.svc.Terminate(r.Context(), id, req.Reason); err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "merchant not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			writeError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to terminate merchant")
		return
	}

	if h.auditLog != nil {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			h.auditLog.Log(r.Context(), audit.LogEntry{
				ActorID: claims.UserID, ActorType: "ADMIN", Action: "merchant.terminated",
				ResourceType: "merchant", ResourceID: &id, IPAddress: stripPort(r.RemoteAddr),
			})
		}
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "terminated"}})
}

// GetMerchantDocuments handles GET /v1/merchants/{id}/documents.
func (h *MerchantHandler) GetMerchantDocuments(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant ID")
		return
	}

	if h.docRepo == nil {
		writeJSON(w, http.StatusOK, envelope{"data": []any{}})
		return
	}

	docs, err := h.docRepo.ListByMerchant(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list documents")
		return
	}

	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		items = append(items, map[string]any{
			"id": d.ID.String(), "category": d.Category,
			"filename": d.Filename, "contentType": d.ContentType,
			"fileSize": d.FileSize, "objectKey": d.ObjectKey,
		})
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func stripPort(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
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

// --- Request types ---

type registerWithUserRequest struct {
	BusinessName string `json:"businessName"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type updateProfileRequest struct {
	BusinessName    *string `json:"businessName"`
	BusinessType    *string `json:"businessType"`
	RegistrationNo  *string `json:"registrationNo"`
	Website         *string `json:"website"`
	ContactPhone    *string `json:"contactPhone"`
	ContactName     *string `json:"contactName"`
	AddressLine1    *string `json:"addressLine1"`
	AddressLine2    *string `json:"addressLine2"`
	City            *string `json:"city"`
	District        *string `json:"district"`
	PostalCode      *string `json:"postalCode"`
	BankName        *string `json:"bankName"`
	BankBranch      *string `json:"bankBranch"`
	BankAccountNo   *string `json:"bankAccountNo"`
	BankAccountName *string `json:"bankAccountName"`
	SubmitKYC       bool    `json:"submitKyc"`
}

// --- Response helpers ---

type envelope map[string]any

func authResponse(r *service.LoginResult) map[string]any {
	return map[string]any{
		"accessToken":  r.AccessToken,
		"refreshToken": r.RefreshToken,
		"user":         userResponse(r.User),
		"merchant":     merchantResponse(r.Merchant),
	}
}

func userResponse(u *domain.User) map[string]any {
	resp := map[string]any{
		"id":       u.ID.String(),
		"email":    u.Email,
		"name":     u.Name,
		"role":     u.Role,
		"isActive": u.IsActive,
	}
	if u.BranchID != nil {
		resp["branchId"] = u.BranchID.String()
	}
	return resp
}

func meResponse(u *domain.User, m *domain.Merchant) map[string]any {
	return map[string]any{
		"user":     userResponse(u),
		"merchant": merchantResponse(m),
	}
}

func merchantResponse(m *domain.Merchant) map[string]any {
	resp := map[string]any{
		"id":              m.ID.String(),
		"businessName":    m.BusinessName,
		"businessType":    m.BusinessType,
		"registrationNo":  m.RegistrationNo,
		"website":         m.Website,
		"contactEmail":    m.ContactEmail,
		"contactPhone":    m.ContactPhone,
		"contactName":     m.ContactName,
		"addressLine1":    m.AddressLine1,
		"addressLine2":    m.AddressLine2,
		"city":            m.City,
		"district":        m.District,
		"postalCode":      m.PostalCode,
		"country":         m.Country,
		"kycStatus":       string(m.KYCStatus),
		"kycRejectionReason": m.KYCRejectionReason,
		"kycReviewNotes":  m.KYCReviewNotes,
		"bankName":        m.BankName,
		"bankBranch":      m.BankBranch,
		"bankAccountNo":   m.BankAccountNo,
		"bankAccountName": m.BankAccountName,
		"defaultCurrency": m.DefaultCurrency,
		"status":          string(m.Status),
		"statusReason":    m.StatusReason,
		"createdAt":       m.CreatedAt.Format(time.RFC3339),
	}
	if m.KYCSubmittedAt != nil {
		resp["kycSubmittedAt"] = m.KYCSubmittedAt.Format(time.RFC3339)
	}
	if m.KYCReviewedAt != nil {
		resp["kycReviewedAt"] = m.KYCReviewedAt.Format(time.RFC3339)
	}
	if m.StatusChangedAt != nil {
		resp["statusChangedAt"] = m.StatusChangedAt.Format(time.RFC3339)
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
