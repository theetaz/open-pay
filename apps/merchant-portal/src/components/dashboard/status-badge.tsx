import { Badge } from '#/components/ui/badge'

type StatusVariant = 'success' | 'warning' | 'error' | 'info' | 'default' | 'purple'

const variantClasses: Record<StatusVariant, string> = {
  success: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-transparent',
  warning: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 border-transparent',
  error: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 border-transparent',
  info: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border-transparent',
  purple: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 border-transparent',
  default: '',
}

const statusMap: Record<string, StatusVariant> = {
  PAID: 'success', SUCCESS: 'success', ACTIVE: 'success', APPROVED: 'success', DELIVERED: 'success',
  INITIATED: 'warning', PENDING: 'warning', UNDER_REVIEW: 'warning', REQUESTED: 'warning', TRIAL: 'warning',
  FAILED: 'error', EXPIRED: 'error', REJECTED: 'error', EXHAUSTED: 'error', PAST_DUE: 'error',
  CREATE: 'purple',
  UPDATE: 'success',
  Admin: 'error',
  Manager: 'info',
  User: 'default',
}

interface StatusBadgeProps {
  status: string
  variant?: StatusVariant
}

export function StatusBadge({ status, variant }: StatusBadgeProps) {
  const v = variant ?? statusMap[status] ?? 'default'
  return (
    <Badge variant="outline" className={variantClasses[v]}>
      {status}
    </Badge>
  )
}
