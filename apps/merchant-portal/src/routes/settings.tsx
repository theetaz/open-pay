import { createFileRoute } from '@tanstack/react-router'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'

export const Route = createFileRoute('/settings')({ component: SettingsPage })

function SettingsPage() {
  return (
    <div className="p-6 max-w-3xl">
      <h2 className="text-2xl font-bold mb-6">Settings</h2>

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>API Keys</CardTitle>
            <CardDescription>Manage API keys for integrating with the payment API.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button>Generate New Key</Button>
            <p className="mt-4 text-sm text-muted-foreground">No API keys created yet.</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Webhook Configuration</CardTitle>
            <CardDescription>Configure your endpoint to receive payment event notifications (ED25519 signed).</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="webhook-url">Webhook URL</Label>
              <Input id="webhook-url" type="url" placeholder="https://your-site.com/api/webhook" />
            </div>
            <Button variant="secondary">Save Webhook</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Bank Details</CardTitle>
            <CardDescription>Bank account for LKR settlements.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="bank-name">Bank Name</Label>
                <Input id="bank-name" placeholder="Bank of Ceylon" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="account-no">Account Number</Label>
                <Input id="account-no" placeholder="1234567890" />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <Label htmlFor="account-name">Account Holder Name</Label>
                <Input id="account-name" placeholder="Your business name" />
              </div>
            </div>
            <Button variant="secondary">Save Bank Details</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
