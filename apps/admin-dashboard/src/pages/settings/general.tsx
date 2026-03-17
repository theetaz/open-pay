import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel, FieldDescription } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { Upload, Loader2, X, ImageIcon } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

export function SettingsGeneralPage() {
  const queryClient = useQueryClient()
  const fileInputRef = React.useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = React.useState(false)

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

  const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploading(true)
    try {
      const result = await api.upload<{ data: { url: string; key: string; filename: string } }>(file, 'branding')
      update('platform_logo_url', result.data.url)
      toast.success('Logo uploaded')
    } catch {
      toast.error('Failed to upload logo')
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const removeLogo = () => {
    update('platform_logo_url', '')
  }

  if (isLoading) return <div className="text-muted-foreground text-sm">Loading...</div>

  const logoUrl = form.platform_logo_url

  return (
    <>
      <PageHeader title="General Settings" description="Platform name, branding, and contact information" />

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Platform Branding</CardTitle>
            <CardDescription>Upload your platform logo and set the display name</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel>Platform Logo</FieldLabel>
                <div className="flex items-start gap-4">
                  {/* Logo preview */}
                  <div className="shrink-0">
                    {logoUrl ? (
                      <div className="relative group">
                        <div className="size-20 rounded-lg border bg-muted/50 flex items-center justify-center overflow-hidden">
                          <img
                            src={logoUrl}
                            alt="Platform logo"
                            className="max-w-full max-h-full object-contain"
                            onError={(e) => {
                              (e.target as HTMLImageElement).style.display = 'none'
                            }}
                          />
                        </div>
                        <button
                          onClick={removeLogo}
                          className="absolute -top-2 -right-2 size-5 rounded-full bg-destructive text-destructive-foreground flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                        >
                          <X className="size-3" />
                        </button>
                      </div>
                    ) : (
                      <div className="size-20 rounded-lg border-2 border-dashed border-border flex items-center justify-center text-muted-foreground">
                        <ImageIcon className="size-8" />
                      </div>
                    )}
                  </div>

                  {/* Upload area */}
                  <div className="flex-1 space-y-2">
                    <div
                      className="border-2 border-dashed border-border rounded-lg p-4 text-center hover:border-primary/50 transition-colors cursor-pointer"
                      onClick={() => !uploading && fileInputRef.current?.click()}
                    >
                      {uploading ? (
                        <Loader2 className="size-5 mx-auto text-primary animate-spin mb-1" />
                      ) : (
                        <Upload className="size-5 mx-auto text-muted-foreground mb-1" />
                      )}
                      <p className="text-sm text-muted-foreground">
                        {uploading ? 'Uploading...' : 'Click to upload logo'}
                      </p>
                      <p className="text-xs text-muted-foreground mt-0.5">PNG, JPG, SVG, WebP (max 10MB)</p>
                    </div>
                    <input
                      ref={fileInputRef}
                      type="file"
                      className="hidden"
                      accept=".png,.jpg,.jpeg,.svg,.webp,.ico"
                      onChange={handleLogoUpload}
                    />
                  </div>
                </div>
                <FieldDescription>
                  This logo appears in the sidebar, login page, and email headers.
                  {logoUrl && <span className="block text-xs font-mono mt-1 truncate">{logoUrl}</span>}
                </FieldDescription>
              </Field>

              <Field>
                <FieldLabel>Platform Name</FieldLabel>
                <Input value={form.platform_name || ''} onChange={(e) => update('platform_name', e.target.value)} />
                <FieldDescription>Displayed in the header, emails, and public pages</FieldDescription>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Contact Information</CardTitle>
            <CardDescription>Support contact details shown to merchants</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel>Support Email</FieldLabel>
                <Input type="email" value={form.support_email || ''} onChange={(e) => update('support_email', e.target.value)} />
                <FieldDescription>Merchants will see this email for support inquiries</FieldDescription>
              </Field>

              <Button onClick={() => mutation.mutate(form)} disabled={mutation.isPending} className="w-fit">
                {mutation.isPending ? <><Loader2 className="mr-2 size-4 animate-spin" />Saving...</> : 'Save Changes'}
              </Button>
            </FieldGroup>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
