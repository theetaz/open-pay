/**
 * Currency symbol and decimal configuration.
 */
const currencyConfig: Record<string, { symbol: string; prefix: boolean; decimals: number }> = {
  LKR: { symbol: 'Rs.', prefix: true, decimals: 2 },
  USDT: { symbol: 'USDT', prefix: false, decimals: 6 },
  USDC: { symbol: 'USDC', prefix: false, decimals: 6 },
  BTC: { symbol: 'BTC', prefix: false, decimals: 8 },
  ETH: { symbol: 'ETH', prefix: false, decimals: 8 },
  BNB: { symbol: 'BNB', prefix: false, decimals: 6 },
}

/**
 * Formats an amount with currency symbol.
 */
export function formatAmount(amount: string | number, currency: string): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  if (isNaN(num)) return `0.00 ${currency}`

  const config = currencyConfig[currency]
  if (!config) {
    return `${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 })} ${currency}`
  }

  const formatted = num.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: config.decimals,
  })

  if (config.prefix) {
    return `${config.symbol} ${formatted}`
  }
  return `${formatted} ${config.symbol}`
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

/**
 * Returns all supported cryptocurrency codes (excluding fiat).
 */
export function getSupportedCryptoCurrencies(): string[] {
  return ['USDT', 'USDC', 'BTC', 'ETH', 'BNB']
}

/**
 * Returns all supported currencies including fiat.
 */
export function getAllSupportedCurrencies(): string[] {
  return ['USDT', 'USDC', 'BTC', 'ETH', 'BNB', 'LKR']
}
