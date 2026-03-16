import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Search, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '#/lib/api'

const PER_PAGE = 20

interface AuditLog {
  id: string
  actorId: string
  actorType: string
  action: string
  resourceType: string
  resourceId: string
  ipAddress: string
  userAgent: string
  merchantId: string
  changes: any
  metadata: any
  createdAt: string
}

interface AuditLogsResponse {
  data: AuditLog[]
  meta: { total: number; page: number; perPage: number }
}

export function AuditLogsPage() {
  const [page, setPage] = React.useState(1)
  const [searchInput, setSearchInput] = React.useState('')
  const [debouncedAction, setDebouncedAction] = React.useState('')

  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedAction(searchInput)
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'audit-logs', { page, perPage: PER_PAGE, action: debouncedAction }],
    queryFn: () => {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('perPage', String(PER_PAGE))
      if (debouncedAction) {
        params.set('action', debouncedAction)
      }
      return api.get<AuditLogsResponse>(`/v1/audit-logs?${params.toString()}`)
    },
    retry: false,
  })

  const logs = data?.data || []
  const total = data?.meta?.total || 0
  const totalPages = Math.ceil(total / PER_PAGE)

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
            placeholder="Search by action..."
            className="pl-9"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
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
                          ? 'Actions like merchant approvals, withdrawal processing, and API key management will be logged here.'
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

          {/* Pagination */}
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
                  Previous
                </Button>
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                  .reduce<(number | 'ellipsis')[]>((acc, p, idx, arr) => {
                    if (idx > 0 && p - (arr[idx - 1] as number) > 1) {
                      acc.push('ellipsis')
                    }
                    acc.push(p)
                    return acc
                  }, [])
                  .map((item, idx) =>
                    item === 'ellipsis' ? (
                      <span key={`ellipsis-${idx}`} className="px-2 text-muted-foreground">...</span>
                    ) : (
                      <Button
                        key={item}
                        variant={page === item ? 'default' : 'outline'}
                        size="sm"
                        className="min-w-8"
                        onClick={() => setPage(item as number)}
                      >
                        {item}
                      </Button>
                    )
                  )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                >
                  Next
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
