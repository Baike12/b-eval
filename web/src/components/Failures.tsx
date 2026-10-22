import type { ReactNode } from 'react'
import type { FailureAnalysis } from '../api/types'
import { useT, useValue } from '../i18n'

// Failures renders the failure section of a report.
//
// Every failed trial is listed with its attribution, because a failed task
// that shows only the word "failed" gives the reader nothing to act on.
export function Failures({ failures }: { failures: FailureAnalysis }): ReactNode {
  const t = useT()
  const value = useValue()
  const classes = Object.entries(failures.failure_classes).sort(([a], [b]) => a.localeCompare(b))
  const reasons = Object.entries(failures.termination_reasons).sort(([a], [b]) => a.localeCompare(b))

  return (
    <section className="failures sheet__body">
      <h2 className="section__title">{t('failures.title')}</h2>
      <div className="failures__grid">
        <div>
          <h3 className="subhead">{t('failures.classes')}</h3>
          <ul className="dist">
            {classes.map(([name, count]) => (
              <li key={name}>
                <span>{value(name)}</span>
                <span className="dist__count">{count}</span>
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h3 className="subhead">{t('failures.reasons')}</h3>
          <ul className="dist">
            {reasons.map(([name, count]) => (
              <li key={name}>
                <span>{value(name)}</span>
                <span className="dist__count">{count}</span>
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h3 className="subhead">{t('failures.scorers')}</h3>
          <ul className="dist">
            {failures.scorers.map((s) => (
              <li key={s.scorer}>
                <span>{s.scorer}</span>
                <span className="dist__count">
                  {t('failures.outcome', { pass: s.pass, fail: s.fail, unscored: s.unscored })}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <h3 className="subhead failures__trials">{t('failures.failedTrials')}</h3>
      {failures.failed_trials.length === 0 ? (
        <p className="empty">{t('failures.noFailedTrials')}</p>
      ) : (
        <div className="table-scroll">
          <table className="table">
            <thead>
              <tr>
                <th>{t('failures.table.trial')}</th>
                <th>{t('failures.table.task')}</th>
                <th>{t('failures.table.domain')}</th>
                <th>{t('failures.table.difficulty')}</th>
                <th>{t('failures.table.reason')}</th>
                <th>{t('failures.table.class')}</th>
              </tr>
            </thead>
            <tbody>
              {failures.failed_trials.map((trial) => (
                <tr key={trial.trial_id}>
                  <td className="mono">{trial.trial_id}</td>
                  <td>{trial.task_id}</td>
                  <td>{trial.domain}</td>
                  <td>{trial.difficulty}</td>
                  <td>{value(trial.termination_reason)}</td>
                  <td>{value(trial.failure_class)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
