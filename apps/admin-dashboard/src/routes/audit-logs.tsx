import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/audit-logs')({ component: AuditLogsPage })

function AuditLogsPage() {
  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">Audit Logs</h2>

      <div className="flex gap-2 mb-4">
        <input
          type="text"
          placeholder="Search by action, actor, or resource..."
          className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        />
        <select className="rounded-md border border-input bg-background px-3 py-2 text-sm">
          <option value="">All Actions</option>
          <option value="merchant.approved">Merchant Approved</option>
          <option value="merchant.rejected">Merchant Rejected</option>
          <option value="withdrawal.approved">Withdrawal Approved</option>
          <option value="withdrawal.rejected">Withdrawal Rejected</option>
          <option value="apikey.created">API Key Created</option>
        </select>
      </div>

      <div className="rounded-lg border border-border bg-card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Timestamp</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Actor</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Action</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Resource</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">IP Address</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={5} className="px-4 py-12 text-center text-sm text-muted-foreground">
                No audit log entries yet.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}
