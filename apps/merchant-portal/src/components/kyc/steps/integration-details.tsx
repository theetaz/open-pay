import type { UseFormReturn } from "react-hook-form"
import { Controller } from "react-hook-form"
import { Globe, Link2 } from "lucide-react"

import { RadioGroup, RadioGroupItem } from "#/components/ui/radio-group"
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from "#/components/ui/field"
import type { KycFormData } from "#/lib/schemas/kyc"

interface IntegrationDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

export function IntegrationDetails({ form }: IntegrationDetailsProps) {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold">Integration Details</h2>
      </div>

      <FieldGroup>
        <Field>
          <FieldLabel className="text-sm font-medium" required>
            Do you need to technically integrate the Open Pay Payment Gateway?
          </FieldLabel>

          <FieldSeparator />

          <Controller
            control={form.control}
            name="integrationType"
            render={({ field }) => (
              <RadioGroup
                value={field.value}
                onValueChange={field.onChange}
                className="grid gap-3"
              >
                <div className="flex items-center gap-3 rounded-lg border border-border p-4 hover:bg-muted/50 transition-colors">
                  <RadioGroupItem value="api" />
                  <FieldLabel className="flex items-center gap-2 cursor-pointer font-normal">
                    <Globe className="size-4 text-muted-foreground" />
                    Yes, we need to integrate it to our Website / Mobile App
                  </FieldLabel>
                </div>

                <div className="flex items-center gap-3 rounded-lg border border-border p-4 hover:bg-muted/50 transition-colors">
                  <RadioGroupItem value="payment_links" />
                  <FieldLabel className="flex items-center gap-2 cursor-pointer font-normal">
                    <Link2 className="size-4 text-muted-foreground" />
                    No, we only need to use Payment Links &amp; Invoices to collect payments
                  </FieldLabel>
                </div>
              </RadioGroup>
            )}
          />

          <FieldError>{form.formState.errors.integrationType?.message}</FieldError>
        </Field>
      </FieldGroup>
    </div>
  )
}
