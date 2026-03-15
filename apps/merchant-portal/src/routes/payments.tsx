import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/payments')({ component: PaymentsPage })

function PaymentsPage() {
  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Payments</h2>
        <button className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors">
          Create Payment
        </button>
      </div>

      <div className="rounded-lg border border-border bg-card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Payment No</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Amount</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Provider</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Status</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Date</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={5} className="px-4 py-12 text-center text-sm text-muted-foreground">
                No payments found. Create your first payment to get started.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}
