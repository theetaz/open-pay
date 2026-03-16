import { useEffect } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { SidebarProvider, SidebarInset } from '#/components/ui/sidebar'
import { TooltipProvider } from '#/components/ui/tooltip'
import { AppSidebar } from '#/components/layout/app-sidebar'
import { SiteHeader } from '#/components/layout/site-header'
import { useAuthStore } from '#/stores/auth'

export function DashboardLayout() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const navigate = useNavigate()

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { replace: true })
    }
  }, [isAuthenticated, navigate])

  if (!isAuthenticated) {
    return null
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
