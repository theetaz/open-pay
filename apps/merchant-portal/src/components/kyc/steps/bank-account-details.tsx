import { useRef, useState } from "react"
import type { UseFormReturn } from "react-hook-form"
import { Controller } from "react-hook-form"
import { Upload, X, FileText } from "lucide-react"

import { Button } from "#/components/ui/button"
import { Input } from "#/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "#/components/ui/select"
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from "#/components/ui/field"
import type { KycFormData } from "#/lib/schemas/kyc"

interface BankAccountDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

const CURRENCY_OPTIONS = [
  { value: "LKR", label: "LKR - Sri Lankan Rupee" },
  { value: "USD", label: "USD - US Dollar" },
]

const BANK_OPTIONS = [
  "Bank of Ceylon",
  "People's Bank",
  "Commercial Bank",
  "Hatton National Bank",
  "Sampath Bank",
  "Seylan Bank",
  "DFCC Bank",
  "NDB Bank",
  "Other",
]

export function BankAccountDetails({ form }: BankAccountDetailsProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [selectedFileName, setSelectedFileName] = useState<string | null>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setSelectedFileName(file.name)
    }
  }

  const handleRemoveFile = () => {
    setSelectedFileName(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ""
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold">Bank Account Details</h2>
      </div>

      <FieldGroup>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel>Currency <span className="text-destructive">*</span></FieldLabel>
            <Controller
              control={form.control}
              name="currency"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select currency" />
                  </SelectTrigger>
                  <SelectContent>
                    {CURRENCY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            <FieldError>{form.formState.errors.currency?.message}</FieldError>
          </Field>

          <Field>
            <FieldLabel>Bank <span className="text-destructive">*</span></FieldLabel>
            <Controller
              control={form.control}
              name="bank"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select bank" />
                  </SelectTrigger>
                  <SelectContent>
                    {BANK_OPTIONS.map((bank) => (
                      <SelectItem key={bank} value={bank}>
                        {bank}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            <FieldError>{form.formState.errors.bank?.message}</FieldError>
          </Field>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel htmlFor="branch">Branch <span className="text-destructive">*</span></FieldLabel>
            <Input
              id="branch"
              placeholder="Enter branch name"
              {...form.register("branch")}
            />
            <FieldError>{form.formState.errors.branch?.message}</FieldError>
          </Field>

          <Field>
            <FieldLabel htmlFor="accountName">Account Name <span className="text-destructive">*</span></FieldLabel>
            <Input
              id="accountName"
              placeholder="Enter account holder name"
              {...form.register("accountName")}
            />
            <FieldError>{form.formState.errors.accountName?.message}</FieldError>
          </Field>
        </div>

        <Field>
          <FieldLabel htmlFor="accountNumber">Account Number <span className="text-destructive">*</span></FieldLabel>
          <Input
            id="accountNumber"
            placeholder="Enter account number"
            {...form.register("accountNumber")}
          />
          <FieldError>{form.formState.errors.accountNumber?.message}</FieldError>
        </Field>
      </FieldGroup>

      <FieldSeparator />

      <div className="flex flex-col gap-4">
        <h3 className="text-sm font-semibold">Supporting Documents</h3>

        <FieldGroup>
          <Field>
            <FieldLabel>
              Bank Statement (Last 3 Months in PDF Format)
              <span className="text-destructive"> *</span>
            </FieldLabel>

            {selectedFileName ? (
              <div className="flex items-center gap-3 rounded-lg border border-border p-4">
                <FileText className="size-5 text-muted-foreground" />
                <span className="flex-1 text-sm truncate">{selectedFileName}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={handleRemoveFile}
                >
                  <X className="size-4" />
                </Button>
              </div>
            ) : (
              <div
                className="border-2 border-dashed border-border rounded-lg p-8 flex flex-col items-center justify-center gap-3 hover:border-primary/50 transition-colors cursor-pointer min-h-[200px]"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="size-8 text-muted-foreground" />
                <p className="text-sm font-medium">Drop your file here</p>
                <p className="text-xs text-muted-foreground">
                  PNG, JPG, or PDF (max. 32MB)
                </p>
                <Button type="button" variant="outline" size="sm">
                  <Upload data-icon="inline-start" />
                  Select File
                </Button>
              </div>
            )}

            <input
              ref={fileInputRef}
              type="file"
              accept=".png,.jpg,.jpeg,.pdf"
              className="hidden"
              onChange={handleFileSelect}
            />
          </Field>
        </FieldGroup>
      </div>
    </div>
  )
}
