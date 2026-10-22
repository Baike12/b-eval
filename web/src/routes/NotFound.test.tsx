import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { App } from '../App'

describe('the route table', () => {
  it('answers an address that names nothing instead of rendering nothing', () => {
    render(
      <MemoryRouter initialEntries={['/task-sets/demo/v1']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByRole('heading', { level: 1, name: 'Nothing here' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Back to the dashboard' })).toBeInTheDocument()
  })
})
