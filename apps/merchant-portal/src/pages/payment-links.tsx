import { useState, useEffect, useRef, useCallback } from 'react'
import { QRCodeSVG } from 'qrcode.react'
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
import { CopyButton } from '#/components/dashboard/copy-button'
import { Plus, Link2, Loader2, Check, X, ExternalLink, Trash2, QrCode } from 'lucide-react'
import { usePaymentLinks, useCreatePaymentLink, useDeletePaymentLink, checkSlugAvailability } from '#/hooks/use-payment-links'
import type { PaymentLink } from '#/hooks/use-payment-links'
import { toast } from 'sonner'

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

function getPaymentLinkUrl(slug: string) {
  return `${window.location.origin}/pay/${slug}`
}

export function PaymentLinksPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [detailLink, setDetailLink] = useState<PaymentLink | null>(null)
  const { data: linksData } = usePaymentLinks({ perPage: 100 })
  const deleteMutation = useDeletePaymentLink()
  const links = linksData?.data || []
  const total = linksData?.meta?.total || 0
  const activeLinks = links.filter((l) => l.status === 'ACTIVE')
  const totalUsed = links.reduce((sum, l) => sum + l.usageCount, 0)

  const handleDelete = (link: PaymentLink) => {
    if (confirm(`Delete "${link.name}"? This cannot be undone.`)) {
      deleteMutation.mutate(link.id)
    }
  }

  return (
    <>
      <PageHeader
        title="Payment Links"
        description="Create and manage shareable payment links for your customers"
        action={
          <Button onClick={() => setCreateOpen(true)}>
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
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {links.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7}>
                    <EmptyState message="No payment links yet. Create your first one." />
                  </TableCell>
                </TableRow>
              ) : (
                links.map((link) => (
                  <TableRow key={link.id} className="cursor-pointer" onClick={() => setDetailLink(link)}>
                    <TableCell className="font-medium">{link.name}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{link.slug}</TableCell>
                    <TableCell>
                      {link.allowCustomAmount ? 'Custom' : `${parseFloat(link.amount).toLocaleString()} ${link.currency}`}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={link.status} />
                    </TableCell>
                    <TableCell>{link.usageCount}</TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {new Date(link.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1" onClick={(e) => e.stopPropagation()}>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="size-8 p-0"
                          onClick={() => {
                            navigator.clipboard.writeText(getPaymentLinkUrl(link.slug))
                            toast.success('Payment link copied to clipboard')
                          }}
                        >
                          <Link2 className="size-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="size-8 p-0"
                          onClick={() => setDetailLink(link)}
                        >
                          <QrCode className="size-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="size-8 p-0 text-destructive hover:text-destructive"
                          onClick={() => handleDelete(link)}
                        >
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <CreatePaymentLinkDialog open={createOpen} onOpenChange={setCreateOpen} />
      <PaymentLinkDetailDialog link={detailLink} onClose={() => setDetailLink(null)} />
    </>
  )
}

function PaymentLinkDetailDialog({ link, onClose }: { link: PaymentLink | null; onClose: () => void }) {
  if (!link) return null

  const payUrl = getPaymentLinkUrl(link.slug)

  return (
    <Dialog open={!!link} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{link.name}</DialogTitle>
          <DialogDescription>{link.description || 'Payment link details'}</DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-2">
          {/* QR Code */}
          <div className="flex flex-col items-center">
            <div className="rounded-lg border bg-white p-4">
              <QRCodeSVG value={payUrl} size={200} level="M" />
            </div>
            <p className="text-xs text-muted-foreground mt-2">Scan to open payment page</p>
          </div>

          {/* Share URL */}
          <div className="space-y-2">
            <Label>Payment URL</Label>
            <div className="flex items-center gap-2">
              <Input value={payUrl} readOnly className="font-mono text-xs" />
              <CopyButton value={payUrl} />
              <Button variant="outline" size="sm" className="shrink-0" onClick={() => window.open(payUrl, '_blank')}>
                <ExternalLink className="size-4" />
              </Button>
            </div>
          </div>

          <Separator />

          {/* Details */}
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <p className="text-xs text-muted-foreground">Amount</p>
              <p className="font-medium">
                {link.allowCustomAmount ? 'Custom' : `${parseFloat(link.amount).toLocaleString()} ${link.currency}`}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Status</p>
              <StatusBadge status={link.status} />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Times Used</p>
              <p className="font-medium">{link.usageCount}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Reusable</p>
              <p className="font-medium">{link.isReusable ? 'Yes' : 'No (single use)'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Created</p>
              <p className="font-medium">{new Date(link.createdAt).toLocaleDateString()}</p>
            </div>
            {link.expireAt && (
              <div>
                <p className="text-xs text-muted-foreground">Expires</p>
                <p className="font-medium">{new Date(link.expireAt).toLocaleDateString()}</p>
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
