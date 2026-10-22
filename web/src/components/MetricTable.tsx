import type { ReactNode } from 'react'
import type { MetricRow } from '../api/types'
import { Coverage, IntervalView, Value } from './common'

// MetricTable lists the metrics of a report.
//
// Every row carries its registry identifier, its version, whether it is a
// built-in metric and its coverage state, so a reader can tell which
// numbers the platform produced and which a registration added.
export function MetricTable({ metrics }: { metrics: MetricRow[] }): ReactNode {
  if (metrics.length === 0) {
    return <p className="empty">No metric was resolved for this run.</p>
  }
  return (
    <table className="table">
      <thead>
        <tr>
          <th>Metric</th>
          <th>Value</th>
          <th>Confidence interval</th>
          <th>Samples</th>
          <th>Numerator / denominator</th>
          <th>Source</th>
        </tr>
      </thead>
      <tbody>
        {metrics.map((m) => (
          <tr key={m.id}>
            <td>
              <span className="metric__name">{m.name}</span>
              <span className="metric__id">{m.id}</span>
              <span className="metric__description">{m.description}</span>
            </td>
            <td>
              <Value value={m.value} hasValue={m.has_value} />
              <span className="metric__aggregation">{m.aggregation}</span>
            </td>
            <td>
              <IntervalView ci={m.ci} />
            </td>
            <td>{m.sample_size}</td>
            <td className="metric__fraction">
              {m.numerator_value} / {m.denominator_value}
              <span className="metric__definition">{m.numerator} ÷ {m.denominator}</span>
            </td>
            <td>
              {m.builtin ? (
                <span className="badge badge--builtin">built-in v{m.version}</span>
              ) : (
                <span className="badge badge--custom" title="A registered metric. It is not a platform metric.">
                  custom v{m.version}
                </span>
              )}
              {m.coverage === 'no coverage' ? <Coverage coverage={m.coverage} /> : null}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
