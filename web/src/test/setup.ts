import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, beforeEach } from 'vitest'
import { setLocale } from '../i18n'

// The auto-cleanup of the testing library only registers itself when the
// test runner exposes the globals. This project does not, so the cleanup is
// registered explicitly: without it a rendered tree stays in the document
// and the next assertion finds the elements of the previous test.
afterEach(() => {
  cleanup()
})

// The runtime the suite runs in exposes no usable local storage unless it
// is started with a storage file, and a browser has one. The interface
// keeps the reader's language in that storage, so the tests install a small
// in-memory store and exercise the behaviour a browser actually offers.
if (typeof window !== 'undefined') {
  const entries = new Map<string, string>()
  const store = {
    get length() {
      return entries.size
    },
    clear: () => entries.clear(),
    getItem: (key: string) => (entries.has(key) ? (entries.get(key) as string) : null),
    key: (index: number) => [...entries.keys()][index] ?? null,
    removeItem: (key: string) => void entries.delete(key),
    setItem: (key: string, value: string) => void entries.set(key, String(value)),
  }
  Object.defineProperty(window, 'localStorage', { value: store, configurable: true })
}

// The suite reads the English catalogue, so that an assertion states the
// message it was written against. The Chinese catalogue has its own test,
// and the language is restored before every test so that one test cannot
// leave the next one in another language.
beforeEach(() => {
  setLocale('en')
})
