import type { ReactNode } from 'react'
import type { Confidence } from '../api/types'
import { useT } from '../i18n'
import { formatNumber } from './common'

// ConfidenceSection is the section that opens every report.
//
// It is not collapsible and it is not optional: the platform's rule is that
// a reader sees how trustworthy the number is before the number itself.
//
// The platform failure rate is drawn against the threshold it is judged by,
// because a rate reported as a bare number hides the one thing the reader
// needs to know about it: how close it sits to the limit.

// THRESHOLD_POSITION is where the limit sits on the failure-rate scale.
//
// The limit is pinned rather than fitted, so that a rate below the limit
// always reads as below it. A scale fitted to the observed rate would move
// the limit as the rate moves, and a reader would then be comparing the
// rate against a mark that moves with it.
const THRESHOLD_POSITION = 0.7

export function ConfidenceSection({ confidence }: { confidence: Confidence }): ReactNode {
  const t = useT()
  const insufficient = confidence.insufficient_confidence
  const overThreshold = confidence.platform_failure_rate > confidence.threshold
  const covered = confidence.valid_sample_size / Math.max(1, confidence.planned_sample_size)

  // A rate above the limit is drawn against the same scale, so the bar
  // visibly crosses the threshold instead of being clipped by it.
  const platformScale = confidence.threshold > 0 ? confidence.threshold / THRESHOLD_POSITION : 1
  const platformFill = Math.min(1, confidence.platform_failure_rate / platformScale)
  const judgementScale = Math.max(confidence.judgement_failure_rate, 0.01) * 1.25

  return (
    <section className={`confidence${insufficient ? ' confidence--insufficient' : ''}`} aria-label={t('confidence.title')}>
      <header className="confidence__head">
        <h2>{t('confidence.title')}</h2>
        {insufficient ? (
          <span className="chip chip--danger">{t('confidence.insufficient')}</span>
        ) : (
          <span className="chip chip--ok">{t('confidence.reportable')}</span>
        )}
      </header>
      {insufficient && confidence.reason ? <p className="confidence__reason">{confidence.reason}</p> : null}
      <div className="confidence__grid">
        <div className="gauge">
          <span className="gauge__label">{t('confidence.platformFailureRate')}</span>
          <span className="gauge__value">{formatNumber(confidence.platform_failure_rate)}</span>
          <span className={`meter${overThreshold ? ' meter--danger' : ''}`} aria-hidden="true">
            <span className="meter__fill" style={{ width: `${platformFill * 100}%` }} />
            <span className="meter__notch" style={{ left: `${THRESHOLD_POSITION * 100}%` }} />
          </span>
          <p className="gauge__note">
            {t('confidence.limitLabel')} <b>{formatNumber(confidence.threshold, 2)}</b>{t('confidence.limitNote')}
          </p>
          <p className="gauge__axis">
            <span>0</span>
            <span>{formatNumber(platformScale, 2)}</span>
          </p>
        </div>

        <div className="gauge">
          <span className="gauge__label">{t('confidence.judgementFailureRate')}</span>
          <span className="gauge__value">{formatNumber(confidence.judgement_failure_rate)}</span>
          <span className="meter" aria-hidden="true">
            <span
              className="meter__fill"
              style={{ width: `${(confidence.judgement_failure_rate / judgementScale) * 100}%` }}
            />
          </span>
          <p className="gauge__note">{t('confidence.unscoredNote')}</p>
          <p className="gauge__axis">
            <span>0</span>
            <span>{formatNumber(judgementScale, 2)}</span>
          </p>
        </div>

        <div className="gauge">
          <span className="gauge__label">{t('confidence.validSampleSize')}</span>
          <span className="gauge__value">
            {confidence.valid_sample_size} / {confidence.planned_sample_size}
          </span>
          <span className="meter" aria-hidden="true">
            <span className="meter__fill" style={{ width: `${Math.min(100, covered * 100)}%` }} />
          </span>
          <p className="gauge__note">
            {t('confidence.coverageLabel')} <b>{formatNumber(confidence.coverage)}</b>
          </p>
          <p className="gauge__axis">
            <span>0</span>
            <span>1</span>
          </p>
        </div>
      </div>
    </section>
  )
}
