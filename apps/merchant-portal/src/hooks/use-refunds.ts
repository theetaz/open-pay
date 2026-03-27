import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '#/lib/api'

export interface Refund {
  id: string
  merchantId: string
  paymentId: string
  paymentNo: string
  amountUsdt: string
  reason: string
  status: string
  approvedBy?: string
  approvedAt?: string
  rejectedReason?: string
  completedAt?: string
  createdAt: string
}

export function useRefunds() {
  return useQuery({
    queryKey: ['refunds'],
    queryFn: () => api.get<{ data: Refund[] }>('/v1/refunds'),
  })
}

export function useRequestRefund() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: {
      paymentId: string
      paymentNo: string
      amountUsdt: string
      reason: string
    }) => api.post<{ data: Refund }>('/v1/refunds', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['refunds'] })
      queryClient.invalidateQueries({ queryKey: ['settlements'] })
    },
  })
}

export function useApproveRefund() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.post(`/v1/refunds/${id}/approve`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['refunds'] }),
  })
}

export function useRejectRefund() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.post(`/v1/refunds/${id}/reject`, { reason }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['refunds'] }),
  })
}
