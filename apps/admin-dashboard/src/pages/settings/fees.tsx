import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel, FieldDescription } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { Loader2 } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

export function SettingsFeesPage() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'settings', 'fees'],
    queryFn: () => api.get<{ data: Array<{ key: string; value: string; description: string }> }>('/v1/admin/settings/fees'),
  })

  const settings = React.useMemo(() => {
    const map: Record<string, string> = {}
    for (const s of data?.data || []) map[s.key] = s.value
    return map
  }, [data])

  const [form, setForm] = React.useState<Record<string, string>>({})
  React.useEffect(() => { if (Object.keys(settings).length) setForm(settings) }, [settings])

  const update = (key: string, value: string) => setForm((p) => ({ ...p, [key]: value }))

  const mutation = useMutation({
    mutationFn: (s: Record<string, string>) => api.put('/v1/admin/settings', { settings: s }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] })
      toast.success('Fee settings saved')
    },
  })

  if (isLoading) return <div className="text-muted-foreground text-sm">Loading...</div>

  return (
    <>
      <PageHeader title="Fees & Pricing" description="Configure transaction fees and pricing" />
      <Card>
        <CardHeader>
          <CardTitle>Fee Configuration</CardTitle>
          <CardDescription>These fees are applied to every transaction processed through the platform</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Platform Fee (%)</FieldLabel>
                <Input type="number" step="0.01" value={form.platform_fee_pct || ''} onChange={(e) => update('platform_fee_pct', e.target.value)} />
                <FieldDescription>Percentage charged on each transaction</FieldDescription>
              </Field>
              <Field>
                <FieldLabel>Exchange Fee (%)</FieldLabel>
                <Input type="number" step="0.01" value={form.exchange_fee_pct || ''} onChange={(e) => update('exchange_fee_pct', e.target.value)} />
                <FieldDescription>Fee for currency conversion</FieldDescription>
              </Field>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Min Withdrawal (USDT)</FieldLabel>
                <Input type="number" step="1" value={form.min_withdrawal_usdt || ''} onChange={(e) => update('min_withdrawal_usdt', e.target.value)} />
              </Field>
              <Field>
                <FieldLabel>Settlement Period (days)</FieldLabel>
                <Input type="number" step="1" value={form.settlement_period_days || ''} onChange={(e) => update('settlement_period_days', e.target.value)} />
              </Field>
            </div>
            <Button onClick={() => mutation.mutate(form)} disabled={mutation.isPending} className="w-fit">
              {mutation.isPending ? <><Loader2 className="mr-2 size-4 animate-spin" />Saving...</> : 'Save Changes'}
            </Button>
          </FieldGroup>
        </CardContent>
      </Card>
    </>
  )
}
