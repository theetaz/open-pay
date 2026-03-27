# @openpay/react-native

Official Open Pay SDK for React Native. Accept crypto payments with automatic LKR conversion in your mobile apps.

## Installation

```bash
npm install @openpay/react-native
# or
yarn add @openpay/react-native
```

## Requirements

- React Native >= 0.70.0
- Hermes engine (recommended, provides Web Crypto API)

## Quick Start

```typescript
import { OpenPay } from '@openpay/react-native'

const openpay = new OpenPay('ak_live_xxx.sk_live_yyy')
```

## Create a Payment

```typescript
const payment = await openpay.payments.create({
  amount: '1000.00',
  currency: 'LKR',
  merchantTradeNo: 'ORDER-123',
  description: 'Premium subscription',
})

console.log(payment.id)
console.log(payment.checkoutLink) // Redirect user here
console.log(payment.deepLink)     // Open wallet app directly
```

## Get Payment Status

```typescript
const payment = await openpay.payments.get('pay_abc123')
console.log(payment.status) // 'INITIATED' | 'PENDING' | 'PAID' | 'CONFIRMED' | ...
```

## List Payments

```typescript
const result = await openpay.payments.list({
  page: 1,
  perPage: 20,
  status: 'PAID',
})

result.data.forEach((payment) => {
  console.log(payment.id, payment.amount, payment.status)
})
```

## Checkout Sessions

Create a hosted checkout session and redirect the user:

```typescript
import { Linking } from 'react-native'

const session = await openpay.checkout.createSession({
  amount: '2500.00',
  currency: 'LKR',
  successUrl: 'myapp://payment/success',
  cancelUrl: 'myapp://payment/cancel',
})

// Open checkout in browser
await Linking.openURL(session.url)
```

## Webhook Verification

Verify incoming webhook signatures in your backend or React Native server:

```typescript
import { verifyWebhookSignature } from '@openpay/react-native'

const event = verifyWebhookSignature(
  rawBody,
  {
    'x-webhook-signature': signature,
    'x-webhook-timestamp': timestamp,
    'x-webhook-event': eventType,
    'x-webhook-id': webhookId,
  },
  publicKey,
)

console.log(event.event) // 'payment.confirmed'
console.log(event.data)  // Payment data
```

## Configuration

```typescript
const openpay = new OpenPay('ak_test_xxx.sk_test_yyy', {
  baseURL: 'https://api.openpay.lk', // Custom API URL
  timeout: 15000,                     // Request timeout in ms
})
```

## API Key Format

API keys follow the format `ak_{env}_{id}.sk_{env}_{secret}` where `env` is either `live` or `test`.

## Error Handling

```typescript
import { OpenPay, APIError, AuthenticationError, OpenPayError } from '@openpay/react-native'

try {
  const payment = await openpay.payments.create({ amount: '1000.00' })
} catch (err) {
  if (err instanceof AuthenticationError) {
    console.error('Bad API key:', err.message)
  } else if (err instanceof APIError) {
    console.error(`API error [${err.code}]: ${err.message} (HTTP ${err.status})`)
  } else if (err instanceof OpenPayError) {
    console.error('SDK error:', err.message)
  }
}
```

## License

MIT
