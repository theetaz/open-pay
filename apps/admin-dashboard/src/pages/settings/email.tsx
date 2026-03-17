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

export function SettingsEmailPage() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'settings', 'email'],
    queryFn: () => api.get<{ data: Array<{ key: string; value: string; description: string }> }>('/v1/admin/settings/email'),
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
      toast.success('Email settings saved')
    },
  })

  if (isLoading) return <div className="text-muted-foreground text-sm">Loading...</div>

  return (
    <>
      <PageHeader title="Email Configuration" description="SMTP server and sender settings" />
      <Card>
        <CardHeader>
          <CardTitle>SMTP Settings</CardTitle>
          <CardDescription>Configure the mail server used for sending notifications</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field>
                <FieldLabel>SMTP Host</FieldLabel>
                <Input value={form.smtp_host || ''} onChange={(e) => update('smtp_host', e.target.value)} />
              </Field>
              <Field>
                <FieldLabel>SMTP Port</FieldLabel>
                <Input value={form.smtp_port || ''} onChange={(e) => update('smtp_port', e.target.value)} />
              </Field>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Sender Name</FieldLabel>
                <Input value={form.smtp_sender_name || ''} onChange={(e) => update('smtp_sender_name', e.target.value)} />
                <FieldDescription>Display name in outgoing emails</FieldDescription>
              </Field>
              <Field>
                <FieldLabel>Sender Email</FieldLabel>
                <Input type="email" value={form.smtp_sender_email || ''} onChange={(e) => update('smtp_sender_email', e.target.value)} />
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
