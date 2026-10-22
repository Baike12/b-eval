import { useState, type FormEvent, type ReactNode } from 'react'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// Agents lists the registered systems under test and their self-check.
//
// The self-check result is shown next to every agent because an agent that
// has not passed it cannot be evaluated, and the reader should not have to
// know that to understand why a run is refused.
export function Agents(): ReactNode {
  const agents = useAsync(() => api.listAgents(), [])
  const [message, setMessage] = useState<string | null>(null)
  const [agentID, setAgentID] = useState('')
  const [shape, setShape] = useState('mock')
  const [endpoint, setEndpoint] = useState('')

  async function register(event: FormEvent) {
    event.preventDefault()
    setMessage(null)
    try {
      await api.createAgent({ agent_id: agentID, name: agentID, shape, endpoint, mode: 'single' })
      setMessage(`Agent ${agentID} registered.`)
      setAgentID('')
      agents.reload()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    }
  }

  async function runSelfCheck(id: string) {
    setMessage(null)
    try {
      const result = await api.selfCheck(id)
      setMessage(
        result?.status === 'passed'
          ? `Self-check passed for ${id}.`
          : `Self-check failed for ${id}: ${(result?.invalid_reasons ?? []).join('; ')}`,
      )
      agents.reload()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <>
      <h1>Agents</h1>
      {agents.loading ? <p className="muted">Loading…</p> : null}
      {agents.error ? <ErrorNote message={agents.error} /> : null}
      {!agents.loading && (agents.data ?? []).length === 0 ? <Empty>No agent is registered yet.</Empty> : null}

      <section className="panel">
        <h2>Register an agent</h2>
        <form className="form" onSubmit={register}>
          <label>
            Agent ID
            <input value={agentID} onChange={(e) => setAgentID(e.target.value)} required />
          </label>
          <label>
            Shape
            <select value={shape} onChange={(e) => setShape(e.target.value)}>
              <option value="mock">mock (deterministic, built in)</option>
              <option value="http">http (an agent service)</option>
            </select>
          </label>
          <label>
            Endpoint
            <input
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="http://127.0.0.1:9000"
              disabled={shape !== 'http'}
            />
          </label>
          <button type="submit">Register</button>
        </form>
        {message ? <p className="notice">{message}</p> : null}
      </section>

      <section className="panel">
        <h2>Registered</h2>
        <table className="table">
          <thead>
            <tr>
              <th>Agent</th>
              <th>Shape</th>
              <th>Self-check</th>
              <th>Violations</th>
              <th>Usage reporting</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {(agents.data ?? []).map((a) => (
              <tr key={a.agent_id}>
                <td>
                  {a.agent_id}
                  <span className="muted"> {a.name}</span>
                </td>
                <td>{a.shape}</td>
                <td>
                  {a.self_check ? (
                    <span className={`badge badge--${a.self_check.status === 'passed' ? 'ok' : 'danger'}`}>
                      {a.self_check.status}
                    </span>
                  ) : (
                    <span className="badge badge--danger">not run</span>
                  )}
                </td>
                <td>{a.self_check?.contract_violations ?? '—'}</td>
                <td>{a.self_check ? (a.self_check.usage_reporting_ok ? 'ok' : 'missing') : '—'}</td>
                <td>
                  <button type="button" onClick={() => runSelfCheck(a.agent_id)}>
                    Run self-check
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  )
}
