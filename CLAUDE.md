# CLAUDE.md — Open Pay Development Guide

## Project Overview

Open Pay is a crypto-to-fiat payment processing platform for Sri Lanka. Merchants accept crypto (USDT, USDC, BTC, ETH, BNB) and settle in LKR. Monorepo with Go microservices, React frontend, Solidity contracts, and multi-language SDKs.

## Quick Start

```bash
make up          # Start infrastructure (Postgres, Redis, NATS, MinIO, Mailpit)
make migrate     # Run all database migrations
make start       # Start all services + frontends with hot reload (uses air)
make stop        # Stop all running services
```

## Architecture

### Go Microservices (hexagonal architecture)

| Service | Port | Database | Purpose |
|---------|------|----------|---------|
| gateway | 8080 | — | API routing, HMAC/JWT auth, rate limiting |
| payment | 8081 | payment_db | Payment creation, tracking, QR codes |
| merchant | 8082 | merchant_db | Registration, KYC, API keys, branches |
| settlement | 8083 | settlement_db | Balances, withdrawals, refunds |
| webhook | 8084 | webhook_db | ED25519-signed delivery with retry |
| exchange | 8085 | exchange_db | Crypto/LKR rates from CoinGecko |
| subscription | 8086 | subscription_db | Recurring billing plans |
| notification | 8087 | notification_db | Email, SMS, push notifications |
| admin | 8088 | admin_db | Merchant approval, audit, settings |
| directdebit | 8089 | directdebit_db | Pre-authorized recurring charges |

### Frontend Apps

| App | Port | Stack |
|-----|------|-------|
| merchant-portal | 4600 | React 19, TanStack Query, shadcn/ui, Tailwind v4 |
| admin-dashboard | 4500 | Same stack |

### Infrastructure (Docker Compose)

- PostgreSQL 16 (port 5433) — multi-database via `scripts/init-multiple-dbs.sh`
- Redis 7 (port 6379)
- NATS with JetStream (port 4222)
- MinIO S3 (ports 9000/9001)
- Mailpit SMTP (ports 1025/8025)

## Service Structure Pattern

Every Go service follows the same hexagonal layout:

```
services/{name}/
├── cmd/main.go                    # Entry point, DI wiring
└── internal/
    ├── domain/                    # Entities, value objects, state machines, errors
    │   ├── {entity}.go
    │   └── {entity}_test.go
    ├── service/                   # Business logic, repository interfaces defined here
    │   └── {name}_service.go
    ├── adapter/
    │   ├── postgres/              # Repository implementations
    │   └── provider/              # External provider adapters (payment only)
    └── handler/
        └── http_handler.go        # Chi router, request/response, JWT middleware
```

## Key Patterns

### Authentication
- **Merchant API**: HMAC-SHA256 with `x-api-key`, `x-timestamp`, `x-signature` headers
- **Dashboard JWT**: Access + refresh tokens via `pkg/auth`. Claims: UserID, MerchantID, Role, BranchID
- **Webhooks**: ED25519 payload signing
- **Gateway proxy**: Sets `X-Internal-Admin: true` header after validating admin JWT; downstream services trust this

### Money & Currencies
- Always use `shopspring/decimal` — never float64 for monetary values
- Supported currencies: USDT, USDC, BTC, ETH, BNB, LKR
- Fee calculation via `pkg/money.CalculateFees()`
- Rates stored as crypto/LKR pairs

### State Machines
- Domain entities use explicit transition maps (e.g., `validTransitions map[Status][]Status`)
- Transition methods validate and return errors for invalid moves
- Terminal states have no outgoing transitions

### Error Handling
- Domain errors defined as `var Err... = errors.New(...)` at top of domain file
- Handlers check `errors.Is()` to map domain errors to HTTP status codes
- JSON error format: `{"error": {"code": "...", "message": "..."}}`

### Repository Pattern
- Interfaces defined in the service package (not a separate port package for newer services)
- Implementations in `adapter/postgres/` using `pgxpool.Pool`
- Use parameterized queries — never string interpolation
- Helper `scan*` functions for row mapping

## Commands

```bash
# Development
make start                    # Start everything with hot reload
make dev                      # Start infra only
make status                   # Check service status

# Build & Test
make build                    # Build all Go services to bin/
make test                     # Run unit tests (short mode)
make test-v                   # Verbose tests
make test-integration         # Integration tests (needs Docker)
make test-coverage            # Generate coverage report

# Code Quality
make lint                     # golangci-lint
make fmt                      # gofmt
make vet                      # go vet

# Database
make migrate                  # Run all migrations
make migrate-down             # Rollback last migration
make migrate-create svc=payment name=add_column  # Create new migration
make db-reset                 # Drop all + re-migrate

# Frontend
cd apps/merchant-portal && pnpm dev    # Vite dev server on :4600
cd apps/admin-dashboard && pnpm dev    # Vite dev server on :4500
```

## Testing Conventions

- Table-driven tests with `t.Run()`
- Use `testify/require` for fatal assertions, `testify/assert` for non-fatal
- Domain tests cover: entity creation validation, state machine transitions, business rules
- Handler tests use `httptest.NewServer` with mock service interfaces
- Test file naming: `{source}_test.go` in the same package

## CI Pipeline

GitHub Actions runs on every PR:
- **Lint** → golangci-lint (staticcheck, errcheck, govet, bodyclose, misspell, gocritic)
- **Test** → `go test -count=1` with Postgres, Redis, NATS services
- **Build** → `go build` all services (depends on Lint passing)
- **Security** → gosec, govulncheck, Trivy container scan, secret detection

## Git Workflow

- Feature branches from `develop`: `feat/{feature-name}`
- PRs target `develop`, never `main` directly
- `main` receives merge PRs from `develop` for releases
- Commit style: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`

## Gateway Routing

All external traffic goes through the gateway on :8080. Routes are defined in `services/gateway/internal/handler/gateway.go`. Three auth patterns:

1. **Public routes** — no middleware (auth endpoints, public payments, checkout)
2. **JWT routes** — `auth.JWTMiddleware` (merchant dashboard operations)
3. **HMAC SDK routes** — `middleware.HMACAuth` under `/v1/sdk/*` (API key authenticated)
4. **Admin routes** — JWT + `RequirePlatformAdmin()` under `/v1/admin/*`, path rewritten before proxying

## Adding a New Feature

1. Create migration: `make migrate-create svc={name} name={description}`
2. Add domain entity with validation, state machine, and tests
3. Add repository interface in service package + postgres implementation
4. Add service methods orchestrating domain logic
5. Add HTTP handler with chi router and JWT middleware
6. Add gateway proxy routes in `services/gateway/internal/handler/gateway.go`
7. Add frontend hooks (`use-{feature}.ts`) and page component
8. Update `App.tsx` routes and `app-sidebar.tsx` navigation
9. If new microservice: update `docker-compose.yml`, `Makefile`, `.env.example`, gateway proxy config

## Environment Variables

See `.env.example` for all configurable values. Key patterns:
- `{SERVICE}_DATABASE_URL` — per-service Postgres connection
- `{SERVICE}_PORT` — service HTTP port
- `{SERVICE}_SERVICE_URL` — used by gateway for proxying
- `JWT_SECRET` — shared JWT signing key (min 32 chars)
