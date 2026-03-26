import { useState, useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { PageHeader } from '#/components/dashboard/page-header'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { DollarSign, CreditCard, Clock, AlertTriangle } from 'lucide-react'
import { usePayments } from '#/hooks/use-payments'
import { useMe } from '#/hooks/use-auth'
import { useExchangeRate } from '#/hooks/use-exchange-rate'
import { formatDualAmount } from '#/lib/currency'

const PER_PAGE = 20

interface Payment {
  id: string
  paymentNo: string
  merchantTradeNo?: string
  createdAt: string
  status: string
  provider: string
  amountUsdt: string
  amount?: string
  currency?: string
  totalFeesUsdt: string
  netAmountUsdt: string
  exchangeRate?: string
}

const filters: FilterConfig[] = [
  {
    id: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { label: 'All', value: '' },
      { label: 'Paid', value: 'PAID' },
      { label: 'Initiated', value: 'INITIATED' },
      { label: 'Failed', value: 'FAILED' },
      { label: 'Expired', value: 'EXPIRED' },
    ],
  },
]

export function PaymentsPage() {
  const [page, setPage] = useState(1)
  const [filterValues, setFilterValues] = useState<Record<string, string | string[]>>({ status: '' })
  const [search, setSearch] = useState('')

  const statusFilter = (filterValues.status as string) || ''

  const { data: meData } = useMe()
  const { data: paymentsData, isLoading } = usePayments({ page, perPage: PER_PAGE, status: statusFilter || undefined })
  const { data: rateData } = useExchangeRate()

  const primaryCurrency = meData?.data?.merchant?.defaultCurrency || 'LKR'
  const liveRate = rateData?.data?.effectiveRate || null
  const payments: Payment[] = paymentsData?.data || []
  const total = paymentsData?.meta?.total || 0

  // Calculate stats from all fetched payments
  const paidPayments = payments.filter((p) => p.status === 'PAID')
  const totalRevenueUsdt = paidPayments.reduce((sum, p) => sum + parseFloat(p.netAmountUsdt || '0'), 0)
  const totalFeesUsdt = paidPayments.reduce((sum, p) => sum + parseFloat(p.totalFeesUsdt || '0'), 0)
  const unsettledPayments = payments.filter((p) => p.status === 'INITIATED' || p.status === 'USER_REVIEW')

  // Sum LKR equivalents from each payment's own exchange rate, fallback to live rate
  const totalRevenueLkr = paidPayments.reduce((sum, p) => {
    const rate = parseFloat(p.exchangeRate || '0') || parseFloat(liveRate || '0')
    return sum + parseFloat(p.netAmountUsdt || '0') * rate
  }, 0)
  const totalFeesLkr = paidPayments.reduce((sum, p) => {
    const rate = parseFloat(p.exchangeRate || '0') || parseFloat(liveRate || '0')
    return sum + parseFloat(p.totalFeesUsdt || '0') * rate
  }, 0)

  const revenueFmt = formatDualAmount(totalRevenueUsdt, totalRevenueLkr > 0 ? totalRevenueLkr : undefined, 'LKR', primaryCurrency, liveRate)
  const feesFmt = formatDualAmount(totalFeesUsdt, totalFeesLkr > 0 ? totalFeesLkr : undefined, 'LKR', primaryCurrency, liveRate)

  const columns: ColumnDef<Payment, any>[] = useMemo(
    () => [
      {
        accessorKey: 'paymentNo',
        header: 'Payment No',
        cell: ({ row }) => (
          <div>
            <p className="font-mono text-sm">{row.original.paymentNo}</p>
            {row.original.merchantTradeNo && (
              <p className="text-xs text-muted-foreground">{row.original.merchantTradeNo}</p>
            )}
          </div>
        ),
      },
      {
        accessorKey: 'createdAt',
        header: 'Date',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {new Date(row.original.createdAt).toLocaleString()}
          </span>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'provider',
        header: 'Provider',
        cell: ({ row }) => <span className="text-sm">{row.original.provider}</span>,
      },
      {
        id: 'amount',
        header: 'Amount',
        cell: ({ row }) => {
          const p = row.original
          const amt = formatDualAmount(p.amountUsdt, p.amount, p.currency, primaryCurrency, p.exchangeRate)
          return (
            <div className="text-right">
              <p className="text-sm font-medium">{amt.primary}</p>
              {amt.secondary && (
                <p className="text-xs text-muted-foreground">({amt.secondary})</p>
              )}
              {p.exchangeRate && (
                <p className="text-xs text-muted-foreground">@ {parseFloat(p.exchangeRate).toFixed(2)}</p>
              )}
            </div>
          )
        },
      },
      {
        id: 'fees',
        header: 'Fees',
        cell: ({ row }) => {
          const p = row.original
          const fee = formatDualAmount(p.totalFeesUsdt, undefined, undefined, primaryCurrency, p.exchangeRate)
          return (
            <div className="text-right">
              <p className="text-sm font-medium">{fee.primary}</p>
              {fee.secondary && (
                <p className="text-xs text-muted-foreground">({fee.secondary})</p>
              )}
            </div>
          )
        },
      },
      {
        id: 'net',
        header: 'Net',
        cell: ({ row }) => {
          const p = row.original
          const net = formatDualAmount(p.netAmountUsdt, undefined, undefined, primaryCurrency, p.exchangeRate)
          return (
            <div className="text-right">
              <p className="text-sm font-medium">{net.primary}</p>
              {net.secondary && (
                <p className="text-xs text-muted-foreground">({net.secondary})</p>
              )}
            </div>
          )
        },
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
      <PageHeader title="Payments" description="View and manage all payment transactions" />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard
          title="Total Revenue"
          value={revenueFmt.primary}
          description={revenueFmt.secondary || 'From paid transactions'}
          icon={DollarSign}
        />
        <StatCard title="Total Payments" value={String(total)} description="All-time transactions" icon={CreditCard} />
        <StatCard
          title="Total Fees"
          value={feesFmt.primary}
          description={feesFmt.secondary || 'Exchange + Platform'}
          icon={Clock}
          valueClassName="text-amber-500"
        />
        <StatCard
          title="Unsettled"
          value={String(unsettledPayments.length)}
          description="Pending transactions"
          icon={AlertTriangle}
          valueClassName="text-amber-500"
        />
      </div>

      <DataTable
        columns={columns}
        data={payments}
        filters={filters}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={handleSearchChange}
        searchPlaceholder="Search payments..."
        pagination={{ page, perPage: PER_PAGE, total }}
        onPageChange={setPage}
        isLoading={isLoading}
      />
    </>
  )
}
