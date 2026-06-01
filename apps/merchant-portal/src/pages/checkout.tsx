import * as React from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { QRCodeSVG } from 'qrcode.react'
import { useCheckout } from '#/hooks/use-payments'
import { CheckCircle2, XCircle, Clock, Smartphone } from 'lucide-react'

export function CheckoutPage() {
  const { paymentId } = useParams<{ paymentId: string }>()
  const [searchParams] = useSearchParams()
  const successUrl = searchParams.get('successUrl')
  const cancelUrl = searchParams.get('cancelUrl')
  const { data, isLoading, isError } = useCheckout(paymentId!)

  const payment = data?.data

  // Auto-redirect after payment completes (must be before early returns — React hooks rules)
  React.useEffect(() => {
    if (!payment) return
    if (payment.status === 'PAID' && successUrl) {
      const timer = setTimeout(() => { window.location.href = successUrl }, 3000)
      return () => clearTimeout(timer)
    }
    if ((payment.status === 'EXPIRED' || payment.status === 'FAILED') && cancelUrl) {
      const timer = setTimeout(() => { window.location.href = cancelUrl }, 3000)
      return () => clearTimeout(timer)
    }
  }, [payment?.status, successUrl, cancelUrl])

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
          <p className="text-muted-foreground mt-2">{payment.amountUsdt} {payment.currency} received</p>
          <p className="text-xs text-muted-foreground mt-1">Payment No: {payment.paymentNo}</p>
          {successUrl && (
            <p className="text-sm text-muted-foreground mt-4">Redirecting back to store in 3 seconds...</p>
          )}
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
          {cancelUrl && (
            <p className="text-sm text-muted-foreground mt-4">Redirecting back to store in 3 seconds...</p>
          )}
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
          {cancelUrl && (
            <p className="text-sm text-muted-foreground mt-4">Redirecting back to store in 3 seconds...</p>
          )}
        </div>
      </div>
    )
  }

  // Sandbox mode: override QR to point to scannable mock wallet page
  const isSandbox = payment.qrContent?.startsWith('mock-qr://')
  let qrValue = payment.qrContent
  let providerPayId = ''

  if (isSandbox) {
    providerPayId = payment.qrContent.replace('mock-qr://', '').split('?')[0]
    qrValue = `${window.location.origin}/sandbox/pay/${providerPayId}?pid=${paymentId}`
  }

  // On-chain (EIP-681) mode: parse the deposit address + token amount so the
  // payer can scan with any wallet OR copy the address manually.
  const onchain = parseEIP681(payment.qrContent)

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
            <p className="text-3xl font-bold mt-1">{payment.amountUsdt} {payment.currency}</p>
            {payment.exchangeRate && (
              <p className="text-xs text-muted-foreground mt-1">
                ≈ {payment.amount} {payment.currency} @ {payment.exchangeRate}
              </p>
            )}
          </div>

          <div className="rounded-lg border border-border bg-muted/50 p-6 flex flex-col items-center">
            <div className="w-48 h-48 bg-white rounded-lg flex items-center justify-center mb-3 p-2">
              {qrValue ? (
                <QRCodeSVG value={qrValue} size={176} level="M" />
              ) : (
                <span className="text-muted-foreground text-xs">No QR data</span>
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {isSandbox
                ? 'Scan with your phone to open mock wallet'
                : onchain
                  ? 'Scan with MetaMask or any crypto wallet'
                  : 'Scan with your wallet app to pay'}
            </p>
          </div>

          {/* On-chain manual payment details (for desktop / wallets that can't scan) */}
          {onchain && (
            <div className="mt-4 rounded-md border border-border bg-muted/50 p-3 space-y-2">
              <p className="text-xs font-medium">Or send manually from any wallet</p>
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Send exactly</span>
                <span className="font-mono font-semibold text-foreground">
                  {onchain.amount} {payment.currency}
                </span>
              </div>
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Network</span>
                <span className="font-mono">BSC Testnet (97)</span>
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">To this address</p>
                <CopyableAddress address={onchain.address} />
              </div>
              <p className="text-[11px] text-amber-500/80">
                Send the exact amount of {payment.currency} on BSC Testnet only. Other tokens or networks will not be detected.
              </p>
            </div>
          )}

          <div className="mt-4 flex items-center justify-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full bg-amber-500 animate-pulse" />
            <span className="text-sm text-muted-foreground">Waiting for payment...</span>
          </div>

          {/* Sandbox hint */}
          {isSandbox && (
            <div className="mt-4 rounded-md border border-blue-500/20 bg-blue-500/5 p-3 flex items-start gap-2.5">
              <Smartphone className="size-4 text-blue-400 shrink-0 mt-0.5" />
              <div>
                <p className="text-xs text-blue-300 font-medium">Sandbox Mode</p>
                <p className="text-xs text-blue-400/70 mt-0.5">
                  Scan the QR code with your phone camera to open a mock wallet page where you can confirm or cancel the payment.
                </p>
              </div>
            </div>
          )}

          <div className="mt-4 rounded-md bg-muted/50 p-3 space-y-1">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Payment No</span>
              <span className="font-mono">{payment.paymentNo}</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Payment ID</span>
              <span className="font-mono">{paymentId!.slice(0, 12)}...</span>
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

// parseEIP681 extracts the deposit address and human-readable token amount from
// an EIP-681 ERC20 transfer URI:
//   ethereum:<token>@<chainId>/transfer?address=<deposit>&uint256=<baseUnits>
// Returns null for non-on-chain QR content (e.g. sandbox or CEX checkout URLs).
function parseEIP681(qr: string | undefined): { address: string; amount: string } | null {
  if (!qr || !qr.startsWith('ethereum:')) return null
  try {
    const url = new URL(qr.replace('ethereum:', 'https://').replace('@', '/chain/'))
    const address = url.searchParams.get('address') || ''
    const raw = url.searchParams.get('uint256') || '0'
    // Tokens here (USDC/USDT) use 6 decimals.
    const amount = (Number(raw) / 1_000_000).toString()
    if (!address) return null
    return { address, amount }
  } catch {
    return null
  }
}

function CopyableAddress({ address }: { address: string }) {
  const [copied, setCopied] = React.useState(false)
  return (
    <button
      onClick={() => {
        navigator.clipboard?.writeText(address)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      }}
      className="w-full rounded-md border border-border bg-background px-3 py-2 text-xs font-mono break-all text-left hover:bg-accent transition-colors"
      title="Click to copy"
    >
      {copied ? 'Copied!' : address}
    </button>
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
