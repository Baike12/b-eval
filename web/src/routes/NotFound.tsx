import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Empty } from '../components/common'
import { useT } from '../i18n'

// NotFound answers an address that matches nothing in the platform.
//
// A router that matches no route renders nothing at all, and an empty
// document cannot be told apart from a screen that is still loading. The
// platform names the address and offers the way back.
export function NotFound(): ReactNode {
  const t = useT()
  return (
    <>
      <header className="page__head">
        <h1 className="page__title">{t('notFound.title')}</h1>
        <p className="page__lede">{t('notFound.lede')}</p>
      </header>
      <div className="sheet">
        <Empty hint={t('notFound.hint')}>{t('notFound.empty')}</Empty>
      </div>
      <p className="notice">
        <Link className="link" to="/">
          {t('notFound.back')}
        </Link>
      </p>
    </>
  )
}
