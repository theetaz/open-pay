import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel, FieldDescription } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Pencil, Loader2 } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface EmailTemplate {
  id: string; eventType: string; name: string; subject: string; bodyHtml: string; variables: string[]; isActive: boolean
}

const EMAIL_TEMPLATE_FILTERS: FilterConfig[] = []

export function SettingsEmailTemplatesPage() {
  const queryClient = useQueryClient()
  const [editTemplate, setEditTemplate] = React.useState<EmailTemplate | null>(null)
  const [form, setForm] = React.useState({ name: '', subject: '', bodyHtml: '' })
  const [search, setSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({})

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

  const filteredTemplates = React.useMemo(() => {
    if (!search) return templates
    const q = search.toLowerCase()
    return templates.filter(
      (t) =>
        t.name.toLowerCase().includes(q) ||
        t.subject.toLowerCase().includes(q) ||
        t.eventType.toLowerCase().includes(q)
    )
  }, [templates, search])

  const openEdit = (t: EmailTemplate) => {
    setEditTemplate(t)
    setForm({ name: t.name, subject: t.subject, bodyHtml: t.bodyHtml })
  }

  const columns: ColumnDef<EmailTemplate>[] = [
    {
      accessorKey: 'eventType',
      header: 'Event Type',
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.eventType}</span>,
    },
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
    },
    {
      accessorKey: 'subject',
      header: 'Subject',
      cell: ({ row }) => <span className="text-sm max-w-[200px] truncate block">{row.original.subject}</span>,
    },
    {
      id: 'variables',
      header: 'Variables',
      cell: ({ row }) => (
        <div className="flex flex-wrap gap-1">
          {(row.original.variables || []).map((v) => (
            <Badge key={v} variant="secondary" className="text-[10px]">{`{{${v}}}`}</Badge>
          ))}
        </div>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      enableHiding: false,
      cell: ({ row }) => (
        <Button variant="ghost" size="sm" onClick={() => openEdit(row.original)}>
          <Pencil className="size-4 mr-1" />Edit
        </Button>
      ),
    },
  ]

  return (
    <>
      <PageHeader title="Email Templates" description="Manage notification email templates and content" />

      <DataTable
        columns={columns}
        data={filteredTemplates}
        filters={EMAIL_TEMPLATE_FILTERS}
        filterValues={filterValues}
        onFilterChange={(id, value) => setFilterValues((prev) => ({ ...prev, [id]: value }))}
        onClearFilters={() => { setFilterValues({}); setSearch('') }}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search by name or subject..."
        pagination={{ page: 1, perPage: 999, total: filteredTemplates.length }}
        onPageChange={() => {}}
        isLoading={isLoading}
      />

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
