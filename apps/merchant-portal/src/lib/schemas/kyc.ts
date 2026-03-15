import { z } from 'zod'

export const personalDetailsSchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  addressLine1: z.string().min(1, 'Address is required'),
  addressLine2: z.string().optional(),
  city: z.string().min(1, 'City is required'),
  postalCode: z.string().min(1, 'Postal code is required'),
  idType: z.enum(['nic', 'passport']),
  idNumber: z.string().min(1, 'ID number is required'),
  dateOfBirth: z.string().min(1, 'Date of birth is required'),
  mobileNo: z.string().min(1, 'Mobile number is required'),
})

export const businessDetailsSchema = z.object({
  businessNature: z.string().min(1, 'Business nature is required'),
  businessCategory: z.string().min(1, 'Business category is required'),
  itemCategory: z.string().min(1, 'Item category is required'),
  itemType: z.string().min(1, 'Item type is required'),
  storeType: z.string().min(1, 'Store type is required'),
  registeredBusinessName: z.string().min(1, 'Business name is required'),
  businessDescription: z.string().min(1, 'Business description is required'),
  registrationNo: z.string().min(1, 'Registration number is required'),
  registeredDate: z.string().min(1, 'Registration date is required'),
  businessEmail: z.string().email('Valid email is required'),
  businessPhone: z.string().min(1, 'Business phone is required'),
  businessAddressLine1: z.string().min(1, 'Business address is required'),
  businessAddressLine2: z.string().optional(),
  businessCity: z.string().min(1, 'City is required'),
  businessPostalCode: z.string().min(1, 'Postal code is required'),
})

export const ownershipDetailsSchema = z.object({
  directors: z.array(z.object({
    email: z.string().email('Valid email is required'),
    verified: z.boolean().default(false),
  })).min(1, 'At least one director is required'),
})

export const integrationDetailsSchema = z.object({
  integrationType: z.enum(['api', 'payment_links']),
})

export const bankAccountDetailsSchema = z.object({
  currency: z.string().min(1, 'Currency is required'),
  bank: z.string().min(1, 'Bank is required'),
  branch: z.string().min(1, 'Branch is required'),
  accountName: z.string().min(1, 'Account name is required'),
  accountNumber: z.string().min(1, 'Account number is required'),
})

export const signAgreementSchema = z.object({
  agreedToTerms: z.boolean().refine(val => val === true, 'You must agree to the terms'),
  signatureName: z.string().min(1, 'Signature is required'),
  signatureDate: z.string().min(1, 'Date is required'),
})

export type PersonalDetails = z.infer<typeof personalDetailsSchema>
export type BusinessDetails = z.infer<typeof businessDetailsSchema>
export type OwnershipDetails = z.infer<typeof ownershipDetailsSchema>
export type IntegrationDetails = z.infer<typeof integrationDetailsSchema>
export type BankAccountDetails = z.infer<typeof bankAccountDetailsSchema>
export type SignAgreement = z.infer<typeof signAgreementSchema>

export const kycFormSchema = personalDetailsSchema
  .merge(businessDetailsSchema)
  .merge(ownershipDetailsSchema)
  .merge(integrationDetailsSchema)
  .merge(bankAccountDetailsSchema)
  .merge(signAgreementSchema)

export type KycFormData = z.infer<typeof kycFormSchema>

export const KYC_STEPS = [
  { id: 1, title: 'Personal Details', schema: personalDetailsSchema },
  { id: 2, title: 'Business Details', schema: businessDetailsSchema },
  { id: 3, title: 'Ownership Details', schema: ownershipDetailsSchema },
  { id: 4, title: 'Integration Details', schema: integrationDetailsSchema },
  { id: 5, title: 'Bank Account Details', schema: bankAccountDetailsSchema },
  { id: 6, title: 'Sign Agreement', schema: signAgreementSchema },
] as const
