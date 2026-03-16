import { useMe } from '#/hooks/use-auth'
import { useAuthStore } from '#/stores/auth'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Badge } from '#/components/ui/badge'
import { Separator } from '#/components/ui/separator'
import { PageHeader } from '#/components/dashboard/page-header'
import { Skeleton } from '#/components/ui/skeleton'
import { Shield, LogOut, Monitor, Clock } from 'lucide-react'

export function SecurityPage() {
  const { data: meData, isLoading } = useMe()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const user = meData?.data?.user

  const handleLogout = () => {
    useAuthStore.getState().logout()
    queryClient.clear()
    navigate('/login')
  }

  if (isLoading) {
    return (
      <>
        <PageHeader title="Security" description="Manage your account security and active sessions" />
        <div className="space-y-6">
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </>
    )
  }

  return (
    <>
      <PageHeader title="Security" description="Manage your account security and active sessions" />

      <div className="space-y-6">
        {/* Account Security Overview */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="size-5" />
              Account Security
            </CardTitle>
            <CardDescription>Overview of your account security settings</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Email Address</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
              <Badge variant="secondary">Verified</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Password</p>
                <p className="text-xs text-muted-foreground">Last changed: Unknown</p>
              </div>
              <Button variant="outline" size="sm" disabled>
                Change Password
              </Button>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Two-Factor Authentication</p>
                <p className="text-xs text-muted-foreground">Add an extra layer of security</p>
              </div>
              <Button variant="outline" size="sm" disabled>
                Enable 2FA
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Current Session */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Monitor className="size-5" />
              Current Session
            </CardTitle>
            <CardDescription>Your active login session</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex size-10 items-center justify-center rounded-full bg-green-500/10">
                    <Monitor className="size-5 text-green-500" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Current Browser Session</p>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Clock className="size-3" />
                      <span>Active now</span>
                    </div>
                  </div>
                </div>
                <Badge variant="default" className="bg-green-500 hover:bg-green-600">Active</Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Danger Zone */}
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
            <CardDescription>Irreversible actions for your account</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Log out of current session</p>
                <p className="text-xs text-muted-foreground">
                  This will end your current session and redirect you to the login page
                </p>
              </div>
              <Button variant="destructive" size="sm" onClick={handleLogout}>
                <LogOut className="mr-2 size-4" />
                Log Out
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
