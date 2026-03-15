import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '#/lib/api'

interface Payment {
  id: string
  merchantId: string
  paymentNo: string
  merchantTradeNo: string
  amount: string
  currency: string
  amountUsdt: string
  exchangeRate?: string
  exchangeFeePct: string
  exchangeFeeUsdt: string
  platformFeePct: string
  platformFeeUsdt: string
  totalFeesUsdt: string
  netAmountUsdt: string
  provider: string
  providerPayId: string
  qrContent: string
  checkoutLink: string
  deepLink: string
  status: string
  customerEmail: string
  txHash: string
  expireTime: string
  paidAt?: string
  createdAt: string
  branchId?: string
}

interface PaymentsResponse {
  data: Payment[]
  meta: {
    total: number
    page: number
    perPage: number
  }
}

interface PaymentResponse {
  data: Payment
}

interface CheckoutResponse {
  data: {
    id: string
    paymentNo: string
    amount: string
    currency: string
    amountUsdt: string
    qrContent: string
    deepLink: string
    status: string
    exchangeRate?: string
    expireTime: string
    paidAt?: string
    createdAt: string
  }
}

export function usePayments(params: { page?: number; perPage?: number; status?: string } = {}) {
  const queryString = new URLSearchParams()
  if (params.page) queryString.set('page', String(params.page))
  if (params.perPage) queryString.set('perPage', String(params.perPage))
  if (params.status) queryString.set('status', params.status)

  const qs = queryString.toString()
  const path = `/v1/payments${qs ? `?${qs}` : ''}`

  return useQuery({
    queryKey: ['payments', params],
    queryFn: () => api.get<PaymentsResponse>(path),
  })
}

export function usePayment(id: string) {
  return useQuery({
    queryKey: ['payments', id],
    queryFn: () => api.get<PaymentResponse>(`/v1/payments/${id}`),
    enabled: !!id,
  })
}

export function useCheckout(id: string) {
  return useQuery({
    queryKey: ['checkout', id],
    queryFn: () => api.get<CheckoutResponse>(`/v1/payments/${id}/checkout`),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.data?.status
      if (status === 'PAID' || status === 'EXPIRED' || status === 'FAILED') return false
      return 3000
    },
  })
}

export function useCreatePayment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: {
      amount: string
      currency?: string
      provider?: string
      merchantTradeNo?: string
      customerEmail?: string
    }) => api.post<PaymentResponse>('/v1/payments', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payments'] })
    },
  })
}
