import { useState, useMemo } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Switch } from '#/components/ui/switch'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '#/components/ui/dialog'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { DataTable, type FilterConfig } from '#/components/data-table'
import { Plus } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'

// Placeholder type until a real hook is available
interface Branch {
  id: string
  code: string
  name: string
  city: string
  postalCode: string
  status: string
  createdAt: string
}

export function BranchesPage() {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [filterValues, setFilterValues] = useState<Record<string, string | string[]>>({ status: '' })

  // Placeholder data - replace with actual hook when available
  const branches: Branch[] = []

  const filters: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { label: 'Active', value: 'ACTIVE' },
        { label: 'Inactive', value: 'INACTIVE' },
      ],
    },
  ]

  const handleFilterChange = (id: string, value: string | string[]) => {
    setFilterValues((prev) => ({ ...prev, [id]: value }))
  }

  const handleClearFilters = () => {
    setFilterValues({ status: '' })
  }

  const filteredData = useMemo(() => {
    let result = branches

    if (search) {
      const q = search.toLowerCase()
      result = result.filter(
        (b) =>
          b.name.toLowerCase().includes(q) ||
          b.code.toLowerCase().includes(q) ||
          b.city.toLowerCase().includes(q),
      )
    }

    const statusFilter = filterValues.status
    if (statusFilter && typeof statusFilter === 'string' && statusFilter !== '') {
      result = result.filter((b) => b.status === statusFilter)
    }

    return result
  }, [branches, search, filterValues])

  const columns: ColumnDef<Branch, any>[] = useMemo(
    () => [
      {
        accessorKey: 'code',
        header: 'Code',
        cell: ({ row }) => (
          <span className="font-mono text-sm">{row.original.code}</span>
        ),
      },
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
      },
      {
        accessorKey: 'city',
        header: 'City',
        cell: ({ row }) => <span className="text-sm">{row.original.city}</span>,
      },
      {
        accessorKey: 'postalCode',
        header: 'Postal Code',
        cell: ({ row }) => <span className="text-sm">{row.original.postalCode}</span>,
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
        cell: () => (
          <Button variant="ghost" size="sm">
            Edit
          </Button>
        ),
      },
    ],
    [],
  )

  return (
    <>
      <PageHeader
        title="Branches"
        description="Manage your branch locations and details"
        action={
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 size-4" /> Add Branch
          </Button>
        }
      />

      <div className="flex items-center justify-between mb-4">
        <div />
        <div className="text-right">
          <p className="text-2xl font-bold">{branches.length}</p>
          <p className="text-xs text-muted-foreground">Total Branches</p>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={filteredData}
        filters={filters}
        filterValues={filterValues}
        onFilterChange={handleFilterChange}
        onClearFilters={handleClearFilters}
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search by name or code..."
        pagination={{ page: 1, perPage: 999, total: filteredData.length }}
        onPageChange={() => {}}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Add New Branch</DialogTitle>
            <DialogDescription>Fill in the details to create a new branch.</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="branch-name">Branch Name *</Label>
                <Input id="branch-name" placeholder="Downtown Branch" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="branch-code">Branch Code *</Label>
                <Input id="branch-code" placeholder="DT-001" />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="branch-address">Address</Label>
              <Textarea id="branch-address" placeholder="123 Main Street" rows={2} />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="branch-city">City</Label>
                <Input id="branch-city" placeholder="Colombo" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="branch-postal">Postal Code</Label>
                <Input id="branch-postal" placeholder="00100" />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="branch-phone">Phone Number</Label>
              <Input id="branch-phone" placeholder="+94 77 123 4567" />
            </div>

            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <Label>Use merchant bank details</Label>
                <p className="text-xs text-muted-foreground">Disable to configure a branch-specific bank account</p>
              </div>
              <Switch defaultChecked />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button>Create Branch</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
