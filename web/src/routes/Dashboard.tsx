import { useState, type FormEvent, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'
import { useT, useValue } from '../i18n'

// Dashboard is the landing screen.
//
// It answers one question first: what is the state of the platform right
// now. The run form is on the same screen because starting a run is the
// thing a user came here to do.
export function Dashboard(): ReactNode {
  const t = useT()
  const value = useValue()
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
      setMessage(t('dashboard.runCompleted', { id: run.run_id }))
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  const recent = (runs.data ?? []).slice(0, 8)

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('dashboard.title')}</h1>
        <p className="page__lede">{t('dashboard.lede')}</p>
      </header>

      {health.error ? <ErrorNote message={health.error} /> : null}
      {health.data ? (
        <div className="status">
          <div className="status__cell">
            <span className="status__label">API</span>
            <span className="status__value">
              <span className="pulse" />
              {health.data.status}
            </span>
          </div>
          <div className="status__cell">
            <span className="status__label">{t('dashboard.taskSets')}</span>
            <span className="status__value">{health.data.task_sets}</span>
          </div>
          <div className="status__cell">
            <span className="status__label">{t('dashboard.agentsRegistered')}</span>
            <span className="status__value">{(agents.data ?? []).length}</span>
          </div>
          <div className="status__cell">
            <span className="status__label">{t('dashboard.serverTime')}</span>
            <span className="status__value status__value--mono">{health.data.time}</span>
          </div>
        </div>
      ) : null}

      <section className="section" style={{ marginTop: '38px' }}>
        <div className="section__head">
          <h2 className="section__title">{t('dashboard.startRun')}</h2>
          <span className="section__note">{t('dashboard.startRunNote')}</span>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            <form className="form" onSubmit={startRun}>
              <label>
                {t('dashboard.taskSet')}
                <select value={taskSetID} onChange={(e) => setTaskSetID(e.target.value)} required>
                  <option value="">{t('dashboard.selectTaskSet')}</option>
                  {(taskSets.data ?? []).map((ts) => (
                    <option key={`${ts.task_set_id}-${ts.version}`} value={ts.task_set_id}>
                      {t('dashboard.taskSetOption', {
                        id: ts.task_set_id,
                        version: ts.version,
                        count: ts.tasks.length,
                      })}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                {t('dashboard.agent')}
                <select value={agentID} onChange={(e) => setAgentID(e.target.value)} required>
                  <option value="">{t('dashboard.selectAgent')}</option>
                  {(agents.data ?? []).map((a) => (
                    <option key={a.agent_id} value={a.agent_id}>
                      {a.agent_id}
                      {a.self_check
                        ? ` ${t('dashboard.selfCheckStatus', { status: value(a.self_check.status) })}`
                        : ` ${t('dashboard.noSelfCheck')}`}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                {t('dashboard.repetitions')}
                <input
                  type="number"
                  min={1}
                  max={20}
                  value={repetitions}
                  onChange={(e) => setRepetitions(Number(e.target.value))}
                />
              </label>
              <label>
                {t('dashboard.aggregation')}
                <select value={aggregation} onChange={(e) => setAggregation(e.target.value)}>
                  <option value="macro">{t('dashboard.macroOption')}</option>
                  <option value="micro">{t('dashboard.microOption')}</option>
                </select>
              </label>
              <button type="submit" disabled={submitting || !taskSetID || !agentID}>
                {submitting ? t('dashboard.running') : t('dashboard.run')}
              </button>
            </form>
            {message ? <p className="notice">{message}</p> : null}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('dashboard.recentRuns')}</h2>
          <span className="section__note">{t('dashboard.newestFirst')}</span>
        </div>
        <div className="sheet sheet--flush">
          {runs.loading ? <p className="empty">{t('common.loading')}</p> : null}
          {runs.error ? <ErrorNote message={runs.error} /> : null}
          {!runs.loading && !runs.error && recent.length === 0 ? (
            <Empty hint={t('dashboard.noRunsHint')}>{t('dashboard.noRuns')}</Empty>
          ) : null}
          {recent.length > 0 ? (
            <ul className="runlist">
              {recent.map((run) => (
                <li key={run.run_id}>
                  <Link className="runlist__id link" to={`/runs/${run.run_id}`}>
                    {run.run_id}
                  </Link>
                  <span className="runlist__facts">
                    <span>{run.plan.task_set_id}</span>
                    <span>{run.plan.agent_id}</span>
                    <span>{t('dashboard.repetition', { count: run.plan.repetitions })}</span>
                    <span>{value(run.plan.aggregation)}</span>
                  </span>
                  <span className={`runlist__status chip ${run.status === 'completed' ? 'chip--ok' : 'chip--running'}`}>
                    {value(run.status)}
                  </span>
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      </section>
    </>
  )
}
