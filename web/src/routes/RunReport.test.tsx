import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { RunReport } from './RunReport'
import type { Report } from '../api/types'

const report: Report = {
  run_id: 'run-1',
  generated_at: '2026-10-22T00:00:00Z',
  plan: {
    plan_id: 'plan-1',
    task_set_id: 'demo',
    task_set_version: 'v1',
    agent_id: 'agent-1',
    repetitions: 2,
    aggregation: 'macro',
    ci_method: 'wilson',
    fingerprint: 'abcdef0123456789',
  },
  confidence: {
    platform_failure_rate: 0,
    judgement_failure_rate: 0,
    valid_sample_size: 8,
    planned_sample_size: 8,
    coverage: 1,
    threshold: 0.05,
    insufficient_confidence: false,
  },
  main_score: {
    aggregation: 'macro',
    total_score: 0.5,
    has_value: true,
    weights: {},
    coverage: 1,
    pass_at_k: 0.75,
    pass_at_k_has_value: true,
    pass_power_k: 0.25,
    pass_power_k_has_value: true,
    continuous_mean: 0.6,
    primary_criterion: 'The task\'s primary criterion is satisfied.',
  },
  metrics: [
    {
      id: 'success_rate',
      name: 'Success rate',
      description: 'Share of valid trials whose primary criterion was satisfied.',
      builtin: true,
      unit: 'ratio',
      version: '1',
      formula: 'macro',
      numerator: 'Passing trials',
      denominator: 'Valid trials',
      aggregation: 'macro',
      value: 0.5,
      has_value: true,
      coverage: 'covered',
      numerator_value: 4,
      denominator_value: 8,
      sample_size: 8,
    },
  ],
  breakdown: [
    {
      dimension: 'domain',
      cells: [
        {
          key: 'D7',
          sample_size: 8,
          success_rate: 0.5,
          has_value: true,
          coverage: 'covered',
          continuous_mean: 0.6,
        },
      ],
    },
  ],
  failures: {
    termination_reasons: { completed: 8 },
    failure_classes: { completed: 8 },
    scorers: [{ scorer: 'contains', pass: 4, fail: 4, unscored: 0, share: {} }],
    failed_trials: [
      {
        trial_id: 'T-1#0#1',
        task_id: 'T-1',
        domain: 'D7',
        difficulty: 'L1',
        termination_reason: 'completed',
        failure_class: 'completed',
      },
    ],
  },
  tasks: [
    {
      task_id: 'T-1',
      display_name: 'Refund eligibility',
      form: 'static-qa',
      domain: 'D7',
      difficulty: 'L1',
      trials: 2,
      passed: 1,
      success_rate: 0.5,
      has_value: true,
      mean_score: 0.5,
    },
  ],
}

describe('RunReport', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify(report), { status: 200, headers: { 'Content-Type': 'application/json' } })),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the report with the confidence section first', async () => {
    render(
      <MemoryRouter initialEntries={['/runs/run-1']}>
        <Routes>
          <Route path="/runs/:runID" element={<RunReport />} />
        </Routes>
      </MemoryRouter>,
    )
    await waitFor(() => expect(screen.getByText('Confidence')).toBeInTheDocument())
    const confidenceHeading = screen.getByRole('heading', { level: 2, name: 'Confidence' })
    const mainScoreHeading = screen.getByRole('heading', { level: 2, name: 'Main score' })
    expect(confidenceHeading).toBeInTheDocument()
    expect(mainScoreHeading).toBeInTheDocument()
    // The confidence section must come before the number, not after it.
    expect(
      confidenceHeading.compareDocumentPosition(mainScoreHeading) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
    expect(screen.getAllByText('0.5000').length).toBeGreaterThan(0)
  })

  it('states that the total is not a single overall capability number', async () => {
    render(
      <MemoryRouter initialEntries={['/runs/run-1']}>
        <Routes>
          <Route path="/runs/:runID" element={<RunReport />} />
        </Routes>
      </MemoryRouter>,
    )
    await waitFor(() => expect(screen.getByText('Confidence')).toBeInTheDocument())
    expect(screen.getByText(/not a single overall capability number/)).toBeInTheDocument()
  })

  it('lists the failed trials with their attribution', async () => {
    render(
      <MemoryRouter initialEntries={['/runs/run-1']}>
        <Routes>
          <Route path="/runs/:runID" element={<RunReport />} />
        </Routes>
      </MemoryRouter>,
    )
    await waitFor(() => expect(screen.getByText('Failure analysis')).toBeInTheDocument())
    expect(screen.getByText('T-1#0#1')).toBeInTheDocument()
  })
})
