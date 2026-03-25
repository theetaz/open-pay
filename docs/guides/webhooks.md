# Webhook Integration Guide

Webhooks notify your application in real-time when payment events occur (paid, expired, failed). Open Pay signs all webhook payloads with ED25519 for security.

## Setup

### 1. Configure Your Endpoint

```typescript
await openpay.webhooks.configure({
  url: 'https://mysite.com/api/openpay-webhook',
  events: ['payment.*'],  // optional: filter specific events
})
```

Or via CLI:
```bash
openpay webhooks configure --url https://mysite.com/api/openpay-webhook
```

### 2. Get Your Public Key

```typescript
const publicKey = await openpay.webhooks.getPublicKey()
// Store this — you'll use it to verify webhook signatures
```

### 3. Handle Incoming Webhooks

Your endpoint will receive POST requests with these headers:

| Header | Description |
|--------|-------------|
| `X-Webhook-Signature` | ED25519 signature (base64) |
| `X-Webhook-Timestamp` | Unix milliseconds when signed |
| `X-Webhook-Event` | Event type (e.g., `payment.paid`) |
| `X-Webhook-ID` | Unique delivery ID |
| `X-Webhook-Attempt` | Delivery attempt number |

### 4. Verify and Process

#### Express.js (TypeScript)
```typescript
import express from 'express'
import { verifyWebhookSignature } from '@openpay/sdk'

const app = express()
app.use(express.raw({ type: 'application/json' }))

const PUBLIC_KEY = 'your-ed25519-public-key-base64'

app.post('/api/openpay-webhook', (req, res) => {
  try {
    const event = verifyWebhookSignature(
      req.body.toString(),
      {
        'x-webhook-signature': req.headers['x-webhook-signature'],
        'x-webhook-timestamp': req.headers['x-webhook-timestamp'],
        'x-webhook-event': req.headers['x-webhook-event'],
        'x-webhook-id': req.headers['x-webhook-id'],
      },
      PUBLIC_KEY,
    )

    switch (event.event) {
      case 'payment.paid':
        const paymentId = event.data.paymentId
        // Mark order as paid in your database
        console.log(`Payment ${paymentId} confirmed!`)
        break

      case 'payment.expired':
        // Release reserved inventory
        break

      case 'payment.failed':
        // Notify customer
        break
    }

    res.sendStatus(200) // Always return 200 to acknowledge
  } catch (err) {
    console.error('Webhook verification failed:', err)
    res.sendStatus(400)
  }
})
```

#### Go
```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    err := openpay.VerifyWebhookSignature(
        publicKeyBase64,
        r.Header.Get("X-Webhook-Timestamp"),
        body,
        r.Header.Get("X-Webhook-Signature"),
    )
    if err != nil {
        http.Error(w, "invalid signature", http.StatusBadRequest)
        return
    }

    var event map[string]interface{}
    json.Unmarshal(body, &event)
    // Process event...

    w.WriteHeader(http.StatusOK)
}
```

#### Python (Flask)
```python
from flask import Flask, request
from openpay import verify_webhook_signature

app = Flask(__name__)
PUBLIC_KEY = "your-ed25519-public-key-base64"

@app.post("/webhook")
def handle_webhook():
    try:
        event = verify_webhook_signature(
            payload=request.data.decode(),
            signature_b64=request.headers["X-Webhook-Signature"],
            timestamp=request.headers["X-Webhook-Timestamp"],
            event=request.headers["X-Webhook-Event"],
            webhook_id=request.headers["X-Webhook-ID"],
            public_key_b64=PUBLIC_KEY,
        )

        if event.event == "payment.paid":
            # Fulfill order
            pass

        return "", 200
    except Exception as e:
        return str(e), 400
```

## Retry Policy

Failed deliveries are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1st retry | 1 minute |
| 2nd retry | 2 minutes |
| 3rd retry | 4 minutes |
| 4th retry | 8 minutes |
| 5th (final) | Marked as exhausted |

The retry worker runs every 30 seconds and picks up pending deliveries.

## Viewing Delivery Logs

Check delivery status and debug failed webhooks:

```typescript
// Via API (JWT auth)
const deliveries = await fetch('/v1/webhooks/deliveries?page=1&perPage=10', {
  headers: { Authorization: `Bearer ${token}` },
})
```

Each delivery shows: event type, status, attempt count, response code, error message, and timestamps.

## Testing

Send a test webhook to verify your endpoint:

```bash
openpay webhooks test
```

This sends a `test.webhook` event to your configured URL.

## Best Practices

1. **Always verify signatures** — never process unverified webhooks
2. **Return 200 quickly** — do heavy processing async (queue the event)
3. **Handle duplicates** — use `X-Webhook-ID` for idempotency
4. **Use HTTPS** — webhook URLs must be HTTPS in production
5. **Log failures** — check delivery logs if webhooks aren't arriving
