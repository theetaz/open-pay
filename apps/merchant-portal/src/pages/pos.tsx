import * as React from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { PageHeader } from '#/components/dashboard/page-header'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Loader2, QrCode, CheckCircle, Copy } from 'lucide-react'
import { useCreatePayment } from '#/hooks/use-payments'
import { formatAmount } from '#/lib/currency'

export function POSPage() {
  const [amount, setAmount] = React.useState('')
  const [currency, setCurrency] = React.useState('USDT')
  const [provider, setProvider] = React.useState('TEST')
  const [customerEmail, setCustomerEmail] = React.useState('')
  const [orderRef, setOrderRef] = React.useState('')
  const [generatedPayment, setGeneratedPayment] = React.useState<any>(null)

  const createPayment = useCreatePayment()

  function handleGenerateQR(e: React.FormEvent) {
    e.preventDefault()
    createPayment.mutate(
      {
        amount,
        currency,
        provider,
        customerEmail: customerEmail || undefined,
        merchantTradeNo: orderRef || undefined,
      },
      {
        onSuccess: (data) => {
          setGeneratedPayment(data.data)
        },
      },
    )
  }

  function handleNewPayment() {
    setGeneratedPayment(null)
    setAmount('')
    setCustomerEmail('')
    setOrderRef('')
  }

  return (
    <>
      <PageHeader
        title="Point of Sale"
        description="Generate QR codes for in-store crypto payments"
      />

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Left: Form */}
        <Card>
          <CardHeader>
            <CardTitle>New POS Payment</CardTitle>
            <CardDescription>Enter the payment details to generate a QR code for the customer</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleGenerateQR}>
              <FieldGroup>
                {createPayment.isError && (
                  <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    {createPayment.error.message}
                  </div>
                )}

                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Amount</FieldLabel>
                    <Input
                      type="number"
                      step="0.01"
                      min="0.01"
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                      placeholder="0.00"
                      required
                      className="text-2xl h-14"
                    />
                  </Field>
                  <Field>
                    <FieldLabel>Currency</FieldLabel>
                    <Select value={currency} onValueChange={(v) => v && setCurrency(v)}>
                      <SelectTrigger className="h-14"><SelectValue /></SelectTrigger>
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
                  <FieldLabel>Payment Provider</FieldLabel>
                  <Select value={provider} onValueChange={(v) => v && setProvider(v)}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="BYBIT">Bybit Pay</SelectItem>
                      <SelectItem value="BINANCE">Binance Pay</SelectItem>
                      <SelectItem value="KUCOIN">KuCoin</SelectItem>
                      <SelectItem value="TEST">Test (Sandbox)</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>

                <Field>
                  <FieldLabel>Customer Email (optional)</FieldLabel>
                  <Input type="email" value={customerEmail} onChange={(e) => setCustomerEmail(e.target.value)} placeholder="customer@example.com" />
                </Field>

                <Field>
                  <FieldLabel>Order Reference (optional)</FieldLabel>
                  <Input value={orderRef} onChange={(e) => setOrderRef(e.target.value)} placeholder="POS-001" />
                </Field>

                <Button type="submit" className="w-full h-12 text-lg" disabled={createPayment.isPending || !amount}>
                  {createPayment.isPending ? (
                    <><Loader2 className="mr-2 h-5 w-5 animate-spin" /> Generating...</>
                  ) : (
                    <><QrCode className="mr-2 h-5 w-5" /> Generate QR Code</>
                  )}
                </Button>
              </FieldGroup>
            </form>
          </CardContent>
        </Card>

        {/* Right: QR Display */}
        <Card>
          <CardHeader>
            <CardTitle>Payment QR Code</CardTitle>
            <CardDescription>Show this to the customer to scan and pay</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center justify-center min-h-[400px]">
            {generatedPayment ? (
              <div className="text-center space-y-4">
                {generatedPayment.status === 'PAID' ? (
                  <div className="flex flex-col items-center gap-4">
                    <CheckCircle className="size-20 text-green-500" />
                    <p className="text-2xl font-bold text-green-600">Payment Received!</p>
                  </div>
                ) : (
                  <>
                    <div className="bg-white p-6 rounded-xl border-2 border-dashed border-muted inline-block">
                      <div className="w-48 h-48 bg-muted/30 rounded-lg flex items-center justify-center">
                        <QrCode className="size-32 text-muted-foreground" />
                      </div>
                    </div>
                    <p className="text-3xl font-bold">
                      {formatAmount(generatedPayment.amount || generatedPayment.amountUsdt, generatedPayment.currency || 'USDT')}
                    </p>
                  </>
                )}

                <div className="text-sm text-muted-foreground space-y-1">
                  <p>Payment: <span className="font-mono">{generatedPayment.paymentNo}</span></p>
                  <p>Status: <span className="font-semibold">{generatedPayment.status}</span></p>
                  {generatedPayment.checkoutLink && (
                    <div className="flex items-center gap-2 justify-center">
                      <p className="text-xs truncate max-w-[200px]">{generatedPayment.checkoutLink}</p>
                      <Button variant="ghost" size="sm" onClick={() => navigator.clipboard.writeText(generatedPayment.checkoutLink)}>
                        <Copy className="size-3" />
                      </Button>
                    </div>
                  )}
                </div>

                <Button variant="outline" onClick={handleNewPayment} className="mt-4">
                  New Payment
                </Button>
              </div>
            ) : (
              <div className="text-center text-muted-foreground">
                <QrCode className="size-24 mx-auto mb-4 opacity-20" />
                <p>Enter payment details and click Generate QR Code</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  )
}
