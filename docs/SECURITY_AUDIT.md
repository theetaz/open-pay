# OWASP Security Audit — Open Pay

**Date:** 2026-03-27
**Auditor:** Automated + Manual Review

## OWASP Top 10 Assessment

### 1. Injection (A03:2021)
**Status:** PASS
- All SQL queries use parameterized queries via `pgx` — no string interpolation
- Search parameters are passed as `$1`, `$2` positional arguments
- No raw SQL concatenation found in any repository file

### 2. Broken Authentication (A07:2021)
**Status:** PASS
- JWT tokens with configurable expiry (access + refresh rotation)
- HMAC-SHA256 API authentication with 5-minute timestamp window
- Refresh token rotation prevents replay
- 2FA/TOTP support for dashboard users
- API secrets shown only once at creation, stored as bcrypt hash

### 3. Sensitive Data Exposure (A02:2021)
**Status:** PASS
- API secrets: bcrypt-hashed at rest, plain text shown once at creation
- ED25519 private keys: stored encrypted
- No PII in structured logs (zerolog configured per-service)
- HTTPS enforced via HSTS header (max-age=31536000)

### 4. XML External Entities (A05:2021)
**Status:** N/A — No XML processing in the platform

### 5. Broken Access Control (A01:2021)
**Status:** PASS
- Merchant isolation: all queries filter by `merchant_id` from JWT claims
- Gateway strips `X-Internal-Admin` from external requests
- Admin routes require `RequirePlatformAdmin()` middleware
- RBAC roles: ADMIN, MANAGER, USER with branch-level scoping

### 6. Security Misconfiguration (A05:2021)
**Status:** FIXED
- **Added:** X-Content-Type-Options: nosniff
- **Added:** X-Frame-Options: DENY
- **Added:** X-XSS-Protection: 1; mode=block
- **Added:** Referrer-Policy: strict-origin-when-cross-origin
- **Added:** Permissions-Policy: camera=(), microphone=(), geolocation=()
- **Added:** Strict-Transport-Security: max-age=31536000; includeSubDomains
- **Added:** Cache-Control: no-store (for API responses)
- CORS currently allows all origins (`*`) — acceptable for public API

### 7. Cross-Site Scripting (A03:2021)
**Status:** PASS
- React frontend auto-escapes JSX output
- No `dangerouslySetInnerHTML` usage found
- API returns JSON only, no HTML rendering

### 8. Insecure Deserialization (A08:2021)
**Status:** PASS
- JSON-only API (encoding/json standard library)
- Strict Go typing prevents unexpected type injection
- No object serialization/deserialization beyond JSON

### 9. Known Vulnerabilities (A06:2021)
**Status:** PASS
- `govulncheck` — no critical vulnerabilities in Go dependencies
- CI pipeline includes gosec, govulncheck, and Trivy container scan
- Dependencies regularly updated

### 10. Insufficient Logging (A09:2021)
**Status:** PASS
- Structured logging via zerolog on all services
- Audit log service tracks all admin actions
- Auth failures logged with request context
- Payment lifecycle events published to NATS + audit log

## Security Headers Added

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Cache-Control: no-store
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

## Recommendations for Future

1. Restrict CORS origins in production (replace `*` with specific domains)
2. Add Content-Security-Policy header for frontend served pages
3. Implement request body size limits per endpoint
4. Add account lockout after N failed login attempts
