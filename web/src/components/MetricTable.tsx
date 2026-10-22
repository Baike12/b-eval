import type { ReactNode } from 'react'
import type { MetricRow } from '../api/types'
import { useT, useValue } from '../i18n'
import { Coverage, IntervalView, Value } from './common'

// MetricTable lists the metrics of a report.
//
// Every row carries its registry identifier, its version, whether it is a
// built-in metric and its coverage state, so a reader can tell which
// numbers the platform produced and which a registration added.
export function MetricTable({ metrics }: { metrics: MetricRow[] }): ReactNode {
  const t = useT()
  const value = useValue()
  if (metrics.length === 0) {
    return <p className="empty">{t('metricTable.noMetrics')}</p>
  }
  return (
    <div className="table-scroll">
      <table className="table">
        <thead>
          <tr>
            <th>{t('metricTable.metric')}</th>
            <th className="is-number">{t('metricTable.value')}</th>
            <th>{t('metricTable.interval')}</th>
            <th className="is-number">{t('metricTable.samples')}</th>
            <th className="is-number">{t('metricTable.fraction')}</th>
            <th>{t('metricTable.source')}</th>
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
              <td className="is-number">
                <Value value={m.value} hasValue={m.has_value} />
                <span className="metric__aggregation">{value(m.aggregation)}</span>
              </td>
              <td>
                <IntervalView ci={m.ci} />
              </td>
              <td className="is-number">{m.sample_size}</td>
              <td className="is-number metric__fraction">
                {m.numerator_value} / {m.denominator_value}
                <span className="metric__definition">
                  {m.numerator} ÷ {m.denominator}
                </span>
              </td>
              <td>
                {m.builtin ? (
                  <span className="chip chip--ok" title={t('common.builtInTitle')}>
                    {t('common.builtIn', { version: m.version })}
                  </span>
                ) : (
                  <span className="chip chip--custom" title={t('common.customNotPlatformTitle')}>
                    {t('common.custom', { version: m.version })}
                  </span>
                )}
                {m.coverage === 'no coverage' ? <Coverage coverage={m.coverage} /> : null}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
