import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import {
  DEFAULT_LOCALE,
  STORAGE_KEY,
  createTranslator,
  getLocale,
  resetLocale,
  setLocale,
  translateNow,
  useT,
  useValue,
} from './index'
import { en } from './locales/en'
import { zhCN } from './locales/zh-CN'

// keyPaths flattens a catalogue into its dotted keys.
function keyPaths(node: unknown, prefix = ''): string[] {
  if (typeof node !== 'object' || node === null) return [prefix]
  return Object.entries(node as Record<string, unknown>).flatMap(([key, value]) =>
    keyPaths(value, prefix ? `${prefix}.${key}` : key),
  )
}

function Probe(): ReactNode {
  const t = useT()
  const value = useValue()
  return (
    <span>
      {t('nav.dashboard')}|{value('completed')}|{value('a_value_the_server_did_not_publish')}
    </span>
  )
}

describe('the language of the interface', () => {
  it('is Chinese until a reader says otherwise', () => {
    resetLocale()
    expect(DEFAULT_LOCALE).toBe('zh-CN')
    expect(getLocale()).toBe('zh-CN')
    expect(createTranslator('zh-CN')('nav.dashboard')).toBe('总览')
  })

  it('is the language the reader chose last time', async () => {
    window.localStorage.setItem(STORAGE_KEY, 'zh-CN')
    vi.resetModules()
    const fresh = await import('./index')
    expect(fresh.getLocale()).toBe('zh-CN')
  })

  it('is remembered for the next visit', () => {
    setLocale('en')
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('en')
    resetLocale()
    expect(window.localStorage.getItem(STORAGE_KEY)).toBeNull()
  })

  it('tells the browser which language the document is in', () => {
    setLocale('zh-CN')
    expect(document.documentElement.lang).toBe('zh-CN')
    setLocale('en')
    expect(document.documentElement.lang).toBe('en')
  })

  it('translates the closed vocabularies the server publishes', () => {
    setLocale('zh-CN')
    render(<Probe />)
    expect(screen.getByText('总览|已完成|a_value_the_server_did_not_publish')).toBeInTheDocument()
  })

  it('carries the same keys in every language', () => {
    expect(keyPaths(zhCN).sort()).toEqual(keyPaths(en).sort())
  })

  it('translates a message that is built outside a screen', () => {
    setLocale('zh-CN')
    expect(translateNow('common.requestFailed', { status: 500 })).toBe('请求失败，状态码 500')
  })
})

describe('a message', () => {
  it('fills its placeholders', () => {
    expect(createTranslator('en')('dashboard.runCompleted', { id: 'run-1' })).toBe('Run run-1 completed.')
    expect(createTranslator('zh-CN')('dashboard.runCompleted', { id: 'run-1' })).toBe('运行 run-1 已完成。')
  })

  it('leaves a placeholder that has no value in place', () => {
    expect(createTranslator('en')('dashboard.runCompleted', {})).toBe('Run {id} completed.')
  })

  it('counts in the singular or in the plural', () => {
    expect(createTranslator('en')('dashboard.repetition', { count: 1 })).toBe('1 repetition')
    expect(createTranslator('en')('dashboard.repetition', { count: 3 })).toBe('3 repetitions')
  })

  it('names itself when no catalogue carries it', () => {
    expect(createTranslator('zh-CN')('nothing.like.this')).toBe('nothing.like.this')
  })
})
