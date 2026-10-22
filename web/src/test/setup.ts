import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// The auto-cleanup of the testing library only registers itself when the
// test runner exposes the globals. This project does not, so the cleanup is
// registered explicitly: without it a rendered tree stays in the document
// and the next assertion finds the elements of the previous test.
afterEach(() => {
  cleanup()
})
