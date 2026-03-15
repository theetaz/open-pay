import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/withdrawal')({
  component: WithdrawalPage,
})

function WithdrawalPage() {
  return (
    <>
      <PageHeader
        title="Withdrawal Requests"
        description="View and manage withdrawal requests for settlements"
        action={
          <Button>
            <Plus className="mr-2 size-4" /> Create Request
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Total Requests" value="0" />
        <StatCard title="Total Amount" value="₹. 0.00" />
        <StatCard title="To Be Claimed" value="₹. 0.00" />
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground mb-2">Status Overview</p>
            <div className="space-y-1 text-sm">
              <div className="flex justify-between">
                <span>Approved:</span>
                <span className="font-medium text-green-500">0</span>
              </div>
              <div className="flex justify-between">
                <span>Pending:</span>
                <span className="font-medium text-amber-500">0</span>
              </div>
              <div className="flex justify-between">
                <span>Rejected:</span>
                <span className="font-medium text-red-500">0</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Settlement Date</TableHead>
                <TableHead>Requested By</TableHead>
                <TableHead>Requested At</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Payments</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Reviewed By</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState message="No withdrawal requests found." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
