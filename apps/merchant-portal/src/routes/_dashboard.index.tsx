import { createFileRoute, Link } from '@tanstack/react-router'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { StatCard } from '#/components/dashboard/stat-card'
import { DollarSign, CreditCard, Clock, AlertTriangle } from 'lucide-react'
import { usePayments } from '#/hooks/use-payments'
import { useMe } from '#/hooks/use-auth'

export const Route = createFileRoute('/_dashboard/')({
  component: DashboardIndex,
})

function DashboardIndex() {
  const { data: meData } = useMe()
  const { data: paymentsData } = usePayments({ perPage: 5 })

  const payments = paymentsData?.data || []
  const totalPayments = paymentsData?.meta?.total || 0
  const merchant = meData?.data?.merchant

  const paidPayments = payments.filter((p) => p.status === 'PAID')
  const totalRevenue = paidPayments.reduce((sum, p) => sum + parseFloat(p.netAmountUsdt || '0'), 0)
  const unsettledPayments = payments.filter((p) => p.status === 'INITIATED' || p.status === 'USER_REVIEW')

  return (
    <>
      {merchant && merchant.kycStatus === 'PENDING' && (
        <div className="mb-4 rounded-md bg-amber-500/10 border border-amber-500/20 p-4">
          <p className="text-sm text-amber-700 dark:text-amber-400">
            Complete your KYC verification to unlock full payment processing.{' '}
            <Link to="/activate" className="underline font-medium">
              Complete KYC
            </Link>
          </p>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Revenue" value={`${totalRevenue.toFixed(2)} USDT`} description="From paid transactions" icon={DollarSign} />
        <StatCard title="Total Payments" value={String(totalPayments)} description="All-time transactions" icon={CreditCard} />
        <StatCard title="Unsettled Amount" value={`${unsettledPayments.length} pending`} description="Awaiting confirmation" icon={Clock} valueClassName="text-amber-500" />
        <StatCard title="Unsettled Payments" value={String(unsettledPayments.length)} description="Pending transactions" icon={AlertTriangle} valueClassName="text-amber-500" />
      </div>

      <div className="mt-8 grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Your latest payment activity</CardDescription>
          </CardHeader>
          <CardContent>
            {payments.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">
                No recent activity. Payments will appear here once processed.
              </p>
            ) : (
              <div className="space-y-3">
                {payments.map((p) => (
                  <div key={p.id} className="flex items-center justify-between text-sm border-b pb-2 last:border-0">
                    <div>
                      <p className="font-medium">{p.paymentNo}</p>
                      <p className="text-muted-foreground text-xs">{new Date(p.createdAt).toLocaleString()}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-medium">{p.amountUsdt} USDT</p>
                      <span className={`text-xs px-2 py-0.5 rounded-full ${
                        p.status === 'PAID' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                        p.status === 'EXPIRED' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
                        'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                      }`}>
                        {p.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common tasks</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Link to="/payments">
              <Button variant="outline" className="w-full justify-start">View All Payments</Button>
            </Link>
            <Button variant="outline" className="w-full justify-start">Create Payment Link</Button>
            <Button variant="outline" className="w-full justify-start">Generate API Key</Button>
            <Button variant="outline" className="w-full justify-start">Configure Webhook</Button>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
