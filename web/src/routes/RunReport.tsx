import type { CSSProperties, ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'
import { useAsync } from '../api/useAsync'
import { Breakdown } from '../components/Breakdown'
import { ConfidenceSection } from '../components/ConfidenceSection'
import { Failures } from '../components/Failures'
import { MetricTable } from '../components/MetricTable'
import { Empty, ErrorNote, Facts, IntervalView, Value } from '../components/common'
import { useT, useValue } from '../i18n'

// RunReport renders the report of one run.
//
// The confidence section comes first and the main score comes second, in
// that order, because the number is only meaningful together with the
// confidence that produced it.
export function RunReport(): ReactNode {
  const t = useT()
  const value = useValue()
  const { runID = '' } = useParams()
  const report = useAsync(() => api.getReport(runID), [runID])

  if (report.loading) return <p className="muted">{t('common.loading')}</p>
  if (report.error) return <ErrorNote message={report.error} />
  const data = report.data
  if (!data) return <Empty>{t('report.noReport')}</Empty>

  const main = data.main_score
  const ci = main.ci
  const bandLeft = ci ? Math.min(ci.lower, ci.upper) : 0
  const bandWidth = ci ? Math.abs(ci.upper - ci.lower) : 0
  const aggregation = value(main.aggregation)

  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('report.title')}</h1>
        <Facts
          items={[
            { label: t('report.run'), value: <span className="mono">{data.run_id}</span> },
            { label: t('report.taskSet'), value: `${data.plan.task_set_id} ${data.plan.task_set_version}` },
            { label: t('report.agent'), value: data.plan.agent_id },
            { label: t('report.repetitions'), value: data.plan.repetitions },
            { label: t('report.fingerprint'), value: <span className="mono">{data.plan.fingerprint.slice(0, 12)}</span> },
          ]}
        />
      </header>

      <ConfidenceSection confidence={data.confidence} />

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('report.mainScore')}</h2>
          <span className="section__note">{t('report.mainScoreNote')}</span>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            <div className="figure">
              <div className="figure__readout">
                <span className="figure__caption">{t('report.totalScore', { aggregation })}</span>
                <span className="figure__value">
                  <Value value={main.total_score} hasValue={main.has_value} />
                </span>
                <span className="figure__interval">
                  <IntervalView ci={main.ci} band={false} />
                </span>
              </div>
              <div className="figure__scale">
                <span className="range range--large" aria-hidden="true">
                  {ci ? (
                    <span
                      className={`range__band${bandWidth > 0 ? '' : ' range__band--point'}`}
                      style={
                        {
                          '--from': `${bandLeft * 100}%`,
                          '--to': `${(bandLeft + bandWidth) * 100}%`,
                        } as CSSProperties
                      }
                    />
                  ) : null}
                  {main.has_value ? (
                    <span
                      className="range__point"
                      style={{ '--at': `${main.total_score * 100}%` } as CSSProperties}
                    />
                  ) : null}
                </span>
                <span className="figure__axis">
                  <span>0</span>
                  <span>{t('report.axis')}</span>
                  <span>1</span>
                </span>
              </div>
            </div>
            <div className="figure__facts">
              <div>
                <span className="fact__label">pass@k</span>
                <span className="fact__value">
                  <Value value={main.pass_at_k} hasValue={main.pass_at_k_has_value} />
                </span>
              </div>
              <div>
                <span className="fact__label">pass^k</span>
                <span className="fact__value">
                  <Value value={main.pass_power_k} hasValue={main.pass_power_k_has_value} />
                </span>
              </div>
              <div>
                <span className="fact__label">{t('report.continuousMean')}</span>
                <span className="fact__value">{main.continuous_mean.toFixed(4)}</span>
              </div>
              <div>
                <span className="fact__label">{t('report.coverage')}</span>
                <span className="fact__value">{main.coverage.toFixed(4)}</span>
              </div>
            </div>
          </div>
          <p className="sheet__foot">
            {t('report.totalNote', { aggregation, criterion: main.primary_criterion })}
          </p>
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('report.metrics')}</h2>
        </div>
        <div className="sheet sheet--flush">
          <MetricTable metrics={data.metrics} />
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('report.breakdown')}</h2>
        </div>
        <div className="sheet">
          <div className="sheet__body">
            {data.breakdown.map((b) => (
              <Breakdown key={b.dimension} breakdown={b} />
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="section__head">
          <h2 className="section__title">{t('report.tasks')}</h2>
        </div>
        <div className="sheet sheet--flush">
          <div className="table-scroll">
            <table className="table">
              <thead>
                <tr>
                  <th>{t('report.table.task')}</th>
                  <th>{t('report.table.form')}</th>
                  <th>{t('report.table.domain')}</th>
                  <th>{t('report.table.difficulty')}</th>
                  <th className="is-number">{t('report.table.trials')}</th>
                  <th className="is-number">{t('report.table.passed')}</th>
                  <th className="is-number">{t('report.table.successRate')}</th>
                  <th className="is-number">{t('report.table.meanScore')}</th>
                </tr>
              </thead>
              <tbody>
                {data.tasks.map((task) => (
                  <tr key={task.task_id}>
                    <td>
                      {task.task_id} <span className="muted">{task.display_name}</span>
                    </td>
                    <td>{task.form}</td>
                    <td>{task.domain}</td>
                    <td>{task.difficulty}</td>
                    <td className="is-number">{task.trials}</td>
                    <td className="is-number">{task.passed}</td>
                    <td className="is-number">
                      <Value value={task.success_rate} hasValue={task.has_value} />
                    </td>
                    <td className="is-number">{task.mean_score.toFixed(4)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="sheet sheet--flush">
          <Failures failures={data.failures} />
        </div>
      </section>
    </>
  )
}
