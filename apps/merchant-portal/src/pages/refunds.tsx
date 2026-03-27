import * as React from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Plus, RotateCcw, DollarSign, Loader2, Clock } from 'lucide-react'
import { useRefunds, useRequestRefund, type Refund } from '#/hooks/use-refunds'
import { formatAmount } from '#/lib/currency'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Textarea } from '#/components/ui/textarea'
import type { ColumnDef } from '@tanstack/react-table'

export function RefundsPage() {
  const { data: refundsData } = useRefunds()
  const [search, setSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({ status: '' })

  const refunds = refundsData?.data || []
  const pending = refunds.filter((r) => r.status === 'PENDING')
  const totalRefunded = refunds
    .filter((r) => r.status === 'COMPLETED')
    .reduce((sum, r) => sum + parseFloat(r.amountUsdt || '0'), 0)

  const filters: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { label: 'Pending', value: 'PENDING' },
        { label: 'Completed', value: 'COMPLETED' },
        { label: 'Rejected', value: 'REJECTED' },
      ],
    },
  ]

  const filtered = React.useMemo(() => {
    let result = refunds
    if (search) {
      const q = search.toLowerCase()
      result = result.filter(
        (r) => r.paymentNo.toLowerCase().includes(q) || r.reason.toLowerCase().includes(q),
      )
    }
    const status = filterValues.status
    if (status && typeof status === 'string' && status !== '') {
      result = result.filter((r) => r.status === status)
    }
    return result
  }, [refunds, search, filterValues])

  const columns: ColumnDef<Refund, any>[] = React.useMemo(
    () => [
      {
        accessorKey: 'paymentNo',
        header: 'Payment',
        cell: ({ row }) => <span className="font-mono text-sm">{row.original.paymentNo}</span>,
      },
      {
        accessorKey: 'amountUsdt',
        header: 'Amount',
        cell: ({ row }) => (
          <span className="font-medium">{formatAmount(row.original.amountUsdt, 'USDT')}</span>
        ),
      },
      {
        accessorKey: 'reason',
        header: 'Reason',
        cell: ({ row }) => (
          <span className="text-sm max-w-[200px] truncate block">{row.original.reason}</span>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'createdAt',
        header: 'Requested',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {new Date(row.original.createdAt).toLocaleDateString()}
          </span>
        ),
      },
    ],
    [],
  )

  return (
    <>
      <PageHeader
        title="Refunds"
        description="Manage refund requests for completed payments"
        action={<CreateRefundDialog />}
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Total Refunds" value={String(refunds.length)} description="All refund requests" icon={RotateCcw} />
        <StatCard title="Pending" value={String(pending.length)} description="Awaiting approval" icon={Clock} />
        <StatCard title="Total Refunded" value={formatAmount(totalRefunded, 'USDT')} description="Completed refunds" icon={DollarSign} />
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        filters={filters}
        filterValues={filterValues}
        onFilterChange={(id, value) => setFilterValues((prev) => ({ ...prev, [id]: value }))}
        onClearFilters={() => setFilterValues({ status: '' })}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search refunds..."
        pagination={{ page: 1, perPage: 999, total: filtered.length }}
        onPageChange={() => {}}
      />
    </>
  )
}

function CreateRefundDialog() {
  const [open, setOpen] = React.useState(false)
  const [paymentId, setPaymentId] = React.useState('')
  const [paymentNo, setPaymentNo] = React.useState('')
  const [amount, setAmount] = React.useState('')
  const [reason, setReason] = React.useState('')

  const requestRefund = useRequestRefund()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    requestRefund.mutate(
      { paymentId, paymentNo, amountUsdt: amount, reason },
      {
        onSuccess: () => {
          setOpen(false)
          setPaymentId('')
          setPaymentNo('')
          setAmount('')
          setReason('')
        },
      },
    )
  }

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="mr-2 size-4" /> Request Refund
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Request Refund</DialogTitle>
              <DialogDescription>Issue a refund for a completed payment. The amount will be debited from your balance.</DialogDescription>
            </DialogHeader>

            <div className="py-4">
              <FieldGroup>
                {requestRefund.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {requestRefund.error.message}
                  </div>
                )}

                <Field>
                  <FieldLabel>Payment ID</FieldLabel>
                  <Input value={paymentId} onChange={(e) => setPaymentId(e.target.value)} placeholder="UUID of the payment" required />
                </Field>

                <Field>
                  <FieldLabel>Payment Number</FieldLabel>
                  <Input value={paymentNo} onChange={(e) => setPaymentNo(e.target.value)} placeholder="PAY-xxxx" required />
                </Field>

                <Field>
                  <FieldLabel>Refund Amount (USDT)</FieldLabel>
                  <Input type="number" step="0.01" min="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="25.00" required />
                </Field>

                <Field>
                  <FieldLabel>Reason</FieldLabel>
                  <Textarea value={reason} onChange={(e) => setReason(e.target.value)} placeholder="Describe why this refund is needed..." required />
                </Field>
              </FieldGroup>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={requestRefund.isPending}>
                {requestRefund.isPending ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Submitting...</> : 'Request Refund'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}
