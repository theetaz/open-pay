import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { StatCard } from '#/components/dashboard/stat-card'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Search, DollarSign, CreditCard, Clock, AlertTriangle } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/payments')({
  component: PaymentsPage,
})

function PaymentsPage() {
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Revenue" value="₹. 0.00" description="₮ 0.00" icon={DollarSign} />
        <StatCard title="Total Payments" value="0" description="All-time transactions" icon={CreditCard} />
        <StatCard title="Unsettled Amount" value="₹. 0.00" description="₮ 0.00" icon={Clock} valueClassName="text-amber-500" />
        <StatCard title="Unsettled Payments" value="0" description="Pending transactions" icon={AlertTriangle} valueClassName="text-amber-500" />
      </div>

      <div className="mt-6">
        <Tabs defaultValue="paid">
          <TabsList>
            <TabsTrigger value="paid">Paid</TabsTrigger>
            <TabsTrigger value="initiated">Initiated</TabsTrigger>
            <TabsTrigger value="failed">Failed</TabsTrigger>
            <TabsTrigger value="expired">Expired</TabsTrigger>
            <TabsTrigger value="all">All</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      <Card className="mt-4">
        <CardContent className="pt-4">
          <div className="flex items-center justify-between gap-4 mb-4">
            <div className="flex items-center gap-2 flex-1 max-w-sm">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                <Input placeholder="Search..." className="pl-9" />
              </div>
            </div>
            <Select defaultValue="7d">
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="Date range" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7d">Last 7 Days</SelectItem>
                <SelectItem value="30d">Last 30 Days</SelectItem>
                <SelectItem value="90d">Last 90 Days</SelectItem>
                <SelectItem value="all">All Time</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-5 gap-4 mb-4 text-sm">
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider">Total Payments</p>
              <p className="font-bold text-lg">0</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider">Gross Amount</p>
              <p className="font-bold text-lg">₮ 0.00</p>
              <p className="text-xs text-muted-foreground">₹. 0.00</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider">Total Fees</p>
              <p className="font-bold text-lg text-red-500">₹. 0.00</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider">Net Amount</p>
              <p className="font-bold text-lg">₹. 0.00</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground uppercase tracking-wider">Fee Breakdown</p>
              <p className="text-xs text-muted-foreground">Exchange: ₹. 0.00</p>
              <p className="text-xs text-muted-foreground">Platform: ₹. 0.00</p>
            </div>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Payment No</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Item</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Associations</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState message="No results." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
