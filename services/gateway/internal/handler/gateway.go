package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/gateway/internal/proxy"
)

// GatewayConfig holds configuration for the gateway router.
type GatewayConfig struct {
	JWTSecret          string
	ServiceProxy       *proxy.ServiceProxy
	RateLimitPerMinute int
}

// NewGatewayRouter creates the main API gateway router that proxies to downstream services.
func NewGatewayRouter(cfg GatewayConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(requestID)
	r.Use(corsMiddleware)
	r.Use(chimw.RealIP)

	// Health endpoints
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz)

	p := cfg.ServiceProxy

	// Auth routes → merchant service (no auth required)
	r.Post("/v1/auth/register", p.ProxyToMerchant)
	r.Post("/v1/auth/login", p.ProxyToMerchant)
	r.Post("/v1/auth/refresh", p.ProxyToMerchant)

	// Merchant routes → merchant service (auth handled by merchant service)
	r.Get("/v1/auth/me", p.ProxyToMerchant)
	r.Put("/v1/merchants/{id}", p.ProxyToMerchant)
	r.Get("/v1/merchants/{id}", p.ProxyToMerchant)

	// Payment routes → payment service (auth handled by payment service)
	r.Post("/v1/payments", p.ProxyToPayment)
	r.Get("/v1/payments", p.ProxyToPayment)
	r.Get("/v1/payments/{id}", p.ProxyToPayment)

	// Public payment routes → payment service (no auth)
	r.Get("/v1/payments/{id}/checkout", p.ProxyToPayment)
	r.Post("/v1/payments/{id}/callback", p.ProxyToPayment)

	// Sandbox routes → payment service (no auth, dev only)
	r.Post("/test/simulate/{providerPayID}", p.ProxyToPayment)

	// Exchange rate routes → exchange service (no auth)
	r.Get("/v1/exchange-rates/active", p.ProxyToExchange)

	// Settlement routes → settlement service (auth handled by settlement service)
	r.Get("/v1/settlements/balance", p.ProxyToSettlement)
	r.Post("/v1/settlements/credit", p.ProxyToSettlement)
	r.Post("/v1/withdrawals", p.ProxyToSettlement)
	r.Get("/v1/withdrawals", p.ProxyToSettlement)
	r.Get("/v1/withdrawals/{id}", p.ProxyToSettlement)
	r.Post("/v1/withdrawals/{id}/approve", p.ProxyToSettlement)
	r.Post("/v1/withdrawals/{id}/reject", p.ProxyToSettlement)
	r.Post("/v1/withdrawals/{id}/complete", p.ProxyToSettlement)

	// Webhook routes → webhook service (auth handled by webhook service)
	r.Post("/v1/webhooks/configure", p.ProxyToWebhook)
	r.Get("/v1/webhooks/public-key", p.ProxyToWebhook)

	// Subscription routes → subscription service
	r.Post("/v1/subscription-plans", p.ProxyToSubscription)
	r.Get("/v1/subscription-plans", p.ProxyToSubscription)
	r.Get("/v1/subscription-plans/{id}", p.ProxyToSubscription)
	r.Post("/v1/subscription-plans/{id}/archive", p.ProxyToSubscription)
	r.Post("/v1/subscription-plans/{id}/subscribe", p.ProxyToSubscription)
	r.Get("/v1/subscriptions", p.ProxyToSubscription)
	r.Get("/v1/subscriptions/{id}", p.ProxyToSubscription)
	r.Post("/v1/subscriptions/{id}/cancel", p.ProxyToSubscription)

	// Notification routes → notification service
	r.Get("/v1/notifications", p.ProxyToNotification)

	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func readyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// requestID adds or preserves X-Request-ID header.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS preflight and response headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key, x-timestamp, x-signature, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
