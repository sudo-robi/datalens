import { describe, it, expect } from 'vitest'
import { store } from './index'

describe('store', () => {
  it('has auth, datasets, and search reducers', () => {
    const state = store.getState()
    expect(state).toHaveProperty('auth')
    expect(state).toHaveProperty('datasets')
    expect(state).toHaveProperty('search')
  })

  it('has correct initial auth state', () => {
    const state = store.getState()
    expect(state.auth).toHaveProperty('token')
    expect(state.auth).toHaveProperty('loading')
    expect(state.auth).toHaveProperty('error')
  })

  it('has correct initial datasets state', () => {
    const state = store.getState()
    expect(state.datasets.items).toEqual([])
    expect(state.datasets.current).toBeNull()
    expect(state.datasets.loading).toBe(false)
  })

  it('has correct initial search state', () => {
    const state = store.getState()
    expect(state.search.query).toBe('')
    expect(state.search.results).toBeNull()
    expect(state.search.loading).toBe(false)
  })

  it('store.dispatch is a function', () => {
    expect(typeof store.dispatch).toBe('function')
  })
})
