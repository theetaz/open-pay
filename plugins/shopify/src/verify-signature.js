const crypto = require('crypto')

let cachedPublicKey = null
let cacheExpiry = 0

/**
 * Fetch the ED25519 public key from Open Pay API (cached for 10 minutes).
 */
async function getPublicKey(apiUrl) {
  if (cachedPublicKey && Date.now() < cacheExpiry) {
    return cachedPublicKey
  }

  try {
    const res = await fetch(`${apiUrl}/v1/webhooks/public-key`)
    if (!res.ok) return null

    const json = await res.json()
    cachedPublicKey = json.data?.publicKey || null
    cacheExpiry = Date.now() + 10 * 60 * 1000 // 10 minutes
    return cachedPublicKey
  } catch {
    return null
  }
}

/**
 * Verify ED25519 webhook signature.
 */
async function verifyWebhookSignature(payload, signatureHex, apiUrl) {
  const publicKeyHex = await getPublicKey(apiUrl)
  if (!publicKeyHex) return false

  try {
    const publicKey = Buffer.from(publicKeyHex, 'hex')
    const signature = Buffer.from(signatureHex, 'hex')
    const data = Buffer.from(payload)

    return crypto.verify(null, data, { key: publicKey, format: 'der', type: 'ed25519' }, signature)
  } catch {
    return false
  }
}

module.exports = { verifyWebhookSignature }
