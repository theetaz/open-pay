import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useCheckout } from '#/hooks/use-payments'
import { CheckCircle2, XCircle, Clock } from 'lucide-react'

export const Route = createFileRoute('/checkout/$paymentId')({ component: CheckoutPage })

function CheckoutPage() {
  const { paymentId } = Route.useParams()
  const { data, isLoading, isError } = useCheckout(paymentId)

  const payment = data?.data

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto" />
          <p className="text-sm text-muted-foreground mt-4">Loading payment...</p>
        </div>
      </div>
    )
  }

  if (isError || !payment) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <XCircle className="h-12 w-12 text-destructive mx-auto" />
          <p className="text-lg font-medium mt-4">Payment not found</p>
          <p className="text-sm text-muted-foreground mt-1">This payment link is invalid or has been removed.</p>
        </div>
      </div>
    )
  }

  if (payment.status === 'PAID') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="w-full max-w-md text-center">
          <CheckCircle2 className="h-16 w-16 text-green-500 mx-auto" />
          <h2 className="text-2xl font-bold mt-4">Payment Successful</h2>
          <p className="text-muted-foreground mt-2">{payment.amountUsdt} USDT received</p>
          <p className="text-xs text-muted-foreground mt-1">Payment No: {payment.paymentNo}</p>
        </div>
      </div>
    )
  }

  if (payment.status === 'EXPIRED') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="w-full max-w-md text-center">
          <Clock className="h-16 w-16 text-amber-500 mx-auto" />
          <h2 className="text-2xl font-bold mt-4">Payment Expired</h2>
          <p className="text-muted-foreground mt-2">This payment link has expired.</p>
        </div>
      </div>
    )
  }

  if (payment.status === 'FAILED') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="w-full max-w-md text-center">
          <XCircle className="h-16 w-16 text-destructive mx-auto" />
          <h2 className="text-2xl font-bold mt-4">Payment Failed</h2>
          <p className="text-muted-foreground mt-2">Something went wrong with this payment.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-primary">Open Pay</h1>
          <p className="text-sm text-muted-foreground mt-1">Secure Crypto Payment</p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          <div className="text-center mb-6">
            <p className="text-sm text-muted-foreground">Amount Due</p>
            <p className="text-3xl font-bold mt-1">{payment.amountUsdt} USDT</p>
            {payment.exchangeRate && (
              <p className="text-xs text-muted-foreground mt-1">
                ≈ {payment.amount} {payment.currency}
              </p>
            )}
          </div>

          <div className="rounded-lg border border-border bg-muted/50 p-6 flex flex-col items-center">
            <div className="w-48 h-48 bg-muted rounded-lg flex items-center justify-center mb-3 break-all p-2">
              <span className="text-muted-foreground text-xs text-center font-mono">
                {payment.qrContent || 'QR Code'}
              </span>
            </div>
            <p className="text-xs text-muted-foreground">Scan with your wallet app or copy the payment URI</p>
          </div>

          <div className="mt-4 flex items-center justify-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full bg-amber-500 animate-pulse" />
            <span className="text-sm text-muted-foreground">Waiting for payment...</span>
          </div>

          <div className="mt-4 rounded-md bg-muted/50 p-3 space-y-1">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Payment No</span>
              <span className="font-mono">{payment.paymentNo}</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Payment ID</span>
              <span className="font-mono">{paymentId.slice(0, 12)}...</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Expires</span>
              <CountdownTimer expireTime={payment.expireTime} />
            </div>
          </div>
        </div>

        <p className="text-center text-xs text-muted-foreground mt-4">
          Powered by Open Pay — Secure crypto payments for Sri Lanka
        </p>
      </div>
    </div>
  )
}

function CountdownTimer({ expireTime }: { expireTime: string }) {
  const [remaining, setRemaining] = React.useState('')

  React.useEffect(() => {
    const update = () => {
      const diff = new Date(expireTime).getTime() - Date.now()
      if (diff <= 0) {
        setRemaining('Expired')
        return
      }
      const mins = Math.floor(diff / 60000)
      const secs = Math.floor((diff % 60000) / 1000)
      setRemaining(`${mins}:${secs.toString().padStart(2, '0')}`)
    }

    update()
    const interval = setInterval(update, 1000)
    return () => clearInterval(interval)
  }, [expireTime])

  return <span className="font-mono">{remaining}</span>
}
