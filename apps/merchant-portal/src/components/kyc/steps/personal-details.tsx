import type { UseFormReturn } from 'react-hook-form'
import { Controller } from 'react-hook-form'
import { User, MapPin, CreditCard, Phone, CalendarDays } from 'lucide-react'
import { CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { RadioGroup, RadioGroupItem } from '#/components/ui/radio-group'
import { Field, FieldGroup, FieldLabel, FieldError } from '#/components/ui/field'
import { DatePicker } from '#/components/ui/date-picker'
import type { KycFormData } from '#/lib/schemas/kyc'

interface PersonalDetailsProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<KycFormData, any, any>
}

export function PersonalDetails({ form }: PersonalDetailsProps) {
  const { register, formState: { errors }, watch, setValue, control } = form
  const idType = watch('idType')

  return (
    <div className="flex flex-col gap-6">
      <CardHeader className="px-0 pt-0">
        <CardTitle className="flex items-center gap-2">
          <User className="size-5" />
          Personal Details
        </CardTitle>
      </CardHeader>

      <FieldGroup>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel htmlFor="firstName">First Name</FieldLabel>
            <Input id="firstName" placeholder="Enter first name" {...register('firstName')} />
            <FieldError>{errors.firstName?.message}</FieldError>
          </Field>
          <Field>
            <FieldLabel htmlFor="lastName">Last Name</FieldLabel>
            <Input id="lastName" placeholder="Enter last name" {...register('lastName')} />
            <FieldError>{errors.lastName?.message}</FieldError>
          </Field>
        </div>
      </FieldGroup>

      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <MapPin className="size-4" />
        Address
      </div>

      <FieldGroup>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel htmlFor="addressLine1">Address Line 1</FieldLabel>
            <Input id="addressLine1" placeholder="Street address" {...register('addressLine1')} />
            <FieldError>{errors.addressLine1?.message}</FieldError>
          </Field>
          <Field>
            <FieldLabel htmlFor="addressLine2">Address Line 2</FieldLabel>
            <Input id="addressLine2" placeholder="Apartment, suite, etc. (optional)" {...register('addressLine2')} />
          </Field>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel htmlFor="city">City</FieldLabel>
            <Input id="city" placeholder="Enter city" {...register('city')} />
            <FieldError>{errors.city?.message}</FieldError>
          </Field>
          <Field>
            <FieldLabel htmlFor="postalCode">Postal Code</FieldLabel>
            <Input id="postalCode" placeholder="Enter postal code" {...register('postalCode')} />
            <FieldError>{errors.postalCode?.message}</FieldError>
          </Field>
        </div>
      </FieldGroup>

      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <CreditCard className="size-4" />
        Identification
      </div>

      <FieldGroup>
        <Field>
          <FieldLabel>ID Type</FieldLabel>
          <RadioGroup
            className="flex flex-row gap-4"
            value={idType}
            onValueChange={(val) => setValue('idType', val as 'nic' | 'passport', { shouldValidate: true })}
          >
            <div className="flex items-center gap-2">
              <RadioGroupItem value="nic" />
              <FieldLabel className="font-normal cursor-pointer">NIC</FieldLabel>
            </div>
            <div className="flex items-center gap-2">
              <RadioGroupItem value="passport" />
              <FieldLabel className="font-normal cursor-pointer">Passport</FieldLabel>
            </div>
          </RadioGroup>
          <FieldError>{errors.idType?.message}</FieldError>
        </Field>

        <Field>
          <FieldLabel htmlFor="idNumber">
            {idType === 'passport' ? 'Passport Number' : 'NIC Number'}
          </FieldLabel>
          <Input
            id="idNumber"
            placeholder={idType === 'passport' ? 'Enter passport number' : 'Enter NIC number'}
            {...register('idNumber')}
          />
          <FieldError>{errors.idNumber?.message}</FieldError>
        </Field>
      </FieldGroup>

      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <CalendarDays className="size-4" />
        Additional Information
      </div>

      <FieldGroup>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel>Date of Birth</FieldLabel>
            <Controller
              name="dateOfBirth"
              control={control}
              render={({ field }) => (
                <DatePicker
                  value={field.value ? new Date(field.value) : undefined}
                  onChange={(date) => field.onChange(date ? date.toISOString().split('T')[0] : '')}
                  placeholder="Select date of birth"
                  toYear={new Date().getFullYear() - 18}
                />
              )}
            />
            <FieldError>{errors.dateOfBirth?.message}</FieldError>
          </Field>
          <Field>
            <FieldLabel htmlFor="mobileNo">
              <span className="inline-flex items-center gap-1">
                <Phone className="size-3.5" />
                Mobile No
              </span>
            </FieldLabel>
            <Input id="mobileNo" placeholder="Enter mobile number" {...register('mobileNo')} />
            <FieldError>{errors.mobileNo?.message}</FieldError>
          </Field>
        </div>
      </FieldGroup>
    </div>
  )
}
