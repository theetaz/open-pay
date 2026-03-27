import { Link, useLocation } from 'react-router-dom'
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
  ShieldCheck,
  RefreshCw,
  Plug,
  RotateCcw,
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '#/components/ui/sidebar'

const navItems = [
  { title: 'Payments', href: '/payments', icon: CreditCard },
  { title: 'Payment Links', href: '/payment-links', icon: Link2 },
  { title: 'Subscriptions', href: '/subscriptions', icon: RefreshCw },
  { title: 'Withdrawal', href: '/withdrawal', icon: ArrowDownToLine },
  { title: 'Refunds', href: '/refunds', icon: RotateCcw },
  { title: 'Branches', href: '/branches', icon: Building2 },
  { title: 'Users', href: '/users', icon: Users },
  { title: 'Integrations', href: '/integrations', icon: Plug },
  { title: 'Audit Log', href: '/audit-log', icon: ScrollText },
  { title: 'Settings', href: '/settings', icon: Settings },
]

const bottomLinks = [
  { title: 'Documentation', href: '/docs', icon: FileText },
  { title: 'Support', href: '/support', icon: HelpCircle },
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
          <span className="text-lg font-semibold tracking-tight">Open Pay</span>
        </Link>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    tooltip={item.title}
                    isActive={currentPath.startsWith(item.href)}
                    render={<Link to={item.href} />}
                  >
                    <item.icon className="size-4" />
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>Account</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  tooltip="Activate"
                  isActive={currentPath.startsWith('/activate')}
                  render={<Link to="/activate" />}
                >
                  <ShieldCheck className="size-4" />
                  <span>Activate</span>
                </SidebarMenuButton>
                <SidebarMenuBadge className="bg-amber-500/10 text-amber-600 dark:text-amber-400">
                  KYC
                </SidebarMenuBadge>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          {bottomLinks.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                tooltip={item.title}
                render={
                  <a href={item.href} target="_blank" rel="noopener noreferrer" />
                }
              >
                <item.icon className="size-4" />
                <span>{item.title}</span>
                <ExternalLink className="ml-auto size-3 text-muted-foreground" />
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}
