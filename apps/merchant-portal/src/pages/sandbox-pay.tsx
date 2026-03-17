import { useState, useEffect } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { CheckCircle2, XCircle, Wallet, ShieldCheck } from 'lucide-react'

const API_BASE = (typeof window !== 'undefined' && import.meta.env.VITE_API_URL) || 'http://localhost:8080'

type Status = 'loading' | 'ready' | 'processing' | 'success' | 'failed' | 'error'

interface PaymentInfo {
  amount: string
  currency: string
  amountUsdt: string
  paymentNo: string
}

export function SandboxPayPage() {
  const { providerPayId } = useParams<{ providerPayId: string }>()
  const [searchParams] = useSearchParams()
  const paymentId = searchParams.get('pid')

  const [status, setStatus] = useState<Status>('loading')
  const [paymentInfo, setPaymentInfo] = useState<PaymentInfo | null>(null)

  useEffect(() => {
    if (!paymentId) {
      setStatus('error')
      return
    }

    fetch(`${API_BASE}/v1/payments/${paymentId}/checkout`)
      .then((res) => res.json())
      .then((data) => {
        if (data.data) {
          setPaymentInfo({
            amount: data.data.amount,
            currency: data.data.currency,
            amountUsdt: data.data.amountUsdt,
            paymentNo: data.data.paymentNo,
          })
          setStatus(data.data.status === 'PAID' ? 'success' : 'ready')
        } else {
          setStatus('error')
        }
      })
      .catch(() => setStatus('error'))
  }, [paymentId])

  const handlePay = async () => {
    if (!providerPayId || !paymentId) return
    setStatus('processing')

    try {
      await fetch(`${API_BASE}/test/simulate/${providerPayId}`, { method: 'POST' })
      await fetch(`${API_BASE}/v1/payments/${paymentId}/callback`, { method: 'POST' })
      setStatus('success')
    } catch {
      setStatus('failed')
    }
  }

  const handleCancel = async () => {
    setStatus('failed')
  }

  if (status === 'loading') {
    return (
      <Shell>
        <div className="flex flex-col items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
          <p className="text-sm text-gray-400 mt-4">Loading payment...</p>
        </div>
      </Shell>
    )
  }

  if (status === 'error') {
    return (
      <Shell>
        <div className="flex flex-col items-center justify-center py-12">
          <XCircle className="h-14 w-14 text-red-500" />
          <p className="text-lg font-semibold text-white mt-4">Invalid Payment</p>
          <p className="text-sm text-gray-400 mt-1">This payment link is invalid or has expired.</p>
        </div>
      </Shell>
    )
  }

  if (status === 'success') {
    return (
      <Shell>
        <div className="flex flex-col items-center justify-center py-12">
          <div className="rounded-full bg-green-500/20 p-4">
            <CheckCircle2 className="h-14 w-14 text-green-500" />
          </div>
          <p className="text-xl font-bold text-white mt-5">Payment Sent!</p>
          <p className="text-sm text-gray-400 mt-2">
            {paymentInfo?.amountUsdt} USDT has been sent successfully
          </p>
          <p className="text-xs text-gray-500 mt-1">Ref: {paymentInfo?.paymentNo}</p>
          <p className="text-xs text-gray-500 mt-4">You can close this page now.</p>
        </div>
      </Shell>
    )
  }

  if (status === 'failed') {
    return (
      <Shell>
        <div className="flex flex-col items-center justify-center py-12">
          <div className="rounded-full bg-red-500/20 p-4">
            <XCircle className="h-14 w-14 text-red-500" />
          </div>
          <p className="text-xl font-bold text-white mt-5">Payment Cancelled</p>
          <p className="text-sm text-gray-400 mt-2">The payment was not completed.</p>
        </div>
      </Shell>
    )
  }

  return (
    <Shell>
      {/* Wallet Header */}
      <div className="flex items-center gap-3 px-5 py-4 border-b border-gray-800">
        <div className="rounded-full bg-blue-500/20 p-2">
          <Wallet className="h-5 w-5 text-blue-400" />
        </div>
        <div>
          <p className="text-sm font-semibold text-white">Open Pay Sandbox Wallet</p>
          <p className="text-xs text-gray-500">Test Environment</p>
        </div>
      </div>

      {/* Payment Details */}
      <div className="px-5 py-6">
        <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Payment Request</p>

        <div className="rounded-xl bg-gray-800/50 border border-gray-700 p-5 mt-3">
          <div className="text-center">
            <p className="text-3xl font-bold text-white">{paymentInfo?.amountUsdt} USDT</p>
            {paymentInfo?.currency !== 'USDT' && (
              <p className="text-sm text-gray-400 mt-1">
                ≈ {parseFloat(paymentInfo?.amount || '0').toLocaleString()} {paymentInfo?.currency}
              </p>
            )}
          </div>

          <div className="mt-5 space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">To</span>
              <span className="text-gray-300">Open Pay Merchant</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Network</span>
              <span className="text-gray-300">USDT (TRC-20)</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Ref</span>
              <span className="text-gray-300 font-mono text-xs">{paymentInfo?.paymentNo}</span>
            </div>
          </div>
        </div>

        {/* Security Notice */}
        <div className="flex items-center gap-2 mt-4 px-1">
          <ShieldCheck className="h-4 w-4 text-green-500 shrink-0" />
          <p className="text-xs text-gray-500">Sandbox transaction — no real funds will be moved</p>
        </div>

        {/* Action Buttons */}
        <div className="mt-6 space-y-3">
          <button
            onClick={handlePay}
            disabled={status === 'processing'}
            className="w-full rounded-xl bg-blue-600 hover:bg-blue-700 text-white py-3.5 text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {status === 'processing' ? (
              <span className="flex items-center justify-center gap-2">
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                Processing...
              </span>
            ) : (
              `Confirm & Pay ${paymentInfo?.amountUsdt} USDT`
            )}
          </button>
          <button
            onClick={handleCancel}
            disabled={status === 'processing'}
            className="w-full rounded-xl border border-gray-700 text-gray-400 py-3.5 text-sm font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
        </div>
      </div>
    </Shell>
  )
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="rounded-2xl bg-gray-900 border border-gray-800 overflow-hidden shadow-2xl">
          {children}
        </div>
        <p className="text-center text-xs text-gray-600 mt-4">
          Open Pay Sandbox — Test Environment Only
        </p>
      </div>
    </div>
  )
}
