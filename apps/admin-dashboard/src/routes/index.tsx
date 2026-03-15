import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: AdminDashboard })

function AdminDashboard() {
  return (
    <div className="flex min-h-screen">
      <AdminSidebar />
      <div className="flex-1">
        <header className="border-b border-border bg-card px-6 py-4 flex items-center justify-between">
          <h1 className="text-lg font-semibold">Admin Dashboard</h1>
          <span className="text-xs px-2 py-1 rounded-full bg-primary/10 text-primary font-medium">
            Internal
          </span>
        </header>
        <main className="p-6">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard title="Total Merchants" value="0" subtitle="Registered" />
            <StatCard title="Pending KYC" value="0" subtitle="Awaiting review" />
            <StatCard title="Today's Volume" value="0.00 USDT" subtitle="Payments processed" />
            <StatCard title="Pending Withdrawals" value="0" subtitle="Needs approval" />
          </div>

          <div className="mt-8 grid gap-4 lg:grid-cols-2">
            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold mb-4">Merchant Approval Queue</h3>
              <p className="text-sm text-muted-foreground">No merchants pending approval.</p>
            </div>
            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold mb-4">Withdrawal Approval Queue</h3>
              <p className="text-sm text-muted-foreground">No withdrawals pending approval.</p>
            </div>
          </div>

          <div className="mt-6 rounded-lg border border-border bg-card p-6">
            <h3 className="font-semibold mb-4">System Health</h3>
            <div className="grid gap-3 sm:grid-cols-3">
              <ServiceStatus name="Gateway" port="8080" />
              <ServiceStatus name="Payment" port="8081" />
              <ServiceStatus name="Merchant" port="8082" />
              <ServiceStatus name="Settlement" port="8083" />
              <ServiceStatus name="Webhook" port="8084" />
              <ServiceStatus name="Exchange" port="8085" />
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}

function AdminSidebar() {
  return (
    <aside className="w-64 border-r border-border bg-card min-h-screen hidden md:block">
      <div className="p-6">
        <h2 className="text-xl font-bold text-primary">Open Pay</h2>
        <p className="text-xs text-muted-foreground mt-1">Admin Console</p>
      </div>
      <nav className="px-3 space-y-1">
        <NavLink href="/" label="Dashboard" />
        <NavLink href="/merchants" label="Merchants" />
        <NavLink href="/withdrawals" label="Withdrawals" />
        <NavLink href="/audit-logs" label="Audit Logs" />
        <NavLink href="/treasury" label="Treasury" />
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

function ServiceStatus({ name, port }: { name: string; port: string }) {
  return (
    <div className="flex items-center gap-2 rounded-md border border-border px-3 py-2">
      <span className="h-2 w-2 rounded-full bg-muted-foreground" />
      <span className="text-sm font-medium">{name}</span>
      <span className="text-xs text-muted-foreground ml-auto">:{port}</span>
    </div>
  )
}
