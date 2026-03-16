import { create } from 'zustand'
import { clearTokens, setTokens, isAuthenticated as checkAuth } from '#/lib/auth'

interface AuthState {
  isAuthenticated: boolean
  login: (accessToken: string, refreshToken: string) => void
  logout: () => void
  checkAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: checkAuth(),
  login: (accessToken, refreshToken) => {
    setTokens(accessToken, refreshToken)
    set({ isAuthenticated: true })
  },
  logout: () => {
    clearTokens()
    set({ isAuthenticated: false })
  },
  checkAuth: () => {
    set({ isAuthenticated: checkAuth() })
  },
}))
