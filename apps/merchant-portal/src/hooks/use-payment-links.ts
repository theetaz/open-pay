import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import { isAuthenticated } from '#/lib/auth'

export interface PaymentLink {
  id: string
  merchantId: string
  branchId?: string
  name: string
  slug: string
  description: string
  currency: string
  amount: string
  allowCustomAmount: boolean
  isReusable: boolean
  showOnQrPage: boolean
  usageCount: number
  status: string
  expireAt?: string
  createdAt: string
  updatedAt: string
}

interface PaymentLinksResponse {
  data: PaymentLink[]
  meta: { total: number; page: number; perPage: number }
}

interface PaymentLinkResponse {
  data: PaymentLink
}

interface SlugCheckResponse {
  data: { available: boolean }
}

export function usePaymentLinks(params: { page?: number; perPage?: number } = {}) {
  const { page = 1, perPage = 20 } = params
  return useQuery({
    queryKey: ['payment-links', page, perPage],
    queryFn: () => api.get<PaymentLinksResponse>(`/v1/payment-links?page=${page}&perPage=${perPage}`),
    enabled: isAuthenticated(),
  })
}

export function usePaymentLink(id: string) {
  return useQuery({
    queryKey: ['payment-links', id],
    queryFn: () => api.get<PaymentLinkResponse>(`/v1/payment-links/${id}`),
    enabled: isAuthenticated() && !!id,
  })
}

export function useCreatePaymentLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: {
      name: string
      slug: string
      description?: string
      currency: string
      amount: string
      allowCustomAmount?: boolean
      isReusable?: boolean
      showOnQrPage?: boolean
      expireAt?: string
    }) => api.post<PaymentLinkResponse>('/v1/payment-links', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-links'] })
      toast.success('Payment link created successfully')
    },
    onError: (err) => {
      toast.error(err.message)
    },
  })
}

export function useDeletePaymentLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => api.delete(`/v1/payment-links/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-links'] })
      toast.success('Payment link deleted')
    },
    onError: (err) => {
      toast.error(err.message)
    },
  })
}

export function checkSlugAvailability(slug: string) {
  return api.get<SlugCheckResponse>(`/v1/payment-links/check-slug/${slug}`)
}

export function usePublicPaymentLink(slug: string) {
  return useQuery({
    queryKey: ['public-payment-link', slug],
    queryFn: () => api.get<PaymentLinkResponse>(`/v1/public/payment-links/by-slug/${slug}`),
    enabled: !!slug,
    retry: false,
  })
}
