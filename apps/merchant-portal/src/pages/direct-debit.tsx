import * as React from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Plus, FileSignature, DollarSign, Loader2, RefreshCw, XCircle, Zap } from 'lucide-react'
import {
  useDirectDebitContracts,
  useScenarioCodes,
  useCreateContract,
  useSyncContract,
  useTerminateContract,
  useExecutePayment,
  type DirectDebitContract,
} from '#/hooks/use-direct-debit'
import { formatAmount } from '#/lib/currency'
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

export function DirectDebitPage() {
  const [statusFilter, setStatusFilter] = React.useState('')
  const [search, setSearch] = React.useState('')
  const { data: contractsData } = useDirectDebitContracts(statusFilter || undefined)
  const syncContract = useSyncContract()
  const terminateContract = useTerminateContract()

  const contracts = contractsData?.data || []
  const signedContracts = contracts.filter((c) => c.status === 'SIGNED')
  const totalCharged = contracts.reduce((sum, c) => sum + parseFloat(c.totalAmountCharged || '0'), 0)

  const filteredContracts = React.useMemo(() => {
    if (!search) return contracts
    const q = search.toLowerCase()
    return contracts.filter(
      (c) =>
        c.serviceName.toLowerCase().includes(q) ||
        c.merchantContractCode.toLowerCase().includes(q),
    )
  }, [contracts, search])

  const filters: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { label: 'Initiated', value: 'INITIATED' },
        { label: 'Signed', value: 'SIGNED' },
        { label: 'Terminated', value: 'TERMINATED' },
        { label: 'Expired', value: 'EXPIRED' },
      ],
    },
  ]
  const filterValues: Record<string, string | string[]> = { status: statusFilter }

  const columns: ColumnDef<DirectDebitContract, any>[] = React.useMemo(
    () => [
      {
        accessorKey: 'serviceName',
        header: 'Service',
        cell: ({ row }) => (
          <div>
            <p className="font-medium">{row.original.serviceName}</p>
            <p className="text-xs text-muted-foreground font-mono">{row.original.merchantContractCode}</p>
          </div>
        ),
      },
      {
        accessorKey: 'paymentProvider',
        header: 'Provider',
        cell: ({ row }) => (
          <span className="text-sm">{row.original.paymentProvider.replace('_', ' ')}</span>
        ),
      },
      {
        accessorKey: 'singleUpperLimit',
        header: 'Limit',
        cell: ({ row }) => (
          <span className="font-medium">{formatAmount(row.original.singleUpperLimit, row.original.currency)}</span>
        ),
      },
      {
        accessorKey: 'paymentCount',
        header: 'Payments',
        cell: ({ row }) => (
          <div className="text-sm">
            <p>{row.original.paymentCount} payments</p>
            <p className="text-xs text-muted-foreground">{formatAmount(row.original.totalAmountCharged, row.original.currency)} charged</p>
          </div>
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
          const c = row.original
          return (
            <div className="flex gap-1">
              {c.status === 'INITIATED' && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => syncContract.mutate(c.id)}
                  disabled={syncContract.isPending}
                  title="Sync status"
                >
                  <RefreshCw className="size-4" />
                </Button>
              )}
              {c.status === 'SIGNED' && (
                <>
                  <ExecutePaymentDialog contract={c} />
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-600"
                    onClick={() => terminateContract.mutate({ id: c.id, terminationNotes: 'Terminated by merchant' })}
                    disabled={terminateContract.isPending}
                    title="Terminate"
                  >
                    <XCircle className="size-4" />
                  </Button>
                </>
              )}
            </div>
          )
        },
      },
    ],
    [syncContract, terminateContract],
  )

  return (
    <>
      <PageHeader
        title="Direct Debit"
        description="Manage pre-authorized contracts and execute recurring charges"
        action={<CreateContractDialog />}
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Total Contracts" value={String(contracts.length)} description="All contracts" icon={FileSignature} />
        <StatCard title="Active Contracts" value={String(signedContracts.length)} description="Signed and active" icon={Zap} />
        <StatCard title="Total Charged" value={formatAmount(totalCharged, 'USDT')} description="From direct debits" icon={DollarSign} />
      </div>

      <DataTable
        columns={columns}
        data={filteredContracts}
        filters={filters}
        filterValues={filterValues}
        onFilterChange={(id, value) => {
          if (id === 'status') setStatusFilter(typeof value === 'string' ? value : '')
        }}
        onClearFilters={() => setStatusFilter('')}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search contracts..."
        pagination={{ page: 1, perPage: 999, total: filteredContracts.length }}
        onPageChange={() => {}}
      />
    </>
  )
}

function CreateContractDialog() {
  const [open, setOpen] = React.useState(false)
  const [serviceName, setServiceName] = React.useState('')
  const [scenarioId, setScenarioId] = React.useState('')
  const [singleUpperLimit, setSingleUpperLimit] = React.useState('')
  const [returnUrl, setReturnUrl] = React.useState('')
  const [cancelUrl, setCancelUrl] = React.useState('')
  const [webhookUrl, setWebhookUrl] = React.useState('')

  const { data: scenariosData } = useScenarioCodes()
  const createContract = useCreateContract()
  const scenarios = scenariosData?.data || []

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    createContract.mutate(
      { serviceName, scenarioId, singleUpperLimit, returnUrl, cancelUrl, webhookUrl: webhookUrl || undefined },
      {
        onSuccess: () => {
          setOpen(false)
          setServiceName('')
          setScenarioId('')
          setSingleUpperLimit('')
          setReturnUrl('')
          setCancelUrl('')
          setWebhookUrl('')
        },
      },
    )
  }

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="mr-2 size-4" /> Create Contract
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Create Direct Debit Contract</DialogTitle>
              <DialogDescription>Set up a pre-authorization for recurring charges</DialogDescription>
            </DialogHeader>

            <div className="py-4">
              <FieldGroup>
                {createContract.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {createContract.error.message}
                  </div>
                )}

                <Field>
                  <FieldLabel>Service Name</FieldLabel>
                  <Input value={serviceName} onChange={(e) => setServiceName(e.target.value)} placeholder="Monthly Subscription" required maxLength={64} />
                </Field>

                <Field>
                  <FieldLabel>Scenario</FieldLabel>
                  <Select value={scenarioId} onValueChange={(v) => v && setScenarioId(v)}>
                    <SelectTrigger><SelectValue placeholder="Select a scenario" /></SelectTrigger>
                    <SelectContent>
                      {scenarios.map((s) => (
                        <SelectItem key={s.id} value={s.id}>
                          {s.scenarioName} ({s.paymentProvider.replace('_', ' ')}) — max {s.maxLimit}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </Field>

                <Field>
                  <FieldLabel>Single Transaction Limit (USDT)</FieldLabel>
                  <Input type="number" step="0.01" min="0.01" value={singleUpperLimit} onChange={(e) => setSingleUpperLimit(e.target.value)} placeholder="100.00" required />
                </Field>

                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Return URL</FieldLabel>
                    <Input type="url" value={returnUrl} onChange={(e) => setReturnUrl(e.target.value)} placeholder="https://..." required />
                  </Field>
                  <Field>
                    <FieldLabel>Cancel URL</FieldLabel>
                    <Input type="url" value={cancelUrl} onChange={(e) => setCancelUrl(e.target.value)} placeholder="https://..." required />
                  </Field>
                </div>

                <Field>
                  <FieldLabel>Webhook URL (optional)</FieldLabel>
                  <Input type="url" value={webhookUrl} onChange={(e) => setWebhookUrl(e.target.value)} placeholder="https://..." />
                </Field>
              </FieldGroup>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={createContract.isPending}>
                {createContract.isPending ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Creating...</> : 'Create Contract'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}

function ExecutePaymentDialog({ contract }: { contract: DirectDebitContract }) {
  const [open, setOpen] = React.useState(false)
  const [amount, setAmount] = React.useState('')
  const [productName, setProductName] = React.useState('')
  const [productDetail, setProductDetail] = React.useState('')

  const executePayment = useExecutePayment()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    executePayment.mutate(
      { contractId: contract.id, amount, productName, productDetail: productDetail || undefined },
      {
        onSuccess: () => {
          setOpen(false)
          setAmount('')
          setProductName('')
          setProductDetail('')
        },
      },
    )
  }

  return (
    <>
      <Button variant="ghost" size="sm" onClick={() => setOpen(true)} title="Execute payment">
        <Zap className="size-4" />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Execute Payment</DialogTitle>
              <DialogDescription>
                Charge against contract: {contract.serviceName} (max {formatAmount(contract.singleUpperLimit, contract.currency)})
              </DialogDescription>
            </DialogHeader>

            <div className="py-4">
              <FieldGroup>
                {executePayment.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {executePayment.error.message}
                  </div>
                )}

                <Field>
                  <FieldLabel>Amount ({contract.currency})</FieldLabel>
                  <Input type="number" step="0.01" min="0.01" max={contract.singleUpperLimit} value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="25.00" required />
                </Field>

                <Field>
                  <FieldLabel>Product Name</FieldLabel>
                  <Input value={productName} onChange={(e) => setProductName(e.target.value)} placeholder="Monthly service fee" required maxLength={256} />
                </Field>

                <Field>
                  <FieldLabel>Product Detail (optional)</FieldLabel>
                  <Input value={productDetail} onChange={(e) => setProductDetail(e.target.value)} placeholder="March 2026 billing cycle" maxLength={256} />
                </Field>
              </FieldGroup>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={executePayment.isPending}>
                {executePayment.isPending ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Processing...</> : 'Execute Payment'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}
