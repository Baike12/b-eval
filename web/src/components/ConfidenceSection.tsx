import type { ReactNode } from 'react'
import type { Confidence } from '../api/types'
import { formatNumber } from './common'

// ConfidenceSection is the section that opens every report.
//
// It is not collapsible and it is not optional: the platform's rule is that
// a reader sees how trustworthy the number is before the number itself.
export function ConfidenceSection({ confidence }: { confidence: Confidence }): ReactNode {
  const insufficient = confidence.insufficient_confidence
  return (
    <section className={`confidence${insufficient ? ' confidence--insufficient' : ''}`} aria-label="Confidence">
      <header className="confidence__header">
        <h2>Confidence</h2>
        {insufficient ? (
          <span className="badge badge--danger">insufficient confidence</span>
        ) : (
          <span className="badge badge--ok">reportable</span>
        )}
      </header>
      {insufficient && confidence.reason ? <p className="confidence__reason">{confidence.reason}</p> : null}
      <div className="confidence__grid">
        <div className="stat">
          <span className="stat__label">Platform failure rate</span>
          <span className="stat__value">{formatNumber(confidence.platform_failure_rate)}</span>
          <span className="stat__note">threshold {formatNumber(confidence.threshold, 2)}</span>
        </div>
        <div className="stat">
          <span className="stat__label">Judgement failure rate</span>
          <span className="stat__value">{formatNumber(confidence.judgement_failure_rate)}</span>
          <span className="stat__note">trials whose verdict is UNSCORED</span>
        </div>
        <div className="stat">
          <span className="stat__label">Valid sample size</span>
          <span className="stat__value">
            {confidence.valid_sample_size} / {confidence.planned_sample_size}
          </span>
          <span className="stat__note">coverage {formatNumber(confidence.coverage)}</span>
        </div>
      </div>
    </section>
  )
}
