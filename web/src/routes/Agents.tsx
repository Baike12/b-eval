import { useState, type FormEvent, type ReactNode } from 'react'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'
import { useT, useValue } from '../i18n'

// Agents lists the registered systems under test and their self-check.
//
// The self-check result is shown next to every agent because an agent that
// has not passed it cannot be evaluated, and the reader should not have to
// know that to understand why a run is refused.
export function Agents(): ReactNode {
  const t = useT()
  const value = useValue()
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
      setMessage(t('agents.registeredMessage', { id: agentID }))
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
          ? t('agents.selfCheckPassed', { id })
          : t('agents.selfCheckFailed', { id, reasons: (result?.invalid_reasons ?? []).join('; ') }),
      )
      agents.reload()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    }
  }

  const registered = agents.data ?? []

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('agents.title')}</h1>
        <p className="page__lede">{t('agents.lede')}</p>
      </header>

      {agents.error ? <ErrorNote message={agents.error} /> : null}

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('agents.register')}</h2>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            <form className="form" onSubmit={register}>
              <label>
                {t('agents.agentID')}
                <input value={agentID} onChange={(e) => setAgentID(e.target.value)} required />
              </label>
              <label>
                {t('agents.shape')}
                <select value={shape} onChange={(e) => setShape(e.target.value)}>
                  <option value="mock">{t('agents.shapeMock')}</option>
                  <option value="http">{t('agents.shapeHttp')}</option>
                </select>
              </label>
              <label>
                {t('agents.endpoint')}
                <input
                  value={endpoint}
                  onChange={(e) => setEndpoint(e.target.value)}
                  placeholder="http://127.0.0.1:9000"
                  disabled={shape !== 'http'}
                />
              </label>
              <button type="submit">{t('agents.registerAction')}</button>
            </form>
            {message ? <p className="notice">{message}</p> : null}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('agents.registered')}</h2>
          <span className="section__note">{t('agents.registeredCount', { count: registered.length })}</span>
        </div>
        <div className="sheet sheet--flush">
          {agents.loading ? <p className="empty">{t('common.loading')}</p> : null}
          {!agents.loading && registered.length === 0 ? (
            <Empty>{t('agents.noAgents')}</Empty>
          ) : null}
          {registered.length > 0 ? (
            <div className="table-scroll">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('agents.agent')}</th>
                    <th>{t('agents.shape')}</th>
                    <th>{t('agents.selfCheck')}</th>
                    <th className="is-number">{t('agents.violations')}</th>
                    <th>{t('agents.usageReporting')}</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {registered.map((a) => (
                    <tr key={a.agent_id}>
                      <td>
                        {a.agent_id}
                        <span className="muted"> {a.name}</span>
                      </td>
                      <td>{a.shape}</td>
                      <td>
                        {a.self_check ? (
                          <span className={`chip ${a.self_check.status === 'passed' ? 'chip--ok' : 'chip--danger'}`}>
                            {value(a.self_check.status)}
                          </span>
                        ) : (
                          <span className="chip chip--danger">{t('common.notRun')}</span>
                        )}
                      </td>
                      <td className="is-number">{a.self_check?.contract_violations ?? '—'}</td>
                      <td>{a.self_check ? value(a.self_check.usage_reporting_ok ? 'ok' : 'missing') : '—'}</td>
                      <td>
                        <button className="secondary" type="button" onClick={() => runSelfCheck(a.agent_id)}>
                          {t('agents.runSelfCheck')}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </div>
      </section>
    </>
  )
}
