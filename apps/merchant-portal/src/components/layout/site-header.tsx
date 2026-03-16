import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { SidebarTrigger } from '#/components/ui/sidebar'
import { Separator } from '#/components/ui/separator'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu'
import { ThemeToggle } from '#/components/theme-toggle'
import { useMe } from '#/hooks/use-auth'
import { useAuthStore } from '#/stores/auth'
import { User, Settings, Shield, LogOut } from 'lucide-react'

export function SiteHeader() {
  const { data: meData } = useMe()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const user = meData?.data?.user
  const merchant = meData?.data?.merchant

  const initials = user?.name
    ? user.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'M'

  const handleLogout = () => {
    useAuthStore.getState().logout()
    queryClient.clear()
    navigate('/login')
  }

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b border-border bg-background px-4 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
      <SidebarTrigger className="-ml-1" />
      <Separator orientation="vertical" className="mr-2 h-4" />
      <div className="flex-1" />
      <ThemeToggle />

      <DropdownMenu>
        <DropdownMenuTrigger className="rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <Avatar className="size-8 cursor-pointer">
            <AvatarFallback className="bg-primary text-primary-foreground text-xs">
              {initials}
            </AvatarFallback>
          </Avatar>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-64" align="end" sideOffset={8}>
          <div className="px-1.5 py-2">
            <p className="text-sm font-medium leading-none">{user?.name || 'User'}</p>
            <p className="text-xs text-muted-foreground mt-1">{user?.email || ''}</p>
            {merchant?.businessName && (
              <p className="text-xs text-muted-foreground mt-0.5">{merchant.businessName}</p>
            )}
          </div>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem onClick={() => navigate('/profile')}>
              <User className="mr-2 size-4" />
              Profile
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate('/security')}>
              <Shield className="mr-2 size-4" />
              Security
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate('/settings')}>
              <Settings className="mr-2 size-4" />
              Settings
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout} variant="destructive">
            <LogOut className="mr-2 size-4" />
            Log out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
