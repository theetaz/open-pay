import { ArrowLeft, ArrowRight, RotateCcw, SkipForward, Save, Send } from 'lucide-react'
import { Button } from '#/components/ui/button'

interface KycStepFooterProps {
  currentStep: number
  totalSteps: number
  onPrevious: () => void
  onNext: () => void
  onClearStep: () => void
  onSave: () => void
  onSkipToDashboard: () => void
  isFirstStep: boolean
  isLastStep: boolean
}

export function KycStepFooter({
  onPrevious,
  onNext,
  onClearStep,
  onSave,
  onSkipToDashboard,
  isFirstStep,
  isLastStep,
}: KycStepFooterProps) {
  return (
    <div className="flex items-center justify-between pt-6 border-t">
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onPrevious}
          disabled={isFirstStep}
        >
          <ArrowLeft data-icon="inline-start" />
          Previous
        </Button>
        <Button type="button" variant="outline" onClick={onClearStep}>
          <RotateCcw data-icon="inline-start" />
          Clear Step
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <Button type="button" variant="ghost" onClick={onSkipToDashboard}>
          <SkipForward data-icon="inline-start" />
          Skip to Dashboard
        </Button>
        <Button type="button" variant="outline" onClick={onSave}>
          <Save data-icon="inline-start" />
          Save
        </Button>
        <Button type="button" size="lg" onClick={onNext}>
          {isLastStep ? (
            <>
              Submit Application
              <Send data-icon="inline-end" />
            </>
          ) : (
            <>
              Save & Next
              <ArrowRight data-icon="inline-end" />
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
