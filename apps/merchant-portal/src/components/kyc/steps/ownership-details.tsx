import { useState, useEffect } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Mail, ShieldCheck, Loader2 } from 'lucide-react'
import { CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Alert, AlertTitle, AlertDescription } from '#/components/ui/alert'
import { Badge } from '#/components/ui/badge'
import { Field, FieldGroup, FieldLabel, FieldError } from '#/components/ui/field'
import type { KycFormData } from '#/lib/schemas/kyc'
import { useMe } from '#/hooks/use-auth'
import { api, ApiRequestError } from '#/lib/api'
import { toast } from 'sonner'

interface Director {
  id: string
  email: string
  fullName: string
  status: 'VERIFIED' | 'PENDING'
  tokenExpired: boolean
  verifiedAt: string | null
  createdAt: string
}

interface OwnershipDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

export function OwnershipDetails({ form }: OwnershipDetailsProps) {
  const { formState: { errors } } = form
  const { data: meData } = useMe()
  const merchantId = meData?.data?.merchant?.id
  const queryClient = useQueryClient()

  const [newEmail, setNewEmail] = useState('')

  const { data: directorsData, isLoading } = useQuery({
    queryKey: ['directors', merchantId],
    queryFn: () => api.get<{ data: Director[] }>(`/v1/merchants/${merchantId}/directors`),
    enabled: !!merchantId,
  })

  const directors = directorsData?.data ?? []

  // Keep form directors field in sync with API data
  useEffect(() => {
    form.setValue(
      'directors',
      directors.map((d) => ({ email: d.email, verified: d.status === 'VERIFIED' })),
    )
  }, [directors, form])

  const addDirectorMutation = useMutation({
    mutationFn: (email: string) =>
      api.post(`/v1/merchants/${merchantId}/directors`, { email }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['directors', merchantId] })
      setNewEmail('')
      toast.success('Director added successfully')
    },
    onError: (error) => {
      if (error instanceof ApiRequestError) {
        if (error.code === 'DUPLICATE_DIRECTOR') {
          toast.error('This director has already been added')
        } else if (error.code === 'MAX_DIRECTORS') {
          toast.error('Maximum number of directors reached')
        } else {
          toast.error(error.message || 'Failed to add director')
        }
      } else {
        toast.error('Failed to add director')
      }
    },
  })

  const resendMutation = useMutation({
    mutationFn: (directorId: string) =>
      api.post(`/v1/merchants/${merchantId}/directors/${directorId}/resend`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['directors', merchantId] })
      toast.success('Verification email resent')
    },
    onError: () => {
      toast.error('Failed to resend verification email')
    },
  })

  const removeMutation = useMutation({
    mutationFn: (directorId: string) =>
      api.delete(`/v1/merchants/${merchantId}/directors/${directorId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['directors', merchantId] })
      toast.success('Director removed')
    },
    onError: () => {
      toast.error('Failed to remove director')
    },
  })

  const handleAddDirector = () => {
    if (!newEmail.trim() || !merchantId) return
    addDirectorMutation.mutate(newEmail.trim())
  }

  return (
    <div className="flex flex-col gap-6">
      <CardHeader className="px-0 pt-0">
        <CardTitle className="flex items-center gap-2">
          <ShieldCheck className="size-5" />
          Ownership Details
        </CardTitle>
      </CardHeader>

      <Alert>
        <Mail className="size-4" />
        <AlertTitle>Identity Verification Required</AlertTitle>
        <AlertDescription>
          Each director must complete identity verification. Click &apos;Send Verification&apos; to
          email a secure verification link. At least one director must be verified.
        </AlertDescription>
      </Alert>

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Provide the details of all the directors.
          </p>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                const emailInput = document.getElementById('new-director-email') as HTMLInputElement | null
                emailInput?.focus()
              }}
            >
              <Plus data-icon="inline-start" />
              Add another
            </Button>
          </div>
        </div>

        <FieldError>{errors.directors?.root?.message}</FieldError>

        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="new-director-email" required>
              Director Email
            </FieldLabel>
            <div className="flex items-center gap-2">
              <Input
                id="new-director-email"
                type="email"
                placeholder="director@example.com"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    handleAddDirector()
                  }
                }}
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddDirector}
                disabled={!newEmail.trim() || addDirectorMutation.isPending}
              >
                {addDirectorMutation.isPending ? (
                  <Loader2 className="size-4 mr-1 animate-spin" />
                ) : (
                  <Plus data-icon="inline-start" />
                )}
                Add Director
              </Button>
            </div>
          </Field>
        </FieldGroup>

        {isLoading && (
          <div className="flex items-center justify-center py-6 text-muted-foreground">
            <Loader2 className="size-4 animate-spin mr-2" />
            <span className="text-sm">Loading directors...</span>
          </div>
        )}

        {!isLoading && directors.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No directors added yet.</p>
          </div>
        )}

        {!isLoading && directors.length > 0 && (
          <FieldGroup>
            <div className="flex flex-col gap-3">
              {directors.map((director, index) => {
                const isVerified = director.status === 'VERIFIED'
                const isExpired = director.status === 'PENDING' && director.tokenExpired
                const isPendingActive = director.status === 'PENDING' && !director.tokenExpired
                const isResending = resendMutation.isPending && resendMutation.variables === director.id
                const isRemoving = removeMutation.isPending && removeMutation.variables === director.id

                return (
                  <div key={director.id} className="flex items-start gap-3">
                    <Field className="flex-1">
                      <FieldLabel htmlFor={`director-${director.id}`} required>
                        Director {index + 1} Email
                      </FieldLabel>
                      <div className="flex items-center gap-2">
                        <Input
                          id={`director-${director.id}`}
                          type="email"
                          value={director.email}
                          readOnly
                          disabled
                        />
                        {isVerified && (
                          <Badge variant="secondary" className="shrink-0 bg-green-500/10 text-green-600 dark:text-green-400">
                            Verified
                          </Badge>
                        )}
                        {isPendingActive && (
                          <Badge variant="secondary" className="shrink-0 bg-amber-500/10 text-amber-600 dark:text-amber-400">
                            Pending
                          </Badge>
                        )}
                        {isExpired && (
                          <Badge variant="secondary" className="shrink-0 bg-red-500/10 text-red-600 dark:text-red-400">
                            Expired
                          </Badge>
                        )}
                      </div>
                    </Field>
                    <div className="flex items-end gap-2 pt-7">
                      {!isVerified && (
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => resendMutation.mutate(director.id)}
                          disabled={isResending}
                        >
                          {isResending ? (
                            <Loader2 className="size-4 mr-1 animate-spin" />
                          ) : (
                            <Mail data-icon="inline-start" />
                          )}
                          Resend
                        </Button>
                      )}
                      {!isVerified && (
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeMutation.mutate(director.id)}
                          disabled={isRemoving}
                        >
                          {isRemoving ? (
                            <Loader2 className="size-4 animate-spin" />
                          ) : (
                            <Trash2 className="size-4" />
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </FieldGroup>
        )}
      </div>
    </div>
  )
}
