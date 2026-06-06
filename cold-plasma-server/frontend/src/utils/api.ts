import { useAuthStore } from '../store/auth'

const API_BASE = (import.meta.env.VITE_API_URL as string | undefined) ?? '/api/v1'

type ApiOk<T> = { data: T }
type ApiErr = { error: { message: string } }

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: 'GET' })
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
}

export async function apiDelete<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: 'DELETE' })
}

export async function apiUpload<T>(path: string, form: FormData): Promise<T> {
  return apiRequest<T>(path, { method: 'POST', body: form })
}

async function apiRequest<T>(path: string, init: RequestInit): Promise<T> {
  const token = useAuthStore.getState().token
  const headers = new Headers(init.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  init.headers = headers

  const res = await fetch(API_BASE + path, init)
  const json = (await res.json().catch(() => ({}))) as Partial<ApiOk<T> & ApiErr>

  if (!res.ok) {
    const msg = json.error?.message || 'Ошибка запроса'
    throw new Error(msg)
  }
  return (json as ApiOk<T>).data
}
