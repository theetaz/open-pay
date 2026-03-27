// ─── Payment Types ───

export interface CreatePaymentInput {
  amount: string
  currency?: string
  provider?: string
  merchantTradeNo?: string
  description?: string
  webhookURL?: string
  successURL?: string
  cancelURL?: string
  customerEmail?: string
  customerBilling?: {
    firstName?: string
    lastName?: string
    phone?: string
  }
  goods?: Array<{
    name: string
    description?: string
    mccCode?: string
  }>
  orderExpireTime?: string
}

export interface Payment {
  id: string
  merchantId: string
  branchId?: string
  amount: string
  currency: string
  status: PaymentStatus
  provider: string
  providerPayId?: string
  merchantTradeNo: string
  qrContent?: string
  checkoutLink?: string
  deepLink?: string
  webhookURL?: string
  successURL?: string
  cancelURL?: string
  customerEmail?: string
  amountLkr?: string
  exchangeRate?: string
  platformFeeLkr?: string
  exchangeFeeLkr?: string
  netAmountLkr?: string
  paidAt?: string
  confirmedAt?: string
  createdAt: string
  updatedAt: string
}

export type PaymentStatus =
  | 'INITIATED'
  | 'PENDING'
  | 'PAID'
  | 'CONFIRMED'
  | 'EXPIRED'
  | 'FAILED'
  | 'REFUNDED'

export interface ListPaymentsParams {
  page?: number
  perPage?: number
  status?: PaymentStatus
  search?: string
  branchId?: string
  dateFrom?: string
  dateTo?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    perPage: number
    total: number
  }
}

// ─── Checkout Types ───

export interface CheckoutSessionInput {
  amount: string
  currency?: string
  provider?: string
  merchantTradeNo?: string
  description?: string
  successUrl?: string
  cancelUrl?: string
  customerEmail?: string
  lineItems?: Array<{ name: string; description?: string; amount?: string }>
  expiresInMinutes?: number
}

export interface CheckoutSession {
  id: string
  paymentId: string
  url: string
  amount: string
  currency: string
  amountUsdt: string
  status: string
  qrContent: string
  deepLink: string
  merchantTradeNo: string
  successUrl: string
  cancelUrl: string
  exchangeRate?: string
  expiresAt: string
  createdAt: string
}

// ─── Webhook Types ───

export interface WebhookEvent {
  id: string
  event: string
  timestamp: string
  data: Record<string, unknown>
}

export interface WebhookHeaders {
  'x-webhook-signature'?: string
  'x-webhook-timestamp'?: string
  'x-webhook-event'?: string
  'x-webhook-id'?: string
}

// ─── Client Types ───

export interface ClientOptions {
  baseURL?: string
  timeout?: number
}

// ─── Error Types ───

export interface APIErrorResponse {
  error?: {
    code: string
    message: string
  }
}
