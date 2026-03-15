import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/treasury')({ component: TreasuryPage })

function TreasuryPage() {
  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">Treasury</h2>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Crypto Holdings</p>
          <p className="text-2xl font-bold mt-1">0.00 USDT</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Fees Earned</p>
          <p className="text-2xl font-bold mt-1">0.00 USDT</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Total Settled</p>
          <p className="text-2xl font-bold mt-1">0.00 LKR</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Simulated Bank Balance</p>
          <p className="text-2xl font-bold mt-1">10,000,000.00 LKR</p>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h3 className="font-semibold mb-4">Treasury Transactions</h3>
        <div className="rounded-lg border border-border overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Date</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Type</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Amount</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-sm text-muted-foreground">
                  No treasury transactions yet.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
