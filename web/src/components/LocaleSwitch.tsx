import type { ReactNode } from 'react'
import { LOCALES, endonyms, setLocale, useLocale, useT } from '../i18n'

// LocaleSwitch chooses the language of the interface.
//
// The control is a pair of buttons rather than a select, because a select
// hides the other language behind a click and a reader who cannot read the
// current language cannot find the option that helps them.
export function LocaleSwitch(): ReactNode {
  const locale = useLocale()
  const t = useT()
  return (
    <div className="locale" role="group" aria-label={t('locale.label')}>
      {LOCALES.map((option) => (
        <button
          key={option}
          type="button"
          className={`locale__option${option === locale ? ' is-active' : ''}`}
          lang={option}
          aria-pressed={option === locale}
          title={t('locale.switchTo', { language: endonyms[option] })}
          onClick={() => setLocale(option)}
        >
          {endonyms[option]}
        </button>
      ))}
    </div>
  )
}
