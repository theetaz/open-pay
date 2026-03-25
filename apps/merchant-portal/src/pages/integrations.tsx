import * as React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Badge } from '#/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Field, FieldGroup, FieldLabel } from '#/components/ui/field'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '#/components/ui/dialog'
import {
  Key, Plus, Copy, Check, Trash2, Globe, Send, AlertTriangle, Eye, EyeOff, Code2,
} from 'lucide-react'
import { api } from '#/lib/api'
import { toast } from 'sonner'

export function IntegrationsPage() {
  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Integrations</h1>
        <p className="text-sm text-muted-foreground">Manage API keys, webhooks, and SDK integration</p>
      </div>

      <Tabs defaultValue="api-keys">
        <TabsList>
          <TabsTrigger value="api-keys">API Keys</TabsTrigger>
          <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
          <TabsTrigger value="quickstart">Quick Start</TabsTrigger>
        </TabsList>

        <TabsContent value="api-keys" className="mt-4">
          <APIKeysSection />
        </TabsContent>

        <TabsContent value="webhooks" className="mt-4">
          <WebhooksSection />
        </TabsContent>

        <TabsContent value="quickstart" className="mt-4">
          <QuickStartSection />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ─── API Keys Section ───

function APIKeysSection() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = React.useState(false)
  const [newKeyName, setNewKeyName] = React.useState('')
  const [newKeyEnv, setNewKeyEnv] = React.useState('live')
  const [createdKey, setCreatedKey] = React.useState<string | null>(null)
  const [copied, setCopied] = React.useState(false)

  const { data } = useQuery({
    queryKey: ['api-keys'],
    queryFn: () => api.get<{ data: any[] }>('/v1/api-keys'),
  })
  const keys = data?.data || []

  const createKey = useMutation({
    mutationFn: (data: { name: string; environment: string }) =>
      api.post<{ data: { secret: string } }>('/v1/api-keys', data),
    onSuccess: (res) => {
      setCreatedKey(res.data.secret)
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      setShowCreate(false)
      setNewKeyName('')
    },
    onError: (err: any) => toast.error(err.message),
  })

  const revokeKey = useMutation({
    mutationFn: (id: string) => api.delete(`/v1/api-keys/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
      toast.success('API key revoked')
    },
    onError: (err: any) => toast.error(err.message),
  })

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>API Keys</CardTitle>
            <CardDescription>Create API keys to authenticate SDK and API requests</CardDescription>
          </div>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus className="size-4 mr-1" /> Create Key
          </Button>
        </CardHeader>
        <CardContent>
          {keys.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Key className="size-10 mx-auto mb-3 opacity-40" />
              <p className="text-sm">No API keys yet. Create one to start integrating.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {keys.map((k: any) => (
                <div key={k.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-3">
                    <Key className="size-5 text-muted-foreground" />
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">{k.name || 'Unnamed Key'}</p>
                        <Badge variant="outline" className="text-xs">{k.environment}</Badge>
                        {!k.isActive && <Badge className="bg-red-500/10 text-red-600 text-xs">Revoked</Badge>}
                      </div>
                      <p className="text-xs text-muted-foreground font-mono mt-0.5">{k.keyId}</p>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        Created {new Date(k.createdAt).toLocaleDateString()}
                        {k.lastUsedAt && ` · Last used ${new Date(k.lastUsedAt).toLocaleDateString()}`}
                      </p>
                    </div>
                  </div>
                  {k.isActive && (
                    <Button variant="ghost" size="sm" className="text-destructive" onClick={() => revokeKey.mutate(k.id)}>
                      <Trash2 className="size-4" />
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Key Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create API Key</DialogTitle>
            <DialogDescription>This key will be used to authenticate SDK and API requests.</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Key Name</FieldLabel>
              <Input placeholder="e.g. Production, Staging" value={newKeyName} onChange={(e) => setNewKeyName(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Environment</FieldLabel>
              <div className="flex gap-2">
                <Button variant={newKeyEnv === 'live' ? 'default' : 'outline'} size="sm" onClick={() => setNewKeyEnv('live')}>Live</Button>
                <Button variant={newKeyEnv === 'test' ? 'default' : 'outline'} size="sm" onClick={() => setNewKeyEnv('test')}>Test</Button>
              </div>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={() => createKey.mutate({ name: newKeyName, environment: newKeyEnv })} disabled={createKey.isPending}>
              Create Key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Created Key Display Dialog */}
      <Dialog open={!!createdKey} onOpenChange={() => setCreatedKey(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Check className="size-5 text-green-600" /> API Key Created
            </DialogTitle>
            <DialogDescription>
              Copy this key now — it won't be shown again.
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg border bg-muted/50 p-4">
            <div className="flex items-center justify-between gap-2">
              <code className="text-xs break-all flex-1">{createdKey}</code>
              <Button variant="ghost" size="sm" onClick={() => copyToClipboard(createdKey!)}>
                {copied ? <Check className="size-4 text-green-600" /> : <Copy className="size-4" />}
              </Button>
            </div>
          </div>
          <div className="rounded-lg bg-amber-500/10 border border-amber-300/30 p-3 flex items-start gap-2">
            <AlertTriangle className="size-4 text-amber-600 mt-0.5 shrink-0" />
            <p className="text-xs text-amber-700">Store this key securely. It cannot be retrieved after closing this dialog.</p>
          </div>
          <DialogFooter>
            <Button onClick={() => setCreatedKey(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ─── Webhooks Section ───

function WebhooksSection() {
  const queryClient = useQueryClient()
  const [webhookUrl, setWebhookUrl] = React.useState('')

  const configureWebhook = useMutation({
    mutationFn: (url: string) => api.post('/v1/webhooks/configure', { url }),
    onSuccess: () => {
      toast.success('Webhook configured')
      queryClient.invalidateQueries({ queryKey: ['webhook-deliveries'] })
    },
    onError: (err: any) => toast.error(err.message),
  })

  const testWebhook = useMutation({
    mutationFn: () => api.post<{ data: any }>('/v1/webhooks/test'),
    onSuccess: (res) => {
      const success = res.data?.success
      if (success) {
        toast.success('Test webhook delivered successfully')
      } else {
        toast.error('Test webhook delivery failed — check your endpoint')
      }
      queryClient.invalidateQueries({ queryKey: ['webhook-deliveries'] })
    },
    onError: (err: any) => toast.error(err.message),
  })

  const { data: publicKeyData } = useQuery({
    queryKey: ['webhook-public-key'],
    queryFn: () => api.get<{ data: { publicKey: string } }>('/v1/webhooks/public-key'),
    retry: false,
  })

  const { data: deliveriesData } = useQuery({
    queryKey: ['webhook-deliveries'],
    queryFn: () => api.get<{ data: any[]; meta: any }>('/v1/webhooks/deliveries?perPage=10'),
    retry: false,
  })

  const publicKey = publicKeyData?.data?.publicKey
  const deliveries = deliveriesData?.data || []

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Webhook Configuration</CardTitle>
          <CardDescription>Receive real-time notifications when payment events occur</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <FieldGroup>
            <Field>
              <FieldLabel>Webhook URL</FieldLabel>
              <div className="flex gap-2">
                <Input
                  placeholder="https://mysite.com/api/webhook"
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  className="flex-1"
                />
                <Button onClick={() => configureWebhook.mutate(webhookUrl)} disabled={!webhookUrl || configureWebhook.isPending}>
                  <Globe className="size-4 mr-1" /> Save
                </Button>
              </div>
            </Field>
          </FieldGroup>

          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => testWebhook.mutate()} disabled={testWebhook.isPending}>
              <Send className="size-4 mr-1" /> Send Test Webhook
            </Button>
          </div>

          {publicKey && (
            <div>
              <p className="text-xs text-muted-foreground mb-1">ED25519 Public Key (for signature verification)</p>
              <div className="rounded-lg border bg-muted/50 p-3">
                <code className="text-xs break-all">{publicKey}</code>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Delivery Log</CardTitle>
          <CardDescription>Recent webhook delivery attempts</CardDescription>
        </CardHeader>
        <CardContent>
          {deliveries.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">No webhook deliveries yet.</p>
          ) : (
            <div className="space-y-2">
              {deliveries.map((d: any) => (
                <div key={d.id} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <div className="flex items-center gap-2">
                      <Badge className={
                        d.status === 'DELIVERED' ? 'bg-green-500/10 text-green-600' :
                        d.status === 'EXHAUSTED' || d.status === 'FAILED' ? 'bg-red-500/10 text-red-600' :
                        'bg-amber-500/10 text-amber-600'
                      }>{d.status}</Badge>
                      <span className="text-sm font-medium">{d.eventType}</span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      Attempts: {d.attemptCount}/{d.maxAttempts}
                      {d.lastResponseCode > 0 && ` · HTTP ${d.lastResponseCode}`}
                      {d.lastError && ` · ${d.lastError}`}
                    </p>
                  </div>
                  <span className="text-xs text-muted-foreground">{new Date(d.createdAt).toLocaleString()}</span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// ─── Quick Start Section ───

function QuickStartSection() {
  const [showSecret, setShowSecret] = React.useState(false)

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><Code2 className="size-5" /> Quick Start Guide</CardTitle>
          <CardDescription>Get started with Open Pay SDK in under 5 minutes</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <h3 className="text-sm font-semibold mb-2">1. Install the SDK</h3>
            <CodeBlock lang="bash" code={`npm install @openpay/sdk    # TypeScript/JavaScript
pip install openpay-sdk     # Python
composer require openpay/sdk # PHP
# Go: go get github.com/openlankapay/openlankapay/sdks/sdk-go`} />
          </div>

          <div>
            <h3 className="text-sm font-semibold mb-2">2. Create a Checkout Session</h3>
            <CodeBlock lang="typescript" code={`import { OpenPay } from '@openpay/sdk'

const openpay = new OpenPay('YOUR_API_KEY', {
  baseURL: '${window.location.origin.replace('merchant', 'api')}',
})

const session = await openpay.checkout.createSession({
  amount: '2500.00',
  currency: 'LKR',
  successUrl: 'https://yoursite.com/success',
  cancelUrl: 'https://yoursite.com/cancel',
})

// Redirect customer to hosted checkout
window.location.href = session.url`} />
          </div>

          <div>
            <h3 className="text-sm font-semibold mb-2">3. Handle Webhooks</h3>
            <CodeBlock lang="typescript" code={`import { verifyWebhookSignature } from '@openpay/sdk'

app.post('/webhook', (req, res) => {
  const event = verifyWebhookSignature(req.body, req.headers, PUBLIC_KEY)

  if (event.event === 'payment.paid') {
    // Fulfill the order
  }

  res.sendStatus(200)
})`} />
          </div>

          <div className="rounded-lg bg-blue-500/10 border border-blue-300/20 p-4">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              <strong>Full documentation:</strong> Check the{' '}
              <a href="https://github.com/theetaz/open-pay/tree/develop/docs" className="underline" target="_blank" rel="noopener">
                developer docs
              </a>
              {' '}for API reference, webhook guide, and POS integration examples.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function CodeBlock({ code, lang }: { code: string; lang: string }) {
  const [copied, setCopied] = React.useState(false)

  const copy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative rounded-lg bg-muted/50 border">
      <div className="flex items-center justify-between px-3 py-1.5 border-b">
        <span className="text-xs text-muted-foreground">{lang}</span>
        <button onClick={copy} className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1">
          {copied ? <Check className="size-3" /> : <Copy className="size-3" />}
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      <pre className="p-3 overflow-x-auto text-xs"><code>{code}</code></pre>
    </div>
  )
}
