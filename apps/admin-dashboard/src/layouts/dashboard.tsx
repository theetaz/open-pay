import { Outlet, Navigate, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { SidebarProvider, SidebarInset } from '#/components/ui/sidebar'
import { TooltipProvider } from '#/components/ui/tooltip'
import { AppSidebar } from '#/components/layout/app-sidebar'
import { SiteHeader } from '#/components/layout/site-header'
import { useAuthStore } from '#/stores/auth'
import { api } from '#/lib/api'

export function DashboardLayout() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const location = useLocation()

  // Fetch current user profile on mount
  useQuery({
    queryKey: ['admin', 'me'],
    queryFn: async () => {
      const res = await api.get<{ data: any }>('/v1/admin/auth/me')
      const u = res.data
      if (u) {
        setUser({
          id: u.id,
          email: u.email,
          name: u.name,
          mustChangePassword: u.mustChangePassword,
          role: u.role || { name: 'ADMIN', permissions: [] },
        })
      }
      return res
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
    retry: false,
  })

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  // Force password change if required
  if (user?.mustChangePassword && location.pathname !== '/change-password') {
    return <Navigate to="/change-password" replace />
  }

  return (
    <TooltipProvider>
      <SidebarProvider
        style={
          {
            '--sidebar-width': 'calc(var(--spacing) * 72)',
            '--header-height': 'calc(var(--spacing) * 12)',
          } as React.CSSProperties
        }
      >
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex flex-1 flex-col">
            <main className="flex-1 p-6">
              <Outlet />
            </main>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}
