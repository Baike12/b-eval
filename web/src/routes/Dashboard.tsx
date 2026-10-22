import { useState, type FormEvent, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// Dashboard is the landing screen.
//
// It answers one question first: what is the state of the platform right
// now. The run form is on the same screen because starting a run is the
// thing a user came here to do.
export function Dashboard(): ReactNode {
  const health = useAsync(() => api.health(), [])
  const taskSets = useAsync(() => api.listTaskSets(), [])
  const agents = useAsync(() => api.listAgents(), [])
  const runs = useAsync(() => api.listRuns(), [])

  const [taskSetID, setTaskSetID] = useState('')
  const [agentID, setAgentID] = useState('')
  const [repetitions, setRepetitions] = useState(1)
  const [aggregation, setAggregation] = useState('macro')
  const [message, setMessage] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function startRun(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setMessage(null)
    try {
      const run = await api.createRun({
        task_set_id: taskSetID,
        agent_id: agentID,
        repetitions,
        aggregation,
      })
      setMessage(`Run ${run.run_id} completed.`)
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <h1>Dashboard</h1>

      {health.error ? <ErrorNote message={health.error} /> : null}
      {health.data ? (
        <p className="muted">
          API {health.data.status} · {health.data.task_sets} task set(s) · {health.data.time}
        </p>
      ) : null}

      <section className="panel">
        <h2>Start a run</h2>
        <form className="form" onSubmit={startRun}>
          <label>
            Task set
            <select value={taskSetID} onChange={(e) => setTaskSetID(e.target.value)} required>
              <option value="">select a task set</option>
              {(taskSets.data ?? []).map((ts) => (
                <option key={`${ts.task_set_id}-${ts.version}`} value={ts.task_set_id}>
                  {ts.task_set_id} ({ts.version}) — {ts.tasks.length} tasks
                </option>
              ))}
            </select>
          </label>
          <label>
            Agent
            <select value={agentID} onChange={(e) => setAgentID(e.target.value)} required>
              <option value="">select an agent</option>
              {(agents.data ?? []).map((a) => (
                <option key={a.agent_id} value={a.agent_id}>
                  {a.agent_id}
                  {a.self_check ? ` (self-check ${a.self_check.status})` : ' (no self-check)'}
                </option>
              ))}
            </select>
          </label>
          <label>
            Repetitions
            <input
              type="number"
              min={1}
              max={20}
              value={repetitions}
              onChange={(e) => setRepetitions(Number(e.target.value))}
            />
          </label>
          <label>
            Aggregation
            <select value={aggregation} onChange={(e) => setAggregation(e.target.value)}>
              <option value="macro">macro (task-equal-weighted)</option>
              <option value="micro">micro (trial-weighted)</option>
            </select>
          </label>
          <button type="submit" disabled={submitting || !taskSetID || !agentID}>
            {submitting ? 'Running…' : 'Run'}
          </button>
        </form>
        {message ? <p className="notice">{message}</p> : null}
      </section>

      <section className="panel">
        <h2>Recent runs</h2>
        {runs.loading ? <p className="muted">Loading…</p> : null}
        {runs.error ? <ErrorNote message={runs.error} /> : null}
        {!runs.loading && !runs.error && (runs.data ?? []).length === 0 ? (
          <Empty>No run has been executed yet.</Empty>
        ) : null}
        <ul className="list">
          {(runs.data ?? []).slice(0, 8).map((run) => (
            <li key={run.run_id}>
              <Link to={`/runs/${run.run_id}`}>{run.run_id}</Link>
              <span className="muted">
                {' '}
                · {run.plan.task_set_id} · {run.plan.agent_id} · {run.plan.repetitions} repetition(s) · {run.status}
              </span>
            </li>
          ))}
        </ul>
      </section>
    </>
  )
}
