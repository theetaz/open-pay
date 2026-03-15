import { SidebarTrigger } from '#/components/ui/sidebar'
import { Separator } from '#/components/ui/separator'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { ThemeToggle } from '#/components/theme-toggle'

export function SiteHeader() {
  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b border-border bg-background px-4 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
      <SidebarTrigger className="-ml-1" />
      <Separator orientation="vertical" className="mr-2 h-4" />
      <div className="flex-1" />
      <ThemeToggle />
      <Avatar className="size-8">
        <AvatarFallback className="bg-primary text-primary-foreground text-xs">
          A
        </AvatarFallback>
      </Avatar>
    </header>
  )
}
