import { useState } from 'react'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { PageHeader } from '#/components/dashboard/page-header'
import { DollarSign, CreditCard, Clock, AlertTriangle, ChevronLeft, ChevronRight } from 'lucide-react'
import { usePayments } from '#/hooks/use-payments'
import { useMe } from '#/hooks/use-auth'
import { useExchangeRate } from '#/hooks/use-exchange-rate'
import { formatDualAmount } from '#/lib/currency'

const PER_PAGE = 20

export function PaymentsPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const { data: meData } = useMe()
  const { data: paymentsData, isLoading } = usePayments({ page, perPage: PER_PAGE, status: statusFilter || undefined })

  const { data: rateData } = useExchangeRate()
  const primaryCurrency = meData?.data?.merchant?.defaultCurrency || 'LKR'
  const liveRate = rateData?.data?.effectiveRate || null
  const payments = paymentsData?.data || []
  const total = paymentsData?.meta?.total || 0
  const totalPages = Math.ceil(total / PER_PAGE)

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

  const statuses = [
    { label: 'All', value: '' },
    { label: 'Paid', value: 'PAID' },
    { label: 'Initiated', value: 'INITIATED' },
    { label: 'Failed', value: 'FAILED' },
    { label: 'Expired', value: 'EXPIRED' },
  ]

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

      {/* Status filter tabs */}
      <div className="flex gap-1 mb-4">
        {statuses.map((s) => (
          <Button
            key={s.value}
            variant={statusFilter === s.value ? 'default' : 'outline'}
            size="sm"
            onClick={() => { setStatusFilter(s.value); setPage(1) }}
          >
            {s.label}
          </Button>
        ))}
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Payment No</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead className="text-right">Fees</TableHead>
                <TableHead className="text-right">Net</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {payments.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7}>
                    <EmptyState
                      message={isLoading ? 'Loading payments...' : 'No payments found.'}
                      description={!isLoading ? 'Payments will appear here once transactions are processed.' : undefined}
                    />
                  </TableCell>
                </TableRow>
              ) : (
                payments.map((p) => {
                  const amt = formatDualAmount(p.amountUsdt, p.amount, p.currency, primaryCurrency, p.exchangeRate)
                  const fee = formatDualAmount(p.totalFeesUsdt, undefined, undefined, primaryCurrency, p.exchangeRate)
                  const net = formatDualAmount(p.netAmountUsdt, undefined, undefined, primaryCurrency, p.exchangeRate)

                  return (
                    <TableRow key={p.id}>
                      <TableCell>
                        <p className="font-mono text-sm">{p.paymentNo}</p>
                        {p.merchantTradeNo && (
                          <p className="text-xs text-muted-foreground">{p.merchantTradeNo}</p>
                        )}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                        {new Date(p.createdAt).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={p.status} />
                      </TableCell>
                      <TableCell className="text-sm">{p.provider}</TableCell>
                      <TableCell className="text-right">
                        <p className="text-sm font-medium">{amt.primary}</p>
                        {amt.secondary && (
                          <p className="text-xs text-muted-foreground">({amt.secondary})</p>
                        )}
                        {p.exchangeRate && (
                          <p className="text-xs text-muted-foreground">@ {parseFloat(p.exchangeRate).toFixed(2)}</p>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <p className="text-sm font-medium">{fee.primary}</p>
                        {fee.secondary && (
                          <p className="text-xs text-muted-foreground">({fee.secondary})</p>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <p className="text-sm font-medium">{net.primary}</p>
                        {net.secondary && (
                          <p className="text-xs text-muted-foreground">({net.secondary})</p>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>

          {totalPages > 1 && (
            <div className="flex items-center justify-between border-t px-4 py-3">
              <p className="text-sm text-muted-foreground">
                Showing {(page - 1) * PER_PAGE + 1}–{Math.min(page * PER_PAGE, total)} of {total}
              </p>
              <div className="flex items-center gap-1">
                <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
                  <ChevronLeft className="size-4" />
                </Button>
                <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages}>
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
