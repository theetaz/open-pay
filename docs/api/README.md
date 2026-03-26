# API Reference

Open Pay provides a REST API for crypto-to-fiat payment processing. All SDK endpoints use HMAC-SHA256 authentication.

## Base URLs

| Environment | URL |
|-------------|-----|
| Production  | `https://olp-api.nipuntheekshana.com` |
| Sandbox     | `http://localhost:8080` |

## Authentication

### HMAC-SHA256 (SDK/API Key)

All `/v1/sdk/*` endpoints require HMAC authentication via three headers:

| Header | Description |
|--------|-------------|
| `x-api-key` | Your API key ID (e.g., `ak_live_xxx`) |
| `x-timestamp` | Current Unix timestamp in milliseconds |
| `x-signature` | HMAC-SHA256 signature |

**Signature computation:**
```
signing_key = SHA256(secret)
message = timestamp + METHOD + path + body
signature = HMAC-SHA256(signing_key, message)
```

The timestamp must be within 5 minutes of the server time.

### JWT (Portal)

Portal endpoints (`/v1/auth/*`, `/v1/merchants/*`) use Bearer JWT tokens from login.

---

## Endpoints

### Checkout Sessions

#### Create Checkout Session

`POST /v1/sdk/checkout/sessions`

Creates a payment and returns a hosted checkout URL.

**Request:**
```json
{
  "amount": "2500.00",
  "currency": "LKR",
  "successUrl": "https://mysite.com/success",
  "cancelUrl": "https://mysite.com/cancel",
  "customerEmail": "buyer@example.com",
  "merchantTradeNo": "ORDER-123",
  "expiresInMinutes": 30,
  "lineItems": [
    { "name": "Premium Widget", "description": "Blue, Large" }
  ]
}
```

**Response (201):**
```json
{
  "data": {
    "id": "cs_a99414e4-0402-4837-8777-93f3f22cb5e8",
    "paymentId": "a99414e4-0402-4837-8777-93f3f22cb5e8",
    "url": "https://checkout.provider.com/pay/...",
    "amount": "2500",
    "currency": "LKR",
    "amountUsdt": "7.95",
    "status": "open",
    "qrContent": "crypto-qr://...",
    "deepLink": "wallet-app://pay/...",
    "exchangeRate": "314.58",
    "expiresAt": "2026-03-25T22:00:00Z",
    "createdAt": "2026-03-25T21:30:00Z"
  }
}
```

---

### Payments

#### Create Payment

`POST /v1/sdk/payments`

**Request:**
```json
{
  "amount": "100.00",
  "currency": "LKR",
  "merchantTradeNo": "ORDER-456",
  "description": "Monthly subscription",
  "webhookURL": "https://mysite.com/webhook",
  "customerEmail": "user@example.com"
}
```

**Response (201):**
```json
{
  "data": {
    "id": "uuid",
    "merchantId": "uuid",
    "amount": "100",
    "currency": "LKR",
    "amountUsdt": "0.32",
    "status": "INITIATED",
    "provider": "TEST",
    "qrContent": "crypto-qr://...",
    "checkoutLink": "https://...",
    "deepLink": "app://...",
    "exchangeRate": "314.58",
    "createdAt": "2026-03-25T21:30:00Z"
  }
}
```

#### List Payments

`GET /v1/sdk/payments?page=1&perPage=20&status=PAID`

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `perPage` | int | Items per page (default: 20, max: 100) |
| `status` | string | Filter by status: INITIATED, PENDING, PAID, CONFIRMED, EXPIRED, FAILED |
| `search` | string | Search by merchant trade no |

#### Get Payment

`GET /v1/sdk/payments/{id}`

Returns the full payment object.

---

### Webhooks

#### Configure Webhook

`POST /v1/sdk/webhooks/configure`

```json
{
  "url": "https://mysite.com/webhook",
  "events": ["payment.*"]
}
```

#### Get Public Key

`GET /v1/sdk/webhooks/public-key`

Returns the ED25519 public key (base64) for verifying webhook signatures.

#### Test Webhook

`POST /v1/sdk/webhooks/test`

Sends a test event to your configured webhook endpoint.

#### List Deliveries

`GET /v1/webhooks/deliveries?page=1&perPage=20`

Returns webhook delivery history with status, attempts, and errors.

---

### API Keys

These endpoints use JWT authentication (from the merchant portal).

#### Create API Key

`POST /v1/api-keys`

```json
{
  "name": "Production Key",
  "environment": "live"
}
```

**Response (201):**
```json
{
  "data": {
    "id": "uuid",
    "keyId": "ak_live_xxx",
    "secret": "ak_live_xxx.sk_live_yyy",
    "name": "Production Key",
    "environment": "live"
  }
}
```

> The `secret` field containing the full compound key is only returned at creation time.

#### List API Keys

`GET /v1/api-keys`

#### Revoke API Key

`DELETE /v1/api-keys/{id}`

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "amount is required"
  }
}
```

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `NOT_FOUND` | 404 | Resource not found |
| `INTERNAL_ERROR` | 500 | Server error |
| `ACCOUNT_TERMINATED` | 403 | Merchant account terminated |
| `ACCOUNT_FROZEN` | 403 | Merchant account frozen |

---

## Payment Statuses

| Status | Description |
|--------|-------------|
| `INITIATED` | Payment created, awaiting customer action |
| `PENDING` | Customer submitted payment, awaiting confirmation |
| `PAID` | Payment confirmed on blockchain |
| `CONFIRMED` | Payment settled to merchant |
| `EXPIRED` | Payment expired (customer didn't pay in time) |
| `FAILED` | Payment failed |
| `REFUNDED` | Payment refunded |

## Webhook Events

| Event | Description |
|-------|-------------|
| `payment.initiated` | Payment created |
| `payment.paid` | Payment confirmed |
| `payment.expired` | Payment expired |
| `payment.failed` | Payment failed |
| `test.webhook` | Test webhook delivery |
