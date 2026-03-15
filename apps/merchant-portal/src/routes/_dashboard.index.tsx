import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { StatCard } from '#/components/dashboard/stat-card'
import { DollarSign, CreditCard, Clock, AlertTriangle } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/')({
  component: DashboardIndex,
})

function DashboardIndex() {
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Revenue" value="0.00 USDT" description="₮ 0.00" icon={DollarSign} />
        <StatCard title="Total Payments" value="0" description="All-time transactions" icon={CreditCard} />
        <StatCard title="Unsettled Amount" value="0.00 USDT" description="₮ 0.00" icon={Clock} valueClassName="text-amber-500" />
        <StatCard title="Unsettled Payments" value="0" description="Pending transactions" icon={AlertTriangle} valueClassName="text-amber-500" />
      </div>

      <div className="mt-8 grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Your latest payment activity</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground py-8 text-center">
              No recent activity. Payments will appear here once processed.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common tasks</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button variant="outline" className="w-full justify-start">Create Payment Link</Button>
            <Button variant="outline" className="w-full justify-start">Generate API Key</Button>
            <Button variant="outline" className="w-full justify-start">Configure Webhook</Button>
            <Button variant="outline" className="w-full justify-start">View Documentation</Button>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
