import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ConfidenceSection } from './ConfidenceSection'
import type { Confidence } from '../api/types'

const base: Confidence = {
  platform_failure_rate: 0.01,
  judgement_failure_rate: 0.02,
  valid_sample_size: 48,
  planned_sample_size: 50,
  coverage: 0.96,
  threshold: 0.05,
  insufficient_confidence: false,
}

describe('ConfidenceSection', () => {
  it('reports the confidence figures', () => {
    render(<ConfidenceSection confidence={base} />)
    expect(screen.getByText('Confidence')).toBeInTheDocument()
    expect(screen.getByText('0.0100')).toBeInTheDocument()
    expect(screen.getByText('48 / 50')).toBeInTheDocument()
    expect(screen.getByText('reportable')).toBeInTheDocument()
  })

  it('flags a report whose platform failure rate exceeds the threshold', () => {
    render(<ConfidenceSection confidence={{ ...base, platform_failure_rate: 0.111, insufficient_confidence: true, reason: 'the platform failure rate 0.1110 exceeds the 0.05 threshold' }} />)
    expect(screen.getByText('insufficient confidence')).toBeInTheDocument()
    expect(screen.getByText(/exceeds the 0.05 threshold/)).toBeInTheDocument()
  })

  it('cannot be collapsed away', () => {
    // The platform's rule is that the confidence section is always visible
    // and never collapsible. There must be no control that hides it.
    const { container } = render(<ConfidenceSection confidence={base} />)
    expect(container.querySelector('details')).toBeNull()
    expect(container.querySelector('button')).toBeNull()
  })
})
