import type { UseFormReturn } from "react-hook-form"
import { Controller } from "react-hook-form"
import { FileText, Shield, Pen } from "lucide-react"

import { Alert, AlertDescription } from "#/components/ui/alert"
import { Checkbox } from "#/components/ui/checkbox"
import { Input } from "#/components/ui/input"
import { ScrollArea } from "#/components/ui/scroll-area"
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from "#/components/ui/field"
import type { KycFormData } from "#/lib/schemas/kyc"

interface SignAgreementProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
  completedSteps: number
}

const TERMS_TEXT = `OPEN PAY PAYMENT GATEWAY — TERMS AND CONDITIONS

Last Updated: March 2026

1. INTRODUCTION
These Terms and Conditions ("Agreement") govern your use of the Open Pay Payment Gateway ("Service") operated by Open Lanka Payment (Pvt) Ltd ("Company"), a registered entity under the laws of the Democratic Socialist Republic of Sri Lanka. By accessing or using our Service, you agree to be bound by these terms.

2. DEFINITIONS
"Merchant" refers to any individual or business entity that registers to use the Open Pay Payment Gateway to accept cryptocurrency payments.
"Transaction" refers to any payment processed through the Service.
"Settlement" refers to the conversion and transfer of funds to the Merchant's designated bank account.

3. ELIGIBILITY
To use the Service, you must:
(a) Be a registered business entity or sole proprietorship in Sri Lanka;
(b) Maintain a valid bank account with a licensed commercial bank in Sri Lanka;
(c) Comply with all applicable laws, including the Payment Devices Fraud Act No. 30 of 2006 and relevant Central Bank of Sri Lanka (CBSL) regulations;
(d) Complete the Know Your Customer (KYC) verification process.

4. PAYMENT PROCESSING
The Company facilitates cryptocurrency payment processing, converting digital assets to fiat currency (LKR or USD) as per the Merchant's preference. Settlement periods are typically 1–3 business days following transaction confirmation on the respective blockchain network.

5. FEES AND CHARGES
Transaction fees are calculated as a percentage of each processed payment and are deducted at the time of settlement. The current fee schedule is available on the Open Pay dashboard. The Company reserves the right to modify fees with 30 days' prior written notice.

6. COMPLIANCE AND ANTI-MONEY LAUNDERING
Merchants must comply with the Financial Transactions Reporting Act No. 6 of 2006 and the Prevention of Money Laundering Act No. 5 of 2006. The Company reserves the right to suspend or terminate accounts suspected of involvement in money laundering, terrorist financing, or other illegal activities.

7. DATA PROTECTION
The Company processes personal data in accordance with the Personal Data Protection Act No. 9 of 2022 of Sri Lanka. Merchant data is encrypted and stored securely in compliance with industry standards.

8. LIMITATION OF LIABILITY
The Company shall not be liable for any indirect, incidental, or consequential damages arising from the use of the Service, including but not limited to losses due to cryptocurrency price volatility, network congestion, or blockchain-related delays.

9. TERMINATION
Either party may terminate this Agreement with 30 days' written notice. The Company may immediately suspend or terminate access if the Merchant breaches any provision of this Agreement or applicable law.

10. GOVERNING LAW
This Agreement shall be governed by and construed in accordance with the laws of the Democratic Socialist Republic of Sri Lanka. Any disputes shall be subject to the exclusive jurisdiction of the courts of Sri Lanka.

11. AMENDMENTS
The Company reserves the right to amend these Terms and Conditions at any time. Continued use of the Service following any amendment constitutes acceptance of the modified terms.

For questions regarding these Terms and Conditions, contact: legal@openpay.lk`

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

      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Shield className="size-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">Terms and Conditions</h3>
        </div>

        <div className="rounded-lg border border-border">
          <ScrollArea className="h-[200px] p-4">
            <p className="text-xs text-muted-foreground whitespace-pre-line leading-relaxed">
              {TERMS_TEXT}
            </p>
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
            <FieldLabel className="text-sm font-normal leading-relaxed cursor-pointer">
              I agree to the Terms and Conditions and Privacy Policy
              <span className="text-destructive"> *</span>
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
              <FieldLabel htmlFor="signatureName">
                Full Name (Electronic Signature) <span className="text-destructive">*</span>
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
