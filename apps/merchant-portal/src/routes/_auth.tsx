import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth')({
  component: AuthLayout,
})

function AuthLayout() {
  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <div className="mb-8 flex items-center gap-2">
        <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
          OP
        </div>
        <span className="text-lg font-bold">Open Pay</span>
      </div>

      <Outlet />

      <p className="mt-8 text-sm text-muted-foreground">
        Secure crypto payment processing
      </p>
    </div>
  )
}
