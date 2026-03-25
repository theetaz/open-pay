import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel, FieldDescription } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Pencil, Loader2 } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface EmailTemplate {
  id: string; eventType: string; name: string; subject: string; bodyHtml: string; variables: string[]; isActive: boolean
}

export function SettingsEmailTemplatesPage() {
  const queryClient = useQueryClient()
  const [editTemplate, setEditTemplate] = React.useState<EmailTemplate | null>(null)
  const [form, setForm] = React.useState({ name: '', subject: '', bodyHtml: '' })

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'email-templates'],
    queryFn: () => api.get<{ data: EmailTemplate[] }>('/v1/admin/email-templates'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...body }: { id: string; name: string; subject: string; bodyHtml: string; variables: string[] }) =>
      api.put(`/v1/admin/email-templates/${id}`, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'email-templates'] })
      setEditTemplate(null)
      toast.success('Template updated')
    },
  })

  const templates = data?.data || []

  const openEdit = (t: EmailTemplate) => {
    setEditTemplate(t)
    setForm({ name: t.name, subject: t.subject, bodyHtml: t.bodyHtml })
  }

  return (
    <>
      <PageHeader title="Email Templates" description="Manage notification email templates and content" />

      <Card>
        <CardHeader>
          <CardTitle>Templates</CardTitle>
          <CardDescription>Edit the subject and body of notification emails. Use {'{{variable}}'} for dynamic content.</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Event Type</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Subject</TableHead>
                <TableHead>Variables</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {templates.length === 0 ? (
                <TableRow><TableCell colSpan={5}><EmptyState message={isLoading ? 'Loading...' : 'No templates.'} /></TableCell></TableRow>
              ) : templates.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-mono text-xs">{t.eventType}</TableCell>
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell className="text-sm max-w-[200px] truncate">{t.subject}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {(t.variables || []).map((v) => (
                        <Badge key={v} variant="secondary" className="text-[10px]">{`{{${v}}}`}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Button variant="ghost" size="sm" onClick={() => openEdit(t)}>
                      <Pencil className="size-4 mr-1" />Edit
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={!!editTemplate} onOpenChange={(v) => !v && setEditTemplate(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Email Template</DialogTitle>
            <DialogDescription>Event: {editTemplate?.eventType}</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Template Name</FieldLabel>
              <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </Field>
            <Field>
              <FieldLabel>Subject Line</FieldLabel>
              <Input value={form.subject} onChange={(e) => setForm({ ...form, subject: e.target.value })} />
            </Field>
            <Field>
              <FieldLabel>Body (HTML)</FieldLabel>
              <Textarea rows={10} className="font-mono text-xs" value={form.bodyHtml} onChange={(e) => setForm({ ...form, bodyHtml: e.target.value })} />
              <FieldDescription>Available variables: {editTemplate?.variables?.map((v) => `{{${v}}}`).join(', ')}</FieldDescription>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTemplate(null)}>Cancel</Button>
            <Button
              onClick={() => editTemplate && updateMutation.mutate({ id: editTemplate.id, ...form, variables: editTemplate.variables })}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}Save Template
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
