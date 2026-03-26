import { createHmac, createHash } from 'node:crypto'
import { AuthenticationError } from './errors.js'

/**
 * Parse a compound API key into key ID and secret.
 * Format: "ak_{env}_{id}.sk_{env}_{secret}"
 */
export function parseAPIKey(apiKey: string): { keyId: string; secret: string } {
  if (!apiKey) throw new AuthenticationError('API key is required')

  const parts = apiKey.split('.')
  if (parts.length !== 2) throw new AuthenticationError('Invalid API key format')

  const [keyId, secret] = parts
  if (!keyId.startsWith('ak_live_') && !keyId.startsWith('ak_test_'))
    throw new AuthenticationError('Invalid API key prefix')
  if (!secret.startsWith('sk_live_') && !secret.startsWith('sk_test_'))
    throw new AuthenticationError('Invalid API secret prefix')

  return { keyId, secret }
}

/**
 * Sign an API request using HMAC-SHA256.
 * Matches the Go implementation: signing key = SHA256(secret), message = timestamp + METHOD + path + body.
 */
export function signRequest(
  secret: string,
  timestamp: string,
  method: string,
  path: string,
  body: string,
): string {
  const signingKey = createHash('sha256').update(secret).digest()
  const message = timestamp + method.toUpperCase() + path + body
  return createHmac('sha256', signingKey).update(message).digest('hex')
}

/**
 * Get current timestamp as Unix milliseconds string.
 */
export function currentTimestamp(): string {
  return Date.now().toString()
}

/**
 * Build authentication headers for an API request.
 */
export function buildAuthHeaders(
  keyId: string,
  secret: string,
  method: string,
  path: string,
  body: string,
): Record<string, string> {
  const timestamp = currentTimestamp()
  const signature = signRequest(secret, timestamp, method, path, body)

  return {
    'x-api-key': keyId,
    'x-timestamp': timestamp,
    'x-signature': signature,
  }
}
