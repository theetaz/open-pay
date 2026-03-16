import { Card, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Search } from 'lucide-react'

export function AuditLogsPage() {
  return (
    <>
      <PageHeader
        title="Audit Logs"
        description="Track all administrative actions and system events"
      />

      <div className="flex gap-2 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground size-4" />
          <Input
            type="text"
            placeholder="Search by action, actor, or resource..."
            className="pl-9"
          />
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>IP Address</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={5}>
                  <EmptyState
                    message="No audit log entries yet."
                    description="Actions like merchant approvals, withdrawal processing, and API key management will be logged here."
                  />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
