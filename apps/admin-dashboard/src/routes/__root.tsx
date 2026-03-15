import { HeadContent, Outlet, Scripts, createRootRoute } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { SidebarProvider, SidebarInset } from '#/components/ui/sidebar'
import { TooltipProvider } from '#/components/ui/tooltip'
import { AppSidebar } from '#/components/layout/app-sidebar'
import { SiteHeader } from '#/components/layout/site-header'
import { queryClient } from '#/lib/query'
import appCss from '../styles.css?url'

const THEME_INIT_SCRIPT = `(function(){try{var stored=localStorage.getItem('theme');var prefersDark=window.matchMedia('(prefers-color-scheme:dark)').matches;var resolved=stored==='dark'?'dark':stored==='light'?'light':prefersDark?'dark':'light';document.documentElement.classList.add(resolved);document.documentElement.style.colorScheme=resolved;}catch(e){}})();`

export const Route = createRootRoute({
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'Open Pay | Admin Dashboard' },
      { name: 'description', content: 'Admin management dashboard for Open Pay platform' },
    ],
    links: [
      { rel: 'stylesheet', href: appCss },
    ],
  }),
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
        <HeadContent />
      </head>
      <body className="min-h-screen bg-background font-mono antialiased" suppressHydrationWarning>
        <QueryClientProvider client={queryClient}>
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
                    {children}
                    <Outlet />
                  </main>
                </div>
              </SidebarInset>
            </SidebarProvider>
          </TooltipProvider>
        </QueryClientProvider>
        <Scripts />
      </body>
    </html>
  )
}
