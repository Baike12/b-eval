import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Breakdown } from '../components/Breakdown'
import { ConfidenceSection } from '../components/ConfidenceSection'
import { Failures } from '../components/Failures'
import { MetricTable } from '../components/MetricTable'
import { Empty, ErrorNote, IntervalView, Value } from '../components/common'

// RunReport renders the report of one run.
//
// The confidence section comes first and the main score comes second, in
// that order, because the number is only meaningful together with the
// confidence that produced it.
export function RunReport(): ReactNode {
  const { runID = '' } = useParams()
  const report = useAsync(() => api.getReport(runID), [runID])

  if (report.loading) return <p className="muted">Loading…</p>
  if (report.error) return <ErrorNote message={report.error} />
  const data = report.data
  if (!data) return <Empty>No report is available for this run.</Empty>

  const main = data.main_score
  return (
    <>
      <h1>
        Run report <span className="muted">{data.run_id}</span>
      </h1>
      <p className="muted">
        {data.plan.task_set_id} {data.plan.task_set_version} · agent {data.plan.agent_id} · {data.plan.repetitions}{' '}
        repetition(s) · fingerprint {data.plan.fingerprint.slice(0, 12)}
      </p>

      <ConfidenceSection confidence={data.confidence} />

      <section className="panel">
        <h2>Main score</h2>
        <div className="main-score">
          <div className="stat stat--large">
            <span className="stat__label">Total score ({main.aggregation})</span>
            <span className="stat__value">
              <Value value={main.total_score} hasValue={main.has_value} />
            </span>
            <span className="stat__note">
              <IntervalView ci={main.ci} />
            </span>
          </div>
          <div className="stat">
            <span className="stat__label">pass@k</span>
            <span className="stat__value">
              <Value value={main.pass_at_k} hasValue={main.pass_at_k_has_value} />
            </span>
          </div>
          <div className="stat">
            <span className="stat__label">pass^k</span>
            <span className="stat__value">
              <Value value={main.pass_power_k} hasValue={main.pass_power_k_has_value} />
            </span>
          </div>
          <div className="stat">
            <span className="stat__label">Continuous average score</span>
            <span className="stat__value">{main.continuous_mean.toFixed(4)}</span>
          </div>
          <div className="stat">
            <span className="stat__label">Coverage</span>
            <span className="stat__value">{main.coverage.toFixed(4)}</span>
          </div>
        </div>
        <p className="muted">
          The total is the weighted view under the selected aggregation method ({main.aggregation}). It is not a
          single overall capability number. Primary criterion: {main.primary_criterion}
        </p>
      </section>

      <section className="panel">
        <h2>Metrics</h2>
        <MetricTable metrics={data.metrics} />
      </section>

      <section className="panel">
        <h2>Breakdown</h2>
        {data.breakdown.map((b) => (
          <Breakdown key={b.dimension} breakdown={b} />
        ))}
      </section>

      <section className="panel">
        <h2>Tasks</h2>
        <table className="table">
          <thead>
            <tr>
              <th>Task</th>
              <th>Form</th>
              <th>Domain</th>
              <th>Difficulty</th>
              <th>Trials</th>
              <th>Passed</th>
              <th>Success rate</th>
              <th>Mean score</th>
            </tr>
          </thead>
          <tbody>
            {data.tasks.map((t) => (
              <tr key={t.task_id}>
                <td>
                  {t.task_id} <span className="muted">{t.display_name}</span>
                </td>
                <td>{t.form}</td>
                <td>{t.domain}</td>
                <td>{t.difficulty}</td>
                <td>{t.trials}</td>
                <td>{t.passed}</td>
                <td>
                  <Value value={t.success_rate} hasValue={t.has_value} />
                </td>
                <td>{t.mean_score.toFixed(4)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <Failures failures={data.failures} />
      </section>
    </>
  )
}
