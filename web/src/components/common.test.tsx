import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Coverage, Value, formatNumber } from './common'

describe('formatNumber', () => {
  it('renders a fixed number of digits', () => {
    expect(formatNumber(0.5)).toBe('0.5000')
    expect(formatNumber(0.5, 2)).toBe('0.50')
  })

  it('renders a dash for a value that is not a number', () => {
    expect(formatNumber(Number.NaN)).toBe('—')
    expect(formatNumber(Number.POSITIVE_INFINITY)).toBe('—')
  })
})

describe('Value', () => {
  it('shows the number when the metric has coverage', () => {
    render(<Value value={0.75} hasValue />)
    expect(screen.getByText('0.7500')).toBeInTheDocument()
  })

  it('shows an explicit label rather than zero when there is no coverage', () => {
    // A metric of zero and a metric with no data are different facts, so
    // the second one must never be rendered as "0.0000".
    render(<Value value={0} hasValue={false} />)
    expect(screen.queryByText('0.0000')).not.toBeInTheDocument()
    expect(screen.getByText('no coverage')).toBeInTheDocument()
  })
})

describe('Coverage', () => {
  it('marks the no-coverage state explicitly', () => {
    render(<Coverage coverage="no coverage" />)
    expect(screen.getByText('no coverage')).toBeInTheDocument()
  })
})
