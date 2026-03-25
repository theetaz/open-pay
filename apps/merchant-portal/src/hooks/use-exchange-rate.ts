import { useQuery } from '@tanstack/react-query'
import { api } from '#/lib/api'

interface ExchangeRate {
  base: string
  quote: string
  rate: string
  spread: string
  effectiveRate: string
  updatedAt: string
}

export function useExchangeRate(base = 'USDT', quote = 'LKR') {
  return useQuery({
    queryKey: ['exchange-rate', base, quote],
    queryFn: () =>
      api.get<{ data: ExchangeRate }>(`/v1/exchange-rates/active?base=${base}&quote=${quote}`),
    staleTime: 60_000, // cache for 1 minute
  })
}
