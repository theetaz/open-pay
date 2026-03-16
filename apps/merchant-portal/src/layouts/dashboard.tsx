import { useEffect } from 'react'
import { Link, Outlet, useNavigate } from 'react-router-dom'
import { SidebarProvider, SidebarInset } from '#/components/ui/sidebar'
import { TooltipProvider } from '#/components/ui/tooltip'
import { Button } from '#/components/ui/button'
import { AppSidebar } from '#/components/layout/app-sidebar'
import { SiteHeader } from '#/components/layout/site-header'
import { useAuthStore } from '#/stores/auth'
import { useSessionValidation } from '#/hooks/use-session-validation'
import { useMe } from '#/hooks/use-auth'
import { AlertTriangle, Clock } from 'lucide-react'

export function DashboardLayout() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const navigate = useNavigate()

  // Validate session against backend on mount - catches stale tokens after DB wipe
  useSessionValidation()

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
            <KycBanner />
            <main className="flex-1 p-6">
              <Outlet />
            </main>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}

function KycBanner() {
  const { data: meData } = useMe()
  const merchant = meData?.data?.merchant

  if (!merchant) return null

  if (merchant.kycStatus === 'PENDING' || merchant.kycStatus === 'REJECTED') {
    return (
      <div className={`mx-6 mt-4 rounded-lg border p-4 flex items-center justify-between ${
        merchant.kycStatus === 'REJECTED'
          ? 'bg-red-600/10 border-red-500/30'
          : 'bg-red-500/10 border-red-500/20'
      }`}>
        <div className="flex items-center gap-3">
          <AlertTriangle className="size-5 flex-shrink-0 text-red-600 dark:text-red-400" />
          <div>
            <p className="text-sm font-medium text-red-700 dark:text-red-300">
              {merchant.kycStatus === 'REJECTED'
                ? 'Your KYC verification was rejected.'
                : 'Your account is not verified yet.'}
            </p>
            <p className="text-xs text-red-600/80 dark:text-red-400/80 mt-0.5">
              {merchant.kycStatus === 'REJECTED'
                ? 'Please update your documents and resubmit for verification.'
                : 'Complete KYC verification to unlock full payment processing features.'}
            </p>
          </div>
        </div>
        <Link to="/activate">
          <Button size="sm" variant="destructive">
            {merchant.kycStatus === 'REJECTED' ? 'Resubmit KYC' : 'Verify Now'}
          </Button>
        </Link>
      </div>
    )
  }

  if (merchant.kycStatus === 'UNDER_REVIEW') {
    return (
      <div className="mx-6 mt-4 rounded-lg bg-blue-500/10 border border-blue-500/20 p-4 flex items-center gap-3">
        <Clock className="size-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
        <p className="text-sm text-blue-700 dark:text-blue-300">
          Your KYC verification is under review. We'll notify you once it's approved.
        </p>
      </div>
    )
  }

  return null
}
