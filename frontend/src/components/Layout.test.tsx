import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import Layout from './Layout'
import authReducer from '../store/slices/authSlice'

function renderWithProviders(ui: React.ReactElement, preloadedState = {}) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState,
  })
  return render(
    <Provider store={store}>
      <MemoryRouter>{ui}</MemoryRouter>
    </Provider>
  )
}

describe('Layout component', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('renders DataLens brand', () => {
    renderWithProviders(<Layout />, {
      auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
    })
    expect(screen.getByText('DataLens')).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    renderWithProviders(<Layout />, {
      auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
    })
    expect(screen.getByRole('link', { name: /dashboard/i })).toHaveAttribute('href', '/')
    expect(screen.getByRole('link', { name: /upload/i })).toHaveAttribute('href', '/upload')
    expect(screen.getByRole('link', { name: /search/i })).toHaveAttribute('href', '/search')
  })

  it('shows username', () => {
    renderWithProviders(<Layout />, {
      auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
    })
    expect(screen.getByText('alice')).toBeInTheDocument()
  })

  it('shows Logout button', () => {
    renderWithProviders(<Layout />, {
      auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
    })
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
  })

  it('clicking Logout dispatches logout action and clears state', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: {
        auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
      },
    })

    render(
      <Provider store={store}>
        <MemoryRouter><Layout /></MemoryRouter>
      </Provider>
    )

    fireEvent.click(screen.getByRole('button', { name: /logout/i }))

    const state = store.getState().auth
    expect(state.token).toBeNull()
    expect(state.username).toBeNull()
    expect(state.email).toBeNull()
  })

  it('renders main element for Outlet', () => {
    renderWithProviders(<Layout />, {
      auth: { token: 'tok', username: 'alice', email: 'alice@test.com', loading: false, error: null },
    })
    expect(screen.getByRole('main')).toBeInTheDocument()
  })
})
