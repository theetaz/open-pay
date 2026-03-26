import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Badge } from '#/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Plus, Loader2, UserX, UserCheck } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface AdminUser {
  id: string; email: string; name: string; isActive: boolean; lastLoginAt?: string; createdAt: string
  role?: { id: string; name: string; permissions: string[] }
}
interface AdminRole { id: string; name: string; description: string; permissions: string[]; isSystem: boolean }

const TEAM_FILTERS: FilterConfig[] = []

export function SettingsTeamPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = React.useState(false)
  const [form, setForm] = React.useState({ email: '', password: '', name: '', roleId: '' })
  const [search, setSearch] = React.useState('')
  const [filterValues, setFilterValues] = React.useState<Record<string, string | string[]>>({})

  const { data: usersData, isLoading } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => api.get<{ data: AdminUser[]; meta: { total: number } }>('/v1/admin/users?perPage=50'),
  })

  const { data: rolesData } = useQuery({
    queryKey: ['admin', 'roles'],
    queryFn: () => api.get<{ data: AdminRole[] }>('/v1/admin/roles'),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/v1/admin/users', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      setCreateOpen(false)
      setForm({ email: '', password: '', name: '', roleId: '' })
      toast.success('Admin user created')
    },
    onError: (err: any) => toast.error(err.message),
  })

  const toggleMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/admin/users/${id}/deactivate`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      toast.success('User status updated')
    },
  })

  const users = usersData?.data || []
  const roles = rolesData?.data || []

  const filteredUsers = React.useMemo(() => {
    if (!search) return users
    const q = search.toLowerCase()
    return users.filter(
      (u) =>
        u.name.toLowerCase().includes(q) ||
        u.email.toLowerCase().includes(q)
    )
  }, [users, search])

  const columns: ColumnDef<AdminUser>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
    },
    {
      accessorKey: 'email',
      header: 'Email',
      cell: ({ row }) => <span className="text-sm">{row.original.email}</span>,
    },
    {
      accessorKey: 'role',
      header: 'Role',
      cell: ({ row }) => <Badge variant="secondary">{row.original.role?.name || 'N/A'}</Badge>,
    },
    {
      id: 'status',
      header: 'Status',
      cell: ({ row }) =>
        row.original.isActive ? (
          <Badge className="bg-green-500/10 text-green-600">Active</Badge>
        ) : (
          <Badge variant="destructive">Inactive</Badge>
        ),
    },
    {
      accessorKey: 'lastLoginAt',
      header: 'Last Login',
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.lastLoginAt ? new Date(row.original.lastLoginAt).toLocaleString() : 'Never'}
        </span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      enableHiding: false,
      cell: ({ row }) => {
        const u = row.original
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => toggleMutation.mutate(u.id)}
            disabled={toggleMutation.isPending}
          >
            {u.isActive ? (
              <>
                <UserX className="size-4 mr-1" />Deactivate
              </>
            ) : (
              <>
                <UserCheck className="size-4 mr-1" />Activate
              </>
            )}
          </Button>
        )
      },
    },
  ]

  return (
    <>
      <PageHeader
        title="Team Members"
        description="Manage platform administrators and their access"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 size-4" />Invite Admin
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={filteredUsers}
        filters={TEAM_FILTERS}
        filterValues={filterValues}
        onFilterChange={(id, value) => setFilterValues((prev) => ({ ...prev, [id]: value }))}
        onClearFilters={() => { setFilterValues({}); setSearch('') }}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search by name or email..."
        pagination={{ page: 1, perPage: 999, total: filteredUsers.length }}
        onPageChange={() => {}}
        isLoading={isLoading}
      />

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invite Admin User</DialogTitle>
            <DialogDescription>Create a new administrator account</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Full Name</FieldLabel>
              <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="John Doe" />
            </Field>
            <Field>
              <FieldLabel>Email</FieldLabel>
              <Input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="john@example.com" />
            </Field>
            <Field>
              <FieldLabel>Password</FieldLabel>
              <Input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="Min 8 characters" />
            </Field>
            <Field>
              <FieldLabel>Role</FieldLabel>
              <Select value={form.roleId} onValueChange={(v) => v && setForm({ ...form, roleId: v })}>
                <SelectTrigger>
                  <SelectValue placeholder="Select role">
                    {form.roleId ? (() => {
                      const selected = roles.find((r) => r.id === form.roleId)
                      return selected ? `${selected.name} — ${selected.description}` : 'Select role'
                    })() : undefined}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {roles.map((r) => (
                    <SelectItem key={r.id} value={r.id}>{r.name} — {r.description}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={() => createMutation.mutate(form)} disabled={createMutation.isPending || !form.email || !form.password || !form.name || !form.roleId}>
              {createMutation.isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}Create Admin
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
