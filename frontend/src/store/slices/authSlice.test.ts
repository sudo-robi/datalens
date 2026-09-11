import { describe, it, expect, beforeEach, vi } from 'vitest'
import authReducer, { logout, clearError, login, register } from './authSlice'

interface AuthState {
  token: string | null
  username: string | null
  email: string | null
  loading: boolean
  error: string | null
}

vi.mock('@/api/client', () => ({
  api: {
    post: vi.fn(),
  },
}))

const { api } = await import('@/api/client') as { api: { post: ReturnType<typeof vi.fn> } }

describe('authSlice', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('has null token/username/email when localStorage is empty', () => {
      const state = authReducer(undefined, { type: 'unknown' })
      expect(state.token).toBeNull()
      expect(state.username).toBeNull()
      expect(state.email).toBeNull()
      expect(state.loading).toBe(false)
      expect(state.error).toBeNull()
    })
  })

  describe('logout', () => {
    it('clears all auth fields and localStorage', () => {
      localStorage.setItem('token', 'tok123')
      localStorage.setItem('username', 'alice')
      localStorage.setItem('email', 'alice@test.com')

      const preState: AuthState = {
        token: 'tok123',
        username: 'alice',
        email: 'alice@test.com',
        loading: false,
        error: null,
      }

      const state = authReducer(preState, logout())

      expect(state.token).toBeNull()
      expect(state.username).toBeNull()
      expect(state.email).toBeNull()
      expect(localStorage.getItem('token')).toBeNull()
      expect(localStorage.getItem('username')).toBeNull()
      expect(localStorage.getItem('email')).toBeNull()
    })
  })

  describe('clearError', () => {
    it('clears error while preserving other fields', () => {
      const preState: AuthState = {
        token: 'tok123',
        username: 'alice',
        email: 'alice@test.com',
        loading: false,
        error: 'Some error',
      }

      const state = authReducer(preState, clearError())

      expect(state.error).toBeNull()
      expect(state.token).toBe('tok123')
    })
  })

  describe('login.pending', () => {
    it('sets loading true and clears error', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: false,
        error: 'old error',
      }

      const state = authReducer(preState, { type: login.pending.type })

      expect(state.loading).toBe(true)
      expect(state.error).toBeNull()
    })
  })

  describe('login.fulfilled', () => {
    it('sets token, username, email and stores in localStorage', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const payload = { token: 'abc123', username: 'bob', email: 'bob@test.com' }
      const state = authReducer(preState, { type: login.fulfilled.type, payload })

      expect(state.loading).toBe(false)
      expect(state.token).toBe('abc123')
      expect(state.username).toBe('bob')
      expect(state.email).toBe('bob@test.com')
      expect(localStorage.getItem('token')).toBe('abc123')
      expect(localStorage.getItem('username')).toBe('bob')
      expect(localStorage.getItem('email')).toBe('bob@test.com')
    })
  })

  describe('login.rejected', () => {
    it('sets error message from payload', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const state = authReducer(preState, {
        type: login.rejected.type,
        error: { message: 'Invalid credentials' },
      })

      expect(state.loading).toBe(false)
      expect(state.error).toBe('Invalid credentials')
    })

    it('uses default error message when none provided', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const state = authReducer(preState, {
        type: login.rejected.type,
        error: {},
      })

      expect(state.error).toBe('Login failed')
    })
  })

  describe('register.pending', () => {
    it('sets loading true and clears error', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: false,
        error: 'old error',
      }

      const state = authReducer(preState, { type: register.pending.type })

      expect(state.loading).toBe(true)
      expect(state.error).toBeNull()
    })
  })

  describe('register.fulfilled', () => {
    it('sets token, username, email and stores in localStorage', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const payload = { token: 'reg-tok', username: 'carol', email: 'carol@test.com' }
      const state = authReducer(preState, { type: register.fulfilled.type, payload })

      expect(state.loading).toBe(false)
      expect(state.token).toBe('reg-tok')
      expect(state.username).toBe('carol')
      expect(state.email).toBe('carol@test.com')
      expect(localStorage.getItem('token')).toBe('reg-tok')
      expect(localStorage.getItem('username')).toBe('carol')
      expect(localStorage.getItem('email')).toBe('carol@test.com')
    })
  })

  describe('register.rejected', () => {
    it('sets error message from payload', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const state = authReducer(preState, {
        type: register.rejected.type,
        error: { message: 'Username taken' },
      })

      expect(state.loading).toBe(false)
      expect(state.error).toBe('Username taken')
    })

    it('uses default error message when none provided', () => {
      const preState: AuthState = {
        token: null,
        username: null,
        email: null,
        loading: true,
        error: null,
      }

      const state = authReducer(preState, {
        type: register.rejected.type,
        error: {},
      })

      expect(state.error).toBe('Registration failed')
    })
  })

  describe('login thunk', () => {
    it('dispatches fulfilled with auth data on success', async () => {
      const mockData = { token: 'jwt-123', username: 'dave', email: 'dave@test.com' }
      api.post.mockResolvedValueOnce({ data: mockData })

      const { default: reducer } = await import('./authSlice')
      const thunk = login({ username: 'dave', password: 'pass123' })
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: login.pending.type }))
      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: login.fulfilled.type, payload: mockData }))
    })

    it('dispatches rejected on API failure', async () => {
      api.post.mockRejectedValueOnce({ response: { data: { detail: 'Bad creds' } } })

      const thunk = login({ username: 'dave', password: 'wrong' })
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: login.pending.type }))
      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: login.rejected.type }))
    })
  })

  describe('register thunk', () => {
    it('dispatches fulfilled with auth data on success', async () => {
      const mockData = { token: 'jwt-reg', username: 'eve', email: 'eve@test.com' }
      api.post.mockResolvedValueOnce({ data: mockData })

      const thunk = register({ username: 'eve', email: 'eve@test.com', password: 'pass123' })
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: register.pending.type }))
      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: register.fulfilled.type, payload: mockData }))
    })

    it('dispatches rejected on API failure', async () => {
      api.post.mockRejectedValueOnce({ response: { data: { detail: 'Email exists' } } })

      const thunk = register({ username: 'eve', email: 'eve@test.com', password: 'pass123' })
      const dispatch = vi.fn()
      const getState = vi.fn()

      await thunk(dispatch, getState, undefined)

      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: register.pending.type }))
      expect(dispatch).toHaveBeenCalledWith(expect.objectContaining({ type: register.rejected.type }))
    })
  })
})
