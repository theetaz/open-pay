import { useRef } from 'react'
import type { UseFormReturn } from 'react-hook-form'
import { Controller } from 'react-hook-form'
import { Building2, Upload, Mail, MapPin } from 'lucide-react'
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
import type { KycFormData } from '#/lib/schemas/kyc'

interface BusinessDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

export function BusinessDetails({ form }: BusinessDetailsProps) {
  const { register, control, formState: { errors } } = form
  const brCopyRef = useRef<HTMLInputElement>(null)
  const formDocRef = useRef<HTMLInputElement>(null)

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
              <FieldLabel>Nature of Business</FieldLabel>
              <Controller
                name="businessNature"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select nature" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="sole_proprietorship">Sole Proprietorship</SelectItem>
                      <SelectItem value="partnership">Partnership</SelectItem>
                      <SelectItem value="private_limited">Private Limited</SelectItem>
                      <SelectItem value="public_limited">Public Limited</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.businessNature?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel>Business Category</FieldLabel>
              <Controller
                name="businessCategory"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="retail">Retail</SelectItem>
                      <SelectItem value="wholesale">Wholesale</SelectItem>
                      <SelectItem value="services">Services</SelectItem>
                      <SelectItem value="manufacturing">Manufacturing</SelectItem>
                      <SelectItem value="technology">Technology</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.businessCategory?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel>Item Category</FieldLabel>
              <Controller
                name="itemCategory"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="physical_goods">Physical Goods</SelectItem>
                      <SelectItem value="digital_goods">Digital Goods</SelectItem>
                      <SelectItem value="services">Services</SelectItem>
                      <SelectItem value="subscriptions">Subscriptions</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.itemCategory?.message}</FieldError>
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel>Item Type</FieldLabel>
              <Controller
                name="itemType"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="electronics">Electronics</SelectItem>
                      <SelectItem value="clothing">Clothing</SelectItem>
                      <SelectItem value="food_beverage">Food & Beverage</SelectItem>
                      <SelectItem value="software">Software</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.itemType?.message}</FieldError>
            </Field>

            <Field>
              <FieldLabel>Store Type</FieldLabel>
              <Controller
                name="storeType"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select store type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="online">Online</SelectItem>
                      <SelectItem value="physical">Physical</SelectItem>
                      <SelectItem value="both">Both</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError>{errors.storeType?.message}</FieldError>
            </Field>
          </div>

          <Field>
            <FieldLabel htmlFor="registeredBusinessName">Registered Business Name</FieldLabel>
            <Input
              id="registeredBusinessName"
              placeholder="Enter registered business name"
              {...register('registeredBusinessName')}
            />
            <FieldError>{errors.registeredBusinessName?.message}</FieldError>
          </Field>

          <Field>
            <FieldLabel htmlFor="businessDescription">Business Description</FieldLabel>
            <Textarea
              id="businessDescription"
              placeholder="Describe your business activities"
              {...register('businessDescription')}
            />
            <FieldError>{errors.businessDescription?.message}</FieldError>
          </Field>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field>
              <FieldLabel htmlFor="registrationNo">Registration No</FieldLabel>
              <Input
                id="registrationNo"
                placeholder="Enter registration number"
                {...register('registrationNo')}
              />
              <FieldError>{errors.registrationNo?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="registeredDate">Registered Date</FieldLabel>
              <Input
                id="registeredDate"
                type="date"
                {...register('registeredDate')}
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
              <FieldLabel>Upload BR Copy</FieldLabel>
              <div
                className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => brCopyRef.current?.click()}
              >
                <Upload className="size-8 mx-auto text-muted-foreground mb-2" />
                <p className="text-sm text-muted-foreground mb-2">
                  Drag and drop or click to upload
                </p>
                <Button type="button" variant="outline" size="sm">
                  Select File
                </Button>
                <input
                  ref={brCopyRef}
                  type="file"
                  className="hidden"
                  accept=".pdf,.jpg,.jpeg,.png"
                />
              </div>
            </Field>

            <Field>
              <FieldLabel>Upload Form 01/20/40</FieldLabel>
              <div
                className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => formDocRef.current?.click()}
              >
                <Upload className="size-8 mx-auto text-muted-foreground mb-2" />
                <p className="text-sm text-muted-foreground mb-2">
                  Drag and drop or click to upload
                </p>
                <Button type="button" variant="outline" size="sm">
                  Select File
                </Button>
                <input
                  ref={formDocRef}
                  type="file"
                  className="hidden"
                  accept=".pdf,.jpg,.jpeg,.png"
                />
              </div>
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
              <FieldLabel htmlFor="businessEmail">Business Email</FieldLabel>
              <Input
                id="businessEmail"
                type="email"
                placeholder="business@example.com"
                {...register('businessEmail')}
              />
              <FieldError>{errors.businessEmail?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="businessPhone">Business Phone</FieldLabel>
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
              <FieldLabel htmlFor="businessAddressLine1">Address Line 1</FieldLabel>
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
              <FieldLabel htmlFor="businessCity">City</FieldLabel>
              <Input
                id="businessCity"
                placeholder="Enter city"
                {...register('businessCity')}
              />
              <FieldError>{errors.businessCity?.message}</FieldError>
            </Field>
            <Field>
              <FieldLabel htmlFor="businessPostalCode">Postal Code</FieldLabel>
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
