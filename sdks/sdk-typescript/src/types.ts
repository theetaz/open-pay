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

// ─── Webhook Types ───

export interface WebhookConfig {
  url: string
  events?: string[]
  secret?: string
}

export interface WebhookEvent {
  id: string
  event: string
  timestamp: string
  data: Record<string, unknown>
}

// ─── API Key Types ───

export interface APIKey {
  id: string
  keyId: string
  name: string
  environment: string
  isActive: boolean
  revokedAt?: string
  lastUsedAt?: string
  createdAt: string
}

export interface CreateAPIKeyInput {
  name: string
  environment?: 'live' | 'test'
}

export interface CreatedAPIKey extends APIKey {
  secret: string
}
