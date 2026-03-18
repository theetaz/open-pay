import * as React from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from '#/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { ScrollArea } from '#/components/ui/scroll-area'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, CheckCircle2, Loader2, Eye, Pencil, FileText, Upload, X, Download } from 'lucide-react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { api } from '#/lib/api'
import { toast } from 'sonner'

const markdownComponents = {
  h1: ({ children, ...props }: React.ComponentProps<'h1'>) => <h1 className="text-xl font-bold mt-4 mb-2" {...props}>{children}</h1>,
  h2: ({ children, ...props }: React.ComponentProps<'h2'>) => <h2 className="text-lg font-semibold mt-3 mb-2" {...props}>{children}</h2>,
  h3: ({ children, ...props }: React.ComponentProps<'h3'>) => <h3 className="text-base font-semibold mt-3 mb-1" {...props}>{children}</h3>,
  p: ({ children, ...props }: React.ComponentProps<'p'>) => <p className="text-sm leading-relaxed mb-2" {...props}>{children}</p>,
  ul: ({ children, ...props }: React.ComponentProps<'ul'>) => <ul className="list-disc pl-5 mb-2 text-sm" {...props}>{children}</ul>,
  ol: ({ children, ...props }: React.ComponentProps<'ol'>) => <ol className="list-decimal pl-5 mb-2 text-sm" {...props}>{children}</ol>,
  li: ({ children, ...props }: React.ComponentProps<'li'>) => <li className="mb-1" {...props}>{children}</li>,
  strong: ({ children, ...props }: React.ComponentProps<'strong'>) => <strong className="font-semibold" {...props}>{children}</strong>,
  a: ({ children, ...props }: React.ComponentProps<'a'>) => <a className="text-primary underline" {...props}>{children}</a>,
  table: ({ children, ...props }: React.ComponentProps<'table'>) => <table className="w-full border-collapse border border-border my-2 text-sm" {...props}>{children}</table>,
  th: ({ children, ...props }: React.ComponentProps<'th'>) => <th className="border border-border px-3 py-1.5 bg-muted/50 text-left font-medium" {...props}>{children}</th>,
  td: ({ children, ...props }: React.ComponentProps<'td'>) => <td className="border border-border px-3 py-1.5" {...props}>{children}</td>,
  blockquote: ({ children, ...props }: React.ComponentProps<'blockquote'>) => <blockquote className="border-l-4 border-primary/30 pl-4 italic my-2 text-muted-foreground" {...props}>{children}</blockquote>,
  hr: (props: React.ComponentProps<'hr'>) => <hr className="my-4 border-border" {...props} />,
  code: ({ children, ...props }: React.ComponentProps<'code'>) => <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-mono" {...props}>{children}</code>,
}

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

const DOCUMENT_TYPES = [
  { value: 'terms_and_conditions', label: 'Terms & Conditions' },
  { value: 'privacy_policy', label: 'Privacy Policy' },
  { value: 'sign_agreement', label: 'Sign Agreement' },
] as const

function typeLabel(type: string) {
  return DOCUMENT_TYPES.find(t => t.value === type)?.label || type
}

export function SettingsLegalDocumentsPage() {
  const queryClient = useQueryClient()

  // Selected document type filter — persisted in URL params
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedType = searchParams.get('type') || 'terms_and_conditions'
  const setSelectedType = (type: string) => setSearchParams({ type }, { replace: true })

  // Create/New version dialog state
  const [createOpen, setCreateOpen] = React.useState(false)
  const [createForm, setCreateForm] = React.useState({ type: '', version: 1, title: '', content: '' })
  const [createPdfKey, setCreatePdfKey] = React.useState<string | undefined>(undefined)
  const [createPdfFilename, setCreatePdfFilename] = React.useState<string | undefined>(undefined)
  const [createUploading, setCreateUploading] = React.useState(false)
  const createFileRef = React.useRef<HTMLInputElement>(null)

  // View dialog state
  const [viewDoc, setViewDoc] = React.useState<LegalDoc | null>(null)

  // Markdown editor mode: 'write' or 'preview'
  const [editorMode, setEditorMode] = React.useState<'write' | 'preview'>('write')

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
      toast.success('New version created successfully')
    },
    onError: () => toast.error('Failed to create document'),
  })

  const activateMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/admin/legal-documents/${id}/activate`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'legal-documents'] })
      toast.success('Document version activated')
    },
    onError: () => toast.error('Failed to activate document'),
  })

  const allDocs = data?.data || []

  // Filter docs by selected type
  const typeDocs = allDocs.filter(d => d.type === selectedType)
  const activeDoc = typeDocs.find(d => d.isActive)
  const versionHistory = typeDocs.filter(d => !d.isActive).sort((a, b) => b.version - a.version)

  // Open new version dialog prepopulated from a source doc (or the active version of selected type)
  function openNewVersion(sourceDoc?: LegalDoc) {
    const type = sourceDoc?.type || selectedType
    const maxVersion = allDocs.filter(d => d.type === type).reduce((max, d) => Math.max(max, d.version), 0)
    setCreateForm({
      type,
      version: maxVersion + 1,
      title: sourceDoc?.title || '',
      content: sourceDoc?.content || '',
    })
    setCreatePdfKey(sourceDoc?.pdfObjectKey)
    setCreatePdfFilename(sourceDoc?.pdfObjectKey ? sourceDoc.pdfObjectKey.split('/').pop() : undefined)
    setEditorMode('write')
    setCreateOpen(true)
  }

  async function handlePdfUpload(e: React.ChangeEvent<HTMLInputElement>) {
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

  return (
    <>
      <PageHeader
        title="Legal Documents"
        description="Manage versioned terms, conditions, and policies"
        action={
          <Button onClick={() => openNewVersion(activeDoc)}>
            <Plus className="mr-2 size-4" />New Version
          </Button>
        }
      />

      {/* Type selector */}
      <div className="flex items-center gap-3 mb-4">
        <span className="text-sm font-medium text-muted-foreground">Document Type:</span>
        <div className="flex gap-1">
          {DOCUMENT_TYPES.map(t => (
            <Button
              key={t.value}
              variant={selectedType === t.value ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedType(t.value)}
            >
              {t.label}
            </Button>
          ))}
        </div>
      </div>

      {/* Active version card */}
      {activeDoc ? (
        <Card className="mb-4">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{activeDoc.title}</CardTitle>
                <Badge className="bg-green-500/10 text-green-600">Active</Badge>
                <Badge variant="secondary">v{activeDoc.version}</Badge>
                {activeDoc.pdfObjectKey && (
                  <Badge variant="secondary" className="gap-1">
                    <FileText className="size-3" />PDF
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-1">
                <Button variant="ghost" size="sm" onClick={() => setViewDoc(activeDoc)} title="View full content">
                  <Eye className="size-4 mr-1" />View
                </Button>
                <Button variant="ghost" size="sm" onClick={() => openNewVersion(activeDoc)} title="Edit and save as new version">
                  <Pencil className="size-4 mr-1" />Edit
                </Button>
                {activeDoc.pdfObjectKey && (
                  <Button variant="ghost" size="sm" onClick={() => window.open(`http://localhost:8080/v1/assets/${activeDoc.pdfObjectKey}`, '_blank')}>
                    <Download className="size-4 mr-1" />PDF
                  </Button>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border bg-muted/30 p-3">
              <p className="text-xs text-muted-foreground line-clamp-4 whitespace-pre-line">{activeDoc.content}</p>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Created: {new Date(activeDoc.createdAt).toLocaleDateString()}
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card className="mb-4">
          <CardContent className="py-8">
            <EmptyState
              message={isLoading ? 'Loading...' : `No ${typeLabel(selectedType)} document yet.`}
            />
            {!isLoading && (
              <div className="flex justify-center mt-3">
                <Button variant="outline" size="sm" onClick={() => openNewVersion()}>
                  <Plus className="size-4 mr-1" />Create First Version
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Version history */}
      {versionHistory.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">Version History</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Version</TableHead>
                  <TableHead>Title</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {versionHistory.map((d) => (
                  <TableRow key={d.id}>
                    <TableCell>v{d.version}</TableCell>
                    <TableCell className="font-medium">{d.title}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(d.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button variant="ghost" size="sm" onClick={() => setViewDoc(d)} title="View">
                          <Eye className="size-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => openNewVersion(d)} title="Create new version from this">
                          <Pencil className="size-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => activateMutation.mutate(d.id)}
                          disabled={activateMutation.isPending}
                          title="Activate this version"
                        >
                          <CheckCircle2 className="size-4 mr-1" />Activate
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* View Dialog */}
      <Dialog open={!!viewDoc} onOpenChange={(open) => { if (!open) setViewDoc(null) }}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {viewDoc?.title}
              <Badge variant="secondary">v{viewDoc?.version}</Badge>
            </DialogTitle>
            <DialogDescription>{viewDoc && typeLabel(viewDoc.type)}</DialogDescription>
          </DialogHeader>
          <ScrollArea className="max-h-[60vh] rounded-md border p-4">
            <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
              {viewDoc?.content || ''}
            </Markdown>
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

      {/* Create / New Version Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {createForm.content ? `Edit ${typeLabel(createForm.type)}` : `Create ${typeLabel(createForm.type)}`}
            </DialogTitle>
            <DialogDescription>
              {createForm.content
                ? 'Make your changes below. This will be saved as a new version.'
                : 'Create the first version of this document.'}
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Document Type</FieldLabel>
                <Select
                  value={createForm.type}
                  onValueChange={(value: string | null) => {
                    if (!value) return
                    // When type changes, prepopulate from that type's active doc
                    const typeActiveDoc = allDocs.find(d => d.type === value && d.isActive)
                    const maxVer = allDocs.filter(d => d.type === value).reduce((max, d) => Math.max(max, d.version), 0)
                    setCreateForm({
                      type: value,
                      version: maxVer + 1,
                      title: typeActiveDoc?.title || '',
                      content: typeActiveDoc?.content || '',
                    })
                    setCreatePdfKey(typeActiveDoc?.pdfObjectKey)
                    setCreatePdfFilename(typeActiveDoc?.pdfObjectKey ? typeActiveDoc.pdfObjectKey.split('/').pop() : undefined)
                  }}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DOCUMENT_TYPES.map(t => (
                      <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>Version</FieldLabel>
                <Input
                  type="number"
                  value={createForm.version}
                  readOnly
                  className="bg-muted"
                />
              </Field>
            </div>
            <Field>
              <FieldLabel>Title</FieldLabel>
              <Input
                value={createForm.title}
                onChange={(e) => setCreateForm({ ...createForm, title: e.target.value })}
                placeholder="e.g. Open Pay — Terms and Conditions"
              />
            </Field>
            <Field>
              <div className="flex items-center justify-between mb-1">
                <FieldLabel>Content (Markdown supported)</FieldLabel>
                <div className="flex gap-1 rounded-md border p-0.5">
                  <button
                    type="button"
                    onClick={() => setEditorMode('write')}
                    className={`px-3 py-1 text-xs rounded transition-colors ${editorMode === 'write' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                  >
                    Write
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditorMode('preview')}
                    className={`px-3 py-1 text-xs rounded transition-colors ${editorMode === 'preview' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                  >
                    Preview
                  </button>
                </div>
              </div>
              {editorMode === 'write' ? (
                <Textarea
                  rows={16}
                  value={createForm.content}
                  onChange={(e) => setCreateForm({ ...createForm, content: e.target.value })}
                  placeholder="Enter document content using Markdown formatting..."
                  className="font-mono text-sm"
                />
              ) : (
                <div className="rounded-md border p-4 min-h-[384px] max-h-[384px] overflow-y-auto">
                  {createForm.content ? (
                    <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                      {createForm.content}
                    </Markdown>
                  ) : (
                    <p className="text-sm text-muted-foreground italic">Nothing to preview yet...</p>
                  )}
                </div>
              )}
            </Field>
            <Field>
              <FieldLabel>PDF Attachment (optional)</FieldLabel>
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
                onChange={handlePdfUpload}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button
              onClick={() => createMutation.mutate({ ...createForm, pdfObjectKey: createPdfKey })}
              disabled={createMutation.isPending || createUploading || !createForm.title || !createForm.content}
            >
              {createMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}
              Save as v{createForm.version}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
