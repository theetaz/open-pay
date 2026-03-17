import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, CheckCircle2, Loader2 } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface LegalDoc {
  id: string; type: string; version: number; title: string; content: string; isActive: boolean; createdAt: string
}

export function SettingsLegalDocumentsPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = React.useState(false)
  const [form, setForm] = React.useState({ type: 'terms_and_conditions', version: 1, title: '', content: '' })

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'legal-documents'],
    queryFn: () => api.get<{ data: LegalDoc[] }>('/v1/admin/legal-documents'),
  })

  const createMutation = useMutation({
    mutationFn: (doc: typeof form) => api.post('/v1/admin/legal-documents', doc),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      setCreateOpen(false)
      toast.success('Document created')
    },
  })

  const activateMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/admin/legal-documents/${id}/activate`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      toast.success('Document activated')
    },
  })

  const docs = data?.data || []

  return (
    <>
      <PageHeader title="Legal Documents" description="Manage versioned terms, conditions, and policies"
        action={<Button onClick={() => setCreateOpen(true)}><Plus className="mr-2 size-4" />New Version</Button>}
      />

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Type</TableHead>
                <TableHead>Version</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {docs.length === 0 ? (
                <TableRow><TableCell colSpan={6}><EmptyState message={isLoading ? 'Loading...' : 'No legal documents yet.'} /></TableCell></TableRow>
              ) : docs.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono text-xs">{d.type}</TableCell>
                  <TableCell>v{d.version}</TableCell>
                  <TableCell className="font-medium">{d.title}</TableCell>
                  <TableCell>{d.isActive ? <Badge className="bg-green-500/10 text-green-600">Active</Badge> : <Badge variant="secondary">Inactive</Badge>}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{new Date(d.createdAt).toLocaleDateString()}</TableCell>
                  <TableCell>
                    {!d.isActive && (
                      <Button variant="ghost" size="sm" onClick={() => activateMutation.mutate(d.id)} disabled={activateMutation.isPending}>
                        <CheckCircle2 className="size-4 mr-1" />Activate
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader><DialogTitle>Create New Document Version</DialogTitle></DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Type</FieldLabel>
                <Input value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} />
              </Field>
              <Field>
                <FieldLabel>Version</FieldLabel>
                <Input type="number" value={form.version} onChange={(e) => setForm({ ...form, version: parseInt(e.target.value) || 1 })} />
              </Field>
            </div>
            <Field>
              <FieldLabel>Title</FieldLabel>
              <Input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
            </Field>
            <Field>
              <FieldLabel>Content</FieldLabel>
              <Textarea rows={12} value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={() => createMutation.mutate(form)} disabled={createMutation.isPending}>
              {createMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
