import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { Separator } from '#/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldDescription,
} from '#/components/ui/field'
import { Checkbox } from '#/components/ui/checkbox'
import { Copy, Plus, Send, Trash2 } from 'lucide-react'

export const Route = createFileRoute('/_dashboard/example')({
  component: ExamplePage,
})

function ExamplePage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold">Component Preview</h1>
        <p className="text-muted-foreground">Testing preset abUBnP — Amber theme, Zinc base, JetBrains Mono</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Topic Card */}
        <Card>
          <CardHeader>
            <CardTitle>Topic</CardTitle>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <Select>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a topic" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="general">General</SelectItem>
                    <SelectItem value="billing">Billing</SelectItem>
                    <SelectItem value="support">Support</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Feedback Card */}
        <Card>
          <CardHeader>
            <CardTitle>Feedback</CardTitle>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="feedback">Feedback</FieldLabel>
                <Textarea id="feedback" placeholder="Your feedback helps us improve..." />
              </Field>
              <Field>
                <Button>Submit</Button>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Environment Variables Card */}
        <Card>
          <CardHeader>
            <CardTitle>Environment Variables</CardTitle>
            <CardDescription>Production · 3 variables</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between rounded-md bg-muted px-3 py-2">
                <span className="text-sm font-medium">DATABASE_URL</span>
                <span className="text-sm text-muted-foreground">••••••••</span>
              </div>
              <div className="flex items-center justify-between rounded-md bg-muted px-3 py-2">
                <span className="text-sm font-medium">NEXT_PUBLIC_API</span>
                <span className="text-sm text-muted-foreground">https://api.example.com</span>
              </div>
              <div className="flex items-center justify-between rounded-md bg-muted px-3 py-2">
                <span className="text-sm font-medium">STRIPE_SECRET</span>
                <span className="text-sm text-muted-foreground">••••••••</span>
              </div>
            </div>
          </CardContent>
          <CardFooter className="gap-2">
            <Button variant="outline">Edit</Button>
            <Button>
              <Send data-icon="inline-start" />
              Deploy
            </Button>
          </CardFooter>
        </Card>

        {/* Invite Team Card */}
        <Card>
          <CardHeader>
            <CardTitle>Invite Team</CardTitle>
            <CardDescription>Add members to your workspace</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <div className="flex items-center gap-2">
                <Input placeholder="alex@example.com" className="flex-1" />
                <Select defaultValue="editor">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="editor">Editor</SelectItem>
                    <SelectItem value="viewer">Viewer</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Input placeholder="sam@example.com" className="flex-1" />
                <Select defaultValue="viewer">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="editor">Editor</SelectItem>
                    <SelectItem value="viewer">Viewer</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button variant="outline" className="w-full">
                <Plus data-icon="inline-start" />
                Add another
              </Button>

              <Separator />

              <Field>
                <FieldLabel>Or share invite link</FieldLabel>
                <div className="flex items-center gap-2">
                  <Input readOnly value="https://app.co/invite/x8f2k" className="flex-1" />
                  <Button variant="outline" size="icon">
                    <Copy />
                  </Button>
                </div>
              </Field>

              <Button className="w-full">
                <Send data-icon="inline-start" />
                Send Invites
              </Button>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Report Bug Card */}
        <Card>
          <CardHeader>
            <CardTitle>Report Bug</CardTitle>
            <CardDescription>Help us fix issues faster.</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="title">Title</FieldLabel>
                <Input id="title" placeholder="Brief description of the issue" />
              </Field>
              <div className="grid grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>Severity</FieldLabel>
                  <Select defaultValue="medium">
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="low">Low</SelectItem>
                      <SelectItem value="medium">Medium</SelectItem>
                      <SelectItem value="high">High</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field>
                  <FieldLabel>Component</FieldLabel>
                  <Select defaultValue="dashboard">
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="dashboard">Dashboard</SelectItem>
                      <SelectItem value="payments">Payments</SelectItem>
                      <SelectItem value="settings">Settings</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </div>
              <Field>
                <FieldLabel htmlFor="steps">Steps to reproduce</FieldLabel>
                <Textarea id="steps" placeholder="1. Go to 2. Click on 3. Observe..." />
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Profile Card */}
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
            <CardDescription>Manage your profile information.</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="name">Name</FieldLabel>
                <Input id="name" defaultValue="shadcn" />
                <FieldDescription>
                  Your name may appear around GitHub where you contribute or are mentioned.
                </FieldDescription>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Button Variants */}
        <Card>
          <CardHeader>
            <CardTitle>Button Variants</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              <Button>Primary</Button>
              <Button variant="secondary">Secondary</Button>
              <Button variant="outline">Outline</Button>
              <Button variant="ghost">Ghost</Button>
              <Button variant="destructive">
                <Trash2 data-icon="inline-start" />
                Delete
              </Button>
              <Button variant="link">Link</Button>
            </div>
          </CardContent>
        </Card>

        {/* Badge Variants */}
        <Card>
          <CardHeader>
            <CardTitle>Badge Variants</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              <Badge>Default</Badge>
              <Badge variant="secondary">Secondary</Badge>
              <Badge variant="outline">Outline</Badge>
              <Badge variant="destructive">Destructive</Badge>
            </div>
          </CardContent>
        </Card>

        {/* Checkbox Group */}
        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
            <CardDescription>Choose what you want to be notified about.</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field orientation="horizontal">
                <Checkbox id="email-notif" defaultChecked />
                <FieldLabel htmlFor="email-notif" className="font-normal">Email notifications</FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="push-notif" />
                <FieldLabel htmlFor="push-notif" className="font-normal">Push notifications</FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="sms-notif" />
                <FieldLabel htmlFor="sms-notif" className="font-normal">SMS notifications</FieldLabel>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
