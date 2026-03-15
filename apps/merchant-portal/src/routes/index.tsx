import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: DashboardPage })

function DashboardPage() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1">
        <header className="border-b border-border bg-card px-6 py-4">
          <h1 className="text-lg font-semibold">Dashboard</h1>
        </header>
        <main className="p-6">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard title="Total Payments" value="0" subtitle="+0% from last month" />
            <StatCard title="Revenue (USDT)" value="0.00" subtitle="Net amount after fees" />
            <StatCard title="Pending Settlements" value="0.00 LKR" subtitle="Ready to withdraw" />
            <StatCard title="Active API Keys" value="0" subtitle="Live + test keys" />
          </div>

          <div className="mt-8 grid gap-4 lg:grid-cols-2">
            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold mb-4">Recent Payments</h3>
              <p className="text-sm text-muted-foreground">
                No payments yet. Integrate via API or create a payment link to get started.
              </p>
            </div>
            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold mb-4">Quick Actions</h3>
              <div className="space-y-2">
                <QuickAction label="Create Payment Link" href="/payments" />
                <QuickAction label="Generate API Key" href="/settings" />
                <QuickAction label="Configure Webhook" href="/settings" />
              </div>
            </div>
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
      <nav className="px-3 space-y-1">
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

function StatCard({ title, value, subtitle }: { title: string; value: string; subtitle: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <p className="text-sm font-medium text-muted-foreground">{title}</p>
      <p className="text-2xl font-bold mt-1">{value}</p>
      <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
    </div>
  )
}

function QuickAction({ label, href }: { label: string; href: string }) {
  return (
    <Link
      to={href}
      className="block w-full text-left px-4 py-2 rounded-md border border-border text-sm hover:bg-accent transition-colors"
    >
      {label}
    </Link>
  )
}
