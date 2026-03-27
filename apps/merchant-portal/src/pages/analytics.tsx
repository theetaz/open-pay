import * as React from 'react'
import { PageHeader } from '#/components/dashboard/page-header'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { StatCard } from '#/components/dashboard/stat-card'
import { Button } from '#/components/ui/button'
import { TrendingUp, PieChart, BarChart3, DollarSign } from 'lucide-react'
import { useRevenueAnalytics, useConversionAnalytics, useProviderAnalytics } from '#/hooks/use-analytics'
import { formatAmount } from '#/lib/currency'

export function AnalyticsPage() {
  const [days, setDays] = React.useState(30)

  const { data: revenueData } = useRevenueAnalytics(days)
  const { data: conversionData } = useConversionAnalytics(days)
  const { data: providerData } = useProviderAnalytics(days)

  const revenue = revenueData?.data || []
  const conversion = conversionData?.data
  const providers = providerData?.data || []

  const totalRevenue = revenue.reduce((sum, r) => sum + r.revenue, 0)
  const totalPayments = revenue.reduce((sum, r) => sum + r.count, 0)
  const avgDaily = days > 0 ? totalRevenue / days : 0

  return (
    <>
      <PageHeader
        title="Analytics"
        description="Revenue trends, conversion rates, and provider performance"
        action={
          <div className="flex gap-2">
            {[7, 30, 90].map((d) => (
              <Button
                key={d}
                variant={days === d ? 'default' : 'outline'}
                size="sm"
                onClick={() => setDays(d)}
              >
                {d}d
              </Button>
            ))}
          </div>
        }
      />

      <div className="grid gap-4 md:grid-cols-4 mb-6">
        <StatCard title="Total Revenue" value={formatAmount(totalRevenue, 'USDT')} description={`Last ${days} days`} icon={DollarSign} />
        <StatCard title="Total Payments" value={String(totalPayments)} description={`Last ${days} days`} icon={BarChart3} />
        <StatCard title="Avg Daily Revenue" value={formatAmount(avgDaily, 'USDT')} description="Per day" icon={TrendingUp} />
        <StatCard title="Success Rate" value={conversion ? `${conversion.successRate.toFixed(1)}%` : '—'} description="Paid / total" icon={PieChart} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2 mb-6">
        {/* Revenue Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Daily Revenue (USDT)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {revenue.slice(-14).map((r) => {
                const maxRevenue = Math.max(...revenue.map((x) => x.revenue), 1)
                const width = (r.revenue / maxRevenue) * 100
                return (
                  <div key={r.date} className="flex items-center gap-3 text-sm">
                    <span className="text-muted-foreground w-20 shrink-0">{r.date.slice(5)}</span>
                    <div className="flex-1 bg-muted rounded-full h-5 overflow-hidden">
                      <div
                        className="bg-primary h-full rounded-full transition-all"
                        style={{ width: `${Math.max(width, 1)}%` }}
                      />
                    </div>
                    <span className="font-mono text-xs w-24 text-right">{formatAmount(r.revenue, 'USDT')}</span>
                  </div>
                )
              })}
              {revenue.length === 0 && (
                <p className="text-muted-foreground text-center py-8">No revenue data for this period</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Conversion Funnel */}
        <Card>
          <CardHeader>
            <CardTitle>Payment Conversion Funnel</CardTitle>
          </CardHeader>
          <CardContent>
            {conversion ? (
              <div className="space-y-4">
                {[
                  { label: 'Total Created', value: conversion.total, color: 'bg-blue-500' },
                  { label: 'Paid', value: conversion.paid, color: 'bg-green-500' },
                  { label: 'Expired', value: conversion.expired, color: 'bg-amber-500' },
                  { label: 'Failed', value: conversion.failed, color: 'bg-red-500' },
                  { label: 'Pending', value: conversion.pending, color: 'bg-gray-400' },
                ].map((item) => {
                  const pct = conversion.total > 0 ? (item.value / conversion.total) * 100 : 0
                  return (
                    <div key={item.label} className="space-y-1">
                      <div className="flex justify-between text-sm">
                        <span>{item.label}</span>
                        <span className="font-medium">{item.value} ({pct.toFixed(1)}%)</span>
                      </div>
                      <div className="h-3 bg-muted rounded-full overflow-hidden">
                        <div className={`h-full rounded-full ${item.color}`} style={{ width: `${Math.max(pct, 1)}%` }} />
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <p className="text-muted-foreground text-center py-8">No conversion data</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Provider Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle>Provider Performance</CardTitle>
        </CardHeader>
        <CardContent>
          {providers.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 font-medium">Provider</th>
                    <th className="text-right py-2 font-medium">Payments</th>
                    <th className="text-right py-2 font-medium">Paid</th>
                    <th className="text-right py-2 font-medium">Volume (USDT)</th>
                    <th className="text-right py-2 font-medium">Success Rate</th>
                  </tr>
                </thead>
                <tbody>
                  {providers.map((p) => (
                    <tr key={p.provider} className="border-b last:border-0">
                      <td className="py-3 font-medium">{p.provider}</td>
                      <td className="py-3 text-right">{p.count}</td>
                      <td className="py-3 text-right text-green-600">{p.paid}</td>
                      <td className="py-3 text-right font-mono">{formatAmount(p.volume, 'USDT')}</td>
                      <td className="py-3 text-right">{p.successRate.toFixed(1)}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground text-center py-8">No provider data for this period</p>
          )}
        </CardContent>
      </Card>
    </>
  )
}
