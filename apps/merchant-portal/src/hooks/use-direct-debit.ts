import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '#/lib/api'

export interface ScenarioCode {
  id: string
  scenarioId: string
  scenarioName: string
  paymentProvider: string
  maxLimit: string
  isActive: boolean
}

export interface DirectDebitContract {
  id: string
  merchantId: string
  branchId?: string
  merchantContractCode: string
  serviceName: string
  scenarioId: string
  paymentProvider: string
  currency: string
  singleUpperLimit: string
  status: string
  qrContent: string
  deepLink: string
  contractId?: string
  openUserId?: string
  webhookUrl?: string
  periodic: boolean
  paymentCount: number
  totalAmountCharged: string
  lastPaymentAt?: string
  terminationTime?: string
  terminationNotes?: string
  createdAt: string
  updatedAt: string
}

export interface DirectDebitPayment {
  id: string
  contractId: string
  merchantId: string
  payId: string
  paymentNo: string
  amount: string
  currency: string
  status: string
  productName: string
  paymentProvider: string
  createdAt: string
  feeBreakdown: {
    grossAmountUSDT: string
    exchangeFeePercentage: string
    exchangeFeeAmountUSDT: string
    platformFeePercentage: string
    platformFeeAmountUSDT: string
    totalFeesUSDT: string
    netAmountUSDT: string
  }
}

interface ContractListResponse {
  data: DirectDebitContract[]
  meta: { page: number; limit: number; totalItems: number; totalPages: number }
}

export function useScenarioCodes(provider?: string) {
  return useQuery({
    queryKey: ['scenario-codes', provider],
    queryFn: () => {
      const params = new URLSearchParams()
      if (provider) params.set('provider', provider)
      const qs = params.toString()
      return api.get<{ data: ScenarioCode[] }>(`/v1/direct-debit/scenario-codes${qs ? `?${qs}` : ''}`)
    },
  })
}

export function useDirectDebitContracts(status?: string, page = 1, limit = 20) {
  return useQuery({
    queryKey: ['direct-debit-contracts', status, page, limit],
    queryFn: () => {
      const params = new URLSearchParams()
      if (status) params.set('status', status)
      params.set('page', String(page))
      params.set('limit', String(limit))
      return api.get<ContractListResponse>(`/v1/direct-debit/list?${params.toString()}`)
    },
  })
}

export function useDirectDebitContract(id: string) {
  return useQuery({
    queryKey: ['direct-debit-contract', id],
    queryFn: () => api.get<{ data: DirectDebitContract }>(`/v1/direct-debit/${id}`),
    enabled: !!id,
  })
}

export function useCreateContract() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: {
      serviceName: string
      scenarioId: string
      singleUpperLimit: string
      returnUrl: string
      cancelUrl: string
      merchantContractCode?: string
      branchId?: string
      webhookUrl?: string
    }) => api.post<{ data: DirectDebitContract }>('/v1/direct-debit', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['direct-debit-contracts'] }),
  })
}

export function useSyncContract() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      api.post<{ data: DirectDebitContract }>(`/v1/direct-debit/${id}/sync`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contracts'] })
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contract'] })
    },
  })
}

export function useTerminateContract() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, terminationNotes }: { id: string; terminationNotes?: string }) =>
      api.post<{ data: DirectDebitContract }>(`/v1/direct-debit/${id}/terminate`, { terminationNotes }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contracts'] })
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contract'] })
    },
  })
}

export function useExecutePayment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      contractId,
      ...data
    }: {
      contractId: string
      amount: string
      productName: string
      productDetail?: string
      webhookUrl?: string
      customerFirstName?: string
      customerLastName?: string
      customerEmail?: string
    }) => api.post<{ data: DirectDebitPayment }>(`/v1/direct-debit/${contractId}/payment`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contracts'] })
      queryClient.invalidateQueries({ queryKey: ['direct-debit-contract'] })
    },
  })
}
