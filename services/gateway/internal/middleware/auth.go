package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/openlankapay/openlankapay/pkg/auth"
)

var (
	ErrKeyNotFound = errors.New("API key not found")

	// timestampWindowMs is the allowed timestamp deviation (5 minutes).
	timestampWindowMs int64 = 5 * 60 * 1000
)

// KeyValidator looks up the secret for a given API key ID.
type KeyValidator interface {
	GetSecretByKeyID(keyID string) (string, error)
}

// HMACAuth returns middleware that validates HMAC-SHA256 signed requests.
func HMACAuth(validator KeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("x-api-key")
			timestamp := r.Header.Get("x-timestamp")
			signature := r.Header.Get("x-signature")

			if apiKey == "" || timestamp == "" || signature == "" {
				writeAuthError(w, "missing required auth headers (x-api-key, x-timestamp, x-signature)")
				return
			}

			if !auth.IsTimestampValid(timestamp, timestampWindowMs) {
				writeAuthError(w, "timestamp expired or invalid")
				return
			}

			secret, err := validator.GetSecretByKeyID(apiKey)
			if err != nil {
				writeAuthError(w, "invalid API key")
				return
			}

			// Read body for signature verification, then restore it
			var body string
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					writeAuthError(w, "failed to read request body")
					return
				}
				body = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			method := strings.ToUpper(r.Method)
			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = path + "?" + r.URL.RawQuery
			}

			if !auth.VerifySignatureWithHMACKey(secret, timestamp, method, path, body, signature) {
				writeAuthError(w, "invalid signature")
				return
			}

			// IP whitelisting check
			if rv, ok := validator.(interface{ GetAllowedIPs(string) []string }); ok {
				allowedIPs := rv.GetAllowedIPs(apiKey)
				if len(allowedIPs) > 0 {
					clientIP := extractClientIP(r)
					if !isIPAllowed(clientIP, allowedIPs) {
						writeForbiddenError(w, "IP address not allowed for this API key")
						return
					}
				}
			}

			// Pass merchant info to downstream services via headers
			if rv, ok := validator.(interface{ GetMerchantID(string) string }); ok {
				r.Header.Set("X-Merchant-ID", rv.GetMerchantID(apiKey))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// RemoteAddr is "ip:port"
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func isIPAllowed(clientIP string, allowedIPs []string) bool {
	for _, ip := range allowedIPs {
		if strings.TrimSpace(ip) == clientIP {
			return true
		}
	}
	return false
}

func writeForbiddenError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "IP_NOT_ALLOWED",
			"message": message,
		},
	})
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "UNAUTHORIZED",
			"message": message,
		},
	})
}
