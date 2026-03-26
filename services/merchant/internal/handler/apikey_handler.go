package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/auth"
)

// CreateAPIKey handles POST /v1/api-keys.
func (h *MerchantHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if req.Environment == "" {
		req.Environment = "live"
	}

	key, plainSecret, err := h.svc.CreateAPIKey(r.Context(), claims.MerchantID, req.Environment, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create API key")
		return
	}

	// The full compound key (ak_xxx.sk_xxx) is only shown once at creation
	compoundKey := key.KeyID + "." + plainSecret

	writeJSON(w, http.StatusCreated, envelope{
		"data": map[string]any{
			"id":          key.ID.String(),
			"keyId":       key.KeyID,
			"secret":      compoundKey,
			"name":        key.Name,
			"environment": key.Environment,
			"createdAt":   key.CreatedAt,
		},
	})
}

// ListAPIKeys handles GET /v1/api-keys.
func (h *MerchantHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	keys, err := h.svc.ListAPIKeys(r.Context(), claims.MerchantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list API keys: "+err.Error())
		return
	}

	items := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		items = append(items, map[string]any{
			"id":          k.ID.String(),
			"keyId":       k.KeyID,
			"name":        k.Name,
			"environment": k.Environment,
			"isActive":    k.IsActive,
			"revokedAt":   k.RevokedAt,
			"lastUsedAt":  k.LastUsedAt,
			"createdAt":   k.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, envelope{"data": items})
}

// RevokeAPIKey handles DELETE /v1/api-keys/{id}.
func (h *MerchantHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid API key ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Revoked by merchant"
	}

	if err := h.svc.RevokeAPIKey(r.Context(), id, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to revoke API key")
		return
	}

	writeJSON(w, http.StatusOK, envelope{"data": map[string]string{"status": "revoked"}})
}
