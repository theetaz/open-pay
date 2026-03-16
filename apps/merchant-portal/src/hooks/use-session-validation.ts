import { useEffect, useRef } from 'react'
import { toast } from 'sonner'
import { isAuthenticated } from '#/lib/auth'
import { useAuthStore } from '#/stores/auth'
import { api } from '#/lib/api'
import { useQueryClient } from '@tanstack/react-query'

/**
 * Validates the current session against the backend on mount.
 * If the token is locally valid but the user/merchant no longer exists
 * (e.g. after a database wipe), this forces a logout.
 */
export function useSessionValidation() {
  const queryClient = useQueryClient()
  const validated = useRef(false)

  useEffect(() => {
    if (validated.current) return
    if (!isAuthenticated()) return

    validated.current = true

    api.get('/v1/auth/me').catch(() => {
      // Backend rejected the token or user no longer exists — force logout
      useAuthStore.getState().logout()
      queryClient.clear()
      toast.error('Your session is no longer valid. Please log in again.')
    })
  }, [queryClient])
}
