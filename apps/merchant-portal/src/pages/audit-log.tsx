import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Card, CardContent } from '#/components/ui/card'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { api } from '#/lib/api'
import { isAuthenticated } from '#/lib/auth'
import { Search, ChevronLeft, ChevronRight, Activity } from 'lucide-react'

const PER_PAGE = 20

interface AuditLog {
  id: string
  actorId: string
  actorType: string
  action: string
  resourceType: string
  resourceId: string
  ipAddress: string
  createdAt: string
}

interface AuditLogsResponse {
  data: AuditLog[]
  meta: { total: number; page: number; perPage: number }
}

export function AuditLogPage() {
  const [page, setPage] = React.useState(1)
  const [searchInput, setSearchInput] = React.useState('')
  const [actionFilter, setActionFilter] = React.useState('')
  const [debouncedSearch, setDebouncedSearch] = React.useState('')

  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput)
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data, isLoading } = useQuery({
    queryKey: ['merchant', 'audit-logs', { page, perPage: PER_PAGE, action: actionFilter || debouncedSearch }],
    queryFn: () => {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('perPage', String(PER_PAGE))
      if (actionFilter && actionFilter !== 'all') params.set('action', actionFilter)
      else if (debouncedSearch) params.set('action', debouncedSearch)
      return api.get<AuditLogsResponse>(`/v1/merchant/audit-logs?${params.toString()}`)
    },
    enabled: isAuthenticated(),
  })

  const logs = data?.data || []
  const total = data?.meta?.total || 0
  const totalPages = Math.ceil(total / PER_PAGE)

  return (
    <>
      <PageHeader title="Audit Log" description="Track all actions and events for your merchant account" />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Total Logs" value={String(total)} description="All-time audit logs" icon={Activity} />
      </div>

      <div className="flex items-center gap-4 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Search by action..."
            className="pl-9"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
        <Select value={actionFilter} onValueChange={(v) => { if (v) { setActionFilter(v); setPage(1) } }}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="All Actions" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Actions</SelectItem>
            <SelectItem value="merchant.registered">Registered</SelectItem>
            <SelectItem value="merchant.login">Login</SelectItem>
            <SelectItem value="merchant.approved">Approved</SelectItem>
            <SelectItem value="merchant.rejected">Rejected</SelectItem>
            <SelectItem value="merchant.deactivated">Deactivated</SelectItem>
            <SelectItem value="payment_link.created">Link Created</SelectItem>
            <SelectItem value="payment_link.deleted">Link Deleted</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Actor Type</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource Type</TableHead>
                <TableHead>IP Address</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <EmptyState
                      message={isLoading ? 'Loading audit logs...' : 'No audit log entries found.'}
                      description={
                        !isLoading
                          ? 'Actions like logins, payment link creation, and account changes will be logged here.'
                          : undefined
                      }
                    />
                  </TableCell>
                </TableRow>
              ) : (
                logs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                      {new Date(log.createdAt).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={log.actorType} />
                    </TableCell>
                    <TableCell className="font-medium">{log.action}</TableCell>
                    <TableCell className="text-sm">{log.resourceType}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{log.ipAddress || '-'}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {totalPages > 1 && (
            <div className="flex items-center justify-between border-t px-4 py-3">
              <p className="text-sm text-muted-foreground">
                Showing {(page - 1) * PER_PAGE + 1}–{Math.min(page * PER_PAGE, total)} of {total} logs
              </p>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  <ChevronLeft className="size-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                >
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </>
  )
}
