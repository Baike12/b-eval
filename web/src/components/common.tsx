import type { CSSProperties, ReactNode } from 'react'
import { useT, useValue } from '../i18n'

// formatNumber renders a metric value.
//
// A value without coverage is rendered as an explicit label rather than as
// zero: a metric of zero and a metric with no data are different facts, and
// a report that shows both as "0" misstates what was measured.
export function formatNumber(value: number, digits = 4): string {
  if (!Number.isFinite(value)) return '—'
  return value.toFixed(digits)
}

// Coverage renders the coverage state of a cell.
//
// The absence of coverage is drawn as an empty hatched track rather than as
// a number, so that a cell with no data can never be read as a cell whose
// value happens to be zero.
export function Coverage({ coverage }: { coverage: string }): ReactNode {
  const t = useT()
  const value = useValue()
  if (coverage === 'no coverage') {
    return (
      <span className="coverage coverage--none" title={t('common.coverageTitle')}>
        {t('common.noCoverage')}
      </span>
    )
  }
  return <span className="coverage coverage--ok">{value(coverage)}</span>
}

// Value renders a metric value together with its coverage state.
export function Value({
  value,
  hasValue,
  digits = 4,
}: {
  value: number
  hasValue: boolean
  digits?: number
}): ReactNode {
  if (!hasValue) return <Coverage coverage="no coverage" />
  return <span className="value">{formatNumber(value, digits)}</span>
}

// IntervalView renders a confidence interval.
//
// The interval is also drawn on a scale, because the width of the interval
// is what tells a reader whether the difference between two numbers is a
// difference or a coincidence. The method that was actually applied is
// shown next to it, because the requested method and the applied method
// differ when the requested one does not apply to the metric's unit of
// observation.
//
// The band is omitted when the caller already draws the interval on a
// scale of its own, because two bands for one interval is one band too
// many and the reader cannot tell which one carries the reading.
export function IntervalView({
  ci,
  band = true,
}: {
  ci?: { lower: number; upper: number; method: string; basis: string }
  band?: boolean
}): ReactNode {
  const t = useT()
  if (!ci) return <span className="muted">—</span>
  const lower = Math.min(ci.lower, ci.upper)
  const upper = Math.max(ci.lower, ci.upper)
  const span = upper - lower
  return (
    <span className="ci" title={t('common.intervalTitle', { method: ci.method, basis: ci.basis })}>
      <span className="ci__text">
        [{formatNumber(lower)}, {formatNumber(upper)}]
      </span>
      <span className="ci__method">{ci.method}</span>
      {band && span > 0 ? (
        <span className="range" aria-hidden="true">
          <span
            className="range__band"
            style={{ '--from': `${lower * 100}%`, '--to': `${upper * 100}%` } as CSSProperties}
          />
        </span>
      ) : null}
    </span>
  )
}

// Facts renders a set of labelled facts.
//
// The facts of a run are shown as separate labelled fields rather than as
// one sentence joined by separators, so that a reader can find the field
// they are looking for without reading the whole line.
export function Facts({ items }: { items: { label: string; value: ReactNode }[] }): ReactNode {
  return (
    <dl className="facts">
      {items.map((item) => (
        <div className="facts__item" key={item.label}>
          <dt className="facts__label">{item.label}</dt>
          <dd className="facts__value">{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}

// Empty renders the placeholder for a list with nothing in it.
//
// An empty screen is an invitation to act, so the placeholder says what to
// do next rather than only reporting that there is nothing.
export function Empty({ children, hint }: { children: ReactNode; hint?: ReactNode }): ReactNode {
  return (
    <p className="empty">
      {children}
      {hint ? <span className="empty__hint">{hint}</span> : null}
    </p>
  )
}

// ErrorNote renders a load error.
export function ErrorNote({ message }: { message: string }): ReactNode {
  return (
    <p className="notice notice--error" role="alert">
      {message}
    </p>
  )
}
