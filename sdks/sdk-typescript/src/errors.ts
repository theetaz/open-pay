/**
 * Base error class for Open Pay SDK errors.
 */
export class OpenPayError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'OpenPayError'
  }
}

/**
 * Thrown when API authentication fails (invalid key, expired timestamp, bad signature).
 */
export class AuthenticationError extends OpenPayError {
  constructor(message = 'Authentication failed') {
    super(message)
    this.name = 'AuthenticationError'
  }
}

/**
 * Thrown when the API returns an error response.
 */
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

/**
 * Thrown when request validation fails.
 */
export class ValidationError extends OpenPayError {
  constructor(message: string) {
    super(message)
    this.name = 'ValidationError'
  }
}
