import type { UseFormReturn } from "react-hook-form"
import { Controller } from "react-hook-form"
import { FileText, Shield, Pen, Loader2, Download, ScrollText } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import Markdown from "react-markdown"
import remarkGfm from "remark-gfm"

import { Alert, AlertDescription } from "#/components/ui/alert"
import { Checkbox } from "#/components/ui/checkbox"
import { Input } from "#/components/ui/input"
import { ScrollArea } from "#/components/ui/scroll-area"
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from "#/components/ui/field"
import { api } from "#/lib/api"
import type { KycFormData } from "#/lib/schemas/kyc"

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const mdComponents = {
  h1: ({ children, ...props }: React.ComponentProps<'h1'>) => <h1 className="text-base font-bold mt-3 mb-1" {...props}>{children}</h1>,
  h2: ({ children, ...props }: React.ComponentProps<'h2'>) => <h2 className="text-sm font-semibold mt-2 mb-1" {...props}>{children}</h2>,
  h3: ({ children, ...props }: React.ComponentProps<'h3'>) => <h3 className="text-sm font-semibold mt-2 mb-1" {...props}>{children}</h3>,
  p: ({ children, ...props }: React.ComponentProps<'p'>) => <p className="text-xs text-muted-foreground leading-relaxed mb-1.5" {...props}>{children}</p>,
  ul: ({ children, ...props }: React.ComponentProps<'ul'>) => <ul className="list-disc ml-5 mb-1.5 text-xs text-muted-foreground" {...props}>{children}</ul>,
  ol: ({ children, ...props }: React.ComponentProps<'ol'>) => <ol className="list-decimal ml-5 mb-1.5 text-xs text-muted-foreground" {...props}>{children}</ol>,
  li: ({ children, ...props }: React.ComponentProps<'li'>) => <li className="mb-0.5" {...props}>{children}</li>,
  strong: ({ children, ...props }: React.ComponentProps<'strong'>) => <strong className="font-semibold text-foreground" {...props}>{children}</strong>,
  a: ({ children, ...props }: React.ComponentProps<'a'>) => <a className="text-primary underline" {...props}>{children}</a>,
  blockquote: ({ children, ...props }: React.ComponentProps<'blockquote'>) => <blockquote className="border-l-2 border-primary/30 pl-3 italic my-1.5 text-muted-foreground" {...props}>{children}</blockquote>,
  hr: (props: React.ComponentProps<'hr'>) => <hr className="my-2 border-border" {...props} />,
}

interface SignAgreementProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
  completedSteps: number
}

interface LegalDocResponse {
  data: { id: string; type: string; version: number; title: string; content: string; pdfObjectKey?: string }
}

function useLegalDocument(type: string) {
  return useQuery({
    queryKey: ['legal-documents', type],
    queryFn: () => api.get<LegalDocResponse>(`/v1/legal-documents/active?type=${type}`),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

function DocumentSection({ type, icon: Icon, fallbackTitle }: { type: string; icon: React.ComponentType<{ className?: string }>; fallbackTitle: string }) {
  const { data, isLoading, isError } = useLegalDocument(type)
  const doc = data?.data

  if (isError || (!isLoading && !doc)) {
    return null // Don't show section if document type doesn't exist yet
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <Icon className="size-4 text-muted-foreground" />
        <h3 className="text-sm font-semibold">
          {doc?.title || fallbackTitle}
          {doc?.version && (
            <span className="ml-2 text-xs font-normal text-muted-foreground">v{doc.version}</span>
          )}
        </h3>
      </div>

      <div className="rounded-lg border border-border">
        {doc?.pdfObjectKey && (
          <div className="flex items-center justify-between border-b px-4 py-2 bg-muted/30">
            <span className="text-xs text-muted-foreground flex items-center gap-1">
              <FileText className="size-3" />
              PDF version available
            </span>
            <button
              type="button"
              onClick={() => window.open(`${API_BASE}/v1/assets/${doc.pdfObjectKey}`, '_blank')}
              className="text-xs text-primary hover:underline flex items-center gap-1"
            >
              <Download className="size-3" />
              Download PDF
            </button>
          </div>
        )}
        <ScrollArea className="h-[200px] p-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Markdown remarkPlugins={[remarkGfm]} components={mdComponents}>
              {doc?.content || `*${fallbackTitle} content is being loaded...*`}
            </Markdown>
          )}
        </ScrollArea>
      </div>
    </div>
  )
}

export function SignAgreement({ form, completedSteps }: SignAgreementProps) {
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

      {/* Terms and Conditions */}
      <DocumentSection
        type="terms_and_conditions"
        icon={Shield}
        fallbackTitle="Terms and Conditions"
      />

      {/* Sign Agreement */}
      <DocumentSection
        type="sign_agreement"
        icon={ScrollText}
        fallbackTitle="Sign Agreement"
      />

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
            I agree to the Terms and Conditions, Privacy Policy, and Sign Agreement
          </FieldLabel>
        </Field>

        <FieldError>{form.formState.errors.agreedToTerms?.message}</FieldError>
      </FieldGroup>

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
