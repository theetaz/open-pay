# Load Tests

k6 load test scripts for benchmarking Open Pay API endpoints.

## Prerequisites

Install k6: https://k6.io/docs/get-started/installation/

## Usage

```bash
# Run all load tests
make load-test

# Run individual tests
k6 run tests/load/exchange-rates.js
k6 run tests/load/create-payment.js --env JWT_TOKEN=<your-token>
k6 run tests/load/list-payments.js --env JWT_TOKEN=<your-token>
k6 run tests/load/checkout-flow.js --env JWT_TOKEN=<your-token>
k6 run tests/load/auth-flow.js --env TEST_EMAIL=<email> --env TEST_PASSWORD=<pass>
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_URL` | Base API URL | `http://localhost:8080` |
| `JWT_TOKEN` | Auth token for protected endpoints | (empty) |
| `TEST_EMAIL` | Email for auth flow test | `loadtest@example.com` |
| `TEST_PASSWORD` | Password for auth flow test | `LoadTest123!` |

## Test Scenarios

| Script | Target | Threshold |
|--------|--------|-----------|
| `exchange-rates.js` | 500 RPS | p95 < 50ms |
| `list-payments.js` | 200 RPS | p95 < 100ms |
| `create-payment.js` | 100 RPS | p95 < 200ms |
| `checkout-flow.js` | 50 concurrent | p95 < 500ms |
| `auth-flow.js` | 50 concurrent | p95 < 300ms |
