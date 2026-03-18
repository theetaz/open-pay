import { useRef, useState } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { Controller } from 'react-hook-form'
import { Building2, Upload, Mail, MapPin, FileText, X, Loader2 } from 'lucide-react'
import { CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'
import { Button } from '#/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Field, FieldGroup, FieldLabel, FieldError, FieldSeparator } from '#/components/ui/field'
import { DatePicker } from '#/components/ui/date-picker'
import type { KycFormData } from '#/lib/schemas/kyc'
import { api } from '#/lib/api'
import { toast } from 'sonner'

const BUSINESS_NATURE_OPTIONS = [
  { value: 'sole_proprietorship', label: 'Sole Proprietorship' },
  { value: 'partnership', label: 'Partnership' },
  { value: 'private_limited', label: 'Private Limited' },
  { value: 'public_limited', label: 'Public Limited' },
]

const BUSINESS_CATEGORY_OPTIONS = [
  { value: 'retail', label: 'Retail' },
  { value: 'wholesale', label: 'Wholesale' },
  { value: 'services', label: 'Services' },
  { value: 'manufacturing', label: 'Manufacturing' },
  { value: 'technology', label: 'Technology' },
]

const ITEM_CATEGORY_OPTIONS = [
  { value: 'physical_goods', label: 'Physical Goods' },
  { value: 'digital_goods', label: 'Digital Goods' },
  { value: 'services', label: 'Services' },
  { value: 'subscriptions', label: 'Subscriptions' },
]

const ITEM_TYPE_OPTIONS = [
  { value: 'electronics', label: 'Electronics' },
  { value: 'clothing', label: 'Clothing' },
  { value: 'food_beverage', label: 'Food & Beverage' },
  { value: 'software', label: 'Software' },
  { value: 'other', label: 'Other' },
]

const STORE_TYPE_OPTIONS = [
  { value: 'online', label: 'Online' },
  { value: 'physical', label: 'Physical' },
  { value: 'both', label: 'Both' },
]

function getLabel(options: { value: string; label: string }[], value: string): string | undefined {
  return options.find((o) => o.value === value)?.label
}

interface BusinessDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

interface UploadedFile {
  name: string
  url: string
  key: string
}

export function BusinessDetails({ form }: BusinessDetailsProps) {
  const { register, control, formState: { errors } } = form
  const brCopyRef = useRef<HTMLInputElement>(null)
  const formDocRef = useRef<HTMLInputElement>(null)
  const [brCopyFile, setBrCopyFile] = useState<UploadedFile | null>(null)
  const [formDocFile, setFormDocFile] = useState<UploadedFile | null>(null)
  const [brCopyUploading, setBrCopyUploading] = useState(false)
  const [formDocUploading, setFormDocUploading] = useState(false)

  const handleFileUpload = async (
    file: File,
    category: string,
    setFile: (f: UploadedFile | null) => void,
    setUploading: (b: boolean) => void,
  ) => {
    setUploading(true)
    try {
      const result = await api.upload<{ data: { url: string; key: string; filename: string } }>(file, category)
      setFile({ name: result.data.filename, url: result.data.url, key: result.data.key })
      toast.success(`${file.name} uploaded successfully`)
    } catch {
      toast.error(`Failed to upload ${file.name}`)
    } finally {
      setUploading(false)
    }
  }

  const handleRemoveFile = (setFile: (f: UploadedFile | null) => void, ref: React.RefObject<HTMLInputElement | null>) => {
    setFile(null)
    if (ref.current) ref.current.value = ''
  }

  return (
    <div className="flex flex-col gap-6">
      <CardHeader className="px-0 pt-0">
        <CardTitle className="flex items-center gap-2">
          <Building2 className="size-5" />
          Business Details
        </CardTitle>
      </CardHeader>

      {/* Section 1: Business Information */}
      <div className="flex flex-col gap-4">
        <h3 className="text-sm font-medium text-muted-foreground">Business Information</h3>

        <FieldGroup>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Field>
              <FieldLabel required>Nature of Business</FieldLabel>
              <Controller
                name="businessNature"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select nature">{getLabel(BUSINESS_NATURE_OPTIONS, field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {BUSINESS_NATURE_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.businessNature?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel required>Business Category</FieldLabel>
              <Controller
                name="businessCategory"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select category">{getLabel(BUSINESS_CATEGORY_OPTIONS, field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {BUSINESS_CATEGORY_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.businessCategory?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel required>Item Category</FieldLabel>
              <Controller
                name="itemCategory"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item category">{getLabel(ITEM_CATEGORY_OPTIONS, field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {ITEM_CATEGORY_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.itemCategory?.message}</FieldError>
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel required>Item Type</FieldLabel>
              <Controller
                name="itemType"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item type">{getLabel(ITEM_TYPE_OPTIONS, field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {ITEM_TYPE_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.itemType?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel required>Store Type</FieldLabel>
              <Controller
                name="storeType"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select store type">{getLabel(STORE_TYPE_OPTIONS, field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {STORE_TYPE_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.storeType?.message}</FieldError>
            </Field>
          </div>

          <Field>
            <FieldLabel htmlFor="registeredBusinessName" required>Registered Business Name</FieldLabel>
            <Input
              id="registeredBusinessName"
              placeholder="Enter registered business name"
              {...register('registeredBusinessName')}
            />
            <FieldError>{errors.registeredBusinessName?.message}</FieldError>
          </Field>

          <Field>
            <FieldLabel htmlFor="businessDescription" required>Business Description</FieldLabel>
            <Textarea
              id="businessDescription"
              placeholder="Describe your business activities"
              {...register('businessDescription')}
            />
            <FieldError>{errors.businessDescription?.message}</FieldError>
          </Field>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="registrationNo" required>Registration No</FieldLabel>
              <Input
                id="registrationNo"
                placeholder="Enter registration number"
                {...register('registrationNo')}
              />
              <FieldError>{errors.registrationNo?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel required>Registered Date</FieldLabel>
              <Controller
                name="registeredDate"
                control={control}
                render={({ field }) => (
                  <DatePicker
                    value={field.value ? new Date(field.value) : undefined}
                    onChange={(date) => {
                      if (!date) return field.onChange('')
                      const y = date.getFullYear()
                      const m = String(date.getMonth() + 1).padStart(2, '0')
                      const d = String(date.getDate()).padStart(2, '0')
                      field.onChange(`${y}-${m}-${d}`)
                    }}
                    placeholder="Select date"
                  />
                )}
              />
              <FieldError>{errors.registeredDate?.message}</FieldError>
            </Field>
          </div>
        </FieldGroup>
      </div>

      <FieldSeparator />

      {/* Section 2: Required Documents */}
      <div className="flex flex-col gap-4">
        <h3 className="text-sm font-medium text-muted-foreground">Required Documents</h3>

        <FieldGroup>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel required>Upload BR Copy</FieldLabel>
              {brCopyFile ? (
                <div className="flex items-center gap-3 rounded-lg border border-border p-4">
                  <FileText className="size-5 text-muted-foreground" />
                  <span className="flex-1 text-sm truncate">{brCopyFile.name}</span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveFile(setBrCopyFile, brCopyRef)}
                  >
                    <X className="size-4" />
                  </Button>
                </div>
              ) : (
                <div
                  className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary/50 transition-colors cursor-pointer"
                  onClick={() => !brCopyUploading && brCopyRef.current?.click()}
                >
                  {brCopyUploading ? (
                    <Loader2 className="size-8 mx-auto text-primary animate-spin mb-2" />
                  ) : (
                    <Upload className="size-8 mx-auto text-muted-foreground mb-2" />
                  )}
                  <p className="text-sm text-muted-foreground mb-2">
                    {brCopyUploading ? 'Uploading...' : 'Drag and drop or click to upload'}
                  </p>
                  {!brCopyUploading && (
                    <Button type="button" variant="outline" size="sm">
                      Select File
                    </Button>
                  )}
                </div>
              )}
              <input
                ref={brCopyRef}
                type="file"
                className="hidden"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0]
                  if (file) handleFileUpload(file, 'br-copy', setBrCopyFile, setBrCopyUploading)
                }}
              />
            </Field>

            <Field>
              <FieldLabel required>Upload Form 01/20/40</FieldLabel>
              {formDocFile ? (
                <div className="flex items-center gap-3 rounded-lg border border-border p-4">
                  <FileText className="size-5 text-muted-foreground" />
                  <span className="flex-1 text-sm truncate">{formDocFile.name}</span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveFile(setFormDocFile, formDocRef)}
                  >
                    <X className="size-4" />
                  </Button>
                </div>
              ) : (
                <div
                  className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary/50 transition-colors cursor-pointer"
                  onClick={() => !formDocUploading && formDocRef.current?.click()}
                >
                  {formDocUploading ? (
                    <Loader2 className="size-8 mx-auto text-primary animate-spin mb-2" />
                  ) : (
                    <Upload className="size-8 mx-auto text-muted-foreground mb-2" />
                  )}
                  <p className="text-sm text-muted-foreground mb-2">
                    {formDocUploading ? 'Uploading...' : 'Drag and drop or click to upload'}
                  </p>
                  {!formDocUploading && (
                    <Button type="button" variant="outline" size="sm">
                      Select File
                    </Button>
                  )}
                </div>
              )}
              <input
                ref={formDocRef}
                type="file"
                className="hidden"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0]
                  if (file) handleFileUpload(file, 'form-doc', setFormDocFile, setFormDocUploading)
                }}
              />
            </Field>
          </div>
        </FieldGroup>
      </div>

      <FieldSeparator />

      {/* Section 3: Business Contact Information */}
      <div className="flex flex-col gap-4">
        <h3 className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <Mail className="size-4" />
          Business Contact Information
        </h3>

        <FieldGroup>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="businessEmail" required>Business Email</FieldLabel>
              <Input
                id="businessEmail"
                type="email"
                placeholder="business@example.com"
                {...register('businessEmail')}
              />
              <FieldError>{errors.businessEmail?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="businessPhone" required>Business Phone</FieldLabel>
              <Input
                id="businessPhone"
                placeholder="Enter business phone"
                {...register('businessPhone')}
              />
              <FieldError>{errors.businessPhone?.message}</FieldError>
            </Field>
          </div>

          <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
            <MapPin className="size-4" />
            Business Address
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="businessAddressLine1" required>Address Line 1</FieldLabel>
              <Input
                id="businessAddressLine1"
                placeholder="Street address"
                {...register('businessAddressLine1')}
              />
              <FieldError>{errors.businessAddressLine1?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="businessAddressLine2">Address Line 2</FieldLabel>
              <Input
                id="businessAddressLine2"
                placeholder="Apartment, suite, etc. (optional)"
                {...register('businessAddressLine2')}
              />
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="businessCity" required>City</FieldLabel>
              <Input
                id="businessCity"
                placeholder="Enter city"
                {...register('businessCity')}
              />
              <FieldError>{errors.businessCity?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="businessPostalCode" required>Postal Code</FieldLabel>
              <Input
                id="businessPostalCode"
                placeholder="Enter postal code"
                {...register('businessPostalCode')}
              />
              <FieldError>{errors.businessPostalCode?.message}</FieldError>
            </Field>
          </div>
        </FieldGroup>
      </div>
    </div>
  )
}
