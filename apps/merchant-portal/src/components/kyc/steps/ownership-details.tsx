import { useState } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { useFieldArray } from 'react-hook-form'
import { Plus, Trash2, Mail, ShieldCheck, Loader2 } from 'lucide-react'
import { CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Alert, AlertTitle, AlertDescription } from '#/components/ui/alert'
import { Badge } from '#/components/ui/badge'
import { Field, FieldGroup, FieldLabel, FieldError } from '#/components/ui/field'
import type { KycFormData } from '#/lib/schemas/kyc'
import { useMe } from '#/hooks/use-auth'
import { api } from '#/lib/api'
import { toast } from 'sonner'

interface OwnershipDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

export function OwnershipDetails({ form }: OwnershipDetailsProps) {
  const { register, control, formState: { errors }, watch } = form
  const { fields, append, remove } = useFieldArray({
    control,
    name: 'directors',
  })
  const { data: meData } = useMe()
  const merchantId = meData?.data?.merchant?.id
  const [sendingIndex, setSendingIndex] = useState<number | null>(null)
  const [sentEmails, setSentEmails] = useState<Set<string>>(new Set())

  const handleSendVerification = async (index: number) => {
    const email = watch(`directors.${index}.email`)
    if (!email || !merchantId) return

    setSendingIndex(index)
    try {
      await api.post(`/v1/merchants/${merchantId}/directors/verify`, { email })
      setSentEmails((prev) => new Set([...prev, email]))
      toast.success(`Verification email sent to ${email}`)
    } catch {
      toast.error('Failed to send verification email')
    } finally {
      setSendingIndex(null)
    }
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
              onClick={() => append({ email: '', verified: false })}
            >
              <Plus data-icon="inline-start" />
              Add another
            </Button>
          </div>
        </div>

        <FieldError>{errors.directors?.root?.message}</FieldError>

        <FieldGroup>
          <div className="flex flex-col gap-3">
            {fields.map((field, index) => {
              const email = watch(`directors.${index}.email`)
              const wasSent = sentEmails.has(email)

              return (
                <div key={field.id} className="flex items-start gap-3">
                  <Field className="flex-1">
                    <FieldLabel htmlFor={`directors.${index}.email`} required>
                      Director {index + 1} Email
                    </FieldLabel>
                    <div className="flex items-center gap-2">
                      <Input
                        id={`directors.${index}.email`}
                        type="email"
                        placeholder="director@example.com"
                        {...register(`directors.${index}.email`)}
                      />
                      {field.verified && (
                        <Badge variant="secondary" className="shrink-0 bg-green-500/10 text-green-600 dark:text-green-400">
                          Verified
                        </Badge>
                      )}
                      {wasSent && !field.verified && (
                        <Badge variant="secondary" className="shrink-0 bg-amber-500/10 text-amber-600 dark:text-amber-400">
                          Sent
                        </Badge>
                      )}
                    </div>
                    <FieldError>{errors.directors?.[index]?.email?.message}</FieldError>
                  </Field>
                  <div className="flex items-end gap-2 pt-7">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleSendVerification(index)}
                      disabled={sendingIndex === index || !email || field.verified}
                    >
                      {sendingIndex === index ? (
                        <Loader2 className="size-4 mr-1 animate-spin" />
                      ) : (
                        <Mail data-icon="inline-start" />
                      )}
                      {wasSent ? 'Resend' : 'Send Verification'}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => remove(index)}
                      disabled={fields.length <= 1}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        </FieldGroup>

        {fields.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No directors added yet.</p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={() => append({ email: '', verified: false })}
            >
              <Plus data-icon="inline-start" />
              Add a director
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
