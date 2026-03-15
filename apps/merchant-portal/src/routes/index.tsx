import { createFileRoute, Link } from '@tanstack/react-router'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Badge } from '#/components/ui/badge'
import { Separator } from '#/components/ui/separator'
import { ThemeToggle } from '#/components/theme-toggle'

export const Route = createFileRoute('/')({ component: DashboardPage })

function DashboardPage() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1">
        <header className="border-b border-border bg-card px-6 py-4 flex items-center justify-between">
          <h1 className="text-lg font-semibold">Dashboard</h1>
          <div className="flex items-center gap-2">
            <Badge variant="outline">Live</Badge>
            <ThemeToggle />
          </div>
        </header>
        <main className="p-6">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard title="Total Payments" value="0" description="+0% from last month" />
            <StatCard title="Revenue (USDT)" value="0.00" description="Net amount after fees" />
            <StatCard title="Pending Settlements" value="0.00 LKR" description="Ready to withdraw" />
            <StatCard title="Active API Keys" value="0" description="Live + test keys" />
          </div>

          <div className="mt-8 grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Recent Payments</CardTitle>
                <CardDescription>Your latest transaction activity</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  No payments yet. Integrate via API or create a payment link to get started.
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common tasks to get started</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button variant="outline" className="w-full justify-start" asChild>
                  <Link to="/payments">Create Payment Link</Link>
                </Button>
                <Button variant="outline" className="w-full justify-start" asChild>
                  <Link to="/settings">Generate API Key</Link>
                </Button>
                <Button variant="outline" className="w-full justify-start" asChild>
                  <Link to="/settings">Configure Webhook</Link>
                </Button>
              </CardContent>
            </Card>
          </div>
        </main>
      </div>
    </div>
  )
}

function Sidebar() {
  return (
    <aside className="w-64 border-r border-border bg-card min-h-screen hidden md:block">
      <div className="p-6">
        <h2 className="text-xl font-bold text-primary">Open Pay</h2>
        <p className="text-xs text-muted-foreground mt-1">Merchant Portal</p>
      </div>
      <Separator />
      <nav className="px-3 py-3 space-y-1">
        <NavLink href="/" label="Dashboard" />
        <NavLink href="/payments" label="Payments" />
        <NavLink href="/settlements" label="Settlements" />
        <NavLink href="/settings" label="Settings" />
      </nav>
    </aside>
  )
}

function NavLink({ href, label }: { href: string; label: string }) {
  return (
    <Link
      to={href}
      className="block px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
      activeProps={{ className: 'bg-primary/10 text-primary font-medium' }}
    >
      {label}
    </Link>
  )
}

function StatCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-2xl">{value}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}
