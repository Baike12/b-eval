import type { ReactNode } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useT } from '../i18n'
import { ErrorBoundary } from './ErrorBoundary'
import { LocaleSwitch } from './LocaleSwitch'

// Layout is the frame every screen sits in.
//
// The navigation order follows the platform's own order of operations:
// tasks first, then the agents, then the runs and their reports, then the
// metric registry.
export function Layout(): ReactNode {
  // The boundary is remounted on every path change, so a screen that failed
  // to render does not keep the next one blank.
  const location = useLocation()
  const t = useT()
  return (
    <div className="app">
      <header className="bezel">
        <div className="bezel__bar">
          <span className="brand">
            <svg className="brand__mark" viewBox="0 0 22 22" aria-hidden="true">
              <rect className="brand__frame" x="1" y="1" width="20" height="20" rx="5" />
              <path className="brand__bars" d="M6.5 14.5V9.5M11 14.5V6.5M15.5 14.5V11.5" />
            </svg>
            <span className="brand__name">b-eval</span>
            <span className="brand__tag">{t('brand.tag')}</span>
          </span>
          <nav className="tabs">
            <NavLink to="/" end className={({ isActive }) => (isActive ? 'tab is-active' : 'tab')}>
              {t('nav.dashboard')}
            </NavLink>
            <NavLink to="/task-sets" className={({ isActive }) => (isActive ? 'tab is-active' : 'tab')}>
              {t('nav.taskSets')}
            </NavLink>
            <NavLink to="/agents" className={({ isActive }) => (isActive ? 'tab is-active' : 'tab')}>
              {t('nav.agents')}
            </NavLink>
            <NavLink to="/runs" className={({ isActive }) => (isActive ? 'tab is-active' : 'tab')}>
              {t('nav.runs')}
            </NavLink>
            <NavLink to="/metrics" className={({ isActive }) => (isActive ? 'tab is-active' : 'tab')}>
              {t('nav.metrics')}
            </NavLink>
          </nav>
          <LocaleSwitch />
        </div>
      </header>
      <main className="page">
        <ErrorBoundary key={location.pathname}>
          <Outlet />
        </ErrorBoundary>
      </main>
    </div>
  )
}
