import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '#/components/ui/tooltip'

interface CopyButtonProps {
  value: string
}

export function CopyButton({ value }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Tooltip>
      <TooltipTrigger render={<Button variant="ghost" size="icon-xs" onClick={handleCopy} />}>
        {copied ? <Check /> : <Copy />}
      </TooltipTrigger>
      <TooltipContent>
        {copied ? 'Copied!' : 'Copy to clipboard'}
      </TooltipContent>
    </Tooltip>
  )
}
