import http from 'k6/http'
import { check, sleep, group } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate = new Rate('errors')
const flowLatency = new Trend('checkout_flow_latency')

export const options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '3m', target: 50 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    errors: ['rate<0.05'],
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'
const JWT_TOKEN = __ENV.JWT_TOKEN || ''

export default function () {
  const start = Date.now()

  let paymentId = ''

  group('create payment', () => {
    const payload = JSON.stringify({
      amount: (Math.random() * 50 + 5).toFixed(2),
      currency: 'USDT',
      provider: 'TEST',
      customerEmail: `checkout-${__VU}@test.com`,
    })

    const res = http.post(`${BASE_URL}/v1/payments`, payload, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${JWT_TOKEN}`,
      },
    })

    check(res, { 'payment created': (r) => r.status === 201 }) || errorRate.add(1)

    try {
      paymentId = JSON.parse(res.body).data.id
    } catch {
      errorRate.add(1)
      return
    }
  })

  if (!paymentId) return

  group('get checkout', () => {
    const res = http.get(`${BASE_URL}/v1/payments/${paymentId}/checkout`)
    check(res, { 'checkout loaded': (r) => r.status === 200 }) || errorRate.add(1)
  })

  group('simulate callback', () => {
    const res = http.post(`${BASE_URL}/v1/payments/${paymentId}/callback`, '{}', {
      headers: { 'Content-Type': 'application/json' },
    })
    check(res, { 'callback accepted': (r) => r.status === 200 }) || errorRate.add(1)
  })

  flowLatency.add(Date.now() - start)
  sleep(0.5)
}
