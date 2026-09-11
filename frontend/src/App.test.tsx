import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import App from './App'
import authReducer from './store/slices/authSlice'
import datasetsReducer from './store/slices/datasetsSlice'
import searchReducer from './store/slices/searchSlice'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn().mockResolvedValue({ data: [] }), post: vi.fn(), delete: vi.fn() },
}))

function renderWithProviders(ui: React.ReactElement, initialRoute = '/', preloadedState = {}) {
  const store = configureStore({
    reducer: { auth: authReducer, datasets: datasetsReducer, search: searchReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={[initialRoute]}>{ui}</MemoryRouter>
    </Provider>
  )
}

describe('App routing', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('redirects unauthenticated user from / to /login', () => {
    renderWithProviders(<App />, '/', {
      auth: { token: null, username: null, email: null, loading: false, error: null },
    })
    expect(screen.getByRole('heading', { name: /datalens/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders Login page at /login', () => {
    renderWithProviders(<App />, '/login', {
      auth: { token: null, username: null, email: null, loading: false, error: null },
    })
    expect(screen.getByRole('heading', { name: /datalens/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders Register page at /register', () => {
    renderWithProviders(<App />, '/register', {
      auth: { token: null, username: null, email: null, loading: false, error: null },
    })
    expect(screen.getByRole('heading', { name: /create account/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign up/i })).toBeInTheDocument()
  })

  it('renders Dashboard for authenticated user at /', async () => {
    renderWithProviders(<App />, '/', {
      auth: { token: 'valid-token', username: 'testuser', email: 'test@test.com', loading: false, error: null },
      datasets: { items: [], current: null, loading: false, error: null },
    })
    expect(screen.getByText('DataLens')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
  })

  it('authenticated user sees Layout with navigation', () => {
    renderWithProviders(<App />, '/', {
      auth: { token: 'valid-token', username: 'testuser', email: 'test@test.com', loading: false, error: null },
      datasets: { items: [], current: null, loading: false, error: null },
    })
    expect(screen.getByText('DataLens')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /dashboard/i })).toBeInTheDocument()
    expect(screen.getAllByRole('link', { name: /upload/i }).length).toBeGreaterThanOrEqual(1)
    expect(screen.getByRole('link', { name: /search/i })).toBeInTheDocument()
  })

  it('unauthenticated user at /upload redirects to login', () => {
    renderWithProviders(<App />, '/upload', {
      auth: { token: null, username: null, email: null, loading: false, error: null },
    })
    expect(screen.getByRole('heading', { name: /datalens/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })
})
