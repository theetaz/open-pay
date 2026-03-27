import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate = new Rate('errors')
const paymentLatency = new Trend('payment_creation_latency')

export const options = {
  stages: [
    { duration: '1m', target: 20 },    // Ramp up
    { duration: '3m', target: 100 },   // Peak load
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],   // 95% of requests under 200ms
    errors: ['rate<0.01'],              // Error rate under 1%
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'
const JWT_TOKEN = __ENV.JWT_TOKEN || ''

export default function () {
  const payload = JSON.stringify({
    amount: (Math.random() * 100 + 1).toFixed(2),
    currency: 'USDT',
    provider: 'TEST',
    customerEmail: `load-test-${__VU}-${__ITER}@test.com`,
  })

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  }

  const res = http.post(`${BASE_URL}/v1/payments`, payload, params)

  paymentLatency.add(res.timings.duration)

  check(res, {
    'status is 201': (r) => r.status === 201,
    'has payment id': (r) => {
      try { return JSON.parse(r.body).data.id !== undefined } catch { return false }
    },
  }) || errorRate.add(1)

  sleep(0.1)
}
