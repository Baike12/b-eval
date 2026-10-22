import type { ReactNode } from 'react'
import type { Breakdown as BreakdownData } from '../api/types'
import { useT, useValue } from '../i18n'
import { Coverage, IntervalView, Value } from './common'

// Breakdown renders one decomposition dimension.
//
// A cell with no coverage is shown as an explicit label, never as a zero,
// because a metric of zero and a metric with no data are different facts.
export function Breakdown({ breakdown }: { breakdown: BreakdownData }): ReactNode {
  const t = useT()
  const value = useValue()
  return (
    <section className="breakdown">
      <h3 className="subhead">{value(breakdown.dimension)}</h3>
      <div className="table-scroll">
        <table className="table">
          <thead>
            <tr>
              <th>{t('breakdown.key')}</th>
              <th className="is-number">{t('breakdown.samples')}</th>
              <th className="is-number">{t('breakdown.successRate')}</th>
              <th className="is-number">{t('breakdown.continuousMean')}</th>
              <th>{t('breakdown.interval')}</th>
            </tr>
          </thead>
          <tbody>
            {breakdown.cells.map((cell) => (
              <tr key={cell.key}>
                <td>{cell.key}</td>
                <td className="is-number">{cell.sample_size}</td>
                <td className="is-number">
                  <Value value={cell.success_rate} hasValue={cell.has_value} />
                </td>
                <td className="is-number">
                  {cell.has_value ? cell.continuous_mean.toFixed(4) : <Coverage coverage="no coverage" />}
                </td>
                <td>
                  <IntervalView ci={cell.ci} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
