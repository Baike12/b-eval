import { useState, type FormEvent, type ReactNode } from 'react'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// Metrics shows the metric registry and accepts a registration.
//
// The built-in metrics are shown as locked. A registered metric is shown as
// custom, everywhere it appears, so a reader never mistakes one for the
// other.
export function Metrics(): ReactNode {
  const registry = useAsync(() => api.listMetrics(), [])
  const [message, setMessage] = useState<string | null>(null)
  const [id, setID] = useState('')
  const [name, setName] = useState('')
  const [kind, setKind] = useState('keyword_coverage')
  const [keywords, setKeywords] = useState('')

  async function register(event: FormEvent) {
    event.preventDefault()
    setMessage(null)
    const params: Record<string, unknown> = {}
    if (kind === 'keyword_coverage') {
      params.keywords = keywords.split(',').map((k) => k.trim()).filter(Boolean)
    }
    try {
      await api.registerMetric({
        id,
        name: name || id,
        kind,
        unit: 'ratio',
        numerator: 'Trials that satisfy the condition',
        denominator: 'Valid trials',
        formula: `${kind} over the valid trials`,
        ci_method: 'wilson',
        params,
      })
      setMessage(`Metric ${id} registered.`)
      setID('')
      setName('')
      registry.reload()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    }
  }

  if (registry.loading) return <p className="muted">Loading…</p>
  if (registry.error) return <ErrorNote message={registry.error} />

  const metrics = registry.data?.metrics ?? []
  const kinds = registry.data?.custom_kinds ?? []

  return (
    <>
      <h1>Metric registry</h1>

      <section className="panel">
        <h2>Registered metrics</h2>
        {metrics.length === 0 ? <Empty>No metric is registered yet.</Empty> : null}
        <table className="table">
          <thead>
            <tr>
              <th>Metric</th>
              <th>Description</th>
              <th>Source</th>
              <th>Unit</th>
              <th>Aggregation</th>
              <th>Numerator</th>
              <th>Denominator</th>
              <th>Formula</th>
            </tr>
          </thead>
          <tbody>
            {metrics.map((m) => (
              <tr key={m.id}>
                <td>
                  {m.name}
                  <span className="metric__id">{m.id}</span>
                </td>
                <td className="metric__definition">{m.description}</td>
                <td>
                  {m.builtin ? (
                    <span className="badge badge--builtin" title="Built-in metrics are locked.">
                      built-in v{m.version}
                    </span>
                  ) : (
                    <span className="badge badge--custom" title="A registered metric.">
                      custom v{m.version}
                    </span>
                  )}
                </td>
                <td>{m.unit}</td>
                <td>{m.aggregations.join(' / ')}</td>
                <td>{m.numerator}</td>
                <td>{m.denominator}</td>
                <td className="metric__formula">{m.formula}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <h2>Register a custom metric</h2>
        <form className="form" onSubmit={register}>
          <label>
            Metric ID
            <input value={id} onChange={(e) => setID(e.target.value)} required />
          </label>
          <label>
            Name
            <input value={name} onChange={(e) => setName(e.target.value)} placeholder="optional" />
          </label>
          <label>
            Computation
            <select value={kind} onChange={(e) => setKind(e.target.value)}>
              {kinds.map((k) => (
                <option key={k.name} value={k.name}>
                  {k.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Keywords
            <input
              value={keywords}
              onChange={(e) => setKeywords(e.target.value)}
              disabled={kind !== 'keyword_coverage'}
              placeholder="comma separated"
            />
          </label>
          <button type="submit">Register</button>
        </form>
        {message ? <p className="notice">{message}</p> : null}
        <p className="muted">
          A registration must declare its unit, its aggregation methods, its numerator, its denominator and its
          formula. A registration that reuses a built-in identifier is refused.
        </p>
      </section>
    </>
  )
}
