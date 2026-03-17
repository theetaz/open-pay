import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { PageHeader } from '#/components/dashboard/page-header'
import { api } from '#/lib/api'

const SERVICE_NAMES = ['gateway', 'payment', 'merchant', 'settlement', 'webhook', 'exchange', 'subscription', 'notification', 'admin'] as const

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

  const bgColor =
    status === 'healthy'
      ? 'bg-green-500/5 border-green-500/20'
      : status === 'unhealthy'
        ? 'bg-red-500/5 border-red-500/20'
        : 'bg-yellow-500/5 border-yellow-500/20'

  return (
    <div className={`flex items-center gap-3 rounded-lg border px-4 py-3 ${bgColor}`}>
      <span className={`h-2.5 w-2.5 rounded-full ${dotColor}`} />
      <div className="flex-1">
        <p className="text-sm font-medium capitalize">{name}</p>
        <p className="text-xs text-muted-foreground">
          {status === 'healthy' ? 'Operational' : status === 'unhealthy' ? 'Down' : 'Checking...'}
        </p>
      </div>
      <span className="text-xs text-muted-foreground">
        {status === 'healthy' && health?.responseTime != null
          ? `${health.responseTime}ms`
          : status === 'unhealthy' ? 'N/A' : '...'}
      </span>
    </div>
  )
}

export function SystemHealthPage() {
  const { data: healthData, dataUpdatedAt } = useQuery({
    queryKey: ['system', 'health'],
    queryFn: () => api.get<{ data: HealthData }>('/v1/system/health'),
    refetchInterval: 15000,
    retry: false,
  })

  const health = healthData?.data || {}
  const healthyCount = Object.values(health).filter((h) => h.status === 'healthy').length
  const totalChecked = Object.keys(health).length

  return (
    <>
      <PageHeader
        title="System Health"
        description="Real-time service status monitoring"
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-3xl font-bold text-green-500">{healthyCount}</p>
              <p className="text-xs text-muted-foreground mt-1">Services Healthy</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <p className={`text-3xl font-bold ${totalChecked - healthyCount > 0 ? 'text-red-500' : 'text-muted-foreground'}`}>
                {totalChecked - healthyCount}
              </p>
              <p className="text-xs text-muted-foreground mt-1">Services Down</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-3xl font-bold">{SERVICE_NAMES.length}</p>
              <p className="text-xs text-muted-foreground mt-1">Total Services</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Service Status</CardTitle>
          <CardDescription>
            Auto-refreshes every 15 seconds
            {dataUpdatedAt && (
              <span className="ml-2">
                — Last checked: {new Date(dataUpdatedAt).toLocaleTimeString()}
              </span>
            )}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {SERVICE_NAMES.map((name) => (
              <ServiceStatus key={name} name={name} health={health[name]} />
            ))}
          </div>
        </CardContent>
      </Card>
    </>
  )
}
