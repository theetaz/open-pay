import { Input } from '#/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { StatCard } from '#/components/dashboard/stat-card'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Search } from 'lucide-react'

export function AuditLogPage() {
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Total Logs" value="0" description="All-time audit logs" />
        <StatCard title="Last 24 Hours" value="0" description="Recent activity" />
        <StatCard title="Newest Log" value="-" description="Most recent entry" />
        <StatCard title="Oldest Log" value="-" description="First recorded entry" />
      </div>

      <div className="flex items-center gap-4 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input placeholder="Search by Entity ID" className="pl-9" />
        </div>
        <Select defaultValue="all-types">
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="All Entity Types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all-types">All Entity Types</SelectItem>
            <SelectItem value="merchant">Merchant</SelectItem>
            <SelectItem value="apikey">Api Key</SelectItem>
            <SelectItem value="payment">Payment</SelectItem>
            <SelectItem value="withdrawal">Withdrawal</SelectItem>
          </SelectContent>
        </Select>
        <Select defaultValue="all-actions">
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="All Actions" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all-actions">All Actions</SelectItem>
            <SelectItem value="CREATE">CREATE</SelectItem>
            <SelectItem value="UPDATE">UPDATE</SelectItem>
            <SelectItem value="DELETE">DELETE</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Action</TableHead>
            <TableHead>Entity Type</TableHead>
            <TableHead>Entity ID</TableHead>
            <TableHead>Actor</TableHead>
            <TableHead>Source</TableHead>
            <TableHead>Timestamp</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell colSpan={6}>
              <EmptyState message="No audit log entries found." />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </>
  )
}
