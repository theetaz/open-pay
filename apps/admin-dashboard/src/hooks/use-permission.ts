import { useAuthStore } from '#/stores/auth'

/**
 * Check if the current admin user has a specific permission.
 * Returns true if the user's role includes the permission.
 */
export function usePermission(permission: string): boolean {
  const user = useAuthStore((s) => s.user)
  return user?.role?.permissions?.includes(permission) ?? false
}

/**
 * Check if the current admin user has any of the given permissions.
 */
export function useAnyPermission(...permissions: string[]): boolean {
  const user = useAuthStore((s) => s.user)
  if (!user?.role?.permissions) return false
  return permissions.some((p) => user.role.permissions.includes(p))
}
