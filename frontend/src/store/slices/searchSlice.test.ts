import { describe, it, expect, vi, beforeEach } from 'vitest'
import searchReducer, { setQuery, clearResults, searchDatasets } from './searchSlice'
import { api } from '@/api/client'

vi.mock('@/api/client', () => ({
  api: { post: vi.fn() },
}))

const initialState = {
  query: '',
  results: null,
  loading: false,
  error: null,
}

describe('searchSlice', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('has correct initial state', () => {
    const state = searchReducer(undefined, { type: '@@INIT' })
    expect(state).toEqual(initialState)
  })

  it('setQuery reducer sets query', () => {
    const state = searchReducer(initialState, setQuery('test query'))
    expect(state.query).toBe('test query')
  })

  it('clearResults reducer clears results and query', () => {
    const stateWithResults = {
      query: 'old query',
      results: { total: 5, hits: [{ a: 1 }], page: 1, per_page: 20 },
      loading: false,
      error: 'some error',
    }
    const state = searchReducer(stateWithResults, clearResults())
    expect(state.results).toBeNull()
    expect(state.query).toBe('')
  })

  it('searchDatasets.pending sets loading true and error null', () => {
    const stateWithErrors = { ...initialState, error: 'old error', loading: false }
    const state = searchReducer(stateWithErrors, searchDatasets.pending({ query: 'q' }))
    expect(state.loading).toBe(true)
    expect(state.error).toBeNull()
  })

  it('searchDatasets.fulfilled sets results and loading false', () => {
    const payload = { total: 10, hits: [{ id: 1 }], page: 1, per_page: 20 }
    const loadingState = { ...initialState, loading: true }
    const state = searchReducer(loadingState, searchDatasets.fulfilled(payload, '', { query: 'q' }))
    expect(state.loading).toBe(false)
    expect(state.results).toEqual(payload)
  })

  it('searchDatasets.rejected sets error message', () => {
    const loadingState = { ...initialState, loading: true }
    const error = { message: 'Network error', name: 'Error', code: 'ERR' }
    const state = searchReducer(loadingState, searchDatasets.rejected(error, '', { query: 'q' }))
    expect(state.loading).toBe(false)
    expect(state.error).toBe('Network error')
  })

  it('searchDatasets.rejected uses default message when error.message is missing', () => {
    const loadingState = { ...initialState, loading: true }
    const error = { message: undefined as unknown as string, name: 'Error', code: 'ERR' }
    const state = searchReducer(loadingState, searchDatasets.rejected(error, '', { query: 'q' }))
    expect(state.error).toBe('Search failed')
  })

  it('searchDatasets async thunk calls api.post and dispatches fulfilled', async () => {
    const { configureStore } = await import('@reduxjs/toolkit')
    const mockResponse = { data: { total: 2, hits: [{ a: 1 }, { a: 2 }], page: 1, per_page: 20 } }
    vi.mocked(api.post).mockResolvedValue(mockResponse)

    const store = configureStore({ reducer: { search: searchReducer } })
    await store.dispatch(searchDatasets({ query: 'test', page: 1 }))

    expect(api.post).toHaveBeenCalledWith('/search', {
      query: 'test',
      index: undefined,
      page: 1,
      per_page: 20,
    })
    const state = store.getState().search
    expect(state.results).toEqual(mockResponse.data)
    expect(state.loading).toBe(false)
  })

  it('searchDatasets async thunk dispatches rejected on error', async () => {
    const { configureStore } = await import('@reduxjs/toolkit')
    vi.mocked(api.post).mockRejectedValue(new Error('Network failure'))

    const store = configureStore({ reducer: { search: searchReducer } })
    await store.dispatch(searchDatasets({ query: 'test' }))

    const state = store.getState().search
    expect(state.error).toBe('Network failure')
    expect(state.loading).toBe(false)
  })
})
