import { useQuery } from '@tanstack/react-query'
import { api } from '#/lib/api'

interface RevenuePoint {
  date: string
  revenue: number
  count: number
}

interface ConversionData {
  total: number
  paid: number
  expired: number
  failed: number
  pending: number
  successRate: number
}

interface ProviderData {
  provider: string
  count: number
  paid: number
  volume: number
  successRate: number
}

export function useRevenueAnalytics(days = 30) {
  return useQuery({
    queryKey: ['analytics-revenue', days],
    queryFn: () => api.get<{ data: RevenuePoint[] }>(`/v1/analytics/revenue?days=${days}`),
  })
}

export function useConversionAnalytics(days = 30) {
  return useQuery({
    queryKey: ['analytics-conversion', days],
    queryFn: () => api.get<{ data: ConversionData }>(`/v1/analytics/conversion?days=${days}`),
  })
}

export function useProviderAnalytics(days = 30) {
  return useQuery({
    queryKey: ['analytics-providers', days],
    queryFn: () => api.get<{ data: ProviderData[] }>(`/v1/analytics/providers?days=${days}`),
  })
}
