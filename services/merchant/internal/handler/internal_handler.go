package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// HMACKeyLookup is the interface for looking up HMAC keys for gateway authentication.
type HMACKeyLookup interface {
	GetHMACKeyByKeyID(ctx context.Context, keyID string) (hmacKey string, merchantID uuid.UUID, err error)
}

// InternalHandler handles internal service-to-service endpoints.
type InternalHandler struct {
	keyLookup HMACKeyLookup
}

// NewInternalHandler creates a new internal handler.
func NewInternalHandler(keyLookup HMACKeyLookup) *InternalHandler {
	return &InternalHandler{keyLookup: keyLookup}
}

// ValidateAPIKey handles POST /internal/api-keys/validate.
// Called by the gateway to get the HMAC signing key for a given API key ID.
func (h *InternalHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		KeyID string `json:"keyId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "INVALID_JSON", "message": "malformed request body"},
		})
		return
	}

	hmacKey, merchantID, err := h.keyLookup.GetHMACKeyByKeyID(r.Context(), req.KeyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "NOT_FOUND", "message": "API key not found or inactive"},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"hmacKey":    hmacKey,
			"merchantId": merchantID.String(),
		},
	})
}
