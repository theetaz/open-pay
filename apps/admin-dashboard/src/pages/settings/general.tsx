import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { Loader2 } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

export function SettingsGeneralPage() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'settings', 'general'],
    queryFn: () => api.get<{ data: Array<{ key: string; value: string; description: string }> }>('/v1/admin/settings/general'),
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
      toast.success('Settings saved')
    },
  })

  if (isLoading) return <div className="text-muted-foreground text-sm">Loading...</div>

  return (
    <>
      <PageHeader title="General Settings" description="Platform name, branding, and contact information" />
      <Card>
        <CardHeader>
          <CardTitle>Platform Identity</CardTitle>
          <CardDescription>Configure how your platform appears to merchants</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel>Platform Name</FieldLabel>
              <Input value={form.platform_name || ''} onChange={(e) => update('platform_name', e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Logo URL</FieldLabel>
              <Input value={form.platform_logo_url || ''} onChange={(e) => update('platform_logo_url', e.target.value)} placeholder="https://..." />
            </Field>
            <Field>
              <FieldLabel>Support Email</FieldLabel>
              <Input type="email" value={form.support_email || ''} onChange={(e) => update('support_email', e.target.value)} />
            </Field>
            <Button onClick={() => mutation.mutate(form)} disabled={mutation.isPending} className="w-fit">
              {mutation.isPending ? <><Loader2 className="mr-2 size-4 animate-spin" />Saving...</> : 'Save Changes'}
            </Button>
          </FieldGroup>
        </CardContent>
      </Card>
    </>
  )
}
