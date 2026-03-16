import { useState, useCallback } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent } from '#/components/ui/card'
import { api } from '#/lib/api'
import { useMe } from '#/hooks/use-auth'
import { KycStepSidebar } from '#/components/kyc/kyc-step-sidebar'
import { KycStepFooter } from '#/components/kyc/kyc-step-footer'
import { PersonalDetails } from '#/components/kyc/steps/personal-details'
import { BusinessDetails } from '#/components/kyc/steps/business-details'
import { OwnershipDetails } from '#/components/kyc/steps/ownership-details'
import { IntegrationDetails } from '#/components/kyc/steps/integration-details'
import { BankAccountDetails } from '#/components/kyc/steps/bank-account-details'
import { SignAgreement } from '#/components/kyc/steps/sign-agreement'
import {
  personalDetailsSchema,
  businessDetailsSchema,
  ownershipDetailsSchema,
  integrationDetailsSchema,
  bankAccountDetailsSchema,
  signAgreementSchema,
  kycFormSchema,
  KYC_STEPS,
} from '#/lib/schemas/kyc'
import type { KycFormData } from '#/lib/schemas/kyc'

const stepSchemas = {
  1: personalDetailsSchema,
  2: businessDetailsSchema,
  3: ownershipDetailsSchema,
  4: integrationDetailsSchema,
  5: bankAccountDetailsSchema,
  6: signAgreementSchema,
} as const

const stepFieldMap: Record<number, (keyof KycFormData)[]> = {
  1: ['firstName', 'lastName', 'addressLine1', 'addressLine2', 'city', 'postalCode', 'idType', 'idNumber', 'dateOfBirth', 'mobileNo'],
  2: ['businessNature', 'businessCategory', 'itemCategory', 'itemType', 'storeType', 'registeredBusinessName', 'businessDescription', 'registrationNo', 'registeredDate', 'businessEmail', 'businessPhone', 'businessAddressLine1', 'businessAddressLine2', 'businessCity', 'businessPostalCode'],
  3: ['directors'],
  4: ['integrationType'],
  5: ['currency', 'bank', 'branch', 'accountName', 'accountNumber'],
  6: ['agreedToTerms', 'signatureName', 'signatureDate'],
}

export function KycWizard() {
  const [currentStep, setCurrentStep] = useState(1)
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set())

  const form = useForm<KycFormData>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(kycFormSchema as any) as Resolver<KycFormData>,
    defaultValues: {
      firstName: '',
      lastName: '',
      addressLine1: '',
      addressLine2: '',
      city: '',
      postalCode: '',
      idType: 'nic',
      idNumber: '',
      dateOfBirth: '',
      mobileNo: '',
      businessNature: '',
      businessCategory: '',
      itemCategory: '',
      itemType: '',
      storeType: '',
      registeredBusinessName: '',
      businessDescription: '',
      registrationNo: '',
      registeredDate: '',
      businessEmail: '',
      businessPhone: '',
      businessAddressLine1: '',
      businessAddressLine2: '',
      businessCity: '',
      businessPostalCode: '',
      directors: [{ email: '', verified: false }],
      integrationType: 'api',
      currency: '',
      bank: '',
      branch: '',
      accountName: '',
      accountNumber: '',
      agreedToTerms: false,
      signatureName: '',
      signatureDate: new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' }),
    },
    mode: 'onTouched',
  })

  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data: meData } = useMe()
  const merchantId = meData?.data?.merchant?.id

  const kycMutation = useMutation({
    mutationFn: (values: KycFormData) => {
      if (!merchantId) throw new Error('Merchant not found')
      return api.put(`/v1/merchants/${merchantId}`, {
        businessName: values.registeredBusinessName || undefined,
        businessType: values.businessNature || undefined,
        registrationNo: values.registrationNo || undefined,
        contactName: `${values.firstName} ${values.lastName}`,
        contactPhone: values.mobileNo || undefined,
        addressLine1: values.addressLine1 || undefined,
        addressLine2: values.addressLine2 || undefined,
        city: values.city || values.businessCity || undefined,
        postalCode: values.postalCode || values.businessPostalCode || undefined,
        bankName: values.bank || undefined,
        bankBranch: values.branch || undefined,
        bankAccountNo: values.accountNumber || undefined,
        bankAccountName: values.accountName || undefined,
        submitKyc: true,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth', 'me'] })
      navigate('/')
    },
  })

  const submitKyc = (values: KycFormData) => kycMutation.mutate(values)

  const validateCurrentStep = useCallback(async () => {
    const schema = stepSchemas[currentStep as keyof typeof stepSchemas]
    const fields = stepFieldMap[currentStep]
    const values: Record<string, unknown> = {}
    for (const field of fields) {
      values[field] = form.getValues(field)
    }

    const result = schema.safeParse(values)
    if (!result.success) {
      await form.trigger(fields)
      return false
    }
    return true
  }, [currentStep, form])

  const handleNext = useCallback(async () => {
    const isValid = await validateCurrentStep()
    if (!isValid) return

    setCompletedSteps((prev) => new Set([...prev, currentStep]))

    if (currentStep < KYC_STEPS.length) {
      setCurrentStep((prev) => prev + 1)
    } else {
      const allValues = form.getValues()
      submitKyc(allValues)
    }
  }, [currentStep, form, validateCurrentStep])

  const handlePrevious = useCallback(() => {
    if (currentStep > 1) {
      setCurrentStep((prev) => prev - 1)
    }
  }, [currentStep])

  const handleClearStep = useCallback(() => {
    const fields = stepFieldMap[currentStep]
    for (const field of fields) {
      if (field === 'directors') {
        form.setValue('directors', [{ email: '', verified: false }])
      } else if (field === 'idType') {
        form.setValue('idType', 'nic')
      } else if (field === 'integrationType') {
        form.setValue('integrationType', 'api')
      } else if (field === 'agreedToTerms') {
        form.setValue('agreedToTerms', false)
      } else {
        form.setValue(field, '')
      }
    }
  }, [currentStep, form])

  const handleSave = useCallback(async () => {
    const isValid = await validateCurrentStep()
    if (isValid) {
      const fields = stepFieldMap[currentStep]
      const values: Record<string, unknown> = {}
      for (const field of fields) {
        values[field] = form.getValues(field)
      }
      // Step saved locally, will be submitted with full form
    }
  }, [currentStep, form, validateCurrentStep])

  const handleSkipToDashboard = useCallback(() => {
    navigate('/')
  }, [navigate])

  const handleStepClick = useCallback((step: number) => {
    setCurrentStep(step)
  }, [])

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return <PersonalDetails form={form} />
      case 2:
        return <BusinessDetails form={form} />
      case 3:
        return <OwnershipDetails form={form} />
      case 4:
        return <IntegrationDetails form={form} />
      case 5:
        return <BankAccountDetails form={form} />
      case 6:
        return <SignAgreement form={form} completedSteps={completedSteps.size} />
      default:
        return null
    }
  }

  return (
    <div className="flex flex-col lg:flex-row gap-8">
      <KycStepSidebar
        currentStep={currentStep}
        completedSteps={completedSteps}
        onStepClick={handleStepClick}
      />

      <div className="flex-1 min-w-0">
        <Card>
          <CardContent className="p-6">
            {renderStepContent()}
            <KycStepFooter
              currentStep={currentStep}
              totalSteps={KYC_STEPS.length}
              onPrevious={handlePrevious}
              onNext={handleNext}
              onClearStep={handleClearStep}
              onSave={handleSave}
              onSkipToDashboard={handleSkipToDashboard}
              isFirstStep={currentStep === 1}
              isLastStep={currentStep === KYC_STEPS.length}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

