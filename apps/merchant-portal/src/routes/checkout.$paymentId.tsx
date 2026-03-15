import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/checkout/$paymentId')({ component: CheckoutPage })

function CheckoutPage() {
  const { paymentId } = Route.useParams()

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
            <p className="text-3xl font-bold mt-1">10.00 USDT</p>
            <p className="text-xs text-muted-foreground mt-1">≈ 3,250.00 LKR</p>
          </div>

          <div className="mb-6">
            <p className="text-sm font-medium mb-3">Select Payment Provider</p>
            <div className="grid grid-cols-3 gap-2">
              <ProviderButton name="Bybit" selected />
              <ProviderButton name="Binance" />
              <ProviderButton name="KuCoin" />
            </div>
          </div>

          <div className="rounded-lg border border-border bg-muted/50 p-6 flex flex-col items-center">
            <div className="w-48 h-48 bg-muted rounded-lg flex items-center justify-center mb-3">
              <span className="text-muted-foreground text-sm">QR Code</span>
            </div>
            <p className="text-xs text-muted-foreground">Scan with your wallet app to pay</p>
          </div>

          <div className="mt-4 flex items-center justify-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full bg-warning animate-pulse" />
            <span className="text-sm text-muted-foreground">Waiting for payment...</span>
          </div>

          <div className="mt-4 rounded-md bg-muted/50 p-3">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Payment ID</span>
              <span className="font-mono">{paymentId.slice(0, 12)}...</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>Expires in</span>
              <span>14:59</span>
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

function ProviderButton({ name, selected }: { name: string; selected?: boolean }) {
  return (
    <button
      className={`rounded-md border px-3 py-2 text-sm font-medium transition-colors ${
        selected
          ? 'border-primary bg-primary/10 text-primary'
          : 'border-border text-muted-foreground hover:bg-accent'
      }`}
    >
      {name}
    </button>
  )
}
