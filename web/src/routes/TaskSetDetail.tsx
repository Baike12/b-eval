import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote, Facts } from '../components/common'
import { useT } from '../i18n'

// TaskDetail shows one task.
//
// The form is shown first, together with the fields the form uses, because
// the point of the form is that a task declares what it needs and nothing
// else.
export function TaskDetail(): ReactNode {
  const t = useT()
  const { taskSetID = '', version = '', taskID = '' } = useParams()
  const taskSet = useAsync(() => api.getTaskSet(taskSetID, version), [taskSetID, version])

  if (taskSet.loading) return <p className="muted">{t('common.loading')}</p>
  if (taskSet.error) return <ErrorNote message={taskSet.error} />

  const task = (taskSet.data?.tasks ?? []).find((item) => item.task_id === taskID)
  if (!task) return <Empty>{t('taskDetail.notInSet', { id: taskID })}</Empty>

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{task.task_id}</h1>
        <p className="page__lede">{task.display_name}</p>
        <Facts
          items={[
            { label: t('taskDetail.form'), value: <span className="chip">{task.form}</span> },
            { label: t('taskDetail.benchmark'), value: task.benchmark },
            { label: t('taskDetail.version'), value: task.task_set_version },
            { label: t('taskDetail.difficulty'), value: task.capability.difficulty },
          ]}
        />
      </header>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('taskDetail.instruction')}</h2>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            <p className="instruction">{task.description.instruction}</p>
            {task.description.evaluator_note ? (
              <>
                <h3 className="subhead">{t('taskDetail.evaluatorNote')}</h3>
                <p className="muted">{task.description.evaluator_note}</p>
              </>
            ) : null}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('taskDetail.contractFields')}</h2>
        </div>
        <div className="sheet sheet--flush">
          <div className="table-scroll">
            <table className="table table--pairs">
              <tbody>
                <tr>
                  <th>{t('taskDetail.capability')}</th>
                  <td>
                    {(task.capability.level1_dimensions ?? []).join(', ')} /{' '}
                    {(task.capability.level2_labels ?? []).join(', ')}
                  </td>
                </tr>
                <tr>
                  <th>{t('taskDetail.domain')}</th>
                  <td>{(task.capability.domains ?? []).join(', ')}</td>
                </tr>
                <tr>
                  <th>{t('taskDetail.difficulty')}</th>
                  <td>{task.capability.difficulty}</td>
                </tr>
                <tr>
                  <th>{t('taskDetail.budget')}</th>
                  <td>
                    <Facts
                      items={[
                        { label: t('taskDetail.steps'), value: task.budget.steps },
                        { label: t('taskDetail.tokens'), value: task.budget.tokens },
                        { label: t('taskDetail.duration'), value: `${task.budget.duration_seconds}s` },
                        { label: t('taskDetail.cost'), value: task.budget.cost_usd },
                      ]}
                    />
                  </td>
                </tr>
                <tr>
                  <th>{t('taskDetail.terminationConditions')}</th>
                  <td>{(task.interaction_contract?.termination_conditions ?? []).join(', ') || '—'}</td>
                </tr>
                <tr>
                  <th>{t('taskDetail.multiTurn')}</th>
                  <td>{task.interaction_contract?.multi_turn ? t('common.yes') : t('common.no')}</td>
                </tr>
                <tr>
                  <th>{t('taskDetail.judgement')}</th>
                  <td>
                    <Facts
                      items={[
                        { label: t('taskDetail.combination'), value: task.judgement.combination },
                        { label: t('taskDetail.primary'), value: `#${task.judgement.primary}` },
                        { label: t('taskDetail.scorers'), value: task.judgement.scorers.map((s) => s.type).join(', ') },
                      ]}
                    />
                  </td>
                </tr>
                <tr>
                  <th>{t('taskDetail.referenceAnswer')}</th>
                  <td>{task.judgement.reference_answer || '—'}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </>
  )
}
