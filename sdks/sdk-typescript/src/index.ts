export { OpenPay } from './client.js'
export type { ClientOptions } from './client.js'

export { verifyWebhookSignature } from './webhook.js'
export { parseAPIKey, signRequest, currentTimestamp } from './auth.js'

export { OpenPayError, APIError, AuthenticationError, ValidationError } from './errors.js'

export type {
  Payment,
  PaymentStatus,
  CreatePaymentInput,
  ListPaymentsParams,
  PaginatedResponse,
  WebhookConfig,
  WebhookEvent,
  APIKey,
  CreateAPIKeyInput,
  CreatedAPIKey,
} from './types.js'
