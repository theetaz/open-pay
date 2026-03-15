package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openlankapay/openlankapay/services/gateway/internal/handler"
	"github.com/openlankapay/openlankapay/services/gateway/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() handler.GatewayConfig {
	// Start a dummy backend that returns 200 for all requests
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"ok"}`))
	}))

	return handler.GatewayConfig{
		JWTSecret: "test-secret-at-least-32-chars-long",
		ServiceProxy: proxy.NewServiceProxy(proxy.Config{
			MerchantServiceURL:     backend.URL,
			PaymentServiceURL:      backend.URL,
			ExchangeServiceURL:     backend.URL,
			SettlementServiceURL:   backend.URL,
			WebhookServiceURL:      backend.URL,
			SubscriptionServiceURL: backend.URL,
			NotificationServiceURL: backend.URL,
			AdminServiceURL:        backend.URL,
		}),
	}
}

func TestHealthEndpoints(t *testing.T) {
	router := handler.NewGatewayRouter(testConfig())

	t.Run("liveness probe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("readiness probe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestCORSHeaders(t *testing.T) {
	router := handler.NewGatewayRouter(testConfig())

	t.Run("options preflight returns CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/v1/payments", nil)
		req.Header.Set("Origin", "https://merchant.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("CORS includes Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/v1/payments", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})
}

func TestRequestID(t *testing.T) {
	router := handler.NewGatewayRouter(testConfig())

	t.Run("adds request ID header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set("X-Request-ID", "my-custom-id")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, "my-custom-id", rec.Header().Get("X-Request-ID"))
	})
}

func TestProxyRoutes(t *testing.T) {
	router := handler.NewGatewayRouter(testConfig())

	t.Run("auth register proxies to merchant service", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Should get 200 from our dummy backend (not 404/405)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("exchange rate proxies to exchange service", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/exchange-rates/active", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
