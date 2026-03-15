package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-jwt-secret-key-at-least-32-chars"

func TestGenerateAndValidateToken(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	branchID := uuid.New()

	t.Run("valid token with all claims", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, merchantID, "ADMIN", &branchID, testSecret, 15*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := auth.ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, merchantID, claims.MerchantID)
		assert.Equal(t, "ADMIN", claims.Role)
		assert.Equal(t, &branchID, claims.BranchID)
		assert.Equal(t, "openlankapay", claims.Issuer)
	})

	t.Run("valid token without branch ID", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, merchantID, "USER", nil, testSecret, 15*time.Minute)
		require.NoError(t, err)

		claims, err := auth.ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.Nil(t, claims.BranchID)
	})

	t.Run("wrong secret fails validation", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, 15*time.Minute)
		require.NoError(t, err)

		_, err = auth.ValidateToken(token, "wrong-secret")
		require.Error(t, err)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("expired token fails validation", func(t *testing.T) {
		token, err := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, -1*time.Minute)
		require.NoError(t, err)

		_, err = auth.ValidateToken(token, testSecret)
		require.Error(t, err)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("malformed token fails validation", func(t *testing.T) {
		_, err := auth.ValidateToken("not.a.valid.token", testSecret)
		require.Error(t, err)
	})

	t.Run("empty token fails validation", func(t *testing.T) {
		_, err := auth.ValidateToken("", testSecret)
		require.Error(t, err)
	})

	t.Run("tokens are unique across different users", func(t *testing.T) {
		t1, _ := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, 15*time.Minute)
		t2, _ := auth.GenerateToken(uuid.New(), merchantID, "ADMIN", nil, testSecret, 15*time.Minute)
		assert.NotEqual(t, t1, t2)
	})
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	userID := uuid.New()

	t.Run("valid refresh token", func(t *testing.T) {
		token, err := auth.GenerateRefreshToken(userID, testSecret, 7*24*time.Hour)
		require.NoError(t, err)

		parsedID, err := auth.ValidateRefreshToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, userID, parsedID)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		token, err := auth.GenerateRefreshToken(userID, testSecret, -1*time.Minute)
		require.NoError(t, err)

		_, err = auth.ValidateRefreshToken(token, testSecret)
		require.Error(t, err)
	})

	t.Run("wrong secret", func(t *testing.T) {
		token, err := auth.GenerateRefreshToken(userID, testSecret, 7*24*time.Hour)
		require.NoError(t, err)

		_, err = auth.ValidateRefreshToken(token, "wrong-secret")
		require.Error(t, err)
	})
}

func TestJWTMiddleware(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	middleware := auth.JWTMiddleware(testSecret)

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.UserID.String()))
	})

	handler := middleware(protectedHandler)

	t.Run("valid token passes middleware", func(t *testing.T) {
		token, _ := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, 15*time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, userID.String(), rr.Body.String())
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid bearer format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic abc123")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("expired token rejected", func(t *testing.T) {
		token, _ := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, -1*time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestRequireRole(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allowed role passes", func(t *testing.T) {
		token, _ := auth.GenerateToken(userID, merchantID, "ADMIN", nil, testSecret, 15*time.Minute)

		handler := auth.JWTMiddleware(testSecret)(auth.RequireRole("ADMIN", "MANAGER")(okHandler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("disallowed role rejected", func(t *testing.T) {
		token, _ := auth.GenerateToken(userID, merchantID, "USER", nil, testSecret, 15*time.Minute)

		handler := auth.JWTMiddleware(testSecret)(auth.RequireRole("ADMIN")(okHandler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("no claims returns unauthorized", func(t *testing.T) {
		handler := auth.RequireRole("ADMIN")(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestClaimsFromContext(t *testing.T) {
	t.Run("no claims in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, ok := auth.ClaimsFromContext(req.Context())
		assert.False(t, ok)
	})
}
