import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'
import { useT, useValue } from '../i18n'

// Runs lists the executed evaluations.
export function Runs(): ReactNode {
  const t = useT()
  const value = useValue()
  const runs = useAsync(() => api.listRuns(), [])

  if (runs.loading) return <p className="muted">{t('common.loading')}</p>
  if (runs.error) return <ErrorNote message={runs.error} />

  const list = runs.data ?? []
  if (list.length === 0) return <Empty>{t('runs.noRuns')}</Empty>

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('runs.title')}</h1>
        <p className="page__lede">{t('runs.lede')}</p>
      </header>
      <div className="sheet sheet--flush">
        <div className="table-scroll">
          <table className="table">
            <thead>
              <tr>
                <th>{t('runs.run')}</th>
                <th>{t('runs.taskSet')}</th>
                <th>{t('runs.agent')}</th>
                <th className="is-number">{t('runs.repetitions')}</th>
                <th>{t('runs.aggregation')}</th>
                <th>{t('runs.status')}</th>
                <th>{t('runs.started')}</th>
              </tr>
            </thead>
            <tbody>
              {list.map((run) => (
                <tr key={run.run_id}>
                  <td>
                    <Link className="link mono" to={`/runs/${run.run_id}`}>
                      {run.run_id}
                    </Link>
                  </td>
                  <td>
                    {run.plan.task_set_id} {run.plan.task_set_version}
                  </td>
                  <td>{run.plan.agent_id}</td>
                  <td className="is-number">{run.plan.repetitions}</td>
                  <td>{value(run.plan.aggregation)}</td>
                  <td>
                    <span className={`chip ${run.status === 'completed' ? 'chip--ok' : 'chip--running'}`}>
                      {value(run.status)}
                    </span>
                  </td>
                  <td className="mono">{run.started_at}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  )
}
