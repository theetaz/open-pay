package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openlankapay/openlankapay/services/gateway/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := middleware.NewInMemoryRateLimiter(5) // 5 requests per window
		mw := middleware.RateLimit(limiter)
		protected := mw(handler)

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		limiter := middleware.NewInMemoryRateLimiter(2)
		mw := middleware.RateLimit(limiter)
		protected := mw(handler)

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Third request should be rate limited
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})

	t.Run("sets rate limit headers", func(t *testing.T) {
		limiter := middleware.NewInMemoryRateLimiter(10)
		mw := middleware.RateLimit(limiter)
		protected := mw(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)

		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("different clients have separate limits", func(t *testing.T) {
		limiter := middleware.NewInMemoryRateLimiter(1)
		mw := middleware.RateLimit(limiter)
		protected := mw(handler)

		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "10.0.0.1:12345"
		rec1 := httptest.NewRecorder()
		protected.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "10.0.0.2:12345"
		rec2 := httptest.NewRecorder()
		protected.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})
}
