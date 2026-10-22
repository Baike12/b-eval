import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from './ErrorBoundary'

function Broken(): never {
  throw new Error('the report carried a null collection')
}

describe('ErrorBoundary', () => {
  it('names the fault instead of leaving a blank screen', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <ErrorBoundary>
        <Broken />
      </ErrorBoundary>,
    )
    expect(screen.getByRole('heading', { name: 'This screen could not be rendered' })).toBeInTheDocument()
    expect(screen.getByText(/null collection/)).toBeInTheDocument()
  })

  it('renders its children while nothing has failed', () => {
    render(
      <ErrorBoundary>
        <p>the report</p>
      </ErrorBoundary>,
    )
    expect(screen.getByText('the report')).toBeInTheDocument()
  })
})
