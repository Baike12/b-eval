import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote, Facts } from '../components/common'
import { useT } from '../i18n'

// TaskSets lists the stored task sets and their tasks.
//
// The form of a task set is shown next to its name, because a reader needs
// to know whether a set needs an environment before they can run it.
export function TaskSets(): ReactNode {
  const t = useT()
  const taskSets = useAsync(() => api.listTaskSets(), [])

  if (taskSets.loading) return <p className="muted">{t('common.loading')}</p>
  if (taskSets.error) return <ErrorNote message={taskSets.error} />

  const sets = taskSets.data ?? []
  if (sets.length === 0) {
    return (
      <>
        <header className="page__head">
          <h1 className="page__title">{t('taskSets.title')}</h1>
        </header>
        <div className="sheet">
          <Empty hint={t('taskSets.noTaskSetsHint')}>{t('taskSets.noTaskSets')}</Empty>
        </div>
      </>
    )
  }

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('taskSets.title')}</h1>
        <p className="page__lede">{t('taskSets.lede')}</p>
      </header>
      {sets.map((ts) => (
        <section className="section" key={`${ts.task_set_id}-${ts.version}`}>
          <div className="section__head">
            <h2 className="section__title">{ts.task_set_id}</h2>
          </div>
          <Facts
            items={[
              { label: t('taskSets.version'), value: ts.version },
              { label: t('taskSets.benchmark'), value: ts.benchmark },
              { label: t('taskSets.tasks'), value: ts.tasks.length },
              { label: t('taskSets.frozen'), value: ts.frozen ? t('common.yes') : t('common.no') },
            ]}
          />
          {ts.description ? <p className="page__lede">{ts.description}</p> : null}
          <div className="sheet sheet--flush">
            <div className="table-scroll">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('taskSets.table.task')}</th>
                    <th>{t('taskSets.table.form')}</th>
                    <th>{t('taskSets.table.capability')}</th>
                    <th>{t('taskSets.table.domain')}</th>
                    <th>{t('taskSets.table.difficulty')}</th>
                    <th>{t('taskSets.table.scorers')}</th>
                  </tr>
                </thead>
                <tbody>
                  {ts.tasks.map((task) => (
                    <tr key={task.task_id}>
                      <td>
                        <Link className="link" to={`/task-sets/${ts.task_set_id}/${ts.version}/${task.task_id}`}>
                          {task.task_id}
                        </Link>
                        <span className="muted"> {task.display_name}</span>
                      </td>
                      <td>
                        <span className="chip">{task.form}</span>
                      </td>
                      <td>{(task.capability.level1_dimensions ?? []).join(', ')}</td>
                      <td>{(task.capability.domains ?? []).join(', ')}</td>
                      <td>{task.capability.difficulty}</td>
                      <td>{task.judgement.scorers.map((s) => s.type).join(', ')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>
      ))}
    </>
  )
}
