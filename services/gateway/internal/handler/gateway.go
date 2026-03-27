package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/gateway/internal/middleware"
	"github.com/openlankapay/openlankapay/services/gateway/internal/proxy"
)

// GatewayConfig holds configuration for the gateway router.
type GatewayConfig struct {
	JWTSecret          string
	ServiceProxy       *proxy.ServiceProxy
	RateLimitPerMinute int
	RateLimiter        middleware.RateLimiter
	KeyValidator       middleware.KeyValidator
	GatewayPort        string
	PlatformFeePct     string
	ExchangeFeePct     string
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

	// Strip internal headers from external requests to prevent forgery
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("X-Internal-Admin")
			next.ServeHTTP(w, r)
		})
	})

	// Apply rate limiting
	if cfg.RateLimiter != nil {
		r.Use(middleware.RateLimit(cfg.RateLimiter))
	}

	// Health endpoints
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz)
	r.Get("/v1/system/health", systemHealth(cfg))
	r.Get("/v1/platform/config", platformConfig(cfg))

	p := cfg.ServiceProxy

	// Auth routes → merchant service (no auth required)
	r.Post("/v1/auth/register", p.ProxyToMerchant)
	r.Post("/v1/auth/login", p.ProxyToMerchant)
	r.Post("/v1/auth/refresh", p.ProxyToMerchant)

	// Merchant routes → merchant service (auth handled by merchant service)
	r.Get("/v1/auth/me", p.ProxyToMerchant)
	r.Post("/v1/auth/change-password", p.ProxyToMerchant)
	r.Post("/v1/auth/2fa/setup", p.ProxyToMerchant)
	r.Post("/v1/auth/2fa/verify", p.ProxyToMerchant)
	r.Post("/v1/auth/2fa/disable", p.ProxyToMerchant)
	r.Get("/v1/merchants", p.ProxyToMerchant)
	r.Put("/v1/merchants/{id}", p.ProxyToMerchant)
	r.Get("/v1/merchants/{id}", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/approve", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/reject", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/deactivate", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/freeze", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/unfreeze", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/terminate", p.ProxyToMerchant)
	r.Get("/v1/merchants/{id}/documents", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/directors", p.ProxyToMerchant)
	r.Get("/v1/merchants/{id}/directors", p.ProxyToMerchant)
	r.Post("/v1/merchants/{id}/directors/{directorId}/resend", p.ProxyToMerchant)
	r.Delete("/v1/merchants/{id}/directors/{directorId}", p.ProxyToMerchant)

	// API key management routes → merchant service (JWT auth)
	r.Post("/v1/api-keys", p.ProxyToMerchant)
	r.Get("/v1/api-keys", p.ProxyToMerchant)
	r.Delete("/v1/api-keys/{id}", p.ProxyToMerchant)

	// File upload routes → merchant service
	r.Post("/v1/uploads", p.ProxyToMerchant)
	r.Delete("/v1/uploads", p.ProxyToMerchant)

	// Branch routes → merchant service
	r.Post("/v1/branches", p.ProxyToMerchant)
	r.Get("/v1/branches", p.ProxyToMerchant)
	r.Get("/v1/branches/{id}", p.ProxyToMerchant)
	r.Delete("/v1/branches/{id}", p.ProxyToMerchant)

	// Payment link routes → merchant service
	r.Post("/v1/payment-links", p.ProxyToMerchant)
	r.Get("/v1/payment-links", p.ProxyToMerchant)
	r.Get("/v1/payment-links/check-slug/{slug}", p.ProxyToMerchant)
	r.Get("/v1/payment-links/{id}", p.ProxyToMerchant)
	r.Put("/v1/payment-links/{id}", p.ProxyToMerchant)
	r.Delete("/v1/payment-links/{id}", p.ProxyToMerchant)
	r.Get("/v1/public/payment-links/by-slug/{slug}", p.ProxyToMerchant)

	// Payment routes → payment service (auth handled by payment service)
	r.Post("/v1/payments", p.ProxyToPayment)
	r.Get("/v1/payments", p.ProxyToPayment)
	r.Get("/v1/payments/{id}", p.ProxyToPayment)
	r.Get("/v1/payments/export/csv", p.ProxyToPayment)

	// Public payment routes → payment service (no auth)
	r.Post("/v1/public/payments", p.ProxyToPayment)
	r.Get("/v1/payments/{id}/checkout", p.ProxyToPayment)
	r.Post("/v1/payments/{id}/callback", p.ProxyToPayment)

	// Public director verification routes → merchant service (no auth)
	r.Get("/v1/public/directors/verify/{token}", p.ProxyToMerchant)
	r.Post("/v1/public/directors/verify/{token}", p.ProxyToMerchant)

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
	r.Post("/v1/webhooks/test", p.ProxyToWebhook)
	r.Get("/v1/webhooks/deliveries", p.ProxyToWebhook)

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
	r.Get("/v1/admin/email-templates", p.ProxyToNotification)
	r.Post("/v1/admin/email-templates", p.ProxyToNotification)
	r.Put("/v1/admin/email-templates/{id}", p.ProxyToNotification)

	// Merchant audit logs → admin service (merchant JWT)
	r.Get("/v1/merchant/audit-logs", p.ProxyToAdmin)

	// Admin auth routes → admin service (public)
	r.Post("/v1/admin/auth/login", p.ProxyToAdmin)
	r.Post("/v1/admin/auth/refresh", p.ProxyToAdmin)

	// Admin protected routes → admin service
	r.Get("/v1/admin/auth/me", p.ProxyToAdmin)
	r.Post("/v1/admin/auth/change-password", p.ProxyToAdmin)
	r.Get("/v1/audit-logs", p.ProxyToAdmin)
	r.Get("/v1/audit-logs/{id}", p.ProxyToAdmin)

	// Admin uploads and assets → admin service
	r.Post("/v1/admin/uploads", p.ProxyToAdmin)
	r.Get("/v1/assets/*", p.ProxyToAdmin)

	// Platform settings → admin service
	r.Get("/v1/admin/settings", p.ProxyToAdmin)
	r.Put("/v1/admin/settings", p.ProxyToAdmin)
	r.Get("/v1/admin/settings/{category}", p.ProxyToAdmin)

	// Admin user management → admin service
	r.Get("/v1/admin/users", p.ProxyToAdmin)
	r.Post("/v1/admin/users", p.ProxyToAdmin)
	r.Put("/v1/admin/users/{id}", p.ProxyToAdmin)
	r.Post("/v1/admin/users/{id}/deactivate", p.ProxyToAdmin)

	// Admin role management → admin service
	r.Get("/v1/admin/roles", p.ProxyToAdmin)
	r.Post("/v1/admin/roles", p.ProxyToAdmin)
	r.Put("/v1/admin/roles/{id}", p.ProxyToAdmin)

	// Legal documents → admin service
	r.Get("/v1/legal-documents/active", p.ProxyToAdmin)
	r.Get("/v1/admin/legal-documents", p.ProxyToAdmin)
	r.Post("/v1/admin/legal-documents", p.ProxyToAdmin)
	r.Put("/v1/admin/legal-documents/{id}", p.ProxyToAdmin)
	r.Post("/v1/admin/legal-documents/{id}/activate", p.ProxyToAdmin)

	// Admin-scoped routes to non-admin services (gateway validates admin JWT)
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(cfg.JWTSecret))
		r.Use(auth.RequirePlatformAdmin())

		// Merchant management → merchant service
		r.Get("/v1/admin/merchants", rewritePath("/v1/merchants", p.ProxyToMerchant))
		r.Get("/v1/admin/merchants/{id}", rewritePath("/v1/merchants/{id}", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/approve", rewritePath("/v1/merchants/{id}/approve", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/reject", rewritePath("/v1/merchants/{id}/reject", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/deactivate", rewritePath("/v1/merchants/{id}/deactivate", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/freeze", rewritePath("/v1/merchants/{id}/freeze", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/unfreeze", rewritePath("/v1/merchants/{id}/unfreeze", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/terminate", rewritePath("/v1/merchants/{id}/terminate", p.ProxyToMerchant))
		r.Get("/v1/admin/merchants/{id}/documents", rewritePath("/v1/merchants/{id}/documents", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/directors", rewritePath("/v1/merchants/{id}/directors", p.ProxyToMerchant))
		r.Get("/v1/admin/merchants/{id}/directors", rewritePath("/v1/merchants/{id}/directors", p.ProxyToMerchant))
		r.Post("/v1/admin/merchants/{id}/directors/{directorId}/resend", rewritePath("/v1/merchants/{id}/directors/{directorId}/resend", p.ProxyToMerchant))
		r.Delete("/v1/admin/merchants/{id}/directors/{directorId}", rewritePath("/v1/merchants/{id}/directors/{directorId}", p.ProxyToMerchant))

		// Settlement/withdrawal management → settlement service
		r.Get("/v1/admin/settlements/balance", rewritePath("/v1/settlements/balance", p.ProxyToSettlement))
		r.Get("/v1/admin/withdrawals", rewritePath("/v1/withdrawals", p.ProxyToSettlement))
		r.Get("/v1/admin/withdrawals/{id}", rewritePath("/v1/withdrawals/{id}", p.ProxyToSettlement))
		r.Post("/v1/admin/withdrawals/{id}/approve", rewritePath("/v1/withdrawals/{id}/approve", p.ProxyToSettlement))
		r.Post("/v1/admin/withdrawals/{id}/reject", rewritePath("/v1/withdrawals/{id}/reject", p.ProxyToSettlement))
		r.Post("/v1/admin/withdrawals/{id}/complete", rewritePath("/v1/withdrawals/{id}/complete", p.ProxyToSettlement))
	})

	// SDK routes — HMAC-authenticated for API key users
	if cfg.KeyValidator != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.HMACAuth(cfg.KeyValidator))

			// Payment operations via SDK
			r.Post("/v1/sdk/payments", rewritePathSDK("/v1/payments", p.ProxyToPayment))
			r.Get("/v1/sdk/payments/{id}", rewritePathSDK("/v1/payments/{id}", p.ProxyToPayment))
			r.Get("/v1/sdk/payments", rewritePathSDK("/v1/payments", p.ProxyToPayment))

			// Checkout sessions via SDK
			r.Post("/v1/sdk/checkout/sessions", rewritePathSDK("/v1/checkout/sessions", p.ProxyToPayment))

			// Webhook configuration via SDK
			r.Post("/v1/sdk/webhooks/configure", rewritePathSDK("/v1/webhooks/configure", p.ProxyToWebhook))
			r.Get("/v1/sdk/webhooks/public-key", rewritePathSDK("/v1/webhooks/public-key", p.ProxyToWebhook))
			r.Post("/v1/sdk/webhooks/test", rewritePathSDK("/v1/webhooks/test", p.ProxyToWebhook))
		})
	}

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

// platformConfig returns platform fee configuration.
func platformConfig(cfg GatewayConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"platformFeePct": cfg.PlatformFeePct,
				"exchangeFeePct": cfg.ExchangeFeePct,
			},
		})
	}
}

// serviceHealthCheck checks a downstream service's health by making a TCP connection.
func serviceHealthCheck(serviceURL string) map[string]any {
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	// Use HEAD to a known base path — even a 404 means the service is up and responding
	resp, err := client.Head(serviceURL + "/")
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return map[string]any{"status": "unhealthy", "responseTime": elapsed}
	}
	defer func() { _ = resp.Body.Close() }()

	// Any HTTP response means the service is running
	return map[string]any{"status": "healthy", "responseTime": elapsed}
}

// systemHealth checks all downstream services and returns their status.
func systemHealth(cfg GatewayConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		services := map[string]string{
			"gateway":      "http://localhost:" + cfg.GatewayPort,
			"payment":      cfg.ServiceProxy.PaymentURL,
			"merchant":     cfg.ServiceProxy.MerchantURL,
			"settlement":   cfg.ServiceProxy.SettlementURL,
			"webhook":      cfg.ServiceProxy.WebhookURL,
			"exchange":     cfg.ServiceProxy.ExchangeURL,
			"subscription": cfg.ServiceProxy.SubscriptionURL,
			"notification": cfg.ServiceProxy.NotificationURL,
			"admin":        cfg.ServiceProxy.AdminURL,
		}

		results := make(map[string]any, len(services))
		type result struct {
			name string
			data map[string]any
		}
		ch := make(chan result, len(services))

		for name, url := range services {
			go func(n, u string) {
				if n == "gateway" {
					ch <- result{n, map[string]any{"status": "healthy", "responseTime": int64(0)}}
					return
				}
				ch <- result{n, serviceHealthCheck(u)}
			}(name, url)
		}

		for range services {
			r := <-ch
			results[r.name] = r.data
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": results})
	}
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

// rewritePath returns a handler that rewrites the request URL path before proxying.
// The template can contain chi URL params like {id} which are replaced from the route context.
// The Authorization header is removed because the gateway already validated the admin JWT;
// downstream services may use a different JWT secret and would reject it.
func rewritePath(template string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newPath := template
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			for i, key := range rctx.URLParams.Keys {
				if i < len(rctx.URLParams.Values) {
					newPath = strings.Replace(newPath, "{"+key+"}", rctx.URLParams.Values[i], 1)
				}
			}
		}
		r.URL.Path = newPath
		r.Header.Del("Authorization")
		r.Header.Set("X-Internal-Admin", "true")
		next.ServeHTTP(w, r)
	}
}

// rewritePathSDK rewrites the URL path for SDK routes.
// Sets X-Internal-Admin to bypass downstream JWT middleware (gateway already validated HMAC).
func rewritePathSDK(template string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newPath := template
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			for i, key := range rctx.URLParams.Keys {
				if i < len(rctx.URLParams.Values) {
					newPath = strings.Replace(newPath, "{"+key+"}", rctx.URLParams.Values[i], 1)
				}
			}
		}
		r.URL.Path = newPath
		r.Header.Del("Authorization")
		r.Header.Set("X-Internal-Admin", "true")
		next.ServeHTTP(w, r)
	}
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
