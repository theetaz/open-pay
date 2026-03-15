import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Switch } from '#/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetFooter } from '#/components/ui/sheet'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Separator } from '#/components/ui/separator'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, Link2 } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/payment-links')({
  component: PaymentLinksPage,
})

function PaymentLinksPage() {
  const [sheetOpen, setSheetOpen] = useState(false)

  return (
    <>
      <PageHeader
        title="Payment Links"
        description="Create and manage shareable payment links for your customers"
        action={
          <Button onClick={() => setSheetOpen(true)}>
            <Plus className="mr-2 size-4" /> Create Link
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Total Links" value="0" description="0 active" icon={Link2} />
        <StatCard title="Active Links" value="0" description="Currently available" icon={Link2} />
        <StatCard title="Total Used" value="0" description="Times links were used" icon={Link2} />
      </div>

      <div className="mb-4">
        <h3 className="text-lg font-semibold">All Payment Links</h3>
        <p className="text-sm text-muted-foreground">Manage and track all your payment links in one place</p>
      </div>

      <Input placeholder="Filter by name..." className="max-w-sm mb-4" />

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Used</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Branch</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={6}>
                  <EmptyState message="No results." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent className="sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Create Payment Link</SheetTitle>
            <SheetDescription>Create a new payment link for your customers</SheetDescription>
          </SheetHeader>

          <div className="space-y-6 py-4">
            <Separator />
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Basic Information</p>

            <div className="space-y-2">
              <Label htmlFor="link-name">Name *</Label>
              <Input id="link-name" placeholder="Item Name" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="link-slug">Slug *</Label>
              <Input id="link-slug" placeholder="item-name" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="link-desc">Description</Label>
              <Textarea id="link-desc" placeholder="Item Description (supports Markdown)" rows={4} />
            </div>

            <div className="space-y-2">
              <Label>Currency *</Label>
              <Select defaultValue="LKR">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="LKR">LKR (₹.)</SelectItem>
                  <SelectItem value="USDT">USDT (₮)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <Separator />
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Amount Settings</p>

            <div className="flex items-center justify-between">
              <div>
                <Label>Allow custom amount</Label>
                <p className="text-xs text-muted-foreground">Let customers enter their own amount</p>
              </div>
              <Switch />
            </div>

            <div className="space-y-2">
              <Label htmlFor="link-amount">Amount *</Label>
              <Input id="link-amount" type="number" placeholder="0.00" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="link-expiry">Expiration Date</Label>
              <Input id="link-expiry" type="datetime-local" />
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Reusable link</Label>
                <p className="text-xs text-muted-foreground">One-time use link (single payment)</p>
              </div>
              <Switch />
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Show on QR page</Label>
                <p className="text-xs text-muted-foreground">Display this payment link on your static QR page</p>
              </div>
              <Switch />
            </div>
          </div>

          <SheetFooter className="gap-2">
            <Button variant="outline" onClick={() => setSheetOpen(false)}>Cancel</Button>
            <Button>Create Payment Link</Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </>
  )
}
