import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Empty, ErrorNote } from '../components/common'

// Runs lists the executed evaluations.
export function Runs(): ReactNode {
  const runs = useAsync(() => api.listRuns(), [])

  if (runs.loading) return <p className="muted">Loading…</p>
  if (runs.error) return <ErrorNote message={runs.error} />

  const list = runs.data ?? []
  if (list.length === 0) return <Empty>No run has been executed yet.</Empty>

  return (
    <>
      <h1>Runs</h1>
      <table className="table">
        <thead>
          <tr>
            <th>Run</th>
            <th>Task set</th>
            <th>Agent</th>
            <th>Repetitions</th>
            <th>Aggregation</th>
            <th>Started</th>
          </tr>
        </thead>
        <tbody>
          {list.map((run) => (
            <tr key={run.run_id}>
              <td>
                <Link to={`/runs/${run.run_id}`}>{run.run_id}</Link>
              </td>
              <td>
                {run.plan.task_set_id} {run.plan.task_set_version}
              </td>
              <td>{run.plan.agent_id}</td>
              <td>{run.plan.repetitions}</td>
              <td>{run.plan.aggregation}</td>
              <td>{run.started_at}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  )
}
