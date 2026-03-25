package middleware_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/gateway/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHMACAuthMiddleware(t *testing.T) {
	secret := "sk_live_test_secret_for_middleware"

	// Pre-compute the HMAC key (SHA256 hex of the secret) as stored in DB
	h := sha256.Sum256([]byte(secret))
	hmacKeyHex := hex.EncodeToString(h[:])

	// A simple handler that returns 200 if auth passes
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Mock key validator that returns the pre-computed HMAC key
	validator := &mockKeyValidator{
		keys: map[string]string{
			"ak_live_testkey": hmacKeyHex,
		},
	}

	mw := middleware.HMACAuth(validator)
	protected := mw(handler)

	t.Run("valid signature passes", func(t *testing.T) {
		timestamp := auth.CurrentTimestamp()
		// Client signs with the plaintext secret (SDK behavior)
		signature := auth.SignRequest(secret, timestamp, "GET", "/v1/payments", "")

		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		req.Header.Set("x-api-key", "ak_live_testkey")
		req.Header.Set("x-timestamp", timestamp)
		req.Header.Set("x-signature", signature)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing api key returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("missing timestamp returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		req.Header.Set("x-api-key", "ak_live_testkey")
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("expired timestamp returns 401", func(t *testing.T) {
		oldTimestamp := "1000000000000"
		signature := auth.SignRequest(secret, oldTimestamp, "GET", "/v1/payments", "")

		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		req.Header.Set("x-api-key", "ak_live_testkey")
		req.Header.Set("x-timestamp", oldTimestamp)
		req.Header.Set("x-signature", signature)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		timestamp := auth.CurrentTimestamp()

		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		req.Header.Set("x-api-key", "ak_live_testkey")
		req.Header.Set("x-timestamp", timestamp)
		req.Header.Set("x-signature", "invalidsignature")
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unknown api key returns 401", func(t *testing.T) {
		timestamp := auth.CurrentTimestamp()
		signature := auth.SignRequest(secret, timestamp, "GET", "/v1/payments", "")

		req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
		req.Header.Set("x-api-key", "ak_live_unknown")
		req.Header.Set("x-timestamp", timestamp)
		req.Header.Set("x-signature", signature)
		rec := httptest.NewRecorder()

		protected.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// mockKeyValidator implements middleware.KeyValidator for testing.
type mockKeyValidator struct {
	keys map[string]string // keyID -> hmacKeyHex
}

func (m *mockKeyValidator) GetSecretByKeyID(keyID string) (string, error) {
	secret, ok := m.keys[keyID]
	if !ok {
		return "", middleware.ErrKeyNotFound
	}
	return secret, nil
}
