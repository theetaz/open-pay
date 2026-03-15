import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Plus, RefreshCw, Users, Loader2 } from 'lucide-react'
import { usePlans, useSubscriptions, useCreatePlan, useArchivePlan, useCancelSubscription } from '#/hooks/use-subscriptions'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'

export const Route = createFileRoute('/_dashboard/subscriptions')({
  component: SubscriptionsPage,
})

function SubscriptionsPage() {
  const { data: plansData } = usePlans()
  const { data: subsData } = useSubscriptions()
  const archivePlan = useArchivePlan()
  const cancelSub = useCancelSubscription()

  const plans = plansData?.data || []
  const subscriptions = subsData?.data || []
  const activePlans = plans.filter((p) => p.status === 'ACTIVE')
  const activeSubs = subscriptions.filter((s) => s.status === 'ACTIVE' || s.status === 'TRIAL')

  return (
    <>
      <PageHeader
        title="Subscriptions"
        description="Manage recurring payment plans and subscribers"
        action={<CreatePlanDialog />}
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Active Plans" value={String(activePlans.length)} description="Subscription plans" icon={RefreshCw} />
        <StatCard title="Active Subscribers" value={String(activeSubs.length)} description="Currently subscribed" icon={Users} />
        <StatCard title="Total Revenue" value="0.00 USDT" description="From subscriptions" />
      </div>

      <Tabs defaultValue="plans">
        <TabsList>
          <TabsTrigger value="plans">Plans ({plans.length})</TabsTrigger>
          <TabsTrigger value="subscribers">Subscribers ({subscriptions.length})</TabsTrigger>
        </TabsList>

        <TabsContent value="plans" className="mt-4">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Interval</TableHead>
                    <TableHead>Trial</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {plans.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7}>
                        <EmptyState message="No subscription plans yet." description="Create your first plan to start accepting recurring payments." />
                      </TableCell>
                    </TableRow>
                  ) : (
                    plans.map((p) => (
                      <TableRow key={p.id}>
                        <TableCell>
                          <div>
                            <p className="font-medium">{p.name}</p>
                            {p.description && <p className="text-xs text-muted-foreground">{p.description}</p>}
                          </div>
                        </TableCell>
                        <TableCell className="font-medium">{p.amount} {p.currency}</TableCell>
                        <TableCell className="text-sm">Every {p.intervalCount > 1 ? `${p.intervalCount} ` : ''}{p.intervalType.toLowerCase()}{p.intervalCount > 1 ? 's' : ''}</TableCell>
                        <TableCell className="text-sm">{p.trialDays > 0 ? `${p.trialDays} days` : '-'}</TableCell>
                        <TableCell><StatusBadge status={p.status} /></TableCell>
                        <TableCell className="text-sm text-muted-foreground">{new Date(p.createdAt).toLocaleDateString()}</TableCell>
                        <TableCell>
                          {p.status === 'ACTIVE' && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => archivePlan.mutate(p.id)}
                              disabled={archivePlan.isPending}
                            >
                              Archive
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="subscribers" className="mt-4">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Email</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Next Billing</TableHead>
                    <TableHead>Total Paid</TableHead>
                    <TableHead>Billing Count</TableHead>
                    <TableHead>Subscribed</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {subscriptions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7}>
                        <EmptyState message="No subscribers yet." description="Subscribers will appear here when customers subscribe to your plans." />
                      </TableCell>
                    </TableRow>
                  ) : (
                    subscriptions.map((s) => (
                      <TableRow key={s.id}>
                        <TableCell className="font-medium">{s.subscriberEmail}</TableCell>
                        <TableCell><StatusBadge status={s.status} /></TableCell>
                        <TableCell className="text-sm">{new Date(s.nextBillingDate).toLocaleDateString()}</TableCell>
                        <TableCell className="text-sm">{s.totalPaidUsdt} USDT</TableCell>
                        <TableCell className="text-sm">{s.billingCount}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{new Date(s.createdAt).toLocaleDateString()}</TableCell>
                        <TableCell>
                          {(s.status === 'ACTIVE' || s.status === 'TRIAL') && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-red-600"
                              onClick={() => cancelSub.mutate({ id: s.id, reason: 'Cancelled by merchant' })}
                              disabled={cancelSub.isPending}
                            >
                              Cancel
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </>
  )
}

function CreatePlanDialog() {
  const [open, setOpen] = React.useState(false)
  const [name, setName] = React.useState('')
  const [description, setDescription] = React.useState('')
  const [amount, setAmount] = React.useState('')
  const [currency, setCurrency] = React.useState('USDT')
  const [intervalType, setIntervalType] = React.useState('MONTHLY')
  const [intervalCount, setIntervalCount] = React.useState(1)
  const [trialDays, setTrialDays] = React.useState(0)

  const createPlan = useCreatePlan()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    createPlan.mutate({ name, description, amount, currency, intervalType, intervalCount, trialDays }, {
      onSuccess: () => {
        setOpen(false)
        setName('')
        setDescription('')
        setAmount('')
      },
    })
  }

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="mr-2 size-4" /> Create Plan
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Create Subscription Plan</DialogTitle>
              <DialogDescription>Set up a new recurring payment plan</DialogDescription>
            </DialogHeader>

            <div className="py-4">
              <FieldGroup>
                {createPlan.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {createPlan.error.message}
                  </div>
                )}

                <Field>
                  <FieldLabel>Plan Name</FieldLabel>
                  <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Premium Monthly" required />
                </Field>

                <Field>
                  <FieldLabel>Description</FieldLabel>
                  <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Access to premium features" />
                </Field>

                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Amount</FieldLabel>
                    <Input type="number" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="10.00" required />
                  </Field>
                  <Field>
                    <FieldLabel>Currency</FieldLabel>
                    <Select value={currency} onValueChange={(v) => v && setCurrency(v)}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="USDT">USDT</SelectItem>
                        <SelectItem value="LKR">LKR</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Billing Interval</FieldLabel>
                    <Select value={intervalType} onValueChange={(v) => v && setIntervalType(v)}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="DAILY">Daily</SelectItem>
                        <SelectItem value="WEEKLY">Weekly</SelectItem>
                        <SelectItem value="MONTHLY">Monthly</SelectItem>
                        <SelectItem value="YEARLY">Yearly</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                  <Field>
                    <FieldLabel>Every N intervals</FieldLabel>
                    <Input type="number" min={1} value={intervalCount} onChange={(e) => setIntervalCount(parseInt(e.target.value) || 1)} />
                  </Field>
                </div>

                <Field>
                  <FieldLabel>Trial Days</FieldLabel>
                  <Input type="number" min={0} value={trialDays} onChange={(e) => setTrialDays(parseInt(e.target.value) || 0)} placeholder="0" />
                </Field>
              </FieldGroup>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={createPlan.isPending}>
                {createPlan.isPending ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Creating...</> : 'Create Plan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}
