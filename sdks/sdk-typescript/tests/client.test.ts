import { describe, it, expect } from 'vitest'
import { OpenPay, AuthenticationError } from '../src/index.js'

describe('OpenPay client', () => {
  it('creates a client with valid API key', () => {
    const client = new OpenPay('ak_live_xxx.sk_live_yyy')
    expect(client).toBeDefined()
    expect(client.payments).toBeDefined()
    expect(client.webhooks).toBeDefined()
  })

  it('throws on empty API key', () => {
    expect(() => new OpenPay('')).toThrow(AuthenticationError)
  })

  it('throws on invalid API key format', () => {
    expect(() => new OpenPay('invalid-key')).toThrow(AuthenticationError)
  })

  it('accepts custom base URL', () => {
    const client = new OpenPay('ak_live_xxx.sk_live_yyy', {
      baseURL: 'http://localhost:8080',
    })
    expect(client).toBeDefined()
  })
})
