import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// TaskSets lists the stored task sets and their tasks.
//
// The form of a task set is shown next to its name, because a reader needs
// to know whether a set needs an environment before they can run it.
export function TaskSets(): ReactNode {
  const taskSets = useAsync(() => api.listTaskSets(), [])

  if (taskSets.loading) return <p className="muted">Loading…</p>
  if (taskSets.error) return <ErrorNote message={taskSets.error} />

  const sets = taskSets.data ?? []
  if (sets.length === 0) {
    return (
      <>
        <h1>Task sets</h1>
        <Empty>No task set is stored yet. Run `b-eval seed` to write the demonstration set.</Empty>
      </>
    )
  }

  return (
    <>
      <h1>Task sets</h1>
      {sets.map((ts) => (
        <section className="panel" key={`${ts.task_set_id}-${ts.version}`}>
          <h2>
            {ts.task_set_id}{' '}
            <span className="muted">
              {ts.version} · {ts.benchmark} · {ts.tasks.length} tasks
            </span>
          </h2>
          {ts.description ? <p className="muted">{ts.description}</p> : null}
          <table className="table">
            <thead>
              <tr>
                <th>Task</th>
                <th>Form</th>
                <th>Capability</th>
                <th>Domain</th>
                <th>Difficulty</th>
                <th>Scorers</th>
              </tr>
            </thead>
            <tbody>
              {ts.tasks.map((task) => (
                <tr key={task.task_id}>
                  <td>
                    <Link to={`/task-sets/${ts.task_set_id}/${ts.version}/${task.task_id}`}>{task.task_id}</Link>
                    <span className="muted"> {task.display_name}</span>
                  </td>
                  <td>
                    <span className={`badge badge--form badge--${task.form}`}>{task.form}</span>
                  </td>
                  <td>{(task.capability.level1_dimensions ?? []).join(', ')}</td>
                  <td>{(task.capability.domains ?? []).join(', ')}</td>
                  <td>{task.capability.difficulty}</td>
                  <td>{task.judgement.scorers.map((s) => s.type).join(', ')}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ))}
    </>
  )
}
