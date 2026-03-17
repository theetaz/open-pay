import { Card, CardContent } from '#/components/ui/card'
import { Clock, CheckCircle2, XCircle, ShieldCheck } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Link } from 'react-router-dom'

interface KycPendingReviewProps {
  status: string
  rejectionReason?: string
}

export function KycPendingReview({ status, rejectionReason }: KycPendingReviewProps) {
  const config = {
    INSTANT_ACCESS: {
      icon: Clock,
      iconColor: 'text-amber-500',
      bgColor: 'bg-amber-50 dark:bg-amber-900/20',
      title: 'Application Under Review',
      description: 'Your KYC application has been submitted and is being reviewed by our team.',
      detail: 'You have been granted instant access with limited transaction volume while we review your application. Full access will be enabled once approved.',
      timeline: 'Expected review time: 1-3 business days',
    },
    UNDER_REVIEW: {
      icon: Clock,
      iconColor: 'text-blue-500',
      bgColor: 'bg-blue-50 dark:bg-blue-900/20',
      title: 'Application Under Review',
      description: 'Your KYC application is currently being reviewed by our compliance team.',
      detail: 'We will notify you via email once the review is complete.',
      timeline: 'Expected review time: 1-3 business days',
    },
    APPROVED: {
      icon: CheckCircle2,
      iconColor: 'text-green-500',
      bgColor: 'bg-green-50 dark:bg-green-900/20',
      title: 'KYC Approved',
      description: 'Your account has been fully verified. You have full access to all features.',
      detail: null,
      timeline: null,
    },
    REJECTED: {
      icon: XCircle,
      iconColor: 'text-red-500',
      bgColor: 'bg-red-50 dark:bg-red-900/20',
      title: 'Application Not Approved',
      description: 'Your KYC application was not approved. Please review the feedback below.',
      detail: rejectionReason || 'No additional details provided.',
      timeline: 'You can resubmit your application after addressing the issues.',
    },
  }

  const c = config[status as keyof typeof config] || config.UNDER_REVIEW
  const Icon = c.icon

  return (
    <div className="max-w-2xl mx-auto">
      <Card>
        <CardContent className="p-8">
          <div className="flex flex-col items-center text-center gap-6">
            <div className={`rounded-full p-4 ${c.bgColor}`}>
              <Icon className={`size-12 ${c.iconColor}`} />
            </div>

            <div className="space-y-2">
              <h1 className="text-2xl font-bold">{c.title}</h1>
              <p className="text-muted-foreground">{c.description}</p>
            </div>

            {c.detail && (
              <div className="w-full rounded-lg border p-4 text-sm text-left">
                {status === 'REJECTED' ? (
                  <div className="space-y-2">
                    <p className="font-medium text-destructive">Reason for rejection:</p>
                    <p className="text-muted-foreground">{c.detail}</p>
                  </div>
                ) : (
                  <p className="text-muted-foreground">{c.detail}</p>
                )}
              </div>
            )}

            {c.timeline && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Clock className="size-4" />
                <span>{c.timeline}</span>
              </div>
            )}

            <div className="flex gap-3 pt-2">
              <Link to="/">
                <Button variant="outline">
                  <ShieldCheck className="mr-2 size-4" />
                  Go to Dashboard
                </Button>
              </Link>
              {status === 'REJECTED' && (
                <Button onClick={() => window.location.reload()}>
                  Resubmit Application
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
