import { useCallback, useEffect, useState } from 'react'

// AsyncState is the state of one load.
//
// The error is kept next to the data rather than in a separate variable, so
// a screen cannot render stale data next to a fresh error.
export interface AsyncState<T> {
  data: T | null
  error: string | null
  loading: boolean
}

// useAsync loads a value from an async function and re-loads it when the
// dependencies change.
//
// The generation counter guards against a slow response overwriting a newer
// one, which is the failure mode that makes a dashboard show yesterday's
// number next to today's error.
export function useAsync<T>(load: () => Promise<T>, deps: unknown[]): AsyncState<T> & { reload: () => void } {
  const [state, setState] = useState<AsyncState<T>>({ data: null, error: null, loading: true })
  const [generation, setGeneration] = useState(0)

  const reload = useCallback(() => setGeneration((g) => g + 1), [])

  useEffect(() => {
    let cancelled = false
    setState((s) => ({ ...s, loading: true, error: null }))
    load()
      .then((data) => {
        if (!cancelled) setState({ data, error: null, loading: false })
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setState({ data: null, error: err instanceof Error ? err.message : String(err), loading: false })
        }
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, generation])

  return { ...state, reload }
}
