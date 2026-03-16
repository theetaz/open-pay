import { useState } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '#/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { UserPlus } from 'lucide-react'

export function UsersPage() {
  const [dialogOpen, setDialogOpen] = useState(false)

  return (
    <>
      <PageHeader
        title="Users"
        description="Manage users and their branch access"
        action={
          <Button onClick={() => setDialogOpen(true)}>
            <UserPlus className="mr-2 size-4" /> Invite User
          </Button>
        }
      />

      <div className="space-y-8">
        <section>
          <h3 className="text-lg font-semibold mb-3">Active Users</h3>
          <Input placeholder="Filter by name or email..." className="max-w-sm mb-4" />

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Branch</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Joined</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>
                  <div className="flex items-center gap-3">
                    <Avatar className="size-8">
                      <AvatarFallback className="text-xs">AD</AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium text-sm">Admin User</p>
                      <p className="text-xs text-muted-foreground">admin@openpay.lk</p>
                    </div>
                  </div>
                </TableCell>
                <TableCell className="text-sm">Main Branch</TableCell>
                <TableCell><StatusBadge status="Admin" /></TableCell>
                <TableCell><StatusBadge status="ACTIVE" /></TableCell>
                <TableCell className="text-sm text-muted-foreground">-</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </section>

        <section>
          <h3 className="text-lg font-semibold mb-3">Pending Invitations</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Branch</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Invited By</TableHead>
                <TableHead>Invited On</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={5}>
                  <EmptyState message="No pending invitations." />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </section>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invite User</DialogTitle>
            <DialogDescription>Send an invitation to a new user to join your organization.</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="user-email">Email</Label>
              <Input id="user-email" type="email" placeholder="user@example.com" />
            </div>

            <div className="space-y-2">
              <Label>Branch</Label>
              <Select defaultValue="main">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="main">Main Branch</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Role</Label>
              <Select defaultValue="user">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="manager">Manager</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button>Invite</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
