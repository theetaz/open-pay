import type {
  CreatePaymentInput,
  Payment,
  ListPaymentsParams,
  PaginatedResponse,
  CheckoutSessionInput,
  CheckoutSession,
  WebhookEvent,
  WebhookHeaders,
  ClientOptions,
  APIErrorResponse,
} from './types'

export * from './types'

const DEFAULT_BASE_URL = 'https://api.openpay.lk'

// ─── Errors ───

export class OpenPayError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'OpenPayError'
  }
}

export class AuthenticationError extends OpenPayError {
  constructor(message = 'Authentication failed') {
    super(message)
    this.name = 'AuthenticationError'
  }
}

export class APIError extends OpenPayError {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'APIError'
    this.code = code
    this.status = status
  }
}

// ─── Crypto Utilities (uses React Native compatible global crypto) ───

/**
 * Convert a string to a Uint8Array.
 */
function stringToBytes(str: string): Uint8Array {
  const encoder = new TextEncoder()
  return encoder.encode(str)
}

/**
 * Convert a Uint8Array to a hex string.
 */
function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

/**
 * SHA-256 hash using Web Crypto API (available in React Native Hermes).
 */
async function sha256(data: Uint8Array): Promise<ArrayBuffer> {
  return crypto.subtle.digest('SHA-256', data)
}

/**
 * HMAC-SHA256 signing using Web Crypto API.
 */
async function hmacSha256(key: ArrayBuffer, message: Uint8Array): Promise<string> {
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const signature = await crypto.subtle.sign('HMAC', cryptoKey, message)
  return bytesToHex(new Uint8Array(signature))
}

// ─── Auth ───

/**
 * Parse a compound API key into key ID and secret.
 * Format: "ak_{env}_{id}.sk_{env}_{secret}"
 */
function parseAPIKey(apiKey: string): { keyId: string; secret: string } {
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
async function signRequest(
  secret: string,
  timestamp: string,
  method: string,
  path: string,
  body: string,
): Promise<string> {
  const signingKey = await sha256(stringToBytes(secret))
  const message = stringToBytes(timestamp + method.toUpperCase() + path + body)
  return hmacSha256(signingKey, message)
}

/**
 * Build authentication headers for an API request.
 */
async function buildAuthHeaders(
  keyId: string,
  secret: string,
  method: string,
  path: string,
  body: string,
): Promise<Record<string, string>> {
  const timestamp = Date.now().toString()
  const signature = await signRequest(secret, timestamp, method, path, body)

  return {
    'x-api-key': keyId,
    'x-timestamp': timestamp,
    'x-signature': signature,
  }
}

// ─── Webhook Verification ───

/**
 * Verify an incoming webhook signature.
 *
 * Note: ED25519 verification requires a crypto library in React Native.
 * This function validates the structure and timestamp, but for full ED25519
 * signature verification, use a library like `react-native-ed25519` or
 * verify on your backend.
 *
 * @param payload - The raw request body string
 * @param headers - The request headers containing signature, timestamp, event, and ID
 * @param _publicKey - The ED25519 public key (base64-encoded) — reserved for future use
 * @returns The parsed webhook event (after timestamp validation)
 */
export function verifyWebhookSignature(
  payload: string,
  headers: WebhookHeaders,
  _publicKey: string,
): WebhookEvent {
  const signature = headers['x-webhook-signature']
  const timestamp = headers['x-webhook-timestamp']
  const event = headers['x-webhook-event']
  const id = headers['x-webhook-id']

  if (!signature || !timestamp || !event || !id) {
    throw new OpenPayError('Missing webhook signature headers')
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

// ─── Payment Resource ───

class PaymentsResource {
  constructor(private client: OpenPay) {}

  /**
   * Create a new payment.
   */
  async create(input: CreatePaymentInput): Promise<Payment> {
    const res = await this.client.request<{ data: Payment }>('POST', '/v1/sdk/payments', input)
    return res.data
  }

  /**
   * Get a payment by ID.
   */
  async get(id: string): Promise<Payment> {
    const res = await this.client.request<{ data: Payment }>('GET', `/v1/sdk/payments/${id}`)
    return res.data
  }

  /**
   * List payments with optional filtering and pagination.
   */
  async list(params: ListPaymentsParams = {}): Promise<PaginatedResponse<Payment>> {
    const query = new URLSearchParams()
    if (params.page) query.set('page', String(params.page))
    if (params.perPage) query.set('perPage', String(params.perPage))
    if (params.status) query.set('status', params.status)
    if (params.search) query.set('search', params.search)
    if (params.branchId) query.set('branchId', params.branchId)
    if (params.dateFrom) query.set('dateFrom', params.dateFrom)
    if (params.dateTo) query.set('dateTo', params.dateTo)

    const qs = query.toString()
    const path = '/v1/sdk/payments' + (qs ? `?${qs}` : '')
    return this.client.request<PaginatedResponse<Payment>>('GET', path)
  }
}

// ─── Checkout Resource ───

class CheckoutResource {
  constructor(private client: OpenPay) {}

  /**
   * Create a checkout session. Returns a hosted checkout URL.
   */
  async createSession(input: CheckoutSessionInput): Promise<CheckoutSession> {
    const res = await this.client.request<{ data: CheckoutSession }>(
      'POST',
      '/v1/sdk/checkout/sessions',
      input,
    )
    return res.data
  }
}

// ─── Main Client ───

/**
 * Open Pay React Native SDK Client.
 *
 * @example
 * ```typescript
 * import { OpenPay } from '@openpay/react-native'
 *
 * const openpay = new OpenPay('ak_live_xxx.sk_live_yyy')
 *
 * const payment = await openpay.payments.create({
 *   amount: '1000.00',
 *   currency: 'LKR',
 *   merchantTradeNo: 'ORDER-123',
 * })
 *
 * console.log(payment.id, payment.checkoutLink)
 * ```
 */
export class OpenPay {
  private readonly keyId: string
  private readonly secret: string
  private readonly baseURL: string
  private readonly timeout: number

  readonly payments: PaymentsResource
  readonly checkout: CheckoutResource

  constructor(apiKey: string, options: ClientOptions = {}) {
    const parsed = parseAPIKey(apiKey)
    this.keyId = parsed.keyId
    this.secret = parsed.secret
    this.baseURL = (options.baseURL || DEFAULT_BASE_URL).replace(/\/$/, '')
    this.timeout = options.timeout || 30_000

    this.payments = new PaymentsResource(this)
    this.checkout = new CheckoutResource(this)
  }

  /**
   * Make an authenticated API request.
   * @internal
   */
  async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const bodyStr = body ? JSON.stringify(body) : ''
    const headers = await buildAuthHeaders(this.keyId, this.secret, method, path, bodyStr)

    const url = this.baseURL + path
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), this.timeout)

    try {
      const res = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers,
        },
        body: bodyStr || undefined,
        signal: controller.signal,
      })

      const json = (await res.json()) as Record<string, unknown>

      if (!res.ok) {
        const error = (json as APIErrorResponse).error
        if (res.status === 401) {
          throw new AuthenticationError(error?.message || 'Authentication failed')
        }
        throw new APIError(
          error?.code || 'UNKNOWN_ERROR',
          error?.message || `HTTP ${res.status}`,
          res.status,
        )
      }

      return json as T
    } catch (err) {
      if (err instanceof OpenPayError) throw err
      if (err instanceof Error && err.name === 'AbortError') {
        throw new OpenPayError('Request timed out')
      }
      throw new OpenPayError(`Request failed: ${err}`)
    } finally {
      clearTimeout(timer)
    }
  }
}
