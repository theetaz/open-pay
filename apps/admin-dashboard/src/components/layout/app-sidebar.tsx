import { Link, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Building2,
  ArrowDownToLine,
  ScrollText,
  Landmark,
  Activity,
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

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  const location = useLocation()
  const currentPath = location.pathname

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
        <SidebarGroup>
          <SidebarGroupLabel>Management</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => {
                const isActive = item.exact
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

      </SidebarContent>

      <SidebarRail />
    </Sidebar>
  )
}
