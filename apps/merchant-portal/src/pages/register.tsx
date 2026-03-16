import * as React from 'react'
import { Link } from 'react-router-dom'
import { User, Building2, Mail, Lock, Loader2 } from 'lucide-react'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Checkbox } from '#/components/ui/checkbox'
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldDescription,
  FieldSeparator,
} from '#/components/ui/field'
import { useRegister } from '#/hooks/use-auth'

export function RegisterPage() {
  const [fullName, setFullName] = React.useState('')
  const [businessName, setBusinessName] = React.useState('')
  const [email, setEmail] = React.useState('')
  const [password, setPassword] = React.useState('')
  const [confirmPassword, setConfirmPassword] = React.useState('')

  const register = useRegister()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (password !== confirmPassword) return
    register.mutate({ businessName, email, password, name: fullName })
  }

  return (
    <Card className="max-w-md w-full">
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Create your account</CardTitle>
        <CardDescription>
          Start accepting crypto payments today
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <FieldGroup>
            {register.isError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {register.error.message}
              </div>
            )}

            <Field>
              <FieldLabel htmlFor="fullName">Full Name</FieldLabel>
              <div className="relative">
                <User className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                <Input
                  id="fullName"
                  type="text"
                  placeholder="John Doe"
                  className="pl-9"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  required
                />
              </div>
            </Field>

            <Field>
              <FieldLabel htmlFor="businessName">Business Name</FieldLabel>
              <div className="relative">
                <Building2 className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                <Input
                  id="businessName"
                  type="text"
                  placeholder="Acme Inc."
                  className="pl-9"
                  value={businessName}
                  onChange={(e) => setBusinessName(e.target.value)}
                  required
                />
              </div>
            </Field>

            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <div className="relative">
                <Mail className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
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
                  type="password"
                  placeholder="Create a password"
                  className="pl-9"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              <FieldDescription>
                Must be at least 8 characters with 1 uppercase and 1 number
              </FieldDescription>
            </Field>

            <Field>
              <FieldLabel htmlFor="confirmPassword">Confirm Password</FieldLabel>
              <div className="relative">
                <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" data-icon="inline-start" />
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="Confirm your password"
                  className="pl-9"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                />
              </div>
              {confirmPassword && password !== confirmPassword && (
                <p className="text-sm text-destructive">Passwords do not match</p>
              )}
            </Field>

            <div className="flex items-center gap-2">
              <Checkbox id="terms" required />
              <FieldLabel htmlFor="terms" className="font-normal">
                I agree to the{' '}
                <Link to="#" className="text-primary hover:underline">
                  Terms of Service
                </Link>{' '}
                and{' '}
                <Link to="#" className="text-primary hover:underline">
                  Privacy Policy
                </Link>
              </FieldLabel>
            </div>

            <Button type="submit" size="lg" className="w-full" disabled={register.isPending || password !== confirmPassword}>
              {register.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating account...
                </>
              ) : (
                'Create Account'
              )}
            </Button>

            <FieldSeparator />

            <FieldDescription className="text-center">
              Already have an account?{' '}
              <Link to="/login" className="text-primary hover:underline text-sm">
                Sign in
              </Link>
            </FieldDescription>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
