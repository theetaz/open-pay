import { createVerify } from 'node:crypto'
import { OpenPayError } from './errors.js'
import type { WebhookEvent } from './types.js'

/**
 * Verify an incoming webhook signature using ED25519.
 * The Open Pay webhook service signs payloads with ED25519.
 *
 * @param payload - The raw request body string
 * @param headers - The request headers containing signature, timestamp, event, and ID
 * @param publicKey - The ED25519 public key (base64-encoded, from GET /v1/webhooks/public-key)
 * @returns The parsed webhook event
 */
export function verifyWebhookSignature(
  payload: string,
  headers: {
    'x-webhook-signature'?: string
    'x-webhook-timestamp'?: string
    'x-webhook-event'?: string
    'x-webhook-id'?: string
  },
  publicKey: string,
): WebhookEvent {
  const signature = headers['x-webhook-signature']
  const timestamp = headers['x-webhook-timestamp']
  const event = headers['x-webhook-event']
  const id = headers['x-webhook-id']

  if (!signature || !timestamp || !event || !id) {
    throw new OpenPayError('Missing webhook signature headers')
  }

  // Verify ED25519 signature: message = timestamp_bytes + payload_bytes
  const message = Buffer.concat([
    Buffer.from(timestamp, 'utf8'),
    Buffer.from(payload, 'utf8'),
  ])

  const sigBytes = Buffer.from(signature, 'base64')
  const pubKeyBytes = Buffer.from(publicKey, 'base64')

  const verify = createVerify('ed25519')
  verify.update(message)

  // ed25519 keys need to be in DER format for Node.js crypto
  const derPrefix = Buffer.from('302a300506032b6570032100', 'hex')
  const derKey = Buffer.concat([derPrefix, pubKeyBytes])

  const isValid = verify.verify(
    { key: derKey, format: 'der', type: 'spki' },
    sigBytes,
  )

  if (!isValid) {
    throw new OpenPayError('Invalid webhook signature')
  }

  // Check timestamp freshness (5 minutes)
  const ts = parseInt(timestamp, 10)
  const now = Date.now()
  if (Math.abs(now - ts) > 5 * 60 * 1000) {
    throw new OpenPayError('Webhook timestamp too old')
  }

  return {
    id,
    event,
    timestamp,
    data: JSON.parse(payload),
  }
}
