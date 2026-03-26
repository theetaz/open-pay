import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '#/components/ui/button'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { CheckCircle2, XCircle, BanknoteIcon, Clock, ArrowDownToLine } from 'lucide-react'
import { api } from '#/lib/api'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '#/components/ui/dialog'
import { DataTable, type FilterConfig } from '#/components/data-table'

const PER_PAGE = 10

interface Withdrawal {
  id: string
  amountUsdt: string
  exchangeRate: string
  amountLkr: string
  bankName: string
  bankAccountNo: string
  bankReference: string
  status: string
  createdAt: string
  [key: string]: any
}

interface WithdrawalsResponse {
  data: Withdrawal[]
  meta: { total: number; page: number; perPage: number }
}

const STATUS_OPTIONS = [
  { label: 'Requested', value: 'REQUESTED' },
  { label: 'Approved', value: 'APPROVED' },
  { label: 'Completed', value: 'COMPLETED' },
  { label: 'Rejected', value: 'REJECTED' },
]

const WITHDRAWAL_FILTERS: FilterConfig[] = [
  { id: 'status', label: 'Status', type: 'select', options: STATUS_OPTIONS },
]

export function WithdrawalsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = React.useState(1)
  const [search, setSearch] = React.useState('')
  const [debouncedSearch, setDebouncedSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({})
  const [confirmDialog, setConfirmDialog] = React.useState<{ type: 'approve' | 'reject' | 'complete'; withdrawal: Withdrawal } | null>(null)

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
    if (debouncedSearch) params.set('search', debouncedSearch)
    for (const [key, value] of Object.entries(filterValues)) {
      if (typeof value === 'string' && value) params.set(key, value)
      if (Array.isArray(value) && value.length) params.set(key, value.join(','))
    }
    return params.toString()
  }, [page, debouncedSearch, filterValues])

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'withdrawals', queryParams],
    queryFn: () => api.get<WithdrawalsResponse>(`/v1/withdrawals?${queryParams}`),
    retry: false,
  })

  const withdrawals = data?.data || []
  const total = data?.meta?.total || 0

  const pending = withdrawals.filter((w) => w.status === 'REQUESTED')
  const approved = withdrawals.filter((w) => w.status === 'APPROVED')
  const completed = withdrawals.filter((w) => w.status === 'COMPLETED')
  const totalSettledLKR = completed.reduce((sum, w) => sum + parseFloat(w.amountLkr || '0'), 0)

  const handleFilterChange = (id: string, value: string | string[]) => {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
    setPage(1)
  }

  const handleClearFilters = () => {
    setFilterValues({})
    setSearch('')
    setPage(1)
  }

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/approve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] })
      setConfirmDialog(null)
    },
  })

  const rejectMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/reject`, { reason: 'Rejected by admin' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] })
      setConfirmDialog(null)
    },
  })

  const completeMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/complete`, { bankReference: `TXN${Date.now()}` }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] })
      setConfirmDialog(null)
    },
  })

  const handleConfirm = () => {
    if (!confirmDialog) return
    const { type, withdrawal } = confirmDialog
    if (type === 'approve') approveMutation.mutate(withdrawal.id)
    if (type === 'reject') rejectMutation.mutate(withdrawal.id)
    if (type === 'complete') completeMutation.mutate(withdrawal.id)
  }

  const actionLabels: Record<string, { title: string; desc: string; btn: string; variant: 'default' | 'destructive' }> = {
    approve: { title: 'Approve Withdrawal', desc: 'Are you sure you want to approve this withdrawal request?', btn: 'Approve', variant: 'default' },
    reject: { title: 'Reject Withdrawal', desc: 'Are you sure you want to reject this withdrawal request?', btn: 'Reject', variant: 'destructive' },
    complete: { title: 'Complete Withdrawal', desc: 'Mark this withdrawal as completed with a bank reference?', btn: 'Complete', variant: 'default' },
  }

  const columns: ColumnDef<Withdrawal>[] = [
    {
      accessorKey: 'createdAt',
      header: 'Date',
      cell: ({ row }) => (
        <span className="text-sm">{new Date(row.original.createdAt).toLocaleDateString()}</span>
      ),
    },
    {
      accessorKey: 'amountUsdt',
      header: 'Amount (USDT)',
      cell: ({ row }) => <span className="font-medium">{row.original.amountUsdt}</span>,
    },
    {
      accessorKey: 'exchangeRate',
      header: 'Rate',
      cell: ({ row }) => <span className="text-sm">{row.original.exchangeRate}</span>,
    },
    {
      accessorKey: 'amountLkr',
      header: 'Amount (LKR)',
      cell: ({ row }) => <span className="font-medium">{row.original.amountLkr}</span>,
    },
    {
      accessorKey: 'bankName',
      header: 'Bank',
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.bankName}
          <br />
          <span className="text-muted-foreground">{row.original.bankAccountNo}</span>
        </span>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => <StatusBadge status={row.original.status} />,
    },
    {
      id: 'actions',
      header: 'Actions',
      enableHiding: false,
      cell: ({ row }) => {
        const w = row.original
        return (
          <div className="flex gap-1">
            {w.status === 'REQUESTED' && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-green-600 hover:text-green-700"
                  onClick={(e) => {
                    e.stopPropagation()
                    setConfirmDialog({ type: 'approve', withdrawal: w })
                  }}
                  disabled={approveMutation.isPending}
                >
                  <CheckCircle2 className="size-4 mr-1" /> Approve
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-600 hover:text-red-700"
                  onClick={(e) => {
                    e.stopPropagation()
                    setConfirmDialog({ type: 'reject', withdrawal: w })
                  }}
                  disabled={rejectMutation.isPending}
                >
                  <XCircle className="size-4 mr-1" /> Reject
                </Button>
              </>
            )}
            {w.status === 'APPROVED' && (
              <Button
                variant="ghost"
                size="sm"
                className="text-blue-600 hover:text-blue-700"
                onClick={(e) => {
                  e.stopPropagation()
                  setConfirmDialog({ type: 'complete', withdrawal: w })
                }}
                disabled={completeMutation.isPending}
              >
                <BanknoteIcon className="size-4 mr-1" /> Complete
              </Button>
            )}
            {w.status === 'COMPLETED' && (
              <span className="text-xs text-muted-foreground font-mono">{w.bankReference}</span>
            )}
          </div>
        )
      },
    },
  ]

  return (
    <>
      <PageHeader
        title="Withdrawal Approvals"
        description="Review and process merchant withdrawal requests"
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Pending" value={String(pending.length)} description="Awaiting approval" icon={Clock} valueClassName={pending.length > 0 ? 'text-amber-500' : ''} />
        <StatCard title="Approved" value={String(approved.length)} description="Ready for bank transfer" icon={ArrowDownToLine} />
        <StatCard title="Total Settled (LKR)" value={totalSettledLKR.toFixed(2)} description="Bank transfers completed" icon={BanknoteIcon} />
      </div>

      <DataTable
        columns={columns}
        data={withdrawals}
        filters={WITHDRAWAL_FILTERS}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search withdrawals..."
        pagination={{ page, perPage: PER_PAGE, total }}
        onPageChange={setPage}
        isLoading={isLoading}
      />

      {/* Confirmation Dialog */}
      <Dialog open={!!confirmDialog} onOpenChange={(v) => !v && setConfirmDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{confirmDialog ? actionLabels[confirmDialog.type]?.title : ''}</DialogTitle>
            <DialogDescription>
              {confirmDialog ? actionLabels[confirmDialog.type]?.desc : ''}
            </DialogDescription>
          </DialogHeader>
          {confirmDialog && (
            <div className="text-sm space-y-1 py-2">
              <p><span className="text-muted-foreground">Amount:</span> {confirmDialog.withdrawal.amountUsdt} USDT</p>
              <p><span className="text-muted-foreground">LKR Amount:</span> {confirmDialog.withdrawal.amountLkr}</p>
              <p><span className="text-muted-foreground">Bank:</span> {confirmDialog.withdrawal.bankName} - {confirmDialog.withdrawal.bankAccountNo}</p>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmDialog(null)}>Cancel</Button>
            <Button
              variant={confirmDialog ? actionLabels[confirmDialog.type]?.variant : 'default'}
              onClick={handleConfirm}
              disabled={approveMutation.isPending || rejectMutation.isPending || completeMutation.isPending}
            >
              {confirmDialog ? actionLabels[confirmDialog.type]?.btn : 'Confirm'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
