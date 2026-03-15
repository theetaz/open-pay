import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useMe } from '#/hooks/use-auth'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '#/components/ui/dialog'
import { Alert, AlertDescription } from '#/components/ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { StatCard } from '#/components/dashboard/stat-card'
import { CopyButton } from '#/components/dashboard/copy-button'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, AlertTriangle } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  return (
    <>
      <PageHeader title="Settings" description="Manage your merchant configuration and integration settings" />

      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="preferences">Preferences</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
          <TabsTrigger value="integration">Integration</TabsTrigger>
          <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-6 space-y-6">
          <GeneralTab />
        </TabsContent>

        <TabsContent value="preferences" className="mt-6">
          <PreferencesTab />
        </TabsContent>

        <TabsContent value="notifications" className="mt-6">
          <NotificationsTab />
        </TabsContent>

        <TabsContent value="integration" className="mt-6">
          <IntegrationTab />
        </TabsContent>

        <TabsContent value="webhooks" className="mt-6">
          <WebhooksTab />
        </TabsContent>
      </Tabs>
    </>
  )
}

function GeneralTab() {
  const { data: meData } = useMe()
  const merchant = meData?.data?.merchant
  const merchantId = merchant?.id || '-'

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Merchant Information</CardTitle>
          <CardDescription>Your business details and account status</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-xs text-muted-foreground">Business Name</p>
              <p className="font-medium">{merchant?.businessName || '-'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Merchant ID</p>
              <div className="flex items-center gap-1">
                <p className="font-mono text-sm">{merchantId}</p>
                <CopyButton value={merchantId} />
              </div>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Email</p>
              <p className="font-medium">{merchant?.contactEmail || '-'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Phone</p>
              <p className="font-medium">{(merchant?.contactPhone as string) || '-'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Status</p>
              <div className="flex gap-2 mt-1">
                <StatusBadge status={merchant?.kycStatus || 'PENDING'} />
                <StatusBadge status={merchant?.status || 'ACTIVE'} />
              </div>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Created At</p>
              <p className="font-medium">{merchant?.createdAt ? new Date(merchant.createdAt as string).toLocaleDateString() : '-'}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Fees</CardTitle>
          <CardDescription>Transaction fee structure for your merchant account</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Fee Type</TableHead>
                <TableHead className="text-right">Percentage</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>Platform Fee</TableCell>
                <TableCell className="text-right font-mono">1.5%</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Bybit Exchange Fee</TableCell>
                <TableCell className="text-right font-mono">0%</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Binance Exchange Fee</TableCell>
                <TableCell className="text-right font-mono">1%</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>KuCoin Exchange Fee</TableCell>
                <TableCell className="text-right font-mono">1%</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}

function PreferencesTab() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Static QR Page</CardTitle>
        <CardDescription>Configure preferences for your static QR payment page</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium text-sm">Default Currency</p>
            <p className="text-xs text-muted-foreground">The currency pre-selected when customers visit your QR page</p>
          </div>
          <Tabs defaultValue="LKR">
            <TabsList className="h-8">
              <TabsTrigger value="LKR" className="text-xs px-3">LKR</TabsTrigger>
              <TabsTrigger value="USDT" className="text-xs px-3">USDT</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium text-sm">Payment Links</p>
            <p className="text-xs text-muted-foreground">Show or hide quick payment links on your QR page</p>
          </div>
          <Tabs defaultValue="show">
            <TabsList className="h-8">
              <TabsTrigger value="show" className="text-xs px-3">Show</TabsTrigger>
              <TabsTrigger value="hide" className="text-xs px-3">Hide</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </CardContent>
    </Card>
  )
}

function NotificationsTab() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Telegram Notifications</CardTitle>
        <CardDescription>Configure telegram group for merchant-level notifications.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <p className="font-medium text-sm">Connect via Verification Code</p>
          <p className="text-xs text-muted-foreground mt-1">
            Use our automated verification system to easily connect your Telegram group
          </p>
        </div>
        <Button variant="secondary">Connect Telegram Group</Button>
      </CardContent>
    </Card>
  )
}

function IntegrationTab() {
  const [dialogOpen, setDialogOpen] = useState(false)

  return (
    <>
      <Tabs defaultValue="api-keys">
        <TabsList className="h-8">
          <TabsTrigger value="api-keys" className="text-xs">API Keys</TabsTrigger>
          <TabsTrigger value="extensions" className="text-xs">Extensions</TabsTrigger>
        </TabsList>

        <TabsContent value="api-keys" className="mt-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>API Integration</CardTitle>
                  <CardDescription>Generate and manage your API keys for payment integration</CardDescription>
                </div>
                <Button onClick={() => setDialogOpen(true)}>
                  <Plus className="mr-2 size-4" /> Generate New Key
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <EmptyState
                message="No API Keys Generated"
                description="Generate your first API key to start integrating Open Pay into your application."
                action={
                  <Button variant="outline" onClick={() => setDialogOpen(true)}>
                    <Plus className="mr-2 size-4" /> Generate Your First Key
                  </Button>
                }
              />
              <p className="text-xs text-muted-foreground text-center mt-4">
                Your API keys are used to authenticate requests to the Open Pay API. Keep them secure and never share them publicly.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="extensions" className="mt-4">
          <Card>
            <CardContent className="pt-6">
              <EmptyState message="No extensions available." />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Generate New API Key</DialogTitle>
            <DialogDescription>
              Create a new API key for your application. Make sure to copy it - it will only be shown once.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="key-name">Key Name</Label>
              <Input id="key-name" placeholder="e.g., Production API Key" />
            </div>

            <div className="space-y-2">
              <Label>Expiration (Optional)</Label>
              <div className="grid grid-cols-2 gap-2">
                <Select>
                  <SelectTrigger>
                    <SelectValue placeholder="Select date" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="30d">30 days</SelectItem>
                    <SelectItem value="90d">90 days</SelectItem>
                    <SelectItem value="1y">1 year</SelectItem>
                    <SelectItem value="never">Never</SelectItem>
                  </SelectContent>
                </Select>
                <Input type="time" defaultValue="23:59:59" />
              </div>
            </div>

            <Alert className="border-amber-500/50 bg-amber-500/10">
              <AlertTriangle className="size-4 text-amber-500" />
              <AlertDescription className="text-amber-600 dark:text-amber-400">
                <strong>Important:</strong> Your API key will only be shown once. Make sure to copy and store it securely.
              </AlertDescription>
            </Alert>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button>Generate Key</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function WebhooksTab() {
  return (
    <>
      <div className="grid gap-4 md:grid-cols-5 mb-6">
        <StatCard title="Total" value="0" />
        <StatCard title="Success" value="0" valueClassName="text-green-500" />
        <StatCard title="Pending" value="0" valueClassName="text-amber-500" />
        <StatCard title="Failed" value="0" valueClassName="text-red-500" />
        <StatCard title="Exhausted" value="0" valueClassName="text-red-500" />
      </div>

      <div className="flex items-center gap-4 mb-4">
        <p className="text-sm">Filter by status:</p>
        <Select defaultValue="all">
          <SelectTrigger className="w-[140px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            <SelectItem value="success">Success</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="exhausted">Exhausted</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Pay ID</TableHead>
            <TableHead>Attempts</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Timestamp</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell colSpan={4}>
              <EmptyState message="No webhooks found." />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </>
  )
}
