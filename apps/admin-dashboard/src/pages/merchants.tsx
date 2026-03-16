import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { CheckCircle2, XCircle, Eye, Trash2, ChevronLeft, ChevronRight, MoreHorizontal } from 'lucide-react'
import { api } from '#/lib/api'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Input } from '#/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '#/components/ui/dropdown-menu'

const PER_PAGE = 10

export function MerchantsPage() {
  const queryClient = useQueryClient()
  const [selectedMerchant, setSelectedMerchant] = React.useState<any>(null)
  const [rejectDialog, setRejectDialog] = React.useState<{ open: boolean; merchantId: string | null }>({ open: false, merchantId: null })
  const [rejectReason, setRejectReason] = React.useState('')
  const [deleteDialog, setDeleteDialog] = React.useState<{ open: boolean; merchant: any | null }>({ open: false, merchant: null })
  const [page, setPage] = React.useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'merchants', { page, perPage: PER_PAGE }],
    queryFn: () => api.get<{ data: any[]; meta: { total: number; page: number; perPage: number } }>(`/v1/merchants?page=${page}&perPage=${PER_PAGE}`),
    retry: false,
  })

  const merchants = data?.data || []
  const total = data?.meta?.total || 0
  const totalPages = Math.ceil(total / PER_PAGE)

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/merchants/${id}/approve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
      setSelectedMerchant(null)
    },
  })

  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.post(`/v1/merchants/${id}/reject`, { reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
      setSelectedMerchant(null)
      setRejectDialog({ open: false, merchantId: null })
      setRejectReason('')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/merchants/${id}/deactivate`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
      setDeleteDialog({ open: false, merchant: null })
    },
  })

  const handleRejectClick = (merchantId: string) => {
    setRejectDialog({ open: true, merchantId })
    setRejectReason('')
  }

  const handleRejectConfirm = () => {
    if (rejectDialog.merchantId) {
      rejectMutation.mutate({ id: rejectDialog.merchantId, reason: rejectReason || 'Does not meet requirements' })
    }
  }

  const handleDeleteClick = (merchant: any) => {
    setDeleteDialog({ open: true, merchant })
  }

  const handleDeleteConfirm = () => {
    if (deleteDialog.merchant) {
      deleteMutation.mutate(deleteDialog.merchant.id)
    }
  }

  return (
    <>
      <PageHeader
        title="Merchants"
        description="Manage merchant accounts and KYC approvals"
      />

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Business Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>KYC Status</TableHead>
                <TableHead>Status</TableHead>
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
                    <TableCell><StatusBadge status={m.status} /></TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(m.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={<Button variant="ghost" size="sm" />}
                        >
                          <MoreHorizontal className="size-4" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="min-w-[160px]">
                          <DropdownMenuItem className="whitespace-nowrap" onClick={() => setSelectedMerchant(m)}>
                            <Eye className="size-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          {(m.kycStatus === 'UNDER_REVIEW' || m.kycStatus === 'INSTANT_ACCESS') && (
                            <>
                              <DropdownMenuItem
                                onClick={() => approveMutation.mutate(m.id)}
                                disabled={approveMutation.isPending}
                              >
                                <CheckCircle2 className="size-4 mr-2" />
                                Approve
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleRejectClick(m.id)}
                                disabled={rejectMutation.isPending}
                              >
                                <XCircle className="size-4 mr-2" />
                                Reject
                              </DropdownMenuItem>
                            </>
                          )}
                          {m.status === 'ACTIVE' && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onClick={() => handleDeleteClick(m)}
                                disabled={deleteMutation.isPending}
                              >
                                <Trash2 className="size-4 mr-2" />
                                Deactivate
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

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between border-t px-4 py-3">
              <p className="text-sm text-muted-foreground">
                Showing {(page - 1) * PER_PAGE + 1}–{Math.min(page * PER_PAGE, total)} of {total} merchants
              </p>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  <ChevronLeft className="size-4" />
                  Previous
                </Button>
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                  .reduce<(number | 'ellipsis')[]>((acc, p, idx, arr) => {
                    if (idx > 0 && p - (arr[idx - 1] as number) > 1) {
                      acc.push('ellipsis')
                    }
                    acc.push(p)
                    return acc
                  }, [])
                  .map((item, idx) =>
                    item === 'ellipsis' ? (
                      <span key={`ellipsis-${idx}`} className="px-2 text-muted-foreground">...</span>
                    ) : (
                      <Button
                        key={item}
                        variant={page === item ? 'default' : 'outline'}
                        size="sm"
                        className="min-w-8"
                        onClick={() => setPage(item as number)}
                      >
                        {item}
                      </Button>
                    )
                  )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                >
                  Next
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Merchant Details Dialog */}
      <Dialog open={!!selectedMerchant} onOpenChange={(open) => !open && setSelectedMerchant(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Merchant Details</DialogTitle>
            <DialogDescription>{selectedMerchant?.businessName}</DialogDescription>
          </DialogHeader>
          {selectedMerchant && (
            <div className="space-y-3 text-sm">
              <DetailRow label="Business Name" value={selectedMerchant.businessName} />
              <DetailRow label="Contact Email" value={selectedMerchant.contactEmail} />
              <DetailRow label="Contact Name" value={selectedMerchant.contactName || '-'} />
              <DetailRow label="Contact Phone" value={selectedMerchant.contactPhone || '-'} />
              <DetailRow label="Business Type" value={selectedMerchant.businessType || '-'} />
              <DetailRow label="Registration No" value={selectedMerchant.registrationNo || '-'} />
              <DetailRow label="Address" value={[selectedMerchant.addressLine1, selectedMerchant.addressLine2, selectedMerchant.city, selectedMerchant.postalCode].filter(Boolean).join(', ') || '-'} />
              <DetailRow label="Bank Name" value={selectedMerchant.bankName || '-'} />
              <DetailRow label="Bank Branch" value={selectedMerchant.bankBranch || '-'} />
              <DetailRow label="Bank Account" value={selectedMerchant.bankAccountNo || '-'} />
              <DetailRow label="Account Name" value={selectedMerchant.bankAccountName || '-'} />
              <div className="flex items-center justify-between pt-2 border-t">
                <span className="text-muted-foreground">KYC Status</span>
                <StatusBadge status={selectedMerchant.kycStatus} />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Status</span>
                <StatusBadge status={selectedMerchant.status} />
              </div>
            </div>
          )}
          <DialogFooter className="gap-2">
            {selectedMerchant && (selectedMerchant.kycStatus === 'UNDER_REVIEW' || selectedMerchant.kycStatus === 'INSTANT_ACCESS') && (
              <>
                <Button
                  className="bg-green-600 hover:bg-green-700 text-white"
                  onClick={() => approveMutation.mutate(selectedMerchant.id)}
                  disabled={approveMutation.isPending}
                >
                  <CheckCircle2 className="size-4 mr-1" />
                  Approve
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => {
                    setSelectedMerchant(null)
                    handleRejectClick(selectedMerchant.id)
                  }}
                  disabled={rejectMutation.isPending}
                >
                  <XCircle className="size-4 mr-1" />
                  Reject
                </Button>
              </>
            )}
            <Button variant="outline" onClick={() => setSelectedMerchant(null)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reject Reason Dialog */}
      <Dialog open={rejectDialog.open} onOpenChange={(open) => { if (!open) { setRejectDialog({ open: false, merchantId: null }); setRejectReason(''); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject Merchant</DialogTitle>
            <DialogDescription>Please provide a reason for rejecting this merchant.</DialogDescription>
          </DialogHeader>
          <div className="py-2">
            <Input
              placeholder="Enter rejection reason..."
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setRejectDialog({ open: false, merchantId: null }); setRejectReason(''); }}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRejectConfirm} disabled={rejectMutation.isPending}>
              {rejectMutation.isPending ? 'Rejecting...' : 'Reject Merchant'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialog.open} onOpenChange={(open) => !open && setDeleteDialog({ open: false, merchant: null })}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Deactivate Merchant</DialogTitle>
            <DialogDescription>
              Are you sure you want to deactivate <strong>{deleteDialog.merchant?.businessName}</strong>? This will disable their account and prevent them from processing payments.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialog({ open: false, merchant: null })}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm} disabled={deleteMutation.isPending}>
              {deleteMutation.isPending ? 'Deactivating...' : 'Deactivate Merchant'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}
