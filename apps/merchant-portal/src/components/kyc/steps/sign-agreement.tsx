import type { UseFormReturn } from "react-hook-form"
import { Controller } from "react-hook-form"
import { FileText, Shield, Pen, Loader2, Download } from "lucide-react"
import { useQuery } from "@tanstack/react-query"

import { Alert, AlertDescription } from "#/components/ui/alert"
import { Checkbox } from "#/components/ui/checkbox"
import { Input } from "#/components/ui/input"
import { ScrollArea } from "#/components/ui/scroll-area"
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from "#/components/ui/field"
import { api } from "#/lib/api"
import type { KycFormData } from "#/lib/schemas/kyc"

interface SignAgreementProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
  completedSteps: number
}

function useTermsAndConditions() {
  return useQuery({
    queryKey: ['legal-documents', 'terms_and_conditions'],
    queryFn: () =>
      api.get<{
        data: { id: string; type: string; version: number; title: string; content: string; pdfObjectKey?: string }
      }>('/v1/legal-documents/active?type=terms_and_conditions'),
    staleTime: 5 * 60 * 1000,
  })
}

export function SignAgreement({ form, completedSteps }: SignAgreementProps) {
  const { data: termsData, isLoading: termsLoading } = useTermsAndConditions()
  const terms = termsData?.data

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold">Sign Agreement</h2>
      </div>

      <Alert className="bg-muted/50">
        <FileText className="size-4" />
        <AlertDescription className="flex flex-col gap-2">
          <p>
            You can review all previous steps before final submission. Please
            ensure all information is accurate as changes after submission may
            require additional review.
          </p>
          <div className="flex items-center justify-between font-medium">
            <span>Total Steps Completed:</span>
            <span className="text-foreground">
              {completedSteps} of 6
            </span>
          </div>
          <p>
            By proceeding, you confirm that all information provided is accurate
            and complete.
          </p>
        </AlertDescription>
      </Alert>

      <FieldSeparator />

      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Shield className="size-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">
            {terms?.title || 'Terms and Conditions'}
            {terms?.version && (
              <span className="ml-2 text-xs font-normal text-muted-foreground">v{terms.version}</span>
            )}
          </h3>
        </div>

        <div className="rounded-lg border border-border">
          {terms?.pdfObjectKey && (
            <div className="flex items-center justify-between border-b px-4 py-2 bg-muted/30">
              <span className="text-xs text-muted-foreground flex items-center gap-1">
                <FileText className="size-3" />
                PDF version available
              </span>
              <button
                type="button"
                onClick={() => window.open(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/v1/assets/${terms.pdfObjectKey}`, '_blank')}
                className="text-xs text-primary hover:underline flex items-center gap-1"
              >
                <Download className="size-3" />
                Download PDF
              </button>
            </div>
          )}
          <ScrollArea className="h-[200px] p-4">
            {termsLoading ? (
              <div className="flex items-center justify-center h-full">
                <Loader2 className="size-6 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <p className="text-xs text-muted-foreground whitespace-pre-line leading-relaxed">
                {terms?.content || 'Terms and conditions are being loaded...'}
              </p>
            )}
          </ScrollArea>
        </div>

        <FieldGroup>
          <Field orientation="horizontal">
            <Controller
              control={form.control}
              name="agreedToTerms"
              render={({ field }) => (
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
            <FieldLabel className="text-sm font-normal leading-relaxed cursor-pointer" required>
              I agree to the Terms and Conditions and Privacy Policy
            </FieldLabel>
          </Field>

          <FieldError>{form.formState.errors.agreedToTerms?.message}</FieldError>
        </FieldGroup>
      </div>

      <FieldSeparator />

      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Pen className="size-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">Electronic Signature</h3>
        </div>

        <FieldGroup>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="signatureName" required>
                Full Name (Electronic Signature)
              </FieldLabel>
              <Input
                id="signatureName"
                placeholder="Enter your full name"
                {...form.register("signatureName")}
              />
              <FieldError>{form.formState.errors.signatureName?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel htmlFor="signatureDate">Signed Date</FieldLabel>
              <Input
                id="signatureDate"
                readOnly
                {...form.register("signatureDate")}
              />
            </Field>
          </div>
        </FieldGroup>

        <Alert className="bg-muted/50">
          <AlertDescription className="text-xs">
            By providing your electronic signature above, you acknowledge that
            you have read, understood, and agree to be bound by the terms and
            conditions of this agreement. Your electronic signature has the same
            legal effect as a handwritten signature.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  )
}
