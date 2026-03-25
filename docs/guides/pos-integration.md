# POS System Integration Guide

This guide shows how to integrate Open Pay into an existing Point of Sale (POS) system to accept crypto payments that settle in LKR.

## Flow Overview

```
Customer checkout → POS creates payment via SDK → Customer scans QR/opens link
→ Customer pays in crypto → Webhook notifies POS → Order fulfilled
→ Settlement in LKR to merchant bank account
```

## Step 1: Setup

Install the SDK for your POS backend language:

```bash
# Node.js POS
npm install @openpay/sdk

# Python POS
pip install openpay-sdk

# PHP POS (WordPress/WooCommerce)
composer require openpay/sdk

# Java POS (Android)
# Add to pom.xml: com.openpay:openpay-sdk:0.1.0
```

## Step 2: Initialize Client

```typescript
// Initialize once at startup
import { OpenPay } from '@openpay/sdk'

const openpay = new OpenPay('ak_live_xxx.sk_live_yyy', {
  baseURL: 'https://olp-api.nipuntheekshana.com',
})
```

## Step 3: Create Payment at Checkout

When a customer is ready to pay:

```typescript
async function handleCheckout(orderId: string, totalLKR: number) {
  // Create a checkout session
  const session = await openpay.checkout.createSession({
    amount: totalLKR.toFixed(2),
    currency: 'LKR',
    merchantTradeNo: orderId,
    successUrl: `https://mypos.com/orders/${orderId}/success`,
    cancelUrl: `https://mypos.com/orders/${orderId}`,
    customerEmail: 'customer@example.com',
  })

  return {
    checkoutUrl: session.url,      // Redirect customer here
    qrCode: session.qrContent,     // Display as QR on POS screen
    paymentId: session.paymentId,  // Track this payment
    expiresAt: session.expiresAt,  // Payment expires after this
  }
}
```

### Display Options

| Method | Use Case |
|--------|----------|
| **QR Code** | Display on POS screen, customer scans with crypto wallet |
| **Checkout URL** | Redirect customer to hosted payment page |
| **Deep Link** | Open customer's wallet app directly (mobile POS) |

## Step 4: Handle Webhook (Payment Confirmation)

Set up a webhook endpoint that Open Pay calls when payment status changes:

```typescript
app.post('/api/openpay-webhook', (req, res) => {
  const event = verifyWebhookSignature(req.body, req.headers, PUBLIC_KEY)

  if (event.event === 'payment.paid') {
    const orderId = event.data.merchantTradeNo
    markOrderAsPaid(orderId)
    printReceipt(orderId)  // POS-specific action
  }

  res.sendStatus(200)
})
```

## Step 5: Check Payment Status (Polling)

For POS systems that can't receive webhooks, poll the payment status:

```typescript
async function waitForPayment(paymentId: string, timeoutMs = 300000) {
  const start = Date.now()

  while (Date.now() - start < timeoutMs) {
    const payment = await openpay.payments.get(paymentId)

    if (payment.status === 'PAID' || payment.status === 'CONFIRMED') {
      return { success: true, payment }
    }
    if (payment.status === 'EXPIRED' || payment.status === 'FAILED') {
      return { success: false, payment }
    }

    await new Promise(r => setTimeout(r, 3000)) // Poll every 3s
  }

  return { success: false, reason: 'timeout' }
}
```

## Example: Express.js POS Backend

```typescript
import express from 'express'
import { OpenPay, verifyWebhookSignature } from '@openpay/sdk'

const app = express()
const openpay = new OpenPay('ak_live_xxx.sk_live_yyy', {
  baseURL: 'https://olp-api.nipuntheekshana.com',
})

// Create payment when customer checks out
app.post('/api/checkout', async (req, res) => {
  const { orderId, amount } = req.body

  const session = await openpay.checkout.createSession({
    amount: amount.toString(),
    currency: 'LKR',
    merchantTradeNo: orderId,
    successUrl: `https://mypos.com/orders/${orderId}/success`,
    cancelUrl: `https://mypos.com/orders/${orderId}`,
  })

  res.json({
    checkoutUrl: session.url,
    qrCode: session.qrContent,
    paymentId: session.paymentId,
  })
})

// Webhook: payment confirmed
app.post('/api/webhook', express.raw({ type: '*/*' }), (req, res) => {
  const event = verifyWebhookSignature(
    req.body.toString(),
    req.headers,
    process.env.OPENPAY_PUBLIC_KEY,
  )

  if (event.event === 'payment.paid') {
    fulfillOrder(event.data.merchantTradeNo)
  }

  res.sendStatus(200)
})

app.listen(3000)
```

## Fee Structure

| Fee | Rate | Description |
|-----|------|-------------|
| Exchange fee | 0.5% | Currency conversion (crypto → LKR) |
| Platform fee | 1.5% | Open Pay processing fee |
| **Total** | **2.0%** | Deducted from settlement amount |

Settlement is made to the merchant's registered bank account in LKR.
