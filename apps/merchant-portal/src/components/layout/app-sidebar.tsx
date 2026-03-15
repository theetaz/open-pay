import { Link } from '@tanstack/react-router'
import {
  CreditCard,
  Link2,
  ArrowDownToLine,
  Building2,
  Users,
  ScrollText,
  Settings,
  FileText,
  HelpCircle,
  ExternalLink,
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '#/components/ui/sidebar'

const navItems = [
  { title: 'Payments', href: '/payments', icon: CreditCard },
  { title: 'Payment Links', href: '/payment-links', icon: Link2 },
  { title: 'Withdrawal', href: '/withdrawal', icon: ArrowDownToLine },
  { title: 'Branches', href: '/branches', icon: Building2 },
  { title: 'Users', href: '/users', icon: Users },
  { title: 'Audit Log', href: '/audit-log', icon: ScrollText },
  { title: 'Settings', href: '/settings', icon: Settings },
]

const bottomLinks = [
  { title: 'Documentation', href: '/docs', icon: FileText, external: true },
  { title: 'Support', href: '/support', icon: HelpCircle, external: true },
]

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarHeader className="p-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
            OP
          </div>
          <span className="text-lg font-bold">Open Pay</span>
        </Link>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <Link
                      to={item.href}
                      activeProps={{ className: 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' }}
                    >
                      <item.icon className="size-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          {bottomLinks.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton asChild>
                <a href={item.href} target="_blank" rel="noopener noreferrer" className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <item.icon className="size-4" />
                    <span>{item.title}</span>
                  </span>
                  <ExternalLink className="size-3 text-muted-foreground" />
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}
