import { Card, CardContent, CardHeader } from '#/components/ui/card'
import type { LucideIcon } from 'lucide-react'

interface StatCardProps {
  title: string
  value: string | number
  description?: string
  icon?: LucideIcon
  valueClassName?: string
}

export function StatCard({ title, value, description, icon: Icon, valueClassName }: StatCardProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">{title}</p>
          {Icon && <Icon className="size-4 text-muted-foreground" />}
        </div>
      </CardHeader>
      <CardContent>
        <p className={`text-2xl font-bold ${valueClassName ?? ''}`}>{value}</p>
        {description && (
          <p className="text-xs text-muted-foreground mt-1">{description}</p>
        )}
      </CardContent>
    </Card>
  )
}
