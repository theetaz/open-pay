import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '#/lib/api'

interface SubscriptionPlan {
  id: string
  merchantId: string
  name: string
  description: string
  amount: string
  currency: string
  intervalType: string
  intervalCount: number
  trialDays: number
  status: string
  createdAt: string
}

interface Subscription {
  id: string
  planId: string
  merchantId: string
  subscriberEmail: string
  status: string
  nextBillingDate: string
  totalPaidUsdt: string
  billingCount: number
  trialEnd?: string
  cancelledAt?: string
  cancellationReason?: string
  createdAt: string
}

export function usePlans() {
  return useQuery({
    queryKey: ['subscription-plans'],
    queryFn: () => api.get<{ data: SubscriptionPlan[] }>('/v1/subscription-plans'),
  })
}

export function useSubscriptions() {
  return useQuery({
    queryKey: ['subscriptions'],
    queryFn: () => api.get<{ data: Subscription[] }>('/v1/subscriptions'),
  })
}

export function useCreatePlan() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: {
      name: string
      description?: string
      amount: string
      currency: string
      intervalType: string
      intervalCount: number
      trialDays?: number
    }) => api.post<{ data: SubscriptionPlan }>('/v1/subscription-plans', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['subscription-plans'] }),
  })
}

export function useArchivePlan() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.post(`/v1/subscription-plans/${id}/archive`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['subscription-plans'] }),
  })
}

export function useCancelSubscription() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.post(`/v1/subscriptions/${id}/cancel`, { reason }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['subscriptions'] }),
  })
}
