import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
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

const ACTION_OPTIONS = [
  { label: 'Payment Initiated', value: 'payment.initiated' },
  { label: 'Payment Paid', value: 'payment.paid' },
  { label: 'Payment Expired', value: 'payment.expired' },
  { label: 'Payment Failed', value: 'payment.failed' },
  { label: 'Checkout Viewed', value: 'payment.checkout_viewed' },
  { label: 'Link Created', value: 'payment_link.created' },
  { label: 'Link Updated', value: 'payment_link.updated' },
  { label: 'Link Deleted', value: 'payment_link.deleted' },
  { label: 'Link Used', value: 'payment_link.used' },
  { label: 'Merchant Registered', value: 'merchant.registered' },
  { label: 'Merchant Approved', value: 'merchant.approved' },
  { label: 'Merchant Login', value: 'merchant.login' },
]

const ACTOR_TYPE_OPTIONS = [
  { label: 'Admin', value: 'ADMIN' },
  { label: 'Merchant User', value: 'MERCHANT_USER' },
  { label: 'System', value: 'SYSTEM' },
]

const RESOURCE_TYPE_OPTIONS = [
  { label: 'Payment', value: 'payment' },
  { label: 'Payment Link', value: 'payment_link' },
  { label: 'Merchant', value: 'merchant' },
  { label: 'Admin User', value: 'admin_user' },
]

const AUDIT_LOG_FILTERS: FilterConfig[] = [
  { id: 'action', label: 'Action', type: 'select', options: ACTION_OPTIONS },
  { id: 'actorType', label: 'Actor Type', type: 'select', options: ACTOR_TYPE_OPTIONS },
  { id: 'resourceType', label: 'Resource Type', type: 'select', options: RESOURCE_TYPE_OPTIONS },
]

export function AuditLogsPage() {
  const [page, setPage] = React.useState(1)
  const [search, setSearch] = React.useState('')
  const [debouncedSearch, setDebouncedSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({})

  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  const queryParams = React.useMemo(() => {
    const params = new URLSearchParams()
    params.set('page', String(page))
    params.set('perPage', String(PER_PAGE))
    if (debouncedSearch) params.set('action', debouncedSearch)
    for (const [key, value] of Object.entries(filterValues)) {
      if (typeof value === 'string' && value) params.set(key, value)
      if (Array.isArray(value) && value.length) params.set(key, value.join(','))
    }
    return params.toString()
  }, [page, debouncedSearch, filterValues])

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'audit-logs', queryParams],
    queryFn: () => api.get<AuditLogsResponse>(`/v1/audit-logs?${queryParams}`),
    retry: false,
  })

  const logs = data?.data || []
  const total = data?.meta?.total || 0

  const handleFilterChange = (id: string, value: string | string[]) => {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
    setPage(1)
  }

  const handleClearFilters = () => {
    setFilterValues({})
    setSearch('')
    setPage(1)
  }

  const columns: ColumnDef<AuditLog>[] = [
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

  return (
    <>
      <PageHeader
        title="Audit Logs"
        description="Track all administrative actions and system events"
      />

      <DataTable
        columns={columns}
        data={logs}
        filters={AUDIT_LOG_FILTERS}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search by action..."
        pagination={{ page, perPage: PER_PAGE, total }}
        onPageChange={setPage}
        isLoading={isLoading}
      />
    </>
  )
}
