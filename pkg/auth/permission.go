package auth

import (
	"net/http"
	"strings"
)

// RequirePlatformAdmin returns middleware that ensures the user has a PLATFORM_ prefixed role.
func RequirePlatformAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"missing authentication"}}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(claims.Role, "PLATFORM_") {
				http.Error(w, `{"error":{"code":"FORBIDDEN","message":"admin access required"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
