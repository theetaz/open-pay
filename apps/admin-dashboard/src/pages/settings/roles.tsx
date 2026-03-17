import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Checkbox } from '#/components/ui/checkbox'
import { Badge } from '#/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { PageHeader } from '#/components/dashboard/page-header'
import { Plus, Pencil, Loader2, Shield } from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface AdminRole {
  id: string; name: string; description: string; permissions: string[]; isSystem: boolean; createdAt: string
}

const PERMISSION_GROUPS: Record<string, { label: string; permissions: string[] }> = {
  merchants: { label: 'Merchants', permissions: ['merchants:read', 'merchants:approve', 'merchants:reject', 'merchants:manage'] },
  withdrawals: { label: 'Withdrawals', permissions: ['withdrawals:read', 'withdrawals:approve', 'withdrawals:reject', 'withdrawals:complete'] },
  payments: { label: 'Payments', permissions: ['payments:read'] },
  treasury: { label: 'Treasury', permissions: ['treasury:read', 'treasury:manage'] },
  audit: { label: 'Audit', permissions: ['audit:read'] },
  subscriptions: { label: 'Subscriptions', permissions: ['subscriptions:read'] },
  notifications: { label: 'Notifications', permissions: ['notifications:read'] },
  system: { label: 'System', permissions: ['system:manage'] },
  team: { label: 'Team', permissions: ['team:read', 'team:manage'] },
  settings: { label: 'Settings', permissions: ['settings:read', 'settings:manage'] },
}

export function SettingsRolesPage() {
  const queryClient = useQueryClient()
  const [editRole, setEditRole] = React.useState<AdminRole | null>(null)
  const [createOpen, setCreateOpen] = React.useState(false)
  const [form, setForm] = React.useState({ name: '', description: '', permissions: [] as string[] })

  const { data } = useQuery({
    queryKey: ['admin', 'roles'],
    queryFn: () => api.get<{ data: AdminRole[] }>('/v1/admin/roles'),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/v1/admin/roles', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'roles'] })
      setCreateOpen(false)
      setForm({ name: '', description: '', permissions: [] })
      toast.success('Role created')
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...body }: { id: string; description: string; permissions: string[] }) =>
      api.put(`/v1/admin/roles/${id}`, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'roles'] })
      setEditRole(null)
      toast.success('Role updated')
    },
  })

  const roles = data?.data || []

  const togglePermission = (perm: string) => {
    setForm((f) => ({
      ...f,
      permissions: f.permissions.includes(perm)
        ? f.permissions.filter((p) => p !== perm)
        : [...f.permissions, perm],
    }))
  }

  const openEdit = (role: AdminRole) => {
    setEditRole(role)
    setForm({ name: role.name, description: role.description, permissions: [...role.permissions] })
  }

  const openCreate = () => {
    setCreateOpen(true)
    setForm({ name: '', description: '', permissions: [] })
  }

  return (
    <>
      <PageHeader title="Roles & Permissions" description="Define access levels for admin team members"
        action={<Button onClick={openCreate}><Plus className="mr-2 size-4" />Create Role</Button>}
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {roles.map((role) => (
          <Card key={role.id}>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base flex items-center gap-2">
                  <Shield className="size-4" />
                  {role.name}
                </CardTitle>
                {role.isSystem && <Badge variant="secondary" className="text-[10px]">System</Badge>}
              </div>
              <CardDescription className="text-xs">{role.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-1 mb-3">
                {role.permissions.map((p) => (
                  <Badge key={p} variant="outline" className="text-[10px]">{p}</Badge>
                ))}
              </div>
              {!role.isSystem && (
                <Button variant="outline" size="sm" className="w-full" onClick={() => openEdit(role)}>
                  <Pencil className="size-3 mr-1" />Edit Permissions
                </Button>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Create/Edit Role Dialog */}
      <Dialog open={createOpen || !!editRole} onOpenChange={(v) => { if (!v) { setCreateOpen(false); setEditRole(null) } }}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editRole ? `Edit Role: ${editRole.name}` : 'Create New Role'}</DialogTitle>
            <DialogDescription>Configure permissions for this role</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            {!editRole && (
              <Field>
                <FieldLabel>Role Name</FieldLabel>
                <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="CUSTOM_ROLE" />
              </Field>
            )}
            <Field>
              <FieldLabel>Description</FieldLabel>
              <Input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="Describe the role's purpose" />
            </Field>

            <div className="space-y-4 pt-2">
              <p className="text-sm font-medium">Permissions</p>
              {Object.entries(PERMISSION_GROUPS).map(([group, { label, permissions }]) => (
                <div key={group} className="rounded-lg border p-3">
                  <p className="text-sm font-medium mb-2">{label}</p>
                  <div className="grid grid-cols-2 gap-2">
                    {permissions.map((perm) => (
                      <label key={perm} className="flex items-center gap-2 text-sm cursor-pointer">
                        <Checkbox
                          checked={form.permissions.includes(perm)}
                          onCheckedChange={() => togglePermission(perm)}
                        />
                        <span className="font-mono text-xs">{perm.split(':')[1]}</span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setCreateOpen(false); setEditRole(null) }}>Cancel</Button>
            <Button
              onClick={() => {
                if (editRole) {
                  updateMutation.mutate({ id: editRole.id, description: form.description, permissions: form.permissions })
                } else {
                  createMutation.mutate(form)
                }
              }}
              disabled={createMutation.isPending || updateMutation.isPending}
            >
              {(createMutation.isPending || updateMutation.isPending) ? <Loader2 className="mr-2 size-4 animate-spin" /> : null}
              {editRole ? 'Save Changes' : 'Create Role'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
