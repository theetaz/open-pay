import { useState, useMemo } from 'react'
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
import { DataTable, type FilterConfig } from '#/components/data-table'
import { UserPlus } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'

// Placeholder types until real hooks are available
interface ActiveUser {
  id: string
  name: string
  email: string
  initials: string
  branch: string
  role: string
  status: string
  joinedAt: string
}

interface PendingInvitation {
  id: string
  email: string
  branch: string
  role: string
  invitedBy: string
  invitedOn: string
}

export function UsersPage() {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [filterValues, setFilterValues] = useState<Record<string, string | string[]>>({ role: '' })

  // Placeholder data - replace with actual hooks when available
  const activeUsers: ActiveUser[] = [
    {
      id: '1',
      name: 'Admin User',
      email: 'admin@openpay.lk',
      initials: 'AD',
      branch: 'Main Branch',
      role: 'Admin',
      status: 'ACTIVE',
      joinedAt: '-',
    },
  ]

  const pendingInvitations: PendingInvitation[] = []

  const filters: FilterConfig[] = [
    {
      id: 'role',
      label: 'Role',
      type: 'select',
      options: [
        { label: 'Admin', value: 'Admin' },
        { label: 'Manager', value: 'Manager' },
        { label: 'User', value: 'User' },
      ],
    },
  ]

  const handleFilterChange = (id: string, value: string | string[]) => {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
  }

  const handleClearFilters = () => {
    setFilterValues({ role: '' })
  }

  const filteredUsers = useMemo(() => {
    let result = activeUsers

    if (search) {
      const q = search.toLowerCase()
      result = result.filter(
        (u) =>
          u.name.toLowerCase().includes(q) ||
          u.email.toLowerCase().includes(q),
      )
    }

    const roleFilter = filterValues.role
    if (roleFilter && typeof roleFilter === 'string' && roleFilter !== '') {
      result = result.filter((u) => u.role === roleFilter)
    }

    return result
  }, [activeUsers, search, filterValues])

  const userColumns: ColumnDef<ActiveUser, any>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: 'User',
        cell: ({ row }) => {
          const u = row.original
          return (
            <div className="flex items-center gap-3">
              <Avatar className="size-8">
                <AvatarFallback className="text-xs">{u.initials}</AvatarFallback>
              </Avatar>
              <div>
                <p className="font-medium text-sm">{u.name}</p>
                <p className="text-xs text-muted-foreground">{u.email}</p>
              </div>
            </div>
          )
        },
      },
      {
        accessorKey: 'branch',
        header: 'Branch',
        cell: ({ row }) => <span className="text-sm">{row.original.branch}</span>,
      },
      {
        accessorKey: 'role',
        header: 'Role',
        cell: ({ row }) => <StatusBadge status={row.original.role} />,
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'joinedAt',
        header: 'Joined',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">{row.original.joinedAt}</span>
        ),
      },
    ],
    [],
  )

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

          <DataTable
            columns={userColumns}
            data={filteredUsers}
            filters={filters}
            filterValues={filterValues}
            onFilterChange={handleFilterChange}
            onClearFilters={handleClearFilters}
            search={search}
            onSearchChange={setSearch}
            searchPlaceholder="Filter by name or email..."
            pagination={{ page: 1, perPage: 999, total: filteredUsers.length }}
            onPageChange={() => {}}
          />
        </section>

        <section>
          <h3 className="text-lg font-semibold mb-3">Pending Invitations</h3>
          <div className="rounded-lg border overflow-hidden bg-card">
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
                {pendingInvitations.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5}>
                      <EmptyState message="No pending invitations." />
                    </TableCell>
                  </TableRow>
                ) : (
                  pendingInvitations.map((inv) => (
                    <TableRow key={inv.id}>
                      <TableCell className="font-medium">{inv.email}</TableCell>
                      <TableCell className="text-sm">{inv.branch}</TableCell>
                      <TableCell><StatusBadge status={inv.role} /></TableCell>
                      <TableCell className="text-sm">{inv.invitedBy}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{inv.invitedOn}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
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
