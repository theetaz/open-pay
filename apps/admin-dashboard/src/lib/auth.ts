const isBrowser = typeof window !== 'undefined'

export function getAccessToken(): string | null {
  if (!isBrowser) return null
  return localStorage.getItem('admin_access_token')
}

export function setTokens(accessToken: string, refreshToken: string): void {
  if (!isBrowser) return
  localStorage.setItem('admin_access_token', accessToken)
  localStorage.setItem('admin_refresh_token', refreshToken)
}

export function clearTokens(): void {
  if (!isBrowser) return
  localStorage.removeItem('admin_access_token')
  localStorage.removeItem('admin_refresh_token')
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
