/**
 * Formats an amount with currency symbol.
 * LKR amounts use "Rs." prefix, USDT uses "T" prefix.
 */
export function formatAmount(amount: string | number, currency: string): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  if (isNaN(num)) return `0.00 ${currency}`

  if (currency === 'LKR') {
    return `Rs. ${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }
  return `${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 })} USDT`
}

/**
 * Displays amount in the merchant's preferred currency with the other in brackets.
 * E.g., if primary is LKR: "Rs. 1,500.00 (4.82 USDT)"
 * If primary is USDT: "4.82 USDT (Rs. 1,500.00)"
 */
export function formatDualAmount(
  amountUsdt: string | number,
  originalAmount?: string | number | null,
  originalCurrency?: string | null,
  primaryCurrency?: string | null,
  exchangeRate?: string | number | null,
): { primary: string; secondary: string } {
  const pc = primaryCurrency || 'LKR'
  const usdt = typeof amountUsdt === 'string' ? parseFloat(amountUsdt) : amountUsdt
  const rate = exchangeRate ? (typeof exchangeRate === 'string' ? parseFloat(exchangeRate) : exchangeRate) : null

  // Calculate LKR amount from USDT if not provided
  let lkr: number | null = null
  if (originalCurrency === 'LKR' && originalAmount) {
    lkr = typeof originalAmount === 'string' ? parseFloat(originalAmount) : originalAmount
  } else if (rate && rate > 0) {
    lkr = usdt * rate
  }

  if (pc === 'LKR' && lkr !== null) {
    return {
      primary: formatAmount(lkr, 'LKR'),
      secondary: formatAmount(usdt, 'USDT'),
    }
  }

  return {
    primary: formatAmount(usdt, 'USDT'),
    secondary: lkr !== null ? formatAmount(lkr, 'LKR') : '',
  }
}
