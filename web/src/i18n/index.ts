import { useMemo, useSyncExternalStore } from 'react'
import { en } from './locales/en'
import { zhCN } from './locales/zh-CN'
import type { Messages } from './messages'
import { DEFAULT_LOCALE, LOCALES, STORAGE_KEY, type Locale } from './types'
import { valueLabels } from './values'

export { DEFAULT_LOCALE, LOCALES, STORAGE_KEY } from './types'
export type { Locale } from './types'
export { endonyms } from './values'

const catalogues: Record<Locale, Messages> = {
  'zh-CN': zhCN,
  en,
}

// TranslateParams are the values that fill a message's placeholders.
export type TranslateParams = Record<string, string | number>

// Translate turns a message key into the text of the current language.
export type Translate = (key: string, params?: TranslateParams) => string

// Plural is a message that depends on the number it counts.
interface Plural {
  one: string
  other: string
}

// lookup resolves a dotted key in a catalogue.
function lookup(catalogue: Messages, path: string): string | Plural | undefined {
  let node: unknown = catalogue
  for (const part of path.split('.')) {
    if (typeof node !== 'object' || node === null) return undefined
    node = (node as Record<string, unknown>)[part]
  }
  if (typeof node === 'string') return node
  if (typeof node === 'object' && node !== null && 'one' in node && 'other' in node) {
    const plural = node as Plural
    return { one: plural.one, other: plural.other }
  }
  return undefined
}

// interpolate fills the placeholders of a message.
//
// A placeholder that has no value is left in place, because a message that
// shows "{id}" is a defect a reader can see, and one that shows nothing at
// all is not.
function interpolate(text: string, params?: TranslateParams): string {
  if (!params) return text
  return text.replace(/\{(\w+)\}/g, (match, name: string) => {
    const value = params[name]
    return value === undefined ? match : String(value)
  })
}

// createTranslator builds the translator of one language.
//
// A key that the language does not carry falls back to the English
// catalogue, and a key that no catalogue carries is returned as itself: the
// reader sees the key, which names the missing message, instead of an empty
// space where a message should be.
export function createTranslator(locale: Locale): Translate {
  return (key, params) => {
    const node = lookup(catalogues[locale], key) ?? lookup(catalogues.en, key)
    if (node === undefined) return key
    if (typeof node === 'string') return interpolate(node, params)
    const count = Number(params?.count ?? 0)
    return interpolate(count === 1 ? node.one : node.other, params)
  }
}

// readStoredLocale reads the language the reader chose last time.
//
// A browser that refuses to hand over its storage still gets a language:
// the default one.
function readStoredLocale(): Locale {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (stored && (LOCALES as string[]).includes(stored)) return stored as Locale
  } catch {
    // The storage is unreachable; the default language is used instead.
  }
  return DEFAULT_LOCALE
}

// persist stores the language the reader chose.
function persist(locale: Locale): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, locale)
  } catch {
    // The choice is kept in memory for this session only.
  }
}

// applyDocument tells the browser which language the document is in.
//
// The lang attribute decides the font the browser picks for the text and
// the rules it applies to line breaking, so a document that claims English
// while showing Chinese is read by a screen reader in the wrong language.
function applyDocument(locale: Locale): void {
  if (typeof document === 'undefined') return
  document.documentElement.lang = locale
  document.title = catalogues[locale].document.title
}

let current: Locale = readStoredLocale()
const listeners = new Set<() => void>()

// The translators of the shipped languages.
//
// They are built once: a translator is a table lookup and the interface
// asks for the same message many times.
const translators: Record<Locale, Translate> = {
  'zh-CN': createTranslator('zh-CN'),
  en: createTranslator('en'),
}

applyDocument(current)

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function getSnapshot(): Locale {
  return current
}

// setLocale changes the language of the interface.
//
// The change is published to every screen at once, because a report whose
// labels are in one language and whose screen is in another is a screen
// nobody can read.
export function setLocale(locale: Locale): void {
  const changed = locale !== current
  current = locale
  persist(locale)
  applyDocument(locale)
  if (changed) listeners.forEach((listener) => listener())
}

// resetLocale returns the interface to the default language and forgets the
// stored choice.
export function resetLocale(): void {
  try {
    window.localStorage.removeItem(STORAGE_KEY)
  } catch {
    // The storage is unreachable; the in-memory language is still reset.
  }
  current = DEFAULT_LOCALE
  applyDocument(current)
  listeners.forEach((listener) => listener())
}

// getLocale returns the language of the interface outside a screen.
export function getLocale(): Locale {
  return current
}

// useLocale returns the language of the interface and re-renders the screen
// when it changes.
export function useLocale(): Locale {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
}

// useT returns the translator of the current language.
export function useT(): Translate {
  const locale = useLocale()
  return useMemo(() => translators[locale], [locale])
}

// translateNow translates a message outside a screen.
//
// The API client builds its own error message and has no screen to read the
// language from, so it reads the current one instead.
export function translateNow(key: string, params?: TranslateParams): string {
  return translators[current](key, params)
}

// useValue returns the translator of the closed vocabularies the server
// publishes, such as a run status or a termination reason.
export function useValue(): (value: string) => string {
  const locale = useLocale()
  return useMemo(() => (value: string) => valueLabels[locale][value] ?? value, [locale])
}
