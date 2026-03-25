import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { Building2, ArrowDownToLine, CreditCard, TrendingUp } from 'lucide-react'
import { api } from '#/lib/api'

const SERVICE_NAMES = ['gateway', 'payment', 'merchant', 'settlement', 'webhook', 'exchange'] as const

interface ServiceHealth {
  status: 'healthy' | 'unhealthy'
  responseTime?: number
  error?: string
}

type HealthData = Record<string, ServiceHealth>

function ServiceStatus({ name, health }: { name: string; health?: ServiceHealth }) {
  const status = health?.status ?? 'checking'

  const dotColor =
    status === 'healthy'
      ? 'bg-green-500'
      : status === 'unhealthy'
        ? 'bg-red-500'
        : 'bg-yellow-500 animate-pulse'

  return (
    <div className="flex items-center gap-2 rounded-md border border-border px-3 py-2">
      <span className={`h-2 w-2 rounded-full ${dotColor}`} />
      <span className="text-sm font-medium capitalize">{name}</span>
      <span className="text-xs text-muted-foreground ml-auto">
        {status === 'healthy' && health?.responseTime != null
          ? `${health.responseTime}ms`
          : status === 'unhealthy' ? 'down' : '...'}
      </span>
    </div>
  )
}

export function DashboardIndex() {
  const { data: merchantsData } = useQuery({
    queryKey: ['admin', 'merchants', 'dashboard'],
    queryFn: () => api.get<{ data: any[]; meta: { total: number } }>('/v1/admin/merchants?perPage=5'),
    retry: false,
  })

  const { data: healthData } = useQuery({
    queryKey: ['system', 'health'],
    queryFn: () => api.get<{ data: HealthData }>('/v1/system/health'),
    refetchInterval: 30000,
    retry: false,
  })

  const health = healthData?.data || {}

  const { data: withdrawalsData } = useQuery({
    queryKey: ['admin', 'withdrawals'],
    queryFn: () => api.get<{ data: any[] }>('/v1/admin/withdrawals?status=REQUESTED'),
    retry: false,
  })

  const merchants = merchantsData?.data || []
  const totalMerchants = merchantsData?.meta?.total || merchants.length
  const pendingKYC = merchants.filter((m: any) => m.kycStatus === 'UNDER_REVIEW' || m.kycStatus === 'INSTANT_ACCESS')
  const pendingWithdrawals = withdrawalsData?.data || []

  return (
    <>
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-sm text-muted-foreground">Platform overview and pending actions</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Merchants" value={String(totalMerchants)} description="Registered" icon={Building2} />
        <StatCard title="Pending KYC" value={String(pendingKYC.length)} description="Awaiting review" icon={CreditCard} valueClassName={pendingKYC.length > 0 ? 'text-amber-500' : ''} />
        <StatCard title="Today's Volume" value="0.00 USDT" description="Payments processed" icon={TrendingUp} />
        <StatCard title="Pending Withdrawals" value={String(pendingWithdrawals.length)} description="Needs approval" icon={ArrowDownToLine} valueClassName={pendingWithdrawals.length > 0 ? 'text-amber-500' : ''} />
      </div>

      <div className="mt-8 grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Merchant Approval Queue</CardTitle>
            <CardDescription>Merchants awaiting KYC review</CardDescription>
          </CardHeader>
          <CardContent>
            {pendingKYC.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4 text-center">No merchants pending approval.</p>
            ) : (
              <div className="space-y-3">
                {pendingKYC.map((m: any) => (
                  <div key={m.id} className="flex items-center justify-between border-b pb-2 last:border-0">
                    <div>
                      <p className="text-sm font-medium">{m.businessName}</p>
                      <p className="text-xs text-muted-foreground">{m.contactEmail}</p>
                    </div>
                    <StatusBadge status={m.kycStatus} />
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Withdrawal Approval Queue</CardTitle>
            <CardDescription>Withdrawals awaiting admin approval</CardDescription>
          </CardHeader>
          <CardContent>
            {pendingWithdrawals.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4 text-center">No withdrawals pending approval.</p>
            ) : (
              <div className="space-y-3">
                {pendingWithdrawals.map((w: any) => (
                  <div key={w.id} className="flex items-center justify-between border-b pb-2 last:border-0">
                    <div>
                      <p className="text-sm font-medium">{w.amountUsdt} USDT</p>
                      <p className="text-xs text-muted-foreground">{w.bankName} - {w.bankAccountNo}</p>
                    </div>
                    <StatusBadge status={w.status} />
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card id="system-health" className="mt-6">
        <CardHeader>
          <CardTitle>System Health</CardTitle>
          <CardDescription>Service status overview</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-3">
            {SERVICE_NAMES.map((name) => (
              <ServiceStatus key={name} name={name} health={health[name]} />
            ))}
          </div>
        </CardContent>
      </Card>
    </>
  )
}
