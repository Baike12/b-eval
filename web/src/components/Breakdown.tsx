import type { ReactNode } from 'react'
import type { Breakdown as BreakdownData } from '../api/types'
import { Coverage, IntervalView, Value } from './common'

// Breakdown renders one decomposition dimension.
//
// A cell with no coverage is shown as an explicit label, never as a zero,
// because a metric of zero and a metric with no data are different facts.
export function Breakdown({ breakdown }: { breakdown: BreakdownData }): ReactNode {
  return (
    <section className="breakdown">
      <h3>{breakdown.dimension}</h3>
      <table className="table">
        <thead>
          <tr>
            <th>Key</th>
            <th>Samples</th>
            <th>Success rate</th>
            <th>Continuous mean</th>
            <th>Confidence interval</th>
          </tr>
        </thead>
        <tbody>
          {breakdown.cells.map((cell) => (
            <tr key={cell.key}>
              <td>{cell.key}</td>
              <td>{cell.sample_size}</td>
              <td>
                <Value value={cell.success_rate} hasValue={cell.has_value} />
              </td>
              <td>{cell.has_value ? cell.continuous_mean.toFixed(4) : <Coverage coverage="no coverage" />}</td>
              <td>
                <IntervalView ci={cell.ci} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}
