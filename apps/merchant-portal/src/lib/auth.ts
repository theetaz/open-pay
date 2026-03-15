const isBrowser = typeof window !== 'undefined'

export function getAccessToken(): string | null {
  if (!isBrowser) return null
  return localStorage.getItem('access_token')
}

export function getRefreshToken(): string | null {
  if (!isBrowser) return null
  return localStorage.getItem('refresh_token')
}

export function setTokens(accessToken: string, refreshToken: string): void {
  if (!isBrowser) return
  localStorage.setItem('access_token', accessToken)
  localStorage.setItem('refresh_token', refreshToken)
}

export function clearTokens(): void {
  if (!isBrowser) return
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

export function isAuthenticated(): boolean {
  if (!isBrowser) return false
  const token = getAccessToken()
  if (!token) return false

  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp * 1000 > Date.now()
  } catch {
    return false
  }
}
