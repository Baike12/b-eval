import type { ReactNode } from 'react'

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
export function Coverage({ coverage }: { coverage: string }): ReactNode {
  if (coverage === 'no coverage') {
    return (
      <span className="coverage coverage--none" title="No data was collected for this cell. It is excluded from the weighted total.">
        no coverage
      </span>
    )
  }
  return <span className="coverage coverage--ok">{coverage}</span>
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

// Interval renders a confidence interval.
//
// The method that was actually applied is shown next to the interval,
// because the requested method and the applied method differ when the
// requested one does not apply to the metric's unit of observation.
export function IntervalView({ ci }: { ci?: { lower: number; upper: number; method: string; basis: string } }): ReactNode {
  if (!ci) return <span className="muted">—</span>
  return (
    <span className="interval" title={`method: ${ci.method} (${ci.basis})`}>
      [{formatNumber(ci.lower)}, {formatNumber(ci.upper)}]
      <span className="interval__method"> {ci.method}</span>
    </span>
  )
}

// Empty renders the placeholder for a list with nothing in it.
export function Empty({ children }: { children: ReactNode }): ReactNode {
  return <p className="empty">{children}</p>
}

// ErrorNote renders a load error.
export function ErrorNote({ message }: { message: string }): ReactNode {
  return (
    <p className="error" role="alert">
      {message}
    </p>
  )
}
