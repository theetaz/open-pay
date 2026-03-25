/**
 * Integration tests — require the local dev environment running (make start).
 * Run with: pnpm test -- --run tests/integration.test.ts
 */
import { describe, it, expect, beforeAll } from 'vitest'
import { OpenPay } from '../src/index.js'

const API_KEY = process.env.OPENPAY_API_KEY || ''
const BASE_URL = process.env.OPENPAY_BASE_URL || 'http://localhost:8080'

describe.skipIf(!API_KEY)('Integration: OpenPay SDK', () => {
  let client: OpenPay

  beforeAll(() => {
    client = new OpenPay(API_KEY, { baseURL: BASE_URL })
  })

  it('lists payments', async () => {
    const result = await client.payments.list()
    expect(result.data).toBeDefined()
    expect(result.meta).toBeDefined()
    expect(result.meta.page).toBe(1)
  })

  it('creates a payment', async () => {
    const payment = await client.payments.create({
      amount: '50.00',
      currency: 'LKR',
      merchantTradeNo: `TS-SDK-${Date.now()}`,
      description: 'TypeScript SDK test payment',
    })

    expect(payment.id).toBeDefined()
    expect(payment.status).toBe('INITIATED')
    expect(payment.amount).toBe('50')
  })

  it('gets a payment by ID', async () => {
    const created = await client.payments.create({
      amount: '25.00',
      currency: 'LKR',
      merchantTradeNo: `TS-GET-${Date.now()}`,
    })

    const fetched = await client.payments.get(created.id)
    expect(fetched.id).toBe(created.id)
    expect(fetched.amount).toBe('25')
  })
})
