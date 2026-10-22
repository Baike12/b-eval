import { Component, type ErrorInfo, type ReactNode } from 'react'
import { useT } from '../i18n'

// ErrorBoundary turns a render failure into a visible statement.
//
// A white page is the worst outcome an evaluation report can have: the
// reader cannot tell a broken screen from a screen with no data, and the
// failure leaves nothing to report. The boundary keeps the navigation alive
// and names the fault instead.
interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('the interface failed to render', error, info.componentStack)
  }

  render(): ReactNode {
    const { error } = this.state
    if (!error) {
      return this.props.children
    }
    return <RenderFailure message={error.message} />
  }
}

// RenderFailure states the fault in the language of the interface.
//
// The boundary is a class component and cannot read the current language, so
// the statement it renders is a function component.
function RenderFailure({ message }: { message: string }): ReactNode {
  const t = useT()
  return (
    <section className="sheet">
      <div className="sheet__body">
        <h2 className="section__title">{t('error.title')}</h2>
        <p className="empty">
          {t('error.body')}
          <span className="empty__hint">{message}</span>
        </p>
      </div>
    </section>
  )
}
