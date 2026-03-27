import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate = new Rate('errors')
const listLatency = new Trend('list_payments_latency')

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '3m', target: 200 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'],
    errors: ['rate<0.01'],
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'
const JWT_TOKEN = __ENV.JWT_TOKEN || ''

export default function () {
  const page = Math.floor(Math.random() * 10) + 1
  const params = {
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  }

  const res = http.get(`${BASE_URL}/v1/payments?page=${page}&perPage=20`, params)

  listLatency.add(res.timings.duration)

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has data array': (r) => {
      try { return Array.isArray(JSON.parse(r.body).data) } catch { return false }
    },
  }) || errorRate.add(1)

  sleep(0.05)
}
