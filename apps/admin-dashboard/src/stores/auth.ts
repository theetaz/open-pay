import { create } from 'zustand'
import { setTokens, clearTokens, isAuthenticated as checkAuth } from '#/lib/auth'

interface AdminUser {
  id: string
  email: string
  name: string
  mustChangePassword?: boolean
  role: {
    name: string
    permissions: string[]
  }
}

interface AuthState {
  isAuthenticated: boolean
  user: AdminUser | null
  login: (accessToken: string, refreshToken: string, user?: AdminUser) => void
  logout: () => void
  setUser: (user: AdminUser) => void
  checkAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: checkAuth(),
  user: null,
  login: (accessToken, refreshToken, user) => {
    setTokens(accessToken, refreshToken)
    set({ isAuthenticated: true, user: user || null })
  },
  logout: () => {
    clearTokens()
    set({ isAuthenticated: false, user: null })
  },
  setUser: (user) => {
    set({ user })
  },
  checkAuth: () => {
    set({ isAuthenticated: checkAuth() })
  },
}))
