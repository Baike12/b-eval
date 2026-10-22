import { translateNow } from '../i18n'
import type {
  Agent,
  MetricsResponse,
  Report,
  Run,
  Task,
  TaskSet,
} from './types'

// ApiError carries the server's message so a screen can show what actually
// went wrong instead of "request failed".
export class ApiError extends Error {
  readonly status: number
  readonly payload: unknown

  constructor(status: number, message: string, payload: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.payload = payload
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  })
  const text = await response.text()
  const payload = text ? safeParse(text) : null
  if (!response.ok) {
    // The server's own message is shown as it is: it names the object the
    // request was about, and the platform speaks one language to its
    // clients. Only the fallback, which the server did not write, is a
    // message of the interface.
    const message =
      payload && typeof payload === 'object' && 'error' in payload
        ? String((payload as { error: unknown }).error)
        : translateNow('common.requestFailed', { status: response.status })
    throw new ApiError(response.status, message, payload)
  }
  return payload as T
}

function safeParse(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

// The API is mounted under the same base path as the bundle. An absolute
// /api/ path would collide with another application when the platform is
// served from a prefix such as /eval/ on a shared host.
const API = `${import.meta.env.BASE_URL}api`

export const api = {
  health: () => request<{ status: string; task_sets: number; time: string }>(`${API}/healthz`),

  listTaskSets: () => request<TaskSet[]>(`${API}/tasksets`),
  getTaskSet: (id: string, version: string) =>
    request<TaskSet>(`${API}/tasksets/${encodeURIComponent(id)}/${encodeURIComponent(version)}`),
  createTaskSet: (payload: unknown) =>
    request<TaskSet>(`${API}/tasksets`, { method: 'POST', body: JSON.stringify(payload) }),
  addTask: (id: string, version: string, task: Task) =>
    request<Task>(`${API}/tasksets/${encodeURIComponent(id)}/${encodeURIComponent(version)}/tasks`, {
      method: 'POST',
      body: JSON.stringify(task),
    }),

  listAgents: () => request<Agent[]>(`${API}/agents`),
  createAgent: (payload: unknown) =>
    request<Agent>(`${API}/agents`, { method: 'POST', body: JSON.stringify(payload) }),
  selfCheck: (id: string) =>
    request<Agent['self_check']>(`${API}/agents/${encodeURIComponent(id)}/selfcheck`, {
      method: 'POST',
      body: '{}',
    }),

  listMetrics: () => request<MetricsResponse>(`${API}/metrics`),
  registerMetric: (payload: unknown) =>
    request<unknown>(`${API}/metrics`, { method: 'POST', body: JSON.stringify(payload) }),

  listRuns: () => request<Run[]>(`${API}/runs`),
  createRun: (payload: unknown) =>
    request<Run>(`${API}/runs`, { method: 'POST', body: JSON.stringify(payload) }),
  getRun: (id: string) => request<Run>(`${API}/runs/${encodeURIComponent(id)}`),
  getReport: (id: string) => request<Report>(`${API}/runs/${encodeURIComponent(id)}/report`),
  compare: (a: string, b: string) =>
    request<{ a: Report; b: Report; difference: Record<string, number> }>(
      `${API}/compare?a=${encodeURIComponent(a)}&b=${encodeURIComponent(b)}`,
    ),
}
