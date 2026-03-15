import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/settlements')({ component: SettlementsPage })

function SettlementsPage() {
  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Settlements</h2>
        <button className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors">
          Request Withdrawal
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Available Balance</p>
          <p className="text-2xl font-bold mt-1">0.00 USDT</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Pending Withdrawal</p>
          <p className="text-2xl font-bold mt-1">0.00 USDT</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Total Withdrawn</p>
          <p className="text-2xl font-bold mt-1">0.00 LKR</p>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h3 className="font-semibold mb-4">Withdrawal History</h3>
        <p className="text-sm text-muted-foreground">No withdrawals yet.</p>
      </div>
    </div>
  )
}
