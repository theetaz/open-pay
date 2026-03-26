import * as React from 'react'
import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { StatCard } from '#/components/dashboard/stat-card'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Plus, Loader2, DollarSign, Clock, ArrowDownToLine } from 'lucide-react'
import { useBalance, useWithdrawals, useRequestWithdrawal } from '#/hooks/use-settlements'
import { useMe } from '#/hooks/use-auth'
import { useExchangeRate } from '#/hooks/use-exchange-rate'
import { formatDualAmount, formatAmount } from '#/lib/currency'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'

interface Withdrawal {
  id: string
  createdAt: string
  amountUsdt: string
  amountLkr: string
  exchangeRate: string
  bankName: string
  status: string
  bankReference?: string
}

const statusFilters: FilterConfig[] = [
  {
    id: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { label: 'All', value: '' },
      { label: 'Requested', value: 'REQUESTED' },
      { label: 'Approved', value: 'APPROVED' },
      { label: 'Completed', value: 'COMPLETED' },
      { label: 'Rejected', value: 'REJECTED' },
    ],
  },
]

export function WithdrawalPage() {
  const [page, setPage] = React.useState(1)
  const [search, setSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({ status: '' })

  const statusFilter = (filterValues.status as string) || ''

  const { data: balanceData } = useBalance()
  const { data: withdrawalsData } = useWithdrawals()
  const { data: meData } = useMe()
  const { data: rateData } = useExchangeRate()

  const balance = balanceData?.data
  const allWithdrawals: Withdrawal[] = withdrawalsData?.data || []
  const merchant = meData?.data?.merchant
  const primaryCurrency = merchant?.defaultCurrency || 'LKR'
  const liveRate = rateData?.data?.effectiveRate || (allWithdrawals.length > 0 ? allWithdrawals[0].exchangeRate : null)

  // Filter withdrawals client-side
  const filteredWithdrawals = React.useMemo(() => {
    let result = allWithdrawals
    if (statusFilter) {
      result = result.filter((w) => w.status === statusFilter)
    }
    if (search) {
      const q = search.toLowerCase()
      result = result.filter(
        (w) =>
          w.bankName?.toLowerCase().includes(q) ||
          w.bankReference?.toLowerCase().includes(q) ||
          w.status.toLowerCase().includes(q),
      )
    }
    return result
  }, [allWithdrawals, statusFilter, search])

  const PER_PAGE = 20
  const total = filteredWithdrawals.length
  const paginatedWithdrawals = filteredWithdrawals.slice((page - 1) * PER_PAGE, page * PER_PAGE)

  const pending = allWithdrawals.filter((w) => w.status === 'REQUESTED')
  const completed = allWithdrawals.filter((w) => w.status === 'COMPLETED')
  const totalWithdrawn = completed.reduce((sum, w) => sum + parseFloat(w.amountLkr), 0)

  const availableFmt = formatDualAmount(balance?.availableUsdt || '0', undefined, undefined, primaryCurrency, liveRate)
  const pendingFmt = formatDualAmount(balance?.pendingUsdt || '0', undefined, undefined, primaryCurrency, liveRate)
  const earnedFmt = formatDualAmount(balance?.totalEarnedUsdt || '0', undefined, undefined, primaryCurrency, liveRate)

  const columns: ColumnDef<Withdrawal, any>[] = useMemo(
    () => [
      {
        accessorKey: 'createdAt',
        header: 'Date',
        cell: ({ row }) => (
          <span className="text-sm">{new Date(row.original.createdAt).toLocaleDateString()}</span>
        ),
      },
      {
        id: 'amount',
        header: 'Amount',
        cell: ({ row }) => {
          const w = row.original
          const wAmt = formatDualAmount(w.amountUsdt, w.amountLkr, 'LKR', primaryCurrency, w.exchangeRate)
          return (
            <div>
              <p className="font-medium">{wAmt.primary}</p>
              {wAmt.secondary && (
                <p className="text-xs text-muted-foreground">({wAmt.secondary})</p>
              )}
            </div>
          )
        },
      },
      {
        accessorKey: 'exchangeRate',
        header: 'Rate',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            @ {parseFloat(row.original.exchangeRate).toFixed(2)}
          </span>
        ),
      },
      {
        accessorKey: 'bankName',
        header: 'Bank',
        cell: ({ row }) => <span className="text-sm">{row.original.bankName}</span>,
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => {
          const status = row.original.status
          return (
            <span
              className={`text-xs px-2 py-0.5 rounded-full ${
                status === 'COMPLETED'
                  ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                  : status === 'REJECTED'
                    ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                    : status === 'APPROVED'
                      ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                      : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
              }`}
            >
              {status}
            </span>
          )
        },
      },
      {
        accessorKey: 'bankReference',
        header: 'Reference',
        cell: ({ row }) => (
          <span className="text-sm font-mono">{row.original.bankReference || '-'}</span>
        ),
      },
    ],
    [primaryCurrency],
  )

  function handleFilterChange(id: string, value: string | string[]) {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
    setPage(1)
  }

  function handleClearFilters() {
    setFilterValues({ status: '' })
    setPage(1)
  }

  function handleSearchChange(value: string) {
    setSearch(value)
    setPage(1)
  }

  return (
    <>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Withdrawals</h1>
          <p className="text-sm text-muted-foreground">Manage your balance and request withdrawals</p>
        </div>
        <WithdrawDialog
          availableUsdt={balance?.availableUsdt || '0'}
          bankName={merchant?.bankName as string || ''}
          bankAccountNo={merchant?.bankAccountNo as string || ''}
          bankAccountName={merchant?.bankAccountName as string || ''}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Available Balance" value={availableFmt.primary} description={availableFmt.secondary} icon={DollarSign} />
        <StatCard title="Pending Withdrawal" value={pendingFmt.primary} description={pendingFmt.secondary} icon={Clock} valueClassName="text-amber-500" />
        <StatCard title="Total Earned" value={earnedFmt.primary} description={earnedFmt.secondary} icon={ArrowDownToLine} />
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground mb-2">Status Overview</p>
            <div className="space-y-1 text-sm">
              <div className="flex justify-between">
                <span>Completed:</span>
                <span className="font-medium text-green-500">{completed.length}</span>
              </div>
              <div className="flex justify-between">
                <span>Pending:</span>
                <span className="font-medium text-amber-500">{pending.length}</span>
              </div>
              <div className="flex justify-between">
                <span>Total Withdrawn:</span>
                <span className="font-medium">{formatAmount(totalWithdrawn, 'LKR')}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <DataTable
        columns={columns}
        data={paginatedWithdrawals}
        filters={statusFilters}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={handleSearchChange}
        searchPlaceholder="Search withdrawals..."
        pagination={{ page, perPage: PER_PAGE, total }}
        onPageChange={setPage}
      />
    </>
  )
}

function WithdrawDialog({ availableUsdt, bankName, bankAccountNo, bankAccountName }: {
  availableUsdt: string
  bankName: string
  bankAccountNo: string
  bankAccountName: string
}) {
  const [open, setOpen] = React.useState(false)
  const [amount, setAmount] = React.useState('')
  const requestWithdrawal = useRequestWithdrawal()

  const exchangeRate = '325'

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    requestWithdrawal.mutate({
      amountUsdt: amount,
      exchangeRate,
      bankName: bankName || 'Commercial Bank',
      bankAccountNo: bankAccountNo || '0000000000',
      bankAccountName: bankAccountName || 'Account Holder',
    }, {
      onSuccess: () => {
        setOpen(false)
        setAmount('')
      },
    })
  }

  const lkrAmount = (parseFloat(amount || '0') * parseFloat(exchangeRate)).toFixed(2)

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90 h-9 px-4 py-2 cursor-pointer"
      >
        <Plus className="size-4" /> Request Withdrawal
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Request Withdrawal</DialogTitle>
            <DialogDescription>
              Available balance: {formatAmount(parseFloat(availableUsdt), 'USDT')}
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <FieldGroup>
              {requestWithdrawal.isError && (
                <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                  {requestWithdrawal.error.message}
                </div>
              )}

              <Field>
                <FieldLabel>Amount (USDT)</FieldLabel>
                <Input
                  type="number"
                  step="0.01"
                  placeholder="Enter amount"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </Field>

              <div className="rounded-md bg-muted/50 p-3 space-y-1 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Exchange Rate</span>
                  <span>1 USDT = {exchangeRate} LKR</span>
                </div>
                <div className="flex justify-between font-medium">
                  <span>You receive</span>
                  <span>{lkrAmount} LKR</span>
                </div>
              </div>

              <div className="rounded-md bg-muted/50 p-3 space-y-1 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Bank</span>
                  <span>{bankName || 'Not configured'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Account</span>
                  <span>{bankAccountNo || 'Not configured'}</span>
                </div>
              </div>
            </FieldGroup>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button type="submit" disabled={requestWithdrawal.isPending || !amount}>
              {requestWithdrawal.isPending ? (
                <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Processing...</>
              ) : (
                'Confirm Withdrawal'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
