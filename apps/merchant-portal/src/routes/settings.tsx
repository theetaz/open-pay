import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/settings')({ component: SettingsPage })

function SettingsPage() {
  return (
    <div className="p-6 max-w-3xl">
      <h2 className="text-2xl font-bold mb-6">Settings</h2>

      <div className="space-y-6">
        <section className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">API Keys</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Manage API keys for integrating with the payment API.
          </p>
          <button className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors">
            Generate New Key
          </button>
          <div className="mt-4 text-sm text-muted-foreground">No API keys created yet.</div>
        </section>

        <section className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">Webhook Configuration</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Configure your endpoint to receive payment event notifications (ED25519 signed).
          </p>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium block mb-1">Webhook URL</label>
              <input
                type="url"
                placeholder="https://your-site.com/api/webhook"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <button className="px-4 py-2 rounded-md bg-secondary text-secondary-foreground text-sm font-medium hover:bg-secondary/80 transition-colors">
              Save Webhook
            </button>
          </div>
        </section>

        <section className="rounded-lg border border-border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">Bank Details</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Bank account for LKR settlements.
          </p>
          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <label className="text-sm font-medium block mb-1">Bank Name</label>
              <input
                type="text"
                placeholder="Bank of Ceylon"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium block mb-1">Account Number</label>
              <input
                type="text"
                placeholder="1234567890"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="text-sm font-medium block mb-1">Account Holder Name</label>
              <input
                type="text"
                placeholder="Your business name"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>
          <button className="mt-3 px-4 py-2 rounded-md bg-secondary text-secondary-foreground text-sm font-medium hover:bg-secondary/80 transition-colors">
            Save Bank Details
          </button>
        </section>
      </div>
    </div>
  )
}
