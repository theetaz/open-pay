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
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// BranchServiceInterface defines branch operations.
type BranchServiceInterface interface {
	CreateBranch(merchantID uuid.UUID, name, description, address, city string) (*domain.Branch, error)
	ListBranches(merchantID uuid.UUID) ([]*domain.Branch, error)
	GetBranch(id uuid.UUID) (*domain.Branch, error)
	DeleteBranch(id uuid.UUID) error
}

// BranchHandler handles branch-related HTTP requests.
type BranchHandler struct {
	repo branchRepo
}

type branchRepo interface {
	Create(ctx context.Context, b *domain.Branch) error
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Branch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// NewBranchHandler creates a new BranchHandler.
func NewBranchHandler(repo branchRepo) *BranchHandler {
	return &BranchHandler{repo: repo}
}

// RegisterBranchRoutes adds branch routes to the router.
func RegisterBranchRoutes(r chi.Router, repo branchRepo, jwtSecret string) {
	h := &BranchHandler{repo: repo}
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtSecret))
		r.Post("/v1/branches", h.CreateBranch)
		r.Get("/v1/branches", h.ListBranches)
		r.Get("/v1/branches/{id}", h.GetBranch)
		r.Delete("/v1/branches/{id}", h.DeleteBranch)
	})
}

func (h *BranchHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Address     string `json:"address"`
		City        string `json:"city"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	branch, err := domain.NewBranch(claims.MerchantID, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	branch.Description = req.Description
	branch.Address = req.Address
	branch.City = req.City

	if err := h.repo.Create(r.Context(), branch); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create branch")
		return
	}

	writeJSON(w, http.StatusCreated, envelope{"data": branchResponse(branch)})
}

func (h *BranchHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	branches, err := h.repo.ListByMerchant(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list branches")
		return
	}

	items := make([]map[string]any, 0, len(branches))
	for _, b := range branches {
		items = append(items, branchResponse(b))
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *BranchHandler) GetBranch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid branch ID")
		return
	}

	branch, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBranchNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "branch not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get branch")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": branchResponse(branch)})
}

func (h *BranchHandler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid branch ID")
		return
	}

	if err := h.repo.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrBranchNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "branch not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete branch")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "deleted"}})
}

func branchResponse(b *domain.Branch) map[string]any {
	return map[string]any{
		"id":          b.ID.String(),
		"merchantId":  b.MerchantID.String(),
		"name":        b.Name,
		"description": b.Description,
		"address":     b.Address,
		"city":        b.City,
		"isActive":    b.IsActive,
		"createdAt":   b.CreatedAt.Format(time.RFC3339),
	}
}
