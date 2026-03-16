import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { CheckCircle2, XCircle, Eye } from 'lucide-react'
import { api } from '#/lib/api'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'

export function MerchantsPage() {
  const queryClient = useQueryClient()
  const [selectedMerchant, setSelectedMerchant] = React.useState<any>(null)

  const { data } = useQuery({
    queryKey: ['admin', 'merchants'],
    queryFn: () => api.get<{ data: any[]; meta: { total: number } }>('/v1/merchants?perPage=50'),
    retry: false,
  })

  const merchants = data?.data || []

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
    },
  })

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
                    <EmptyState message="No merchants registered yet." />
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
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedMerchant(m)}
                        >
                          <Eye className="size-4" />
                        </Button>
                        {(m.kycStatus === 'UNDER_REVIEW' || m.kycStatus === 'INSTANT_ACCESS') && (
                          <>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-green-600 hover:text-green-700"
                              onClick={() => approveMutation.mutate(m.id)}
                              disabled={approveMutation.isPending}
                            >
                              <CheckCircle2 className="size-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-red-600 hover:text-red-700"
                              onClick={() => rejectMutation.mutate({ id: m.id, reason: 'Does not meet requirements' })}
                              disabled={rejectMutation.isPending}
                            >
                              <XCircle className="size-4" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

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
              <DetailRow label="City" value={selectedMerchant.city || '-'} />
              <DetailRow label="Bank Name" value={selectedMerchant.bankName || '-'} />
              <DetailRow label="Bank Account" value={selectedMerchant.bankAccountNo || '-'} />
              <div className="flex items-center justify-between pt-2 border-t">
                <span className="text-muted-foreground">KYC Status</span>
                <StatusBadge status={selectedMerchant.kycStatus} />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setSelectedMerchant(null)}>Close</Button>
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
