import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MetricTable } from './MetricTable'
import type { MetricRow } from '../api/types'

const builtin: MetricRow = {
  id: 'success_rate',
  name: 'Success rate',
  description: 'Share of valid trials whose primary criterion was satisfied.',
  builtin: true,
  unit: 'ratio',
  version: '1',
  formula: 'micro: passing trials / valid trials',
  numerator: 'Valid trials judged as passing',
  denominator: 'Valid trials',
  aggregation: 'macro',
  value: 0.5,
  has_value: true,
  coverage: 'covered',
  numerator_value: 4,
  denominator_value: 8,
  sample_size: 8,
  ci: { lower: 0.2, upper: 0.8, method: 'wilson', requested: 'wilson', basis: 'pooled', level: 0.95 },
}

const custom: MetricRow = {
  ...builtin,
  id: 'keyword_coverage',
  name: 'Keyword coverage',
  builtin: false,
  version: '1',
  value: 0,
  has_value: false,
  coverage: 'no coverage',
  sample_size: 0,
  ci: undefined,
}

describe('MetricTable', () => {
  it('marks a registered metric as custom', () => {
    render(<MetricTable metrics={[builtin, custom]} />)
    expect(screen.getByText('built-in v1')).toBeInTheDocument()
    expect(screen.getByText('custom v1')).toBeInTheDocument()
  })

  it('shows a metric with no coverage as an explicit label', () => {
    render(<MetricTable metrics={[custom]} />)
    expect(screen.getAllByText('no coverage').length).toBeGreaterThan(0)
  })

  it('shows the numerator and the denominator of every metric', () => {
    render(<MetricTable metrics={[builtin]} />)
    expect(screen.getByText('Valid trials judged as passing ÷ Valid trials')).toBeInTheDocument()
  })

  it('shows the confidence interval method that was applied', () => {
    render(<MetricTable metrics={[builtin]} />)
    expect(screen.getByText('wilson')).toBeInTheDocument()
  })
})
