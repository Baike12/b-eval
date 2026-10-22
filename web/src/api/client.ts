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
    const message =
      payload && typeof payload === 'object' && 'error' in payload
        ? String((payload as { error: unknown }).error)
        : `request failed with status ${response.status}`
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

export const api = {
  health: () => request<{ status: string; task_sets: number; time: string }>('/api/healthz'),

  listTaskSets: () => request<TaskSet[]>('/api/tasksets'),
  getTaskSet: (id: string, version: string) =>
    request<TaskSet>(`/api/tasksets/${encodeURIComponent(id)}/${encodeURIComponent(version)}`),
  createTaskSet: (payload: unknown) =>
    request<TaskSet>('/api/tasksets', { method: 'POST', body: JSON.stringify(payload) }),
  addTask: (id: string, version: string, task: Task) =>
    request<Task>(`/api/tasksets/${encodeURIComponent(id)}/${encodeURIComponent(version)}/tasks`, {
      method: 'POST',
      body: JSON.stringify(task),
    }),

  listAgents: () => request<Agent[]>('/api/agents'),
  createAgent: (payload: unknown) =>
    request<Agent>('/api/agents', { method: 'POST', body: JSON.stringify(payload) }),
  selfCheck: (id: string) =>
    request<Agent['self_check']>(`/api/agents/${encodeURIComponent(id)}/selfcheck`, {
      method: 'POST',
      body: '{}',
    }),

  listMetrics: () => request<MetricsResponse>('/api/metrics'),
  registerMetric: (payload: unknown) =>
    request<unknown>('/api/metrics', { method: 'POST', body: JSON.stringify(payload) }),

  listRuns: () => request<Run[]>('/api/runs'),
  createRun: (payload: unknown) =>
    request<Run>('/api/runs', { method: 'POST', body: JSON.stringify(payload) }),
  getRun: (id: string) => request<Run>(`/api/runs/${encodeURIComponent(id)}`),
  getReport: (id: string) => request<Report>(`/api/runs/${encodeURIComponent(id)}/report`),
  compare: (a: string, b: string) =>
    request<{ a: Report; b: Report; difference: Record<string, number> }>(
      `/api/compare?a=${encodeURIComponent(a)}&b=${encodeURIComponent(b)}`,
    ),
}
