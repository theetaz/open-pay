import { createFileRoute } from '@tanstack/react-router'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'

export const Route = createFileRoute('/settlements')({ component: SettlementsPage })

function SettlementsPage() {
  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Settlements</h2>
        <Button>Request Withdrawal</Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Available Balance</CardDescription>
            <CardTitle className="text-2xl">0.00 USDT</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Pending Withdrawal</CardDescription>
            <CardTitle className="text-2xl">0.00 USDT</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Withdrawn</CardDescription>
            <CardTitle className="text-2xl">0.00 LKR</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Withdrawal History</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No withdrawals yet.</p>
        </CardContent>
      </Card>
    </div>
  )
}
