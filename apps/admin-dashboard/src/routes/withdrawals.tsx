import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/withdrawals')({ component: WithdrawalsPage })

function WithdrawalsPage() {
  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">Withdrawal Approvals</h2>

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Pending</p>
          <p className="text-2xl font-bold mt-1 text-warning">0</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Approved Today</p>
          <p className="text-2xl font-bold mt-1 text-success">0</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-6">
          <p className="text-sm text-muted-foreground">Total Settled (LKR)</p>
          <p className="text-2xl font-bold mt-1">0.00</p>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Merchant</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Amount (USDT)</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Amount (LKR)</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Bank</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Status</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={6} className="px-4 py-12 text-center text-sm text-muted-foreground">
                No pending withdrawals.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}
