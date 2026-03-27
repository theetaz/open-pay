import * as React from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Plus, RefreshCw, Users, Loader2, DollarSign } from 'lucide-react'
import { usePlans, useSubscriptions, useCreatePlan, useArchivePlan, useCancelSubscription } from '#/hooks/use-subscriptions'
import { useMe } from '#/hooks/use-auth'
import { useExchangeRate } from '#/hooks/use-exchange-rate'
import { formatAmount, formatDualAmount } from '#/lib/currency'
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
import type { ColumnDef } from '@tanstack/react-table'

export function SubscriptionsPage() {
  const { data: plansData } = usePlans()
  const { data: subsData } = useSubscriptions()
  const { data: meData } = useMe()
  const { data: rateData } = useExchangeRate()
  const archivePlan = useArchivePlan()
  const cancelSub = useCancelSubscription()

  const plans = plansData?.data || []
  const subscriptions = subsData?.data || []
  const primaryCurrency = meData?.data?.merchant?.defaultCurrency || 'LKR'
  const liveRate = rateData?.data?.effectiveRate || null
  const activePlans = plans.filter((p) => p.status === 'ACTIVE')
  const activeSubs = subscriptions.filter((s) => s.status === 'ACTIVE' || s.status === 'TRIAL')
  const totalSubRevenue = subscriptions.reduce((sum, s) => sum + parseFloat(s.totalPaidUsdt || '0'), 0)
  const subRevenueFmt = formatDualAmount(totalSubRevenue, undefined, undefined, primaryCurrency, liveRate)

  // Plans tab state
  const [plansSearch, setPlansSearch] = React.useState('')
  const [plansFilterValues, setPlansFilterValues] = React.useState<Record<string, string | string[]>>({ status: '' })

  // Subscribers tab state
  const [subsSearch, setSubsSearch] = React.useState('')
  const [subsFilterValues, setSubsFilterValues] = React.useState<Record<string, string | string[]>>({ status: '' })

  const plansFilters: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { label: 'Active', value: 'ACTIVE' },
        { label: 'Archived', value: 'ARCHIVED' },
      ],
    },
  ]

  const subsFilters: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { label: 'Active', value: 'ACTIVE' },
        { label: 'Trial', value: 'TRIAL' },
        { label: 'Cancelled', value: 'CANCELLED' },
        { label: 'Expired', value: 'EXPIRED' },
      ],
    },
  ]

  const filteredPlans = React.useMemo(() => {
    let result = plans

    if (plansSearch) {
      const q = plansSearch.toLowerCase()
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          (p.description && p.description.toLowerCase().includes(q)),
      )
    }

    const statusFilter = plansFilterValues.status
    if (statusFilter && typeof statusFilter === 'string' && statusFilter !== '') {
      result = result.filter((p) => p.status === statusFilter)
    }

    return result
  }, [plans, plansSearch, plansFilterValues])

  const filteredSubs = React.useMemo(() => {
    let result = subscriptions

    if (subsSearch) {
      const q = subsSearch.toLowerCase()
      result = result.filter((s) => s.subscriberEmail.toLowerCase().includes(q))
    }

    const statusFilter = subsFilterValues.status
    if (statusFilter && typeof statusFilter === 'string' && statusFilter !== '') {
      result = result.filter((s) => s.status === statusFilter)
    }

    return result
  }, [subscriptions, subsSearch, subsFilterValues])

  const planColumns: ColumnDef<(typeof plans)[number], any>[] = React.useMemo(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div>
            <p className="font-medium">{row.original.name}</p>
            {row.original.description && (
              <p className="text-xs text-muted-foreground">{row.original.description}</p>
            )}
          </div>
        ),
      },
      {
        accessorKey: 'amount',
        header: 'Amount',
        cell: ({ row }) => (
          <span className="font-medium">{formatAmount(row.original.amount, row.original.currency)}</span>
        ),
      },
      {
        id: 'interval',
        header: 'Interval',
        cell: ({ row }) => {
          const p = row.original
          return (
            <span className="text-sm">
              Every {p.intervalCount > 1 ? `${p.intervalCount} ` : ''}
              {p.intervalType.toLowerCase()}
              {p.intervalCount > 1 ? 's' : ''}
            </span>
          )
        },
      },
      {
        accessorKey: 'trialDays',
        header: 'Trial',
        cell: ({ row }) => (
          <span className="text-sm">
            {row.original.trialDays > 0 ? `${row.original.trialDays} days` : '-'}
          </span>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'createdAt',
        header: 'Created',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {new Date(row.original.createdAt).toLocaleDateString()}
          </span>
        ),
      },
      {
        id: 'actions',
        header: 'Actions',
        enableHiding: false,
        cell: ({ row }) => {
          const p = row.original
          return p.status === 'ACTIVE' ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => archivePlan.mutate(p.id)}
              disabled={archivePlan.isPending}
            >
              Archive
            </Button>
          ) : null
        },
      },
    ],
    [archivePlan],
  )

  const subColumns: ColumnDef<(typeof subscriptions)[number], any>[] = React.useMemo(
    () => [
      {
        accessorKey: 'subscriberEmail',
        header: 'Email',
        cell: ({ row }) => <span className="font-medium">{row.original.subscriberEmail}</span>,
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        accessorKey: 'nextBillingDate',
        header: 'Next Billing',
        cell: ({ row }) => (
          <span className="text-sm">{new Date(row.original.nextBillingDate).toLocaleDateString()}</span>
        ),
      },
      {
        id: 'totalPaid',
        header: 'Total Paid',
        cell: ({ row }) => {
          const paid = formatDualAmount(row.original.totalPaidUsdt, undefined, undefined, primaryCurrency, liveRate)
          return (
            <div className="text-sm">
              <p className="font-medium">{paid.primary}</p>
              {paid.secondary && <p className="text-xs text-muted-foreground">({paid.secondary})</p>}
            </div>
          )
        },
      },
      {
        accessorKey: 'billingCount',
        header: 'Billing Count',
        cell: ({ row }) => <span className="text-sm">{row.original.billingCount}</span>,
      },
      {
        accessorKey: 'createdAt',
        header: 'Subscribed',
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {new Date(row.original.createdAt).toLocaleDateString()}
          </span>
        ),
      },
      {
        id: 'actions',
        header: 'Actions',
        enableHiding: false,
        cell: ({ row }) => {
          const s = row.original
          return s.status === 'ACTIVE' || s.status === 'TRIAL' ? (
            <Button
              variant="ghost"
              size="sm"
              className="text-red-600"
              onClick={() => cancelSub.mutate({ id: s.id, reason: 'Cancelled by merchant' })}
              disabled={cancelSub.isPending}
            >
              Cancel
            </Button>
          ) : null
        },
      },
    ],
    [cancelSub, primaryCurrency, liveRate],
  )

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
        <StatCard title="Total Revenue" value={subRevenueFmt.primary} description={subRevenueFmt.secondary || 'From subscriptions'} icon={DollarSign} />
      </div>

      <Tabs defaultValue="plans">
        <TabsList>
          <TabsTrigger value="plans">Plans ({plans.length})</TabsTrigger>
          <TabsTrigger value="subscribers">Subscribers ({subscriptions.length})</TabsTrigger>
        </TabsList>

        <TabsContent value="plans" className="mt-4">
          <DataTable
            columns={planColumns}
            data={filteredPlans}
            filters={plansFilters}
            filterValues={plansFilterValues}
            onFilterChange={(id, value) => setPlansFilterValues((prev) => ({ ...prev, [id]: value }))}
            onClearFilters={() => setPlansFilterValues({ status: '' })}
            search={plansSearch}
            onSearchChange={setPlansSearch}
            searchPlaceholder="Search plans..."
            pagination={{ page: 1, perPage: 999, total: filteredPlans.length }}
            onPageChange={() => {}}
          />
        </TabsContent>

        <TabsContent value="subscribers" className="mt-4">
          <DataTable
            columns={subColumns}
            data={filteredSubs}
            filters={subsFilters}
            filterValues={subsFilterValues}
            onFilterChange={(id, value) => setSubsFilterValues((prev) => ({ ...prev, [id]: value }))}
            onClearFilters={() => setSubsFilterValues({ status: '' })}
            search={subsSearch}
            onSearchChange={setSubsSearch}
            searchPlaceholder="Search subscribers..."
            pagination={{ page: 1, perPage: 999, total: filteredSubs.length }}
            onPageChange={() => {}}
          />
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
                        <SelectItem value="USDC">USDC</SelectItem>
                        <SelectItem value="BTC">BTC</SelectItem>
                        <SelectItem value="ETH">ETH</SelectItem>
                        <SelectItem value="BNB">BNB</SelectItem>
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
