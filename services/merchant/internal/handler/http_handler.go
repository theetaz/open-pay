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
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
)

// MerchantServiceInterface defines the operations the handler depends on.
type MerchantServiceInterface interface {
	Register(ctx context.Context, input service.RegisterInput) (*domain.Merchant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
}

// MerchantHandler handles HTTP requests for merchant operations.
type MerchantHandler struct {
	svc MerchantServiceInterface
}

// NewMerchantHandler creates a new MerchantHandler.
func NewMerchantHandler(svc MerchantServiceInterface) *MerchantHandler {
	return &MerchantHandler{svc: svc}
}

// NewRouter creates a chi router with merchant routes.
func NewRouter(h *MerchantHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/v1/merchants", func(r chi.Router) {
		r.Post("/", h.Register)
		r.Get("/{id}", h.GetByID)
	})

	return r
}

// Register handles POST /v1/merchants.
func (h *MerchantHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	merchant, err := h.svc.Register(r.Context(), service.RegisterInput{
		BusinessName: req.BusinessName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		ContactName:  req.ContactName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidMerchant) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, domain.ErrDuplicateEmail) {
			writeError(w, http.StatusConflict, "DUPLICATE_EMAIL", "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register merchant")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": merchantResponse(merchant)})
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

// --- Request/Response types ---

type registerRequest struct {
	BusinessName string `json:"businessName"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	ContactName  string `json:"contactName"`
}

type envelope map[string]any

func merchantResponse(m *domain.Merchant) map[string]any {
	return map[string]any{
		"id":            m.ID.String(),
		"businessName":  m.BusinessName,
		"contactEmail":  m.ContactEmail,
		"contactPhone":  m.ContactPhone,
		"contactName":   m.ContactName,
		"kycStatus":     string(m.KYCStatus),
		"status":        string(m.Status),
		"country":       m.Country,
		"createdAt":     m.CreatedAt.Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
