// Locale is the identifier of an interface language.
//
// The identifier is a BCP 47 tag, because the same tag is written into the
// document's lang attribute and the browser reads it to pick the font and
// the line-breaking rules.
export type Locale = 'zh-CN' | 'en'

// DEFAULT_LOCALE is the language the interface speaks until a reader says
// otherwise.
//
// The platform's readers are Chinese-speaking, so Chinese is the default
// and English is the alternative, not the other way round.
export const DEFAULT_LOCALE: Locale = 'zh-CN'

// LOCALES lists the languages the interface ships.
export const LOCALES: Locale[] = ['zh-CN', 'en']

// STORAGE_KEY is where the reader's choice is kept.
//
// The choice belongs to the browser and not to the account: the interface
// language is a property of the reader's session, not of the data the
// platform holds.
export const STORAGE_KEY = 'b-eval.locale'
