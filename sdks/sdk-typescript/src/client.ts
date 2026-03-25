import { parseAPIKey, buildAuthHeaders } from './auth.js'
import { APIError, AuthenticationError, OpenPayError } from './errors.js'
import type {
  CreatePaymentInput,
  Payment,
  ListPaymentsParams,
  PaginatedResponse,
  WebhookConfig,
} from './types.js'

const DEFAULT_BASE_URL = 'https://api.openpay.lk'

export interface ClientOptions {
  baseURL?: string
  timeout?: number
}

/**
 * Open Pay API Client.
 *
 * @example
 * ```typescript
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
  readonly webhooks: WebhooksResource

  constructor(apiKey: string, options: ClientOptions = {}) {
    const parsed = parseAPIKey(apiKey)
    this.keyId = parsed.keyId
    this.secret = parsed.secret
    this.baseURL = (options.baseURL || DEFAULT_BASE_URL).replace(/\/$/, '')
    this.timeout = options.timeout || 30_000

    this.payments = new PaymentsResource(this)
    this.webhooks = new WebhooksResource(this)
  }

  /** @internal Make an authenticated API request */
  async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const bodyStr = body ? JSON.stringify(body) : ''
    const headers = buildAuthHeaders(this.keyId, this.secret, method, path, bodyStr)

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

      const json = await res.json() as Record<string, unknown>

      if (!res.ok) {
        const error = json.error as { code: string; message: string } | undefined
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

// ─── Webhooks Resource ───

class WebhooksResource {
  constructor(private client: OpenPay) {}

  /**
   * Configure the webhook endpoint for your merchant account.
   */
  async configure(config: WebhookConfig): Promise<void> {
    await this.client.request('POST', '/v1/sdk/webhooks/configure', config)
  }

  /**
   * Get the ED25519 public key for verifying webhook signatures.
   */
  async getPublicKey(): Promise<string> {
    const res = await this.client.request<{ data: { publicKey: string } }>('GET', '/v1/sdk/webhooks/public-key')
    return res.data.publicKey
  }

  /**
   * Send a test webhook to your configured endpoint.
   */
  async test(): Promise<void> {
    await this.client.request('POST', '/v1/sdk/webhooks/test')
  }
}
