import { useState, type FormEvent, type ReactNode } from 'react'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'
import { useT, useValue } from '../i18n'

// Metrics shows the metric registry and accepts a registration.
//
// The built-in metrics are shown as locked. A registered metric is shown as
// custom, everywhere it appears, so a reader never mistakes one for the
// other.
export function Metrics(): ReactNode {
  const t = useT()
  const value = useValue()
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
      setMessage(t('metrics.registeredMessage', { id }))
      setID('')
      setName('')
      registry.reload()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    }
  }

  if (registry.loading) return <p className="muted">{t('common.loading')}</p>
  if (registry.error) return <ErrorNote message={registry.error} />

  const metrics = registry.data?.metrics ?? []
  const kinds = registry.data?.custom_kinds ?? []

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('metrics.title')}</h1>
        <p className="page__lede">{t('metrics.lede')}</p>
      </header>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('metrics.registered')}</h2>
          <span className="section__note">{t('metrics.registeredCount', { count: metrics.length })}</span>
        </div>
        <div className="sheet sheet--flush">
          {metrics.length === 0 ? <Empty>{t('metrics.noMetrics')}</Empty> : null}
          {metrics.length > 0 ? (
            <div className="table-scroll">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('metrics.metric')}</th>
                    <th>{t('metrics.description')}</th>
                    <th>{t('metrics.source')}</th>
                    <th>{t('metrics.unit')}</th>
                    <th>{t('metrics.aggregation')}</th>
                    <th>{t('metrics.numerator')}</th>
                    <th>{t('metrics.denominator')}</th>
                    <th>{t('metrics.formula')}</th>
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
                          <span className="chip chip--ok" title={t('common.builtInTitle')}>
                            {t('common.builtIn', { version: m.version })}
                          </span>
                        ) : (
                          <span className="chip chip--custom" title={t('common.customTitle')}>
                            {t('common.custom', { version: m.version })}
                          </span>
                        )}
                      </td>
                      <td>{value(m.unit)}</td>
                      <td>{m.aggregations.map(value).join(' / ')}</td>
                      <td>{m.numerator}</td>
                      <td>{m.denominator}</td>
                      <td className="metric__formula">{m.formula}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('metrics.registerCustom')}</h2>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            <form className="form" onSubmit={register}>
              <label>
                {t('metrics.metricID')}
                <input value={id} onChange={(e) => setID(e.target.value)} required />
              </label>
              <label>
                {t('metrics.name')}
                <input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('metrics.optional')} />
              </label>
              <label>
                {t('metrics.computation')}
                <select value={kind} onChange={(e) => setKind(e.target.value)}>
                  {kinds.map((k) => (
                    <option key={k.name} value={k.name}>
                      {k.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                {t('metrics.keywords')}
                <input
                  value={keywords}
                  onChange={(e) => setKeywords(e.target.value)}
                  disabled={kind !== 'keyword_coverage'}
                  placeholder={t('metrics.commaSeparated')}
                />
              </label>
              <button type="submit">{t('metrics.register')}</button>
            </form>
            {message ? <p className="notice">{message}</p> : null}
          </div>
          <p className="sheet__foot">{t('metrics.foot')}</p>
        </div>
      </section>
    </>
  )
}
