import type { ReactNode } from 'react'
import { NavLink, Outlet } from 'react-router-dom'

// Layout is the frame every screen sits in.
//
// The navigation order follows the platform's own order of operations:
// tasks first, then the agents, then the runs and their reports, then the
// metric registry.
export function Layout(): ReactNode {
  return (
    <div className="layout">
      <header className="layout__header">
        <span className="layout__brand">b-eval</span>
        <span className="layout__tagline">agent evaluation platform</span>
      </header>
      <nav className="layout__nav">
        <NavLink to="/" end>
          Dashboard
        </NavLink>
        <NavLink to="/task-sets">Task sets</NavLink>
        <NavLink to="/agents">Agents</NavLink>
        <NavLink to="/runs">Runs</NavLink>
        <NavLink to="/metrics">Metrics</NavLink>
      </nav>
      <main className="layout__main">
        <Outlet />
      </main>
    </div>
  )
}
