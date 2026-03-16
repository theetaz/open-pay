import { useMe } from '#/hooks/use-auth'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { Badge } from '#/components/ui/badge'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { CopyButton } from '#/components/dashboard/copy-button'
import { Skeleton } from '#/components/ui/skeleton'

export function ProfilePage() {
  const { data: meData, isLoading } = useMe()
  const user = meData?.data?.user
  const merchant = meData?.data?.merchant

  if (isLoading) {
    return (
      <>
        <PageHeader title="Profile" description="Your account and business details" />
        <div className="space-y-6">
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </>
    )
  }

  const initials = user?.name
    ? user.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'M'

  return (
    <>
      <PageHeader title="Profile" description="Your account and business details" />

      <div className="space-y-6">
        {/* User Profile Card */}
        <Card>
          <CardHeader>
            <CardTitle>Account Information</CardTitle>
            <CardDescription>Your personal account details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-start gap-6">
              <Avatar className="size-16">
                <AvatarFallback className="bg-primary text-primary-foreground text-lg">
                  {initials}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-muted-foreground">Full Name</p>
                    <p className="font-medium">{user?.name || '-'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Email</p>
                    <p className="font-medium">{user?.email || '-'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Role</p>
                    <Badge variant="secondary">{user?.role || '-'}</Badge>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Account Status</p>
                    <Badge variant={user?.isActive ? 'default' : 'destructive'}>
                      {user?.isActive ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">User ID</p>
                    <div className="flex items-center gap-1">
                      <p className="font-mono text-sm">{user?.id || '-'}</p>
                      {user?.id && <CopyButton value={user.id} />}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Merchant Info Card */}
        <Card>
          <CardHeader>
            <CardTitle>Business Information</CardTitle>
            <CardDescription>Your merchant account details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">Business Name</p>
                <p className="font-medium">{merchant?.businessName || '-'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Merchant ID</p>
                <div className="flex items-center gap-1">
                  <p className="font-mono text-sm">{merchant?.id || '-'}</p>
                  {merchant?.id && <CopyButton value={merchant.id} />}
                </div>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Contact Email</p>
                <p className="font-medium">{merchant?.contactEmail || '-'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Contact Phone</p>
                <p className="font-medium">{(merchant?.contactPhone as string) || '-'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">KYC Status</p>
                <StatusBadge status={merchant?.kycStatus || 'PENDING'} />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Account Status</p>
                <StatusBadge status={merchant?.status || 'ACTIVE'} />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
