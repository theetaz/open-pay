import * as React from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { Mail, Lock, Eye, EyeOff, Loader2, Shield } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { api } from '#/lib/api'
import { useAuthStore } from '#/stores/auth'

export function LoginPage() {
  const [email, setEmail] = React.useState('')
  const [password, setPassword] = React.useState('')
  const [showPassword, setShowPassword] = React.useState(false)
  const navigate = useNavigate()

  const login = useMutation({
    mutationFn: (data: { email: string; password: string }) =>
      api.post<{ data: { accessToken: string; refreshToken: string; user: any } }>('/v1/admin/auth/login', data),
    onSuccess: (res) => {
      useAuthStore.getState().login(res.data.accessToken, res.data.refreshToken)
      navigate('/')
    },
  })

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    login.mutate({ email, password })
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
              <Shield className="size-8 text-primary" />
            </div>
            <CardTitle className="text-xl">Admin Login</CardTitle>
            <CardDescription>Sign in to the platform admin dashboard</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit}>
              <FieldGroup>
                {login.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {login.error.message}
                  </div>
                )}

                <Field>
                  <FieldLabel htmlFor="email">Email</FieldLabel>
                  <div className="relative">
                    <Mail className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                    <Input
                      id="email"
                      type="email"
                      placeholder="admin@openlankapay.lk"
                      className="pl-9"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      required
                    />
                  </div>
                </Field>

                <Field>
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                    <Input
                      id="password"
                      type={showPassword ? 'text' : 'password'}
                      placeholder="Enter your password"
                      className="pl-9 pr-9"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      required
                    />
                    <button
                      type="button"
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? <EyeOff data-icon="inline-end" /> : <Eye data-icon="inline-end" />}
                    </button>
                  </div>
                </Field>

                <Button type="submit" size="lg" className="w-full" disabled={login.isPending}>
                  {login.isPending ? (
                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Signing in...</>
                  ) : (
                    'Sign In'
                  )}
                </Button>
              </FieldGroup>
            </form>
          </CardContent>
        </Card>

        <p className="text-center text-xs text-muted-foreground mt-4">
          Platform Administration — Authorized Personnel Only
        </p>
      </div>
    </div>
  )
}
