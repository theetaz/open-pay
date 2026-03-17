import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Building2,
  ArrowDownToLine,
  ScrollText,
  Landmark,
  Activity,
  Settings2,
  Receipt,
  Mail,
  Users,
  Shield,
  FileText,
  MailPlus,
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '#/components/ui/sidebar'

const navItems = [
  { title: 'Dashboard', href: '/', icon: LayoutDashboard, exact: true },
  { title: 'Merchants', href: '/merchants', icon: Building2 },
  { title: 'Withdrawals', href: '/withdrawals', icon: ArrowDownToLine },
  { title: 'Treasury', href: '/treasury', icon: Landmark },
  { title: 'Audit Logs', href: '/audit-logs', icon: ScrollText },
  { title: 'System Health', href: '/system-health', icon: Activity },
]

const settingsItems = [
  { title: 'General', href: '/settings/general', icon: Settings2 },
  { title: 'Fees & Pricing', href: '/settings/fees', icon: Receipt },
  { title: 'Email', href: '/settings/email', icon: Mail },
  { title: 'Team', href: '/settings/team', icon: Users },
  { title: 'Roles', href: '/settings/roles', icon: Shield },
  { title: 'Legal Documents', href: '/settings/legal-documents', icon: FileText },
  { title: 'Email Templates', href: '/settings/email-templates', icon: MailPlus },
]

function NavGroup({ label, items }: { label: string; items: typeof navItems }) {
  const location = useLocation()
  const currentPath = location.pathname

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{label}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => {
            const isActive = 'exact' in item && item.exact
              ? currentPath === item.href
              : currentPath.startsWith(item.href)

            return (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  tooltip={item.title}
                  isActive={isActive}
                  render={<Link to={item.href} />}
                >
                  <item.icon className="size-4" />
                  <span>{item.title}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar {...props}>
      <SidebarHeader className="p-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
            OP
          </div>
          <div className="flex flex-col">
            <span className="text-lg font-semibold tracking-tight leading-none">Open Pay</span>
            <span className="text-[10px] text-muted-foreground tracking-widest uppercase">Admin</span>
          </div>
        </Link>
      </SidebarHeader>

      <SidebarContent>
        <NavGroup label="Management" items={navItems} />
        <NavGroup label="Settings" items={settingsItems} />
      </SidebarContent>

      <SidebarRail />
    </Sidebar>
  )
}
