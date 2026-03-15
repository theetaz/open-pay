import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/pay/$slug')({ component: PaymentLinkPage })

function PaymentLinkPage() {
  const { slug } = Route.useParams()

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-primary">Open Pay</h1>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          <div className="text-center mb-6">
            <h2 className="text-xl font-semibold">Payment Link</h2>
            <p className="text-sm text-muted-foreground mt-1">Ref: {slug}</p>
          </div>

          <div className="space-y-4">
            <div className="rounded-lg border border-border p-4 text-center">
              <p className="text-sm text-muted-foreground">Amount</p>
              <p className="text-2xl font-bold mt-1">25.00 USDT</p>
            </div>

            <div>
              <label className="text-sm font-medium block mb-1">Quantity</label>
              <div className="flex items-center gap-2">
                <button className="rounded-md border border-border px-3 py-1 text-sm hover:bg-accent">-</button>
                <span className="text-sm font-medium w-8 text-center">1</span>
                <button className="rounded-md border border-border px-3 py-1 text-sm hover:bg-accent">+</button>
              </div>
            </div>

            <div>
              <label className="text-sm font-medium block mb-1">Email</label>
              <input
                type="email"
                placeholder="your@email.com"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            <div>
              <p className="text-sm font-medium mb-2">Select Provider</p>
              <div className="grid grid-cols-3 gap-2">
                <ProviderOption name="Bybit" />
                <ProviderOption name="Binance" />
                <ProviderOption name="KuCoin" />
              </div>
            </div>

            <button className="w-full rounded-md bg-primary text-primary-foreground py-3 text-sm font-medium hover:bg-primary/90 transition-colors">
              Pay 25.00 USDT
            </button>
          </div>
        </div>

        <p className="text-center text-xs text-muted-foreground mt-4">
          Secure payment powered by Open Pay
        </p>
      </div>
    </div>
  )
}

function ProviderOption({ name }: { name: string }) {
  return (
    <button className="rounded-md border border-border px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors">
      {name}
    </button>
  )
}
