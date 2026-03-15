import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { CheckCircle2, XCircle, BanknoteIcon, Clock, ArrowDownToLine } from 'lucide-react'
import { api } from '#/lib/api'

export const Route = createFileRoute('/_dashboard/withdrawals')({ component: WithdrawalsPage })

function WithdrawalsPage() {
  const queryClient = useQueryClient()

  const { data } = useQuery({
    queryKey: ['admin', 'withdrawals'],
    queryFn: () => api.get<{ data: any[] }>('/v1/withdrawals'),
    retry: false,
  })

  const withdrawals = data?.data || []
  const pending = withdrawals.filter((w: any) => w.status === 'REQUESTED')
  const approved = withdrawals.filter((w: any) => w.status === 'APPROVED')
  const completed = withdrawals.filter((w: any) => w.status === 'COMPLETED')
  const totalSettledLKR = completed.reduce((sum: number, w: any) => sum + parseFloat(w.amountLkr || '0'), 0)

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/approve`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] }),
  })

  const rejectMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/reject`, { reason: 'Rejected by admin' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] }),
  })

  const completeMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/withdrawals/${id}/complete`, { bankReference: `TXN${Date.now()}` }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'withdrawals'] }),
  })

  return (
    <>
      <PageHeader
        title="Withdrawal Approvals"
        description="Review and process merchant withdrawal requests"
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Pending" value={String(pending.length)} description="Awaiting approval" icon={Clock} valueClassName={pending.length > 0 ? 'text-amber-500' : ''} />
        <StatCard title="Approved" value={String(approved.length)} description="Ready for bank transfer" icon={ArrowDownToLine} />
        <StatCard title="Total Settled (LKR)" value={totalSettledLKR.toFixed(2)} description="Bank transfers completed" icon={BanknoteIcon} />
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Amount (USDT)</TableHead>
                <TableHead>Rate</TableHead>
                <TableHead>Amount (LKR)</TableHead>
                <TableHead>Bank</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {withdrawals.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7}>
                    <EmptyState message="No withdrawal requests." />
                  </TableCell>
                </TableRow>
              ) : (
                withdrawals.map((w: any) => (
                  <TableRow key={w.id}>
                    <TableCell className="text-sm">{new Date(w.createdAt).toLocaleDateString()}</TableCell>
                    <TableCell className="font-medium">{w.amountUsdt}</TableCell>
                    <TableCell className="text-sm">{w.exchangeRate}</TableCell>
                    <TableCell className="font-medium">{w.amountLkr}</TableCell>
                    <TableCell className="text-sm">{w.bankName}<br /><span className="text-muted-foreground">{w.bankAccountNo}</span></TableCell>
                    <TableCell><StatusBadge status={w.status} /></TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {w.status === 'REQUESTED' && (
                          <>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-green-600 hover:text-green-700"
                              onClick={() => approveMutation.mutate(w.id)}
                              disabled={approveMutation.isPending}
                            >
                              <CheckCircle2 className="size-4 mr-1" /> Approve
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-red-600 hover:text-red-700"
                              onClick={() => rejectMutation.mutate(w.id)}
                              disabled={rejectMutation.isPending}
                            >
                              <XCircle className="size-4 mr-1" /> Reject
                            </Button>
                          </>
                        )}
                        {w.status === 'APPROVED' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-blue-600 hover:text-blue-700"
                            onClick={() => completeMutation.mutate(w.id)}
                            disabled={completeMutation.isPending}
                          >
                            <BanknoteIcon className="size-4 mr-1" /> Complete
                          </Button>
                        )}
                        {w.status === 'COMPLETED' && (
                          <span className="text-xs text-muted-foreground font-mono">{w.bankReference}</span>
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
    </>
  )
}
