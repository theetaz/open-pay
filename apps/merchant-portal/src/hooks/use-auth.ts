import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { api } from '#/lib/api'
import { isAuthenticated } from '#/lib/auth'
import { useAuthStore } from '#/stores/auth'

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
      defaultCurrency: string
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
      useAuthStore.getState().login(res.data.accessToken, res.data.refreshToken)
      queryClient.setQueryData(['auth', 'me'], { data: { user: res.data.user, merchant: res.data.merchant } })
      toast.success('Welcome back!')
      navigate('/')
    },
    onError: (err) => {
      toast.error(err.message)
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
      useAuthStore.getState().login(res.data.accessToken, res.data.refreshToken)
      queryClient.setQueryData(['auth', 'me'], { data: { user: res.data.user, merchant: res.data.merchant } })
      toast.success('Account created! You can explore the portal and complete KYC when ready.')
      navigate('/')
    },
    onError: (err) => {
      toast.error(err.message)
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
    useAuthStore.getState().logout()
    queryClient.clear()
    navigate('/login')
  }
}
