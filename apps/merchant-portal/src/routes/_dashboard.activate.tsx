import { createFileRoute } from '@tanstack/react-router'
import { KycWizard } from '#/components/kyc/kyc-wizard'

export const Route = createFileRoute('/_dashboard/activate')({
  component: ActivatePage,
})

function ActivatePage() {
  return <KycWizard />
}
