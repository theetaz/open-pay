const crypto = require('crypto')

class OpenPayClient {
  constructor(baseUrl, apiKey, apiSecret) {
    this.baseUrl = baseUrl
    this.apiKey = apiKey
    this.apiSecret = apiSecret
  }

  /**
   * Generate HMAC-SHA256 signature for API request.
   */
  sign(method, path, body, timestamp) {
    const message = `${method}${path}${body}${timestamp}`
    const hmacKey = crypto.createHash('sha256').update(this.apiSecret).digest('hex')
    return crypto.createHmac('sha256', hmacKey).update(message).digest('hex')
  }

  /**
   * Make an authenticated API request.
   */
  async request(method, path, data = null) {
    const timestamp = String(Date.now())
    const body = data ? JSON.stringify(data) : ''
    const signature = this.sign(method, `/v1/sdk${path}`, body, timestamp)

    const url = `${this.baseUrl}/v1/sdk${path}`
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': this.apiKey,
        'x-timestamp': timestamp,
        'x-signature': signature,
      },
    }

    if (body && method !== 'GET') {
      options.body = body
    }

    const res = await fetch(url, options)
    const json = await res.json()

    if (!res.ok) {
      throw new Error(json.error?.message || `API error: ${res.status}`)
    }

    return json.data
  }

  async createPayment(params) {
    return this.request('POST', '/payments', params)
  }

  async getPayment(id) {
    return this.request('GET', `/payments/${id}`)
  }
}

module.exports = { OpenPayClient }
