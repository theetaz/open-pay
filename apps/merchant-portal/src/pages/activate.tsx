import { KycWizard } from '#/components/kyc/kyc-wizard'
import { KycPendingReview } from '#/components/kyc/kyc-pending-review'
import { useMe } from '#/hooks/use-auth'

export function ActivatePage() {
  const { data: meData, isLoading } = useMe()
  const kycStatus = meData?.data?.merchant?.kycStatus

  if (isLoading) return null

  if (kycStatus && kycStatus !== 'PENDING') {
    return <KycPendingReview status={kycStatus} rejectionReason={meData?.data?.merchant?.kycRejectionReason as string | undefined} />
  }

  return <KycWizard />
}
