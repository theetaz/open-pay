import * as React from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '#/components/ui/dropdown-menu'
import { Search, Filter, ChevronLeft, ChevronRight, Columns3, Loader2 } from 'lucide-react'
import { cn } from '#/lib/utils'
import { FilterSidebar } from './filter-sidebar'
import type { DataTableProps } from './types'

export function DataTable<TData>({
  columns,
  data,
  filters,
  filterValues,
  onFilterChange,
  onClearFilters,
  search,
  onSearchChange,
  searchPlaceholder = 'Search...',
  pagination,
  onPageChange,
  isLoading,
  onRowClick,
}: DataTableProps<TData>) {
  const [showFilters, setShowFilters] = React.useState(false)
  const [columnVisibility, setColumnVisibility] = React.useState<Record<string, boolean>>({})

  const table = useReactTable({
    data,
    columns,
    state: { columnVisibility },
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualFiltering: true,
  })

  const { page, perPage, total } = pagination
  const totalPages = Math.ceil(total / perPage)
  const from = total === 0 ? 0 : (page - 1) * perPage + 1
  const to = Math.min(page * perPage, total)

  const activeFilterCount = Object.values(filterValues).filter((v) =>
    Array.isArray(v) ? v.length > 0 : v !== '',
  ).length

  // Generate page numbers with ellipsis
  const pageNumbers = React.useMemo(() => {
    const pages: (number | '...')[] = []
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
      pages.push(1)
      if (page > 3) pages.push('...')
      for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
        pages.push(i)
      }
      if (page < totalPages - 2) pages.push('...')
      pages.push(totalPages)
    }
    return pages
  }, [page, totalPages])

  return (
    <div className="space-y-3">
      {/* Toolbar: Search + Filter toggle + Column visibility */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={searchPlaceholder}
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>

        <Button
          variant={showFilters ? 'default' : 'outline'}
          size="sm"
          onClick={() => setShowFilters((v) => !v)}
          className="shrink-0"
        >
          <Filter className="size-4 mr-1.5" />
          Filters
          {activeFilterCount > 0 && (
            <span className="ml-1.5 size-5 rounded-full bg-primary-foreground text-primary text-xs flex items-center justify-center font-medium">
              {activeFilterCount}
            </span>
          )}
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger render={<Button variant="outline" size="sm" />}>
            <Columns3 className="size-4 mr-1.5" /> Columns
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="min-w-[160px]">
            {table.getAllColumns().filter((col) => col.getCanHide()).map((col) => (
              <DropdownMenuItem
                key={col.id}
                onClick={() => col.toggleVisibility(!col.getIsVisible())}
              >
                <span className={cn(
                  'size-3.5 rounded-sm border mr-2 flex items-center justify-center text-[10px]',
                  col.getIsVisible() ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/40',
                )}>
                  {col.getIsVisible() && '✓'}
                </span>
                {typeof col.columnDef.header === 'string' ? col.columnDef.header : col.id}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Content: Filter sidebar + Table */}
      <div className="flex rounded-lg border overflow-hidden bg-card">
        {showFilters && (
          <FilterSidebar
            filters={filters}
            values={filterValues}
            onChange={onFilterChange}
            onClear={onClearFilters}
          />
        )}

        <div className="flex-1 min-w-0">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-32 text-center">
                    <div className="flex items-center justify-center gap-2 text-muted-foreground">
                      <Loader2 className="size-4 animate-spin" /> Loading...
                    </div>
                  </TableCell>
                </TableRow>
              ) : table.getRowModel().rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-32 text-center text-muted-foreground">
                    No results found.
                  </TableCell>
                </TableRow>
              ) : (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    className={cn(onRowClick && 'cursor-pointer')}
                    onClick={() => onRowClick?.(row.original)}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          {total > 0 && (
            <div className="flex items-center justify-between border-t px-4 py-3">
              <p className="text-sm text-muted-foreground">
                {from}–{to} of {total} results
              </p>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(page - 1)}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="size-4" /> Previous
                </Button>

                {pageNumbers.map((p, i) =>
                  p === '...' ? (
                    <span key={`ellipsis-${i}`} className="px-2 text-sm text-muted-foreground">
                      ...
                    </span>
                  ) : (
                    <Button
                      key={p}
                      variant={p === page ? 'default' : 'outline'}
                      size="sm"
                      className="min-w-[32px]"
                      onClick={() => onPageChange(p)}
                    >
                      {p}
                    </Button>
                  ),
                )}

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(page + 1)}
                  disabled={page >= totalPages}
                >
                  Next <ChevronRight className="size-4" />
                </Button>
              </div>
              <p className="text-sm text-muted-foreground">
                Page {page} of {totalPages}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
