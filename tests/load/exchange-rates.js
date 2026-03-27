import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate = new Rate('errors')
const rateLatency = new Trend('exchange_rate_latency')

export const options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '3m', target: 500 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<50'],
    errors: ['rate<0.001'],
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'

export default function () {
  const res = http.get(`${BASE_URL}/v1/exchange-rates/active`)

  rateLatency.add(res.timings.duration)

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has rate data': (r) => {
      try { return JSON.parse(r.body).data !== undefined } catch { return false }
    },
  }) || errorRate.add(1)

  sleep(0.01)
}
