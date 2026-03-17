import { Link } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { StatCard } from '#/components/dashboard/stat-card'
import { DollarSign, CreditCard, Clock, AlertTriangle } from 'lucide-react'
import { usePayments } from '#/hooks/use-payments'
import { useMe } from '#/hooks/use-auth'
import { CreatePaymentDialog } from '#/components/dashboard/create-payment-dialog'
import { formatDualAmount } from '#/lib/currency'

export function DashboardIndex() {
  const { data: meData } = useMe()
  const { data: paymentsData } = usePayments({ perPage: 5 })

  const primaryCurrency = meData?.data?.merchant?.defaultCurrency || 'LKR'
  const payments = paymentsData?.data || []
  const totalPayments = paymentsData?.meta?.total || 0

  const paidPayments = payments.filter((p) => p.status === 'PAID')
  const totalRevenue = paidPayments.reduce((sum, p) => sum + parseFloat(p.netAmountUsdt || '0'), 0)
  const unsettledPayments = payments.filter((p) => p.status === 'INITIATED' || p.status === 'USER_REVIEW')

  const revenueFmt = formatDualAmount(totalRevenue, undefined, undefined, primaryCurrency)

  return (
    <>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Revenue"
          value={revenueFmt.primary}
          description={revenueFmt.secondary ? `${revenueFmt.secondary}` : 'From paid transactions'}
          icon={DollarSign}
        />
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
                {payments.map((p) => {
                  const amt = formatDualAmount(p.amountUsdt, p.amount, p.currency, primaryCurrency, p.exchangeRate)
                  return (
                    <div key={p.id} className="flex items-center justify-between text-sm border-b pb-2 last:border-0">
                      <div>
                        <p className="font-medium">{p.paymentNo}</p>
                        <p className="text-muted-foreground text-xs">{new Date(p.createdAt).toLocaleString()}</p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{amt.primary}</p>
                        {amt.secondary && (
                          <p className="text-xs text-muted-foreground">({amt.secondary})</p>
                        )}
                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                          p.status === 'PAID' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                          p.status === 'EXPIRED' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
                          'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                        }`}>
                          {p.status}
                        </span>
                      </div>
                    </div>
                  )
                })}
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
            <CreatePaymentDialog />
            <Link to="/payments">
              <Button variant="outline" className="w-full justify-start">View All Payments</Button>
            </Link>
            <Link to="/subscriptions">
              <Button variant="outline" className="w-full justify-start">Manage Subscriptions</Button>
            </Link>
            <Link to="/withdrawal">
              <Button variant="outline" className="w-full justify-start">Request Withdrawal</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
