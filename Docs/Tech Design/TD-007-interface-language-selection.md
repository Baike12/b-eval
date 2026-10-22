# TD-007: Interface Language Selection

| Field | Value |
|---|---|
| **ID** | TD-007 |
| **Status** | Implemented |
| **Implements** | [FEAT-010 Interface Language Selection](../Features/experience/interface-language-selection.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the interface changes language. The *what* and *why* live in the Feature document it implements.

## 1. Problem Restatement

The single-page application writes every label as a literal in the component that shows it. There are seven screens, roughly 190 pieces of interface text between them, and each of them is a string in a `.tsx` file.

A language switch therefore cannot be a label on the switch: the reader who presses it has to find a screen full of English. Every piece of interface text has to be addressed by a key, and every key has to exist in every language the platform ships.

Two properties make this expensive if they are not designed for:

**The count is large and the text is small.** A catalogue of 190 short messages is a table, not prose. It wants to be checked by a compiler rather than by a reader, because a reader misses the eleventh row of a diff.

**The platform publishes closed vocabularies.** A run status, a coverage state, a termination reason and a failure class are values the server sends. They are not interface text — they are part of the platform's contract — but they are shown to the reader as words, and a reader who cannot read English cannot read `budget_exhausted` either.

## 2. Design Goals

- Every piece of interface text is addressed by a key, and the key set is fixed by the compiler.
- A language change reaches every screen at once, without a reload and without a route change.
- The choice survives the next visit.
- A value the server publishes is shown in the reader's language when the platform knows the value, and unchanged when it does not.
- The change adds no dependency to the frontend bundle.

## 3. Non-Goals

- A message that is missing from a language shows the English message rather than a blank. The Chinese catalogue is typed against the English one, so this is a fallback, not a workflow.
- The catalogue is not translated at runtime and not extracted from the sources. There is no build step between a literal and a key, because a build step that fails on an untranslated string is a gate that has to be maintained, and the compiler already does the same job for free.
- The wire contract does not change. The API keeps publishing `no coverage`, `completed` and `budget_exhausted`.

## 4. The Architecture

The module is `web/src/i18n/`, and it has five parts.

| File | Holds |
|---|---|
| `types.ts` | The `Locale` type, the default locale, the list of locales, the storage key |
| `locales/en.ts` | The English catalogue — the reference that fixes the key set |
| `locales/zh-CN.ts` | The Chinese catalogue, typed against the reference |
| `messages.ts` | The `Messages` type, derived from the reference catalogue |
| `values.ts` | The closed vocabularies the server publishes, and the endonyms of the languages |
| `index.ts` | The store, the translator, the hooks |

### 4.1 The catalogue is typed against the reference

`Messages` is derived from the English catalogue by a mapped type that widens the literal types of the reference to `string`:

```ts
type Widen<T> = { [K in keyof T]: T[K] extends string ? string : Widen<T[K]> }
export type Messages = Widen<typeof en>
```

The Chinese catalogue is declared as `const zhCN: Messages`. A key that the Chinese catalogue does not carry is a compile error, and a key that the Chinese catalogue carries and the English one does not is a compile error as well, because the mapped type has no index signature.

The widening matters for the second half of the contract: without it, the Chinese value of a message would have to be the English string character for character, which is not translation.

### 4.2 The translator resolves a key in three steps

`createTranslator(locale)` returns a function `(key, params?) => string`:

1. The key is looked up in the catalogue of the locale. A key that is a plural message resolves to its `one` or its `other` form, chosen by `params.count`.
2. A key the locale does not carry falls back to the English catalogue.
3. A key no catalogue carries is returned as itself, so that the reader sees the name of the missing message instead of an empty space.

Placeholders are written `{name}` and are filled from the parameters. A placeholder with no value is left in place, because a message that shows `{id}` names the defect and one that shows nothing does not.

### 4.3 The store is a module, not a provider

The locale lives in a module-level variable with a listener set, and the hook reads it through `useSyncExternalStore`:

```ts
let current: Locale = readStoredLocale()
const listeners = new Set<() => void>()
export function setLocale(locale: Locale): void { /* ... */ }
export function useLocale(): Locale {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
}
```

A provider would have been the other choice. The store was chosen because the language is a property of the reader's session and not of the screen: a provider would put the language in the tree of every screen and would have to be re-created when the address changes, and the error boundary — a class component, because React has no hook form of it — could not read it at all.

The consequence of the choice is that the store is a module-level singleton. It is read by `useLocale` and written by `setLocale`, and nothing else touches it.

### 4.4 Persistence and the document

The choice is stored under `b-eval.locale` in `localStorage`. The storage is read once, when the module is loaded, and the reader's stored value wins over the default. A browser that refuses the storage — a private window, a locked-down profile — is handled by a `try` around every access: the interface then speaks the default language for the session and loses nothing but the choice.

The same module writes the document: `document.documentElement.lang` carries the locale and `document.title` carries the catalogue's title. The `lang` attribute is not decoration: it decides the font the browser picks for the text, the rules it applies to line breaking, and the language a screen reader announces.

### 4.5 The closed vocabularies

`values.ts` maps the values the server publishes to the word the reader sees. The mapping is per locale, and a value that is not listed is returned unchanged.

This is the one place where the platform's contract meets the reader's language, and the rule is that the wire value never changes: the server publishes `no coverage`, the Chinese interface shows that value in Chinese, and the report artefact exported from the platform carries the canonical value in both languages.

## 5. The Switch

The switch lives in the bezel bar, to the right of the navigation, because the language is a property of the reader and not of the screen.

It is a pair of buttons rather than a select. A select hides the other language behind a click, and a reader who cannot read the current language cannot find the option that helps them. Each button is named in its own language — the Chinese endonym and the English one — which is the convention a reader of either language recognises.

The buttons carry `lang`, so that the browser applies the right font to the endonym itself, and `aria-pressed`, so that a screen reader states which language the interface is in.

## 6. Type and Layout for Chinese

The self-hosted faces carry no CJK glyphs, so the Chinese font of the reader's system follows them in the stack:

```css
--sans: 'Archivo', ui-sans-serif, system-ui, -apple-system, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
  'Microsoft YaHei', 'Noto Sans CJK SC', 'Source Han Sans SC', sans-serif;
```

A Latin word still comes from Archivo, because it comes first in the stack; a Chinese word falls through to the system font.

The Latin type scale carries a negative `letter-spacing`, which opens up the word spacing of a Latin word. A Chinese glyph is square and dense already, and a negative tracking crowds it, so the Chinese document carries no tracking on the type scale:

```css
html[lang='zh-CN'] .page__title,
html[lang='zh-CN'] .section__title,
html[lang='zh-CN'] .table th { letter-spacing: 0; }
```

## 7. The English-Only Gate

The repository is written in English, and the `docs-english` gate enforces it on every file it scans. The Chinese catalogue is the one text in the tree that is not English, so the gate skips the directory that holds it:

```go
var translationDirs = []string{
	filepath.Join("web", "src", "i18n"),
}
```

The exemption is a path, not a file extension and not a language: nothing else in the tree may hold a translation, and the keys of the catalogue stay English.

## 8. Verification

| Property | How it is checked |
|---|---|
| The key sets agree | The `Messages` type, at compile time, and a test that compares the flattened key sets of the two catalogues |
| The default language is Chinese | A test that clears the stored choice and reads the language of a fresh module |
| The choice is remembered | A test that writes the choice, re-imports the module and reads the language back |
| The change reaches every screen | The switch test, which changes the language and reads the text of a screen |
| The document declares its language | A test that reads `document.documentElement.lang` after a change |
| A value that no vocabulary covers is unchanged | The switch test's probe, which renders an unknown value |
| The gate exempts the catalogue alone | Two gate tests: the catalogue passes, and Chinese text outside it fails |
