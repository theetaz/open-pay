import { useState, useEffect, useRef, useCallback } from 'react'
import { Card, CardContent } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Switch } from '#/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '#/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from '#/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '#/components/ui/table'
import { Separator } from '#/components/ui/separator'
import { PageHeader } from '#/components/dashboard/page-header'
import { StatCard } from '#/components/dashboard/stat-card'
import { StatusBadge } from '#/components/dashboard/status-badge'
import { EmptyState } from '#/components/dashboard/empty-state'
import { Plus, Link2, Loader2, Check, X } from 'lucide-react'
import { usePaymentLinks, useCreatePaymentLink, checkSlugAvailability } from '#/hooks/use-payment-links'

function toSlug(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

type SlugStatus = 'idle' | 'checking' | 'available' | 'taken' | 'invalid'

export function PaymentLinksPage() {
  const [dialogOpen, setDialogOpen] = useState(false)
  const { data: linksData } = usePaymentLinks({ perPage: 100 })
  const links = linksData?.data || []
  const total = linksData?.meta?.total || 0
  const activeLinks = links.filter((l) => l.status === 'ACTIVE')
  const totalUsed = links.reduce((sum, l) => sum + l.usageCount, 0)

  return (
    <>
      <PageHeader
        title="Payment Links"
        description="Create and manage shareable payment links for your customers"
        action={
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 size-4" /> Create Link
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <StatCard title="Total Links" value={String(total)} description={`${activeLinks.length} active`} icon={Link2} />
        <StatCard title="Active Links" value={String(activeLinks.length)} description="Currently available" icon={Link2} />
        <StatCard title="Total Used" value={String(totalUsed)} description="Times links were used" icon={Link2} />
      </div>

      <div className="mb-4">
        <h3 className="text-lg font-semibold">All Payment Links</h3>
        <p className="text-sm text-muted-foreground">Manage and track all your payment links in one place</p>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Slug</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Used</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {links.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>
                    <EmptyState message="No payment links yet. Create your first one." />
                  </TableCell>
                </TableRow>
              ) : (
                links.map((link) => (
                  <PaymentLinkRow key={link.id} link={link} />
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <CreatePaymentLinkDialog open={dialogOpen} onOpenChange={setDialogOpen} />
    </>
  )
}

function PaymentLinkRow({ link }: { link: import('#/hooks/use-payment-links').PaymentLink }) {
  return (
    <TableRow>
      <TableCell className="font-medium">{link.name}</TableCell>
      <TableCell className="font-mono text-xs text-muted-foreground">{link.slug}</TableCell>
      <TableCell>
        {link.allowCustomAmount ? 'Custom' : `${link.amount} ${link.currency}`}
      </TableCell>
      <TableCell>
        <StatusBadge status={link.status} />
      </TableCell>
      <TableCell>{link.usageCount}</TableCell>
      <TableCell className="text-muted-foreground text-sm">
        {new Date(link.createdAt).toLocaleDateString()}
      </TableCell>
    </TableRow>
  )
}

function CreatePaymentLinkDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const createMutation = useCreatePaymentLink()

  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false)
  const [slugStatus, setSlugStatus] = useState<SlugStatus>('idle')
  const [description, setDescription] = useState('')
  const [currency, setCurrency] = useState('LKR')
  const [amount, setAmount] = useState('')
  const [allowCustomAmount, setAllowCustomAmount] = useState(false)
  const [isReusable, setIsReusable] = useState(false)
  const [showOnQrPage, setShowOnQrPage] = useState(false)
  const [expireAt, setExpireAt] = useState('')

  const debounceRef = useRef<ReturnType<typeof setTimeout>>(null)

  const resetForm = useCallback(() => {
    setName('')
    setSlug('')
    setSlugManuallyEdited(false)
    setSlugStatus('idle')
    setDescription('')
    setCurrency('LKR')
    setAmount('')
    setAllowCustomAmount(false)
    setIsReusable(false)
    setShowOnQrPage(false)
    setExpireAt('')
  }, [])

  // Auto-generate slug from name
  useEffect(() => {
    if (!slugManuallyEdited) {
      const generated = toSlug(name)
      setSlug(generated)
    }
  }, [name, slugManuallyEdited])

  // Debounced slug availability check
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)

    if (!slug || slug.length < 2) {
      setSlugStatus('idle')
      return
    }

    if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug)) {
      setSlugStatus('invalid')
      return
    }

    setSlugStatus('checking')
    debounceRef.current = setTimeout(async () => {
      try {
        const res = await checkSlugAvailability(slug)
        setSlugStatus(res.data.available ? 'available' : 'taken')
      } catch {
        setSlugStatus('idle')
      }
    }, 500)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [slug])

  const handleSlugChange = (value: string) => {
    setSlugManuallyEdited(true)
    setSlug(value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
  }

  const canSubmit =
    name.trim() &&
    slug.trim() &&
    (allowCustomAmount || (amount && parseFloat(amount) > 0)) &&
    slugStatus !== 'taken' &&
    slugStatus !== 'invalid' &&
    slugStatus !== 'checking' &&
    !createMutation.isPending

  const handleSubmit = () => {
    if (!canSubmit) return

    createMutation.mutate(
      {
        name: name.trim(),
        slug: slug.trim(),
        description: description.trim(),
        currency,
        amount: allowCustomAmount ? '0.01' : amount,
        allowCustomAmount,
        isReusable,
        showOnQrPage,
        expireAt: expireAt ? new Date(expireAt).toISOString() : undefined,
      },
      {
        onSuccess: () => {
          resetForm()
          onOpenChange(false)
        },
      },
    )
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) resetForm()
        onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Payment Link</DialogTitle>
          <DialogDescription>Create a new payment link for your customers</DialogDescription>
        </DialogHeader>

        <div className="space-y-5 py-2">
          <Separator />
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Basic Information
          </p>

          <div className="space-y-2">
            <Label htmlFor="link-name">Name *</Label>
            <Input
              id="link-name"
              placeholder="e.g. Premium Plan"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="link-slug">Slug *</Label>
            <div className="relative">
              <Input
                id="link-slug"
                placeholder="premium-plan"
                value={slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                className={
                  slugStatus === 'taken' || slugStatus === 'invalid'
                    ? 'border-destructive pr-9'
                    : slugStatus === 'available'
                      ? 'border-green-500 pr-9'
                      : 'pr-9'
                }
              />
              <div className="absolute right-3 top-1/2 -translate-y-1/2">
                {slugStatus === 'checking' && <Loader2 className="size-4 animate-spin text-muted-foreground" />}
                {slugStatus === 'available' && <Check className="size-4 text-green-500" />}
                {slugStatus === 'taken' && <X className="size-4 text-destructive" />}
                {slugStatus === 'invalid' && <X className="size-4 text-destructive" />}
              </div>
            </div>
            {slugStatus === 'taken' && (
              <p className="text-xs text-destructive">This slug is already in use. Choose a different one.</p>
            )}
            {slugStatus === 'invalid' && (
              <p className="text-xs text-destructive">Slug must be lowercase letters, numbers, and hyphens only.</p>
            )}
            {slugStatus === 'available' && (
              <p className="text-xs text-green-600 dark:text-green-400">Slug is available.</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="link-desc">Description</Label>
            <Textarea
              id="link-desc"
              placeholder="Item Description (supports Markdown)"
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>Currency *</Label>
            <Select value={currency} onValueChange={(v) => v && setCurrency(v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="LKR">LKR (Rs.)</SelectItem>
                <SelectItem value="USDT">USDT (T)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Separator />
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Amount Settings
          </p>

          <div className="flex items-center justify-between">
            <div>
              <Label>Allow custom amount</Label>
              <p className="text-xs text-muted-foreground">Let customers enter their own amount</p>
            </div>
            <Switch checked={allowCustomAmount} onCheckedChange={setAllowCustomAmount} />
          </div>

          {!allowCustomAmount && (
            <div className="space-y-2">
              <Label htmlFor="link-amount">Amount *</Label>
              <Input
                id="link-amount"
                type="number"
                step="0.01"
                min="0"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="link-expiry">Expiration Date</Label>
            <Input
              id="link-expiry"
              type="datetime-local"
              value={expireAt}
              onChange={(e) => setExpireAt(e.target.value)}
            />
          </div>

          <div className="flex items-center justify-between">
            <div>
              <Label>Reusable link</Label>
              <p className="text-xs text-muted-foreground">Allow multiple payments with this link</p>
            </div>
            <Switch checked={isReusable} onCheckedChange={setIsReusable} />
          </div>

          <div className="flex items-center justify-between">
            <div>
              <Label>Show on QR page</Label>
              <p className="text-xs text-muted-foreground">Display this payment link on your static QR page</p>
            </div>
            <Switch checked={showOnQrPage} onCheckedChange={setShowOnQrPage} />
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => {
              resetForm()
              onOpenChange(false)
            }}
          >
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!canSubmit}>
            {createMutation.isPending && <Loader2 className="mr-2 size-4 animate-spin" />}
            Create Payment Link
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
