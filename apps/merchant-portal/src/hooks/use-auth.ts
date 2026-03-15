import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { api } from '#/lib/api'
import { setTokens, clearTokens, isAuthenticated } from '#/lib/auth'

interface AuthResponse {
  data: {
    accessToken: string
    refreshToken: string
    user: {
      id: string
      email: string
      name: string
      role: string
      isActive: boolean
      branchId?: string
    }
    merchant: {
      id: string
      businessName: string
      contactEmail: string
      kycStatus: string
      status: string
      [key: string]: unknown
    }
  }
}

interface MeResponse {
  data: {
    user: AuthResponse['data']['user']
    merchant: AuthResponse['data']['merchant']
  }
}

export function useLogin() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: { email: string; password: string }) =>
      api.post<AuthResponse>('/v1/auth/login', data),
    onSuccess: (res) => {
      setTokens(res.data.accessToken, res.data.refreshToken)
      queryClient.setQueryData(['auth', 'me'], { data: { user: res.data.user, merchant: res.data.merchant } })
      navigate({ to: '/' })
    },
  })
}

export function useRegister() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: { businessName: string; email: string; password: string; name: string; phone?: string }) =>
      api.post<AuthResponse>('/v1/auth/register', data),
    onSuccess: (res) => {
      setTokens(res.data.accessToken, res.data.refreshToken)
      queryClient.setQueryData(['auth', 'me'], { data: { user: res.data.user, merchant: res.data.merchant } })
      navigate({ to: '/activate' })
    },
  })
}

export function useMe() {
  return useQuery({
    queryKey: ['auth', 'me'],
    queryFn: () => api.get<MeResponse>('/v1/auth/me'),
    enabled: isAuthenticated(),
    staleTime: 5 * 60 * 1000,
  })
}

export function useLogout() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  return () => {
    clearTokens()
    queryClient.clear()
    navigate({ to: '/login' })
  }
}
