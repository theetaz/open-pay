import type { ColumnDef } from '@tanstack/react-table'

export interface FilterOption {
  label: string
  value: string
}

export interface FilterConfig {
  id: string
  label: string
  type: 'select' | 'multi-select' | 'search'
  options?: FilterOption[]
  placeholder?: string
}

export interface ServerPaginationState {
  page: number
  perPage: number
  total: number
}

export interface DataTableProps<TData> {
  columns: ColumnDef<TData, any>[]
  data: TData[]
  filters: FilterConfig[]
  filterValues: Record<string, string | string[]>
  onFilterChange: (id: string, value: string | string[]) => void
  onClearFilters: () => void
  search: string
  onSearchChange: (value: string) => void
  searchPlaceholder?: string
  pagination: ServerPaginationState
  onPageChange: (page: number) => void
  isLoading?: boolean
  onRowClick?: (row: TData) => void
}
