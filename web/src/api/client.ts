import type {
  AssignCredentialRequest,
  CreateCredentialRequest,
  CreateMemberRequest,
  CredentialItem,
  MemberResponse,
  MetricsSummaryResponse,
} from './types'

const TOKEN_KEY = 'angelix_admin_token'

export const auth = {
  getToken: () => localStorage.getItem(TOKEN_KEY) ?? '',
  setToken: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  clearToken: () => localStorage.removeItem(TOKEN_KEY),
  isLoggedIn: () => Boolean(localStorage.getItem(TOKEN_KEY)),
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Token': auth.getToken(),
      ...init.headers,
    },
  })

  if (res.status === 204) return undefined as T

  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? res.statusText)
  return body as T
}

export const api = {
  members: {
    list: (): Promise<{ members: MemberResponse[] }> =>
      request('/api/v1/members'),

    create: (data: CreateMemberRequest): Promise<MemberResponse> =>
      request('/api/v1/members', { method: 'POST', body: JSON.stringify(data) }),

    delete: (id: string): Promise<void> =>
      request(`/api/v1/members/${id}`, { method: 'DELETE' }),

    metrics: (id: string): Promise<MetricsSummaryResponse> =>
      request(`/api/v1/members/${id}/metrics`),

    listCredentials: (id: string): Promise<{ credentials: CredentialItem[] }> =>
      request(`/api/v1/members/${id}/credentials`),

    assignCredential: (id: string, data: AssignCredentialRequest): Promise<void> =>
      request(`/api/v1/members/${id}/credentials`, { method: 'POST', body: JSON.stringify(data) }),

    unassignCredential: (id: string, credId: string): Promise<void> =>
      request(`/api/v1/members/${id}/credentials/${credId}`, { method: 'DELETE' }),
  },

  catalog: {
    list: (): Promise<{ credentials: CredentialItem[] }> =>
      request('/api/v1/catalog'),

    create: (data: CreateCredentialRequest): Promise<CredentialItem> =>
      request('/api/v1/catalog', { method: 'POST', body: JSON.stringify(data) }),

    delete: (id: string): Promise<void> =>
      request(`/api/v1/catalog/${id}`, { method: 'DELETE' }),
  },
}
