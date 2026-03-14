# OpenLankaPayment

> A production-grade, open-source cryptocurrency payment processing platform for Sri Lanka.

[![CI](https://github.com/theetaz/open-pay/actions/workflows/ci.yml/badge.svg)](https://github.com/theetaz/open-pay/actions/workflows/ci.yml)
[![Security](https://github.com/theetaz/open-pay/actions/workflows/security.yml/badge.svg)](https://github.com/theetaz/open-pay/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/openlankapay/openlankapay)](https://goreportcard.com/report/github.com/openlankapay/openlankapay)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Overview

OpenLankaPayment enables merchants to accept cryptocurrency payments and receive settlement in Sri Lankan Rupees (LKR). Built as a microservices architecture with enterprise-grade security, observability, and fault tolerance.

### Key Features

- **50+ Cryptocurrencies** — Accept BTC, ETH, USDT, USDC, and more
- **LKR Settlement** — Merchants receive Sri Lankan Rupees in their bank accounts
- **Zero Chargebacks** — Crypto transactions are irreversible (>99.9% finality)
- **Multi-Provider** — Abstracted payment provider layer (Bybit, Binance Pay, KuCoin)
- **Subscription Payments** — Recurring crypto billing (off-chain + on-chain via smart contracts)
- **Real-Time Updates** — WebSocket-based payment status streaming
- **Multi-Tenant** — Branch system with role-based access control
- **Plugin Ecosystem** — WooCommerce plugin, with SDKs for Go, TypeScript, Python, PHP

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                            │
│            (HMAC-SHA256 Auth · Rate Limiting)               │
└──────────┬──────────┬──────────┬──────────┬────────────────┘
           │          │          │          │
    ┌──────▼───┐ ┌────▼────┐ ┌──▼────┐ ┌───▼──────┐
    │ Payment  │ │Merchant │ │Settle-│ │Subscrip- │
    │ Service  │ │ Service │ │ ment  │ │  tion    │  ...6 more
    └──────────┘ └─────────┘ └───────┘ └──────────┘
           │          │          │          │
    ┌──────▼──────────▼──────────▼──────────▼────────────────┐
    │              NATS JetStream (Event Bus)                 │
    └────────────────────────────────────────────────────────┘
    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐
    │PostgreSQL│  │  Redis   │  │ Hardhat  │  │ Prometheus │
    │ (per-svc)│  │ (cache)  │  │(testnet) │  │ + Grafana  │
    └──────────┘  └──────────┘  └──────────┘  └────────────┘
```

### Services

| Service | Port | Responsibility |
|---------|------|----------------|
| **Gateway** | 8080 | Public REST API, authentication, rate limiting |
| **Payment** | 8081 | Payment creation, status tracking, QR generation |
| **Merchant** | 8082 | Registration, KYC, API keys, branches, users |
| **Settlement** | 8083 | Balance tracking, withdrawals, treasury |
| **Webhook** | 8084 | ED25519-signed delivery with exponential backoff |
| **Exchange** | 8085 | Real-time exchange rates (USDT/LKR) |
| **Subscription** | 8086 | Recurring payments, billing cycles, smart contracts |
| **Notification** | 8087 | Email, SMS, push notifications |
| **Admin** | 8088 | Merchant approval, audit logs, system health |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.25 · chi · ConnectRPC · sqlc · pgx/v5 |
| **Database** | PostgreSQL 16 (database-per-service) |
| **Messaging** | NATS JetStream |
| **Cache** | Redis 7 |
| **Frontend** | TanStack Start · shadcn/ui · Tailwind CSS v4 |
| **Blockchain** | Solidity · Hardhat · go-ethereum |
| **Observability** | zerolog · OpenTelemetry · Prometheus · Grafana |
| **Testing** | testify · testcontainers-go · gomock |
| **Infrastructure** | Docker · Docker Compose |

---

## Quick Start

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) (for database migrations)
- [Node.js 20+](https://nodejs.org/) & [pnpm](https://pnpm.io/) (for frontend)

### Setup

```bash
# Clone the repository
git clone https://github.com/nipuntheekshana/open-lanka-payment.git
cd open-lanka-payment

# Start infrastructure (PostgreSQL, Redis, NATS)
make dev

# Run database migrations
make migrate

# Run all tests
make test

# Build all services
make build
```

### Available Commands

```bash
make help              # Show all available commands
make dev               # Start infrastructure containers
make down              # Stop all containers
make test              # Run unit tests
make test-v            # Run unit tests (verbose)
make test-integration  # Run integration tests (requires Docker)
make test-coverage     # Generate HTML coverage report
make lint              # Run golangci-lint
make fmt               # Format Go code
make build             # Build all services
make migrate-create    # Create new migration (svc=payment name=create_table)
```

---

## Security

Security is a first-class concern in OpenLankaPayment.

### Authentication

| Layer | Mechanism | Purpose |
|-------|-----------|---------|
| Merchant API | HMAC-SHA256 | Request signing with derived keys |
| Webhook Delivery | ED25519 | Asymmetric payload signing |
| Dashboard | JWT | Access + refresh token rotation |
| Admin API | Admin Secret + JWT | Internal operations |

### Security Measures

- **No floating-point arithmetic** for monetary values (uses `shopspring/decimal`)
- **Parameterized queries** via sqlc (compile-time SQL safety)
- **API secrets** bcrypt-hashed at rest, shown once at creation
- **ED25519 private keys** encrypted at rest
- **Timestamp validation** (5-minute window) prevents replay attacks
- **Rate limiting** per-merchant per-endpoint (sliding window)
- **RBAC** with branch-level scoping (Admin, Manager, User)
- **Soft-delete** architecture preserves audit trails
- **PII isolation** — customer data stays on merchant servers
- **Automated security scanning** via GitHub Actions (gosec, govulncheck, trivy)

### Reporting Vulnerabilities

If you discover a security vulnerability, please report it responsibly via email. Do **not** open a public issue.

---

## Project Structure

```
open-lanka-payment/
├── services/           # Go microservices (9 services)
│   ├── gateway/        #   API Gateway
│   ├── payment/        #   Payment processing
│   ├── merchant/       #   Merchant management
│   ├── settlement/     #   Settlement engine
│   ├── webhook/        #   Webhook delivery
│   ├── exchange/       #   Exchange rates
│   ├── subscription/   #   Recurring payments
│   ├── notification/   #   Notifications
│   └── admin/          #   Admin operations
├── pkg/                # Shared Go packages
│   ├── auth/           #   HMAC-SHA256 + ED25519
│   ├── money/          #   Currency + fee calculation
│   ├── database/       #   PostgreSQL connection pool
│   └── observability/  #   Structured logging
├── migrations/         # SQL migrations (per service)
├── apps/               # Frontend applications
│   ├── merchant-portal/
│   └── admin-dashboard/
├── contracts/          # Solidity smart contracts
├── plugins/            # E-commerce plugins
├── sdks/               # Client SDKs
├── docs/               # OpenAPI specifications
├── config/             # Infrastructure config
└── scripts/            # Utility scripts
```

---

## Development Workflow

This project follows a strict development workflow:

1. **Feature branches** — `feature/<phase-description>` from `main`
2. **TDD** — Every feature starts with a failing test
3. **Conventional commits** — `feat:`, `fix:`, `test:`, `chore:`, `refactor:`, etc.
4. **Security audit** — Automated scanning before merge
5. **Pull requests** — All changes reviewed and merged via PR

### Branch Naming

```
feature/merchant-service
feature/api-gateway
feature/payment-service
feature/webhook-system
feature/settlement-engine
```

---

## Implementation Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 0 | Project Foundation | Done |
| 1 | Merchant Service | Planned |
| 2 | API Gateway | Planned |
| 3 | Payment Service | Planned |
| 4 | Exchange Rate Service | Planned |
| 5 | Webhook Service | Planned |
| 6 | Settlement Service | Planned |
| 7 | Merchant Portal (Frontend) | Planned |
| 8 | Admin Dashboard | Planned |
| 9 | Checkout Experience | Planned |
| 10 | Subscription Service | Planned |
| 11 | Smart Contracts | Planned |
| 12 | Notification Service | Planned |
| 13 | WooCommerce Plugin | Planned |
| 14 | Client SDKs | Planned |
| 15 | Hardening & Polish | Planned |

---

## License

This project is licensed under the [MIT License](LICENSE).
