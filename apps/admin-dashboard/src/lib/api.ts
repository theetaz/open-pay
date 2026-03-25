const isBrowser = typeof window !== 'undefined'
const API_BASE_URL = (isBrowser && import.meta.env.VITE_API_URL) || 'http://localhost:8080'

export class ApiRequestError extends Error {
  code: string
  status: number
  constructor(code: string, message: string, status: number) {
    super(message)
    this.code = code
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = isBrowser ? localStorage.getItem('admin_access_token') : null
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...options, headers })
  const body = await res.json()

  if (!res.ok) {
    const error = body.error as { code: string; message: string } | undefined
    if (res.status === 401 && isBrowser) {
      localStorage.removeItem('admin_access_token')
      localStorage.removeItem('admin_refresh_token')
      window.location.href = '/login'
    }
    throw new ApiRequestError(
      error?.code || 'UNKNOWN_ERROR',
      error?.message || 'An unexpected error occurred',
      res.status,
    )
  }
  return body
}

async function uploadFile<T>(path: string, file: File, category: string): Promise<T> {
  const token = isBrowser ? localStorage.getItem('admin_access_token') : null
  const formData = new FormData()
  formData.append('file', file)
  formData.append('category', category)

  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers,
    body: formData,
  })

  const body = await res.json()
  if (!res.ok) {
    const error = body.error as { code: string; message: string } | undefined
    throw new ApiRequestError(
      error?.code || 'UPLOAD_FAILED',
      error?.message || 'Failed to upload file',
      res.status,
    )
  }
  return body
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, data?: unknown) =>
    request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
  put: <T>(path: string, data?: unknown) =>
    request<T>(path, { method: 'PUT', body: data ? JSON.stringify(data) : undefined }),
  upload: <T>(file: File, category: string) => uploadFile<T>('/v1/admin/uploads', file, category),
}
