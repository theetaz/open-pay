import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { api } from '#/lib/api'
import { isAuthenticated } from '#/lib/auth'
import { Activity } from 'lucide-react'

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

const filters: FilterConfig[] = [
  {
    id: 'action',
    label: 'Action',
    type: 'select',
    options: [
      { label: 'All Actions', value: '' },
      { label: 'Registered', value: 'merchant.registered' },
      { label: 'Login', value: 'merchant.login' },
      { label: 'Approved', value: 'merchant.approved' },
      { label: 'Rejected', value: 'merchant.rejected' },
      { label: 'Deactivated', value: 'merchant.deactivated' },
      { label: 'Payment Initiated', value: 'payment.initiated' },
      { label: 'Payment Paid', value: 'payment.paid' },
      { label: 'Payment Expired', value: 'payment.expired' },
      { label: 'Payment Failed', value: 'payment.failed' },
      { label: 'Link Created', value: 'payment_link.created' },
      { label: 'Link Updated', value: 'payment_link.updated' },
      { label: 'Link Deleted', value: 'payment_link.deleted' },
      { label: 'Link Used', value: 'payment_link.used' },
    ],
  },
]

const columns: ColumnDef<AuditLog, any>[] = [
  {
    accessorKey: 'createdAt',
    header: 'Timestamp',
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground whitespace-nowrap">
        {new Date(row.original.createdAt).toLocaleString()}
      </span>
    ),
  },
  {
    accessorKey: 'actorType',
    header: 'Actor Type',
    cell: ({ row }) => <StatusBadge status={row.original.actorType} />,
  },
  {
    accessorKey: 'action',
    header: 'Action',
    cell: ({ row }) => <span className="font-medium">{row.original.action}</span>,
  },
  {
    accessorKey: 'resourceType',
    header: 'Resource Type',
    cell: ({ row }) => <span className="text-sm">{row.original.resourceType}</span>,
  },
  {
    accessorKey: 'ipAddress',
    header: 'IP Address',
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">{row.original.ipAddress || '-'}</span>
    ),
  },
]

export function AuditLogPage() {
  const [page, setPage] = React.useState(1)
  const [search, setSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({ action: '' })
  const [debouncedSearch, setDebouncedSearch] = React.useState('')

  const actionFilter = (filterValues.action as string) || ''

  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  const { data, isLoading } = useQuery({
    queryKey: ['merchant', 'audit-logs', { page, perPage: PER_PAGE, action: actionFilter || debouncedSearch }],
    queryFn: () => {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('perPage', String(PER_PAGE))
      if (actionFilter) params.set('action', actionFilter)
      else if (debouncedSearch) params.set('action', debouncedSearch)
      return api.get<AuditLogsResponse>(`/v1/merchant/audit-logs?${params.toString()}`)
    },
    enabled: isAuthenticated(),
  })

  const logs = data?.data || []
  const total = data?.meta?.total || 0

  function handleFilterChange(id: string, value: string | string[]) {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
    setPage(1)
  }

  function handleClearFilters() {
    setFilterValues({ action: '' })
    setPage(1)
  }

  function handleSearchChange(value: string) {
    setSearch(value)
  }

  return (
    <>
      <PageHeader title="Audit Log" description="Track all actions and events for your merchant account" />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Total Logs" value={String(total)} description="All-time audit logs" icon={Activity} />
      </div>

      <DataTable
        columns={columns}
        data={logs}
        filters={filters}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={handleSearchChange}
        searchPlaceholder="Search by action..."
        pagination={{ page, perPage: PER_PAGE, total }}
        onPageChange={setPage}
        isLoading={isLoading}
      />
    </>
  )
}
