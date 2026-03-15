import type { ReactNode } from 'react'

interface EmptyStateProps {
  message: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ message, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <p className="text-sm font-medium text-muted-foreground">{message}</p>
      {description && (
        <p className="text-xs text-muted-foreground mt-1 max-w-sm">{description}</p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
