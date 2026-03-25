import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Landmark, TrendingUp, ArrowDownToLine, Banknote } from 'lucide-react'
import { api } from '#/lib/api'

interface BalanceData {
  data: {
    availableUsdt?: string
    pendingUsdt?: string
    totalEarnedUsdt?: string
    totalWithdrawnUsdt?: string
    totalFeesUsdt?: string
    totalEarnedLkr?: string
    totalWithdrawnLkr?: string
  }
}

export function TreasuryPage() {
  const { data: rateData } = useQuery({
    queryKey: ['admin', 'exchange-rate'],
    queryFn: () => api.get<{ data: { rate: string; source: string; fetchedAt: string } }>('/v1/exchange-rates/active'),
    retry: false,
  })

  const { data: balanceData } = useQuery({
    queryKey: ['admin', 'settlements-balance'],
    queryFn: () => api.get<BalanceData>('/v1/admin/settlements/balance'),
    retry: false,
  })

  const { data: configData } = useQuery({
    queryKey: ['platform', 'config'],
    queryFn: () => api.get<{ data: { platformFeePct: string; exchangeFeePct: string } }>('/v1/platform/config'),
    retry: false,
  })

  const rate = rateData?.data
  const balance = balanceData?.data
  const feeConfig = configData?.data

  const availableUsdt = balance?.availableUsdt ?? '0.00'
  const totalFeesUsdt = balance?.totalFeesUsdt ?? '0.00'
  const totalWithdrawnLkr = balance?.totalWithdrawnLkr ?? '0.00'
  const pendingUsdt = balance?.pendingUsdt ?? '0.00'

  const rateValue = rate?.rate ? parseFloat(rate.rate) : null
  const lkrFor100Usdt = rateValue ? (100 * rateValue).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : null

  return (
    <>
      <PageHeader
        title="Treasury"
        description="Platform financial overview and exchange rates"
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <StatCard title="Crypto Holdings" value={`${availableUsdt} USDT`} description="Platform balance" icon={Landmark} />
        <StatCard title="Fees Earned" value={`${totalFeesUsdt} USDT`} description="Platform + exchange fees" icon={TrendingUp} />
        <StatCard title="Total Settled" value={`${totalWithdrawnLkr} LKR`} description="Bank transfers completed" icon={ArrowDownToLine} />
        <StatCard title="Pending Settlement" value={`${pendingUsdt} USDT`} description="Awaiting settlement" icon={Banknote} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2 mb-6">
        <Card>
          <CardHeader>
            <CardTitle>Current Exchange Rate</CardTitle>
            <CardDescription>USDT / LKR</CardDescription>
          </CardHeader>
          <CardContent>
            {rate ? (
              <div className="space-y-2">
                <p className="text-3xl font-bold">1 USDT = {rate.rate} LKR</p>
                {lkrFor100Usdt && (
                  <p className="text-lg text-muted-foreground">100 USDT = {lkrFor100Usdt} LKR</p>
                )}
                <div className="text-sm text-muted-foreground space-y-1">
                  <p>Source: {rate.source}</p>
                  <p>Last updated: {new Date(rate.fetchedAt).toLocaleString()}</p>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Unable to fetch exchange rate</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Fee Structure</CardTitle>
            <CardDescription>Current platform fee configuration</CardDescription>
          </CardHeader>
          <CardContent>
            {feeConfig ? (
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Platform Fee</span>
                  <span className="font-medium">{feeConfig.platformFeePct}%</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Exchange Fee</span>
                  <span className="font-medium">{feeConfig.exchangeFeePct}%</span>
                </div>
                <div className="flex justify-between text-sm border-t pt-2">
                  <span className="text-muted-foreground">Total Fee</span>
                  <span className="font-bold">{(parseFloat(feeConfig.platformFeePct) + parseFloat(feeConfig.exchangeFeePct)).toFixed(1)}%</span>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Loading fee configuration...</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Treasury Transactions</CardTitle>
          <CardDescription>Platform-level financial events</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Description</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={4}>
                  <EmptyState message="No treasury transactions yet." description="Transactions will appear as payments are processed and fees are collected." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
