import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate } from 'k6/metrics'

const errorRate = new Rate('errors')

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '2m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    errors: ['rate<0.05'],
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'
const TEST_EMAIL = __ENV.TEST_EMAIL || 'loadtest@example.com'
const TEST_PASSWORD = __ENV.TEST_PASSWORD || 'LoadTest123!'

export default function () {
  // Login
  const loginRes = http.post(`${BASE_URL}/v1/auth/login`, JSON.stringify({
    email: TEST_EMAIL,
    password: TEST_PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } })

  const loginOk = check(loginRes, {
    'login succeeds': (r) => r.status === 200,
  })

  if (!loginOk) {
    errorRate.add(1)
    sleep(1)
    return
  }

  let token = ''
  let refreshToken = ''
  try {
    const body = JSON.parse(loginRes.body)
    token = body.data.accessToken
    refreshToken = body.data.refreshToken
  } catch {
    errorRate.add(1)
    return
  }

  // Get profile
  const meRes = http.get(`${BASE_URL}/v1/auth/me`, {
    headers: { 'Authorization': `Bearer ${token}` },
  })

  check(meRes, { 'profile loaded': (r) => r.status === 200 }) || errorRate.add(1)

  // Refresh token
  const refreshRes = http.post(`${BASE_URL}/v1/auth/refresh`, JSON.stringify({
    refreshToken: refreshToken,
  }), { headers: { 'Content-Type': 'application/json' } })

  check(refreshRes, { 'token refreshed': (r) => r.status === 200 }) || errorRate.add(1)

  sleep(1)
}
