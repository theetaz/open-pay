import { Check } from 'lucide-react'
import { cn } from '#/lib/utils'
import { KYC_STEPS } from '#/lib/schemas/kyc'

interface KycStepSidebarProps {
  currentStep: number
  completedSteps: Set<number>
  onStepClick: (step: number) => void
}

export function KycStepSidebar({ currentStep, completedSteps, onStepClick }: KycStepSidebarProps) {
  return (
    <div className="w-64 shrink-0 hidden lg:block">
      <div className="mb-8">
        <h2 className="text-lg font-semibold">Activation Form</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Submit the following details for review to start accepting payments.
        </p>
      </div>

      <nav className="relative">
        {KYC_STEPS.map((step, index) => {
          const isCompleted = completedSteps.has(step.id)
          const isCurrent = currentStep === step.id
          const isLast = index === KYC_STEPS.length - 1

          return (
            <div key={step.id} className="relative">
              <button
                type="button"
                onClick={() => onStepClick(step.id)}
                className={cn(
                  'flex items-center gap-3 w-full text-left py-3 group',
                  isCurrent && 'cursor-default',
                )}
              >
                <div
                  className={cn(
                    'relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full text-sm font-medium transition-colors',
                    isCompleted && 'bg-green-500 text-white',
                    isCurrent && 'bg-primary text-primary-foreground ring-4 ring-primary/20 animate-pulse',
                    !isCompleted && !isCurrent && 'bg-muted text-muted-foreground',
                  )}
                >
                  {isCompleted ? <Check className="size-4" /> : step.id}
                </div>
                <span
                  className={cn(
                    'text-sm font-medium transition-colors',
                    isCompleted && 'text-green-600 dark:text-green-400',
                    isCurrent && 'text-primary',
                    !isCompleted && !isCurrent && 'text-muted-foreground',
                  )}
                >
                  {step.title}
                </span>
              </button>

              {!isLast && (
                <div
                  className={cn(
                    'absolute left-[15px] top-11 h-[calc(100%-20px)] w-0.5',
                    isCompleted ? 'bg-green-500' : 'bg-border',
                  )}
                />
              )}
            </div>
          )
        })}
      </nav>
    </div>
  )
}
