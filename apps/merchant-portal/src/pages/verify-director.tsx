import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Loader2, CheckCircle, XCircle, AlertTriangle, ShieldCheck } from 'lucide-react'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

interface DirectorVerificationInfo {
  status: string
  businessName: string
  email: string
}

type PageState =
  | { kind: 'loading' }
  | { kind: 'invalid' }
  | { kind: 'expired' }
  | { kind: 'already_verified' }
  | { kind: 'pending'; info: DirectorVerificationInfo }
  | { kind: 'success' }
  | { kind: 'error'; message: string }

export function VerifyDirectorPage() {
  const { token } = useParams<{ token: string }>()
  const [state, setState] = useState<PageState>({ kind: 'loading' })

  const [fullName, setFullName] = useState('')
  const [dateOfBirth, setDateOfBirth] = useState('')
  const [nicPassport, setNicPassport] = useState('')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [document, setDocument] = useState<File | null>(null)
  const [consent, setConsent] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState('')

  useEffect(() => {
    if (!token) {
      setState({ kind: 'invalid' })
      return
    }

    fetch(`${API_BASE}/v1/public/directors/verify/${token}`)
      .then(async (res) => {
        if (res.status === 404) {
          setState({ kind: 'invalid' })
          return
        }
        if (res.status === 410) {
          setState({ kind: 'expired' })
          return
        }
        if (!res.ok) {
          const body = await res.json().catch(() => ({}))
          setState({ kind: 'error', message: body.message || 'An unexpected error occurred.' })
          return
        }
        const body = await res.json()
        const info: DirectorVerificationInfo = {
          status: body.data?.status ?? body.status,
          businessName: body.data?.businessName ?? body.businessName ?? '',
          email: body.data?.email ?? body.email ?? '',
        }
        if (info.status === 'VERIFIED') {
          setState({ kind: 'already_verified' })
        } else {
          setState({ kind: 'pending', info })
        }
      })
      .catch(() => {
        setState({ kind: 'error', message: 'Failed to load verification details. Please try again.' })
      })
  }, [token])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!consent) return

    setSubmitting(true)
    setSubmitError('')

    const formData = new FormData()
    formData.append('fullName', fullName)
    formData.append('dateOfBirth', dateOfBirth)
    formData.append('nicPassport', nicPassport)
    formData.append('phone', phone)
    formData.append('address', address)
    if (document) {
      formData.append('document', document)
    }

    try {
      const res = await fetch(`${API_BASE}/v1/public/directors/verify/${token}`, {
        method: 'POST',
        body: formData,
      })

      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setSubmitError(body.message || 'Submission failed. Please try again.')
        setSubmitting(false)
        return
      }

      setState({ kind: 'success' })
    } catch {
      setSubmitError('Network error. Please check your connection and try again.')
      setSubmitting(false)
    }
  }

  // ── Loading ──────────────────────────────────────────────────────────────
  if (state.kind === 'loading') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto" />
          <p className="text-sm text-muted-foreground mt-4">Loading...</p>
        </div>
      </div>
    )
  }

  // ── Invalid link ─────────────────────────────────────────────────────────
  if (state.kind === 'invalid') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="text-center">
          <XCircle className="h-12 w-12 text-destructive mx-auto" />
          <p className="text-lg font-medium mt-4">Invalid Link</p>
          <p className="text-sm text-muted-foreground mt-1">
            This verification link is invalid or does not exist.
          </p>
        </div>
      </div>
    )
  }

  // ── Expired ──────────────────────────────────────────────────────────────
  if (state.kind === 'expired') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="text-center">
          <AlertTriangle className="h-12 w-12 text-yellow-500 mx-auto" />
          <p className="text-lg font-medium mt-4">Verification Link Expired</p>
          <p className="text-sm text-muted-foreground mt-1">
            This verification link has expired. Please contact the merchant to request a new one.
          </p>
        </div>
      </div>
    )
  }

  // ── Already verified ─────────────────────────────────────────────────────
  if (state.kind === 'already_verified') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="text-center">
          <CheckCircle className="h-12 w-12 text-green-500 mx-auto" />
          <p className="text-lg font-medium mt-4">Identity Verified</p>
          <p className="text-sm text-muted-foreground mt-1">
            Your identity has already been verified. No further action is required.
          </p>
        </div>
      </div>
    )
  }

  // ── Success after submission ──────────────────────────────────────────────
  if (state.kind === 'success') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="w-full max-w-lg">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-primary">Open Pay</h1>
            <p className="text-sm text-muted-foreground mt-1">Director Verification</p>
          </div>
          <div className="rounded-lg border border-border bg-card p-6 shadow-sm text-center">
            <CheckCircle className="h-12 w-12 text-green-500 mx-auto" />
            <h2 className="text-xl font-semibold mt-4">Verification Submitted</h2>
            <p className="text-sm text-muted-foreground mt-2">
              Thank you. Your identity details have been submitted successfully and are under review.
              You will be notified once the verification is complete.
            </p>
          </div>
          <p className="text-center text-xs text-muted-foreground mt-4">Powered by Open Pay</p>
        </div>
      </div>
    )
  }

  // ── Generic error ────────────────────────────────────────────────────────
  if (state.kind === 'error') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="text-center">
          <XCircle className="h-12 w-12 text-destructive mx-auto" />
          <p className="text-lg font-medium mt-4">Something went wrong</p>
          <p className="text-sm text-muted-foreground mt-1">{state.message}</p>
        </div>
      </div>
    )
  }

  // ── Pending — show the verification form ─────────────────────────────────
  const { info } = state

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        {/* Branding header */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-primary">Open Pay</h1>
          <p className="text-sm text-muted-foreground mt-1">Director Verification</p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          {/* Page title */}
          <div className="flex items-center gap-2 mb-5">
            <ShieldCheck className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">Identity Verification</h2>
          </div>

          {/* Read-only context box */}
          <div className="rounded-md bg-muted/50 border border-border px-4 py-3 mb-6 space-y-1">
            <p className="text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Business:</span> {info.businessName}
            </p>
            <p className="text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Email:</span> {info.email}
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Full Name */}
            <div>
              <label className="text-sm font-medium block mb-1">
                Full Name <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="As shown on your national ID or passport"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {/* Date of Birth */}
            <div>
              <label className="text-sm font-medium block mb-1">
                Date of Birth <span className="text-destructive">*</span>
              </label>
              <input
                type="date"
                required
                value={dateOfBirth}
                onChange={(e) => setDateOfBirth(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {/* NIC / Passport */}
            <div>
              <label className="text-sm font-medium block mb-1">
                NIC / Passport Number <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="e.g. 199012345678 or A12345678"
                value={nicPassport}
                onChange={(e) => setNicPassport(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {/* Phone Number */}
            <div>
              <label className="text-sm font-medium block mb-1">
                Phone Number <span className="text-destructive">*</span>
              </label>
              <input
                type="tel"
                required
                placeholder="+94 77 123 4567"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {/* Address */}
            <div>
              <label className="text-sm font-medium block mb-1">
                Address <span className="text-destructive">*</span>
              </label>
              <textarea
                required
                rows={3}
                placeholder="Your current residential address"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none"
              />
            </div>

            {/* Document Upload */}
            <div>
              <label className="text-sm font-medium block mb-1">
                Identity Document <span className="text-destructive">*</span>
              </label>
              <p className="text-xs text-muted-foreground mb-2">
                Upload a clear copy of your NIC (front &amp; back) or passport photo page.
              </p>
              <input
                type="file"
                required
                accept=".pdf,.png,.jpg,.jpeg"
                onChange={(e) => setDocument(e.target.files?.[0] ?? null)}
                className="w-full text-sm text-muted-foreground file:mr-3 file:rounded-md file:border file:border-input file:bg-background file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-foreground hover:file:bg-accent"
              />
            </div>

            {/* Consent */}
            <div className="flex items-start gap-3 pt-1">
              <input
                type="checkbox"
                id="consent"
                required
                checked={consent}
                onChange={(e) => setConsent(e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-input accent-primary"
              />
              <label htmlFor="consent" className="text-sm text-muted-foreground leading-snug">
                I confirm that the information provided is accurate and I consent to Open Pay
                processing my personal data for identity verification purposes.
              </label>
            </div>

            {/* Inline error */}
            {submitError && (
              <div className="rounded-md bg-destructive/10 border border-destructive/20 p-3">
                <p className="text-sm text-destructive">{submitError}</p>
              </div>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={submitting || !consent}
              className="w-full rounded-md bg-primary text-primary-foreground py-3 text-sm font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {submitting ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <ShieldCheck className="size-4" />
              )}
              {submitting ? 'Submitting...' : 'Submit Verification'}
            </button>
          </form>
        </div>

        <p className="text-center text-xs text-muted-foreground mt-4">Powered by Open Pay</p>
      </div>
    </div>
  )
}
