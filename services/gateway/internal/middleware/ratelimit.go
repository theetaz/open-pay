package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter defines the contract for rate limiting.
type RateLimiter interface {
	Allow(key string) (allowed bool, limit int, remaining int, resetAt time.Time)
}

// RateLimit returns middleware that enforces per-client request limits.
func RateLimit(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractIP(r)

			allowed, limit, remaining, resetAt := limiter.Allow(clientIP)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetAt).Seconds())+1, 10))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "RATE_LIMITED",
						"message": "too many requests",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InMemoryRateLimiter is a simple sliding window rate limiter for development.
// In production, use Redis-backed rate limiting.
type InMemoryRateLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	clients map[string]*clientWindow
}

type clientWindow struct {
	count    int
	windowStart time.Time
}

// NewInMemoryRateLimiter creates a rate limiter with the given limit per minute.
func NewInMemoryRateLimiter(limitPerMinute int) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		limit:   limitPerMinute,
		window:  time.Minute,
		clients: make(map[string]*clientWindow),
	}
}

func (l *InMemoryRateLimiter) Allow(key string) (allowed bool, limit int, remaining int, resetAt time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	cw, exists := l.clients[key]
	if !exists || now.Sub(cw.windowStart) >= l.window {
		l.clients[key] = &clientWindow{count: 1, windowStart: now}
		return true, l.limit, l.limit - 1, now.Add(l.window)
	}

	cw.count++
	remaining = l.limit - cw.count
	resetAt = cw.windowStart.Add(l.window)

	if cw.count > l.limit {
		return false, l.limit, 0, resetAt
	}

	return true, l.limit, remaining, resetAt
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (behind load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := net.ParseIP(xff)
		if parts != nil {
			return xff
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
