import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Switch } from '#/components/ui/switch'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '#/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, Search } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/branches')({
  component: BranchesPage,
})

function BranchesPage() {
  const [dialogOpen, setDialogOpen] = useState(false)

  return (
    <>
      <PageHeader
        title="Branches"
        description="Manage your branch locations and details"
        action={
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 size-4" /> Add Branch
          </Button>
        }
      />

      <div className="flex items-center justify-between mb-4">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input placeholder="Search by name or code..." className="pl-9" />
        </div>
        <div className="text-right">
          <p className="text-2xl font-bold">0</p>
          <p className="text-xs text-muted-foreground">Total Branches</p>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Code</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>City</TableHead>
                <TableHead>Postal Code</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState message="No branches found." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Add New Branch</DialogTitle>
            <DialogDescription>Fill in the details to create a new branch.</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="branch-name">Branch Name *</Label>
                <Input id="branch-name" placeholder="Downtown Branch" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="branch-code">Branch Code *</Label>
                <Input id="branch-code" placeholder="DT-001" />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="branch-address">Address</Label>
              <Textarea id="branch-address" placeholder="123 Main Street" rows={2} />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="branch-city">City</Label>
                <Input id="branch-city" placeholder="Colombo" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="branch-postal">Postal Code</Label>
                <Input id="branch-postal" placeholder="00100" />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="branch-phone">Phone Number</Label>
              <Input id="branch-phone" placeholder="+94 77 123 4567" />
            </div>

            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <Label>Use merchant bank details</Label>
                <p className="text-xs text-muted-foreground">Disable to configure a branch-specific bank account</p>
              </div>
              <Switch defaultChecked />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button>Create Branch</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
