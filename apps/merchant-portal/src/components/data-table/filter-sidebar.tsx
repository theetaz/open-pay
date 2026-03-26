import * as React from 'react'
import { ChevronDown, Search, X } from 'lucide-react'
import { Input } from '#/components/ui/input'
import { ScrollArea } from '#/components/ui/scroll-area'
import { cn } from '#/lib/utils'
import type { FilterConfig } from './types'

interface FilterSidebarProps {
  filters: FilterConfig[]
  values: Record<string, string | string[]>
  onChange: (id: string, value: string | string[]) => void
  onClear: () => void
}

export function FilterSidebar({ filters, values, onChange, onClear }: FilterSidebarProps) {
  const [collapsed, setCollapsed] = React.useState<Record<string, boolean>>({})
  const hasActiveFilters = Object.values(values).some((v) =>
    Array.isArray(v) ? v.length > 0 : v !== '',
  )

  const toggle = (id: string) =>
    setCollapsed((prev) => ({ ...prev, [id]: !prev[id] }))

  return (
    <div className="w-[240px] shrink-0 border-r">
      <div className="flex items-center justify-between px-3 py-2 border-b">
        <span className="text-sm font-medium">Filters</span>
        {hasActiveFilters && (
          <button
            onClick={onClear}
            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1"
          >
            <X className="size-3" /> Clear all
          </button>
        )}
      </div>
      <ScrollArea className="h-[calc(100vh-280px)]">
        <div className="p-1">
          {filters.map((filter) => {
            const isOpen = !collapsed[filter.id]
            const currentValue = values[filter.id]

            return (
              <div key={filter.id} className="border-b last:border-0">
                <button
                  onClick={() => toggle(filter.id)}
                  className="flex w-full items-center justify-between px-3 py-2.5 text-sm font-medium hover:bg-muted/50 rounded-md"
                >
                  <span className="flex items-center gap-2">
                    {filter.label}
                    {((typeof currentValue === 'string' && currentValue) ||
                      (Array.isArray(currentValue) && currentValue.length > 0)) && (
                      <span className="size-1.5 rounded-full bg-primary" />
                    )}
                  </span>
                  <ChevronDown
                    className={cn('size-4 text-muted-foreground transition-transform', isOpen && 'rotate-180')}
                  />
                </button>

                {isOpen && (
                  <div className="px-2 pb-2">
                    {filter.type === 'search' && (
                      <FilterSearch
                        value={(currentValue as string) || ''}
                        onChange={(v) => onChange(filter.id, v)}
                        placeholder={filter.placeholder || `Search ${filter.label.toLowerCase()}...`}
                      />
                    )}
                    {filter.type === 'select' && filter.options && (
                      <FilterRadioList
                        options={filter.options}
                        value={(currentValue as string) || ''}
                        onChange={(v) => onChange(filter.id, v)}
                      />
                    )}
                    {filter.type === 'multi-select' && filter.options && (
                      <FilterCheckboxList
                        options={filter.options}
                        values={(currentValue as string[]) || []}
                        onChange={(v) => onChange(filter.id, v)}
                      />
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}

function FilterSearch({
  value,
  onChange,
  placeholder,
}: {
  value: string
  onChange: (v: string) => void
  placeholder: string
}) {
  return (
    <div className="relative">
      <Search className="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
      <Input
        className="h-8 pl-7 text-xs"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  )
}

function FilterRadioList({
  options,
  value,
  onChange,
}: {
  options: { label: string; value: string }[]
  value: string
  onChange: (v: string) => void
}) {
  return (
    <div className="space-y-0.5">
      <button
        onClick={() => onChange('')}
        className={cn(
          'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-xs',
          value === '' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted/50',
        )}
      >
        <span
          className={cn(
            'size-3.5 rounded-full border-2 flex items-center justify-center',
            value === '' ? 'border-primary' : 'border-muted-foreground/40',
          )}
        >
          {value === '' && <span className="size-1.5 rounded-full bg-primary" />}
        </span>
        All
      </button>
      {options.map((opt) => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value === value ? '' : opt.value)}
          className={cn(
            'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-xs',
            value === opt.value
              ? 'bg-primary/10 text-primary font-medium'
              : 'text-muted-foreground hover:bg-muted/50',
          )}
        >
          <span
            className={cn(
              'size-3.5 rounded-full border-2 flex items-center justify-center',
              value === opt.value ? 'border-primary' : 'border-muted-foreground/40',
            )}
          >
            {value === opt.value && <span className="size-1.5 rounded-full bg-primary" />}
          </span>
          {opt.label}
        </button>
      ))}
    </div>
  )
}

function FilterCheckboxList({
  options,
  values,
  onChange,
}: {
  options: { label: string; value: string }[]
  values: string[]
  onChange: (v: string[]) => void
}) {
  const toggle = (val: string) => {
    onChange(values.includes(val) ? values.filter((v) => v !== val) : [...values, val])
  }

  return (
    <div className="space-y-0.5">
      {options.map((opt) => {
        const checked = values.includes(opt.value)
        return (
          <button
            key={opt.value}
            onClick={() => toggle(opt.value)}
            className={cn(
              'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-xs',
              checked ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted/50',
            )}
          >
            <span
              className={cn(
                'size-3.5 rounded-sm border-2 flex items-center justify-center text-[10px]',
                checked ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/40',
              )}
            >
              {checked && '✓'}
            </span>
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}
