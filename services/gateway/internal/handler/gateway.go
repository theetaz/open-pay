package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// GatewayConfig holds configuration for the gateway router.
type GatewayConfig struct {
	// RateLimitPerMinute is the default rate limit for public endpoints.
	RateLimitPerMinute int
}

// NewGatewayRouter creates the main API gateway router with all middleware.
func NewGatewayRouter(cfg GatewayConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(requestID)
	r.Use(cors)
	r.Use(chimw.RealIP)

	// Health endpoints (no auth)
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz)

	// API v1 routes (auth + rate limiting applied per group)
	r.Route("/v1", func(r chi.Router) {
		// Placeholder routes — will proxy to internal services
		r.Get("/payments", placeholder("payments.list"))
		r.Post("/payments", placeholder("payments.create"))
		r.Get("/payments/{id}", placeholder("payments.get"))
	})

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

func placeholder(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": name + " endpoint - not yet connected to service",
		})
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

// cors handles CORS preflight and response headers.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-api-key, x-timestamp, x-signature, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
