import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// TaskDetail shows one task.
//
// The form is shown first, together with the fields the form uses, because
// the point of the form is that a task declares what it needs and nothing
// else.
export function TaskDetail(): ReactNode {
  const { taskSetID = '', version = '', taskID = '' } = useParams()
  const taskSet = useAsync(() => api.getTaskSet(taskSetID, version), [taskSetID, version])

  if (taskSet.loading) return <p className="muted">Loading…</p>
  if (taskSet.error) return <ErrorNote message={taskSet.error} />

  const task = (taskSet.data?.tasks ?? []).find((t) => t.task_id === taskID)
  if (!task) return <Empty>Task {taskID} is not in this task set.</Empty>

  return (
    <>
      <h1>
        {task.task_id} <span className="muted">{task.display_name}</span>
      </h1>
      <p>
        <span className={`badge badge--form badge--${task.form}`}>{task.form}</span>{' '}
        <span className="badge badge--builtin">benchmark {task.benchmark}</span>{' '}
        <span className="badge badge--builtin">version {task.task_set_version}</span>
      </p>

      <section className="panel">
        <h2>Instruction</h2>
        <p className="instruction">{task.description.instruction}</p>
        {task.description.evaluator_note ? (
          <>
            <h3>Evaluator note</h3>
            <p className="muted">{task.description.evaluator_note}</p>
          </>
        ) : null}
      </section>

      <section className="panel">
        <h2>Contract fields</h2>
        <table className="table table--pairs">
          <tbody>
            <tr>
              <th>Capability</th>
              <td>
                {(task.capability.level1_dimensions ?? []).join(', ')} /{' '}
                {(task.capability.level2_labels ?? []).join(', ')}
              </td>
            </tr>
            <tr>
              <th>Domain</th>
              <td>{(task.capability.domains ?? []).join(', ')}</td>
            </tr>
            <tr>
              <th>Difficulty</th>
              <td>{task.capability.difficulty}</td>
            </tr>
            <tr>
              <th>Budget</th>
              <td>
                steps {task.budget.steps} · tokens {task.budget.tokens} · duration {task.budget.duration_seconds}s · cost{' '}
                {task.budget.cost_usd}
              </td>
            </tr>
            <tr>
              <th>Termination conditions</th>
              <td>{(task.interaction_contract?.termination_conditions ?? []).join(', ') || '—'}</td>
            </tr>
            <tr>
              <th>Multi-turn</th>
              <td>{task.interaction_contract?.multi_turn ? 'yes' : 'no'}</td>
            </tr>
            <tr>
              <th>Judgement</th>
              <td>
                {task.judgement.combination} · primary #{task.judgement.primary} ·{' '}
                {task.judgement.scorers.map((s) => s.type).join(', ')}
              </td>
            </tr>
            <tr>
              <th>Reference answer</th>
              <td>{task.judgement.reference_answer || '—'}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </>
  )
}
