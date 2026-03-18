import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog'
import { ScrollArea } from '#/components/ui/scroll-area'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, CheckCircle2, Loader2, Eye, Pencil, FileText, Upload, X, Download } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface LegalDoc {
  id: string
  type: string
  version: number
  title: string
  content: string
  isActive: boolean
  createdAt: string
  pdfObjectKey?: string
}

export function SettingsLegalDocumentsPage() {
  const queryClient = useQueryClient()

  // Create dialog state
  const [createOpen, setCreateOpen] = React.useState(false)
  const [createForm, setCreateForm] = React.useState({ type: 'terms_and_conditions', version: 1, title: '', content: '' })
  const [createPdfKey, setCreatePdfKey] = React.useState<string | undefined>(undefined)
  const [createPdfFilename, setCreatePdfFilename] = React.useState<string | undefined>(undefined)
  const [createUploading, setCreateUploading] = React.useState(false)
  const createFileRef = React.useRef<HTMLInputElement>(null)

  // View dialog state
  const [viewDoc, setViewDoc] = React.useState<LegalDoc | null>(null)

  // Edit dialog state
  const [editDoc, setEditDoc] = React.useState<LegalDoc | null>(null)
  const [editForm, setEditForm] = React.useState({ title: '', content: '' })
  const [editPdfKey, setEditPdfKey] = React.useState<string | undefined>(undefined)
  const [editPdfFilename, setEditPdfFilename] = React.useState<string | undefined>(undefined)
  const [editUploading, setEditUploading] = React.useState(false)
  const editFileRef = React.useRef<HTMLInputElement>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'legal-documents'],
    queryFn: () => api.get<{ data: LegalDoc[] }>('/v1/admin/legal-documents'),
  })

  const createMutation = useMutation({
    mutationFn: (payload: typeof createForm & { pdfObjectKey?: string }) =>
      api.post('/v1/admin/legal-documents', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      setCreateOpen(false)
      setCreateForm({ type: 'terms_and_conditions', version: 1, title: '', content: '' })
      setCreatePdfKey(undefined)
      setCreatePdfFilename(undefined)
      toast.success('Document created')
    },
    onError: () => toast.error('Failed to create document'),
  })

  const editMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: { title: string; content: string; pdfObjectKey?: string } }) =>
      api.put(`/v1/admin/legal-documents/${id}`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      setEditDoc(null)
      toast.success('Document updated')
    },
    onError: () => toast.error('Failed to update document'),
  })

  const activateMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/admin/legal-documents/${id}/activate`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      toast.success('Document activated')
    },
    onError: () => toast.error('Failed to activate document'),
  })

  const docs = data?.data || []

  // Open edit dialog and pre-populate form
  function openEdit(doc: LegalDoc) {
    setEditDoc(doc)
    setEditForm({ title: doc.title, content: doc.content })
    setEditPdfKey(doc.pdfObjectKey)
    setEditPdfFilename(doc.pdfObjectKey ? doc.pdfObjectKey.split('/').pop() : undefined)
  }

  // Handle PDF upload for create dialog
  async function handleCreatePdfUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setCreateUploading(true)
    try {
      const result = await api.upload<{ data: { key: string; filename: string } }>(file, 'legal-documents')
      setCreatePdfKey(result.data.key)
      setCreatePdfFilename(result.data.filename ?? file.name)
      toast.success('PDF uploaded')
    } catch {
      toast.error('PDF upload failed')
    } finally {
      setCreateUploading(false)
      if (createFileRef.current) createFileRef.current.value = ''
    }
  }

  // Handle PDF upload for edit dialog
  async function handleEditPdfUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setEditUploading(true)
    try {
      const result = await api.upload<{ data: { key: string; filename: string } }>(file, 'legal-documents')
      setEditPdfKey(result.data.key)
      setEditPdfFilename(result.data.filename ?? file.name)
      toast.success('PDF uploaded')
    } catch {
      toast.error('PDF upload failed')
    } finally {
      setEditUploading(false)
      if (editFileRef.current) editFileRef.current.value = ''
    }
  }

  return (
    <>
      <PageHeader
        title="Legal Documents"
        description="Manage versioned terms, conditions, and policies"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 size-4" />New Version
          </Button>
        }
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
                <TableRow>
                  <TableCell colSpan={6}>
                    <EmptyState message={isLoading ? 'Loading...' : 'No legal documents yet.'} />
                  </TableCell>
                </TableRow>
              ) : docs.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono text-xs">{d.type}</TableCell>
                  <TableCell>v{d.version}</TableCell>
                  <TableCell className="font-medium">{d.title}</TableCell>
                  <TableCell>
                    {d.isActive
                      ? <Badge className="bg-green-500/10 text-green-600">Active</Badge>
                      : <Badge variant="secondary">Inactive</Badge>}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(d.createdAt).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button variant="ghost" size="sm" onClick={() => setViewDoc(d)}>
                        <Eye className="size-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => openEdit(d)}>
                        <Pencil className="size-4" />
                      </Button>
                      {!d.isActive && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => activateMutation.mutate(d.id)}
                          disabled={activateMutation.isPending}
                        >
                          <CheckCircle2 className="size-4 mr-1" />Activate
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* View Dialog */}
      <Dialog open={!!viewDoc} onOpenChange={(open) => { if (!open) setViewDoc(null) }}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {viewDoc?.title}
              <Badge variant="secondary">v{viewDoc?.version}</Badge>
            </DialogTitle>
          </DialogHeader>
          <ScrollArea className="max-h-[60vh] rounded-md border p-4">
            <pre className="whitespace-pre-wrap text-sm font-sans leading-relaxed">{viewDoc?.content}</pre>
          </ScrollArea>
          <DialogFooter className="flex items-center justify-between sm:justify-between">
            <div>
              {viewDoc?.pdfObjectKey && (
                <Button
                  variant="outline"
                  onClick={() => window.open(`http://localhost:8080/v1/assets/${viewDoc.pdfObjectKey}`, '_blank')}
                >
                  <Download className="mr-2 size-4" />Download PDF
                </Button>
              )}
            </div>
            <Button variant="outline" onClick={() => setViewDoc(null)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={!!editDoc} onOpenChange={(open) => { if (!open) setEditDoc(null) }}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Edit Document</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Title</FieldLabel>
              <Input
                value={editForm.title}
                onChange={(e) => setEditForm({ ...editForm, title: e.target.value })}
              />
            </Field>
            <Field>
              <FieldLabel>Content</FieldLabel>
              <Textarea
                rows={16}
                value={editForm.content}
                onChange={(e) => setEditForm({ ...editForm, content: e.target.value })}
              />
            </Field>
            <Field>
              <FieldLabel>PDF Attachment</FieldLabel>
              {editPdfKey ? (
                <div className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                  <FileText className="size-4 text-muted-foreground shrink-0" />
                  <span className="flex-1 truncate text-muted-foreground">{editPdfFilename ?? editPdfKey}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto px-2 py-0.5 text-xs"
                    onClick={() => editFileRef.current?.click()}
                    disabled={editUploading}
                  >
                    Replace
                  </Button>
                  <button
                    className="text-muted-foreground hover:text-destructive"
                    onClick={() => { setEditPdfKey(undefined); setEditPdfFilename(undefined) }}
                    type="button"
                    aria-label="Remove PDF"
                  >
                    <X className="size-4" />
                  </button>
                </div>
              ) : (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => editFileRef.current?.click()}
                  disabled={editUploading}
                >
                  {editUploading
                    ? <Loader2 className="mr-2 size-4 animate-spin" />
                    : <Upload className="mr-2 size-4" />}
                  {editUploading ? 'Uploading...' : 'Upload PDF'}
                </Button>
              )}
              <input
                ref={editFileRef}
                type="file"
                accept=".pdf"
                className="hidden"
                onChange={handleEditPdfUpload}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDoc(null)}>Cancel</Button>
            <Button
              onClick={() => {
                if (!editDoc) return
                editMutation.mutate({
                  id: editDoc.id,
                  payload: { title: editForm.title, content: editForm.content, pdfObjectKey: editPdfKey },
                })
              }}
              disabled={editMutation.isPending || editUploading}
            >
              {editMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader><DialogTitle>Create New Document Version</DialogTitle></DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Type</FieldLabel>
                <Input
                  value={createForm.type}
                  onChange={(e) => setCreateForm({ ...createForm, type: e.target.value })}
                />
              </Field>
              <Field>
                <FieldLabel>Version</FieldLabel>
                <Input
                  type="number"
                  value={createForm.version}
                  onChange={(e) => setCreateForm({ ...createForm, version: parseInt(e.target.value) || 1 })}
                />
              </Field>
            </div>
            <Field>
              <FieldLabel>Title</FieldLabel>
              <Input
                value={createForm.title}
                onChange={(e) => setCreateForm({ ...createForm, title: e.target.value })}
              />
            </Field>
            <Field>
              <FieldLabel>Content</FieldLabel>
              <Textarea
                rows={12}
                value={createForm.content}
                onChange={(e) => setCreateForm({ ...createForm, content: e.target.value })}
              />
            </Field>
            <Field>
              <FieldLabel>PDF Attachment</FieldLabel>
              {createPdfKey ? (
                <div className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                  <FileText className="size-4 text-muted-foreground shrink-0" />
                  <span className="flex-1 truncate text-muted-foreground">{createPdfFilename ?? createPdfKey}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto px-2 py-0.5 text-xs"
                    onClick={() => createFileRef.current?.click()}
                    disabled={createUploading}
                  >
                    Replace
                  </Button>
                  <button
                    className="text-muted-foreground hover:text-destructive"
                    onClick={() => { setCreatePdfKey(undefined); setCreatePdfFilename(undefined) }}
                    type="button"
                    aria-label="Remove PDF"
                  >
                    <X className="size-4" />
                  </button>
                </div>
              ) : (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => createFileRef.current?.click()}
                  disabled={createUploading}
                >
                  {createUploading
                    ? <Loader2 className="mr-2 size-4 animate-spin" />
                    : <Upload className="mr-2 size-4" />}
                  {createUploading ? 'Uploading...' : 'Upload PDF'}
                </Button>
              )}
              <input
                ref={createFileRef}
                type="file"
                accept=".pdf"
                className="hidden"
                onChange={handleCreatePdfUpload}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button
              onClick={() => createMutation.mutate({ ...createForm, pdfObjectKey: createPdfKey })}
              disabled={createMutation.isPending || createUploading}
            >
              {createMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
