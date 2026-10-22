import type { ReactNode } from 'react'
import type { FailureAnalysis } from '../api/types'

// Failures renders the failure section of a report.
//
// Every failed trial is listed with its attribution, because a failed task
// that shows only the word "failed" gives the reader nothing to act on.
export function Failures({ failures }: { failures: FailureAnalysis }): ReactNode {
  const classes = Object.entries(failures.failure_classes).sort(([a], [b]) => a.localeCompare(b))
  const reasons = Object.entries(failures.termination_reasons).sort(([a], [b]) => a.localeCompare(b))

  return (
    <section className="failures">
      <h3>Failure analysis</h3>
      <div className="failures__grid">
        <div>
          <h4>Failure classes</h4>
          <ul className="dist">
            {classes.map(([name, count]) => (
              <li key={name}>
                <span>{name}</span>
                <span className="dist__count">{count}</span>
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h4>Termination reasons</h4>
          <ul className="dist">
            {reasons.map(([name, count]) => (
              <li key={name}>
                <span>{name}</span>
                <span className="dist__count">{count}</span>
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h4>Scorer outcomes</h4>
          <ul className="dist">
            {failures.scorers.map((s) => (
              <li key={s.scorer}>
                <span>{s.scorer}</span>
                <span className="dist__count">
                  {s.pass} pass / {s.fail} fail / {s.unscored} unscored
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <h4>Failed trials</h4>
      {failures.failed_trials.length === 0 ? (
        <p className="empty">No trial failed.</p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>Trial</th>
              <th>Task</th>
              <th>Domain</th>
              <th>Difficulty</th>
              <th>Reason</th>
              <th>Class</th>
            </tr>
          </thead>
          <tbody>
            {failures.failed_trials.map((t) => (
              <tr key={t.trial_id}>
                <td>{t.trial_id}</td>
                <td>{t.task_id}</td>
                <td>{t.domain}</td>
                <td>{t.difficulty}</td>
                <td>{t.termination_reason}</td>
                <td>{t.failure_class}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  )
}
