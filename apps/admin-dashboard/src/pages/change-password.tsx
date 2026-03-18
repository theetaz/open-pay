import * as React from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { Lock, Loader2, ShieldAlert } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { api } from '#/lib/api'
import { useAuthStore } from '#/stores/auth'
import { toast } from 'sonner'

export function ChangePasswordPage() {
  const [currentPassword, setCurrentPassword] = React.useState('')
  const [newPassword, setNewPassword] = React.useState('')
  const [confirmPassword, setConfirmPassword] = React.useState('')
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  const user = useAuthStore((s) => s.user)

  const mutation = useMutation({
    mutationFn: (data: { currentPassword: string; newPassword: string }) =>
      api.post('/v1/admin/auth/change-password', data),
    onSuccess: () => {
      // Clear the mustChangePassword flag in the store
      if (user) {
        setUser({ ...user, mustChangePassword: false } as any)
      }
      toast.success('Password changed successfully')
      navigate('/')
    },
  })

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match')
      return
    }
    if (newPassword.length < 8) {
      toast.error('Password must be at least 8 characters')
      return
    }
    mutation.mutate({ currentPassword, newPassword })
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
              OP
            </div>
            <div>
              <h1 className="text-xl font-bold tracking-tight">Open Pay</h1>
              <p className="text-[10px] text-muted-foreground tracking-widest uppercase">Admin Console</p>
            </div>
          </div>
        </div>

        <Card>
          <CardHeader className="text-center">
            <div className="flex justify-center mb-2">
              <ShieldAlert className="size-8 text-amber-500" />
            </div>
            <CardTitle className="text-xl">Change Your Password</CardTitle>
            <CardDescription>You must change your temporary password before continuing</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit}>
              <FieldGroup>
                {mutation.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {(mutation.error as any)?.message || 'Failed to change password'}
                  </div>
                )}

                <Field>
                  <FieldLabel htmlFor="current">Current Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground size-4" />
                    <Input
                      id="current"
                      type="password"
                      className="pl-9"
                      placeholder="Enter current password"
                      value={currentPassword}
                      onChange={(e) => setCurrentPassword(e.target.value)}
                      required
                    />
                  </div>
                </Field>

                <Field>
                  <FieldLabel htmlFor="new">New Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground size-4" />
                    <Input
                      id="new"
                      type="password"
                      className="pl-9"
                      placeholder="Min 8 characters"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      required
                    />
                  </div>
                </Field>

                <Field>
                  <FieldLabel htmlFor="confirm">Confirm New Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground size-4" />
                    <Input
                      id="confirm"
                      type="password"
                      className="pl-9"
                      placeholder="Re-enter new password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      required
                    />
                  </div>
                </Field>

                <Button type="submit" size="lg" className="w-full" disabled={mutation.isPending}>
                  {mutation.isPending ? (
                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Changing password...</>
                  ) : (
                    'Change Password & Continue'
                  )}
                </Button>
              </FieldGroup>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
