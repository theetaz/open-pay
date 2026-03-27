const express = require('express')
const crypto = require('crypto')
const { OpenPayClient } = require('./openpay-client')
const { verifyWebhookSignature } = require('./verify-signature')

const app = express()
app.use(express.json({ verify: (req, _res, buf) => { req.rawBody = buf.toString() } }))

const PORT = process.env.PORT || 3100
const OPENPAY_API_URL = process.env.OPENPAY_API_URL || 'http://localhost:8080'
const OPENPAY_API_KEY = process.env.OPENPAY_API_KEY || ''
const OPENPAY_API_SECRET = process.env.OPENPAY_API_SECRET || ''
const OPENPAY_PROVIDER = process.env.OPENPAY_PROVIDER || 'TEST'
const APP_URL = process.env.APP_URL || `http://localhost:${PORT}`

const client = new OpenPayClient(OPENPAY_API_URL, OPENPAY_API_KEY, OPENPAY_API_SECRET)

/**
 * POST /payment/create
 * Called by Shopify when customer selects Open Pay at checkout.
 * Creates a payment and returns the checkout redirect URL.
 */
app.post('/payment/create', async (req, res) => {
  try {
    const { order_id, amount, currency, customer_email, return_url, cancel_url } = req.body

    const payment = await client.createPayment({
      amount: String(amount),
      currency: currency || 'USDT',
      provider: OPENPAY_PROVIDER,
      merchantTradeNo: `SHOPIFY-${order_id}`,
      customerEmail: customer_email || '',
      webhookUrl: `${APP_URL}/webhook/payment`,
      successUrl: return_url || `${APP_URL}/payment/success`,
      cancelUrl: cancel_url || `${APP_URL}/payment/cancel`,
    })

    res.json({
      success: true,
      payment_id: payment.id,
      checkout_url: payment.checkoutLink || `${OPENPAY_API_URL.replace('/api', '')}/checkout/${payment.id}`,
      qr_content: payment.qrContent,
    })
  } catch (err) {
    console.error('Payment creation failed:', err.message)
    res.status(500).json({ success: false, error: err.message })
  }
})

/**
 * GET /payment/status/:id
 * Returns current payment status for order status polling.
 */
app.get('/payment/status/:id', async (req, res) => {
  try {
    const payment = await client.getPayment(req.params.id)
    res.json({
      success: true,
      status: payment.status,
      payment_no: payment.paymentNo,
      paid_at: payment.paidAt || null,
    })
  } catch (err) {
    res.status(500).json({ success: false, error: err.message })
  }
})

/**
 * POST /webhook/payment
 * Receives payment status updates from Open Pay.
 * Verifies ED25519 signature and maps to Shopify order status.
 */
app.post('/webhook/payment', async (req, res) => {
  const signature = req.headers['x-signature'] || ''
  const event = req.headers['x-webhook-event'] || ''

  // Verify signature
  if (signature && OPENPAY_API_URL) {
    const isValid = await verifyWebhookSignature(req.rawBody, signature, OPENPAY_API_URL)
    if (!isValid) {
      console.warn('Webhook signature verification failed')
      return res.status(401).json({ error: 'Invalid signature' })
    }
  }

  const data = req.body
  const orderId = (data.merchantTradeNo || '').replace('SHOPIFY-', '')

  console.log(`Webhook received: ${event} for order ${orderId}, status: ${data.status}`)

  switch (event) {
    case 'payment.paid':
      // In production: call Shopify Admin API to mark order as paid
      console.log(`Order ${orderId} marked as PAID. Tx: ${data.txHash || 'N/A'}`)
      break
    case 'payment.expired':
      console.log(`Order ${orderId} payment EXPIRED`)
      break
    case 'payment.failed':
      console.log(`Order ${orderId} payment FAILED`)
      break
    default:
      console.log(`Unknown event: ${event}`)
  }

  res.json({ received: true })
})

/**
 * GET /health
 */
app.get('/health', (_req, res) => {
  res.json({ status: 'ok', provider: OPENPAY_PROVIDER })
})

/**
 * GET /config
 * Returns current plugin configuration (non-sensitive).
 */
app.get('/config', (_req, res) => {
  res.json({
    apiUrl: OPENPAY_API_URL,
    provider: OPENPAY_PROVIDER,
    hasApiKey: !!OPENPAY_API_KEY,
    appUrl: APP_URL,
  })
})

app.listen(PORT, () => {
  console.log(`Open Pay Shopify plugin running on port ${PORT}`)
  console.log(`  API URL: ${OPENPAY_API_URL}`)
  console.log(`  Provider: ${OPENPAY_PROVIDER}`)
})
