import { describe, expect, it } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { endonyms, useT } from '../i18n'
import { zhCN } from '../i18n/locales/zh-CN'
import { LocaleSwitch } from './LocaleSwitch'

// The expectations are taken from the catalogue rather than written here,
// because the catalogue is the one place in the tree that holds the text of
// a translation.
const CHINESE = endonyms['zh-CN']
const ENGLISH = endonyms.en

function Probe(): ReactNode {
  const t = useT()
  return <span>{t('nav.dashboard')}</span>
}

describe('the language switch', () => {
  it('offers every language by the name that language gives itself', () => {
    render(<LocaleSwitch />)
    expect(screen.getByRole('button', { name: CHINESE })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: ENGLISH })).toBeInTheDocument()
  })

  it('marks the language the interface is in', () => {
    render(<LocaleSwitch />)
    expect(screen.getByRole('button', { name: ENGLISH })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: CHINESE })).toHaveAttribute('aria-pressed', 'false')
  })

  it('changes the language of the screen at once', () => {
    render(
      <>
        <LocaleSwitch />
        <Probe />
      </>,
    )
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: CHINESE }))
    expect(screen.getByText(zhCN.nav.dashboard)).toBeInTheDocument()
    expect(document.documentElement.lang).toBe('zh-CN')
  })
})
