import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Textarea } from '#/components/ui/textarea'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Badge } from '#/components/ui/badge'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import {
  CheckCircle2, XCircle, Eye, MoreHorizontal, ChevronLeft, ChevronRight,
  Snowflake, Ban, Unlock, FileText, Download,
} from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator,
} from '#/components/ui/dropdown-menu'

const PER_PAGE = 10

interface MerchantDoc {
  id: string; category: string; filename: string; contentType: string; fileSize: number; objectKey: string
}

export function MerchantsPage() {
  const queryClient = useQueryClient()
  const [selectedMerchant, setSelectedMerchant] = React.useState<any>(null)
  const [actionDialog, setActionDialog] = React.useState<{ type: string; merchant: any } | null>(null)
  const [reason, setReason] = React.useState('')
  const [page, setPage] = React.useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'merchants', { page, perPage: PER_PAGE }],
    queryFn: () => api.get<{ data: any[]; meta: { total: number } }>(`/v1/admin/merchants?page=${page}&perPage=${PER_PAGE}`),
    retry: false,
  })

  const merchants = data?.data || []
  const total = data?.meta?.total || 0
  const totalPages = Math.ceil(total / PER_PAGE)

  const doAction = useMutation({
    mutationFn: ({ id, action, body }: { id: string; action: string; body?: any }) =>
      api.post(`/v1/admin/merchants/${id}/${action}`, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
      setSelectedMerchant(null)
      setActionDialog(null)
      setReason('')
      toast.success('Action completed')
    },
    onError: (err: any) => toast.error(err.message),
  })

  const openAction = (type: string, merchant: any) => {
    setActionDialog({ type, merchant })
    setReason('')
  }

  const confirmAction = () => {
    if (!actionDialog) return
    const { type, merchant } = actionDialog
    const actionMap: Record<string, { action: string; body?: any }> = {
      approve: { action: 'approve' },
      reject: { action: 'reject', body: { reason: reason || 'Does not meet requirements' } },
      freeze: { action: 'freeze', body: { reason: reason || 'Suspicious activity' } },
      unfreeze: { action: 'unfreeze' },
      terminate: { action: 'terminate', body: { reason: reason || 'Terms violation' } },
      deactivate: { action: 'deactivate' },
    }
    const config = actionMap[type]
    if (config) doAction.mutate({ id: merchant.id, ...config })
  }

  const needsReason = actionDialog?.type === 'reject' || actionDialog?.type === 'freeze' || actionDialog?.type === 'terminate'
  const actionLabels: Record<string, { title: string; desc: string; btn: string; variant: string }> = {
    approve: { title: 'Approve Merchant', desc: 'Approve this merchant\'s KYC application.', btn: 'Approve', variant: 'default' },
    reject: { title: 'Reject Merchant', desc: 'Reject with feedback so the merchant can resubmit.', btn: 'Reject', variant: 'destructive' },
    freeze: { title: 'Freeze Account', desc: 'Freeze all funds and disable payment processing.', btn: 'Freeze Account', variant: 'destructive' },
    unfreeze: { title: 'Unfreeze Account', desc: 'Restore account and re-enable payment processing.', btn: 'Unfreeze', variant: 'default' },
    terminate: { title: 'Terminate Account', desc: 'Permanently terminate this merchant account. This cannot be undone.', btn: 'Terminate', variant: 'destructive' },
    deactivate: { title: 'Deactivate Account', desc: 'Temporarily deactivate this merchant account.', btn: 'Deactivate', variant: 'destructive' },
  }

  return (
    <>
      <PageHeader title="Merchants" description="Manage merchant accounts and KYC approvals" />

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Business Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>KYC Status</TableHead>
                <TableHead>Account Status</TableHead>
                <TableHead>Registered</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {merchants.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>
                    <EmptyState message={isLoading ? 'Loading merchants...' : 'No merchants registered yet.'} />
                  </TableCell>
                </TableRow>
              ) : (
                merchants.map((m: any) => (
                  <TableRow key={m.id}>
                    <TableCell className="font-medium">{m.businessName}</TableCell>
                    <TableCell className="text-sm">{m.contactEmail}</TableCell>
                    <TableCell><StatusBadge status={m.kycStatus} /></TableCell>
                    <TableCell>
                      <Badge className={
                        m.status === 'ACTIVE' ? 'bg-green-500/10 text-green-600' :
                        m.status === 'FROZEN' ? 'bg-blue-500/10 text-blue-600' :
                        m.status === 'TERMINATED' ? 'bg-red-500/10 text-red-600' :
                        'bg-muted text-muted-foreground'
                      }>{m.status}</Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(m.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger render={<Button variant="ghost" size="sm" />}>
                          <MoreHorizontal className="size-4" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="min-w-[180px]">
                          <DropdownMenuItem onClick={() => setSelectedMerchant(m)}>
                            <Eye className="size-4 mr-2" /> View & Review
                          </DropdownMenuItem>
                          {(m.kycStatus === 'UNDER_REVIEW' || m.kycStatus === 'INSTANT_ACCESS') && (
                            <>
                              <DropdownMenuItem onClick={() => openAction('approve', m)}>
                                <CheckCircle2 className="size-4 mr-2" /> Approve KYC
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => openAction('reject', m)}>
                                <XCircle className="size-4 mr-2" /> Reject KYC
                              </DropdownMenuItem>
                            </>
                          )}
                          <DropdownMenuSeparator />
                          {m.status === 'ACTIVE' && (
                            <>
                              <DropdownMenuItem onClick={() => openAction('freeze', m)}>
                                <Snowflake className="size-4 mr-2" /> Freeze Funds
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => openAction('terminate', m)} className="text-destructive focus:text-destructive">
                                <Ban className="size-4 mr-2" /> Terminate
                              </DropdownMenuItem>
                            </>
                          )}
                          {m.status === 'FROZEN' && (
                            <>
                              <DropdownMenuItem onClick={() => openAction('unfreeze', m)}>
                                <Unlock className="size-4 mr-2" /> Unfreeze
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => openAction('terminate', m)} className="text-destructive focus:text-destructive">
                                <Ban className="size-4 mr-2" /> Terminate
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {totalPages > 1 && (
            <div className="flex items-center justify-between border-t px-4 py-3">
              <p className="text-sm text-muted-foreground">
                Showing {(page - 1) * PER_PAGE + 1}–{Math.min(page * PER_PAGE, total)} of {total}
              </p>
              <div className="flex items-center gap-1">
                <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
                  <ChevronLeft className="size-4" /> Previous
                </Button>
                <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages}>
                  Next <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Merchant Detail / KYC Review Panel */}
      <MerchantDetailDialog
        merchant={selectedMerchant}
        onClose={() => setSelectedMerchant(null)}
        onAction={openAction}
      />

      {/* Action Confirmation Dialog */}
      <Dialog open={!!actionDialog} onOpenChange={(v) => !v && setActionDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{actionDialog ? actionLabels[actionDialog.type]?.title : ''}</DialogTitle>
            <DialogDescription>
              {actionDialog?.merchant?.businessName} — {actionDialog ? actionLabels[actionDialog.type]?.desc : ''}
            </DialogDescription>
          </DialogHeader>
          {needsReason && (
            <div className="py-2">
              <FieldGroup>
                <Field>
                  <FieldLabel>Reason / Feedback</FieldLabel>
                  <Textarea
                    placeholder="Enter reason..."
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                    rows={3}
                  />
                </Field>
              </FieldGroup>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setActionDialog(null)}>Cancel</Button>
            <Button
              variant={actionLabels[actionDialog?.type || '']?.variant === 'destructive' ? 'destructive' : 'default'}
              onClick={confirmAction}
              disabled={doAction.isPending}
            >
              {actionLabels[actionDialog?.type || '']?.btn || 'Confirm'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function MerchantDetailDialog({ merchant, onClose, onAction }: {
  merchant: any | null
  onClose: () => void
  onAction: (type: string, merchant: any) => void
}) {
  const queryClient = useQueryClient()
  const [forceApproveOpen, setForceApproveOpen] = React.useState(false)
  const [forceReason, setForceReason] = React.useState('')

  const { data: docsData } = useQuery({
    queryKey: ['admin', 'merchant-documents', merchant?.id],
    queryFn: () => api.get<{ data: MerchantDoc[] }>(`/v1/admin/merchants/${merchant?.id}/documents`),
    enabled: !!merchant,
  })

  const { data: directorsData } = useQuery({
    queryKey: ['admin', 'directors', merchant?.id],
    queryFn: () => api.get<{ data: any[] }>(`/v1/admin/merchants/${merchant?.id}/directors`),
    enabled: !!merchant,
  })

  const docs = docsData?.data || []
  const directors = directorsData?.data ?? []
  const verifiedCount = directors.filter((d: any) => d.status === 'VERIFIED').length
  const allVerified = directors.length === 0 || verifiedCount === directors.length
  const pendingCount = directors.length - verifiedCount

  const forceApproveMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: any }) =>
      api.post(`/v1/admin/merchants/${id}/approve`, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
      setForceApproveOpen(false)
      setForceReason('')
      onClose()
      toast.success('Force approved successfully')
    },
    onError: (err: any) => toast.error(err.message),
  })

  const handleForceApprove = () => {
    if (!merchant || !forceReason.trim()) return
    forceApproveMutation.mutate({ id: merchant.id, body: { force: true, forceReason } })
  }

  if (!merchant) return null

  return (
    <Dialog open={!!merchant} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            {merchant.businessName}
            <Badge className={
              merchant.status === 'ACTIVE' ? 'bg-green-500/10 text-green-600' :
              merchant.status === 'FROZEN' ? 'bg-blue-500/10 text-blue-600' :
              merchant.status === 'TERMINATED' ? 'bg-red-500/10 text-red-600' :
              'bg-muted text-muted-foreground'
            }>{merchant.status}</Badge>
            <StatusBadge status={merchant.kycStatus} />
          </DialogTitle>
          <DialogDescription>{merchant.contactEmail}</DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="business">
          <TabsList>
            <TabsTrigger value="business">Business Info</TabsTrigger>
            <TabsTrigger value="banking">Banking</TabsTrigger>
            <TabsTrigger value="documents">Documents ({docs.length})</TabsTrigger>
            <TabsTrigger value="directors">Directors ({verifiedCount}/{directors.length})</TabsTrigger>
            <TabsTrigger value="timeline">Timeline</TabsTrigger>
          </TabsList>

          <TabsContent value="business" className="mt-4 space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <DetailRow label="Business Name" value={merchant.businessName} />
              <DetailRow label="Business Type" value={merchant.businessType || '-'} />
              <DetailRow label="Registration No" value={merchant.registrationNo || '-'} />
              <DetailRow label="Contact Name" value={merchant.contactName || '-'} />
              <DetailRow label="Contact Phone" value={merchant.contactPhone || '-'} />
              <DetailRow label="Email" value={merchant.contactEmail} />
              <DetailRow label="Address" value={[merchant.addressLine1, merchant.city, merchant.postalCode].filter(Boolean).join(', ') || '-'} />
              <DetailRow label="Currency" value={merchant.defaultCurrency || 'LKR'} />
            </div>
            {merchant.statusReason && (
              <div className="rounded-md border p-3 mt-2">
                <p className="text-xs text-muted-foreground mb-1">Status Reason</p>
                <p className="text-sm">{merchant.statusReason}</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="banking" className="mt-4 space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <DetailRow label="Bank Name" value={merchant.bankName || '-'} />
              <DetailRow label="Branch" value={merchant.bankBranch || '-'} />
              <DetailRow label="Account No" value={merchant.bankAccountNo || '-'} />
              <DetailRow label="Account Name" value={merchant.bankAccountName || '-'} />
            </div>
          </TabsContent>

          <TabsContent value="documents" className="mt-4">
            {docs.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-6">No documents uploaded.</p>
            ) : (
              <div className="space-y-2">
                {docs.map((doc) => (
                  <div key={doc.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div className="flex items-center gap-3">
                      <FileText className="size-5 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">{doc.filename}</p>
                        <p className="text-xs text-muted-foreground">
                          {doc.category} — {(doc.fileSize / 1024).toFixed(1)} KB
                        </p>
                      </div>
                    </div>
                    <Button variant="ghost" size="sm" onClick={() => window.open(`http://localhost:8080/v1/assets/${doc.objectKey}`, '_blank')}>
                      <Download className="size-4 mr-1" /> View
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="directors" className="mt-4">
            {directors.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4">No directors registered.</p>
            ) : (
              <div className="flex flex-col gap-3">
                {directors.map((d: any) => (
                  <div key={d.id} className="rounded-lg border p-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium">{d.fullName || d.email}</p>
                        {d.fullName && <p className="text-xs text-muted-foreground">{d.email}</p>}
                      </div>
                      {d.status === 'VERIFIED' ? (
                        <Badge className="bg-green-500/10 text-green-600">Verified</Badge>
                      ) : d.tokenExpired ? (
                        <Badge className="bg-red-500/10 text-red-600">Expired</Badge>
                      ) : (
                        <Badge className="bg-amber-500/10 text-amber-600">Pending</Badge>
                      )}
                    </div>
                    {d.status === 'VERIFIED' && (
                      <div className="mt-2 grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
                        <div><span className="text-muted-foreground">NIC/Passport:</span> {d.nicPassportNumber}</div>
                        <div><span className="text-muted-foreground">Phone:</span> {d.phone}</div>
                        <div><span className="text-muted-foreground">DOB:</span> {d.dateOfBirth}</div>
                        <div><span className="text-muted-foreground">Verified:</span> {d.verifiedAt ? new Date(d.verifiedAt).toLocaleDateString() : '-'}</div>
                        {d.address && <div className="col-span-2"><span className="text-muted-foreground">Address:</span> {d.address}</div>}
                        {d.documentFilename && (
                          <div className="col-span-2">
                            <span className="text-muted-foreground">Document:</span>{' '}
                            <button className="text-primary underline" onClick={() => window.open(`http://localhost:8080/v1/assets/${d.documentObjectKey}`, '_blank')}>
                              {d.documentFilename}
                            </button>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="timeline" className="mt-4 text-sm space-y-3">
            <TimelineItem label="Registered" date={merchant.createdAt} />
            {merchant.kycSubmittedAt && <TimelineItem label="KYC Submitted" date={merchant.kycSubmittedAt} />}
            {merchant.kycReviewedAt && <TimelineItem label="KYC Reviewed" date={merchant.kycReviewedAt} status={merchant.kycStatus} />}
            {merchant.kycRejectionReason && (
              <div className="rounded-md bg-red-500/10 p-3">
                <p className="text-xs text-muted-foreground mb-1">Rejection Reason</p>
                <p className="text-sm text-red-600">{merchant.kycRejectionReason}</p>
              </div>
            )}
            {merchant.statusChangedAt && <TimelineItem label={`Status: ${merchant.status}`} date={merchant.statusChangedAt} />}
            {merchant.kycReviewNotes && (
              <div className="rounded-md border p-3">
                <p className="text-xs text-muted-foreground mb-1">Review Notes</p>
                <p className="text-sm">{merchant.kycReviewNotes}</p>
              </div>
            )}
          </TabsContent>
        </Tabs>

        {forceApproveOpen && (
          <div className="rounded-md border border-amber-300 bg-amber-500/10 p-4 space-y-3">
            <p className="text-sm font-medium text-amber-700">
              Warning: {pendingCount} of {directors.length} directors have not completed verification.
            </p>
            <Textarea
              placeholder="Enter reason for force approval (required)..."
              value={forceReason}
              onChange={(e) => setForceReason(e.target.value)}
              rows={3}
            />
            <div className="flex gap-2">
              <Button
                size="sm"
                className="bg-amber-600 hover:bg-amber-700 text-white"
                onClick={handleForceApprove}
                disabled={!forceReason.trim() || forceApproveMutation.isPending}
              >
                Confirm Force Approve
              </Button>
              <Button size="sm" variant="outline" onClick={() => { setForceApproveOpen(false); setForceReason('') }}>
                Cancel
              </Button>
            </div>
          </div>
        )}

        <DialogFooter className="gap-2 flex-wrap">
          {(merchant.kycStatus === 'UNDER_REVIEW' || merchant.kycStatus === 'INSTANT_ACCESS') && (
            <>
              <Button
                className="bg-green-600 hover:bg-green-700 text-white"
                onClick={() => { onClose(); onAction('approve', merchant) }}
                disabled={!allVerified}
                title={allVerified ? '' : 'All directors must verify first'}
              >
                <CheckCircle2 className="size-4 mr-1" /> Approve
              </Button>
              <Button
                variant="outline"
                className="border-amber-400 text-amber-700 hover:bg-amber-500/10"
                onClick={() => setForceApproveOpen((v) => !v)}
              >
                Force Approve
              </Button>
              <Button variant="destructive" onClick={() => { onClose(); onAction('reject', merchant) }}>
                <XCircle className="size-4 mr-1" /> Reject
              </Button>
            </>
          )}
          {merchant.status === 'ACTIVE' && (
            <Button variant="outline" onClick={() => { onClose(); onAction('freeze', merchant) }}>
              <Snowflake className="size-4 mr-1" /> Freeze
            </Button>
          )}
          {merchant.status === 'FROZEN' && (
            <Button variant="outline" onClick={() => { onClose(); onAction('unfreeze', merchant) }}>
              <Unlock className="size-4 mr-1" /> Unfreeze
            </Button>
          )}
          <Button variant="outline" onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="font-medium">{value}</p>
    </div>
  )
}

function TimelineItem({ label, date, status }: { label: string; date: string; status?: string }) {
  return (
    <div className="flex items-center gap-3">
      <div className="size-2 rounded-full bg-primary shrink-0" />
      <div className="flex-1">
        <p className="font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{new Date(date).toLocaleString()}</p>
      </div>
      {status && <StatusBadge status={status} />}
    </div>
  )
}
