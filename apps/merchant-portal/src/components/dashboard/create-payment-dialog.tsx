import * as React from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import { Plus, Loader2, Copy, CheckCircle2 } from 'lucide-react'
import { useCreatePayment } from '#/hooks/use-payments'
import { useMe } from '#/hooks/use-auth'
import { formatDualAmount } from '#/lib/currency'

export function CreatePaymentDialog() {
  const [open, setOpen] = React.useState(false)
  const [amount, setAmount] = React.useState('')
  const [currency, setCurrency] = React.useState('USDT')
  const [merchantTradeNo, setMerchantTradeNo] = React.useState('')
  const [customerEmail, setCustomerEmail] = React.useState('')
  const [createdPayment, setCreatedPayment] = React.useState<any>(null)
  const [copied, setCopied] = React.useState(false)

  const createPayment = useCreatePayment()
  const { data: meData } = useMe()
  const primaryCurrency = meData?.data?.merchant?.defaultCurrency || 'LKR'

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    createPayment.mutate(
      {
        amount,
        currency,
        provider: 'TEST',
        merchantTradeNo: merchantTradeNo || undefined,
        customerEmail: customerEmail || undefined,
      },
      {
        onSuccess: (res) => {
          setCreatedPayment(res.data)
        },
      },
    )
  }

  function handleClose() {
    setOpen(false)
    setCreatedPayment(null)
    setAmount('')
    setMerchantTradeNo('')
    setCustomerEmail('')
    setCopied(false)
  }

  function copyCheckoutUrl() {
    if (!createdPayment) return
    const url = `${window.location.origin}/checkout/${createdPayment.id}`
    navigator.clipboard.writeText(url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="mr-2 size-4" /> New Payment
      </Button>
      <Dialog open={open} onOpenChange={(v) => !v && handleClose()}>
        <DialogContent>
          {createdPayment ? (
            <>
              <DialogHeader>
                <DialogTitle>Payment Created</DialogTitle>
                <DialogDescription>Share the checkout link with your customer</DialogDescription>
              </DialogHeader>
              <div className="py-4 space-y-4">
                <div className="rounded-md bg-green-50 dark:bg-green-900/20 p-4 text-center">
                  <CheckCircle2 className="size-8 text-green-500 mx-auto mb-2" />
                  <p className="font-medium">{createdPayment.paymentNo}</p>
                  {(() => {
                    const amt = formatDualAmount(createdPayment.amountUsdt, createdPayment.amount, createdPayment.currency, primaryCurrency, createdPayment.exchangeRate)
                    return (
                      <>
                        <p className="text-2xl font-bold mt-1">{amt.primary}</p>
                        {amt.secondary && <p className="text-sm text-muted-foreground">({amt.secondary})</p>}
                      </>
                    )
                  })()}
                </div>

                <div className="rounded-md bg-muted/50 p-3 space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Status</span>
                    <span className="font-medium">{createdPayment.status}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Provider</span>
                    <span>{createdPayment.provider}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Expires</span>
                    <span>{new Date(createdPayment.expireTime).toLocaleTimeString()}</span>
                  </div>
                </div>

                <Button className="w-full" variant="outline" onClick={copyCheckoutUrl}>
                  {copied ? (
                    <><CheckCircle2 className="mr-2 size-4 text-green-500" /> Copied!</>
                  ) : (
                    <><Copy className="mr-2 size-4" /> Copy Checkout Link</>
                  )}
                </Button>
              </div>
              <DialogFooter>
                <Button onClick={handleClose}>Done</Button>
              </DialogFooter>
            </>
          ) : (
            <form onSubmit={handleSubmit}>
              <DialogHeader>
                <DialogTitle>Create Payment</DialogTitle>
                <DialogDescription>Create a new payment for a customer</DialogDescription>
              </DialogHeader>
              <div className="py-4">
                <FieldGroup>
                  {createPayment.isError && (
                    <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                      {createPayment.error.message}
                    </div>
                  )}
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
                  <Field>
                    <FieldLabel>Order Reference (optional)</FieldLabel>
                    <Input value={merchantTradeNo} onChange={(e) => setMerchantTradeNo(e.target.value)} placeholder="ORDER-001" />
                  </Field>
                  <Field>
                    <FieldLabel>Customer Email (optional)</FieldLabel>
                    <Input type="email" value={customerEmail} onChange={(e) => setCustomerEmail(e.target.value)} placeholder="customer@example.com" />
                  </Field>
                </FieldGroup>
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={handleClose}>Cancel</Button>
                <Button type="submit" disabled={createPayment.isPending || !amount}>
                  {createPayment.isPending ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Creating...</> : 'Create Payment'}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
