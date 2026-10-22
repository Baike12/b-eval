import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App } from './App'
import './styles.css'

const container = document.getElementById('root')
if (!container) {
  throw new Error('the root element is missing from index.html')
}

// The application is mounted under the base path the bundle was built with,
// so a deployment can serve it from a prefix such as /eval/ instead of the
// root. React Router's basename must not carry a trailing slash.
const basePath = import.meta.env.BASE_URL
const basename = basePath === '/' ? '/' : basePath.replace(/\/+$/, '')

createRoot(container).render(
  <StrictMode>
    <BrowserRouter basename={basename}>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
