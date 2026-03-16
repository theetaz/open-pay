import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { usePublicPaymentLink } from '#/hooks/use-payment-links'
import { api } from '#/lib/api'
import { Loader2, XCircle, CreditCard } from 'lucide-react'

interface CreatePaymentResponse {
  data: {
    id: string
    paymentNo: string
    status: string
  }
}

export function PaymentLinkCheckout() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { data, isLoading, isError } = usePublicPaymentLink(slug!)

  const [email, setEmail] = useState('')
  const [provider, setProvider] = useState('TEST')
  const [customAmount, setCustomAmount] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const paymentLink = data?.data

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto" />
          <p className="text-sm text-muted-foreground mt-4">Loading...</p>
        </div>
      </div>
    )
  }

  if (isError || !paymentLink) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="text-center">
          <XCircle className="h-12 w-12 text-destructive mx-auto" />
          <p className="text-lg font-medium mt-4">Payment Link Not Found</p>
          <p className="text-sm text-muted-foreground mt-1">This payment link is invalid, expired, or has been deactivated.</p>
        </div>
      </div>
    )
  }

  const displayAmount = paymentLink.allowCustomAmount
    ? customAmount || '0.00'
    : paymentLink.amount

  const canSubmit = !submitting &&
    (paymentLink.allowCustomAmount ? parseFloat(customAmount) > 0 : true)

  const handlePay = async () => {
    setSubmitting(true)
    setError('')

    try {
      const res = await api.post<CreatePaymentResponse>('/v1/public/payments', {
        merchantId: paymentLink.merchantId,
        amount: paymentLink.allowCustomAmount ? customAmount : paymentLink.amount,
        currency: paymentLink.currency,
        provider,
        merchantTradeNo: `PL-${paymentLink.slug}`,
        customerEmail: email || undefined,
      })

      navigate(`/checkout/${res.data.id}`)
    } catch (err: any) {
      setError(err.message || 'Failed to create payment. Please try again.')
      setSubmitting(false)
    }
  }

  const providers = [
    { id: 'TEST', name: 'Test Pay', desc: 'Sandbox' },
    { id: 'BYBIT', name: 'Bybit', desc: 'Exchange' },
    { id: 'BINANCE', name: 'Binance', desc: 'Exchange' },
    { id: 'KUCOIN', name: 'KuCoin', desc: 'Exchange' },
  ]

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-primary">Open Pay</h1>
          <p className="text-sm text-muted-foreground mt-1">Secure Crypto Payment</p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          {/* Payment Link Info */}
          <div className="text-center mb-6">
            <h2 className="text-xl font-semibold">{paymentLink.name}</h2>
            {paymentLink.description && (
              <p className="text-sm text-muted-foreground mt-1">{paymentLink.description}</p>
            )}
          </div>

          {/* Amount */}
          <div className="rounded-lg border border-border p-4 text-center mb-6">
            {paymentLink.allowCustomAmount ? (
              <div>
                <p className="text-sm text-muted-foreground mb-2">Enter Amount ({paymentLink.currency})</p>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="0.00"
                  value={customAmount}
                  onChange={(e) => setCustomAmount(e.target.value)}
                  className="w-full text-center text-2xl font-bold bg-transparent border-none outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                />
              </div>
            ) : (
              <div>
                <p className="text-sm text-muted-foreground">Amount</p>
                <p className="text-2xl font-bold mt-1">
                  {parseFloat(paymentLink.amount).toLocaleString()} {paymentLink.currency}
                </p>
              </div>
            )}
          </div>

          <div className="space-y-4">
            {/* Email */}
            <div>
              <label className="text-sm font-medium block mb-1">Email (optional)</label>
              <input
                type="email"
                placeholder="your@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {/* Provider Selection */}
            <div>
              <p className="text-sm font-medium mb-2">Select Provider</p>
              <div className="grid grid-cols-2 gap-2">
                {providers.map((p) => (
                  <button
                    key={p.id}
                    onClick={() => setProvider(p.id)}
                    className={`rounded-md border px-3 py-2.5 text-sm transition-colors ${
                      provider === p.id
                        ? 'border-primary bg-primary/10 text-primary font-medium'
                        : 'border-border text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                    }`}
                  >
                    <span className="block font-medium">{p.name}</span>
                    <span className="block text-xs opacity-70">{p.desc}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Error */}
            {error && (
              <div className="rounded-md bg-destructive/10 border border-destructive/20 p-3">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}

            {/* Pay Button */}
            <button
              onClick={handlePay}
              disabled={!canSubmit}
              className="w-full rounded-md bg-primary text-primary-foreground py-3 text-sm font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {submitting ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <CreditCard className="size-4" />
              )}
              {submitting
                ? 'Creating payment...'
                : `Pay ${parseFloat(displayAmount).toLocaleString()} ${paymentLink.currency}`}
            </button>
          </div>
        </div>

        <p className="text-center text-xs text-muted-foreground mt-4">
          Powered by Open Pay — Secure crypto payments for Sri Lanka
        </p>
      </div>
    </div>
  )
}
