import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '#/lib/api'

interface Balance {
  merchantId: string
  availableUsdt: string
  pendingUsdt: string
  totalEarnedUsdt: string
  totalWithdrawnUsdt: string
  totalFeesUsdt: string
  updatedAt: string
}

interface Withdrawal {
  id: string
  merchantId: string
  amountUsdt: string
  exchangeRate: string
  amountLkr: string
  bankName: string
  bankAccountNo: string
  bankAccountName: string
  status: string
  rejectedReason?: string
  bankReference?: string
  completedAt?: string
  createdAt: string
}

export function useBalance() {
  return useQuery({
    queryKey: ['settlements', 'balance'],
    queryFn: () => api.get<{ data: Balance }>('/v1/settlements/balance'),
  })
}

export function useWithdrawals() {
  return useQuery({
    queryKey: ['withdrawals'],
    queryFn: () => api.get<{ data: Withdrawal[] }>('/v1/withdrawals'),
  })
}

export function useRequestWithdrawal() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: {
      amountUsdt: string
      exchangeRate: string
      bankName: string
      bankAccountNo: string
      bankAccountName: string
    }) => api.post<{ data: Withdrawal }>('/v1/withdrawals', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settlements'] })
      queryClient.invalidateQueries({ queryKey: ['withdrawals'] })
    },
  })
}
